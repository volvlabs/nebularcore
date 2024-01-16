package models

type Auth struct {
	BaseModel

	Identity     string `gorm:"uniqueIndex"`
	PasswordHash string
}
