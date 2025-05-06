package authorization

import (
	"context"
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories"
	"gorm.io/gorm"
)

// AuthorizationManager handles role and permission management
type AuthorizationManager struct {
	enforcer *casbin.Enforcer
	roleRepo *repositories.RoleRepository
	db       *gorm.DB
}

// NewAuthorizationManager creates a new authorization manager
func NewAuthorizationManager(db *gorm.DB) (*AuthorizationManager, error) {
	// Initialize Casbin adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Initialize enforcer with RBAC model
	enforcer, err := casbin.NewEnforcer("auth_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load RBAC model
	err = enforcer.LoadModel()
	if err != nil {
		return nil, fmt.Errorf("failed to load RBAC model: %w", err)
	}

	// Initialize role repository
	roleRepo := repositories.NewRoleRepository(db)

	return &AuthorizationManager{
		enforcer: enforcer,
		roleRepo: roleRepo,
		db:       db,
	}, nil
}

// CreateRole creates a new role
func (m *AuthorizationManager) CreateRole(ctx context.Context, name, description string, metadata map[string]interface{}) error {
	_, err := m.roleRepo.CreateRole(ctx, map[string]interface{}{
		"name":        name,
		"description": description,
		"metadata":    metadata,
	})
	return err
}

// AssignRole assigns a role to a user
func (m *AuthorizationManager) AssignRole(ctx context.Context, userID, roleName string, duration *time.Duration) error {
	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find role
	var role repositories.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Calculate expiry if duration is provided
	var expiresAt *time.Time
	if duration != nil {
		t := time.Now().Add(*duration)
		expiresAt = &t
	}

	// Assign role in database
	if err := m.roleRepo.AssignRole(ctx, userID, role.ID, expiresAt); err != nil {
		tx.Rollback()
		return err
	}

	// Add role to Casbin
	if _, err := m.enforcer.AddRoleForUser(userID, roleName); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UnassignRole removes a role from a user
func (m *AuthorizationManager) UnassignRole(ctx context.Context, userID, roleName string) error {
	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find role
	var role repositories.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Remove role assignment from database
	if err := m.roleRepo.UnassignRole(ctx, userID, role.ID); err != nil {
		tx.Rollback()
		return err
	}

	// Remove role from Casbin
	if _, err := m.enforcer.DeleteRoleForUser(userID, roleName); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GrantPermission grants a permission to a role
func (m *AuthorizationManager) GrantPermission(ctx context.Context, roleName, resource, action string) error {
	_, err := m.enforcer.AddPolicy(roleName, resource, action)
	return err
}

// RevokePermission revokes a permission from a role
func (m *AuthorizationManager) RevokePermission(ctx context.Context, roleName, resource, action string) error {
	_, err := m.enforcer.RemovePolicy(roleName, resource, action)
	return err
}

// HasPermission checks if a user has a specific permission
func (m *AuthorizationManager) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	return m.enforcer.Enforce(userID, resource, action)
}

// GetUserRoles gets all roles assigned to a user
func (m *AuthorizationManager) GetUserRoles(ctx context.Context, userID string) ([]*repositories.Role, error) {
	return m.roleRepo.GetUserRoles(ctx, userID)
}

// GetRolePermissions gets all permissions assigned to a role
func (m *AuthorizationManager) GetRolePermissions(ctx context.Context, roleName string) ([][]string, error) {
	permissions, err := m.enforcer.GetPermissionsForUser(roleName)
	return permissions, err
}

// HasRole checks if a user has a specific role
func (m *AuthorizationManager) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	return m.roleRepo.HasRole(ctx, userID, roleName)
}
