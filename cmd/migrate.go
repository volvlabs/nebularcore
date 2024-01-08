package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/volvlabs/nebularcore/core"
	"gitlab.com/volvlabs/nebularcore/tools/migrate"
)

func NewMigrateCommand(app core.App) *cobra.Command {
	const cmdDesc = `Supported commands are:
- up			- runs migrtations
- down [number] - revert last number of migrations
- create [name] - creates new blank migration file`
	command := &cobra.Command{
		Use:       "migrate",
		Short:     "Execute app DB migration scripts",
		Long:      cmdDesc,
		ValidArgs: []string{"up", "down", "create"},
		RunE: func(command *cobra.Command, args []string) error {
			cmd := ""
			if len(args) > 0 {
				cmd = args[0]
			}
			switch cmd {
			case "create":
				if err := createMigrationFileHandler(app.MigrationsDir(), args[1]); err != nil {
					return err
				}
			default:
				runner, err := migrate.NewRunner("", "")
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
