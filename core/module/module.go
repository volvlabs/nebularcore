package module

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/core/config"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
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
	NewConfig() config.Config
	Configure(config config.Config) error
	Shutdown(ctx context.Context) error

	// Namespace returns the module's namespace (public or tenant)
	Namespace() ModuleNamespace

	// Optional methods - modules can implement these interfaces if needed
	// MigrationProvider for modules that need database migrations
	ProvidesMigrations() bool
	MigrationsDir() string
	GetMigrationSources(projectRoot string) []migrationRunner.Source
}

// GlobalMiddlewareModule is an optional interface for modules that need to
// add gin middleware to the router itself, rather than to one of their own
// route groups — e.g. auth's soft-identify middleware (populates the
// authenticated user on the context without requiring one, so later
// middleware can see it), or localization's AccountCountry/DetectedCountry
// resolution (which needs the authenticated user already on the context to
// resolve AccountCountry, so it must run after auth's soft-identify).
//
// A module's own Initialize(ctx, db, router) is the wrong place for this:
// gin.RouterGroup.Group() copies the engine's current handler chain at
// creation time, so router.Use(...) only affects groups created
// afterward. Since every module's Initialize also registers that module's
// own route groups, a router.Use call inside one module's Initialize only
// reliably runs before another module's routes if that module happens to
// be initialized first — an implicit, unenforced ordering dependency
// bugs hide behind easily (see the AccountCountry incident this interface
// was added to fix: it depended on auth's Initialize having already run,
// which registration order didn't guarantee).
//
// GlobalMiddlewareModule fixes this by giving these modules a dedicated
// hook, called for every implementing module — in dependency/registration
// order (see Registry.GetModulesInOrder) — in its own phase, before any
// module's Initialize runs and thus before any module creates a route
// Group. Implement it only for middleware that must apply globally, ahead
// of every route; anything scoped to one module's own routes belongs in
// that module's Initialize as before.
type GlobalMiddlewareModule interface {
	Module

	// GlobalMiddleware returns the gin middleware this module needs
	// registered on the router before any module's Initialize runs. May
	// perform the same lazy, nil-guarded setup Initialize would otherwise
	// do (constructing resolvers, etc.) so the module is fully usable
	// regardless of whether Initialize or GlobalMiddleware runs first —
	// see auth.Module and localization.Module for the pattern.
	GlobalMiddleware(ctx context.Context, db *gorm.DB) ([]gin.HandlerFunc, error)
}
