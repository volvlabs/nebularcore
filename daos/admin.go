package daos

import (
	"errors"
	"strings"

	"gitlab.com/volvlabs/nebularcore/models"
	"gitlab.com/volvlabs/nebularcore/tools/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (d *Dao) SaveAdmin(admin *models.Admin) error {
	if admin.Id == uuid.Nil {
		return &types.UserError{Message: "use create method to add new admins"}
	}
	return d.Save(admin)
}

func (d *Dao) CreateAdmin(admin *models.Admin) error {
	err := d.Save(admin)
	if err != nil {
		if strings.Contains(err.Error(), "admins.email") {
			return &types.UserError{Message: "email already registered"}
		}

		return &types.SystemError{Message: err.Error()}
	}

	return nil
}

func (d *Dao) FindAdminByEmail(email string) (*models.Admin, error) {
	admin := models.Admin{}
	err := d.FindBy(&admin, &models.Admin{Email: email})
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &types.UserError{Message: "admin not found"}
	}

	return &admin, err
}

func (d *Dao) FindAdminById(id uuid.UUID) (*models.Admin, error) {
	admin := &models.Admin{}
	where := &models.Admin{}
	where.Id = id
	err := d.FindBy(admin, where)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &types.UserError{Message: "admin not found"}
	}

	return admin, err
}
