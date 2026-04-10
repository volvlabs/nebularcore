package authentication

import (
	"github.com/volvlabs/nebularcore/entities"
	"github.com/volvlabs/nebularcore/tools/auth"
)

func (a *Auth) CreateWithOAuth2(oauth2Request OAuth2Request) (*auth.AuthUser, error) {
	err := a.ValidateOAuth2Request(oauth2Request)
	if err != nil {
		return nil, err
	}

	provider, err := a.getProvider(oauth2Request.Provider)
	if err != nil {
		return nil, err
	}

	authUser, err := a.getAuthUser(provider, oauth2Request.Code)
	if err != nil {
		return nil, err
	}

	err = a.dao.CreateAuth(&entities.Auth{
		Identity:     authUser.Email,
		PasswordHash: "",
	})
	if err != nil {
		return nil, err
	}

	return authUser, nil
}
