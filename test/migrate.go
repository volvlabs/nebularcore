package test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/volvlabs/nebularcore/tools/filesystem"
	"github.com/volvlabs/nebularcore/tools/migrate"
)

func RunMigrationWithCustomPaths(t *testing.T, migrationDir, dataDir string) func(*testing.T) {
	migrationDir = filesystem.EncodeFilePathAsFileURL(migrationDir)
	connectionString := fmt.Sprintf("sqlite://%s", filepath.Join(dataDir, "data.db"))
	runner, err := migrate.NewRunner(migrationDir, connectionString)
	if err != nil {
		t.Fatalf("Error creating migration runner: %v", err)
	}

	err = runner.Run("up")
	if err != nil {
		t.Fatalf("Error running 'up' migration: %v", err)
	}

	return func(t *testing.T) {
		err := runner.Run("down")
		if err != nil {
			t.Fatalf("Error running 'down' migration: %v", err)
		}
		runner.Close()
	}
}

func RunMigration(t *testing.T, baseDir, dataDir string) func(*testing.T) {
	migrationDir := filepath.Join(baseDir, "test/data/migrations")
	return RunMigrationWithCustomPaths(t, migrationDir, dataDir)
}
