package repositories

import (
	"context"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	CreateGroup(ctx context.Context, group *models.PermissionGroup) (*models.PermissionGroup, error)
	AssignPermissionToUser(ctx context.Context, userID, permissionID, grantedBy uuid.UUID) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID, grantedByID uuid.UUID, roleName string) error
	AssignPermissionToGroup(ctx context.Context, groupID, permissionID string, expiresAt *time.Time) error
	AddUserToGroup(ctx context.Context, userID, groupID string, expiresAt *time.Time) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error)
	GetRolePermissions(
		ctx context.Context,
		roleID uuid.UUID,
	) ([]*models.Permission, error)
	RevokePermissionToRole(
		ctx context.Context,
		roleID, permissionID uuid.UUID,
		roleName string,
	) error
	RevokePermissionToUser(
		ctx context.Context,
		userID, permissionID uuid.UUID,
	) error
	GetGroupPermissions(ctx context.Context, groupID string) ([]*models.Permission, error)
	GetUserGroups(ctx context.Context, userID string) ([]*models.UserGroup, error)
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// permissionRepository handles permission-related database operations
type permissionRepository struct {
	db       *gorm.DB
	enforcer *casbin.Enforcer
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB, enforcer *casbin.Enforcer) PermissionRepository {
	return &permissionRepository{
		db:       db,
		enforcer: enforcer,
	}
}

// CreatePermission creates a new permission
func (r *permissionRepository) CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	if err := r.db.WithContext(ctx).Model(permission).Create(permission).Error; err != nil {
		return nil, err
	}
	return permission, nil
}

// CreateGroup creates a new permission group
func (r *permissionRepository) CreateGroup(ctx context.Context, group *models.PermissionGroup) (*models.PermissionGroup, error) {
	if err := r.db.WithContext(ctx).Model(group).Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

// AssignPermissionToUser assigns a permission directly to a user
func (r *permissionRepository) AssignPermissionToUser(
	ctx context.Context,
	userID, permissionID, grantedBy uuid.UUID,
) error {
	permission := &models.Permission{}
	if err := r.db.WithContext(ctx).Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return err
	}

	if _, err := r.enforcer.AddPolicy(userID.String(), permission.Name, permission.Action); err != nil {
		return err
	}

	assignment := &models.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
		GrantedBy:    &grantedBy,
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// AssignPermissionToRole assigns a permission to a role
func (r *permissionRepository) AssignPermissionToRole(
	ctx context.Context,
	roleID, permissionID, grantedBy uuid.UUID,
	roleName string,
) error {
	permission := &models.Permission{}
	if err := r.db.WithContext(ctx).Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return err
	}

	if _, err := r.enforcer.AddPolicy(roleName, permission.Name, permission.Action); err != nil {
		return err
	}

	assignment := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		GrantedByID:  grantedBy,
	}

	return r.db.WithContext(ctx).Create(assignment).Error
}

func (r *permissionRepository) RevokePermissionToRole(
	ctx context.Context,
	roleID, permissionID uuid.UUID,
	roleName string,
) error {
	permission := &models.Permission{}
	if err := r.db.WithContext(ctx).Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return err
	}

	if _, err := r.enforcer.RemovePolicy(roleName, permission.Name, permission.Action); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.RolePermission{}, "role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

func (r *permissionRepository) RevokePermissionToUser(
	ctx context.Context,
	userID, permissionID uuid.UUID,
) error {
	permission := &models.Permission{}
	if err := r.db.WithContext(ctx).Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return err
	}

	if _, err := r.enforcer.RemovePolicy(userID.String(), permission.Name, permission.Action); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.UserPermission{}, "user_id = ? AND permission_id = ?", userID, permissionID).Error
}

// AssignPermissionToGroup assigns a permission to a group
func (r *permissionRepository) AssignPermissionToGroup(ctx context.Context, groupID, permissionID string, expiresAt *time.Time) error {
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return err
	}
	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return err
	}
	assignment := &models.GroupPermission{
		GroupID:      groupUUID,
		PermissionID: permUUID,
		ExpiresAt:    expiresAt,
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// AddUserToGroup adds a user to a permission group
func (r *permissionRepository) AddUserToGroup(ctx context.Context, userID, groupID string, expiresAt *time.Time) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return err
	}
	assignment := &models.UserGroup{
		ID:        uuid.New(),
		UserID:    userUUID,
		GroupID:   groupUUID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// GetUserPermissions gets all permissions assigned to a user (direct, role-based, and group-based)
func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).
		Distinct().
		Select("permissions.*").
		Joins("LEFT JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Joins("LEFT JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("LEFT JOIN role_assignments ON role_permissions.role_id = role_assignments.role_id").
		Joins("LEFT JOIN group_permissions ON permissions.id = group_permissions.permission_id").
		Joins("LEFT JOIN user_groups ON group_permissions.group_id = user_groups.group_id").
		Where("(user_permissions.user_id = ? OR "+
			"role_assignments.user_id = ? OR "+
			"user_groups.user_id = ?)",
			userID, userID, userID).Find(&permissions).Error
	return permissions, err
}

// GetRolePermissions gets all permissions assigned to a role
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND (role_permissions.expires_at IS NULL OR role_permissions.expires_at > ?)",
			roleID, time.Now()).
		Find(&permissions).Error
	return permissions, err
}

// GetGroupPermissions gets all permissions assigned to a group
func (r *permissionRepository) GetGroupPermissions(ctx context.Context, groupID string) ([]*models.Permission, error) {
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, err
	}
	var permissions []*models.Permission
	err = r.db.WithContext(ctx).
		Joins("JOIN group_permissions ON permissions.id = group_permissions.permission_id").
		Where("group_permissions.group_id = ? AND (group_permissions.expires_at IS NULL OR group_permissions.expires_at > ?)",
			groupUUID, time.Now()).
		Find(&permissions).Error
	return permissions, err
}

// GetUserGroups gets all groups a user belongs to
func (r *permissionRepository) GetUserGroups(ctx context.Context, userID string) ([]*models.UserGroup, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	var groups []*models.UserGroup
	err = r.db.WithContext(ctx).
		Joins("JOIN user_groups ON user_groups.group_id = permission_groups.id").
		Where("user_groups.user_id = ? AND (user_groups.expires_at IS NULL OR user_groups.expires_at > ?)",
			userUUID, time.Now()).
		Find(&groups).Error
	return groups, err
}

func (r *permissionRepository) HasPermission(
	ctx context.Context,
	userID uuid.UUID,
	resource, action string,
) (bool, error) {
	return r.enforcer.Enforce(userID.String(), resource, action)
}
