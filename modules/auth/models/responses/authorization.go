package responses

import "time"

// RoleResponse mirrors repositories.Role's exported fields for the generic
// role-management API. Deliberately not constructed from a
// *repositories.Role here (see handlers.NewRoleResponse instead) — this
// package sits downstream of modules/auth/repositories in the import graph
// (via interfaces -> models/responses), so importing repositories back
// from here would be a cycle.
type RoleResponse struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"createdAt"`
} // @name RoleResponse

// RoleListResponse is the response for GET /auth/roles and GET
// /auth/users/:userID/roles.
type RoleListResponse struct {
	Roles []RoleResponse `json:"roles"`
} // @name RoleListResponse

// PolicyListResponse is the response for GET /auth/roles/:name/permissions
// and GET /auth/permissions — each entry is a [role, resource, action]
// casbin policy line.
type PolicyListResponse struct {
	Permissions [][]string `json:"permissions"`
} // @name PolicyListResponse

// PermissionCheckResponse is the response for
// GET /auth/users/:userID/permissions/check.
type PermissionCheckResponse struct {
	Allowed bool `json:"allowed"`
} // @name PermissionCheckResponse

// ResourceResponse mirrors repositories.Resource's exported fields for the
// generic resource-management API. Not constructed from a
// *repositories.Resource here — same import-cycle reasoning as
// RoleResponse above (see handlers.newResourceResponse instead).
type ResourceResponse struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Actions     []string  `json:"actions"`
	CreatedAt   time.Time `json:"createdAt"`
} // @name ResourceResponse

// ResourceListResponse is the response for GET /auth/resources.
type ResourceListResponse struct {
	Resources []ResourceResponse `json:"resources"`
} // @name ResourceListResponse
