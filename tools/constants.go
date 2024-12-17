package tools

type ContextKey string

const (
	ContextClaimsKey           ContextKey = "claims"
	ContextTenantIdKey         ContextKey = "tenantId"
	ContextTenantSchemaNameKey ContextKey = "tenantSchemaName"
	ContextDBSessionKey        ContextKey = "dbSession"
)
