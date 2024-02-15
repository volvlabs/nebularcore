package models

type Auth struct {
	BaseModel

	Identity     string `gorm:"uniqueIndex"`
	Role         string `json:"role"`
	PasswordHash string
}
