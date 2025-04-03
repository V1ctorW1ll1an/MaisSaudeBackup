package files

import (
	"os"
	"path/filepath"

	"github.com/V1ctorW1ll1an/MaisSaudeBackup/config/logger"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/pkg/gdrive"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/api/drive/v3"
)

var (
	l *logger.Logger
)

// LogFileMetadata logs the metadata of a given file, including its name, size, and last modification time.
//
// It retrieves the file information using os.Stat and logs relevant details.
// If the file does not exist or an error occurs while retrieving its metadata, an error message is logged and returned.
//
// Parameters:
//   - path: The path to the file whose metadata will be logged.
func LogFileMetadata(path string) error {
	f, err := os.Stat(path)
	if err != nil {
		l.Infof("Erro ao obter informações do arquivo: %v", err)
		return err
	}

	l.Info("Detalhes do arquivo:")
	l.Infof("Nome: %s\n", f.Name())
	l.Infof("Tamanho: %d bytes\n", f.Size())
	l.Infof("Última modificação: %v\n", f.ModTime())
	return nil
}

// WatchFolder sets up a file system watcher on the specified folder and listens for file events.
//
// It creates a new fsnotify watcher to monitor changes in the given folder. Whenever an event occurs,
// the provided callback function `f` is invoked with the event details.
//
// The function runs indefinitely, blocking on a channel (`done`) to keep the watcher active.
// If an error occurs while creating or adding the watcher, it is returned.
//
// Parameters:
//   - folderPath: The path of the folder to be monitored.
//   - ds: The Google Drive service to be used for uploading files.
//
// Returns:
//   - error: An error if the watcher fails to initialize or attach to the folder.
func WatchFolder(folderPath string, ds *drive.Service) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)

	// This goroutine runs an infinite loop to listen for file system events and errors.
	// The `select` statement continuously checks both channels (`watcher.Events` and `watcher.Errors`).
	// - If an event is received, it is passed to the callback function `f(event)`.
	// - If an error occurs, it is logged.
	// - If either channel is closed (`!ok`), the goroutine exits, stopping the event listener.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				handleFsNotifyEvent(event, ds)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				l.Errorf("Erro: %v", err)
			}
		}
	}()

	err = watcher.Add(folderPath)
	if err != nil {
		return err
	}

	// Keeps the function running indefinitely
	<-done
	return nil
}

func handleFsNotifyEvent(event fsnotify.Event, ds *drive.Service) {
	absPath, err := filepath.Abs(event.Name)
	if err != nil {
		l.Errorf("Erro ao obter caminho absoluto: %v", err)
		absPath = event.Name
	}

	// Limpa o caminho para garantir consistência
	cleanPath := filepath.Clean(absPath)

	switch event.Op {
	case fsnotify.Create:
		l.Infof("Novo arquivo criado: %s", cleanPath)
		LogFileMetadata(cleanPath)
		err := gdrive.UploadFileToDrive(ds, cleanPath)
		if err != nil {
			l.Errorf("Erro ao fazer upload do arquivo para o Google Drive: %v", err)
		}
	case fsnotify.Remove:
		l.Infof("Arquivo removido: %s", cleanPath)
	case fsnotify.Rename:
		l.Infof("Arquivo renomeado: %s", cleanPath)
	case fsnotify.Write:
		l.Infof("Arquivo modificado: %s", cleanPath)
	}
}
