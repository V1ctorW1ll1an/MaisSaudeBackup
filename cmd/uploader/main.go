package main

import (
	"os"

	"github.com/V1ctorW1ll1an/MaisSaudeBackup/config/logger"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/pkg/files"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/pkg/gdrive"
)

var (
	backupPath = "./backups"
	logDir     = "./logs"
)

func main() {
	l, err := logger.New(logDir)
	if err != nil {
		l.Errorf("Falha ao inicializar o logger: %v", err)
		os.Exit(1)
	}
	defer l.Close()

	ds, err := gdrive.New()
	if err != nil {
		l.Errorf("Erro ao Inicializar o google drive: %v", err)
		//TODO: Send email or notification to the system manager
		os.Exit(1)
	}

	l.Infof("Monitorando pasta: %s\n", backupPath)

	// watchFolder will watch the backupPath, the only argument is the path, the function handleFsNotifyEvent will be called internally
	err = files.WatchFolder(backupPath, ds)
	if err != nil {
		l.Errorf("Erro ao iniciar monitoramento: %v", err)
		//TODO: Send email or notification to the system manager
		os.Exit(1)
	}
}
