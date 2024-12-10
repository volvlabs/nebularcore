package authentication

import "gitlab.com/jideobs/nebularcore/tools/types"

var (
	ErrInvalidPasswordToken = &types.UserError{
		Message: "invalid password token",
	}
	ErrTokenExpired = &types.UserError{
		Message: "password reset token expired",
	}
)
