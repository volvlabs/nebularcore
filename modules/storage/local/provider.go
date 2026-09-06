package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/volvlabs/nebularcore/modules/storage/models"
)

type Provider struct {
	basePath string
	baseURL  string
}

func isWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}
	mode := info.Mode()
	return mode&0200 != 0
}

func New(basePath, baseURL string) (*Provider, error) {
	if info, err := os.Stat(basePath); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("path exists but is not a directory: %s", basePath)
		}
		if !isWritable(basePath) {
			return nil, fmt.Errorf("directory is not writable: %s", basePath)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create base directory: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to access base directory: %w", err)
	}

	return &Provider{
		basePath: basePath,
		baseURL:  strings.TrimRight(baseURL, "/"),
	}, nil
}

func (p *Provider) Upload(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
	if input.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filepath.Join(p.basePath, input.Key))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a temporary file for upload
	tempFile, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	// Calculate hash while copying
	hash := sha256.New()
	writer := io.MultiWriter(tempFile, hash)

	size, err := io.Copy(writer, input.File)
	if err != nil {
		_ = tempFile.Close()
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Generate content hash
	contentHash := hex.EncodeToString(hash.Sum(nil))

	// Determine content type
	contentType := input.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(input.FileName))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// Create metadata file
	metadata := map[string]string{
		"originalName": input.FileName,
		"contentHash":  contentHash,
		"uploadTime":   time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range input.Metadata {
		metadata[k] = v
	}

	targetPath := filepath.Join(p.basePath, input.Key)
	metadataPath := targetPath + ".meta"

	// Write metadata file
	if err := p.writeMetadata(metadataPath, metadata); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempFile.Name(), targetPath); err != nil {
		_ = os.Remove(metadataPath)
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	return &models.UploadOutput{
		Path:        input.Key,
		URL:         p.getPublicURL(input.Key),
		ContentType: contentType,
		Size:        size,
		Metadata:    metadata,
	}, nil
}

func (p *Provider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (p *Provider) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(p.basePath, path)
	metadataPath := fullPath + ".meta"

	// Delete metadata file first
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// Delete the actual file
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (p *Provider) List(ctx context.Context, prefix string) ([]models.FileInfo, error) {
	fullPath := filepath.Join(p.basePath, prefix)

	// Check if directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Return empty list for non-existent directory
		return []models.FileInfo{}, nil
	}

	var files []models.FileInfo

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip metadata files and directories
		if strings.HasSuffix(path, ".meta") || info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(p.basePath, path)
		if err != nil {
			return err
		}

		// Skip if not under prefix
		if !strings.HasPrefix(relPath, prefix) {
			return nil
		}

		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		files = append(files, models.FileInfo{
			Path:        relPath,
			Size:        info.Size(),
			ContentType: contentType,
			ModTime:     info.ModTime(),
			IsDir:       false,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

func (p *Provider) writeMetadata(path string, metadata map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	for k, v := range metadata {
		if _, err := fmt.Fprintf(file, "%s=%s\n", k, v); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) getPublicURL(key string) string {
	if p.baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", p.baseURL, key)
}

// Helper function to serve files via HTTP
func (p *Provider) ServeFile(w http.ResponseWriter, r *http.Request, key string) {
	fullPath := filepath.Join(p.basePath, key)

	// Prevent directory traversal
	if !strings.HasPrefix(fullPath, p.basePath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Don't serve directories
	if info.IsDir() {
		http.Error(w, "Cannot serve directory", http.StatusBadRequest)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}
