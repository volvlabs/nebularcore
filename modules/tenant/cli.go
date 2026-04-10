package tenant

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/core/config"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"gorm.io/gorm"
)

const cmdDesc = `Supported commands are:
- tenant [schema] [command] - runs migrations for a specific tenant schema
- all-tenants   - runs migrations for all tenant schemas`

func getAllSchemas(db *gorm.DB) ([]string, error) {
	tenantSchemas := []string{}
	err := db.Model(&Tenant{}).Select("distinct schema").Scan(&tenantSchemas).Error
	return tenantSchemas, err
}

func NewTenantMigrateCommand[T config.Settings](app core.App[T], dbCfg config.DatabaseConfig) *cobra.Command {
	command := &cobra.Command{
		Use:       "migrate_schemas",
		Short:     "execute app tenant db migration scripts",
		Long:      cmdDesc,
		ValidArgs: []string{"tenant", "all-tenants"},
		RunE: func(command *cobra.Command, args []string) error {
			cmd := ""
			if len(args) > 0 {
				cmd = args[0]
			}

			projectRoot := app.Config().ProjectRoot
			switch cmd {
			case "tenant":
				if len(args) < 2 {
					return fmt.Errorf("schema name is required")
				}

				schema := args[1]
				dbString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
					dbCfg.Username, dbCfg.Password, dbCfg.Host, dbCfg.Port, schema, dbCfg.SSLMode)
				for name, module := range app.GetModulesByNamespace(module.TenantNamespace) {
					if !module.ProvidesMigrations() {
						continue
					}

					if err := runTenantMigrations(module, projectRoot, name, dbString); err != nil {
						log.Err(err).Msgf("error occurred running migrations for %s", name)
					}
				}
			case "all-tenants":
			default:
				schemas, err := getAllSchemas(app.DB())
				if err != nil {
					return err
				}

				for _, schema := range schemas {
					dbString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
						dbCfg.Username, dbCfg.Password, dbCfg.Host, dbCfg.Port, schema, dbCfg.SSLMode)
					for name, module := range app.GetModulesByNamespace(module.TenantNamespace) {
						if !module.ProvidesMigrations() {
							continue
						}

						if err := runTenantMigrations(module, projectRoot, name, dbString); err != nil {
							log.Err(err).Msgf("error occurred running migrations for %s", name)
						}
					}
				}

				return nil
			}

			return nil
		},
	}

	return command
}

func runTenantMigrations(module module.Module, projectRoot, name, connectionString string) error {
	runner, err := migrationRunner.New(
		module.GetMigrationSources(projectRoot),
		connectionString,
		fmt.Sprintf("schema_migrations_%s", name),
	)
	if err != nil {
		return err
	}

	return runner.Up()
}
