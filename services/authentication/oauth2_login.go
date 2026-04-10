package authentication

import "github.com/volvlabs/nebularcore/tools/auth"

func (a *Auth) LoginWithOAuth2(oauth2Request OAuth2Request) (*auth.AuthUser, error) {
	err := a.ValidateOAuth2Request(oauth2Request)
	if err != nil {
		return nil, err
	}

	provider, err := a.getProvider(oauth2Request.Provider)
	if err != nil {
		return nil, err
	}

	return a.getAuthUser(provider, oauth2Request.Code)
}
