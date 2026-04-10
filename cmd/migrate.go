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
	"github.com/volvlabs/nebularcore/core/config"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
)

func NewMigrateCommand[T config.Settings](
	app core.App[T],
	dbCfg config.DatabaseConfig,
) *cobra.Command {
	const cmdDesc = `Supported commands are:
- create [module] [name] - creates new blank migration file
- up - runs all pending migrations`
	command := &cobra.Command{
		Use:       "migrate",
		Short:     "All DB migration commands",
		Long:      cmdDesc,
		ValidArgs: []string{"create", "up"},
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
	defer func() { _ = upFile.Close() }()

	downFile, err := os.Create(downFileName)
	if err != nil {
		return err
	}
	defer func() { _ = downFile.Close() }()

	log.Info().Msgf("Created %s and %s\n", upFileName, downFileName)
	return nil
}
