package tenant

import (
	"errors"
	"fmt"

	"github.com/volvlabs/nebularcore/core/config"
	"github.com/volvlabs/nebularcore/core/module"
	"gorm.io/gorm"
)

// ProvisionSchema is the generic half of "onboard a new tenant": create its
// Postgres schema, register it in the tenants table, and run every
// TenantNamespace module's migrations against it (tenantModules — e.g. from
// app.GetModulesInOrder(module.TenantNamespace), resolved once by the
// caller) — the mechanical, application-agnostic steps any schema-per-tenant
// onboarding flow needs, regardless of what the tenant is called in the
// calling app's domain language (farm, workspace, account, ...). Idempotent:
// safe to retry (CREATE SCHEMA IF NOT EXISTS, and each module's migration
// Up() is a no-op once current).
//
// Callers own tenant-naming concerns entirely — schemaName here must
// already be validated and fully formed (e.g. "farm_acme"); this function
// does no slug validation or prefixing of its own, since that's
// application-specific, not framework-generic. Takes db/dbCfg/tenantModules
// directly rather than an app.App[T] handle, so callers deep inside
// application code (with only a *gorm.DB, not the app instance) can use it
// without threading the whole app through their own call chains.
func ProvisionSchema(db *gorm.DB, dbCfg config.DatabaseConfig, tenantModules []module.OrderedModule, projectRoot, id, code, name, schemaName string) (Tenant, error) {
	if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q", schemaName)).Error; err != nil {
		return Tenant{}, fmt.Errorf("tenant: creating schema %s: %w", schemaName, err)
	}

	var t Tenant
	err := db.Where("schema_name = ?", schemaName).First(&t).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		t = Tenant{ID: id, Code: code, Name: name, SchemaName: schemaName, Active: true}
		if err := db.Create(&t).Error; err != nil {
			return Tenant{}, fmt.Errorf("tenant: registering tenant for schema %s: %w", schemaName, err)
		}
	case err != nil:
		return Tenant{}, fmt.Errorf("tenant: looking up tenant for schema %s: %w", schemaName, err)
	}

	if err := migrateSchema(dbCfg, tenantModules, projectRoot, schemaName); err != nil {
		return Tenant{}, err
	}

	return t, nil
}

// MigrateAllSchemas runs every TenantNamespace module's migrations
// (tenantModules) against every known tenant schema — the generic
// replacement for a bespoke "migrate all my per-tenant schemas" loop. Meant
// to run on every deploy.
func MigrateAllSchemas(db *gorm.DB, dbCfg config.DatabaseConfig, tenantModules []module.OrderedModule, projectRoot string) error {
	schemas, err := getAllSchemas(db)
	if err != nil {
		return fmt.Errorf("tenant: listing schemas: %w", err)
	}
	for _, schema := range schemas {
		if err := migrateSchema(dbCfg, tenantModules, projectRoot, schema); err != nil {
			return err
		}
	}
	return nil
}
