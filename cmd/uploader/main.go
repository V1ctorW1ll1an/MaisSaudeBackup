package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	// Importa os pacotes internos usando o path do módulo definido no go.mod
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/config"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/gdrive"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/logger"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/watcher"
)

func main() {
	// Parsear TODOS os flags definidos na aplicação
	flag.Parse()

	// Criar a configuração usando os valores parseados
	cfg, err := config.NewConfig()
	if err != nil {
		log.Println("Erro ao carregar a configuração:")
		flag.Usage() // Mostrar ajuda se a validação em NewConfig falhar
		log.Fatalf("Erro: %v", err)
	}

	// Setup Logger
	appLogger, logFile, err := logger.Setup(cfg.LogDir, cfg.LogLevel)
	if err != nil {
		// Tenta logar no stderr se o logger falhou
		fmt.Fprintf(os.Stderr, "Erro crítico ao inicializar logger: %v\n", err)
		os.Exit(1)
	}
	// Fecha o arquivo de log ao sair
	if logFile != nil {
		defer func() {
			appLogger.Debug("Fechando arquivo de log...")
			if err := logFile.Close(); err != nil {
				// Loga no logger que ainda deve estar funcional (só o file handle fecha)
				appLogger.Error("Erro ao fechar arquivo de log", slog.Any("error", err))
			}
		}()
	}
	appLogger.Info("Logger inicializado", slog.String("level", cfg.LogLevel), slog.String("log_dir", cfg.LogDir))

	// Setup Contexto e Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop() // Libera o listener de sinais ao sair

	appLogger.Info("Aplicação iniciada.", slog.String("pid", fmt.Sprintf("%d", os.Getpid())))
	appLogger.Info("Pressione Ctrl+C para sair...")

	// Goroutine separada para lidar com o cancelamento explícito e logar
	go func() {
		<-ctx.Done() // Espera o contexto ser cancelado
		appLogger.Warn("Sinal de shutdown recebido, iniciando finalização suave...")
		// stop() já foi chamado pelo defer ou pelo NotifyContext, não precisa chamar de novo
	}()

	// Setup Google Drive Uploader
	uploader, err := gdrive.NewDriveUploader(ctx, appLogger, cfg.CredentialsFile, cfg.TokenFile)
	if err != nil {
		appLogger.Error("Falha ao inicializar Google Drive Uploader", slog.Any("error", err))
		os.Exit(1)
	}

	// Setup e Run Folder Watcher
	folderWatcher := watcher.NewFolderWatcher(appLogger, uploader, cfg.WatchDir)

	// Executa o watcher. Ele bloqueará até o contexto ser cancelado.
	if err := folderWatcher.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		// Loga apenas se o erro NÃO for de cancelamento (que é esperado no shutdown)
		appLogger.Error("Folder Watcher encerrou com erro inesperado", slog.Any("error", err))
		os.Exit(1) // Sai com erro
	}

	// Shutdown Completo
	appLogger.Info("Aplicação finalizada com sucesso.")
	os.Exit(0)
}
