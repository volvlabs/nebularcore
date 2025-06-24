package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/jideobs/nebularcore/modules/auth/authorization/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/handlers"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/tools/test"
)

func setupAuthorizationHandler(t *testing.T) (*handlers.AuthorizationManager, *mocks.Manager) {
	manager := new(mocks.Manager)
	handler := handlers.NewAuthorizationManager(manager)
	return handler, manager
}

func TestAuthorizationManager_CreateRole(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:   "create role - success",
			Method: http.MethodPost,
			Url:    "/api/v1/authorization/roles",
			Body: strings.NewReader(`{
				"name": "admin",
				"description": "Administrator role",
				"metadata": {"key": "value"}
			}`),
			ExpectedStatus: http.StatusCreated,
		},
		{
			Name:           "create role - invalid payload",
			Method:         http.MethodPost,
			Url:            "/api/v1/authorization/roles",
			Body:           strings.NewReader(`{`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, s := range tests {
		s := s // capture range variable
		t.Run(s.Name, func(t *testing.T) {
			router := gin.Default()
			handler, manager := setupAuthorizationHandler(t)
			manager.On("CreateRole", mock.Anything, mock.Anything).Return(nil)
			api := router.Group("/api/v1")
			handler.RegisterRoutes(api)

			req := httptest.NewRequest(s.Method, s.Url, s.Body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, s.ExpectedStatus, w.Code)
			for _, expectedContent := range s.ExpectedContent {
				assert.Contains(t, w.Body.String(), expectedContent)
			}
		})
	}
}

func TestAuthorizationManager_CreatePermission(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:   "successful permission creation",
			Url:    "/api/v1/authorization/permissions",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"name": "read",
				"resourceId": "` + uuid.New().String() + `",
				"action": "read",
				"description": "Read permission",
				"metadata": {"key": "value"}
			}`),
			ExpectedStatus:  http.StatusCreated,
			ExpectedContent: []string{"name", "resourceId", "action"},
		},
	}

	for _, s := range tests {
		s := s
		t.Run(s.Name, func(t *testing.T) {
			router := gin.Default()
			handler, manager := setupAuthorizationHandler(t)
			manager.On("CreatePermission", mock.Anything, mock.Anything).Return(&models.Permission{
				Name:   "read",
				Action: "read",
			}, nil)
			api := router.Group("/api/v1")
			handler.RegisterRoutes(api)

			req := httptest.NewRequest(s.Method, s.Url, s.Body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, s.ExpectedStatus, w.Code)
			for _, expectedContent := range s.ExpectedContent {
				assert.Contains(t, w.Body.String(), expectedContent)
			}
		})
	}
}

func TestAuthorizationManager_AssignRole(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()

	tests := []test.ApiScenario{
		{
			Name:           "successful role assignment",
			Url:            "/api/v1/authorization/roles/" + roleID.String() + "/users/" + userID.String() + "?roleName=admin",
			Method:         http.MethodPost,
			ExpectedStatus: 204,
		},
		{
			Name:            "invalid user ID",
			Url:             "/api/v1/authorization/roles/" + roleID.String() + "/users/invalid-uuid?roleName=admin",
			Method:          http.MethodPost,
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"invalid user ID"},
		},
	}

	for _, s := range tests {
		s := s // capture range variable
		t.Run(s.Name, func(t *testing.T) {
			router := gin.Default()
			handler, manager := setupAuthorizationHandler(t)
			if s.ExpectedStatus == http.StatusNoContent {
				manager.On("AssignRole", mock.Anything, userID, roleID, "admin").Return(nil)
			}
			api := router.Group("/api/v1")
			handler.RegisterRoutes(api)

			req := httptest.NewRequest(s.Method, s.Url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, s.ExpectedStatus, w.Code)
			for _, expectedContent := range s.ExpectedContent {
				assert.Contains(t, w.Body.String(), expectedContent)
			}
		})
	}
}

func TestAuthorizationManager_HasPermission(t *testing.T) {
	userID := uuid.New()

	tests := []test.ApiScenario{
		{
			Name:            "has permission - success",
			Method:          http.MethodGet,
			Url:             fmt.Sprintf("/api/v1/authorization/users/%s/hasPermission?resource=posts&action=read", userID),
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"hasPermission":true`},
		},
		{
			Name:            "has permission - missing parameters",
			Method:          http.MethodGet,
			Url:             fmt.Sprintf("/api/v1/authorization/users/%s/hasPermission", userID),
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"resource query parameter is required"},
		},
	}

	for _, s := range tests {
		s := s // capture range variable
		t.Run(s.Name, func(t *testing.T) {
			router := gin.Default()
			handler, manager := setupAuthorizationHandler(t)
			if s.ExpectedStatus == http.StatusOK {
				manager.On("HasPermission", mock.Anything, userID, "posts", "read").Return(true, nil)
			}
			api := router.Group("/api/v1")
			handler.RegisterRoutes(api)

			req := httptest.NewRequest(s.Method, s.Url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, s.ExpectedStatus, w.Code)
			for _, expectedContent := range s.ExpectedContent {
				assert.Contains(t, w.Body.String(), expectedContent)
			}
		})
	}
}

func TestAuthorizationManager_GetUserRoles(t *testing.T) {
	userID := uuid.New()

	tests := []test.ApiScenario{
		{
			Name:            "get user roles - success",
			Method:          http.MethodGet,
			Url:             fmt.Sprintf("/api/v1/authorization/users/%s/roles", userID),
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"name":"admin"`, `"name":"user"`},
		},
		{
			Name:            "get user roles - invalid user ID",
			Method:          http.MethodGet,
			Url:             "/api/v1/authorization/users/invalid-uuid/roles",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"invalid user ID"},
		},
	}

	for _, s := range tests {
		s := s // capture range variable
		t.Run(s.Name, func(t *testing.T) {
			router := gin.Default()
			handler, manager := setupAuthorizationHandler(t)
			if s.ExpectedStatus == http.StatusOK {
				manager.On("GetUserRoles", mock.Anything, userID).Return([]*models.Role{
					{Name: "admin"},
					{Name: "user"},
				}, nil)
			}
			api := router.Group("/api/v1")
			handler.RegisterRoutes(api)

			req := httptest.NewRequest(s.Method, s.Url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, s.ExpectedStatus, w.Code)
			for _, expectedContent := range s.ExpectedContent {
				assert.Contains(t, w.Body.String(), expectedContent)
			}
		})
	}
}
