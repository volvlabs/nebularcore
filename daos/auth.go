package daos

import (
	"errors"
	"gitlab.com/jideobs/nebularcore/entities"
	"strings"

	"gitlab.com/jideobs/nebularcore/tools/types"
	"gorm.io/gorm"
)

func (d *Dao) CreateAuth(auth *entities.Auth) error {
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
	return d.Update(&entities.Auth{}, &entities.Auth{Identity: identity}, "password_hash", newPasswordHash)
}

func (d *Dao) FindAuthByIdentity(identity string) (*entities.Auth, error) {
	auth := &entities.Auth{}
	where := &entities.Auth{}
	where.Identity = identity
	err := d.FindBy(auth, where)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &types.UserError{Message: "authentication not found"}
	}

	return auth, err
}
