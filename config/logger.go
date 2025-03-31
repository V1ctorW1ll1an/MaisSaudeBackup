package config

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
	writter io.Writer
}

func NewLogger(p string) *Logger {
	writer := io.Writer(os.Stdout)
	logger := log.New(writer, p, log.Ldate|log.Ltime)

	return &Logger{
		debug:   log.New(writer, ">>> DEBUG: ", logger.Flags()),
		info:    log.New(writer, ">>> INFO: ", logger.Flags()),
		warning: log.New(writer, ">>> WARNING: ", logger.Flags()),
		err:     log.New(writer, ">>> ERROR: ", logger.Flags()),
		writter: writer,
	}
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
