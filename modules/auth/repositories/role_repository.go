package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a role in the system
type Role struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex:idx_roles_name_tenant"`
	Description string
	Metadata    map[string]interface{} `gorm:"type:jsonb"`
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
