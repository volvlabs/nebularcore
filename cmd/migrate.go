package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/models/config"
	"github.com/volvlabs/nebularcore/tools/migrate"
	migrationRunner "github.com/vovlabs/nebularcore/tools/migrate/runner"
)

func NewMigrateCommand[T config.Settings](
	app core.App[T],
	dbCfg config.DatabaseConfig,
) *cobra.Command {
	const cmdDesc = `Supported commands are:
- up			- runs migrations
- down [number]		- revert last [number] of migrations, or all if omitted
- goto <version>	- migrate to a specific version number
- create [name]		- creates new blank migration file
- tenant [schema] [command] [args] - runs migrations for a specific tenant schema
- all-tenants [command] [args]     - runs migrations for all tenant schemas`
	command := &cobra.Command{
		Use:       "migrate",
		Short:     "All DB migration commands",
		Long:      cmdDesc,
		ValidArgs: []string{"up", "down", "goto", "create", "tenant", "all-tenants"},
		RunE: func(command *cobra.Command, args []string) error {
			cmd := ""
			if len(args) > 0 {
				cmd = args[0]
			}
			switch cmd {
			case "create":
				if len(args) < 3 {
					return fmt.Errorf("module and migration name is required")
				}
				module, ok := app.GetModule(args[1])
				if !ok {
					return fmt.Errorf("module %s not found", args[1])
				}
				if !module.ProvidesMigrations() {
					return fmt.Errorf("module %s does not provide migrations", args[1])
				}

				directory := fmt.Sprintf("%s/modules/%s/%s",
					app.Config().ProjectRoot, module.Name(), module.MigrationsDir())
				if err := createMigrationFileHandler(directory, args[2]); err != nil {
					return err
				}
			case "up":
				dbString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
					dbCfg.Username, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode)
				for _, om := range app.GetModulesInOrder(module.PublicNamespace) {
					if !om.Module.ProvidesMigrations() {
						continue
					}

					projectRoot := app.Config().ProjectRoot
					sources := om.Module.GetMigrationSources(projectRoot)
					runner, err := migrationRunner.New(
						sources,
						dbString,
						fmt.Sprintf("schema_migrations_%s", om.Name),
					)
					if err != nil {
						log.Err(err).Msg("error creating migration runner")
						return err
					}
					if err := runner.Up(); err != nil {
						if err == migrate.ErrNoChange || err == migrate.ErrNilVersion {
							log.Err(err).Msgf("no migrations to run for module %s", om.Name)
							continue
						}
						return err
					}

					log.Info().Msgf("migration for module %s ran successfully", om.Name)
				}

				schemaName := args[1]

				// Create a new migration runner
				runner, err := migrate.NewRunner(
					fmt.Sprintf("file:///%s", app.MigrationsDir()),
					fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
						dbCfg.Username, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode))
				if err != nil {
					return err
				}

				// Create a schema-specific runner
				schemaRunner, err := runner.WithSchema(schemaName)
				if err != nil {
					return err
				}

				// Run the migration with the remaining args
				var migrationArgs []string
				if len(args) > 2 {
					migrationArgs = args[2:]
				} else {
					migrationArgs = []string{"up"}
				}

				return schemaRunner.Run(migrationArgs...)
			case "all-tenants":
				// Create a DAO to access the database
				dao := app.Dao()
				if dao == nil {
					return fmt.Errorf("database access is not available")
				}

				// Run migrations for all tenants, passing any sub-command and args
				return dao.AutoMigrateSchemas(args[1:]...)
			default:
				return fmt.Errorf("unknown command %s", cmd)
			}
			return nil
		},
	}

	return command
}

func createMigrationFileHandler(migrationsDir, name string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	highestVersion := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			parts := strings.Split(entry.Name(), "_")
			if len(parts) > 0 {
				if version, err := strconv.Atoi(parts[0]); err == nil {
					if version > highestVersion {
						highestVersion = version
					}
				}
			}
		}
	}

	newVersion := highestVersion + 1
	versionPrefix := fmt.Sprintf("%06d", newVersion)

	upFileName := fmt.Sprintf("%s/%s_%s.up.sql", migrationsDir, versionPrefix, name)
	downFileName := fmt.Sprintf("%s/%s_%s.down.sql", migrationsDir, versionPrefix, name)

	upFile, err := os.Create(upFileName)
	if err != nil {
		return err
	}
	defer upFile.Close()

	downFile, err := os.Create(downFileName)
	if err != nil {
		return err
	}
	defer downFile.Close()

	log.Info().Msgf("Created %s and %s\n", upFileName, downFileName)
	return nil
}
