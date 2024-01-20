package services

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

type Auth struct {
	dao *daos.Dao
}

func NewAuth(dao *daos.Dao) *Auth {
	return &Auth{dao: dao}
}

func (a *Auth) Create(identity, password string) error {
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		log.Err(err).Msgf("AuthCreate: could not hash password")
		return err
	}

	auth := &models.Auth{
		Identity:     identity,
		PasswordHash: hashedPassword,
	}

	return a.dao.CreateAuth(auth)
}

func (a *Auth) ChangePassword(identity, oldPassword, password string) error {
	auth, err := a.dao.FindAuthByIdentity(identity)
	if err != nil {
		return err
	}

	if !security.ValidatePassword(auth.PasswordHash, oldPassword) {
		return &types.UserError{Message: "current password is incorrect"}
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		log.Err(err).Msgf("AuthChangePassword: could not hash password")
		return err
	}

	return a.dao.UpdatePassword(identity, hashedPassword)
}

func (a *Auth) PasswordLogin(identity, password string) error {
	auth, err := a.dao.FindAuthByIdentity(identity)
	if err != nil {
		if err.Error() == "auth not found" {
			return &types.UserError{Message: "invalid login credentials"}
		}
		return err
	}

	if auth == nil || !security.ValidatePassword(auth.PasswordHash, password) {
		return &types.UserError{Message: "invalid login credentials"}
	}

	return nil
}
