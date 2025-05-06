package tenant

import (
	"context"
	"fmt"

	"gitlab.com/jideobs/nebularcore/core/model"
	"gorm.io/gorm"
)

// TenantAwareRepository provides tenant-scoped database operations
type TenantAwareRepository[T model.TenantBound] struct {
	db     *gorm.DB
	schema string
}

// NewTenantAware creates a new tenant-aware repository
func NewTenantAware[T model.TenantBound](db *gorm.DB, schema string) *TenantAwareRepository[T] {
	if schema == "" {
		schema = DefaultSchema
	}
	return &TenantAwareRepository[T]{
		db:     db,
		schema: schema,
	}
}

// withTenant adds tenant schema to the query
func (r *TenantAwareRepository[T]) withTenant(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.Table(fmt.Sprintf("%s.%s", r.schema, db.Statement.Table))
}

// Create creates a new record with tenant scoping
func (r *TenantAwareRepository[T]) Create(ctx context.Context, model *T) error {
	return r.withTenant(ctx, r.db).Create(model).Error
}

// First finds the first record with tenant scoping
func (r *TenantAwareRepository[T]) First(ctx context.Context, dest *T, conds ...interface{}) error {
	return r.withTenant(ctx, r.db).First(dest, conds...).Error
}

// Find finds records with tenant scoping
func (r *TenantAwareRepository[T]) Find(ctx context.Context, dest *[]T, conds ...interface{}) error {
	return r.withTenant(ctx, r.db).Find(dest, conds...).Error
}

// Update updates records with tenant scoping
func (r *TenantAwareRepository[T]) Update(ctx context.Context, model *T, attrs interface{}) error {
	return r.withTenant(ctx, r.db).Model(model).Updates(attrs).Error
}

// Delete deletes records with tenant scoping
func (r *TenantAwareRepository[T]) Delete(ctx context.Context, model *T) error {
	return r.withTenant(ctx, r.db).Delete(model).Error
}

// Transaction executes operations in a transaction with tenant scoping
func (r *TenantAwareRepository[T]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)
		return fn(tx)
	})
}

// Query returns a new query with tenant scoping
func (r *TenantAwareRepository[T]) Query(ctx context.Context) *gorm.DB {
	return r.withTenant(ctx, r.db.Model(new(T)))
}
