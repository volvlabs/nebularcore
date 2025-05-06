package models

import (
	"time"

	"github.com/google/uuid"
)

// UserGroup represents a user's membership in a group
type UserGroup struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"index:idx_user_groups_user"`
	GroupID   uuid.UUID `gorm:"index:idx_user_groups_group"`
	CreatedAt time.Time
	ExpiresAt *time.Time
}
