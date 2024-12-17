package tools

type ContextKey string

const ContextClaimsKey = "claims"
const (
	ContextTenantIdKey         ContextKey = "tenantId"
	ContextTenantSchemaNameKey ContextKey = "tenantSchemaName"
	ContextDBSessionKey        ContextKey = "dbSession"
)
