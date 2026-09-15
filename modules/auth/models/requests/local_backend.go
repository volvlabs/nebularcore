package requests

type LocalBackendPayload struct {
	Username    string `json:"username"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
	Password    string `json:"password" validate:"required"`
}
