package nebularcore

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gitlab.com/volvlabs/nebularcore/cmd"
	"gitlab.com/volvlabs/nebularcore/core"
	"gitlab.com/volvlabs/nebularcore/models/config"
	"gitlab.com/volvlabs/nebularcore/tools/filesystem"

	"github.com/spf13/cobra"
)

type appWrapper struct {
	core.App
}

type NebularCore struct {
	*appWrapper

	RootCmd *cobra.Command
}

func New(cfg *config.AppConfig) *NebularCore {
	backendApp := &NebularCore{
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
		MigrationsDir:  filepath.Join(filesystem.GetRootDir(""), "migrations"),
	})}

	backendApp.RootCmd.AddCommand(cmd.NewServeCommand(backendApp, cfg.Server))
	backendApp.RootCmd.AddCommand(cmd.NewMigrateCommand(backendApp))
	return backendApp
}

func (b *NebularCore) Execute() error {
	if err := b.appWrapper.Bootstrap(); err != nil {
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
		b.RootCmd.Execute()

		done <- true
	}()

	<-done

	return b.Terminate()
}
