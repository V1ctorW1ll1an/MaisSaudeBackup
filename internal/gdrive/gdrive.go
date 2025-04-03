package gdrive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveUploader encapsula a lógica de interação com o Google Drive.
// Ele satisfaz a interface watcher.Uploader.
type DriveUploader struct {
	logger  *slog.Logger
	service *drive.Service
}

// NewDriveUploader cria e configura um novo cliente para a API do Google Drive.
func NewDriveUploader(ctx context.Context, logger *slog.Logger, credentialsFile, tokenFile string) (*DriveUploader, error) {
	log := logger.With(slog.String("component", "DriveUploader"))

	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Error("Não foi possível ler o arquivo de credenciais", slog.String("path", credentialsFile), slog.Any("error", err))
		return nil, fmt.Errorf("leitura de %s falhou: %w", credentialsFile, err)
	}

	// Use drive.DriveFileScope para acesso limitado ou drive.DriveScope para acesso total
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Error("Não foi possível parsear o arquivo de credenciais", slog.String("path", credentialsFile), slog.Any("error", err))
		return nil, fmt.Errorf("parse de %s falhou: %w", credentialsFile, err)
	}

	client, err := getOAuthClient(ctx, log, config, tokenFile)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter cliente OAuth: %w", err)
	}

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Error("Não foi possível criar o serviço do Drive", slog.Any("error", err))
		return nil, fmt.Errorf("criação do serviço Drive falhou: %w", err)
	}

	log.Info("Serviço Google Drive inicializado com sucesso.")
	return &DriveUploader{logger: log, service: driveService}, nil
}

// UploadFile envia um arquivo para o Google Drive. Satisfaz watcher.Uploader.
func (du *DriveUploader) UploadFile(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		du.logger.Error("Erro ao abrir arquivo local para upload", slog.String("path", filePath), slog.Any("error", err))
		return fmt.Errorf("abrir arquivo %s falhou: %w", filePath, err)
	}
	defer file.Close()

	fileInfo, statErr := file.Stat()
	logAttrs := []any{slog.String("path", filePath)}
	if statErr == nil {
		logAttrs = append(logAttrs, slog.Int64("size", fileInfo.Size()))
	} else {
		du.logger.Warn("Não foi possível obter metadados do arquivo local", slog.String("path", filePath), slog.Any("error", statErr))
	}
	du.logger.Debug("Iniciando upload", logAttrs...)

	driveFile := &drive.File{
		Name: filepath.Base(filePath),
		// Parents: []string{"YOUR_FOLDER_ID"}, // Opcional: Adicione ID da pasta aqui
	}

	// Usa Context(ctx) para propagar cancelamento
	_, err = du.service.Files.Create(driveFile).Media(file).Context(ctx).Do()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			du.logger.Warn("Upload cancelado ou timeout", slog.String("path", filePath), slog.Any("error", err))
			// Retorna o erro original do contexto
			return err
		}
		du.logger.Error("Erro durante upload para o Google Drive", slog.String("path", filePath), slog.Any("error", err))
		return fmt.Errorf("upload de %s falhou: %w", filePath, err)
	}

	du.logger.Info("Arquivo enviado com sucesso para o Google Drive", slog.String("path", filePath))
	return nil
}
