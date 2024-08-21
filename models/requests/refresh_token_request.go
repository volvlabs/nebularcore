package requests

type RefreshTokenRequest struct {
	Token string `json:"refreshToken" validate:"required"`
}
