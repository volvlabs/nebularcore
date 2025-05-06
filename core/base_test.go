package core_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/core/module"
	moduleMocks "gitlab.com/jideobs/nebularcore/core/module/mocks"
)

// TestSettings implements config.Settings interface for testing
type TestSettings struct {
	TestValue string `yaml:"testValue"`
}

func (s TestSettings) Validate() error {
	return nil
}

func (s TestSettings) IsProduction() bool {
	return false
}

// Helper function to create a temporary config file
func createTempConfig(t *testing.T) string {
	content := `
core:
  environment: test
  server:
    host: localhost
    port: 8080
    readTimeout: 30s
    writeTimeout: 30s
    shutdownTimeout: 30s
  database:
    driver: sqlite
    host: localhost
    port: 5432
    name: test_db
    username: test
    password: test
    sslmode: disable
project:
  testValue: test
modules:
  test-module:
    enabled: true
`

	tmpfile, err := os.CreateTemp("", "config*.yaml")
	assert.NoError(t, err)

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)

	return tmpfile.Name()
}

func TestNew(t *testing.T) {
	configPath := createTempConfig(t)
	defer os.Remove(configPath)

	// Test successful initialization
	opts := core.Options[TestSettings]{
		ConfigPath: configPath,
		EnvPrefix:  "TEST",
	}

	app, err := core.New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, app)

	// Test with invalid config path
	opts.ConfigPath = "nonexistent.yaml"
	app, err = core.New(opts)
	assert.Error(t, err)
	assert.Nil(t, app)
}

func TestModuleRegistration(t *testing.T) {
	configPath := createTempConfig(t)
	defer os.Remove(configPath)

	opts := core.Options[TestSettings]{
		ConfigPath: configPath,
		EnvPrefix:  "TEST",
	}

	app, err := core.New(opts)
	assert.NoError(t, err)

	// Test module registration
	mockModule := moduleMocks.NewModule(t)
	mockModule.On("Namespace").Return(module.PublicNamespace)
	mockModule.On("Name").Return("test-module")
	mockModule.On("Dependencies").Return([]string{})
	err = app.RegisterModule(mockModule)
	assert.NoError(t, err)

	// Test module retrieval
	modules := app.GetModules()
	assert.Len(t, modules, 1)
	assert.Contains(t, modules, mockModule.Name())

	// Test getting module by name
	module, exists := app.GetModule(mockModule.Name())
	assert.True(t, exists)
	assert.Equal(t, mockModule, module)

	// Test getting modules by namespace
	namespaceModules := app.GetModulesByNamespace(mockModule.Namespace())
	assert.Len(t, namespaceModules, 1)
}

func TestBootstrapAndShutdown(t *testing.T) {
	type ModuleConfig struct {
		Enabled bool `yaml:"enabled"`
	}

	configPath := createTempConfig(t)
	defer os.Remove(configPath)

	opts := core.Options[TestSettings]{
		ConfigPath: configPath,
		EnvPrefix:  "TEST",
	}

	app, err := core.New(opts)
	assert.NoError(t, err)

	// Setup mock module
	mockModule := moduleMocks.NewModule(t)
	mockModule.On("Namespace").Return(module.PublicNamespace)
	mockModule.On("Name").Return("test-module")
	mockModule.On("NewConfig").Return(&ModuleConfig{})
	mockModule.On("Configure", mock.Anything).Return(nil)
	mockModule.On("Dependencies").Return([]string{})
	mockModule.On("Initialize", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockModule.On("Shutdown", mock.Anything).Return(nil)

	err = app.RegisterModule(mockModule)
	assert.NoError(t, err)

	// Test bootstrap
	ctx := context.Background()
	err = app.Bootstrap(ctx)
	assert.NoError(t, err)

	// Verify database initialization
	db := app.DB()
	assert.NotNil(t, db)

	// Verify router initialization
	router := app.Router()
	assert.NotNil(t, router)

	// Test shutdown
	err = app.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify mock expectations
	mockModule.AssertExpectations(t)
}

func TestConfiguration(t *testing.T) {
	configPath := createTempConfig(t)
	defer os.Remove(configPath)

	opts := core.Options[TestSettings]{
		ConfigPath: configPath,
		EnvPrefix:  "TEST",
	}

	app, err := core.New(opts)
	assert.NoError(t, err)

	// Test core config
	coreConfig := app.Config()
	assert.NotNil(t, coreConfig)
	assert.Equal(t, "localhost", coreConfig.Server.Host)
	assert.Equal(t, "8080", coreConfig.Server.Port)
	assert.Equal(t, 30*time.Second, coreConfig.Server.ReadTimeout)

	// Test project settings
	settings := app.Settings()
	assert.Equal(t, "test", settings.TestValue)

	// Test project root
	assert.NotEmpty(t, coreConfig.ProjectRoot)
	_, err = os.Stat(coreConfig.ProjectRoot)
	assert.NoError(t, err)
}
