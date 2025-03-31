package main

import (
	"os"
	"path/filepath"

	"github.com/V1ctorW1ll1an/MaisSaudeBackup/config"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/api/drive/v3"
)

var (
	srv    *drive.Service
	logger *config.Logger
)

func handleFsNotifyEvent(event fsnotify.Event) {
	absPath, err := filepath.Abs(event.Name)
	if err != nil {
		logger.Errorf("Erro ao obter caminho absoluto: %v", err)
		absPath = event.Name
	}

	// Limpa o caminho para garantir consistência
	cleanPath := filepath.Clean(absPath)

	switch event.Op {
	case fsnotify.Create:
		logger.Infof("Novo arquivo criado: %s", cleanPath)
		LogFileMetadata(cleanPath)
		err := config.UploadFileToDrive(srv, cleanPath)
		if err != nil {
			logger.Errorf("Erro no upload do arquivo para o google drive: %v", err)
		}
	case fsnotify.Remove:
		logger.Infof("Arquivo removido: %s", cleanPath)
	case fsnotify.Rename:
		logger.Infof("Arquivo renomeado: %s", cleanPath)
	case fsnotify.Write:
		logger.Infof("Arquivo modificado: %s", cleanPath)
	}
}

func main() {
	folderPath := "./backups"
	logger = config.NewLogger("MAIN")
	var err error

	srv, err = config.GoogleDriveSetup()
	if err != nil {
		logger.Errorf("Erro ao fazer setup do google drive: %v", err)
	}

	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		logger.Errorf("Erro ao criar pasta: %v", err)
	}

	logger.Infof("Monitorando pasta: %s\n", folderPath)

	err = WatchFolder(folderPath, handleFsNotifyEvent)
	if err != nil {
		logger.Errorf("Erro ao iniciar monitoramento: %v", err)
	}
}
