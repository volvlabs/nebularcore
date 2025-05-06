package backends

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
)

// LocalBackend implements the AuthenticationBackend interface for local authentication
type LocalBackend struct {
	userRepo    interfaces.UserRepository
	tokenIssuer interfaces.TokenIssuer
}

// NewLocalBackend creates a new local authentication backend
func NewLocalBackend(userRepo interfaces.UserRepository, tokenIssuer interfaces.TokenIssuer) *LocalBackend {
	return &LocalBackend{
		userRepo:    userRepo,
		tokenIssuer: tokenIssuer,
	}
}

// Name returns the name of the backend
func (b *LocalBackend) Name() string {
	return "local"
}

// Priority returns the priority of the backend
func (b *LocalBackend) Priority() int {
	return 1 // Highest priority
}

// Supports returns whether the backend supports a given authentication method
func (b *LocalBackend) Supports(method string) bool {
	return method == "username_password" ||
		method == "email_password" ||
		method == "phone_password"
}

// Authenticate authenticates a user with username/email and password
func (b *LocalBackend) Authenticate(ctx context.Context, credentials map[string]any) (interfaces.User, error) {
	localBackendPayload := requests.LocalBackendPayload{}

	if username, ok := credentials["username"].(string); ok {
		localBackendPayload.Username = username
	}
	if email, ok := credentials["email"].(string); ok {
		localBackendPayload.Email = email
	}
	if phoneNumber, ok := credentials["phoneNumber"].(string); ok {
		localBackendPayload.PhoneNumber = phoneNumber
	}
	if password, ok := credentials["password"].(string); ok {
		localBackendPayload.Password = password
	}

	if localBackendPayload.Username == "" && localBackendPayload.Email == "" &&
		localBackendPayload.PhoneNumber == "" && localBackendPayload.Password == "" {
		return nil, nil
	}

	identifierExists := localBackendPayload.Username != "" || localBackendPayload.Email != "" ||
		localBackendPayload.PhoneNumber != ""
	if !identifierExists {
		return nil, fmt.Errorf("username or email is required")
	}
	if localBackendPayload.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	var user interfaces.User
	var err error

	if localBackendPayload.Username != "" {
		user, err = b.userRepo.FindByUsername(ctx, localBackendPayload.Username)
	} else if localBackendPayload.Email != "" {
		user, err = b.userRepo.FindByEmail(ctx, localBackendPayload.Email)
	} else if localBackendPayload.PhoneNumber != "" {
		user, err = b.userRepo.FindByPhoneNumber(ctx, localBackendPayload.PhoneNumber)
	}
	if err != nil {
		return nil, autherrors.ErrInvalidCredentials
	}

	// Check if user is disabled
	if !user.IsActive() {
		return nil, autherrors.ErrUserDisabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(localBackendPayload.Password)); err != nil {
		return nil, autherrors.ErrInvalidCredentials
	}

	return user, nil
}

// ValidateToken validates a JWT token and returns the associated user
func (b *LocalBackend) ValidateToken(ctx context.Context, tokenString string) (interfaces.User, error) {
	// Parse token
	claims, err := b.tokenIssuer.ValidateToken(tokenString)
	if err != nil {
		return nil, autherrors.ErrInvalidOrExpiredToken
	}

	// Extract user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, autherrors.ErrInvalidOrExpiredToken
	}

	// Find user
	user, err := b.userRepo.FindByID(ctx, uuid.Must(uuid.Parse(userID)))
	if err != nil {
		return nil, autherrors.ErrInvalidOrExpiredToken
	}

	// Check if user is disabled
	if !user.IsActive() {
		return nil, autherrors.ErrUserDisabled
	}

	return user, nil
}
