package requests

// PasswordResetRequest represents a request to reset password
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetVerifyRequest represents a request to verify password reset token
type PasswordResetVerifyRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// PasswordChangeRequest represents a request to change password
type PasswordChangeRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}
