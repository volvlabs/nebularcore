package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID `gorm:"primaryKey"`
	Email              string    `gorm:"uniqueIndex"`
	PhoneNumber        string    `gorm:"uniqueIndex"`
	Username           string    `gorm:"uniqueIndex"`
	Password           string
	Metadata           map[string]any `gorm:"type:jsonb"`
	Active             bool           `gorm:"default:true"`
	PasswordResetToken *string
	Token              *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastLoginAt        *time.Time
	PasswordResetAt    *time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
	Role               string
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
