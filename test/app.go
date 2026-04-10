package test

import (
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/models/config"
	"github.com/volvlabs/nebularcore/tools/auth"
	"github.com/volvlabs/nebularcore/tools/eventclient"
)

func getTempDataDirName(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var output []byte
	for i := 0; i < length; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomIndex := r.Intn(len(charset))
		output = append(output, charset[randomIndex])
	}
	return string(output)
}

type TestApp struct {
	*core.BaseApp
}

func (t *TestApp) CleanUp() {
	os.RemoveAll(t.DataDir())
}

func (t *TestApp) EventClient() eventclient.Client {
	return &eventClientMock{}
}

func NewTestApp() (*TestApp, error) {
	app := &TestApp{}

	cfg, _ := config.New("")

	_, currentFile, _, _ := runtime.Caller(0)
	cfg.BaseDir = filepath.Join(path.Dir(currentFile), "../")
	cfg.TestDir = filepath.Join(cfg.BaseDir, "test/data")

	dataDir := filepath.Join(os.TempDir(), getTempDataDirName(12))

	// Create data directory for each app created, this is to prevent multiple unit tests
	// trying to perform operations on the same data.
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return nil, err
	}

	app.BaseApp = core.NewBaseApp(
		core.BaseAppConfig{
			Env:        "test",
			DataDir:    dataDir,
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
