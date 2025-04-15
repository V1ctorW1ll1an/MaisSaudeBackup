package gdrive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

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

	// Use drive.DriveScope para acesso total, incluindo criação de pastas
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
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

// createBackupFolder creates a new folder in Google Drive with the current date as name if it doesn't exist
func (du *DriveUploader) createBackupFolder(ctx context.Context) (string, error) {
	folderName := time.Now().Format("02-01-2006")
	du.logger.Info("Verificando pasta de backup", slog.String("folder", folderName))

	// First, check if the folder already exists
	query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", folderName)
	files, err := du.service.Files.List().
		Q(query).
		Fields("files(id, name, mimeType)").
		Context(ctx).
		Do()
	if err != nil {
		du.logger.Error("Erro ao buscar pasta existente", slog.String("folder", folderName), slog.Any("error", err))
		return "", fmt.Errorf("busca de pasta %s falhou: %w", folderName, err)
	}

	// Log all found folders for debugging
	if len(files.Files) > 0 {
		du.logger.Info("Pastas encontradas:", slog.Int("count", len(files.Files)))
		for i, f := range files.Files {
			du.logger.Info(fmt.Sprintf("Pasta %d:", i+1),
				slog.String("id", f.Id),
				slog.String("name", f.Name),
				slog.String("mimeType", f.MimeType))
		}
	}

	// If folder exists, verify it's actually a folder and return its ID
	if len(files.Files) > 0 {
		for _, f := range files.Files {
			if f.MimeType == "application/vnd.google-apps.folder" {
				du.logger.Info("Pasta válida encontrada",
					slog.String("folder", folderName),
					slog.String("id", f.Id))
				return f.Id, nil
			}
		}
	}

	du.logger.Info("Nenhuma pasta válida encontrada, criando nova pasta", slog.String("folder", folderName))

	// If folder doesn't exist, create it
	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	createdFolder, err := du.service.Files.Create(folder).
		Fields("id, name, mimeType").
		Context(ctx).
		Do()
	if err != nil {
		du.logger.Error("Erro ao criar pasta no Google Drive",
			slog.String("folder", folderName),
			slog.Any("error", err),
			slog.String("error_type", fmt.Sprintf("%T", err)))
		return "", fmt.Errorf("criação de pasta %s falhou: %w", folderName, err)
	}

	du.logger.Info("Pasta criada com sucesso no Google Drive",
		slog.String("folder", folderName),
		slog.String("id", createdFolder.Id),
		slog.String("mimeType", createdFolder.MimeType))
	return createdFolder.Id, nil
}

// deleteOldBackupFolder deletes the backup folder from two days ago
func (du *DriveUploader) deleteOldBackupFolder(ctx context.Context) error {
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("02-01-2006")

	// Search for the folder from two days ago
	query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder'", twoDaysAgo)
	files, err := du.service.Files.List().Q(query).Context(ctx).Do()
	if err != nil {
		du.logger.Error("Erro ao buscar pasta antiga", slog.String("folder", twoDaysAgo), slog.Any("error", err))
		return fmt.Errorf("busca de pasta %s falhou: %w", twoDaysAgo, err)
	}

	if len(files.Files) == 0 {
		du.logger.Debug("Nenhuma pasta antiga encontrada para deletar", slog.String("folder", twoDaysAgo))
		return nil
	}

	// Delete the folder
	err = du.service.Files.Delete(files.Files[0].Id).Context(ctx).Do()
	if err != nil {
		du.logger.Error("Erro ao deletar pasta antiga", slog.String("folder", twoDaysAgo), slog.Any("error", err))
		return fmt.Errorf("deleção de pasta %s falhou: %w", twoDaysAgo, err)
	}

	du.logger.Info("Pasta antiga deletada com sucesso", slog.String("folder", twoDaysAgo))
	return nil
}

// UploadFile envia um arquivo para o Google Drive. Satisfaz watcher.Uploader.
func (du *DriveUploader) UploadFile(ctx context.Context, filePath string) error {
	// Create new backup folder
	folderID, err := du.createBackupFolder(ctx)
	if err != nil {
		return fmt.Errorf("falha ao criar pasta de backup: %w", err)
	}

	// Delete old backup folder
	if err := du.deleteOldBackupFolder(ctx); err != nil {
		du.logger.Warn("Falha ao deletar pasta antiga, continuando com upload", slog.Any("error", err))
	}

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
		Name:    filepath.Base(filePath),
		Parents: []string{folderID},
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
