package authentication

import (
	"errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/entities"
	"gitlab.com/jideobs/nebularcore/models/requests"
	"gorm.io/gorm"
)

func (a *Auth) ResetPassword(resetPasswordRequest requests.ResetPasswordRequest) error {
	if err := a.Validate(resetPasswordRequest); err != nil {
		return err
	}

	user := &entities.Auth{}
	err := a.dao.FindBy(user, &entities.Auth{Identity: resetPasswordRequest.Email})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		log.Err(err).Msgf("PasswordReset: error occurred fetching user: %s", resetPasswordRequest.Email)
		return err
	}

	return nil
}
