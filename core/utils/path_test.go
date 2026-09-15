package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volvlabs/nebularcore/core/utils"
)

func TestGetProjectRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "project-root-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test cases
	tests := []struct {
		name           string
		setupFunc      func(string) string
		expectedError  bool
		expectedMarker string
	}{
		{
			name: "find project root with go.mod",
			setupFunc: func(baseDir string) string {
				// Create a nested directory structure with go.mod
				testDir := filepath.Join(baseDir, "project", "src", "pkg")
				if err := os.MkdirAll(testDir, 0755); err != nil {
					panic(err)
				}
				goModPath := filepath.Join(baseDir, "project", "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
					panic(err)
				}
				return testDir
			},
			expectedMarker: "go.mod",
		},
		{
			name: "find project root with .git",
			setupFunc: func(baseDir string) string {
				// Create a nested directory structure with .git
				testDir := filepath.Join(baseDir, "project", "src", "pkg")
				if err := os.MkdirAll(testDir, 0755); err != nil {
					panic(err)
				}
				gitDir := filepath.Join(baseDir, "project", ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					panic(err)
				}
				return testDir
			},
			expectedMarker: ".git",
		},
		{
			name: "no project root found",
			setupFunc: func(baseDir string) string {
				// Create a nested directory without any project markers
				testDir := filepath.Join(baseDir, "no-project", "src", "pkg")
				if err := os.MkdirAll(testDir, 0755); err != nil {
					panic(err)
				}
				return testDir
			},
			expectedError: true,
		},
	}

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := tt.setupFunc(tmpDir)

			// Change to test directory
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			// Run test
			root, err := utils.GetProjectRoot()

			// Restore original working directory
			if err := os.Chdir(originalWd); err != nil {
				t.Fatal(err)
			}

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, "", root)
			} else {
				assert.NoError(t, err)
				// Verify the marker file/directory exists in the returned root
				markerPath := filepath.Join(root, tt.expectedMarker)
				_, err := os.Stat(markerPath)
				assert.NoError(t, err)
			}
		})
	}
}
