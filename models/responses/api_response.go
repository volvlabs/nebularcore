package responses

type ApiResponsePayload struct {
	Status  bool     `json:"status"`
	Message string   `json:"message"`
	Data    any      `json:"data,omitempty"`
	Errors  []string `json:"errors,omitempty"`
} // @name ApiResponsePayload
