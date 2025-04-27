package responses

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         any    `json:"user"`
} // @name AuthResponse
