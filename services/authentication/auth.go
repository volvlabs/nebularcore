package authentication

import (
	"errors"
	"github.com/volvlabs/nebularcore/entities"

	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/daos"
	"github.com/volvlabs/nebularcore/models/requests"
	"github.com/volvlabs/nebularcore/tools/auth"
	"github.com/volvlabs/nebularcore/tools/security"
	"github.com/volvlabs/nebularcore/tools/types"
	"github.com/volvlabs/nebularcore/tools/validation"
	"gorm.io/gorm"
)

type OAuth2Request struct {
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
	Provider string `json:"provider" validate:"required"`
}

type Auth struct {
	app       core.App
	dao       *daos.Dao
	validator *validation.Validator
}

func New(app core.App) *Auth {
	return &Auth{app: app, dao: app.Dao(), validator: app.Validator()}
}

func (a *Auth) ValidateOAuth2Request(oauth2Request OAuth2Request) error {
	fieldErrs, err := a.validator.Validate(oauth2Request)
	if err != nil {
		return &types.RequestBodyError{
			Message: "error validating request",
			Errors:  fieldErrs,
		}
	}

	providerConfig, ok := a.app.Settings().NamedAuthProviderConfig(oauth2Request.Provider)
	if !ok {
		return errors.New("invalid provider provided")
	}

	if !providerConfig.Enabled {
		return errors.New("provider not enabled")
	}

	return nil
}

func (a *Auth) getProvider(providerName string) (auth.Provider, error) {
	provider, _ := auth.NewProviderByName(providerName)
	providerConfig, _ := a.app.Settings().NamedAuthProviderConfig(providerName)
	err := providerConfig.SetupProvider(provider)
	return provider, err
}

func (a *Auth) getAuthUser(provider auth.Provider, code string) (*auth.AuthUser, error) {
	token, err := provider.FetchToken(code)
	if err != nil {
		return nil, err
	}

	return provider.FetchAuthUser(token)
}

func (a *Auth) Validate(request any) error {
	fieldErrs, err := a.app.Validator().Validate(request)
	if err != nil {
		return &types.RequestBodyError{
			Message: "error validating request body",
			Errors:  fieldErrs,
		}
	}

	return nil
}

func (a *Auth) RefreshToken(request requests.RefreshTokenRequest) (map[string]any, error) {
	if err := a.Validate(request); err != nil {
		return nil, err
	}

	claims, err := security.ParseJWT(request.Token, a.app.Settings().AuthTokenRefreshSecret)
	if err != nil {
		return nil, security.ErrInvalidRefreshToken
	}

	userId := claims["id"].(string)
	authEntity := &entities.Auth{}
	userInfo := map[string]any{}
	err = a.app.Dao().DB().Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&entities.Auth{}).Where("user_id = ?", userId).First(authEntity).Error
		if err != nil {
			return err
		}

		return tx.Table(authEntity.UserTableName).Where("id = ?", userId).Scan(&userInfo).Error
	})
	if err != nil {
		return nil, err
	}

	userInfo["role"] = authEntity.Role
	return userInfo, nil
}
