package models

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/tools/types"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID       `gorm:"primaryKey" json:"id"`
	Email              string          `gorm:"uniqueIndex" json:"email"`
	PhoneNumber        string          `gorm:"uniqueIndex" json:"phoneNumber"`
	Username           string          `gorm:"uniqueIndex" json:"username"`
	Password           string          `json:"-"`
	Metadata           json.RawMessage `gorm:"type:jsonb" json:"-"`
	Active             bool            `gorm:"default:true" json:"active"`
	PasswordResetToken *string         `json:"-"`
	Token              *string         `json:"-"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	LastLoginAt        *time.Time      `json:"-"`
	PasswordResetAt    *time.Time      `json:"-"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
	Role               string          `json:"role"`
	EmailVerified      bool            `json:"emailVerified"`
	EmailVerifiedAt    *types.DateTime `json:"-"`
} // @name User

func (u *User) GetID() uuid.UUID {
	return u.ID
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetPhoneNumber() string {
	return u.PhoneNumber
}

func (u *User) GetPasswordHash() string {
	return u.Password
}

func (u *User) IsActive() bool {
	return u.Active
}

func (u *User) GetLastLoginAt() *time.Time {
	return u.LastLoginAt
}

func (u *User) GetMetadata() map[string]any {
	metadata := map[string]any{}
	if len(bytes.TrimSpace(u.Metadata)) == 0 {
		return metadata
	}

	err := json.Unmarshal(u.Metadata, &metadata)
	if err != nil {
		log.Err(err).Bytes("metadata", u.Metadata).
			Any("ID", u.GetID()).
			Msg("unmarhsalling of user metadata failed")
	}
	return metadata
}

func (u *User) GetPasswordResetToken() *string {
	return u.PasswordResetToken
}

func (u *User) GetPasswordResetAt() *time.Time {
	return u.PasswordResetAt
}

func (u *User) GetRole() string {
	return u.Role
}

func (u *User) GetEmailVerified() bool {
	return u.EmailVerified
}

func (u *User) SetPasswordResetToken(token *string) {
	u.PasswordResetToken = token
}

func (u *User) SetPasswordResetAt(at *time.Time) {
	u.PasswordResetAt = at
}

func (u *User) SetPassword(password string) {
	u.Password = password
}

func (u *User) GetEmailVerifiedAt() *types.DateTime {
	return u.EmailVerifiedAt
}
