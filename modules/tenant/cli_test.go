package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volvlabs/nebularcore/core/config"
)

func TestConnectionString_TargetsConfiguredDatabaseWithSearchPath(t *testing.T) {
	dbCfg := config.DatabaseConfig{
		Username: "mori",
		Password: "secret",
		Host:     "localhost",
		Port:     "5432",
		Name:     "mori",
		SSLMode:  "disable",
	}

	got := connectionString(dbCfg, "farm_acme")

	// The schema must never end up in the dbname position — schemas live
	// inside dbCfg.Name's single database (D3a), they are not separate
	// databases. It must appear only as search_path.
	assert.Equal(t, "postgres://mori:secret@localhost:5432/mori?sslmode=disable&search_path=farm_acme", got)
	assert.Contains(t, got, "/mori?", "must connect to the configured database, not one named after the schema")
	assert.Contains(t, got, "search_path=farm_acme")
}
