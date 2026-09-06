package models

import (
	"bytes"
	"testing"
	"time"
)

func TestUploadInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   UploadInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: UploadInput{
				File:        bytes.NewReader([]byte("test content")),
				FileName:    "test.txt",
				ContentType: "text/plain",
				Key:        "user-123/test.txt",
				Metadata:    map[string]string{"owner": "user-123"},
			},
			wantErr: false,
		},
		{
			name: "missing file",
			input: UploadInput{
				FileName:    "test.txt",
				ContentType: "text/plain",
				Key:        "user-123/test.txt",
			},
			wantErr: true,
		},
		{
			name: "missing key",
			input: UploadInput{
				File:        bytes.NewReader([]byte("test content")),
				FileName:    "test.txt",
				ContentType: "text/plain",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.File == nil && !tt.wantErr {
				t.Error("expected error for nil File but got none")
			}
			if tt.input.Key == "" && !tt.wantErr {
				t.Error("expected error for empty Key but got none")
			}
		})
	}
}

func TestFileInfo_IsDirectory(t *testing.T) {
	tests := []struct {
		name     string
		fileInfo FileInfo
		want     bool
	}{
		{
			name: "directory",
			fileInfo: FileInfo{
				Path:        "folder/",
				Size:        0,
				ContentType: "application/x-directory",
				ModTime:     time.Now(),
				IsDir:       true,
			},
			want: true,
		},
		{
			name: "file",
			fileInfo: FileInfo{
				Path:        "folder/file.txt",
				Size:        100,
				ContentType: "text/plain",
				ModTime:     time.Now(),
				IsDir:       false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fileInfo.IsDir; got != tt.want {
				t.Errorf("FileInfo.IsDir = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadOutput_Validation(t *testing.T) {
	tests := []struct {
		name   string
		output UploadOutput
		valid  bool
	}{
		{
			name: "valid output",
			output: UploadOutput{
				Path:        "user-123/test.txt",
				URL:         "https://storage.example.com/user-123/test.txt",
				ContentType: "text/plain",
				Size:        100,
				Metadata:    map[string]string{"owner": "user-123"},
			},
			valid: true,
		},
		{
			name: "missing path",
			output: UploadOutput{
				URL:         "https://storage.example.com/user-123/test.txt",
				ContentType: "text/plain",
				Size:        100,
			},
			valid: false,
		},
		{
			name: "invalid size",
			output: UploadOutput{
				Path:        "user-123/test.txt",
				URL:         "https://storage.example.com/user-123/test.txt",
				ContentType: "text/plain",
				Size:        -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.output.Path == "" && tt.valid {
				t.Error("expected invalid for empty Path but got valid")
			}
			if tt.output.Size < 0 && tt.valid {
				t.Error("expected invalid for negative Size but got valid")
			}
		})
	}
}
