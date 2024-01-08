package models

type Admin struct {
	BaseModel

	Avatar       string `json:"avatar"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `gorm:"uniqueIndex" json:"email"`
	Role         string `json:"role"`
	TokenKey     string `json:"-"`
	PasswordHash string `json:"-"`
}
