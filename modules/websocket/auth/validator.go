package auth

import (
	"fmt"
	"net/url"

	"github.com/golang-jwt/jwt/v4"
	"github.com/volvlabs/nebularcore/tools/security"
)

// Claims holds the validated user claims from a JWT.
type Claims struct {
	UserID   string
	TenantID string
	OrgID    string
	Raw      jwt.MapClaims
}

// ValidateToken parses and validates a JWT, extracting userID, tenantID, and orgID.
func ValidateToken(token, secret string) (*Claims, error) {
	mapClaims, err := security.ParseJWT(token, secret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	userID, _ := mapClaims["user_id"].(string)
	if userID == "" {
		// Try "sub" as a fallback.
		userID, _ = mapClaims["sub"].(string)
	}

	tenantID, _ := mapClaims["tenant_id"].(string)

	// org_id is a distinct claim from tenant_id: tenant_id is used by some
	// callers for schema/DB-level multi-tenancy, while org_id identifies the
	// user's organization for application-level authorization (e.g. scoping
	// WebSocket topic subscriptions to one org).
	orgID, _ := mapClaims["org_id"].(string)

	return &Claims{
		UserID:   userID,
		TenantID: tenantID,
		OrgID:    orgID,
		Raw:      mapClaims,
	}, nil
}

// ValidateOrigin checks if a request origin is in the allowed origins list.
// An empty allowlist permits all origins.
func ValidateOrigin(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// ExtractTokenFromQuery extracts a "token" query parameter from a URL.
func ExtractTokenFromQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("token")
}
