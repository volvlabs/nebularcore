package daos

import (
	"fmt"

	"gitlab.com/jideobs/nebularcore/tools/migrate"
	"gorm.io/gorm"
)

func (d *Dao) WithSchemaSession(schemaName string) (*gorm.DB, error) {
	dbSession := d.DB().Session(&gorm.Session{NewDB: true})
	err := dbSession.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
	if err != nil {
		return nil, err
	}

	return dbSession, nil
}

func (d *Dao) ResetSchema() error {
	return d.DB().Exec("SET search_path TO public").Error
}

func (d *Dao) CreateSchema(schemaName string) error {
	return d.DB().Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
}

func (d *Dao) DropSchema(schemaName string) error {
	return d.DB().Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error
}

func (d *Dao) MigrateSchema(schemaName string, args ...string) error {
	if len(args) == 0 {
		args = []string{"up"}
	}

	dbConfig := d.databaseConfig
	// Create a base runner
	runner, err := migrate.NewRunner(
		fmt.Sprintf("file:///%s/%s", d.tenantConfig.BaseDir, d.tenantConfig.MigrationPath),
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Port,
			dbConfig.Name,
			dbConfig.SSLMode,
		))
	if err != nil {
		return err
	}

	// Create a schema-specific runner
	schemaRunner, err := runner.WithSchema(schemaName)
	if err != nil {
		return err
	}

	// Run the migration
	return schemaRunner.Run(args...)
}

func (d *Dao) AutoMigrateSchemas(args ...string) error {
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
			if err := d.MigrateSchema(schemaName, args...); err != nil {
				return err
			}
		}

		offset += batchSize
	}

	return nil
}
