package migration_runner_test

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/jideobs/nebularcore/core/migration_runner"
)

//go:embed testdata/migrations/*.sql
var testMigrations embed.FS

func TestNew(t *testing.T) {
	// Create a temporary directory for test migrations
	tmpDir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test migration files
	migrationContent := []byte("-- Test migration")
	err = os.WriteFile(filepath.Join(tmpDir, "001_test.up.sql"), migrationContent, 0644)
	require.NoError(t, err)

	tests := []struct {
		name             string
		sources          []migration_runner.Source
		connectionString string
		tableName        string
		expectError      bool
	}{
		{
			name:             "empty sources",
			sources:          []migration_runner.Source{},
			connectionString: "postgres://mock:mock@localhost:5432/mockdb?sslmode=disable",
			tableName:        "migrations",
			expectError:      true,
		},
		{
			name: "invalid source path",
			sources: []migration_runner.Source{
				{
					Path:     "/invalid/path",
					Priority: 1,
					Exclude:  []string{},
				},
			},
			connectionString: "postgres://mock:mock@localhost:5432/mockdb?sslmode=disable",
			tableName:        "migrations",
			expectError:      true,
		},
		{
			name: "invalid connection string",
			sources: []migration_runner.Source{
				{
					Path:     tmpDir,
					Priority: 1,
					Exclude:  []string{},
				},
			},
			connectionString: "invalid://connection:string",
			tableName:        "migrations",
			expectError:      true,
		},
		{
			name: "embedded filesystem source",
			sources: []migration_runner.Source{
				{
					Path:     "testdata/migrations",
					Priority: 1,
					Exclude:  []string{},
					FS:       testMigrations,
				},
			},
			connectionString: "postgres://mock:mock@localhost:5432/mockdb?sslmode=disable",
			tableName:        "migrations",
			expectError:      true, // This should fail since we're not actually connecting to a DB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := migration_runner.New(tt.sources, tt.connectionString, tt.tableName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, runner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, runner)
			}
		})
	}
}

func TestRunner_Integration(t *testing.T) {
	// Skip if not integration test environment
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test")
	}

	// Test cases for both file system and embedded migrations
	tests := []struct {
		name    string
		sources []migration_runner.Source
	}{
		{
			name: "file system migrations",
			sources: func() []migration_runner.Source {
				// Create a temporary directory for test migrations
				tmpDir, err := os.MkdirTemp("", "migrations")
				require.NoError(t, err)
				t.Cleanup(func() { os.RemoveAll(tmpDir) })

				// Create test migration files
				upContent := []byte(`
					CREATE TABLE test_table (
						id SERIAL PRIMARY KEY,
						name TEXT
					);
				`)
				downContent := []byte(`DROP TABLE IF EXISTS test_table;`)

				err = os.WriteFile(filepath.Join(tmpDir, "001_test.up.sql"), upContent, 0644)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(tmpDir, "001_test.down.sql"), downContent, 0644)
				require.NoError(t, err)

				return []migration_runner.Source{
					{
						Path:     tmpDir,
						Priority: 1,
						Exclude:  []string{},
					},
				}
			}(),
		},
		{
			name: "embedded migrations",
			sources: []migration_runner.Source{
				{
					Path:     "testdata/migrations",
					Priority: 1,
					Exclude:  []string{},
					FS:       testMigrations,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use test database connection string from environment
			connStr := os.Getenv("TEST_DB_CONNECTION")
			if connStr == "" {
				connStr = "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
			}

			runner, err := migration_runner.New(tt.sources, connStr, "migrations")
			require.NoError(t, err)
			defer runner.Close()

			// Test Up
			err = runner.Up()
			assert.NoError(t, err)

			// Test Down
			err = runner.Down(0)
			assert.NoError(t, err)

			// Test Steps
			err = runner.Steps(1)
			assert.NoError(t, err)

			err = runner.Steps(-1)
			assert.NoError(t, err)
		})
	}
}
