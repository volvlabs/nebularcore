package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gorm.io/gorm"
)

// PermissionRepository handles permission-related database operations
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB, schema string) *PermissionRepository {
	return &PermissionRepository{
		db: db,
	}
}

// CreatePermission creates a new permission
func (r *PermissionRepository) CreatePermission(ctx context.Context, data map[string]interface{}) (*models.Permission, error) {
	perm := &models.Permission{}
	if err := r.db.WithContext(ctx).Model(perm).Create(data).Error; err != nil {
		return nil, err
	}
	return perm, nil
}

// CreateGroup creates a new permission group
func (r *PermissionRepository) CreateGroup(ctx context.Context, data map[string]interface{}) (*models.PermissionGroup, error) {
	group := &models.PermissionGroup{}
	if err := r.db.WithContext(ctx).Model(group).Create(data).Error; err != nil {
		return nil, err
	}
	return group, nil
}

// AssignPermissionToUser assigns a permission directly to a user
func (r *PermissionRepository) AssignPermissionToUser(ctx context.Context, userID, permissionID string, grantedBy string, expiresAt *time.Time) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return err
	}
	var grantedByUUID *uuid.UUID
	if grantedBy != "" {
		if parsed, err := uuid.Parse(grantedBy); err == nil {
			grantedByUUID = &parsed
		}
	}
	assignment := &models.UserPermission{
		ID:           uuid.New(),
		UserID:       userUUID,
		PermissionID: permUUID,
		GrantedBy:    grantedByUUID,
		ExpiresAt:    expiresAt,
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// AssignPermissionToRole assigns a permission to a role
func (r *PermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID string, expiresAt *time.Time) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return err
	}
	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return err
	}
	assignment := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       roleUUID,
		PermissionID: permUUID,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// AssignPermissionToGroup assigns a permission to a group
func (r *PermissionRepository) AssignPermissionToGroup(ctx context.Context, groupID, permissionID string, expiresAt *time.Time) error {
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return err
	}
	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return err
	}
	assignment := &models.GroupPermission{
		ID:           uuid.New(),
		GroupID:      groupUUID,
		PermissionID: permUUID,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(assignment).Error
}

// AddUserToGroup adds a user to a permission group
func (r *PermissionRepository) AddUserToGroup(ctx context.Context, userID, groupID string, expiresAt *time.Time) error {
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
func (r *PermissionRepository) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	var permissions []*models.Permission
	err = r.db.WithContext(ctx).
		Distinct().
		Select("permissions.*").
		Joins("LEFT JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Joins("LEFT JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("LEFT JOIN role_assignments ON role_permissions.role_id = role_assignments.role_id").
		Joins("LEFT JOIN group_permissions ON permissions.id = group_permissions.permission_id").
		Joins("LEFT JOIN user_groups ON group_permissions.group_id = user_groups.group_id").
		Where("(user_permissions.user_id = ? AND (user_permissions.expires_at IS NULL OR user_permissions.expires_at > ?)) OR "+
			"(role_assignments.user_id = ? AND (role_assignments.expires_at IS NULL OR role_assignments.expires_at > ?) AND (role_permissions.expires_at IS NULL OR role_permissions.expires_at > ?)) OR "+
			"(user_groups.user_id = ? AND (user_groups.expires_at IS NULL OR user_groups.expires_at > ?) AND (group_permissions.expires_at IS NULL OR group_permissions.expires_at > ?))",
			userUUID, time.Now(),
			userUUID, time.Now(), time.Now(),
			userUUID, time.Now(), time.Now()).
		Find(&permissions).Error
	return permissions, err
}

// GetRolePermissions gets all permissions assigned to a role
func (r *PermissionRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*models.Permission, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return nil, err
	}
	var permissions []*models.Permission
	err = r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND (role_permissions.expires_at IS NULL OR role_permissions.expires_at > ?)",
			roleUUID, time.Now()).
		Find(&permissions).Error
	return permissions, err
}

// GetGroupPermissions gets all permissions assigned to a group
func (r *PermissionRepository) GetGroupPermissions(ctx context.Context, groupID string) ([]*models.Permission, error) {
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
func (r *PermissionRepository) GetUserGroups(ctx context.Context, userID string) ([]*models.UserGroup, error) {
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
