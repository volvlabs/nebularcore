package models

import "github.com/volvlabs/nebularcore/modules/auth/interfaces"

// CustomUserFactory creates instances of CustomUser
type CustomUserFactory struct{}

// NewCustomUserFactory creates a new CustomUserFactory
func NewCustomUserFactory() *CustomUserFactory {
	return &CustomUserFactory{}
}

// NewUser implements interfaces.UserRepositoryFactory
func (f *CustomUserFactory) NewUser() interfaces.User {
	return &CustomUser{}
}

// GetTableName implements interfaces.UserRepositoryFactory
func (f *CustomUserFactory) GetTableName() string {
	return "users"
}

// GetSchema implements interfaces.UserRepositoryFactory
func (f *CustomUserFactory) GetSchema() any {
	return &CustomUser{}
}
