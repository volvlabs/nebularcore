package entities

import (
	"github.com/volvlabs/nebularcore/tools/types"
)

type Admin struct {
	UserBase

	Avatar    string         `json:"avatar"`
	IsDeleted bool           `json:"-"`
	DeletedAt types.DateTime `json:"-"`
}

func (a *Admin) GetType() string {
	return "admin"
}
