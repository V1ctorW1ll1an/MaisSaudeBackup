package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
	writter io.Writer
	logFile *os.File
}

// New creates a logger that writes to both stdout and a file in the specified directory.
func New(logDir string) (*Logger, error) {
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar diretório de logs %s: %w", logDir, err)
	}

	dataCompleta, horario := getFormattedDateTime()
	logFilePath := createLogFilePath(logDir, dataCompleta, horario)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir arquivo de log %s: %w", logFilePath, err)
	}

	writer := io.MultiWriter(os.Stdout, file)

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

// Gera strings formatadas para data e hora com base no tempo atual
func getFormattedDateTime() (string, string) {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	hour := now.Format("15")
	minute := now.Format("04")
	second := now.Format("05")

	dataCompleta := fmt.Sprintf("%s%s%s", year, month, day)
	horario := fmt.Sprintf("%s%s%s", hour, minute, second)

	return dataCompleta, horario
}

// Cria o caminho para o arquivo de log
func createLogFilePath(logDir string, dataCompleta string, horario string) string {
	return filepath.Join(logDir, fmt.Sprintf("log_backup_%s_%s.log", dataCompleta, horario))
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
