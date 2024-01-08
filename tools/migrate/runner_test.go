package migrate_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/volvlabs/nebularcore/tools/filesystem"
	"gitlab.com/volvlabs/nebularcore/tools/migrate"
)

func TestNewRunner(t *testing.T) {
	// Arrange:
	tempDir := os.TempDir()
	scenarios := []struct {
		name             string
		migrationDir     string
		connectionString string
		wantErr          bool
	}{
		{
			name:             "should successfully return a runner",
			migrationDir:     fmt.Sprintf("file://%s", tempDir),
			connectionString: fmt.Sprintf("sqlite://%s", filepath.Join(tempDir, "test.db")),
			wantErr:          false,
		},
		{
			name:             "should return an error because of bad connection string",
			migrationDir:     fmt.Sprintf("file://%s", tempDir),
			connectionString: "bad connection string",
			wantErr:          true,
		},
	}

	for _, scenario := range scenarios {
		// Act:
		runner, err := migrate.NewRunner(
			scenario.migrationDir,
			scenario.connectionString)

		// Assert:
		if (err != nil) != scenario.wantErr {
			t.Errorf("NewRunner() error = %v, wantErr %v", err, scenario.wantErr)
			return
		}
		if !scenario.wantErr && runner == nil {
			t.Errorf("expected runner, got nil")
		}
	}
}

func TestUpAndDown(t *testing.T) {
	// Arrange:
	rootDir := filesystem.GetRootDir("../../")
	dbPath := filepath.Join(rootDir, "test/data/data.db")
	runner, _ := migrate.NewRunner(
		fmt.Sprintf("file://%s", filepath.Join(rootDir, "test/data/migrations")),
		fmt.Sprintf("sqlite://%s", dbPath))
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	// Act-Assert:
	if tableExists(db, "users") || tableExists(db, "admins") {
		t.Fatalf("expected tables to not exist")
	}

	err = runner.Run("up")
	if err != nil {
		t.Fatalf("Run('up') error = %v", err)
	}

	if !tableExists(db, "users") || !tableExists(db, "admins") {
		t.Fatalf("expected tables exist")
	}

	err = runner.Run("down", "1")
	if err != nil {
		t.Fatalf("Run('down') error = %v", err)
	}

	if !tableExists(db, "admins") {
		t.Fatalf("expected admins table to not exist")
	}
	if tableExists(db, "users") {
		t.Fatalf("expected users table to exist")
	}

	err = runner.Run("down")
	if err != nil {
		t.Fatalf("Run('down') error = %v", err)
	}
	if tableExists(db, "admins") {
		t.Fatal("expected admins table to not exist")
	}
}

func tableExists(db *sql.DB, tableName string) bool {
	var name string
	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	err := db.QueryRow(query).Scan(&name)
	return err == nil && name == tableName
}
