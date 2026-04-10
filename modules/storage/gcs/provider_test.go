package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/volvlabs/nebularcore/modules/storage/models"
)

// mockStorage implements a simple mock for GCS storage operations
type mockStorage struct {
	objects map[string][]byte
	attrs   map[string]*storage.ObjectAttrs
	errors  map[string]error
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		objects: make(map[string][]byte),
		attrs:   make(map[string]*storage.ObjectAttrs),
		errors:  make(map[string]error),
	}
}

func (m *mockStorage) writeObject(name string, data []byte, attrs *storage.ObjectAttrs) error {
	if err := m.errors["write"+name]; err != nil {
		return err
	}
	m.objects[name] = data
	m.attrs[name] = attrs
	return nil
}

func (m *mockStorage) readObject(name string) ([]byte, *storage.ObjectAttrs, error) {
	if err := m.errors["read"+name]; err != nil {
		return nil, nil, err
	}
	data, ok := m.objects[name]
	if !ok {
		return nil, nil, fmt.Errorf("object not found: %s", name)
	}
	return data, m.attrs[name], nil
}

func (m *mockStorage) deleteObject(name string) error {
	if err := m.errors["delete"+name]; err != nil {
		return err
	}
	if _, ok := m.objects[name]; !ok {
		return fmt.Errorf("object not found: %s", name)
	}
	delete(m.objects, name)
	delete(m.attrs, name)
	return nil
}

func (m *mockStorage) listObjects(prefix string) ([]*storage.ObjectAttrs, error) {
	if err := m.errors["list"+prefix]; err != nil {
		return nil, err
	}
	var result []*storage.ObjectAttrs
	for name, attrs := range m.attrs {
		if strings.HasPrefix(name, prefix) {
			result = append(result, attrs)
		}
	}
	return result, nil
}

func TestProvider_Upload(t *testing.T) {
	mock := newMockStorage()
	provider := &Provider{
		client: mock,
		bucket: "test-bucket",
	}

	tests := []struct {
		name      string
		input     *models.UploadInput
		wantError bool
	}{
		{
			name: "valid upload",
			input: &models.UploadInput{
				File:        bytes.NewReader([]byte("test content")),
				FileName:    "test.txt",
				ContentType: "text/plain",
				Key:         "test/file.txt",
				Metadata:    map[string]string{"test": "value"},
			},
			wantError: false,
		},
		{
			name: "missing key",
			input: &models.UploadInput{
				File:        bytes.NewReader([]byte("test content")),
				FileName:    "test.txt",
				ContentType: "text/plain",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := provider.Upload(context.Background(), tt.input)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expectedURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", provider.bucket, tt.input.Key)
			if output.URL != expectedURL {
				t.Errorf("URL = %v, want %v", output.URL, expectedURL)
			}
			if output.Path != tt.input.Key {
				t.Errorf("Path = %v, want %v", output.Path, tt.input.Key)
			}
		})
	}
}

func TestProvider_Download(t *testing.T) {
	testContent := []byte("test content")
	mock := newMockStorage()
	mock.writeObject("test/file.txt", testContent, &storage.ObjectAttrs{
		Name:        "test/file.txt",
		ContentType: "text/plain",
	})

	provider := &Provider{
		client: mock,
		bucket: "test-bucket",
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "existing file",
			path:      "test/file.txt",
			wantError: false,
		},
		{
			name:      "non-existent file",
			path:      "nonexistent.txt",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := provider.Download(context.Background(), tt.path)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			defer reader.Close()

			content, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("failed to read content: %v", err)
				return
			}
			if !bytes.Equal(content, testContent) {
				t.Errorf("content = %v, want %v", string(content), string(testContent))
			}
		})
	}
}

func TestProvider_Delete(t *testing.T) {
	mock := newMockStorage()
	// Add test file
	mock.writeObject("test/file.txt", []byte("test content"), &storage.ObjectAttrs{
		Name:        "test/file.txt",
		ContentType: "text/plain",
	})

	provider := &Provider{
		client: mock,
		bucket: "test-bucket",
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "existing file",
			path:      "test/file.txt",
			wantError: false,
		},
		{
			name:      "non-existent file",
			path:      "nonexistent.txt",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Delete(context.Background(), tt.path)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestProvider_List(t *testing.T) {
	now := time.Now()
	mock := newMockStorage()

	// Add test files
	mock.writeObject("test/file1.txt", []byte("content1"), &storage.ObjectAttrs{
		Name:        "test/file1.txt",
		Size:        100,
		Updated:     now,
		ContentType: "text/plain",
	})
	mock.writeObject("test/folder/", []byte{}, &storage.ObjectAttrs{
		Name:        "test/folder/",
		Size:        0,
		Updated:     now,
		ContentType: "application/x-directory",
	})

	provider := &Provider{
		client: mock,
		bucket: "test-bucket",
	}

	tests := []struct {
		name      string
		prefix    string
		wantCount int
	}{
		{
			name:      "list all",
			prefix:    "",
			wantCount: 2,
		},
		{
			name:      "list with prefix",
			prefix:    "test/",
			wantCount: 2,
		},
		{
			name:      "list non-existent prefix",
			prefix:    "nonexistent/",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := provider.List(context.Background(), tt.prefix)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(files) != tt.wantCount {
				t.Errorf("got %d files, want %d", len(files), tt.wantCount)
			}

			for _, file := range files {
				if tt.prefix != "" && !strings.HasPrefix(file.Path, tt.prefix) {
					t.Errorf("file path %s doesn't have prefix %s", file.Path, tt.prefix)
				}
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "jpeg file",
			filename: "test.jpg",
			want:     "image/jpeg",
		},
		{
			name:     "png file",
			filename: "test.png",
			want:     "image/png",
		},
		{
			name:     "gif file",
			filename: "test.gif",
			want:     "image/gif",
		},
		{
			name:     "pdf file",
			filename: "test.pdf",
			want:     "application/pdf",
		},
		{
			name:     "unknown extension",
			filename: "test.xyz",
			want:     "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getContentType(tt.filename); got != tt.want {
				t.Errorf("getContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
