package core

import (
	"fmt"

	"github.com/volvlabs/nebularcore/models/config"
	"github.com/volvlabs/nebularcore/tools/types"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func connectPostgresDB(config config.DatabaseConfig) (*gorm.DB, error) {
	if config.Type == types.GoogleCloudsqlPostgres {
		return gorm.Open(postgres.New(postgres.Config{
			DriverName: "cloudsqlpostgres",
			DSN: fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s TimeZone=UTC",
				config.Host, config.Username, config.Name, config.Password, config.SSLMode),
		}))
	}

	return gorm.Open(
		postgres.Open(fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			config.Host, config.Username, config.Password, config.Name, config.Port, config.SSLMode)),
		&gorm.Config{
			SkipDefaultTransaction: true,
		},
	)
}

func connectSqliteDB(dbPath string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		SkipDefaultTransaction: true,
	})
}
