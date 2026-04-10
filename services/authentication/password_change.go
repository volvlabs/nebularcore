package authentication

import (
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/tools/security"
	"github.com/volvlabs/nebularcore/tools/types"
)

func (a *Auth) ChangePassword(identity, currentPassword, password string) error {
	auth, err := a.dao.FindAuthByIdentity(identity)
	if err != nil {
		return err
	}

	if !security.ValidatePassword(auth.PasswordHash, currentPassword) {
		return &types.UserError{Message: "current password is incorrect"}
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		log.Err(err).Msgf("AuthChangePassword: could not hash password")
		return err
	}

	return a.dao.UpdatePassword(identity, hashedPassword)
}
