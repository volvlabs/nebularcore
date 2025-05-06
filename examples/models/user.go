package models

import (
	"time"

	"gitlab.com/jideobs/nebularcore/modules/auth/models"
)

// CustomUser extends the base user model with additional fields
type CustomUser struct {
	models.User        // Embed base user model
	FirstName    string
	LastName     string
	DateOfBirth  *time.Time
	Address      string
	CompanyName  string
	Department   string
	Role         string
}

// GetID implements interfaces.User
func (u *CustomUser) GetID() string {
	return u.ID.String()
}

// GetUsername implements interfaces.User
func (u *CustomUser) GetUsername() string {
	return u.Username
}

// GetEmail implements interfaces.User
func (u *CustomUser) GetEmail() string {
	return u.Email
}

// GetPhoneNumber implements interfaces.User
func (u *CustomUser) GetPhoneNumber() string {
	return u.PhoneNumber
}

// GetPasswordHash implements interfaces.User
func (u *CustomUser) GetPasswordHash() string {
	return u.Password
}

// IsActive implements interfaces.User
func (u *CustomUser) IsActive() bool {
	return u.Active
}

// GetLastLoginAt implements interfaces.User
func (u *CustomUser) GetLastLoginAt() *time.Time {
	return u.LastLoginAt
}

// GetMetadata implements interfaces.User
func (u *CustomUser) GetMetadata() map[string]any {
	return u.Metadata
}

// GetFullName is a custom method specific to CustomUser
func (u *CustomUser) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// GetCompanyInfo is a custom method specific to CustomUser
func (u *CustomUser) GetCompanyInfo() (string, string) {
	return u.CompanyName, u.Department
}
