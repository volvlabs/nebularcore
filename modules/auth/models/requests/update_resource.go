package requests

type UpdateResourcePayload struct {
	Resource    string                 `json:"resource" validate:"required"`
	Actions     []string               `json:"actions" validate:"required"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}
