package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/volvlabs/nebularcore/modules/auth/authorization"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models/requests"
	"github.com/volvlabs/nebularcore/modules/auth/models/responses"
	"github.com/volvlabs/nebularcore/modules/auth/repositories"
	"github.com/volvlabs/nebularcore/tools/handlers"
)

// newRoleResponse builds a responses.RoleResponse from a repository row.
func newRoleResponse(role *repositories.Role) responses.RoleResponse {
	return responses.RoleResponse{
		Name:        role.Name,
		Description: role.Description,
		Metadata:    role.Metadata,
		CreatedAt:   role.CreatedAt,
	}
}

// newResourceResponse builds a responses.ResourceResponse from a repository row.
func newResourceResponse(resource *repositories.Resource) responses.ResourceResponse {
	return responses.ResourceResponse{
		Name:        resource.Name,
		Description: resource.Description,
		Actions:     resource.Actions,
		CreatedAt:   resource.CreatedAt,
	}
}

// AuthorizationHandler exposes generic role/permission management over
// HTTP, riding entirely on AuthorizationManager's casbin-backed methods.
// Only registered when config.AuthorizationConfig.ExposeManagementAPI is
// true (see module.go's Initialize) — this is a privileged CRUD surface,
// not on by default.
type AuthorizationHandler struct {
	authMgr        *authorization.AuthorizationManager
	authMiddleware interfaces.AuthMiddleware
}

// NewAuthorizationHandler creates a new authorization management handler.
func NewAuthorizationHandler(authMgr *authorization.AuthorizationManager, authMiddleware interfaces.AuthMiddleware) *AuthorizationHandler {
	return &AuthorizationHandler{authMgr: authMgr, authMiddleware: authMiddleware}
}

// RegisterRoutes registers the role/permission management routes. Every
// route is gated by requireManagePermission — see its doc comment for how
// the bootstrap chicken-and-egg (who can call this before any permission
// exists) is resolved.
func (h *AuthorizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	gate := h.requireManagePermission()

	roles := router.Group("/auth/roles", h.authMiddleware.RequireAuth(), gate)
	roles.GET("", h.ListRoles)
	roles.POST("", h.CreateRole)
	roles.GET("/:name", h.GetRole)
	roles.PATCH("/:name", h.UpdateRole)
	roles.DELETE("/:name", h.DeleteRole)
	roles.POST("/:name/assignments", h.AssignRole)
	roles.DELETE("/:name/assignments/:userID", h.UnassignRole)
	roles.GET("/:name/permissions", h.GetRolePermissions)
	roles.POST("/:name/permissions", h.GrantPermission)
	roles.DELETE("/:name/permissions", h.RevokePermission)

	router.GET("/auth/permissions", h.authMiddleware.RequireAuth(), gate, h.ListAllPolicies)
	router.GET("/auth/users/:userID/roles", h.authMiddleware.RequireAuth(), gate, h.GetUserRoles)
	router.GET("/auth/users/:userID/permissions/check", h.authMiddleware.RequireAuth(), gate, h.CheckPermission)

	resources := router.Group("/auth/resources", h.authMiddleware.RequireAuth(), gate)
	resources.GET("", h.ListResources)
	resources.POST("", h.CreateResource)
	resources.GET("/:name", h.GetResource)
	resources.PATCH("/:name", h.UpdateResource)
	resources.DELETE("/:name", h.DeleteResource)
}

// requireManagePermission gates every route in this handler on a single
// self-referential casbin permission (resource="roles", action="manage"),
// checked through the same enforcer as everything else here — no special-
// cased bootstrap logic. The first grant of this permission necessarily
// happens out-of-band, at host-app startup (the same way bootstrapping the
// very first super_admin role assignment already has to happen outside
// any HTTP path) — see AuthorizationConfig.ExposeManagementAPI's doc
// comment.
func (h *AuthorizationHandler) requireManagePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			handlers.NewUnauthorizedError(c)
			return
		}
		allowed, err := h.authMgr.HasPermission(c.Request.Context(), user.(interfaces.User).GetID().String(), "roles", "manage")
		if err != nil || !allowed {
			handlers.NewForbiddenError(c)
			return
		}
		c.Next()
	}
}

func (h *AuthorizationHandler) ListRoles(c *gin.Context) {
	roles, err := h.authMgr.ListRoles(c.Request.Context())
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	out := make([]responses.RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = newRoleResponse(r)
	}
	c.JSON(http.StatusOK, responses.RoleListResponse{Roles: out})
}

func (h *AuthorizationHandler) CreateRole(c *gin.Context) {
	var req requests.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	if err := h.authMgr.CreateRole(c.Request.Context(), req.Name, req.Description, req.Metadata); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	role, err := h.authMgr.GetRole(c.Request.Context(), req.Name)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusCreated, newRoleResponse(role))
}

func (h *AuthorizationHandler) GetRole(c *gin.Context) {
	role, err := h.authMgr.GetRole(c.Request.Context(), c.Param("name"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handlers.NewNotFoundError(c)
			return
		}
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, newRoleResponse(role))
}

func (h *AuthorizationHandler) UpdateRole(c *gin.Context) {
	var req requests.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	name := c.Param("name")
	if err := h.authMgr.UpdateRole(c.Request.Context(), name, req.Description, req.Metadata); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	role, err := h.authMgr.GetRole(c.Request.Context(), name)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, newRoleResponse(role))
}

func (h *AuthorizationHandler) DeleteRole(c *gin.Context) {
	if err := h.authMgr.DeleteRole(c.Request.Context(), c.Param("name")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handlers.NewNotFoundError(c)
			return
		}
		handlers.NewInternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthorizationHandler) AssignRole(c *gin.Context) {
	var req requests.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	var duration *time.Duration
	if req.ExpiresAt != nil {
		d := time.Until(*req.ExpiresAt)
		duration = &d
	}
	if err := h.authMgr.AssignRole(c.Request.Context(), req.UserID, c.Param("name"), duration); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthorizationHandler) UnassignRole(c *gin.Context) {
	if err := h.authMgr.UnassignRole(c.Request.Context(), c.Param("userID"), c.Param("name")); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthorizationHandler) GetRolePermissions(c *gin.Context) {
	perms, err := h.authMgr.ListPermissionsForRole(c.Request.Context(), c.Param("name"))
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, responses.PolicyListResponse{Permissions: perms})
}

func (h *AuthorizationHandler) GrantPermission(c *gin.Context) {
	var req requests.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	if err := h.authMgr.GrantPermission(c.Request.Context(), c.Param("name"), req.Resource, req.Action); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthorizationHandler) RevokePermission(c *gin.Context) {
	var req requests.RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	if err := h.authMgr.RevokePermission(c.Request.Context(), c.Param("name"), req.Resource, req.Action); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthorizationHandler) ListAllPolicies(c *gin.Context) {
	policies, err := h.authMgr.ListAllPolicies(c.Request.Context())
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, responses.PolicyListResponse{Permissions: policies})
}

func (h *AuthorizationHandler) GetUserRoles(c *gin.Context) {
	roles, err := h.authMgr.GetUserRoles(c.Request.Context(), c.Param("userID"))
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	out := make([]responses.RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = newRoleResponse(r)
	}
	c.JSON(http.StatusOK, responses.RoleListResponse{Roles: out})
}

func (h *AuthorizationHandler) ListResources(c *gin.Context) {
	resources, err := h.authMgr.ListResources(c.Request.Context())
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	out := make([]responses.ResourceResponse, len(resources))
	for i, r := range resources {
		out[i] = newResourceResponse(r)
	}
	c.JSON(http.StatusOK, responses.ResourceListResponse{Resources: out})
}

func (h *AuthorizationHandler) CreateResource(c *gin.Context) {
	var req requests.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	if err := h.authMgr.CreateResource(c.Request.Context(), req.Name, req.Description, req.Actions); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	resource, err := h.authMgr.GetResource(c.Request.Context(), req.Name)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusCreated, newResourceResponse(resource))
}

func (h *AuthorizationHandler) GetResource(c *gin.Context) {
	resource, err := h.authMgr.GetResource(c.Request.Context(), c.Param("name"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handlers.NewNotFoundError(c)
			return
		}
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, newResourceResponse(resource))
}

func (h *AuthorizationHandler) UpdateResource(c *gin.Context) {
	var req requests.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, err.Error(), nil)
		return
	}
	name := c.Param("name")
	if err := h.authMgr.UpdateResource(c.Request.Context(), name, req.Description, req.Actions); err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	resource, err := h.authMgr.GetResource(c.Request.Context(), name)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, newResourceResponse(resource))
}

func (h *AuthorizationHandler) DeleteResource(c *gin.Context) {
	if err := h.authMgr.DeleteResource(c.Request.Context(), c.Param("name")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handlers.NewNotFoundError(c)
			return
		}
		handlers.NewInternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthorizationHandler) CheckPermission(c *gin.Context) {
	resource := c.Query("resource")
	action := c.Query("action")
	if resource == "" || action == "" {
		handlers.NewBadRequestError(c, "resource and action query params are required", nil)
		return
	}
	allowed, err := h.authMgr.HasPermission(c.Request.Context(), c.Param("userID"), resource, action)
	if err != nil {
		handlers.NewInternalServerError(c)
		return
	}
	c.JSON(http.StatusOK, responses.PermissionCheckResponse{Allowed: allowed})
}
