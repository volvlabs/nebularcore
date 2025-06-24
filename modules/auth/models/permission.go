package models

import (
	"time"

	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/core/model"
	"gorm.io/gorm"
)

// Permission represents a permission in the system
type Permission struct {
	model.Model
	Name        string         `json:"name" gorm:"index"`
	ResourceID  uuid.UUID      `json:"resourceId"`
	Resource    string         `json:"resource" gorm:"foreignKey:ModuleResource;references:ID"`
	Action      string         `json:"action" gorm:"index"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata" gorm:"type:jsonb"`
}

// PermissionGroup represents a group of permissions
type PermissionGroup struct {
	model.Model
	Name        string `gorm:"uniqueIndex:idx_permission_groups_name"`
	Description string
	Metadata    map[string]any `gorm:"type:jsonb"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// UserPermission represents a direct permission assignment to a user
type UserPermission struct {
	model.Model
	UserID       uuid.UUID `gorm:"index:idx_user_permissions_user_perm"`
	PermissionID uuid.UUID `gorm:"index:idx_user_permissions_user_perm"`
	GrantedBy    *uuid.UUID
	ExpiresAt    *time.Time
}

// RolePermission represents a permission assignment to a role
type RolePermission struct {
	model.Model
	RoleID       uuid.UUID `gorm:"index:idx_role_permissions_role_perm"`
	PermissionID uuid.UUID `gorm:"index:idx_role_permissions_role_perm"`
	GrantedByID  uuid.UUID
	ExpiresAt    *time.Time
}

// GroupPermission represents a permission assignment to a group
type GroupPermission struct {
	model.Model
	GroupID      uuid.UUID `gorm:"index:idx_group_permissions_group_perm"`
	PermissionID uuid.UUID `gorm:"index:idx_group_permissions_group_perm"`
	ExpiresAt    *time.Time
}
