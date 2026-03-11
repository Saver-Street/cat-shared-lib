package validation

import "testing"

func FuzzEmail(f *testing.F) {
	f.Add("user@example.com")
	f.Add("bad-email")
	f.Add("")
	f.Add("a@b.c")
	f.Add("user+tag@sub.domain.co.uk")
	f.Add("@missing-local.com")
	f.Add("missing-domain@")
	f.Add("user@.leading-dot.com")
	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic on any input.
		Email("email", input)
	})
}

func FuzzUUID(f *testing.F) {
	f.Add("550e8400-e29b-41d4-a716-446655440000")
	f.Add("not-a-uuid")
	f.Add("")
	f.Add("550E8400-E29B-41D4-A716-446655440000")
	f.Add("550e8400e29b41d4a716446655440000")
	f.Fuzz(func(t *testing.T, input string) {
		UUID("id", input)
	})
}

func FuzzPhone(f *testing.F) {
	f.Add("+1-555-123-4567")
	f.Add("555-123-4567")
	f.Add("")
	f.Add("+44 20 7946 0958")
	f.Add("(555) 123-4567")
	f.Add("123")
	f.Fuzz(func(t *testing.T, input string) {
		Phone("phone", input)
	})
}

func FuzzURL(f *testing.F) {
	f.Add("https://example.com")
	f.Add("http://localhost:8080/path?q=1")
	f.Add("not-a-url")
	f.Add("")
	f.Add("ftp://files.example.com/pub")
	f.Fuzz(func(t *testing.T, input string) {
		URL("url", input)
	})
}

func FuzzMinLength(f *testing.F) {
	f.Add("hello", 3)
	f.Add("", 1)
	f.Add("ab", 5)
	f.Fuzz(func(t *testing.T, input string, min int) {
		MinLength("field", input, min)
	})
}

func FuzzMaxLength(f *testing.F) {
	f.Add("hello", 10)
	f.Add("toolong", 3)
	f.Add("", 0)
	f.Fuzz(func(t *testing.T, input string, max int) {
		MaxLength("field", input, max)
	})
}
