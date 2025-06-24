package repositories

import (
	"context"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role repository
type RoleRepository interface {
	CreateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID, roleName string) error
	UnassignRole(ctx context.Context, userID, roleID uuid.UUID, roleName string) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]string, error)
	HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error)
}

// roleRepository handles role-related database operations
type roleRepository struct {
	db       *gorm.DB
	enforcer *casbin.Enforcer
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB, enforcer *casbin.Enforcer) RoleRepository {
	return &roleRepository{
		db:       db,
		enforcer: enforcer,
	}
}

// CreateRole creates a new role
func (r *roleRepository) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// AssignRole assigns a role to a user
func (r *roleRepository) AssignRole(
	ctx context.Context,
	userID, roleID uuid.UUID,
	roleName string,
) error {
	if _, err := r.enforcer.AddRoleForUser(userID.String(), roleName); err != nil {
		return err
	}
	assignment := &models.RoleAssignment{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(assignment).Error; err != nil {
			return err
		}
		return nil
	})
}

// UnassignRole removes a role from a user
func (r *roleRepository) UnassignRole(
	ctx context.Context,
	userID, roleID uuid.UUID,
	roleName string,
) error {
	if _, err := r.enforcer.DeleteRoleForUser(userID.String(), roleName); err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).
			Delete(&models.RoleAssignment{}, "user_id = ? AND role_id = ?", userID, roleID).Error
	})
}

// GetUserRoles gets all roles assigned to a user
func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN role_assignments ON roles.id = role_assignments.role_id").
		Where("role_assignments.user_id = ? AND (role_assignments.expires_at IS NULL OR role_assignments.expires_at > ?)",
			userID, time.Now()).
		Find(&roles).Error
	return roles, err
}

// GetRoleUsers gets all users assigned to a role
func (r *roleRepository) GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Model(&models.RoleAssignment{}).
		Where("role_id = ? AND (expires_at IS NULL OR expires_at > ?)",
			roleID, time.Now()).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// HasRole checks if a user has a specific role
func (r *roleRepository) HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RoleAssignment{}).
		Joins("JOIN roles ON role_assignments.role_id = roles.id").
		Where("role_assignments.user_id = ? AND roles.name = ? AND (role_assignments.expires_at IS NULL OR role_assignments.expires_at > ?)",
			userID, roleName, time.Now()).
		Count(&count).Error
	return count > 0, err
}
