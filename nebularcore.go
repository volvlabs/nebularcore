package nebularcore

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gitlab.com/jideobs/nebularcore/cmd"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models/config"

	"github.com/spf13/cobra"
)

type appWrapper struct {
	core.App
}

type NebularCore struct {
	*appWrapper

	cfg     *config.AppConfig
	RootCmd *cobra.Command
}

func New(cfg *config.AppConfig) core.App {
	backendApp := &NebularCore{
		cfg: cfg,
		RootCmd: &cobra.Command{
			Use:   fmt.Sprintf("%s [configfilepath]", filepath.Base(os.Args[0])),
			Short: "Backend CLI",
		},
	}
	backendApp.appWrapper = &appWrapper{core.NewBaseApp(core.BaseAppConfig{
		Env:            cfg.Env,
		IsDev:          cfg.IsDev,
		EnforceAcl:     cfg.EnforceAcl,
		DatabaseConfig: cfg.Database,
		MigrationsDir:  cfg.MigrationsDir,
	})}

	return backendApp
}

func (n *NebularCore) Start() error {
	n.RootCmd.AddCommand(cmd.NewServeCommand(n, n.cfg.Endpoints, n.cfg.Server))
	n.RootCmd.AddCommand(cmd.NewMigrateCommand(n, n.cfg.Database))

	return n.Execute()
}

func (n *NebularCore) Execute() error {
	if err := n.appWrapper.Bootstrap(); err != nil {
		return err
	}

	done := make(chan bool, 1)

	// listen for signal interrupt.
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	go func() {
		n.RootCmd.Execute()

		done <- true
	}()

	<-done

	return n.Terminate()
}
