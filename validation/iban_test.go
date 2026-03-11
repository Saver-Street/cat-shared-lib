package validation

import "testing"

func TestIBAN(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"GB valid", "GB29NWBK60161331926819", true},
		{"DE valid", "DE89370400440532013000", true},
		{"FR valid", "FR7630006000011234567890189", true},
		{"ES valid", "ES9121000418450200051332", true},
		{"NL valid", "NL91ABNA0417164300", true},
		{"IT valid", "IT60X0542811101000000123456", true},
		{"BE valid", "BE68539007547034", true},
		{"CH valid", "CH9300762011623852957", true},
		{"AT valid", "AT611904300234573201", true},
		{"with spaces", "GB29 NWBK 6016 1331 9268 19", true},
		{"lowercase", "gb29nwbk60161331926819", true},
		{"empty", "", false},
		{"too short", "GB29", false},
		{"bad check digits", "GB00NWBK60161331926819", false},
		{"wrong length for country", "GB29NWBK6016133192681", false},
		{"all digits", "1234567890123456", false},
		{"special chars", "GB29!WBK60161331926819", false},
		{"no letter start", "12NWBK60161331926819XX", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IBAN("iban", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("IBAN(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("IBAN(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestIBANMod97(t *testing.T) {
	tests := []struct {
		iban string
		want bool
	}{
		{"GB29NWBK60161331926819", true},
		{"GB00NWBK60161331926819", false},
	}
	for _, tt := range tests {
		t.Run(tt.iban, func(t *testing.T) {
			got := ibanMod97(tt.iban)
			if got != tt.want {
				t.Fatalf("ibanMod97(%q) = %v, want %v", tt.iban, got, tt.want)
			}
		})
	}
}

func BenchmarkIBAN(b *testing.B) {
	for b.Loop() {
		IBAN("iban", "DE89370400440532013000")
	}
}

func BenchmarkIBANWithSpaces(b *testing.B) {
	for b.Loop() {
		IBAN("iban", "GB29 NWBK 6016 1331 9268 19")
	}
}

func FuzzIBAN(f *testing.F) {
	f.Add("GB29NWBK60161331926819")
	f.Add("DE89370400440532013000")
	f.Add("")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, s string) {
		_ = IBAN("iban", s)
	})
}
