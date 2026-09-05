package geoip

import "context"

// CloudflareHeaderResolver trusts a header injected by Cloudflare's edge
// (or a compatible upstream proxy) instead of doing its own IP lookup.
// Only correct to use when the app's traffic genuinely all passes through
// Cloudflare (or something else that sets and can be trusted to set this
// header) — otherwise a client can spoof it.
type CloudflareHeaderResolver struct {
	headerName string
}

func NewCloudflareHeaderResolver(headerName string) *CloudflareHeaderResolver {
	if headerName == "" {
		headerName = "CF-IPCountry"
	}
	return &CloudflareHeaderResolver{headerName: headerName}
}

// HeaderName is the header this resolver reads — callers extract it from
// the incoming request and pass the value to Lookup as ip (despite the
// parameter name, this resolver ignores the actual IP and only cares about
// the pre-resolved header value the caller hands it).
func (r *CloudflareHeaderResolver) HeaderName() string {
	return r.headerName
}

func (r *CloudflareHeaderResolver) Lookup(_ context.Context, headerValue string) (string, bool) {
	if headerValue == "" || headerValue == "XX" { // Cloudflare uses "XX" for unknown
		return "", false
	}
	return headerValue, true
}
