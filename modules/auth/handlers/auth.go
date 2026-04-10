package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/models/responses"
	"github.com/volvlabs/nebularcore/modules/auth/backends"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	autherrors "github.com/volvlabs/nebularcore/modules/auth/errors"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models/requests"
	authresponses "github.com/volvlabs/nebularcore/modules/auth/models/responses"
	"github.com/volvlabs/nebularcore/modules/auth/types"
	"github.com/volvlabs/nebularcore/tools/handlers"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	basePath string

	authManager  backends.AuthenticationManager
	tokenIssuer  interfaces.TokenIssuer
	googleSignin interfaces.GoogleSignin
	config       *config.Config
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authManager backends.AuthenticationManager,
	tokenIssuer interfaces.TokenIssuer,
	googleSignin interfaces.GoogleSignin,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		basePath:     "/auth",
		authManager:  authManager,
		tokenIssuer:  tokenIssuer,
		googleSignin: googleSignin,
		config:       config,
	}
}

func validateAndGetProvider(c *gin.Context) (string, error) {
	providerName := c.Param("provider")
	if providerName == "" {
		return providerName, fmt.Errorf("provider is required")
	}

	provider, err := types.ParseAuthProvider(providerName)
	if err != nil {
		return providerName, fmt.Errorf("invalid provider")
	}

	if provider != types.AuthProviderGoogle {
		return providerName, fmt.Errorf("provider not supported")
	}

	return providerName, nil
}

// RegisterRoutes registers the authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, socialSignupHandler gin.HandlerFunc) {
	auth := router.Group(h.basePath)
	{
		if h.config.Social.Enabled {
			auth.GET("/providers/:provider/initiate", h.InitiateSocialLogin)

			h.RegisterSocialSignupRoutes(router, socialSignupHandler)
		}

		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
	}
}

func (h *AuthHandler) RegisterSocialSignupRoutes(router *gin.RouterGroup, handler gin.HandlerFunc) {
	auth := router.Group(h.basePath)
	if handler != nil {
		auth.POST("/providers/:provider/signup", h.SocialSignup(handler))
		auth.POST("/providers/:provider/login", h.SocialSignInOrSignUp)
	}
}

// @ID initiate-social-login
// @Summary      Initiate Social Login
// @Description  Initiates a social login process
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body map[string]any true "Initiate social login request"
// @Success      202  {object}  authresponses.TokenResponse
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/providers/:provider/initiate [post]
func (h *AuthHandler) InitiateSocialLogin(c *gin.Context) {
	providerName := c.Param("provider")
	if providerName == "" {
		handlers.NewBadRequestError(c, "provider is required", nil)
		return
	}

	provider, err := types.ParseAuthProvider(providerName)
	if err != nil {
		handlers.NewBadRequestError(c, "invalid provider", nil)
		return
	}

	if provider != types.AuthProviderGoogle {
		handlers.NewBadRequestError(c, "provider not supported", nil)
		return
	}

	state := uuid.New().String()
	url := h.googleSignin.GetAuthURL(state)

	c.JSON(http.StatusOK, responses.ApiResponsePayload{
		Status:  true,
		Message: "Successfully initiated social login",
		Data:    map[string]any{"url": url},
	})
}

// @ID socialSignup
// @Summary      Signup using an oauth provider e.g Google
// @Description  Create an account via an oauth provider
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body map[string]any true "Login request"
// @Success      202  {object}  authresponses.TokenResponse
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/providers/:provider/signup [post]
func (h *AuthHandler) SocialSignup(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := validateAndGetProvider(c)
		if err != nil {
			handlers.NewBadRequestError(c, err.Error(), nil)
			return
		}
		next(c)
	}
}

// @ID login
// @Summary      Login
// @Description  Logs in a user
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body map[string]any true "Login request"
// @Success      202  {object}  authresponses.TokenResponse
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var credentials map[string]any

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var (
		user interfaces.User
		err  error
	)
	user, err = h.authManager.Authenticate(c.Request.Context(), credentials)
	if err != nil {
		if authErr, ok := err.(*autherrors.AuthError); ok {
			handlers.NewBadRequestError(c, authErr.Message, nil)
			return
		}

		isBadRequestErr := errors.Is(err, autherrors.ErrInvalidCredentials) ||
			errors.Is(err, autherrors.ErrUserDisabled) ||
			errors.Is(err, autherrors.ErrUserNotFound)
		if isBadRequestErr {
			handlers.NewBadRequestError(c, err.Error(), nil)
			return
		}
		handlers.NewInternalServerError(c)
		return
	}

	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	tokenResp, err := h.tokenIssuer.IssueToken(user)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// @ID socialSigninOrSignup
// @Summary      Via an oauth provider login or create an account for a user
// @Description  This would either log a user in or create an account for a user using the provided provider's creds.
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body map[string]any true "Login request"
// @Success      202  {object}  authresponses.TokenResponse
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/providers/:provider/login [post]
func (h *AuthHandler) SocialSignInOrSignUp(c *gin.Context) {
	if _, err := validateAndGetProvider(c); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}

	var credentials map[string]any

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var (
		user interfaces.User
		err  error
	)
	user, err = h.authManager.Authenticate(c.Request.Context(), credentials)
	if err != nil {
		if errors.Is(err, autherrors.ErrSocialEmailDoesNotExist) {
			c.Redirect(http.StatusPermanentRedirect, "/providers/:provider/signup")
			return
		}

		if authErr, ok := err.(*autherrors.AuthError); ok {
			handlers.NewBadRequestError(c, authErr.Message, nil)
			return
		}

		isBadRequestErr := errors.Is(err, autherrors.ErrInvalidCredentials) ||
			errors.Is(err, autherrors.ErrUserDisabled) ||
			errors.Is(err, autherrors.ErrUserNotFound)
		if isBadRequestErr {
			handlers.NewBadRequestError(c, err.Error(), nil)
			return
		}
		handlers.NewInternalServerError(c)
		return
	}

	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	tokenResp, err := h.tokenIssuer.IssueToken(user)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// @ID refresh-token
// @Summary      Refresh Token
// @Description  Refreshes a user's token
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body requests.RefreshTokenPayload true "Refresh token request"
// @Success      202  {object}  authresponses.TokenResponse
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req requests.RefreshTokenPayload

	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, "bad request payload", nil)
		return
	}

	var (
		tokenResp *authresponses.TokenResponse
		err       error
	)
	tokenResp, err = h.tokenIssuer.RefreshToken(req.RefreshToken)
	if err != nil {
		handlers.NewBadRequestError(c, "invalid refresh token", nil)
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// @ID logout
// @Summary      Logout
// @Description  Logs out a user
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.ApiResponsePayload
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Security 	 BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		handlers.NewBadRequestError(c, "missing token", nil)
		return
	}

	// Remove "Bearer " prefix
	token = token[7:]

	if err := h.tokenIssuer.RevokeToken(token); err != nil {
		handlers.NewInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponsePayload{
		Status:  true,
		Message: "Successfully logged out",
	})
}
