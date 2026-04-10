package factories

import (
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models"
)

type DefaultUserFactory struct {
}

func NewDefaultUserFactory() interfaces.UserRepositoryFactory {
	return &DefaultUserFactory{}
}

func (f *DefaultUserFactory) NewUser() interfaces.User {
	return &models.User{}
}

func (f *DefaultUserFactory) GetTableName() string {
	return "users"
}

func (f *DefaultUserFactory) GetSchema() any {
	return &models.User{}
}
