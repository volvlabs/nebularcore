package nebularcore

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/volvlabs/nebularcore/cmd"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/models/config"

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
		TenantConfig:   cfg.TenantConfig,
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
		if err := n.RootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "command error: %v\n", err)
		}

		done <- true
	}()

	<-done

	return n.Terminate()
}
