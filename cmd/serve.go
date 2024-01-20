package cmd

import (
	"log"
	"net/http"

	"gitlab.com/jideobs/nebularcore/apis"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models/config"

	"github.com/spf13/cobra"
)

func NewServeCommand(app core.App, serveConfig config.ServeConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "serve",
		Args:  cobra.ArbitraryArgs,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			err := apis.Serve(app, serveConfig)
			if err != http.ErrServerClosed {
				log.Fatalln(err)
			}
		},
	}

	return command
}
