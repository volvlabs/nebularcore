package backends

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/modules/auth/pkg"
	"gitlab.com/jideobs/nebularcore/modules/auth/types"
	"gorm.io/gorm"
)

type GoogleBackend struct {
	socialRepo   interfaces.SocialAccountRepository
	userRepo     interfaces.UserRepository
	tokenIssuer  interfaces.TokenIssuer
	googleSignin interfaces.GoogleSignin
}

func NewGoogleBackend(
	socialRepo interfaces.SocialAccountRepository,
	userRepo interfaces.UserRepository,
	tokenIssuer interfaces.TokenIssuer,
	googleSignin interfaces.GoogleSignin,
) *GoogleBackend {
	return &GoogleBackend{
		socialRepo:   socialRepo,
		userRepo:     userRepo,
		tokenIssuer:  tokenIssuer,
		googleSignin: googleSignin,
	}
}

func (g *GoogleBackend) Name() string {
	return "google"
}

func (g *GoogleBackend) Priority() int {
	return 2
}

func (g *GoogleBackend) Supports(method string) bool {
	return method == "google"
}

func (g *GoogleBackend) Authenticate(
	ctx context.Context,
	credentials map[string]any,
) (interfaces.User, error) {
	code, codeOk := credentials["code"].(string)
	idToken, idTokenOk := credentials["idToken"].(string)
	if !codeOk && !idTokenOk || (code == "" && idToken == "") {
		log.Warn().Msg("GoogleBackend: missing code and idToken in credentials")
		return nil, autherrors.ErrInvalidCredentials
	}

	var googleUser *pkg.GoogleUser
	var err error
	if code != "" {
		googleUser, err = g.googleSignin.Exchange(ctx, code)
		if err != nil {
			log.Debug().Err(err).Msg("GoogleBackend: failed to exchange code for token")
			return nil, err
		}
	} else if idToken != "" {
		googleUser, err = g.googleSignin.VerifyGoogleIDToken(ctx, idToken)
		if err != nil {
			log.Debug().Err(err).Msg("GoogleBackend: failed to exchange code for token")
			return nil, err
		}
	}

	socialAccount, err := g.socialRepo.FindByProvider(
		ctx,
		types.AuthProviderGoogle,
		googleUser.ID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			socialAccount, err = g.LinkSocialAccount(ctx, googleUser)
			if err != nil {
				return nil, err
			}
		} else {
			log.Debug().Err(err).Msg("GoogleBackend: failed to find social account by provider")
			return nil, err
		}
	}

	user := socialAccount.User
	if !user.IsActive() {
		log.Debug().Str("userID", user.ID.String()).Msg("GoogleBackend: user is disabled")
		return nil, autherrors.ErrUserDisabled
	}

	return &user, nil
}

func (g *GoogleBackend) LinkSocialAccount(ctx context.Context, googleUser *pkg.GoogleUser) (*models.SocialAccount, error) {
	user, err := g.userRepo.FindByEmail(ctx, googleUser.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrSocialEmailDoesNotExist
		}

		return nil, err
	}

	socialAccount, err := g.socialRepo.Create(ctx, &models.SocialAccount{
		UserID:         user.GetID(),
		Provider:       types.AuthProviderGoogle,
		ProviderUserID: googleUser.ID,
		Email:          googleUser.Email,
	})
	if err != nil {
		return nil, err
	}

	socialAccount.User = *user.(*models.User)
	return socialAccount, nil
}

func (g *GoogleBackend) ValidateToken(
	ctx context.Context,
	tokenString string,
) (interfaces.User, error) {
	claims, err := g.tokenIssuer.ValidateToken(tokenString)
	if err != nil {
		return nil, autherrors.ErrInvalidOrExpiredToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, autherrors.ErrInvalidOrExpiredToken
	}

	socialAccount, err := g.socialRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, autherrors.ErrInvalidOrExpiredToken
	}

	if !socialAccount.User.IsActive() {
		return nil, autherrors.ErrUserDisabled
	}

	return &socialAccount.User, nil
}
