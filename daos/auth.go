package daos

import (
	"errors"
	"strings"

	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gorm.io/gorm"
)

func (d *Dao) CreateAuth(auth *models.Auth) error {
	err := d.Save(auth)
	if err != nil {
		if strings.Contains(err.Error(), "auths.identity") {
			return &types.UserError{Message: "identity already created"}
		}

		return err
	}

	return nil
}

func (d *Dao) UpdatePassword(identity, newPasswordHash string) error {
	return d.Update(&models.Auth{}, &models.Auth{Identity: identity}, "password_hash", newPasswordHash)
}

func (d *Dao) FindAuthByIdentity(identity string) (*models.Auth, error) {
	auth := &models.Auth{}
	where := &models.Auth{}
	where.Identity = identity
	err := d.FindBy(auth, where)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &types.UserError{Message: "auth not found"}
	}

	return auth, err
}
