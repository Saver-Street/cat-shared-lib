package validation

import (
	"errors"
	"testing"
)

func TestPortNumberValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"min", "1"},
		{"common", "80"},
		{"https", "443"},
		{"high", "8080"},
		{"max", "65535"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := PortNumber("port", tt.value); err != nil {
				t.Errorf("PortNumber(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestPortNumberInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"zero", "0"},
		{"negative", "-1"},
		{"too high", "65536"},
		{"not a number", "abc"},
		{"float", "80.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := PortNumber("port", tt.value); err == nil {
				t.Errorf("PortNumber(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestPortNumberFieldName(t *testing.T) {
	t.Parallel()
	err := PortNumber("server_port", "")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("type = %T; want *ValidationError", err)
	}
	if ve.Field != "server_port" {
		t.Errorf("Field = %q; want server_port", ve.Field)
	}
}

func TestHostPortValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"hostname", "localhost:8080"},
		{"ipv4", "127.0.0.1:443"},
		{"ipv6", "[::1]:80"},
		{"domain", "example.com:3000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := HostPort("addr", tt.value); err != nil {
				t.Errorf("HostPort(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestHostPortInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"no port", "localhost"},
		{"empty host", ":8080"},
		{"bad port", "localhost:abc"},
		{"port zero", "localhost:0"},
		{"port too high", "localhost:99999"},
		{"just colon", ":"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := HostPort("addr", tt.value); err == nil {
				t.Errorf("HostPort(%q) = nil; want error", tt.value)
			}
		})
	}
}

func BenchmarkPortNumber(b *testing.B) {
	for range b.N {
		_ = PortNumber("p", "8080")
	}
}

func BenchmarkHostPort(b *testing.B) {
	for range b.N {
		_ = HostPort("a", "localhost:8080")
	}
}

func FuzzPortNumber(f *testing.F) {
	f.Add("80")
	f.Add("0")
	f.Add("")
	f.Fuzz(func(t *testing.T, value string) {
		_ = PortNumber("p", value)
	})
}
