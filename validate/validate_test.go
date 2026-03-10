package validate

import (
	"fmt"
	"testing"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"user@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"user@sub.domain.com", true},
		{"a@b.cc", true},
		{"", false},
		{"user", false},
		{"user@", false},
		{"@domain.com", false},
		{"user@domain", false},
		{"user domain.com", false},
		{"user@.com", false},
		{" user@example.com ", true}, // trimmed
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			if got := Email(tt.input); got != tt.want {
				t.Errorf("Email(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEmail_TooLong(t *testing.T) {
	long := make([]byte, 250)
	for i := range long {
		long[i] = 'a'
	}
	addr := string(long) + "@b.cc"
	if Email(addr) {
		t.Error("expected false for >254 char email")
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550E8400-E29B-41D4-A716-446655440000", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"", false},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716", false},
		{"550e8400e29b41d4a716446655440000", false}, // no dashes
		{"550e8400-e29b-41d4-a716-44665544000g", false},
		{" 550e8400-e29b-41d4-a716-446655440000 ", true}, // trimmed
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			if got := UUID(tt.input); got != tt.want {
				t.Errorf("UUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPhone(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"+12025551234", true},
		{"12025551234", true},
		{"+442071234567", true},
		{"+1 (202) 555-1234", true}, // formatting stripped
		{"+1-202-555-1234", true},
		{"1234567", true},         // 7 digits minimum
		{"123456789012345", true}, // 15 digits max
		{"", false},
		{"+123", false},             // too short
		{"1234567890123456", false}, // too long
		{"abcdefghijk", false},
		{"+1abc2345678", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			if got := Phone(tt.input); got != tt.want {
				t.Errorf("Phone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://example.com/path?q=1", true},
		{"https://sub.domain.co.uk/path", true},
		{"http://localhost:8080", true},
		{"https://example.com:443/api/v1", true},
		{"", false},
		{"example.com", false},
		{"ftp://example.com", false},
		{"javascript:alert(1)", false},
		{"data:text/html,<h1>hi</h1>", false},
		{"/relative/path", false},
		{"://missing-scheme.com", false},
		{" https://example.com ", true}, // trimmed
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			if got := URL(tt.input); got != tt.want {
				t.Errorf("URL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkEmail(b *testing.B) {
	for b.Loop() {
		Email("user@example.com")
	}
}

func BenchmarkUUID(b *testing.B) {
	for b.Loop() {
		UUID("550e8400-e29b-41d4-a716-446655440000")
	}
}

func BenchmarkPhone(b *testing.B) {
	for b.Loop() {
		Phone("+12025551234")
	}
}

func BenchmarkURL(b *testing.B) {
	for b.Loop() {
		URL("https://example.com/path?q=1")
	}
}
