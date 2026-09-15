package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/modules/storage/local"
)

// mockLocalProvider implements a mock for the local.Provider
type mockLocalProvider struct {
	serveFileFunc func(w http.ResponseWriter, r *http.Request, path string)
}

func (m *mockLocalProvider) ServeFile(w http.ResponseWriter, r *http.Request, path string) {
	if m.serveFileFunc != nil {
		m.serveFileFunc(w, r, path)
	}
}

func TestHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		prefix       string
		path         string
		expectedPath string
		expectedCode int
		mockResponse string
		mockProvider *mockLocalProvider
	}{
		{
			name:         "serve file without prefix",
			prefix:       "",
			path:         "/test.txt",
			expectedPath: "test.txt",
			expectedCode: http.StatusOK,
			mockResponse: "test content",
			mockProvider: &mockLocalProvider{
				serveFileFunc: func(w http.ResponseWriter, r *http.Request, path string) {
					if path != "test.txt" {
						t.Errorf("expected path %s, got %s", "test.txt", path)
					}
					w.WriteHeader(http.StatusOK)
					if _, err := io.WriteString(w, "test content"); err != nil {
						t.Errorf("failed to write response: %v", err)
					}
				},
			},
		},
		{
			name:         "serve file with prefix",
			prefix:       "/files",
			path:         "/files/test.txt",
			expectedPath: "test.txt",
			expectedCode: http.StatusOK,
			mockResponse: "test content",
			mockProvider: &mockLocalProvider{
				serveFileFunc: func(w http.ResponseWriter, r *http.Request, path string) {
					if path != "test.txt" {
						t.Errorf("expected path %s, got %s", "test.txt", path)
					}
					w.WriteHeader(http.StatusOK)
					if _, err := io.WriteString(w, "test content"); err != nil {
						t.Errorf("failed to write response: %v", err)
					}
				},
			},
		},
		{
			name:         "root path returns 404",
			prefix:       "",
			path:         "/",
			expectedCode: http.StatusNotFound,
			mockProvider: &mockLocalProvider{},
		},
		{
			name:         "path traversal attempt",
			prefix:       "",
			path:         "/../secret.txt",
			expectedPath: "secret.txt",
			expectedCode: http.StatusOK,
			mockResponse: "test content",
			mockProvider: &mockLocalProvider{
				serveFileFunc: func(w http.ResponseWriter, r *http.Request, path string) {
					if path != "secret.txt" {
						t.Errorf("expected path %s, got %s", "secret.txt", path)
					}
					w.WriteHeader(http.StatusOK)
					if _, err := io.WriteString(w, "test content"); err != nil {
						t.Errorf("failed to write response: %v", err)
					}
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()

			// Create and register the handler
			handler := NewHandler(tt.mockProvider, tt.prefix)
			handler.Register(router)

			// Create a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			// Check response body if expected
			if tt.mockResponse != "" && w.Body.String() != tt.mockResponse {
				t.Errorf("expected response body %s, got %s", tt.mockResponse, w.Body.String())
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name       string
		provider   *local.Provider
		prefix     string
		wantPrefix string
	}{
		{
			name:       "empty prefix",
			provider:   &local.Provider{},
			prefix:     "",
			wantPrefix: "",
		},
		{
			name:       "prefix with trailing slash",
			provider:   &local.Provider{},
			prefix:     "/files/",
			wantPrefix: "/files",
		},
		{
			name:       "prefix without trailing slash",
			provider:   &local.Provider{},
			prefix:     "/files",
			wantPrefix: "/files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.provider, tt.prefix)
			if handler.prefix != tt.wantPrefix {
				t.Errorf("NewHandler() prefix = %v, want %v", handler.prefix, tt.wantPrefix)
			}
			if handler.provider != tt.provider {
				t.Error("NewHandler() provider not set correctly")
			}
		})
	}
}
