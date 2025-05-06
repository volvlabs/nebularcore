package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents a permission in the system
type Permission struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"index"`
	Resource    string `gorm:"index"`
	Action      string `gorm:"index"`
	Description string
	Metadata    map[string]any `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// PermissionGroup represents a group of permissions
type PermissionGroup struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Name        string    `gorm:"uniqueIndex:idx_permission_groups_name"`
	Description string
	Metadata    map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// UserPermission represents a direct permission assignment to a user
type UserPermission struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	UserID       uuid.UUID `gorm:"index:idx_user_permissions_user_perm"`
	PermissionID uuid.UUID `gorm:"index:idx_user_permissions_user_perm"`
	CreatedAt    time.Time
	ExpiresAt    *time.Time
	GrantedBy    *uuid.UUID
}

// RolePermission represents a permission assignment to a role
type RolePermission struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	RoleID       uuid.UUID `gorm:"index:idx_role_permissions_role_perm"`
	PermissionID uuid.UUID `gorm:"index:idx_role_permissions_role_perm"`
	CreatedAt    time.Time
	ExpiresAt    *time.Time
}

// GroupPermission represents a permission assignment to a group
type GroupPermission struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	GroupID      uuid.UUID `gorm:"index:idx_group_permissions_group_perm"`
	PermissionID uuid.UUID `gorm:"index:idx_group_permissions_group_perm"`
	CreatedAt    time.Time
	ExpiresAt    *time.Time
}
