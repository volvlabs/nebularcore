package core_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/volvlabs/nebularcore/core"
	coreConfig "github.com/volvlabs/nebularcore/core/config"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	moduleMocks "github.com/volvlabs/nebularcore/core/module/mocks"
	"gorm.io/gorm"
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

type ModuleConfig struct {
	Enabled bool `yaml:"enabled"`
}

func (m *ModuleConfig) Key() string {
	return "test-module"
}

func (m *ModuleConfig) Validate() error {
	return nil
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
	defer func() { _ = os.Remove(configPath) }()

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
	defer func() { _ = os.Remove(configPath) }()

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
	configPath := createTempConfig(t)
	defer func() { _ = os.Remove(configPath) }()

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
	defer func() { _ = os.Remove(configPath) }()

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

// globalMiddlewareFakeModule implements module.GlobalMiddlewareModule: its
// GlobalMiddleware stamps every request's gin context with a marker before
// any module's Initialize runs, simulating auth's soft-identify middleware.
type globalMiddlewareFakeModule struct {
	name  string
	order *[]string
}

func (m *globalMiddlewareFakeModule) Name() string                      { return m.name }
func (m *globalMiddlewareFakeModule) Version() string                   { return "1.0.0" }
func (m *globalMiddlewareFakeModule) Dependencies() []string            { return nil }
func (m *globalMiddlewareFakeModule) NewConfig() coreConfig.Config      { return nil }
func (m *globalMiddlewareFakeModule) Configure(coreConfig.Config) error { return nil }
func (m *globalMiddlewareFakeModule) Shutdown(context.Context) error    { return nil }
func (m *globalMiddlewareFakeModule) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}
func (m *globalMiddlewareFakeModule) ProvidesMigrations() bool { return false }
func (m *globalMiddlewareFakeModule) MigrationsDir() string    { return "" }
func (m *globalMiddlewareFakeModule) GetMigrationSources(string) []migrationRunner.Source {
	return nil
}
func (m *globalMiddlewareFakeModule) Initialize(_ context.Context, _ *gorm.DB, _ *gin.Engine) error {
	*m.order = append(*m.order, "initialize:"+m.name)
	return nil
}
func (m *globalMiddlewareFakeModule) GlobalMiddleware(context.Context, *gorm.DB) ([]gin.HandlerFunc, error) {
	*m.order = append(*m.order, "globalMiddleware:"+m.name)
	return []gin.HandlerFunc{
		func(c *gin.Context) {
			c.Set("marker", "set-by-"+m.name)
			c.Next()
		},
	}, nil
}

// routeFakeModule does NOT implement GlobalMiddlewareModule. Its Initialize
// registers a route group — mirroring how a real module (e.g. billing)
// both creates its own groups and relies on whatever global middleware
// already ran to have populated the context.
type routeFakeModule struct {
	name  string
	order *[]string
}

func (m *routeFakeModule) Name() string                      { return m.name }
func (m *routeFakeModule) Version() string                   { return "1.0.0" }
func (m *routeFakeModule) Dependencies() []string            { return nil }
func (m *routeFakeModule) NewConfig() coreConfig.Config      { return nil }
func (m *routeFakeModule) Configure(coreConfig.Config) error { return nil }
func (m *routeFakeModule) Shutdown(context.Context) error    { return nil }
func (m *routeFakeModule) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}
func (m *routeFakeModule) ProvidesMigrations() bool { return false }
func (m *routeFakeModule) MigrationsDir() string    { return "" }
func (m *routeFakeModule) GetMigrationSources(string) []migrationRunner.Source {
	return nil
}
func (m *routeFakeModule) Initialize(_ context.Context, _ *gorm.DB, router *gin.Engine) error {
	*m.order = append(*m.order, "initialize:"+m.name)
	group := router.Group("/probe")
	group.GET("/marker", func(c *gin.Context) {
		marker, _ := c.Get("marker")
		c.JSON(http.StatusOK, gin.H{"marker": marker})
	})
	return nil
}

// TestBootstrapAppliesGlobalMiddlewareBeforeAnyModuleInitialize is a
// regression test for the AccountCountry bug this interface was added to
// fix: a module's own router.Use(...) inside Initialize only reaches route
// groups created by modules initialized afterward, which made
// localization's AccountCountry resolution depend on registration order
// relative to auth instead of being guaranteed. GlobalMiddlewareModule
// gives modules a dedicated phase, run for every implementing module
// before ANY module's Initialize (and thus before any module creates a
// route group) — this asserts both the ordering and that a later,
// non-GlobalMiddlewareModule's route group actually inherits it.
func TestBootstrapAppliesGlobalMiddlewareBeforeAnyModuleInitialize(t *testing.T) {
	configPath := createTempConfig(t)
	defer func() { _ = os.Remove(configPath) }()

	app, err := core.New(core.Options[TestSettings]{ConfigPath: configPath, EnvPrefix: "TEST"})
	assert.NoError(t, err)

	var order []string
	mw := &globalMiddlewareFakeModule{name: "mw-module", order: &order}
	route := &routeFakeModule{name: "route-module", order: &order}

	// route-module registered first: if global-middleware ordering were
	// still tied to registration/Initialize order the way the old
	// router.Use-inside-Initialize approach was, its group would be
	// created before mw-module's middleware ever ran.
	assert.NoError(t, app.RegisterModule(route))
	assert.NoError(t, app.RegisterModule(mw))

	assert.NoError(t, app.Bootstrap(context.Background()))

	// Both GlobalMiddleware calls must happen before either Initialize
	// call, regardless of registration order.
	assert.Equal(t, []string{
		"globalMiddleware:mw-module",
		"initialize:route-module",
		"initialize:mw-module",
	}, order)

	req := httptest.NewRequest(http.MethodGet, "/probe/marker", nil)
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"marker":"set-by-mw-module"}`, w.Body.String())
}
