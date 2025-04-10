package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FolderWatcher monitora um diretório por eventos de criação de arquivos.
type FolderWatcher struct {
	logger   *slog.Logger
	uploader Uploader // Depende da interface, não da implementação concreta
	watchDir string
}

// NewFolderWatcher cria uma nova instância do monitor de pastas.
func NewFolderWatcher(logger *slog.Logger, uploader Uploader, watchDir string) *FolderWatcher {
	return &FolderWatcher{
		logger:   logger.With(slog.String("component", "FolderWatcher")),
		uploader: uploader,
		watchDir: watchDir,
	}
}

// Run inicia o processo de monitoramento e bloqueia até que o contexto seja cancelado.
func (fw *FolderWatcher) Run(ctx context.Context) error {
	fw.logger.Info("Iniciando monitoramento", slog.String("directory", fw.watchDir))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fw.logger.Error("Falha ao criar watcher fsnotify", slog.Any("error", err))
		return fmt.Errorf("criar watcher falhou: %w", err)
	}
	defer func() {
		fw.logger.Debug("Fechando watcher fsnotify...")
		if err := watcher.Close(); err != nil {
			fw.logger.Error("Erro ao fechar watcher fsnotify", slog.Any("error", err))
		} else {
			fw.logger.Info("Watcher fsnotify fechado.")
		}
	}()

	eventLoopDone := make(chan struct{})
	go func() {
		defer close(eventLoopDone)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					fw.logger.Info("Canal de eventos do watcher fechado.")
					return
				}
				// Usar event.Has() é mais robusto para operações combinadas
				// Vamos focar na CRIAÇÃO de arquivos .zip
				if event.Has(fsnotify.Create) { // Verificar evento de CRIAÇÃO
					filePath := event.Name
					fileExt := filepath.Ext(filePath)

					// Processar SOMENTE se for um arquivo .zip
					if fileExt == ".zip" {
						fw.logger.Info("Novo arquivo .zip detectado", slog.String("path", filePath))

						cleanPath, absErr := filepath.Abs(filepath.Clean(filePath))
						if absErr != nil {
							fw.logger.Error("Erro ao obter caminho absoluto", slog.String("raw_path", filePath), slog.Any("error", absErr))
							cleanPath = filePath // Tenta continuar mesmo assim
						}

						// Lança upload em goroutine separada
						go fw.handleUpload(ctx, cleanPath)
					} else {
						fw.logger.Debug("Evento de criação ignorado (não é .zip)", slog.String("path", filePath), slog.String("ext", fileExt))
					}
				} else {
					fw.logger.Debug("Evento fsnotify ignorado (não é Create)", slog.String("path", event.Name), slog.String("op", event.Op.String()))
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					fw.logger.Info("Canal de erros do watcher fechado.")
					return
				}
				fw.logger.Error("Erro no watcher fsnotify", slog.Any("error", err))
				// Poderia ter lógica para tentar re-adicionar o watch ou parar dependendo do erro

			case <-ctx.Done():
				fw.logger.Info("Recebido sinal de cancelamento. Encerrando loop de eventos do watcher.")
				return
			}
		}
	}()

	fw.logger.Debug("Adicionando diretório ao watcher", slog.String("directory", fw.watchDir))
	err = watcher.Add(fw.watchDir)
	if err != nil {
		fw.logger.Error("Falha ao adicionar diretório ao watcher", slog.String("directory", fw.watchDir), slog.Any("error", err))
		// Não precisa cancelar o contexto aqui, o retorno do erro fará main sair
		return fmt.Errorf("adicionar %s ao watcher falhou: %w", fw.watchDir, err)
	}
	fw.logger.Info("Monitoramento iniciado com sucesso.", slog.String("directory", fw.watchDir))

	// Aguarda o contexto ser cancelado (shutdown) ou o loop de eventos terminar
	select {
	case <-ctx.Done():
		fw.logger.Info("Sinal de shutdown recebido. Aguardando loop de eventos terminar...")
	case <-eventLoopDone:
		// Isso não deveria acontecer a menos que os canais do watcher fechem inesperadamente
		fw.logger.Warn("Loop de eventos do watcher terminou inesperadamente.")
	}

	// Espera a goroutine do loop realmente terminar antes de retornar
	<-eventLoopDone
	fw.logger.Info("Processo de monitoramento finalizado.")
	// Retorna o erro do contexto se foi cancelado, ou nil se saiu de outra forma (improvável)
	return ctx.Err()
}

// handleUpload é chamado em uma goroutine separada para fazer upload de um arquivo.
// (função não exportada)
func (fw *FolderWatcher) handleUpload(ctx context.Context, filePath string) {
	uploadLogger := fw.logger.With(slog.String("upload_file", filepath.Base(filePath)))
	uploadLogger.Debug("Iniciando processamento de upload")

	// Delay opcional - útil se arquivos são criados e escritos em etapas
	time.Sleep(2 * time.Second)

	// Verifica se o contexto já foi cancelado antes de tentar o upload
	if err := ctx.Err(); err != nil {
		uploadLogger.Warn("Upload cancelado antes de iniciar devido ao contexto", slog.Any("error", err))
		return
	}

	err := fw.uploader.UploadFile(ctx, filePath)
	if err != nil {
		// Erro já logado dentro de UploadFile (ou será logado se for de contexto)
		// Apenas loga a falha geral aqui
		uploadLogger.Error("Falha no upload", slog.Any("error", err))
		// TODO: Lógica de retentativa/notificação
	} else {
		uploadLogger.Info("Upload concluído com sucesso")

		// Obter o diretório onde o arquivo está
		dirPath := filepath.Dir(filePath)
		uploadLogger.Info("Iniciando limpeza pós-upload", slog.String("directory_to_clean", dirPath))

		// Chamar a função auxiliar para deletar os arquivos nesse diretório
		// Passa o logger principal ou o uploadLogger, dependendo do nível de detalhe desejado nos logs de exclusão
		deleteAllFilesInDir(dirPath, fw.logger)
	}
}

// deleteAllFilesInDir é uma função auxiliar para limpar os arquivos de um diretório.
func deleteAllFilesInDir(dirPath string, logger *slog.Logger) {
	logger = logger.With(slog.String("directory", dirPath)) // Adiciona contexto do diretório ao logger
	logger.Info("Iniciando tentativa de exclusão de todos os arquivos no diretório")

	// Lê todas as entradas (arquivos e subdiretórios) no diretório especificado
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("Falha ao listar conteúdo do diretório para exclusão", slog.Any("error", err))
		return // Não podemos prosseguir se não conseguirmos ler o diretório
	}

	deletedCount := 0
	errorCount := 0

	for _, entry := range entries {
		// Constrói o caminho completo para a entrada atual
		fullPath := filepath.Join(dirPath, entry.Name())

		// Verifica se a entrada NÃO é um diretório (ou seja, é um arquivo)
		if !entry.IsDir() {
			logger.Debug("Tentando excluir arquivo", slog.String("file", fullPath))
			err := os.Remove(fullPath)
			if err != nil {
				// Loga o erro mas continua tentando excluir outros arquivos
				logger.Warn("Falha ao excluir arquivo", slog.String("file", fullPath), slog.Any("error", err))
				errorCount++
			} else {
				logger.Info("Arquivo excluído com sucesso", slog.String("file", fullPath))
				deletedCount++
			}
		} else {
			// Apenas loga que está pulando um subdiretório
			logger.Debug("Ignorando subdiretório durante a exclusão", slog.String("subdir", fullPath))
		}
	}

	if errorCount > 0 {
		logger.Warn("Finalizada a exclusão de arquivos com erros", slog.Int("deleted_count", deletedCount), slog.Int("error_count", errorCount))
	} else {
		logger.Info("Finalizada a exclusão de arquivos com sucesso", slog.Int("deleted_count", deletedCount))
	}
}
