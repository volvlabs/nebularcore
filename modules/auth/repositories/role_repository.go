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

// RoleMetadata is arbitrary per-role metadata stored as jsonb — e.g. the
// "delegable" convention consuming apps use to mark a role as assignable
// by a tenant-level admin, not just a platform superadmin (see
// AuthorizationManager.GetRole's doc comment). A named type (rather than
// a bare map[string]interface{}) is required here so it can implement
// sql.Scanner/driver.Valuer — the plain map type has no way to
// (de)serialize a jsonb column on its own, which is what previously made
// reading a role's Metadata back out of the database fail.
type RoleMetadata map[string]interface{}

// Value implements driver.Valuer.
func (m RoleMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner.
func (m *RoleMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported Scan type for RoleMetadata: %T", value)
	}
	return json.Unmarshal(bytes, m)
}

// Role represents a role in the system
type Role struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex:idx_roles_name_tenant"`
	Description string
	Metadata    RoleMetadata `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate generates the primary key when the caller didn't set one.
// Needed because GORM only omits zero-value primary keys from struct-based
// INSERTs for recognized numeric/autoincrement types — a zero-value string
// PK is sent as an explicit empty string, which defeats the column's
// DEFAULT gen_random_uuid() and fails outright against a uuid column type.
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

// RoleAssignment represents a user-role assignment
type RoleAssignment struct {
	ID        string `gorm:"primaryKey"`
	UserID    string `gorm:"index:idx_role_assignments_user_role"`
	RoleID    string `gorm:"index:idx_role_assignments_user_role"`
	CreatedAt time.Time
	ExpiresAt *time.Time
}

// BeforeCreate generates the primary key when the caller didn't set one. See
// Role.BeforeCreate for why this is necessary.
func (a *RoleAssignment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

// RoleRepository handles role-related database operations
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

// CreateRole creates a new role
func (r *RoleRepository) CreateRole(ctx context.Context, data map[string]interface{}) (*Role, error) {
	role := &Role{}
	if err := r.db.WithContext(ctx).Model(role).Create(data).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// AssignRole assigns a role to a user
func (r *RoleRepository) AssignRole(ctx context.Context, userID, roleID string, expiresAt *time.Time) error {
	assignment := &RoleAssignment{
		UserID:    userID,
		RoleID:    roleID,
		ExpiresAt: expiresAt,
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// UnassignRole removes a role from a user
func (r *RoleRepository) UnassignRole(ctx context.Context, userID, roleID string) error {
	return r.db.WithContext(ctx).Delete(&RoleAssignment{}, "user_id = ? AND role_id = ?", userID, roleID).Error
}

// GetUserRoles gets all roles assigned to a user
func (r *RoleRepository) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	var roles []*Role
	err := r.db.WithContext(ctx).
		Joins("JOIN role_assignments ON roles.id = role_assignments.role_id").
		Where("role_assignments.user_id = ? AND (role_assignments.expires_at IS NULL OR role_assignments.expires_at > ?)",
			userID, time.Now()).
		Find(&roles).Error
	return roles, err
}

// GetRoleUsers gets all users assigned to a role
func (r *RoleRepository) GetRoleUsers(ctx context.Context, roleID string) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Model(&RoleAssignment{}).
		Where("role_id = ? AND (expires_at IS NULL OR expires_at > ?)",
			roleID, time.Now()).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// GetRoleByName returns a role by name.
func (r *RoleRepository) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles returns every role.
func (r *RoleRepository) ListRoles(ctx context.Context) ([]*Role, error) {
	var roles []*Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

// UpdateRole updates description/metadata for an existing role by name.
// Name is intentionally not updatable here — casbin's policy/grouping rows
// reference roles by name string, not ID, so renaming would mean rewriting
// every p/g row that references the old name; delete+recreate is the
// supported way to change a role's identity.
func (r *RoleRepository) UpdateRole(ctx context.Context, name string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Role{}).Where("name = ?", name).Updates(updates).Error
}

// DeleteRole deletes a role's assignments then the role row itself. Casbin-
// side cascade (removing the role's p/g policy rows) is the caller's
// (AuthorizationManager's) responsibility — this only handles the
// roles/role_assignments tables.
func (r *RoleRepository) DeleteRole(ctx context.Context, name string) error {
	role, err := r.GetRoleByName(ctx, name)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Delete(&RoleAssignment{}, "role_id = ?", role.ID).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(role).Error
}

// HasRole checks if a user has a specific role
func (r *RoleRepository) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&RoleAssignment{}).
		Joins("JOIN roles ON role_assignments.role_id = roles.id").
		Where("role_assignments.user_id = ? AND roles.name = ? AND (role_assignments.expires_at IS NULL OR role_assignments.expires_at > ?)",
			userID, roleName, time.Now()).
		Count(&count).Error
	return count > 0, err
}
