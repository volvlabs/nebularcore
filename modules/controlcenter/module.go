package controlcenter

import (
	"context"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/core/config"
	"gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gitlab.com/jideobs/nebularcore/core/module"
	"gorm.io/gorm"
)

type coreApp interface {
	GetModulesByNamespace(namespace module.ModuleNamespace) map[string]module.Module
	Config() *config.CoreConfig
}

// Module implements the event system using Watermill
type Module struct {
	app coreApp
}

// New creates a new event module
func New(app coreApp) *Module {
	return &Module{app: app}
}

// Configure implements module.Module.
func (m *Module) Configure(config any) error {
	return nil
}

// Dependencies implements module.Module.
func (m *Module) Dependencies() []string {
	return nil
}

// GetMigrationSources implements module.Module.
func (m *Module) GetMigrationSources(projectRoot string) []migration_runner.Source {
	return nil
}

// MigrationsDir implements module.Module.
func (m *Module) MigrationsDir() string {
	return ""
}

// Namespace implements module.Module.
func (m *Module) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}

// NewConfig implements module.Module.
func (m *Module) NewConfig() any {
	return nil
}

// ProvidesMigrations implements module.Module.
func (m *Module) ProvidesMigrations() bool {
	return false
}

// Version implements module.Module.
func (m *Module) Version() string {
	return "1.0.0"
}

// Name returns the module name
func (m *Module) Name() string {
	return "controlcenter"
}

// Initialize implements core.Module interface
func (m *Module) Initialize(
	ctx context.Context,
	db *gorm.DB,
	router *gin.Engine,
) error {
	newHandler(m.app).RegisterRoutes(router.Group("/controlcenter"))
	return nil
}

// Shutdown implements core.Module interface
func (m *Module) Shutdown(ctx context.Context) error {
	return nil
}
