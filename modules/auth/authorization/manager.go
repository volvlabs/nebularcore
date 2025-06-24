package authorization

import (
	"context"

	"github.com/google/uuid"

	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories"
	"gitlab.com/jideobs/nebularcore/tools/validation"
)

// Manager handles role and permission management
type Manager interface {
	// HasPermission checks if a user has a specific permission
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)

	// HasRole checks if a user has a specific role
	HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error)

	CreateRole(ctx context.Context, payload *requests.CreateRolePayload) error

	// AssignRole assigns a role to a user
	AssignRole(ctx context.Context, userID, roleID uuid.UUID, roleName string) error

	// UnassignRole removes a role from a user
	UnassignRole(ctx context.Context, userID, roleID uuid.UUID, roleName string) error

	CreatePermission(ctx context.Context, payload *requests.CreatePermissionPayload) (*models.Permission, error)

	// GrantRolePermission grants a permission to a role
	GrantRolePermission(ctx context.Context, roleID, permissionID, grantedByID uuid.UUID, roleName string) error

	// RevokeRolePermission revokes a permission from a role
	RevokeRolePermission(ctx context.Context, roleID, permissionID uuid.UUID, roleName string) error

	// GrantUserPermission grants a permission directly to a user
	GrantUserPermission(ctx context.Context, userID, permissionID, grantedByID uuid.UUID) error

	// RevokeUserPermission revokes a direct permission from a user
	RevokeUserPermission(ctx context.Context, userID, permissionID uuid.UUID) error

	// GetUserRoles gets all roles assigned to a user
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error)

	// GetUserDirectPermissions gets all direct permissions assigned to a user
	GetUserDirectPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error)

	// GetRolePermissions gets all permissions assigned to a role
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)
}

// manager handles role and permission management
type manager struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	validator      *validation.Validator
}

// NewAuthorizationManager creates a new authorization manager
func NewAuthorizationManager(
	roleRepository repositories.RoleRepository,
	permissionRepository repositories.PermissionRepository,
) Manager {
	return &manager{
		roleRepo:       roleRepository,
		permissionRepo: permissionRepository,
		validator:      validation.New(),
	}
}

// CreateRole creates a new role
func (m *manager) CreateRole(
	ctx context.Context,
	payload *requests.CreateRolePayload,
) error {
	if err := m.validator.Validate(payload); err != nil {
		return err
	}

	role := &models.Role{
		Name:        payload.Name,
		Description: payload.Description,
		Metadata:    payload.Metadata,
	}
	_, err := m.roleRepo.CreateRole(ctx, role)

	return err
}

// AssignRole assigns a role to a user
func (m *manager) AssignRole(
	ctx context.Context,
	userID, roleID uuid.UUID,
	roleName string,
) error {
	return m.roleRepo.AssignRole(ctx, userID, roleID, roleName)
}

// UnassignRole removes a role from a user
func (m *manager) UnassignRole(
	ctx context.Context,
	userID, roleID uuid.UUID,
	roleName string,
) error {
	return m.roleRepo.UnassignRole(ctx, userID, roleID, roleName)
}

func (m *manager) CreatePermission(
	ctx context.Context,
	payload *requests.CreatePermissionPayload,
) (*models.Permission, error) {
	if err := m.validator.Validate(payload); err != nil {
		return nil, err
	}

	permission := &models.Permission{
		Name:        payload.Name,
		ResourceID:  payload.ResourceID,
		Action:      payload.Action,
		Description: payload.Description,
		Metadata:    payload.Metadata,
	}
	return m.permissionRepo.CreatePermission(ctx, permission)
}

// GrantRolePermission grants a permission to a role
func (m *manager) GrantRolePermission(
	ctx context.Context,
	roleID, permissionID, grantedByID uuid.UUID,
	roleName string,
) error {
	return m.permissionRepo.AssignPermissionToRole(
		ctx, roleID, permissionID, grantedByID, roleName)
}

// RevokeRolePermission revokes a permission from a role
func (m *manager) RevokeRolePermission(
	ctx context.Context,
	roleID, permissionID uuid.UUID,
	roleName string,
) error {
	return m.permissionRepo.RevokePermissionToRole(ctx, roleID, permissionID, roleName)
}

// GrantUserPermission grants a permission directly to a user
func (m *manager) GrantUserPermission(
	ctx context.Context,
	userID, permissionID, grantedByID uuid.UUID,
) error {
	return m.permissionRepo.AssignPermissionToUser(ctx, userID, permissionID, grantedByID)
}

// RevokeUserPermission revokes a direct permission from a user
func (m *manager) RevokeUserPermission(
	ctx context.Context,
	userID, permissionID uuid.UUID,
) error {
	return m.permissionRepo.RevokePermissionToUser(ctx, userID, permissionID)
}

// GetUserRoles gets all roles assigned to a user
func (m *manager) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	return m.roleRepo.GetUserRoles(ctx, userID)
}

// GetRolePermissions gets all permissions assigned to a role
func (m *manager) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	return m.permissionRepo.GetRolePermissions(ctx, roleID)
}

// GetUserDirectPermissions gets all direct permissions assigned to a user
func (m *manager) GetUserDirectPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	return m.permissionRepo.GetUserPermissions(ctx, userID)
}

// HasRole checks if a user has a specific role
func (m *manager) HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	return m.roleRepo.HasRole(ctx, userID, roleName)
}

// HasPermission checks if a subject has a specific permission
func (m *manager) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	return m.permissionRepo.HasPermission(ctx, userID, resource, action)
}
