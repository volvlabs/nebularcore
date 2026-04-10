package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/core/model"
	"github.com/volvlabs/nebularcore/modules/auth/types"
)

// SocialAccount represents a social media account linked to a user
type SocialAccount struct {
	model.Model
	UserID         uuid.UUID          `gorm:"not null;index"`
	User           User               `gorm:"foreignKey:UserID;references:ID"`
	Provider       types.AuthProvider `gorm:"not null;index"`
	ProviderUserID string             `gorm:"not null;uniqueIndex:idx_social_accounts_provider_id"`
	Email          string             `gorm:"unqiueIndex:idx_social_accounts_provider_email"`
	Name           string             `gorm:"size:255"`
	Metadata       json.RawMessage    `gorm:"type:jsonb" json:"-"`
} // @name SocialAccount

// TableName returns the table name for the social account model
func (sa *SocialAccount) TableName() string {
	return "social_accounts"
}
