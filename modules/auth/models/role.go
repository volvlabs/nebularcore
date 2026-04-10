package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Name        string    `gorm:"uniqueIndex:idx_roles_name_tenant"`
	Description string
	Metadata    map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
} // @name Role

// RoleAssignment represents a role assignment to a user
type RoleAssignment struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"index:idx_role_assignments_user_role"`
	RoleID    uuid.UUID `gorm:"index:idx_role_assignments_user_role"`
	CreatedAt time.Time
	ExpiresAt *time.Time
} // @name RoleAssignment
