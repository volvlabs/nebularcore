package nebularcore

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/volvlabs/nebularcore/cmd"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/core/config"

	"github.com/spf13/cobra"
)

type appWrapper[T config.Settings] struct {
	core.App[T]
}

type NebularCore[T config.Settings] struct {
	appWrapper[T]
	RootCmd *cobra.Command
}

func New[T config.Settings](options core.Options[T]) *NebularCore[T] {
	backendApp := &NebularCore[T]{
		RootCmd: &cobra.Command{
			Use:   fmt.Sprintf("%s [configfilepath]", filepath.Base(os.Args[0])),
			Short: "Backend CLI",
		},
	}
	app, err := core.New(options)
	if err != nil {
		panic(err)
	}
	backendApp.appWrapper = appWrapper[T]{app}

	return backendApp
}

func (n *NebularCore[T]) Start(ctx context.Context) error {
	n.RootCmd.AddCommand(cmd.NewServeCommand(n))
	n.RootCmd.AddCommand(cmd.NewMigrateCommand(n, n.Config().Database))

	return n.Execute(ctx)
}

func (n *NebularCore[T]) Execute(ctx context.Context) error {
	done := make(chan bool, 1)

	// listen for signal interrupt.
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	go func() {
		if err := n.RootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "command error: %v\n", err)
		}

		done <- true
	}()

	<-done

	return n.Shutdown(ctx)
}
