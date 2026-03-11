package validation

import "testing"

func TestIPAddress(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"ipv4", "192.168.1.1", true},
		{"ipv6", "::1", true},
		{"ipv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"empty", "", false},
		{"invalid", "not-an-ip", false},
		{"overflow", "999.999.999.999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IPAddress("ip", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("IPAddress(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("IPAddress(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestIPv6(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"loopback", "::1", true},
		{"full", "2001:db8::1", true},
		{"ipv4 rejected", "192.168.1.1", false},
		{"empty", "", false},
		{"invalid", "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IPv6("ip", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("IPv6(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("IPv6(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestPrivateIP(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"10.x", "10.0.0.1", true},
		{"172.16.x", "172.16.0.1", true},
		{"192.168.x", "192.168.1.1", true},
		{"public", "8.8.8.8", false},
		{"invalid", "bad", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrivateIP("ip", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("PrivateIP(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("PrivateIP(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestIPInRange(t *testing.T) {
	tests := []struct {
		name  string
		value string
		cidr  string
		ok    bool
	}{
		{"in range", "192.168.1.5", "192.168.1.0/24", true},
		{"out of range", "10.0.0.1", "192.168.1.0/24", false},
		{"exact", "10.0.0.1", "10.0.0.1/32", true},
		{"invalid ip", "bad", "192.168.1.0/24", false},
		{"invalid cidr", "10.0.0.1", "bad-cidr", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IPInRange("ip", tt.value, tt.cidr)
			if tt.ok && err != nil {
				t.Fatalf("IPInRange(%q, %q) = %v, want nil", tt.value, tt.cidr, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("IPInRange(%q, %q) = nil, want error", tt.value, tt.cidr)
			}
		})
	}
}

func BenchmarkIPAddress(b *testing.B) {
	for b.Loop() {
		IPAddress("ip", "192.168.1.1")
	}
}

func BenchmarkIPInRange(b *testing.B) {
	for b.Loop() {
		IPInRange("ip", "192.168.1.5", "192.168.1.0/24")
	}
}

func FuzzIPAddress(f *testing.F) {
	f.Add("192.168.1.1")
	f.Add("::1")
	f.Add("")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, s string) {
		_ = IPAddress("ip", s)
	})
}
