package services

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

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
