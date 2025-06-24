package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID      `json:"id" gorm:"primaryKey"`
	Email              string         `json:"email" gorm:"uniqueIndex"`
	PhoneNumber        string         `json:"phoneNumber" gorm:"uniqueIndex"`
	Username           string         `json:"username" gorm:"uniqueIndex"`
	Password           string         `json:"-"`
	Metadata           map[string]any `json:"metadata" gorm:"type:jsonb"`
	Active             bool           `json:"active" gorm:"default:true"`
	PasswordResetToken *string        `json:"-"`
	Token              *string        `json:"-"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	LastLoginAt        *time.Time     `json:"lastLoginAt"`
	PasswordResetAt    *time.Time     `json:"passwordResetAt"`
	DeletedAt          gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Role               string         `json:"role"`
}

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
	return u.Metadata
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

func (u *User) SetPasswordResetToken(token *string) {
	u.PasswordResetToken = token
}

func (u *User) SetPasswordResetAt(at *time.Time) {
	u.PasswordResetAt = at
}

func (u *User) SetPassword(password string) {
	u.Password = password
}
