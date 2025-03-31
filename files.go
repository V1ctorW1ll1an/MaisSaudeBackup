package main

import (
	"os"

	"github.com/fsnotify/fsnotify"
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
		logger.Infof("Erro ao obter informações do arquivo: %v", err)
		return err
	}

	logger.Info("Detalhes do arquivo:")
	logger.Infof("Nome: %s\n", f.Name())
	logger.Infof("Tamanho: %d bytes\n", f.Size())
	logger.Infof("Última modificação: %v\n", f.ModTime())
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
//   - f: A callback function that processes file system events.
//
// Returns:
//   - error: An error if the watcher fails to initialize or attach to the folder.
func WatchFolder(folderPath string, f func(fsnotify.Event)) error {
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
				f(event)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("Erro: %v", err)
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
