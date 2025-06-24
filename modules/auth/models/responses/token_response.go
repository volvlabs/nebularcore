package responses

// TokenResponse represents the response after successful authentication
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	RefreshToken string `json:"refreshToken,omitempty"`
}
