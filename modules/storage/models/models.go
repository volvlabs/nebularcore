package models

import (
	"io"
	"time"
)

// UploadInput contains all necessary information for file upload
type UploadInput struct {
	File        io.Reader
	FileName    string
	ContentType string
	// For idempotency and updates
	Key      string // Unique identifier for the file (e.g., "user-123/avatar")
	Metadata map[string]string
}

// UploadOutput contains the result of a successful upload
type UploadOutput struct {
	Path        string
	URL         string // Public URL if available
	ContentType string
	Size        int64
	Metadata    map[string]string
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	Path        string
	Size        int64
	ContentType string
	ModTime     time.Time
	IsDir       bool
}
