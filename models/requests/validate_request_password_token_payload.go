package requests

type ValidateRequestPasswordTokenPayload struct {
	Token string `json:"token" validate:"required"`
} // @name ValidateRequestPasswordTokenPayload
