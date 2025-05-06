package utils

import (
	"os"
	"path/filepath"
)

// GetProjectRoot returns the absolute path to the project root directory.
// It traverses up the directory tree until it finds a directory containing
// either a go.mod file (for Go projects) or a .git directory (for Git repositories).
func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}

		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			return currentDir, nil
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", os.ErrNotExist
		}
		currentDir = parentDir
	}
}
