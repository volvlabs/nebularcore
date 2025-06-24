package models

import (
	"gitlab.com/jideobs/nebularcore/core/model"
)

// Resource represents a resource provided by a module
type Resource struct {
	model.Model
	ModuleName  string                 `json:"moduleName" gorm:"index:idx_module_resources_name"`
	Resource    string                 `json:"resource" gorm:"index:idx_module_resources_resource"`
	Actions     []string               `json:"actions" gorm:"type:jsonb"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}
