package backends

import (
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories"
)

// SocialBackend handles social login authentication
type SocialBackend struct {
	userRepo        *repositories.UserRepository
	socialRepo      *repositories.SocialAccountRepository
	config          config.SocialConfig
	tokenIssuer     interfaces.TokenIssuer
	providerClients map[string]SocialProviderClient
}

// NewSocialBackend creates a new social authentication backend
func NewSocialBackend(
	userRepo *repositories.UserRepository,
	socialRepo *repositories.SocialAccountRepository,
	tokenIssuer interfaces.TokenIssuer,
	config config.SocialConfig,
) *SocialBackend {
	backend := &SocialBackend{
		userRepo:        userRepo,
		socialRepo:      socialRepo,
		config:          config,
		tokenIssuer:     tokenIssuer,
		providerClients: make(map[string]SocialProviderClient),
	}

	// Initialize provider clients
	for provider, cfg := range config.Providers {
		var client SocialProviderClient
		switch provider {
		case "google":
			client = NewGoogleClient(cfg)
		case "github":
			client = NewGithubClient(cfg)
		case "facebook":
			client = NewFacebookClient(cfg)
		}
		if client != nil {
			backend.providerClients[provider] = client
		}
	}

	return backend
}

// func (b *SocialBackend) Name() string {
// 	return "social"
// }

// func (b *SocialBackend) Priority() int {
// 	return 3
// }

// func (b *SocialBackend) Authenticate(ctx context.Context, credentials map[string]interface{}) (interfaces.User, error) {
// 	// Extract credentials
// 	provider, ok := credentials["provider"].(string)
// 	if !ok || provider == "" {
// 		return nil, &autherrors.AuthError{
// 			Code:    autherrors.ErrInvalidCredentials,
// 			Message: "provider is required",
// 		}
// 	}

// 	accessToken, ok := credentials["access_token"].(string)
// 	if !ok || accessToken == "" {
// 		return nil, &autherrors.AuthError{
// 			Code:    autherrors.ErrInvalidCredentials,
// 			Message: "access_token is required",
// 		}
// 	}

// 	// Get provider client
// 	client, ok := b.providerClients[provider]
// 	if !ok {
// 		return nil, &autherrors.AuthError{
// 			Code:    autherrors.ErrUnsupportedMethod,
// 			Message: fmt.Sprintf("unsupported provider: %s", provider),
// 		}
// 	}

// 	// Verify token and get user info
// 	providerUser, err := client.GetUserInfo(accessToken)
// 	if err != nil {
// 		return nil, &autherrors.AuthError{
// 			Code:    http.StatusBadRequest,
// 			Message: "invalid or expired social token",
// 		}
// 	}

// 	// Find existing social account
// 	socialAccount, err := b.socialRepo.FindByProvider(ctx, provider, providerUser.ID)
// 	if err == nil {
// 		// Social account exists, get associated user
// 		user, err := b.userRepo.FindByID(ctx, socialAccount.ID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return user, nil
// 	}

// 	// Create new user and social account
// 	userData := map[string]interface{}{
// 		"email":     providerUser.Email,
// 		"full_name": providerUser.Name,
// 		"status":    "active",
// 	}

// 	user, err := b.userRepo.Create(ctx, userData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create social account
// 	socialData := map[string]interface{}{
// 		"user_id":          user.GetID(),
// 		"provider":         provider,
// 		"provider_user_id": providerUser.ID,
// 		"metadata":         providerUser.Raw,
// 	}

// 	_, err = b.socialRepo.Create(ctx, socialData)
// 	if err != nil {
// 		// TODO: Should we delete the user if social account creation fails?
// 		return nil, err
// 	}

// 	return user, nil
// }

// func (b *SocialBackend) ValidateToken(ctx context.Context, tokenString string) (interfaces.User, error) {
// 	return nil, nil
// }
