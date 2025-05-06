package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/core/config"
	migrationRunner "gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gitlab.com/jideobs/nebularcore/core/module"
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
				for name, module := range app.GetModulesByNamespace(module.PublicNamespace) {
					if !module.ProvidesMigrations() {
						continue
					}

					projectRoot := app.Config().ProjectRoot
					sources := module.GetMigrationSources(projectRoot)
					runner, err := migrationRunner.New(
						sources,
						dbString,
						fmt.Sprintf("schema_migrations_%s", name),
					)
					if err != nil {
						log.Err(err).Msg("error creating migration runner")
						return err
					}
					if err := runner.Up(); err != nil {
						return err
					}

					log.Info().Msgf("migration for module %s ran successfully", name)
				}
			default:
				return fmt.Errorf("unknown command %s", cmd)
			}
			return nil
		},
	}

	return command
}

func createMigrationFileHandler(migrationsDir, name string) error {
	// Read all files in migrations directory
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Find highest version number
	highestVersion := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			// Extract version number from filename (e.g., "000001" from "000001_init.up.sql")
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

	// Generate new version number
	newVersion := highestVersion + 1
	versionPrefix := fmt.Sprintf("%06d", newVersion) // Format as 000001, 000002, etc.

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
