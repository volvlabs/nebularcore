package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  any    `json:"errors"`
} // @name ApiError

func (e *ApiError) Error() string {
	return e.Message
}

func NewApiError(code int, message string, errors any) *ApiError {
	return &ApiError{
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}

func NewBadRequestError(c *gin.Context, message string, errors any) {
	apiError := NewApiError(http.StatusBadRequest, message, errors)
	c.AbortWithStatusJSON(http.StatusBadRequest, apiError)
}

func NewInternalServerError(c *gin.Context) {
	apiError := NewApiError(http.StatusInternalServerError,
		"internal server error", nil)
	c.AbortWithStatusJSON(http.StatusInternalServerError, apiError)
}

func NewNotFoundError(c *gin.Context) {
	apiError := NewApiError(http.StatusNotFound, "not found", nil)
	c.AbortWithStatusJSON(http.StatusNotFound, apiError)
}

func NewUnauthorizedError(c *gin.Context) {
	apiError := NewApiError(http.StatusUnauthorized, "unauthorized", nil)
	c.AbortWithStatusJSON(http.StatusUnauthorized, apiError)
}

func NewForbiddenError(c *gin.Context) {
	apiError := NewApiError(http.StatusForbidden, "forbidden", nil)
	c.AbortWithStatusJSON(http.StatusForbidden, apiError)
}

func HandleError(c *gin.Context, err error) {
	if errors.Is(err, types.ErrRecordNotFound) {
		NewNotFoundError(c)
	} else if types.ErrIsUserError(err) {
		var errs any = nil
		if appErr, ok := err.(*types.AppError); ok && appErr.Type == types.ErrorTypeValidation {
			errs = appErr.Errors
		}
		NewBadRequestError(c, err.Error(), errs)
	} else {
		NewInternalServerError(c)
	}
}
