package requests

import "github.com/google/uuid"

type CreatePermissionPayload struct {
	Name        string         `json:"name" validate:"required"`
	ResourceID  uuid.UUID      `json:"resourceId" validate:"required"`
	Action      string         `json:"action" validate:"required"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata"`
}
