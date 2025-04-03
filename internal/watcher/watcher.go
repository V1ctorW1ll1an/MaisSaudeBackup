package watcher

import (
	"context"
	"fmt"
	"log/slog"
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
				// Usar event.Has() é mais robusto para operações combinadas (embora Create seja simples)
				if event.Has(fsnotify.Create) {
					filePath := event.Name
					fw.logger.Info("Novo arquivo detectado", slog.String("path", filePath))

					cleanPath, absErr := filepath.Abs(filepath.Clean(filePath))
					if absErr != nil {
						fw.logger.Error("Erro ao obter caminho absoluto", slog.String("raw_path", filePath), slog.Any("error", absErr))
						cleanPath = filePath // Tenta continuar mesmo assim
					}

					// Lança upload em goroutine separada
					go fw.handleUpload(ctx, cleanPath)
				} else {
					fw.logger.Debug("Evento fsnotify ignorado", slog.String("path", event.Name), slog.String("op", event.Op.String()))
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
		// TODO: Mover/deletar arquivo local opcionalmente
	}
}
