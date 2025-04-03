package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Setup inicializa um logger slog que escreve para stdout e para um arquivo.
// Retorna o logger, o handle do arquivo (para fechamento posterior) e um erro.
func Setup(logDir, levelStr string) (*slog.Logger, *os.File, error) {
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		fmt.Fprintf(os.Stderr, "Nível de log inválido '%s', usando 'info'\n", levelStr)
		level = slog.LevelInfo
	}

	err := os.MkdirAll(logDir, 0750) // Permissões rwxr-x---
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao criar diretório de log %s: %w", logDir, err)
	}

	logFilePath := filepath.Join(logDir, "uploader_app.log")

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640) // Permissões rw-r-----
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao abrir arquivo de log %s: %w", logFilePath, err)
	}

	writer := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Ajuda na depuração
	})

	logger := slog.New(handler)

	return logger, logFile, nil
}
