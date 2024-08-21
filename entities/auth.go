package entities

import "github.com/google/uuid"

type Auth struct {
	BaseModel

	Identity      string `gorm:"uniqueIndex"`
	UserTableName string
	UserId        uuid.UUID `gorm:"uniqueIndex"`
	Role          string    `json:"role"`
	PasswordHash  string
}
