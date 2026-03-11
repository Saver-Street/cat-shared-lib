package validation

import (
	"testing"
)

func TestCreditCard(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"valid visa", "4111111111111111", true},
		{"valid visa with spaces", "4111 1111 1111 1111", true},
		{"valid visa with dashes", "4111-1111-1111-1111", true},
		{"valid mastercard", "5500000000000004", true},
		{"valid amex", "378282246310005", true},
		{"valid discover", "6011111111111117", true},
		{"invalid luhn", "4111111111111112", false},
		{"too short", "41111111", false},
		{"too long", "41111111111111111111", false},
		{"non-digit chars", "4111abcd11111111", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreditCard("card", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("CreditCard(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("CreditCard(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestCreditCardNetwork(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		networks []CardNetwork
		ok       bool
	}{
		{"visa accepted", "4111111111111111", []CardNetwork{CardVisa}, true},
		{"visa not accepted", "4111111111111111", []CardNetwork{CardMastercard}, false},
		{"mastercard accepted", "5500000000000004", []CardNetwork{CardMastercard, CardVisa}, true},
		{"amex accepted", "378282246310005", []CardNetwork{CardAmex}, true},
		{"invalid number", "1234567890123", []CardNetwork{CardVisa}, false},
		{"discover accepted", "6011111111111117", []CardNetwork{CardDiscover}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreditCardNetwork("card", tt.value, tt.networks...)
			if tt.ok && err != nil {
				t.Fatalf("CreditCardNetwork(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("CreditCardNetwork(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestDetectCardNetwork(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   CardNetwork
	}{
		{"visa 16", "4111111111111111", CardVisa},
		{"visa 13", "4111111111111", CardVisa},
		{"visa 19", "4111111111111111111", CardVisa},
		{"mastercard 51", "5100000000000000", CardMastercard},
		{"mastercard 55", "5500000000000004", CardMastercard},
		{"mastercard 2221", "2221000000000000", CardMastercard},
		{"mastercard 2720", "2720000000000000", CardMastercard},
		{"amex 34", "340000000000000", CardAmex},
		{"amex 37", "378282246310005", CardAmex},
		{"discover 6011", "6011111111111117", CardDiscover},
		{"discover 65", "6500000000000000", CardDiscover},
		{"unknown", "9999999999999999", CardUnknown},
		{"empty", "", CardUnknown},
		{"short mastercard prefix", "5", CardUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCardNetwork(tt.number)
			if got != tt.want {
				t.Fatalf("DetectCardNetwork(%q) = %q, want %q", tt.number, got, tt.want)
			}
		})
	}
}

func TestLuhn(t *testing.T) {
	tests := []struct {
		number string
		valid  bool
	}{
		{"4111111111111111", true},
		{"79927398713", true},
		{"79927398710", false},
		{"0", true},
	}
	for _, tt := range tests {
		t.Run(tt.number, func(t *testing.T) {
			got := luhn(tt.number)
			if got != tt.valid {
				t.Fatalf("luhn(%q) = %v, want %v", tt.number, got, tt.valid)
			}
		})
	}
}

func TestStripCardChars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"4111 1111 1111 1111", "4111111111111111"},
		{"4111-1111-1111-1111", "4111111111111111"},
		{"4111111111111111", "4111111111111111"},
		{"  ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripCardChars(tt.input)
			if got != tt.want {
				t.Fatalf("stripCardChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsMastercardPrefix(t *testing.T) {
	tests := []struct {
		number string
		want   bool
	}{
		{"5100000000000000", true},
		{"5500000000000004", true},
		{"2221000000000000", true},
		{"2720000000000000", true},
		{"4111111111111111", false},
		{"5", false},
		{"50", false},
		{"2220000000000000", false},
	}
	for _, tt := range tests {
		t.Run(tt.number, func(t *testing.T) {
			got := isMastercardPrefix(tt.number)
			if got != tt.want {
				t.Fatalf("isMastercardPrefix(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func BenchmarkCreditCard(b *testing.B) {
	for b.Loop() {
		CreditCard("card", "4111111111111111")
	}
}

func BenchmarkDetectCardNetwork(b *testing.B) {
	for b.Loop() {
		DetectCardNetwork("4111111111111111")
	}
}

func BenchmarkLuhn(b *testing.B) {
	for b.Loop() {
		luhn("4111111111111111")
	}
}

func FuzzCreditCard(f *testing.F) {
	f.Add("4111111111111111")
	f.Add("5500000000000004")
	f.Add("378282246310005")
	f.Add("6011111111111117")
	f.Add("")
	f.Add("not-a-card")

	f.Fuzz(func(t *testing.T, s string) {
		// Must not panic.
		_ = CreditCard("card", s)
	})
}

func FuzzDetectCardNetwork(f *testing.F) {
	f.Add("4111111111111111")
	f.Add("5500000000000004")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		net := DetectCardNetwork(s)
		_ = net
	})
}
