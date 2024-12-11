package daos

import (
	"fmt"

	"gitlab.com/jideobs/nebularcore/tools/migrate"
	"gorm.io/gorm"
)

func (d *Dao) WithTenant(schemaName string) (*gorm.DB, error) {
	dbSession := d.DB().Session(&gorm.Session{NewDB: true})
	err := dbSession.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
	if err != nil {
		return nil, err
	}

	return dbSession, nil
}

func (d *Dao) Create(schemaName string) error {
	return d.DB().Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
}

func (d *Dao) Drop(schemaName string) error {
	return d.DB().Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error
}

func (d *Dao) Migrate(schemaName string) error {
	dbConfig := d.databaseConfig
	runner, err := migrate.NewRunner(
		fmt.Sprintf("file:///%s/%s", d.tenantConfig.BaseDir, d.tenantConfig.MigrationPath),
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.Name,
			dbConfig.SSLMode,
			schemaName,
		))
	if err != nil {
		return err
	}
	return runner.Run("up")
}

func (d *Dao) AutoMigrate() error {
	const batchSize = 1000
	var offset int

	for {
		var schemaNames []string
		err := d.DB().Transaction(func(tx *gorm.DB) error {
			return tx.Table(d.tenantConfig.TenantsTableName).
				Select(d.tenantConfig.TenantSchemaColName).
				Offset(offset).
				Limit(batchSize).
				Pluck(d.tenantConfig.TenantSchemaColName, &schemaNames).Error
		})
		if err != nil {
			return err
		}

		if len(schemaNames) == 0 {
			break
		}

		for _, schemaName := range schemaNames {
			if err := d.Migrate(schemaName); err != nil {
				return err
			}
		}

		offset += batchSize
	}

	return nil
}
