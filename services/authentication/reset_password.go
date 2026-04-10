package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/entities"
	"github.com/volvlabs/nebularcore/models/requests"
	"github.com/volvlabs/nebularcore/tools/security"
	"gorm.io/gorm"
)

func (a *Auth) InitiateResetPassword(resetPasswordRequest *requests.InitiateResetPasswordPayload) error {
	if err := a.Validate(resetPasswordRequest); err != nil {
		return err
	}

	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		log.Err(err).Msgf("error occurred trying to generate token")
		return err
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return a.app.Dao().DB().Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entities.Auth{}).Where("identity = ?", resetPasswordRequest.Email).
			Updates(map[string]any{
				"reset_password_token":             token,
				"reset_password_token_expiry_date": expiresAt,
			})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			log.Warn().Msgf("unknown user reset password initiated for, email: %s", resetPasswordRequest.Email)
		}
		return nil
	})
}

func (a *Auth) validateToken(token string) error {
	user := &entities.Auth{}
	err := a.app.Dao().DB().Transaction(func(tx *gorm.DB) error {
		return tx.Model(&entities.Auth{}).Select(
			"reset_password_token_expiry_date",
		).Where("reset_password_token = ?", token).
			First(user).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Msgf("could not find user with token: %s", token)
			return ErrInvalidPasswordToken
		}

		log.Err(err).Msgf("error occurred getting user with token: %s", token)
		return err
	}

	if time.Now().After(user.ResetPasswordTokenExpiryDate.Time()) {
		return ErrTokenExpired
	}
	return nil
}

func (a *Auth) ValidateToken(payload *requests.ValidateRequestPasswordTokenPayload) error {
	return a.validateToken(payload.Token)
}

func (a *Auth) ResetPassword(payload *requests.ResetPasswordPayload) error {
	if err := a.validateToken(payload.Token); err != nil {
		return err
	}

	hashedPassword, err := security.HashPassword(
		payload.Password,
	)
	if err != nil {
		log.Err(err).Msgf("AuthService: could not hash password, token: %s", payload.Token)
		return err
	}

	err = a.app.Dao().DB().Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&entities.Auth{}).Where("reset_password_token = ?", payload.Token).
			Updates(map[string]any{
				"reset_password_token":             "",
				"reset_password_token_expiry_date": nil,
				"password_hash":                    hashedPassword,
			}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Err(err).Msgf("AuthService: error occurred changing password, token: %s", payload.Token)
	}

	return err
}
