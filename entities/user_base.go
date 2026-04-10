package entities

import (
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/tools/types"
)

type User interface {
	GetId() uuid.UUID
	GetAvatar() string
	GetRole() types.Role
	GetEmail() string
	GetFirstName() string
	GetLastName() string
	GetPhoneNumber() string
	GetIsActive() bool
	GetDeletedAt() types.DateTime
}

type UserBase struct {
	BaseModel

	Avatar      string         `json:"avatar"`
	FirstName   string         `json:"firstName"`
	LastName    string         `json:"lastName"`
	Email       string         `json:"email"`
	PhoneNumber string         `json:"phoneNumber"`
	Role        types.Role     `json:"role"`
	IsActive    bool           `json:"isActive"`
	DeletedAt   types.DateTime `json:"deletedAt"`
}

func (u *UserBase) GetAvatar() string {
	return u.Avatar
}

func (u *UserBase) GetFirstName() string {
	return u.FirstName
}

func (u *UserBase) GetLastName() string {
	return u.LastName
}

func (u *UserBase) GetEmail() string {
	return u.Email
}

func (u *UserBase) GetPhoneNumber() string {
	return u.PhoneNumber
}

func (u *UserBase) GetRole() types.Role {
	return u.Role
}

func (u *UserBase) GetIsActive() bool {
	return u.IsActive
}

func (u *UserBase) GetDeletedAt() types.DateTime {
	return u.DeletedAt
}
