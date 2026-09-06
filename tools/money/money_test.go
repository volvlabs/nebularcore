package money

import "testing"

func TestFormatAmount(t *testing.T) {
	cases := []struct {
		amount   int64
		currency string
		factor   int
		want     string
	}{
		{500000, "NGN", 100, "₦5,000.00"},
		{5000, "GHS", 100, "GH₵50.00"},
		{100, "NGN", 100, "₦1.00"},
		{0, "NGN", 100, "₦0.00"},
		{-500000, "NGN", 100, "-₦5,000.00"},
		{123456789, "NGN", 100, "₦1,234,567.89"},
		{5000, "XYZ", 100, "XYZ 50.00"},
		{50, "TND", 1000, "DT0.050"},
	}

	for _, c := range cases {
		got := FormatAmount(c.amount, c.currency, c.factor)
		if got != c.want {
			t.Errorf("FormatAmount(%d, %q, %d) = %q, want %q", c.amount, c.currency, c.factor, got, c.want)
		}
	}
}

func TestMinorUnitsToDecimal(t *testing.T) {
	if got := MinorUnitsToDecimal(500000, 100); got != 5000 {
		t.Errorf("got %v, want 5000", got)
	}
	if got := MinorUnitsToDecimal(500000, 0); got != 5000 {
		t.Errorf("zero factor should default to 100: got %v, want 5000", got)
	}
}
