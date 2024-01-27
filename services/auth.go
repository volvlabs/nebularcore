package services

import (
	"errors"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/tools/auth"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gitlab.com/jideobs/nebularcore/tools/validation"
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

func NewAuth(app core.App) *Auth {
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
