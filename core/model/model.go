package model

import (
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/tools/types"
	"gorm.io/gorm"
)

// TenantBound defines an interface for models that belong to a tenant
type TenantBound interface {
	IsTenantBound() bool
}

type Model struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primarykey;default:gen_random_uuid()"`
	CreatedAt types.DateTime `json:"createdAt" gorm:"type:timestamp with time zone;default:current_timestamp"`
	UpdatedAt types.DateTime `json:"updatedAt" gorm:"type:timestamp with time zone;default:current_timestamp"`
	DeletedAt types.DateTime `json:"deletedAt" gorm:"index type:timestamp with time zone"`
}

func (m *Model) HasId() bool {
	return m.ID != uuid.Nil
}

func (m *Model) GetId() uuid.UUID {
	return m.ID
}

func (m *Model) SetId(id uuid.UUID) {
	m.ID = id
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if !m.HasId() {
		m.ID = uuid.New()
	}
	m.CreatedAt = types.NowDateTime()
	return nil
}

func (m *Model) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = types.NowDateTime()
	return nil
}
