package tenant

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type contextKey string

const (
	TenantIDKey   contextKey = "tenant_id"
	TenantCodeKey contextKey = "tenant_code"
	SchemaNameKey contextKey = "schema_name"
	DefaultSchema string     = "public"

	// DefaultHeader is the HTTP header the middleware reads by default.
	// Callers with a two-level hierarchy (e.g. organization -> farm, where
	// the farm is the schema-bearing unit) should configure a different
	// header via WithHeader and keep this name free for their own
	// org-level authorization layer, so the two don't collide.
	DefaultHeader string = "X-Tenant-ID"
)

// Tenant represents the schema-bearing unit in the system. In a two-level
// tenancy model (e.g. organization -> farm) this maps to the *leaf* level
// that actually owns a schema; the level above it is application-specific
// and lives outside this module.
type Tenant struct {
	ID         string `gorm:"type:uuid;primary_key"`
	Code       string `gorm:"type:varchar(50);unique;not null"`
	Name       string `gorm:"type:varchar(255);not null"`
	SchemaName string `gorm:"type:varchar(50);unique;not null"`
	Active     bool   `gorm:"default:true"`
}

func (Tenant) TableName() string {
	return "tenants"
}

// AuthorizeFunc validates that the caller (typically resolved from a JWT
// earlier in the middleware chain) is actually a member of the requested
// tenant. It runs after the tenant is looked up and before it's placed in
// context. Returning false aborts the request with 403. When nil, the
// middleware only checks that the tenant exists and is active — callers
// that need membership enforcement (see the identity/auth module) must
// supply this.
type AuthorizeFunc func(c *gin.Context, tenant Tenant) bool

// Module implements tenant schema registration, migrations, and request
// resolution.
type Module struct {
	db            *gorm.DB
	migrationsDir string
	header        string
	authorize     AuthorizeFunc
}

// Option configures a Module.
type Option func(*Module)

// WithHeader overrides the HTTP header the middleware reads to resolve the
// active tenant. Defaults to DefaultHeader.
func WithHeader(header string) Option {
	return func(m *Module) { m.header = header }
}

// WithAuthorize sets the membership check run after tenant lookup. See
// AuthorizeFunc.
func WithAuthorize(fn AuthorizeFunc) Option {
	return func(m *Module) { m.authorize = fn }
}

// New creates a new tenant module.
func New(db *gorm.DB, migrationsDir string, opts ...Option) *Module {
	m := &Module{
		db:            db,
		migrationsDir: migrationsDir,
		header:        DefaultHeader,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// ProvidesMigrations implements module.MigrationProvider.
func (m *Module) ProvidesMigrations() bool {
	return true
}

// MigrationDir implements module.MigrationProvider.
func (m *Module) MigrationDir() string {
	return m.migrationsDir
}

// Initialize registers the tenant-scoping GORM plugin. Call this once
// during app startup, after the db connection is established.
func (m *Module) Initialize(ctx context.Context) error {
	if err := m.db.Use(NewPlugin()); err != nil {
		return fmt.Errorf("tenant module: registering plugin: %w", err)
	}
	return nil
}

// Middleware returns a gin middleware that resolves the tenant named by the
// configured header, checks it exists and is active, runs the optional
// AuthorizeFunc, and places its schema into the request context for the
// tenant plugin to pick up. Requests without the header pass through
// unscoped — routes that require a tenant should guard for that themselves
// (a tenant-bound query with no schema in context fails closed, it does not
// silently serve public data).
func (m *Module) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader(m.header)
		if tenantID == "" {
			c.Next()
			return
		}

		var t Tenant
		if err := m.db.Where("id = ? AND active = ?", tenantID, true).First(&t).Error; err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "invalid or inactive tenant"})
			return
		}

		if m.authorize != nil && !m.authorize(c, t) {
			c.AbortWithStatusJSON(403, gin.H{"error": "not a member of this tenant"})
			return
		}

		ctx := WithTenant(c.Request.Context(), t.ID, t.Code, t.SchemaName)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// SchemaFromContext extracts just the resolved schema name, if any. This is
// what the tenant plugin uses; it does not require id/code to also be set.
func SchemaFromContext(ctx context.Context) (schema string, ok bool) {
	if ctx == nil {
		return "", false
	}
	schema, ok = ctx.Value(SchemaNameKey).(string)
	return schema, ok
}

// GetTenantFromContext extracts the full tenant identity from context. All
// three values must be present for ok to be true; use SchemaFromContext if
// you only need the schema.
func GetTenantFromContext(ctx context.Context) (id, code, schema string, ok bool) {
	id, ok1 := ctx.Value(TenantIDKey).(string)
	code, ok2 := ctx.Value(TenantCodeKey).(string)
	schema, ok3 := ctx.Value(SchemaNameKey).(string)
	return id, code, schema, ok1 && ok2 && ok3
}

// WithTenant adds tenant information to context.
func WithTenant(ctx context.Context, id, code, schema string) context.Context {
	ctx = context.WithValue(ctx, TenantIDKey, id)
	ctx = context.WithValue(ctx, TenantCodeKey, code)
	return context.WithValue(ctx, SchemaNameKey, schema)
}

// ResolveSchemaContext looks up tenantID's schema name and returns a
// context carrying it (via WithTenant) — for callers with no HTTP request
// to hang the tenant-resolving Middleware off of (e.g. a Kafka consumer),
// which otherwise have no way to get a tenant-scoped context for
// TenantBound model reads/writes.
func ResolveSchemaContext(ctx context.Context, db *gorm.DB, tenantID string) (context.Context, error) {
	var t Tenant
	if err := db.WithContext(ctx).Where("id = ?", tenantID).First(&t).Error; err != nil {
		return nil, fmt.Errorf("tenant: resolving schema for tenant %s: %w", tenantID, err)
	}
	return WithTenant(ctx, t.ID, t.Code, t.SchemaName), nil
}
