package responses

type FilterResourcePayload struct {
	ModuleName string   `json:"moduleName"`
	Resource   string   `json:"resource"`
	Actions    []string `json:"actions"`
}
