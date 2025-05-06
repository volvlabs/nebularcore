package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialAccount represents a social media account linked to a user
type SocialAccount struct {
	ID          uuid.UUID      `gorm:"primaryKey"`
	UserID      string         `gorm:"not null;index"`
	Provider    string         `gorm:"not null;index"`
	ProviderID  string         `gorm:"not null;uniqueIndex:idx_provider_id"`
	Email       string         `gorm:"size:255"`
	Name        string         `gorm:"size:255"`
	AccessToken string         `gorm:"size:1024"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the social account model
func (sa *SocialAccount) TableName() string {
	return "social_accounts"
}
