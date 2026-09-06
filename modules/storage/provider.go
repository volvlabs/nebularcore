package storage

import (
	"context"
	"io"

	"github.com/volvlabs/nebularcore/modules/storage/models"
)

// StorageProvider defines the interface for storage operations
type StorageProvider interface {
	// Upload handles file upload with idempotency
	Upload(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error)

	// Download retrieves file content
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file
	Delete(ctx context.Context, path string) error

	// List returns files in a directory
	List(ctx context.Context, prefix string) ([]models.FileInfo, error)
}
