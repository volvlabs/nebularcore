package services

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/tools/security"
)

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
