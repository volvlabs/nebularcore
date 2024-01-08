package models

import (
	"gitlab.com/volvlabs/nebularcore/tools/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model interface {
	HasId() bool
	SetId(id uuid.UUID)
	GetId() uuid.UUID
	BeforeCreate(tx *gorm.DB) error
	BeforeUpdate(tx *gorm.DB) error
}

type BaseModel struct {
	Id      uuid.UUID      `gorm:"default:uuid_generate_v4()" json:"id"`
	Created types.DateTime `json:"created"`
	Updated types.DateTime `json:"updated"`
}

func (m *BaseModel) HasId() bool {
	return m.Id != uuid.Nil
}

func (m *BaseModel) GetId() uuid.UUID {
	return m.Id
}

func (m *BaseModel) SetId(id uuid.UUID) {
	m.Id = id
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	m.Id = uuid.New()
	m.Created = types.NowDateTime()
	return nil
}

func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.Updated = types.NowDateTime()
	return nil
}
