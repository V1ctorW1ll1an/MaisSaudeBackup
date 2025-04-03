package watcher

import "context"

// Uploader define a interface necessária para fazer upload de arquivos.
// Isso desacopla o watcher da implementação específica (ex: Google Drive).
type Uploader interface {
	UploadFile(ctx context.Context, filePath string) error
}
