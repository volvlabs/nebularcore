package core

import (
	"context"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/core/config"
	"gitlab.com/jideobs/nebularcore/core/module"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"gorm.io/gorm"
)

// App defines the core application interface
type App[T config.Settings] interface {
	// Core functionality
	Bootstrap(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Run(ctx context.Context) error

	// Configuration access
	Config() *config.CoreConfig
	Settings() T

	// Module management
	RegisterModule(m module.Module) error
	GetModule(name string) (module.Module, bool)
	GetModulesByNamespace(namespace module.ModuleNamespace) map[string]module.Module

	// Core services
	Router() *gin.Engine
	DB() *gorm.DB
}
