package models

// Country is generic, app-agnostic reference data describing one of the
// countries a nebularcore-based product can operate in: which currency to
// bill in, what timezone to default to, and how to validate/format phone
// numbers there. It deliberately carries no payment-provider concept
// (e.g. Paystack support) — that is app-specific business data and belongs
// in the host app's own tables/config, keyed by Code.
type Country struct {
	// Code is the ISO 3166-1 alpha-2 country code (e.g. "NG", "GH") and the
	// primary key.
	Code string `gorm:"primaryKey;size:2" json:"code"`

	Name string `json:"name"`

	// CurrencyCode is the ISO 4217 currency code (e.g. "NGN", "GHS").
	CurrencyCode string `json:"currencyCode"`

	// CurrencyMinorUnitFactor is the number of minor units per major unit
	// for CurrencyCode (e.g. 100 for NGN/GHS kobo/pesewas), matching the
	// ISO 4217 minor-unit exponent for that currency.
	CurrencyMinorUnitFactor int `json:"currencyMinorUnitFactor"`

	// DefaultTimezone is an IANA timezone name (e.g. "Africa/Lagos") used
	// only as a fallback default — per-entity timezones (set explicitly by
	// the user) always take precedence over this.
	DefaultTimezone string `json:"defaultTimezone"`

	// PhoneDialCode is the international dial prefix (e.g. "+234").
	PhoneDialCode string `json:"phoneDialCode"`

	// PhoneRegion is the region code passed to libphonenumber for parsing
	// and validating phone numbers for this country (usually equal to Code).
	PhoneRegion string `json:"phoneRegion"`

	// IsActive gates whether a host app currently operates in this country.
	// The table is seeded broadly (see migrations); most rows start
	// inactive, and a host app activates only the countries it actually
	// serves — no schema change needed to add a country later.
	IsActive bool `gorm:"default:false" json:"isActive"`
} // @name Country

func (Country) TableName() string {
	return "countries"
}
