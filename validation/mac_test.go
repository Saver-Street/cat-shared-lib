package validation

import "testing"

func TestMACAddress(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"colon upper", "01:23:45:67:89:AB", true},
		{"colon lower", "aa:bb:cc:dd:ee:ff", true},
		{"colon mixed", "Aa:Bb:Cc:Dd:Ee:Ff", true},
		{"hyphen", "01-23-45-67-89-AB", true},
		{"dot", "0123.4567.89AB", true},
		{"all zeros", "00:00:00:00:00:00", true},
		{"broadcast", "FF:FF:FF:FF:FF:FF", true},
		{"empty", "", false},
		{"short", "01:23:45:67:89", false},
		{"long", "01:23:45:67:89:AB:CD", false},
		{"no separator", "0123456789AB", false},
		{"invalid hex", "GG:HH:II:JJ:KK:LL", false},
		{"mixed sep", "01:23-45:67-89:AB", false},
		{"spaces", "01 23 45 67 89 AB", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MACAddress("mac", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("MACAddress(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("MACAddress(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestMACAddressColon(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"valid colon", "01:23:45:67:89:AB", true},
		{"hyphen rejected", "01-23-45-67-89-AB", false},
		{"dot rejected", "0123.4567.89AB", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MACAddressColon("mac", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("MACAddressColon(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("MACAddressColon(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestNormalizeMACAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{"from colon lower", "aa:bb:cc:dd:ee:ff", "AA:BB:CC:DD:EE:FF", true},
		{"from colon upper", "01:23:45:67:89:AB", "01:23:45:67:89:AB", true},
		{"from hyphen", "01-23-45-67-89-ab", "01:23:45:67:89:AB", true},
		{"from dot", "0123.4567.89ab", "01:23:45:67:89:AB", true},
		{"invalid", "not-a-mac", "", false},
		{"empty", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMACAddress(tt.input)
			if tt.ok {
				if err != nil {
					t.Fatalf("NormalizeMACAddress(%q) error: %v", tt.input, err)
				}
				if got != tt.want {
					t.Fatalf("NormalizeMACAddress(%q) = %q, want %q", tt.input, got, tt.want)
				}
			} else {
				if err == nil {
					t.Fatalf("NormalizeMACAddress(%q) = nil error, want error", tt.input)
				}
			}
		})
	}
}

func BenchmarkMACAddress(b *testing.B) {
	for b.Loop() {
		MACAddress("mac", "01:23:45:67:89:AB")
	}
}

func BenchmarkNormalizeMACAddress(b *testing.B) {
	for b.Loop() {
		NormalizeMACAddress("aa-bb-cc-dd-ee-ff")
	}
}

func FuzzMACAddress(f *testing.F) {
	f.Add("01:23:45:67:89:AB")
	f.Add("01-23-45-67-89-AB")
	f.Add("0123.4567.89AB")
	f.Add("")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, s string) {
		_ = MACAddress("mac", s)
	})
}

func FuzzNormalizeMACAddress(f *testing.F) {
	f.Add("01:23:45:67:89:AB")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = NormalizeMACAddress(s)
	})
}
