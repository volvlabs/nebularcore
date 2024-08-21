package apis

import (
	"errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/entities"
	"gitlab.com/jideobs/nebularcore/services/authentication"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"net/http"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models/requests"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func BindAuthApi(app core.App, rg *gin.RouterGroup) {
	api := authApi{app: app}

	subGroup := rg.Group("")
	subGroup.POST("/login", api.login)
	subGroup.PUT("/refresh-token", api.refreshToken)

	authGroup := subGroup.Group("")
	authGroup.Use(AuthenticateRequestThenLoadAuthContext(app))
	authGroup.PUT("/change-password", api.changePassword)
}

type authApi struct {
	app core.App
}

func (api authApi) getTokenAndRefreshToken(id uuid.UUID, identity string, role types.Role) (string, string, error) {
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
func (api *authApi) authResponseWithUserInfoMap(c *gin.Context, identity string, userInfo map[string]any) {
	token, refreshToken, err := api.getTokenAndRefreshToken(
		uuid.MustParse(userInfo["id"].(string)), identity, types.StringToRole[userInfo["role"].(string)])
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

func (api *authApi) authResponseWithUserType(c *gin.Context, user entities.User) {
	token, refreshToken, err := api.getTokenAndRefreshToken(user.GetId(), user.GetEmail(), user.GetRole())
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

func (api *authApi) login(c *gin.Context) {
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

	api.authResponseWithUserInfoMap(c, loginRequest.Identity, userInfo)
}

func (api *authApi) changePassword(c *gin.Context) {
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

	c.JSON(http.StatusOK, map[string]string{})
}

func (api *authApi) refreshToken(c *gin.Context) {
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

	api.authResponseWithUserInfoMap(c, userInfo["email"].(string), userInfo)
}
