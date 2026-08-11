package requests

import "time"

// CreateRoleRequest is the payload for POST /auth/roles.
type CreateRoleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
} // @name CreateRoleRequest

// UpdateRoleRequest is the payload for PATCH /auth/roles/:name — name
// itself is not updatable, see RoleRepository.UpdateRole's doc comment.
type UpdateRoleRequest struct {
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
} // @name UpdateRoleRequest

// AssignRoleRequest is the payload for POST /auth/roles/:name/assignments.
type AssignRoleRequest struct {
	UserID    string     `json:"userID" binding:"required"`
	ExpiresAt *time.Time `json:"expiresAt"`
} // @name AssignRoleRequest

// GrantPermissionRequest is the payload for POST
// /auth/roles/:name/permissions.
type GrantPermissionRequest struct {
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
} // @name GrantPermissionRequest

// RevokePermissionRequest is the payload for DELETE
// /auth/roles/:name/permissions.
type RevokePermissionRequest struct {
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
} // @name RevokePermissionRequest

// CreateResourceRequest is the payload for POST /auth/resources.
type CreateResourceRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
} // @name CreateResourceRequest

// UpdateResourceRequest is the payload for PATCH /auth/resources/:name —
// name itself is not updatable, see ResourceRepository.UpdateResource's
// doc comment.
type UpdateResourceRequest struct {
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
} // @name UpdateResourceRequest
