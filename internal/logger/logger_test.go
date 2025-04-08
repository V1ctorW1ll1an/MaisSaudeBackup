package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "logger_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name      string
		logDir    string
		levelStr  string
		wantLevel string
		wantErr   bool
	}{
		{
			name:      "valid debug level",
			logDir:    tempDir,
			levelStr:  "debug",
			wantLevel: "debug",
			wantErr:   false,
		},
		{
			name:      "valid info level",
			logDir:    tempDir,
			levelStr:  "info",
			wantLevel: "info",
			wantErr:   false,
		},
		{
			name:      "valid warn level",
			logDir:    tempDir,
			levelStr:  "warn",
			wantLevel: "warn",
			wantErr:   false,
		},
		{
			name:      "valid error level",
			logDir:    tempDir,
			levelStr:  "error",
			wantLevel: "error",
			wantErr:   false,
		},
		{
			name:      "invalid level defaults to info",
			logDir:    tempDir,
			levelStr:  "invalid",
			wantLevel: "info",
			wantErr:   false,
		},
		{
			name:      "invalid directory",
			logDir:    "/nonexistent/path",
			levelStr:  "info",
			wantLevel: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logFile, err := Setup(tt.logDir, tt.levelStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
				assert.Nil(t, logFile)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, logger)
			require.NotNil(t, logFile)

			// Verify log file was created
			logFilePath := filepath.Join(tt.logDir, "uploader_app.log")
			_, err = os.Stat(logFilePath)
			assert.NoError(t, err)

			// Clean up
			logFile.Close()
		})
	}
}

func TestSetup_LogFilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logger_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	_, logFile, err := Setup(tempDir, "info")
	require.NoError(t, err)
	defer logFile.Close()

	// Get file info
	logFilePath := filepath.Join(tempDir, "uploader_app.log")
	fileInfo, err := os.Stat(logFilePath)
	require.NoError(t, err)

	// Check file permissions
	// Expected: -rw-r-----
	expectedMode := os.FileMode(0640)
	assert.Equal(t, expectedMode, fileInfo.Mode().Perm(), "Log file permissions should be 0640")
}

func TestSetup_DirectoryPermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logger_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a subdirectory for logs
	logDir := filepath.Join(tempDir, "logs")
	_, _, err = Setup(logDir, "info")
	require.NoError(t, err)

	// Get directory info
	dirInfo, err := os.Stat(logDir)
	require.NoError(t, err)

	// Check directory permissions
	// Expected: drwxr-x---
	expectedMode := os.FileMode(0750)
	assert.Equal(t, expectedMode, dirInfo.Mode().Perm(), "Log directory permissions should be 0750")
}

func TestSetup_LogFileAppend(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logger_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// First setup
	logger1, logFile1, err := Setup(tempDir, "info")
	require.NoError(t, err)
	logger1.Info("First message")
	logFile1.Close()

	// Second setup - should append to existing file
	logger2, logFile2, err := Setup(tempDir, "info")
	require.NoError(t, err)
	logger2.Info("Second message")
	logFile2.Close()

	// Read the log file
	logFilePath := filepath.Join(tempDir, "uploader_app.log")
	content, err := os.ReadFile(logFilePath)
	require.NoError(t, err)

	// Check if both messages are present
	assert.Contains(t, string(content), "First message")
	assert.Contains(t, string(content), "Second message")
}
