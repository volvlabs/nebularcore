package authentication

import "github.com/volvlabs/nebularcore/tools/types"

var (
	ErrInvalidPasswordToken = &types.UserError{
		Message: "invalid password token",
	}
	ErrTokenExpired = &types.UserError{
		Message: "password reset token expired",
	}
)
