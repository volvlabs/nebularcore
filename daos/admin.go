package daos

import (
	"errors"
	"gitlab.com/jideobs/nebularcore/entities"
	"strings"

	"gitlab.com/jideobs/nebularcore/tools/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (d *Dao) SaveAdmin(admin *entities.Admin) error {
	if admin.Id == uuid.Nil {
		return &types.UserError{Message: "use create method to add new admins"}
	}
	return d.Save(admin)
}

func (d *Dao) CreateAdmin(admin *entities.Admin) error {
	err := d.Save(admin)
	if err != nil {
		if strings.Contains(err.Error(), "admins.email") {
			return &types.UserError{Message: "email already registered"}
		}

		return &types.SystemError{Message: err.Error()}
	}

	return nil
}

func (d *Dao) FindAdminByEmail(email string) (*entities.Admin, error) {
	admin := entities.Admin{}
	err := d.FindBy(&admin, &entities.Admin{UserBase: entities.UserBase{Email: email}})
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &types.UserError{Message: "admin not found"}
	}

	return &admin, err
}

func (d *Dao) FindAdminById(id uuid.UUID) (*entities.Admin, error) {
	admin := &entities.Admin{}
	where := &entities.Admin{}
	where.Id = id
	err := d.FindBy(admin, where)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &types.UserError{Message: "admin not found"}
	}

	return admin, err
}
