package local

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/jideobs/nebularcore/modules/storage/models"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestNew(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name      string
		basePath  string
		setup     func(t *testing.T) string
		cleanup   func(t *testing.T, path string)
		baseURL   string
		wantError bool
	}{
		{
			name:      "valid configuration",
			basePath:  tempDir,
			baseURL:   "http://example.com",
			wantError: false,
		},
		{
			name:      "creates non-existent directory in existing parent",
			basePath:  filepath.Join(tempDir, "new-directory"),
			baseURL:   "http://example.com",
			wantError: false,
		},
		{
			name: "fails when path exists but is not a directory",
			setup: func(t *testing.T) string {
				testFile := filepath.Join(tempDir, "testfile")
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
				return testFile
			},
			baseURL:   "http://example.com",
			wantError: true,
		},
		{
			name: "fails when directory is not writable",
			setup: func(t *testing.T) string {
				testDir := filepath.Join(tempDir, "readonly-dir")
				require.NoError(t, os.MkdirAll(testDir, 0755))
				err := os.Chmod(testDir, 0555)
				if err != nil {
					t.Skip("Skipping read-only directory test: ", err)
				}
				return testDir
			},
			baseURL:   "http://example.com",
			wantError: true,
			cleanup: func(t *testing.T, path string) {
				// Restore permissions for cleanup
				_ = os.Chmod(path, 0755)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath := tt.basePath
			if tt.setup != nil {
				basePath = tt.setup(t)
			}
			if tt.cleanup != nil {
				defer tt.cleanup(t, basePath)
			}

			provider, err := New(basePath, tt.baseURL)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if provider.basePath != tt.basePath {
				t.Errorf("basePath = %v, want %v", provider.basePath, tt.basePath)
			}
			if provider.baseURL != tt.baseURL {
				t.Errorf("baseURL = %v, want %v", provider.baseURL, tt.baseURL)
			}
		})
	}
}

func TestProvider_Upload(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	provider, err := New(tempDir, "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
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

			// Verify file exists
			filePath := filepath.Join(tempDir, tt.input.Key)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Error("file was not created")
			}

			// Verify metadata file exists
			metaPath := filePath + ".meta"
			if _, err := os.Stat(metaPath); os.IsNotExist(err) {
				t.Error("metadata file was not created")
			}

			// Verify output
			if output.Path != tt.input.Key {
				t.Errorf("Path = %v, want %v", output.Path, tt.input.Key)
			}
			if output.ContentType != tt.input.ContentType {
				t.Errorf("ContentType = %v, want %v", output.ContentType, tt.input.ContentType)
			}
		})
	}
}

func TestProvider_Download(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	provider, err := New(tempDir, "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create a test file
	testContent := []byte("test content")
	testPath := "test/download.txt"
	fullPath := filepath.Join(tempDir, testPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(fullPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "existing file",
			path:      testPath,
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
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	provider, err := New(tempDir, "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create a test file
	testPath := "test/delete.txt"
	fullPath := filepath.Join(tempDir, testPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(fullPath+".meta", []byte("test=value"), 0644); err != nil {
		t.Fatalf("Failed to create test metadata: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "existing file",
			path:      testPath,
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
				return
			}

			// Verify file is deleted
			if _, err := os.Stat(filepath.Join(tempDir, tt.path)); !os.IsNotExist(err) {
				t.Error("file was not deleted")
			}
			// Verify metadata is deleted
			if _, err := os.Stat(filepath.Join(tempDir, tt.path+".meta")); !os.IsNotExist(err) {
				t.Error("metadata file was not deleted")
			}
		})
	}
}

func TestProvider_List(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	provider, err := New(tempDir, "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create test directory structure
	files := map[string]string{
		"test/file1.txt":       "content1",
		"test/file2.txt":       "content2",
		"test/sub/file3.txt":   "content3",
		"other/file4.txt":      "content4",
		"test/file1.txt.meta":  "test=value",
		"test/file2.txt.meta":  "test=value",
		"test/sub/file3.meta":  "test=value",
		"other/file4.txt.meta": "test=value",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tests := []struct {
		name       string
		prefix     string
		wantCount  int
		wantPrefix string
	}{
		{
			name:       "list test directory",
			prefix:     "test",
			wantCount:  3,
			wantPrefix: "test",
		},
		{
			name:       "list subdirectory",
			prefix:     "test/sub",
			wantCount:  1,
			wantPrefix: "test/sub",
		},
		{
			name:       "list non-existent directory",
			prefix:     "nonexistent",
			wantCount:  0,
			wantPrefix: "nonexistent",
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
				if !strings.HasPrefix(file.Path, tt.wantPrefix) {
					t.Errorf("file path %s doesn't have prefix %s", file.Path, tt.wantPrefix)
				}
				if strings.HasSuffix(file.Path, ".meta") {
					t.Error("metadata file should not be included in listing")
				}
			}
		})
	}
}

func TestProvider_ServeFile(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	provider, err := New(tempDir, "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create a test file
	testContent := []byte("test content")
	testPath := "test/serve.txt"
	fullPath := filepath.Join(tempDir, testPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(fullPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "serve existing file",
			path:       testPath,
			wantStatus: http.StatusOK,
			wantBody:   string(testContent),
		},
		{
			name:       "serve non-existent file",
			path:       "nonexistent.txt",
			wantStatus: http.StatusNotFound,
			wantBody:   "",
		},
		{
			name:       "prevent directory traversal",
			path:       "../../../etc/passwd",
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/"+tt.path, nil)
			provider.ServeFile(w, r, tt.path)

			if w.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				if body := w.Body.String(); body != tt.wantBody {
					t.Errorf("body = %s, want %s", body, tt.wantBody)
				}
			}
		})
	}
}
