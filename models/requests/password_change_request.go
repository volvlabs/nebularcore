package requests

type PasswordChangeRequest struct {
	CurrentPassword    string `json:"currentPassword" validate:"required"`
	NewPassword        string `json:"newPassword" validate:"required"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required"`
} // @name PasswordChangeRequest
