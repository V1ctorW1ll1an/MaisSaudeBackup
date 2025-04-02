package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
	writter io.Writer
	logFile *os.File
}

// NewLogger creates a logger that writes to both stdout and a file in the specified directory.
func NewLogger(logDir string) (*Logger, error) {
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar diretório de logs %s: %w", logDir, err)
	}

	// Construct log file path (e.g., app_2023-10-27.log)
	// Using a fixed name for simplicity now: app.log
	logFilePath := filepath.Join(logDir, "app.log")

	// Open/Create the log file
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir arquivo de log %s: %w", logFilePath, err)
	}

	// Create a multi-writer to write to both stdout and the file
	writer := io.MultiWriter(os.Stdout, file)

	// Use standard log flags
	flags := log.Ldate | log.Ltime | log.Lshortfile

	return &Logger{
		debug:   log.New(writer, ">>> DEBUG: ", flags),
		info:    log.New(writer, ">>> INFO: ", flags),
		warning: log.New(writer, ">>> WARNING: ", flags),
		err:     log.New(writer, ">>> ERROR: ", flags),
		writter: writer,
		logFile: file,
	}, nil
}

// Close closes the log file handle.
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Create Non-Formatted Logs
func (l *Logger) Debug(v ...any) {
	l.debug.Println(v...)
}
func (l *Logger) Info(v ...any) {
	l.info.Println(v...)
}
func (l *Logger) Warn(v ...any) {
	l.warning.Println(v...)
}
func (l *Logger) Error(v ...any) {
	l.err.Println(v...)
}

// Create Formatted Logs
func (l *Logger) Debugf(f string, v ...any) {
	l.debug.Printf(f, v...)
}
func (l *Logger) Infof(f string, v ...any) {
	l.info.Printf(f, v...)
}
func (l *Logger) Warnf(f string, v ...any) {
	l.warning.Printf(f, v...)
}
func (l *Logger) Errorf(f string, v ...any) {
	l.err.Printf(f, v...)
}

// LogBackupStart logs the start of a backup process.
func (l *Logger) LogBackupStart(completedDate string, horario string) {
	l.info.Printf("Inicio do backup em %s %s", completedDate, horario)
}
