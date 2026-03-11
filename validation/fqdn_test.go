package validation

import (
	"strings"
	"testing"
)

func TestFQDN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"simple", "example.com", false},
		{"subdomain", "api.example.com", false},
		{"deep subdomain", "a.b.c.example.com", false},
		{"trailing dot", "example.com.", false},
		{"hyphen label", "my-site.example.com", false},
		{"numeric label", "123.example.com", false},
		{"empty", "", true},
		{"single label", "localhost", true},
		{"leading hyphen", "-example.com", true},
		{"trailing hyphen", "example-.com", true},
		{"double dot", "example..com", true},
		{"dot only", ".", true},
		{"space", "example .com", true},
		{"underscore", "my_site.example.com", true},
		{"numeric TLD", "example.123", true},
		{"too long label", strings.Repeat("a", 64) + ".com", true},
		{"too long total", strings.Repeat("a.", 128) + "com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := FQDN("domain", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("FQDN(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestDataURI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"plain text", "data:text/plain,hello", false},
		{"base64 image", "data:image/png;base64,iVBOR", false},
		{"no media type", "data:,hello", false},
		{"base64 only", "data:base64,abc", false},
		{"html", "data:text/html,<h1>hi</h1>", false},
		{"empty data", "data:text/plain,", false},
		{"empty", "", true},
		{"no data prefix", "text/plain,hello", true},
		{"no comma", "data:text/plain", true},
		{"invalid media", "data:invalid;base64,abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := DataURI("uri", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataURI(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func BenchmarkFQDN(b *testing.B) {
	for range b.N {
		_ = FQDN("domain", "api.example.com")
	}
}

func BenchmarkDataURI(b *testing.B) {
	for range b.N {
		_ = DataURI("uri", "data:image/png;base64,iVBOR")
	}
}

func FuzzFQDN(f *testing.F) {
	f.Add("example.com")
	f.Add("a.b.c.example.com.")
	f.Add("")
	f.Add("localhost")
	f.Fuzz(func(t *testing.T, s string) {
		_ = FQDN("domain", s)
	})
}

func FuzzDataURI(f *testing.F) {
	f.Add("data:text/plain,hello")
	f.Add("data:,")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		_ = DataURI("uri", s)
	})
}
