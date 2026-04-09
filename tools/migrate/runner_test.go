package migrate_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/migrate"
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
	dbPath := filepath.Join(os.TempDir(), "/data.db")

	os.Remove(dbPath)

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
	if !tableExists(db, "users") {
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

func TestGoTo(t *testing.T) {
	// Arrange:
	rootDir := filesystem.GetRootDir("../../")
	dbPath := filepath.Join(os.TempDir(), "goto_data.db")

	os.Remove(dbPath)

	runner, err := migrate.NewRunner(
		fmt.Sprintf("file://%s", filepath.Join(rootDir, "test/data/migrations")),
		fmt.Sprintf("sqlite://%s", dbPath))
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	defer runner.Close()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	// Go to version 2: admins (v1) and users (v2) should exist; auths (v3) should not.
	if err := runner.Run("goto", "2"); err != nil {
		t.Fatalf("Run('goto', '2') error = %v", err)
	}
	if !tableExists(db, "admins") {
		t.Error("expected admins table to exist after goto 2")
	}
	if !tableExists(db, "users") {
		t.Error("expected users table to exist after goto 2")
	}
	if tableExists(db, "auths") {
		t.Error("expected auths table to not exist after goto 2")
	}

	// Go to version 1: only admins should exist.
	if err := runner.Run("goto", "1"); err != nil {
		t.Fatalf("Run('goto', '1') error = %v", err)
	}
	if !tableExists(db, "admins") {
		t.Error("expected admins table to exist after goto 1")
	}
	if tableExists(db, "users") {
		t.Error("expected users table to not exist after goto 1")
	}

	// Go to version 3: all tables should exist.
	if err := runner.Run("goto", "3"); err != nil {
		t.Fatalf("Run('goto', '3') error = %v", err)
	}
	if !tableExists(db, "admins") || !tableExists(db, "users") || !tableExists(db, "auths") {
		t.Error("expected all tables to exist after goto 3")
	}
}

func TestRunErrors(t *testing.T) {
	// Arrange: a valid runner is needed even for arg-validation errors.
	tempDir := os.TempDir()
	runner, err := migrate.NewRunner(
		fmt.Sprintf("file://%s", tempDir),
		fmt.Sprintf("sqlite://%s", filepath.Join(tempDir, "errors_test.db")))
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	defer runner.Close()

	scenarios := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "goto missing version",
			args:    []string{"goto"},
			wantErr: "goto requires a target version number",
		},
		{
			name:    "goto invalid version string",
			args:    []string{"goto", "abc"},
			wantErr: "invalid version",
		},
		{
			name:    "unknown command",
			args:    []string{"migrate"},
			wantErr: "command not supported",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := runner.Run(scenario.args...)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", scenario.wantErr)
			}
			if !strings.Contains(err.Error(), scenario.wantErr) {
				t.Errorf("expected error containing %q, got %q", scenario.wantErr, err.Error())
			}
		})
	}
}

func tableExists(db *sql.DB, tableName string) bool {
	var name string
	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	err := db.QueryRow(query).Scan(&name)
	return err == nil && name == tableName
}
