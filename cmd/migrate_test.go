package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestCreateMigrationFileHandler(t *testing.T) {
	// Arrange:
	migrationsDir := os.TempDir()
	migrationName := "test_migration"

	timestamp := time.Now().Unix()
	upFileName := fmt.Sprintf("%s/%d_%s.up.sql", migrationsDir, timestamp, migrationName)
	downFileName := fmt.Sprintf("%s/%d_%s.down.sql", migrationsDir, timestamp, migrationName)
	buf := new(bytes.Buffer)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: buf})

	// Act:
	err := createMigrationFileHandler(migrationsDir, migrationName)

	// Assert:
	assert.NoError(t, err)
	assert.FileExists(t, upFileName)
	assert.FileExists(t, downFileName)
	assert.Contains(t, buf.String(), fmt.Sprintf("Created %s and %s\n", upFileName, downFileName))
}
