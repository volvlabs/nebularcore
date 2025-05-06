package requests

// PasswordResetPayload represents a request to reset password
type PasswordResetPayload struct {
	Email string `json:"email" binding:"required,email"`
} // @name PasswordResetPayload

// PasswordResetVerifyPayload represents a request to verify password reset token
type PasswordResetVerifyPayload struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
} // @name PasswordResetVerifyPayload

// PasswordChangePayload represents a request to change password
type PasswordChangePayload struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
} // @name PasswordChangePayload
