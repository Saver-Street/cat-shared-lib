package validation

import "testing"

func TestPort(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"min", "1", true},
		{"max", "65535", true},
		{"common", "8080", true},
		{"zero", "0", false},
		{"negative", "-1", false},
		{"too high", "65536", false},
		{"empty", "", false},
		{"letters", "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Port("port", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("Port(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("Port(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestPortInt(t *testing.T) {
	tests := []struct {
		name string
		n    int
		ok   bool
	}{
		{"min", 1, true},
		{"max", 65535, true},
		{"zero", 0, false},
		{"negative", -1, false},
		{"too high", 65536, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PortInt("port", tt.n)
			if tt.ok && err != nil {
				t.Fatalf("PortInt(%d) = %v, want nil", tt.n, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("PortInt(%d) = nil, want error", tt.n)
			}
		})
	}
}

func TestHostPort(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"valid", "example.com:8080", true},
		{"localhost", "localhost:3000", true},
		{"no port", "example.com", false},
		{"no host", ":8080", false},
		{"empty", "", false},
		{"bad port", "example.com:abc", false},
		{"bad host", "-bad:8080", false},
		{"colon only", ":", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HostPort("addr", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("HostPort(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("HostPort(%q) = nil, want error", tt.value)
			}
		})
	}
}

func BenchmarkPort(b *testing.B) {
	for b.Loop() {
		Port("port", "8080")
	}
}

func BenchmarkHostPort(b *testing.B) {
	for b.Loop() {
		HostPort("addr", "example.com:8080")
	}
}

func FuzzPort(f *testing.F) {
	f.Add("8080")
	f.Add("0")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		_ = Port("p", s)
	})
}

func FuzzHostname(f *testing.F) {
	f.Add("example.com")
	f.Add("")
	f.Add("-bad")

	f.Fuzz(func(t *testing.T, s string) {
		_ = Hostname("h", s)
	})
}
