package requests

type InitiateResetPasswordPayload struct {
	Email string `json:"email" validate:"required,email"`
}
