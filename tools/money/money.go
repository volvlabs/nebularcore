// Package money provides small, generic minor-unit currency helpers —
// formatting and conversion only, no payment-provider logic (that's
// app-specific and belongs in the host app).
package money

import (
	"fmt"
	"strings"
)

// symbols covers the currencies nebularcore's localization module seeds by
// default (see modules/localization/migrations/000001_init_localization.up.sql).
// An unlisted currency code falls back to "<CODE> <amount>" in FormatAmount
// rather than guessing a symbol.
var symbols = map[string]string{
	"NGN": "₦",   // ₦
	"GHS": "GH₵", // GH₵
	"KES": "KSh",
	"ZAR": "R",
	"EGP": "E£",
	"TZS": "TSh",
	"UGX": "USh",
	"RWF": "FRw",
	"XOF": "CFA",
	"XAF": "FCFA",
	"ETB": "Br",
	"MAD": "DH",
	"DZD": "DA",
	"TND": "DT",
	"ZMW": "ZK",
	"ZWL": "Z$",
	"MZN": "MT",
	"BWP": "P",
	"NAD": "N$",
	"USD": "$",
	"GBP": "£",
}

// MinorUnitsToDecimal converts an integer amount in a currency's minor
// units (e.g. kobo, pesewas, cents) to its major-unit decimal value, given
// that currency's minor-unit factor (e.g. 100 for NGN/GHS/USD, 1000 for
// currencies like TND with three decimal places).
func MinorUnitsToDecimal(amount int64, minorUnitFactor int) float64 {
	if minorUnitFactor <= 0 {
		minorUnitFactor = 100
	}
	return float64(amount) / float64(minorUnitFactor)
}

// FormatAmount renders a minor-units amount as a human-readable string in
// the given currency, e.g. FormatAmount(500000, "NGN", 100) -> "₦5,000.00".
// This is deliberately simple (no locale-specific grouping/decimal
// separators, no negative-amount styling beyond a leading "-") — a host
// app's UI layer should prefer its platform's native locale-aware
// formatter (e.g. Intl.NumberFormat in a webapp) where one is available;
// this exists for contexts without one, such as backend-rendered strings
// (emails, receipts, logs).
func FormatAmount(amount int64, currencyCode string, minorUnitFactor int) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	decimal := MinorUnitsToDecimal(amount, minorUnitFactor)
	decimalPlaces := decimalPlacesFor(minorUnitFactor)
	formatted := groupThousands(fmt.Sprintf("%.*f", decimalPlaces, decimal))

	symbol, ok := symbols[strings.ToUpper(currencyCode)]
	var result string
	if ok {
		result = symbol + formatted
	} else {
		result = strings.ToUpper(currencyCode) + " " + formatted
	}

	if negative {
		return "-" + result
	}
	return result
}

func decimalPlacesFor(minorUnitFactor int) int {
	switch minorUnitFactor {
	case 1:
		return 0
	case 10:
		return 1
	case 1000:
		return 3
	default:
		return 2
	}
}

// groupThousands inserts "," thousands separators into the integer part of
// a decimal string like "5000.00" -> "5,000.00".
func groupThousands(s string) string {
	intPart, fracPart, hasFrac := strings.Cut(s, ".")

	if len(intPart) <= 3 {
		if hasFrac {
			return intPart + "." + fracPart
		}
		return intPart
	}

	var b strings.Builder
	offset := len(intPart) % 3
	if offset == 0 {
		offset = 3
	}
	b.WriteString(intPart[:offset])
	for i := offset; i < len(intPart); i += 3 {
		b.WriteByte(',')
		b.WriteString(intPart[i : i+3])
	}

	if hasFrac {
		return b.String() + "." + fracPart
	}
	return b.String()
}
