package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddedMigrations(t *testing.T) {
	files, err := migrations.ReadDir("migrations")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	// Check for specific migration files
	expectedFiles := []string{
		"000001_init_auth.up.sql",
		"000001_init_auth.down.sql",
		"000002_roles.up.sql",
		"000002_roles.down.sql",
		"000003_permissions.up.sql",
		"000003_permissions.down.sql",
		"000004_groups.up.sql",
		"000004_groups.down.sql",
	}

	foundFiles := make([]string, 0)
	for _, file := range files {
		foundFiles = append(foundFiles, file.Name())
	}

	for _, expected := range expectedFiles {
		assert.Contains(t, foundFiles, expected)
	}
}
