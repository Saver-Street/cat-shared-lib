package validation

import (
	"testing"
)

func TestPort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"min", "1", false},
		{"max", "65535", false},
		{"http", "80", false},
		{"https", "443", false},
		{"high port", "8080", false},
		{"zero", "0", true},
		{"negative", "-1", true},
		{"too high", "65536", true},
		{"not a number", "abc", true},
		{"empty", "", true},
		{"float", "80.5", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Port("port", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Port(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestPortInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"min", 1, false},
		{"max", 65535, false},
		{"common", 3000, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too high", 65536, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := PortInt("port", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("PortInt(%d) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestHostPort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"domain and port", "example.com:8080", false},
		{"localhost", "localhost:3000", false},
		{"subdomain", "api.example.com:443", false},
		{"empty", "", true},
		{"no port", "example.com", true},
		{"no host", ":8080", true},
		{"bad port", "example.com:abc", true},
		{"port too high", "example.com:99999", true},
		{"only colon", ":", true},
		{"trailing colon", "example.com:", true},
		{"invalid hostname", "-invalid:8080", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := HostPort("addr", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostPort(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func BenchmarkPort(b *testing.B) {
	for range b.N {
		_ = Port("port", "8080")
	}
}

func BenchmarkHostPort(b *testing.B) {
	for range b.N {
		_ = HostPort("addr", "example.com:8080")
	}
}

func FuzzPort(f *testing.F) {
	f.Add("80")
	f.Add("65535")
	f.Add("")
	f.Add("abc")
	f.Fuzz(func(t *testing.T, s string) {
		_ = Port("port", s)
	})
}
