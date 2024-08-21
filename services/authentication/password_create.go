package authentication

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/entities"
	"gitlab.com/jideobs/nebularcore/tools/security"
)

func (a *Auth) Create(identity, password, userTableName string, userId uuid.UUID) error {
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		log.Err(err).Msgf("AuthCreate: could not hash password")
		return err
	}

	auth := &entities.Auth{
		Identity:      identity,
		PasswordHash:  hashedPassword,
		UserTableName: userTableName,
		UserId:        userId,
	}

	return a.dao.CreateAuth(auth)
}
