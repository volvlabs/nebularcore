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
	TenantHeader  string     = "X-Tenant-ID"
)

// Model represents a database model that can be tenant-scoped
type Model interface {
	IsTenantBound() bool
}

// Tenant represents a tenant in the system
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

// Module implements the tenant management functionality
type Module struct {
	db            *gorm.DB
	migrationsDir string
}

// ProvidesMigrations implements module.Module
func (m *Module) ProvidesMigrations() bool {
	return true
}

// MigrationDir implements module.MigrationProvider
func (m *Module) MigrationDir() string {
	return m.migrationsDir
}

// New creates a new tenant module
func New(db *gorm.DB, migrationsDir string) *Module {
	return &Module{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// Initialize implements module.Module
func (m *Module) Initialize(ctx context.Context) error {
	// Register GORM callbacks for tenant scoping
	if err := m.db.Callback().Query().Before("gorm:query").Register("tenant:before_query", m.beforeQuery); err != nil {
		return fmt.Errorf("registering query callback: %w", err)
	}

	if err := m.db.Callback().Create().Before("gorm:create").Register("tenant:before_create", m.beforeCreate); err != nil {
		return fmt.Errorf("registering create callback: %w", err)
	}

	return nil
}

// Middleware returns a gin middleware that extracts tenant information from requests
func (m *Module) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader(TenantHeader)
		if tenantID == "" {
			c.Next()
			return
		}

		var tenant Tenant
		if err := m.db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "invalid tenant"})
			return
		}

		ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenant.ID)
		ctx = context.WithValue(ctx, TenantCodeKey, tenant.Code)
		ctx = context.WithValue(ctx, SchemaNameKey, tenant.SchemaName)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// beforeQuery is a GORM callback that adds tenant schema to queries
func (m *Module) beforeQuery(db *gorm.DB) {
	if model, ok := db.Statement.Model.(Model); ok && model.IsTenantBound() {
		if schema, ok := db.Statement.Context.Value(SchemaNameKey).(string); ok && schema != "" {
			db.Statement.Table = fmt.Sprintf("%s.%s", schema, db.Statement.Table)
		} else {
			db.Statement.Table = fmt.Sprintf("%s.%s", DefaultSchema, db.Statement.Table)
		}
	}
}

// beforeCreate is a GORM callback that ensures tenant schema for new records
func (m *Module) beforeCreate(db *gorm.DB) {
	if model, ok := db.Statement.Model.(Model); ok && model.IsTenantBound() {
		if schema, ok := db.Statement.Context.Value(SchemaNameKey).(string); ok && schema != "" {
			db.Statement.Table = fmt.Sprintf("%s.%s", schema, db.Statement.Table)
		} else {
			db.Statement.Table = fmt.Sprintf("%s.%s", DefaultSchema, db.Statement.Table)
		}
	}
}

// GetTenantFromContext extracts tenant information from context
func GetTenantFromContext(ctx context.Context) (id, code, schema string, ok bool) {
	id, ok1 := ctx.Value(TenantIDKey).(string)
	code, ok2 := ctx.Value(TenantCodeKey).(string)
	schema, ok3 := ctx.Value(SchemaNameKey).(string)
	return id, code, schema, ok1 && ok2 && ok3
}

// WithTenant adds tenant information to context
func WithTenant(ctx context.Context, id, code, schema string) context.Context {
	ctx = context.WithValue(ctx, TenantIDKey, id)
	ctx = context.WithValue(ctx, TenantCodeKey, code)
	return context.WithValue(ctx, SchemaNameKey, schema)
}
