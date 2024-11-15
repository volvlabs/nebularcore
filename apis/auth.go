package apis

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/entities"
	"gitlab.com/jideobs/nebularcore/models/responses"
	"gitlab.com/jideobs/nebularcore/services/authentication"
	"gitlab.com/jideobs/nebularcore/tools/types"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models/requests"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func BindAuthApi(app core.App, rg *gin.RouterGroup) {
	api := AuthApi{app: app}

	subGroup := rg.Group("")
	subGroup.POST("/login", api.Login)
	subGroup.PUT("/reset-password", api.ResetPassword)
	subGroup.PUT("/refresh-token", api.RefreshToken)

	authGroup := subGroup.Group("")
	authGroup.Use(AuthenticateRequestThenLoadAuthContext(app))
	authGroup.PUT("/change-password", api.ChangePassword)
}

type AuthApi struct {
	app core.App
}

func NewAuthApi(app core.App) *AuthApi {
	return &AuthApi{app: app}
}

func (api AuthApi) GetTokenAndRefreshToken(id uuid.UUID, identity string, role types.Role) (string, string, error) {
	var token, refreshToken string
	token, err := security.NewJWT(
		jwt.MapClaims{"id": id, "role": role},
		api.app.Settings().AuthTokenSecret,
		api.app.Settings().AuthTokenDuration,
	)
	if err != nil {
		return token, refreshToken, err
	}
	refreshToken, err = security.NewJWT(
		jwt.MapClaims{"id": id, "identity": identity, "role": role},
		api.app.Settings().AuthTokenRefreshSecret,
		api.app.Settings().AuthRefreshTokenExpiryDuration,
	)
	if err != nil {
		return refreshToken, token, err
	}
	return token, refreshToken, nil
}

// Todo: look for better way to represent user information
func (api *AuthApi) AuthResponseWithUserInfoMap(c *gin.Context, identity string, userInfo map[string]any) {
	token, refreshToken, err := api.GetTokenAndRefreshToken(
		uuid.MustParse(userInfo["id"].(string)), identity, userInfo["role"].(types.Role))
	if err != nil {
		log.Err(err).Msgf("LoginApi: error occurred getting token and refresh token")
		NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"token":        token,
		"user":         userInfo,
		"refreshToken": refreshToken,
	})
}

func (api *AuthApi) AuthResponseWithUserType(c *gin.Context, user entities.User) {
	token, refreshToken, err := api.GetTokenAndRefreshToken(user.GetId(), user.GetEmail(), user.GetRole())
	if err != nil {
		log.Err(err).Msgf("LoginApi: error occurred getting token and refresh token")
		NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"token":        token,
		"user":         user,
		"refreshToken": refreshToken,
	})
}

// Login godoc
// @Summary      Login a user
// @Description  authenticate a user returning auth token, refresh token and user information
// @Tags         auth
// @Accept       json
// @Param 		  request body requests.LoginRequest true "Login request"
// @Produce      json
// @Success      200  {object}  responses.AuthResponse
// @Failure      500  {object}  apis.ApiError
// @Router       /Login [post]
func (api *AuthApi) Login(c *gin.Context) {
	var loginRequest requests.LoginRequest
	if err := c.BindJSON(&loginRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	authService := authentication.New(api.app)
	userInfo, err := authService.PasswordLogin(loginRequest)
	if err != nil {
		HandleError(c, err)
		return
	}

	api.AuthResponseWithUserInfoMap(c, loginRequest.Identity, userInfo)
}

// ChangePassword godoc
// @Summary      change user's password
// @Description  allow a user to change its password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param 		 request body requests.PasswordChangeRequest true "Password Change request"
// @Security 	 BearerAuth
// @Success      200  {object}  responses.ApiResponse
// @Failure      401  {object}  apis.ApiError
// @Failure      500  {object}  apis.ApiError
// @Router       /change-password [put]
func (api *AuthApi) ChangePassword(c *gin.Context) {
	var passwordChangeRequest requests.PasswordChangeRequest
	if err := c.BindJSON(&passwordChangeRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	authService := authentication.New(api.app)
	claims, _ := c.Get("claims")
	identity := claims.(jwt.MapClaims)["identity"]

	err := authService.ChangePassword(
		identity.(string), passwordChangeRequest.CurrentPassword, passwordChangeRequest.NewPassword)
	if err != nil {
		HandleError(c, err)

		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Code: "00", Message: "password changed successfully"})
}

// RefreshToken godoc
// @Summary      refresh an auth token
// @Description  get a new auth token with the refresh token for user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param 		 request body requests.RefreshTokenRequest true "Refresh token request"
// @Security 	 BearerAuth
// @Success      200  {object}  responses.AuthResponse
// @Failure      401  {object}  apis.ApiError
// @Failure      500  {object}  apis.ApiError
// @Router       /refresh-token [put]
func (api *AuthApi) RefreshToken(c *gin.Context) {
	var refreshTokenRequest requests.RefreshTokenRequest
	if err := c.BindJSON(&refreshTokenRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	authService := authentication.New(api.app)
	userInfo, err := authService.RefreshToken(refreshTokenRequest)
	if err != nil {
		if errors.Is(err, security.ErrInvalidRefreshToken) {
			NewBadRequestError(c, "invalid refresh token", nil)
			return
		}

		HandleError(c, err)
		return
	}

	api.AuthResponseWithUserInfoMap(c, userInfo["email"].(string), userInfo)
}

// ResetPassword godoc
// @Summary      initiate password reset process for user
// @Description  start the process of resetting user's password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param 		 request body requests.ResetPasswordRequest true "Reset password request"
// @Security 	 BearerAuth
// @Success      200  {object}  responses.ApiResponse
// @Failure      401  {object}  apis.ApiError
// @Failure      500  {object}  apis.ApiError
// @Router       /reset-password [put]
func (api *AuthApi) ResetPassword(c *gin.Context) {
	var resetPasswordRequest requests.ResetPasswordRequest
	if err := c.BindJSON(&resetPasswordRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	authService := authentication.New(api.app)
	err := authService.ResetPassword(resetPasswordRequest)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Code:    "00",
		Message: "password reset successful",
	})
}
