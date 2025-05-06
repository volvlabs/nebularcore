package model

// TenantBound defines an interface for models that belong to a tenant
type TenantBound interface {
	IsTenantBound() bool
}
