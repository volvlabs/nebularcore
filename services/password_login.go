package services

import (
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

func (a *Auth) PasswordLogin(identity, password string) error {
	auth, err := a.dao.FindAuthByIdentity(identity)
	if err != nil {
		if err.Error() == "auth not found" {
			return &types.UserError{Message: "invalid login credentials"}
		}
		return err
	}

	if auth == nil || !security.ValidatePassword(auth.PasswordHash, password) {
		return &types.UserError{Message: "invalid login credentials"}
	}

	return nil
}
