package test

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"gitlab.com/volvlabs/nebularcore/core"
	"gitlab.com/volvlabs/nebularcore/models/config"
	"gitlab.com/volvlabs/nebularcore/tools/auth"
)

type TestApp struct {
	*core.BaseApp
}

func NewTestApp() (*TestApp, error) {
	app := &TestApp{}

	cfg, _ := config.New("")

	_, currentFile, _, _ := runtime.Caller(0)
	cfg.BaseDir = filepath.Join(path.Dir(currentFile), "../")
	cfg.TestDir = filepath.Join(cfg.BaseDir, "test/data")

	app.BaseApp = core.NewBaseApp(
		core.BaseAppConfig{
			Env:        "test",
			DataDir:    os.TempDir(),
			IsDev:      true,
			EnforceAcl: true,
		})

	policyFilePath := filepath.Join(cfg.TestDir, "policy.csv")
	modelFilePath := filepath.Join(cfg.TestDir, "model.conf")
	app.Acm().RegisterAll([]auth.AclConfig{
		{Role: "admin", PolicyPath: policyFilePath, ConfPath: modelFilePath},
		{Role: "user", PolicyPath: policyFilePath, ConfPath: modelFilePath},
	})

	if err := app.Bootstrap(); err != nil {
		return nil, err
	}

	return app, nil
}
