package tenant

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	coreModel "github.com/volvlabs/nebularcore/core/model"
	"gorm.io/gorm"
)

// ErrNoTenantInContext is returned when a tenant-bound model is touched
// without a resolved schema in the statement context. The plugin fails
// closed: it never falls back to the public schema for tenant-bound data,
// because a silent fallback would mean a missing tenant context reads or
// writes into the wrong schema instead of failing loudly.
var ErrNoTenantInContext = errors.New("tenant: no schema resolved in context for tenant-bound model")

// Plugin is a GORM plugin that schema-qualifies tenant-bound models based on
// the schema stored in the query context (see WithTenant / SchemaNameKey).
// Register it once on the *gorm.DB and every Create/Query/Update/Delete/Row
// against a model implementing model.TenantBound is routed to the caller's
// tenant schema automatically. Repositories and handlers never need to
// concatenate schema names onto table names themselves.
//
// Raw SQL (db.Raw / db.Exec) is intentionally NOT covered — the caller wrote
// the SQL string, so schema-qualification is the caller's job. Use
// QualifiedTable(ctx, tableName) to build the right identifier by hand, and
// treat any raw SQL against a tenant-bound table as a review flag.
type Plugin struct{}

// NewPlugin constructs the tenant scoping plugin.
func NewPlugin() *Plugin { return &Plugin{} }

// Name implements gorm.Plugin.
func (p *Plugin) Name() string { return "nebularcore:tenant" }

// Initialize implements gorm.Plugin. It registers resolveSchema at the
// earliest point in each callback chain where Statement.Table is still safe
// to rewrite, before GORM builds any SQL from it.
func (p *Plugin) Initialize(db *gorm.DB) error {
	registrations := []struct {
		name   string
		before string
		reg    func(name string, fn func(*gorm.DB)) error
	}{
		{"tenant:before_create", "gorm:before_create", db.Callback().Create().Before("gorm:before_create").Register},
		{"tenant:before_query", "gorm:query", db.Callback().Query().Before("gorm:query").Register},
		{"tenant:before_update", "gorm:before_update", db.Callback().Update().Before("gorm:before_update").Register},
		{"tenant:before_delete", "gorm:before_delete", db.Callback().Delete().Before("gorm:before_delete").Register},
		{"tenant:before_row", "gorm:row", db.Callback().Row().Before("gorm:row").Register},
	}

	for _, r := range registrations {
		if err := r.reg(r.name, resolveSchema); err != nil {
			return fmt.Errorf("tenant plugin: registering %s before %s: %w", r.name, r.before, err)
		}
	}

	return nil
}

// resolveSchema is the shared callback body for every stage the plugin
// hooks. It only acts on models that opt in via model.TenantBound, and it
// is idempotent: if the table is already schema-qualified (e.g. a caller
// used db.Table("x.y") explicitly) it leaves it alone rather than
// double-prefixing.
func resolveSchema(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Table == "" {
		return
	}

	if !isTenantBound(db) {
		return
	}

	if alreadyQualified(db) {
		return
	}

	schema, ok := SchemaFromContext(db.Statement.Context)
	if !ok || schema == "" {
		_ = db.AddError(fmt.Errorf("%w: table %q", ErrNoTenantInContext, db.Statement.Table))
		return
	}

	db.Statement.Table = schema + "." + db.Statement.Table
}

// isTenantBound determines whether the current statement's model type opts
// into tenant scoping. GORM's processor sets Statement.Model = Statement.Dest
// before any callback runs, which for slice destinations (e.g. a plain
// Find(&[]T{}) with no explicit .Model() call) makes Model a *[]T rather
// than a *T — a direct type assertion on Model would miss that case
// entirely. Statement.Schema.ModelType is populated from the same parse
// step and reliably reflects the *element* type regardless of how Model was
// shaped, so it is checked first; the direct assertion remains as a
// fallback for the rare case Schema isn't set yet.
func isTenantBound(db *gorm.DB) bool {
	if db.Statement.Schema != nil {
		v := reflect.New(db.Statement.Schema.ModelType).Interface()
		if bound, ok := v.(coreModel.TenantBound); ok {
			return bound.IsTenantBound()
		}
		return false
	}
	if bound, ok := db.Statement.Model.(coreModel.TenantBound); ok {
		return bound.IsTenantBound()
	}
	return false
}

// alreadyQualified reports whether the caller already fully specified the
// table, so the plugin should not touch it. A dotted name passed to
// db.Table("schema.table") is NOT reflected in Statement.Table itself —
// GORM splits it, keeping the bare table name there and moving the quoted
// "schema"."table" expression into Statement.TableExpr — so both fields
// must be checked or a caller-specified schema gets silently overridden by
// the context-resolved one.
func alreadyQualified(db *gorm.DB) bool {
	if db.Statement.TableExpr != nil && strings.Contains(db.Statement.TableExpr.SQL, ".") {
		return true
	}
	return strings.Contains(db.Statement.Table, ".")
}

// QualifiedTable returns "<schema>.<table>" for the tenant resolved in ctx,
// for the rare case a caller must write raw SQL against a tenant-bound
// table. It returns an error under the same fail-closed rule as the plugin.
func QualifiedTable(ctx context.Context, table string) (string, error) {
	schema, ok := SchemaFromContext(ctx)
	if !ok || schema == "" {
		return "", fmt.Errorf("%w: table %q", ErrNoTenantInContext, table)
	}
	return schema + "." + table, nil
}
