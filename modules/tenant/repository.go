package tenant

import (
	"context"

	"github.com/volvlabs/nebularcore/core/model"
	"gorm.io/gorm"
)

// TenantAwareRepository provides tenant-scoped database operations. It does
// not know or store a schema itself: every method threads ctx through to
// GORM via WithContext, and the Plugin resolves the schema from context on
// each call. This is what keeps repositories free of schema string-building
// — construct one with NewTenantAware(db) and call it the same way
// regardless of which tenant is active.
type TenantAwareRepository[T model.TenantBound] struct {
	db *gorm.DB
}

// NewTenantAware creates a new tenant-aware repository for model T.
func NewTenantAware[T model.TenantBound](db *gorm.DB) *TenantAwareRepository[T] {
	return &TenantAwareRepository[T]{db: db}
}

// Create creates a new record with tenant scoping.
func (r *TenantAwareRepository[T]) Create(ctx context.Context, model *T) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// First finds the first record with tenant scoping.
func (r *TenantAwareRepository[T]) First(ctx context.Context, dest *T, conds ...interface{}) error {
	return r.db.WithContext(ctx).First(dest, conds...).Error
}

// Find finds records with tenant scoping.
func (r *TenantAwareRepository[T]) Find(ctx context.Context, dest *[]T, conds ...interface{}) error {
	return r.db.WithContext(ctx).Find(dest, conds...).Error
}

// Update updates records with tenant scoping.
func (r *TenantAwareRepository[T]) Update(ctx context.Context, model *T, attrs interface{}) error {
	return r.db.WithContext(ctx).Model(model).Updates(attrs).Error
}

// Delete deletes records with tenant scoping.
func (r *TenantAwareRepository[T]) Delete(ctx context.Context, model *T) error {
	return r.db.WithContext(ctx).Delete(model).Error
}

// Transaction executes operations in a transaction. The ctx passed in must
// carry the resolved tenant; it is applied to the transaction's *gorm.DB
// before fn runs so scoped queries inside fn resolve the same tenant.
func (r *TenantAwareRepository[T]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Query returns a tenant-scoped query builder for further chaining.
func (r *TenantAwareRepository[T]) Query(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(new(T))
}
