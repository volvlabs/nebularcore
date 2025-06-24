package models

import (
	"time"

	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/core/model"
)

// Role represents a role in the system
type Role struct {
	model.Model
	Name        string                 `json:"name" gorm:"uniqueIndex:idx_roles_name_tenant"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// RoleAssignment represents a role assignment to a user
type RoleAssignment struct {
	model.Model
	UserID    uuid.UUID  `json:"userId" gorm:"index:idx_role_assignments_user_role"`
	RoleID    uuid.UUID  `json:"roleId" gorm:"index:idx_role_assignments_user_role"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
