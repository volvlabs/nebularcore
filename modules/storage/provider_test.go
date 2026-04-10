package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/volvlabs/nebularcore/modules/storage/models"
)

// mockStorageProvider implements StorageProvider interface for testing
type mockStorageProvider struct {
	uploadFunc   func(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error)
	downloadFunc func(ctx context.Context, path string) (io.ReadCloser, error)
	deleteFunc   func(ctx context.Context, path string) error
	listFunc     func(ctx context.Context, prefix string) ([]models.FileInfo, error)
}

func (m *mockStorageProvider) Upload(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, input)
	}
	return nil, fmt.Errorf("upload not implemented")
}

func (m *mockStorageProvider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, path)
	}
	return nil, fmt.Errorf("download not implemented")
}

func (m *mockStorageProvider) Delete(ctx context.Context, path string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, path)
	}
	return fmt.Errorf("delete not implemented")
}

func (m *mockStorageProvider) List(ctx context.Context, prefix string) ([]models.FileInfo, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, prefix)
	}
	return nil, fmt.Errorf("list not implemented")
}

func TestStorageProviderInterface(t *testing.T) {
	ctx := context.Background()
	testContent := []byte("test content")
	testTime := time.Now()

	tests := []struct {
		name     string
		provider StorageProvider
		wantErr  bool
	}{
		{
			name: "successful operations",
			provider: &mockStorageProvider{
				uploadFunc: func(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
					return &models.UploadOutput{
						Path:        input.Key,
						URL:         "https://example.com/" + input.Key,
						ContentType: input.ContentType,
						Size:        int64(len(testContent)),
						Metadata:    input.Metadata,
					}, nil
				},
				downloadFunc: func(ctx context.Context, path string) (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(testContent)), nil
				},
				deleteFunc: func(ctx context.Context, path string) error {
					return nil
				},
				listFunc: func(ctx context.Context, prefix string) ([]models.FileInfo, error) {
					return []models.FileInfo{
						{
							Path:        prefix + "/file1.txt",
							Size:        100,
							ContentType: "text/plain",
							ModTime:     testTime,
							IsDir:       false,
						},
					}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "failing operations",
			provider: &mockStorageProvider{
				uploadFunc: func(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
					return nil, fmt.Errorf("upload failed")
				},
				downloadFunc: func(ctx context.Context, path string) (io.ReadCloser, error) {
					return nil, fmt.Errorf("download failed")
				},
				deleteFunc: func(ctx context.Context, path string) error {
					return fmt.Errorf("delete failed")
				},
				listFunc: func(ctx context.Context, prefix string) ([]models.FileInfo, error) {
					return nil, fmt.Errorf("list failed")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Upload
			input := &models.UploadInput{
				File:        bytes.NewReader(testContent),
				FileName:    "test.txt",
				ContentType: "text/plain",
				Key:         "test/file.txt",
				Metadata:    map[string]string{"test": "value"},
			}
			output, err := tt.provider.Upload(ctx, input)
			if tt.wantErr {
				if err == nil {
					t.Error("Upload() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("Upload() error = %v", err)
				}
				if output.Path != input.Key {
					t.Errorf("Upload() path = %v, want %v", output.Path, input.Key)
				}
				if output.ContentType != input.ContentType {
					t.Errorf("Upload() contentType = %v, want %v", output.ContentType, input.ContentType)
				}
			}

			// Test Download
			reader, err := tt.provider.Download(ctx, "test/file.txt")
			if tt.wantErr {
				if err == nil {
					t.Error("Download() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("Download() error = %v", err)
				} else {
					defer reader.Close()
					content, err := io.ReadAll(reader)
					if err != nil {
						t.Errorf("Failed to read downloaded content: %v", err)
					}
					if !bytes.Equal(content, testContent) {
						t.Errorf("Download() content = %v, want %v", string(content), string(testContent))
					}
				}
			}

			// Test Delete
			err = tt.provider.Delete(ctx, "test/file.txt")
			if tt.wantErr {
				if err == nil {
					t.Error("Delete() error = nil, want error")
				}
			} else if err != nil {
				t.Errorf("Delete() error = %v", err)
			}

			// Test List
			files, err := tt.provider.List(ctx, "test/")
			if tt.wantErr {
				if err == nil {
					t.Error("List() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("List() error = %v", err)
				}
				if len(files) == 0 {
					t.Error("List() returned no files")
				}
				for _, file := range files {
					if file.Path == "" {
						t.Error("List() returned file with empty path")
					}
					if file.ContentType == "" {
						t.Error("List() returned file with empty content type")
					}
				}
			}
		})
	}
}

func TestStorageProviderEdgeCases(t *testing.T) {
	ctx := context.Background()
	provider := &mockStorageProvider{}

	t.Run("upload with nil input", func(t *testing.T) {
		_, err := provider.Upload(ctx, nil)
		if err == nil {
			t.Error("expected error for nil input")
		}
	})

	t.Run("upload with nil file", func(t *testing.T) {
		_, err := provider.Upload(ctx, &models.UploadInput{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Key:         "test/file.txt",
		})
		if err == nil {
			t.Error("expected error for nil file")
		}
	})

	t.Run("download empty path", func(t *testing.T) {
		_, err := provider.Download(ctx, "")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("delete empty path", func(t *testing.T) {
		err := provider.Delete(ctx, "")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("list with invalid prefix", func(t *testing.T) {
		_, err := provider.List(ctx, "../invalid")
		if err == nil {
			t.Error("expected error for invalid prefix")
		}
	})
}
