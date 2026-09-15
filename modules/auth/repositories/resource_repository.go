package repositories

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResourceActions is the set of action names meaningful for a resource
// (e.g. "read", "write", "delete", "manage", or any consuming-app-defined
// action string like "feedback") — stored as jsonb. A named type (rather
// than a bare []string) is required here so it can implement
// sql.Scanner/driver.Valuer, mirroring RoleMetadata's exact rationale: the
// plain slice type has no way to (de)serialize a jsonb column on its own.
type ResourceActions []string

// Value implements driver.Valuer.
func (a ResourceActions) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner.
func (a *ResourceActions) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported Scan type for ResourceActions: %T", value)
	}
	return json.Unmarshal(bytes, a)
}

// Resource represents a resource in the system that roles can be granted
// permissions against (e.g. "zones", "devices", "staff") — the catalog a
// dashboard uses to build a role×resource×action permission editor. This
// is deliberately not a casbin principal itself; permissions reference a
// resource by its Name string directly in casbin's own p rows.
type Resource struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	Description string
	Actions     ResourceActions `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate generates the primary key when the caller didn't set one.
// See Role.BeforeCreate for why this is necessary.
func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

// ResourceRepository handles resource-related database operations
type ResourceRepository struct {
	db *gorm.DB
}

// NewResourceRepository creates a new resource repository
func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{
		db: db,
	}
}

// CreateResource creates a new resource
func (r *ResourceRepository) CreateResource(ctx context.Context, data map[string]interface{}) (*Resource, error) {
	resource := &Resource{}
	if err := r.db.WithContext(ctx).Model(resource).Create(data).Error; err != nil {
		return nil, err
	}
	return resource, nil
}

// GetResourceByName returns a resource by name.
func (r *ResourceRepository) GetResourceByName(ctx context.Context, name string) (*Resource, error) {
	var resource Resource
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&resource).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

// ListResources returns every resource.
func (r *ResourceRepository) ListResources(ctx context.Context) ([]*Resource, error) {
	var resources []*Resource
	err := r.db.WithContext(ctx).Find(&resources).Error
	return resources, err
}

// UpdateResource updates description/actions for an existing resource by
// name. Name is intentionally not updatable here — permission grants
// reference resources by name string, not ID; delete+recreate is the
// supported way to change a resource's identity.
func (r *ResourceRepository) UpdateResource(ctx context.Context, name string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Resource{}).Where("name = ?", name).Updates(updates).Error
}

// DeleteResource deletes a resource by name. Casbin-side cascade (removing
// any p rows that reference this resource) is the caller's
// (AuthorizationManager's) responsibility.
func (r *ResourceRepository) DeleteResource(ctx context.Context, name string) error {
	resource, err := r.GetResourceByName(ctx, name)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(resource).Error
}
