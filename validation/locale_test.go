package validation

import (
	"errors"
	"testing"
)

func TestLocaleValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"lang only", "en"},
		{"lang region", "en-US"},
		{"lang 3 letter", "fra"},
		{"lang script region", "zh-Hant-TW"},
		{"lang region numeric", "es-419"},
		{"with variant", "sl-rozaj-biske"},
		{"script only", "sr-Latn"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := Locale("locale", tt.value); err != nil {
				t.Errorf("Locale(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestLocaleInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too short", "e"},
		{"too long primary", "abcde"},
		{"numbers only", "12"},
		{"special chars", "en_US"},
		{"trailing dash", "en-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := Locale("locale", tt.value); err == nil {
				t.Errorf("Locale(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestLocaleFieldName(t *testing.T) {
	t.Parallel()
	err := Locale("lang", "bad!")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("type = %T; want *ValidationError", err)
	}
	if ve.Field != "lang" {
		t.Errorf("Field = %q; want lang", ve.Field)
	}
}

func TestCountryCodeValid(t *testing.T) {
	t.Parallel()
	tests := []string{"US", "DE", "JP", "GB", "BR"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			if err := CountryCode("country", v); err != nil {
				t.Errorf("CountryCode(%q) = %v; want nil", v, err)
			}
		})
	}
}

func TestCountryCodeInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"one letter", "U"},
		{"three letters", "USA"},
		{"numbers", "12"},
		{"lowercase rejected by regex but upper'd", "us"}, // uppercase internally
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CountryCode("country", tt.value)
			// "us" uppercased to "US" should be valid
			if tt.value == "us" {
				if err != nil {
					t.Errorf("CountryCode(%q) = %v; want nil (uppercased)", tt.value, err)
				}
				return
			}
			if err == nil {
				t.Errorf("CountryCode(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestCurrencyCodeValid(t *testing.T) {
	t.Parallel()
	tests := []string{"USD", "EUR", "JPY", "GBP", "BRL"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			if err := CurrencyCode("currency", v); err != nil {
				t.Errorf("CurrencyCode(%q) = %v; want nil", v, err)
			}
		})
	}
}

func TestCurrencyCodeInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"two letters", "US"},
		{"four letters", "USDX"},
		{"numbers", "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := CurrencyCode("currency", tt.value); err == nil {
				t.Errorf("CurrencyCode(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestLanguageCodeValid(t *testing.T) {
	t.Parallel()
	tests := []string{"en", "fr", "de", "fra", "deu"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			if err := LanguageCode("lang", v); err != nil {
				t.Errorf("LanguageCode(%q) = %v; want nil", v, err)
			}
		})
	}
}

func TestLanguageCodeInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"one letter", "e"},
		{"four letters", "abcd"},
		{"numbers", "12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := LanguageCode("lang", tt.value); err == nil {
				t.Errorf("LanguageCode(%q) = nil; want error", tt.value)
			}
		})
	}
}

func BenchmarkLocale(b *testing.B) {
	for range b.N {
		_ = Locale("l", "en-US")
	}
}

func BenchmarkCountryCode(b *testing.B) {
	for range b.N {
		_ = CountryCode("c", "US")
	}
}

func FuzzLocale(f *testing.F) {
	f.Add("en")
	f.Add("en-US")
	f.Add("")
	f.Fuzz(func(t *testing.T, value string) {
		_ = Locale("f", value)
	})
}
