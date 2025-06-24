package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/modules/auth/authorization"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"gitlab.com/jideobs/nebularcore/tools/handlers"
)

type AuthorizationManager struct {
	authManager authorization.Manager
}

func NewAuthorizationManager(authManager authorization.Manager) *AuthorizationManager {
	return &AuthorizationManager{
		authManager: authManager,
	}
}

// RegisterRoutes registers the authorization routes
func (h *AuthorizationManager) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/authorization")
	{
		// Role management
		auth.POST("/roles", h.CreateRole)
		auth.POST("/roles/:roleId/users/:userId", h.AssignRole)
		auth.DELETE("/roles/:roleId/users/:userId", h.UnassignRole)
		auth.GET("/users/:userId/roles", h.GetUserRoles)
		auth.GET("/roles/:roleId/permissions", h.GetRolePermissions)

		// Permission management
		auth.POST("/permissions", h.CreatePermission)
		auth.POST("/roles/:roleId/permissions/:permissionId", h.GrantRolePermission)
		auth.DELETE("/roles/:roleId/permissions/:permissionId", h.RevokeRolePermission)
		auth.POST("/users/:userId/permissions/:permissionId", h.GrantUserPermission)
		auth.DELETE("/users/:userId/permissions/:permissionId", h.RevokeUserPermission)
		auth.GET("/users/:userId/permissions", h.GetUserDirectPermissions)

		// Checks
		auth.GET("/users/:userId/hasRole/:roleName", h.HasRole)
		auth.GET("/users/:userId/hasPermission", h.HasPermission)
	}
}

func (h *AuthorizationManager) CreateRole(c *gin.Context) {
	var payload requests.CreateRolePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authManager.CreateRole(c.Request.Context(), &payload); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *AuthorizationManager) CreatePermission(c *gin.Context) {
	var payload requests.CreatePermissionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission, err := h.authManager.CreatePermission(c.Request.Context(), &payload)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, permission)
}

func (h *AuthorizationManager) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	roleName := c.Query("roleName")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleName query parameter is required"})
		return
	}

	if err := h.authManager.AssignRole(c.Request.Context(), userID, roleID, roleName); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthorizationManager) UnassignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	roleName := c.Query("roleName")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleName query parameter is required"})
		return
	}

	if err := h.authManager.UnassignRole(c.Request.Context(), userID, roleID, roleName); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthorizationManager) GrantRolePermission(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("permissionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	grantedByID, err := uuid.Parse(c.Query("grantedBy"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grantedBy query parameter is required"})
		return
	}

	roleName := c.Query("roleName")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleName query parameter is required"})
		return
	}

	if err := h.authManager.GrantRolePermission(c.Request.Context(), roleID, permissionID, grantedByID, roleName); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthorizationManager) RevokeRolePermission(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("permissionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	roleName := c.Query("roleName")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleName query parameter is required"})
		return
	}

	if err := h.authManager.RevokeRolePermission(c.Request.Context(), roleID, permissionID, roleName); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthorizationManager) GrantUserPermission(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("permissionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	grantedByID, err := uuid.Parse(c.Query("grantedBy"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grantedBy query parameter is required"})
		return
	}

	if err := h.authManager.GrantUserPermission(c.Request.Context(), userID, permissionID, grantedByID); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthorizationManager) RevokeUserPermission(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("permissionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	if err := h.authManager.RevokeUserPermission(c.Request.Context(), userID, permissionID); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthorizationManager) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	roles, err := h.authManager.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, roles)
}

func (h *AuthorizationManager) GetUserDirectPermissions(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	permissions, err := h.authManager.GetUserDirectPermissions(c.Request.Context(), userID)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func (h *AuthorizationManager) GetRolePermissions(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	permissions, err := h.authManager.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func (h *AuthorizationManager) HasRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	roleName := c.Param("roleName")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleName parameter is required"})
		return
	}

	hasRole, err := h.authManager.HasRole(c.Request.Context(), userID, roleName)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"hasRole": hasRole})
}

func (h *AuthorizationManager) HasPermission(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	resource := c.Query("resource")
	if resource == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource query parameter is required"})
		return
	}

	action := c.Query("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action query parameter is required"})
		return
	}

	hasPermission, err := h.authManager.HasPermission(c.Request.Context(), userID, resource, action)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"hasPermission": hasPermission})
}
