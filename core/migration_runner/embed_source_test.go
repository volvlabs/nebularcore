package migration_runner

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/migrations/*.sql
var theTestMigrations embed.FS

func TestEmbedSource_First(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	version, err := source.First()
	assert.NoError(t, err)
	assert.Equal(t, uint(1), version)
}

func TestEmbedSource_Next(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	version, err := source.Next(1)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), version)

	_, err = source.Next(2)
	assert.Error(t, err)
}

func TestEmbedSource_Prev(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	version, err := source.Prev(2)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), version)

	_, err = source.Prev(1)
	assert.Error(t, err)
}

func TestEmbedSource_ReadUp(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	reader, identifier, err := source.ReadUp(1)
	assert.NoError(t, err)
	assert.Equal(t, "000001_test.up.sql", identifier)
	assert.NotNil(t, reader)
	reader.Close()

	_, _, err = source.ReadUp(999)
	assert.Error(t, err)
}

func TestEmbedSource_ReadDown(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	reader, identifier, err := source.ReadDown(1)
	assert.NoError(t, err)
	assert.Equal(t, "000001_test.down.sql", identifier)
	assert.NotNil(t, reader)
	reader.Close()

	_, _, err = source.ReadDown(999)
	assert.Error(t, err)
}

func TestEmbedSource_Open(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	driver, err := source.Open("test-url")
	assert.NoError(t, err)
	assert.Equal(t, source, driver)
}

func TestEmbedSource_Close(t *testing.T) {
	source := NewEmbedSource(theTestMigrations, "testdata/migrations")

	err := source.Close()
	assert.NoError(t, err)
}
