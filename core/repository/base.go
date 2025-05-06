package repository

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines the base interface for all repositories
type Repository[T any] interface {
	// Basic CRUD operations
	Create(ctx context.Context, model *T) error
	First(ctx context.Context, dest *T, conds ...interface{}) error
	Find(ctx context.Context, dest *[]T, conds ...interface{}) error
	Update(ctx context.Context, model *T, attrs interface{}) error
	Delete(ctx context.Context, model *T) error

	// Transaction support
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// Raw query access
	Query(ctx context.Context) *gorm.DB

	// DB access for custom repositories
	DB() *gorm.DB
}

// BaseRepository provides a base repository implementation
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBase creates a new base repository instance
func NewBase[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

// Create creates a new record
func (r *BaseRepository[T]) Create(ctx context.Context, model *T) error {
	return r.db.Create(model).Error
}

// First finds the first record
func (r *BaseRepository[T]) First(ctx context.Context, dest *T, conds ...interface{}) error {
	return r.db.WithContext(ctx).First(dest, conds...).Error
}

// Find finds records
func (r *BaseRepository[T]) Find(ctx context.Context, dest *[]T, conds ...interface{}) error {
	return r.db.WithContext(ctx).Find(dest, conds...).Error
}

// Update updates records
func (r *BaseRepository[T]) Update(ctx context.Context, model *T, attrs interface{}) error {
	return r.db.WithContext(ctx).Model(model).Updates(attrs).Error
}

// Delete deletes records
func (r *BaseRepository[T]) Delete(ctx context.Context, model *T) error {
	return r.db.WithContext(ctx).Delete(model).Error
}

// Transaction executes operations in a transaction
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// Query returns a new query
func (r *BaseRepository[T]) Query(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(new(T))
}

// DB returns the underlying database connection
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}
