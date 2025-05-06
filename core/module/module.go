package module

import (
	"context"

	"github.com/gin-gonic/gin"
	migrationRunner "gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gorm.io/gorm"
)

// Module defines the interface that all modules must implement
// ModuleNamespace defines the type of schema a module operates in
type ModuleNamespace string

const (
	// PublicNamespace represents modules that operate in the public schema
	PublicNamespace ModuleNamespace = "public"
	// TenantNamespace represents modules that operate in tenant-specific schemas
	TenantNamespace ModuleNamespace = "tenant"
)

type Module interface {
	// Basic module information
	Name() string
	Version() string
	Dependencies() []string

	// Lifecycle methods
	Initialize(ctx context.Context, db *gorm.DB, router *gin.Engine) error
	NewConfig() any
	Configure(config any) error
	Shutdown(ctx context.Context) error

	// Namespace returns the module's namespace (public or tenant)
	Namespace() ModuleNamespace

	// Optional methods - modules can implement these interfaces if needed
	// MigrationProvider for modules that need database migrations
	ProvidesMigrations() bool
	MigrationsDir() string
	GetMigrationSources(projectRoot string) []migrationRunner.Source
}
