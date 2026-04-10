package authentication

import (
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/models/requests"
	"github.com/volvlabs/nebularcore/tools/security"
	"github.com/volvlabs/nebularcore/tools/types"
	"gorm.io/gorm"
)

func (a *Auth) PasswordLogin(loginRequest requests.LoginRequest) (map[string]any, error) {
	identity := loginRequest.Identity
	password := loginRequest.Password
	auth, err := a.dao.FindAuthByIdentity(identity)
	if err != nil {
		if err.Error() == "authentication not found" {
			return nil, &types.UserError{Message: "invalid login credentials"}
		}
		return nil, err
	}

	if auth == nil || !security.ValidatePassword(auth.PasswordHash, password) {
		log.Info().Msgf("Auth: user password invalid, user: %s", identity)
		return nil, &types.UserError{Message: "invalid login credentials"}
	}

	userInfo := map[string]any{}
	err = a.app.Dao().DB().Transaction(func(tx *gorm.DB) error {
		return tx.Table(auth.UserTableName).Where("id = ?", auth.UserId).Scan(&userInfo).Error
	})
	if err != nil {
		return nil, err
	}

	userInfo["role"] = auth.Role
	return userInfo, nil
}
