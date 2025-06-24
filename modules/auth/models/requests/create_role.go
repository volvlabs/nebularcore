package requests

type CreateRolePayload struct {
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
}
