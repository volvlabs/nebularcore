package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/models/config"
)

func NewServeCommand[T config.Settings](app core.App[T]) *cobra.Command {
	command := &cobra.Command{
		Use:   "serve",
		Args:  cobra.ArbitraryArgs,
		Short: "Bootstrap and start the application",
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Name() != "serve" {
				return
			}

			if err := app.Bootstrap(context.Background()); err != nil {
				log.Fatalln(err)
			}

			if err := app.Run(context.Background()); err != nil {
				log.Fatalln(err)
			}
		},
	}

	return command
}
