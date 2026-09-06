package tenant

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
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
	err := db.Model(&Tenant{}).Select("distinct schema_name").Scan(&tenantSchemas).Error
	return tenantSchemas, err
}

// connectionString builds a DSN that connects to the configured database and
// scopes the session to the given schema via search_path. Schemas in this
// framework are not separate databases — they're schemas *within* dbCfg's
// single database (see D3a in the mori platform plan) — so the schema name
// must never be substituted into the DSN's dbname position.
func connectionString(dbCfg config.DatabaseConfig, schema string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
		dbCfg.Username, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode, schema)
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
			tenantModules := app.GetModulesInOrder(module.TenantNamespace)
			switch cmd {
			case "tenant":
				if len(args) < 2 {
					return fmt.Errorf("schema name is required")
				}
				return migrateSchema(dbCfg, tenantModules, projectRoot, args[1])
			case "all-tenants", "":
				return MigrateAllSchemas(app.DB(), dbCfg, tenantModules, projectRoot)
			default:
				return fmt.Errorf("unknown command %s", cmd)
			}
		},
	}

	return command
}

// migrateSchema runs every TenantNamespace module's migrations against one
// schema, in dependency order (GetModulesInOrder — not
// GetModulesByNamespace, which returns an unordered map and would let a
// module's migration run before one it depends on, e.g. a module whose
// tables FK-reference another tenant module's tables). Takes the ordered
// module list directly rather than an app.App[T] — generic callers like
// ProvisionSchema (provision.go) may have no app.App[T] handle at all, only
// a db connection and the module list resolved once at startup.
func migrateSchema(dbCfg config.DatabaseConfig, tenantModules []module.OrderedModule, projectRoot, schema string) error {
	dbString := connectionString(dbCfg, schema)
	for _, om := range tenantModules {
		if !om.Module.ProvidesMigrations() {
			continue
		}
		if err := runTenantMigrations(om.Module, projectRoot, om.Name, dbString); err != nil {
			return fmt.Errorf("running migrations for module %s against schema %s: %w", om.Name, schema, err)
		}
	}
	return nil
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

	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	log.Info().Msgf("migrations for module %s ran successfully", name)
	return nil
}
