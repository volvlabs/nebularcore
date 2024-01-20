package core

import (
	"fmt"

	"gitlab.com/jideobs/nebularcore/models/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func connectPostgresDB(config config.DatabaseConfig) (*gorm.DB, error) {
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
