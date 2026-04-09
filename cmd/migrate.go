package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models/config"
	"gitlab.com/jideobs/nebularcore/tools/migrate"
)

func NewMigrateCommand(app core.App, dbCfg config.DatabaseConfig) *cobra.Command {
	const cmdDesc = `Supported commands are:
- up			- runs migrations
- down [number]		- revert last [number] of migrations, or all if omitted
- goto <version>	- migrate to a specific version number
- create [name]		- creates new blank migration file
- tenant [schema] [command] [args] - runs migrations for a specific tenant schema
- all-tenants [command] [args]     - runs migrations for all tenant schemas`
	command := &cobra.Command{
		Use:       "migrate",
		Short:     "Execute app DB migration scripts",
		Long:      cmdDesc,
		ValidArgs: []string{"up", "down", "goto", "create", "tenant", "all-tenants"},
		RunE: func(command *cobra.Command, args []string) error {
			cmd := ""
			if len(args) > 0 {
				cmd = args[0]
			}
			switch cmd {
			case "create":
				if len(args) < 2 {
					return fmt.Errorf("migration name is required")
				}
				if err := createMigrationFileHandler(app.MigrationsDir(), args[1]); err != nil {
					return err
				}
			case "tenant":
				if len(args) < 2 {
					return fmt.Errorf("schema name is required")
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
				runner, err := migrate.NewRunner(
					fmt.Sprintf("file:///%s", app.MigrationsDir()),
					fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
						dbCfg.Username, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode))
				if err != nil {
					return err
				}
				return runner.Run(args...)
			}
			return nil
		},
	}

	return command
}

func createMigrationFileHandler(migrationsDir, name string) error {
	timestamp := time.Now().Unix()

	upFileName := fmt.Sprintf("%s/%d_%s.up.sql", migrationsDir, timestamp, name)
	downFileName := fmt.Sprintf("%s/%d_%s.down.sql", migrationsDir, timestamp, name)

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
