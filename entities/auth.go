package entities

import (
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

type Auth struct {
	BaseModel

	Identity      string `gorm:"uniqueIndex"`
	UserTableName string
	UserId        uuid.UUID  `gorm:"uniqueIndex"`
	Role          types.Role `json:"role"`
	PasswordHash  string
}
