package cmd

import (
	"errors"
	"log"
	"net/http"

	"github.com/volvlabs/nebularcore/apis"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/models/config"

	"github.com/spf13/cobra"
)

func NewServeCommand(app core.App, endpointsConfig config.Endpoints, serveConfig config.ServeConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "serve",
		Args:  cobra.ArbitraryArgs,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			apis.Endpoints(app, endpointsConfig)
			err := apis.Serve(app, serveConfig)
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln(err)
			}
		},
	}

	return command
}
