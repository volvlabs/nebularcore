package requests

type RefreshTokenPayload struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
} // @name RefreshTokenPayload
