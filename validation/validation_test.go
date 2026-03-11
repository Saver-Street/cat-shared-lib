package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestEmail_Valid(t *testing.T) {
	valid := []string{
		"user@example.com",
		"alice.bob@company.co.uk",
		"test+tag@gmail.com",
		"a@b.co",
		"user123@test-domain.org",
	}
	for _, e := range valid {
		if err := Email("email", e); err != nil {
			t.Errorf("Email(%q) = %v, want nil", e, err)
		}
	}
}

func TestEmail_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"notanemail",
		"@example.com",
		"user@",
		"user@.com",
		"user@com",
		"user@exam ple.com",
		strings.Repeat("a", 250) + "@b.co", // > 254 chars
	}
	for _, e := range invalid {
		err := Email("email", e)
		if err == nil {
			t.Errorf("Email(%q) = nil, want error", e)
			continue
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("Email(%q) error type = %T, want *ValidationError", e, err)
		}
	}
}

func TestUUID_Valid(t *testing.T) {
	valid := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"F47AC10B-58CC-4372-A567-0E02B2C3D479",
	}
	for _, u := range valid {
		if err := UUID("id", u); err != nil {
			t.Errorf("UUID(%q) = %v, want nil", u, err)
		}
	}
}

func TestUUID_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"not-a-uuid",
		"550e8400e29b41d4a716446655440000",      // no dashes
		"550e8400-e29b-41d4-a716-44665544000",   // too short
		"550e8400-e29b-41d4-a716-4466554400000", // too long
		"gggggggg-gggg-gggg-gggg-gggggggggggg",  // invalid hex
	}
	for _, u := range invalid {
		err := UUID("id", u)
		if err == nil {
			t.Errorf("UUID(%q) = nil, want error", u)
			continue
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("UUID(%q) error type = %T, want *ValidationError", u, err)
		}
	}
}

func TestPhone_Valid(t *testing.T) {
	valid := []string{
		"+14155551234",
		"4155551234",
		"+1-415-555-1234",
		"+44 20 7946 0958",
		"123.456.7890",
		"1234567",
	}
	for _, p := range valid {
		if err := Phone("phone", p); err != nil {
			t.Errorf("Phone(%q) = %v, want nil", p, err)
		}
	}
}

func TestPhone_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"123",                  // too few digits
		"abc",                  // not digits
		"12345678901234567890", // too many digits
		"(123)4567890",         // enough digits but invalid format (parentheses)
	}
	for _, p := range invalid {
		err := Phone("phone", p)
		if err == nil {
			t.Errorf("Phone(%q) = nil, want error", p)
			continue
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("Phone(%q) error type = %T, want *ValidationError", p, err)
		}
	}
}

func TestURL_Valid(t *testing.T) {
	valid := []string{
		"https://example.com",
		"http://localhost:8080/path",
		"https://sub.domain.com/path?q=1",
		"http://192.168.1.1",
	}
	for _, u := range valid {
		if err := URL("website", u); err != nil {
			t.Errorf("URL(%q) = %v, want nil", u, err)
		}
	}
}

func TestURL_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"ftp://example.com",
		"example.com",
		"://missing-scheme",
		"https://",
	}
	for _, u := range invalid {
		err := URL("website", u)
		if err == nil {
			t.Errorf("URL(%q) = nil, want error", u)
			continue
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("URL(%q) error type = %T, want *ValidationError", u, err)
		}
	}
}

func TestRequired(t *testing.T) {
	if err := Required("name", "hello"); err != nil {
		t.Errorf("Required(hello) = %v, want nil", err)
	}
	if err := Required("name", ""); err == nil {
		t.Error("Required('') = nil, want error")
	}
	if err := Required("name", "   "); err == nil {
		t.Error("Required('   ') = nil, want error")
	}
}

func TestMinLength(t *testing.T) {
	if err := MinLength("name", "hello", 3); err != nil {
		t.Errorf("MinLength(hello, 3) = %v, want nil", err)
	}
	if err := MinLength("name", "hi", 3); err == nil {
		t.Error("MinLength(hi, 3) = nil, want error")
	}
	if err := MinLength("name", "   hi   ", 3); err == nil {
		t.Error("MinLength trimmed should fail for 2 chars min 3")
	}
}

func TestMaxLength(t *testing.T) {
	if err := MaxLength("name", "hi", 5); err != nil {
		t.Errorf("MaxLength(hi, 5) = %v, want nil", err)
	}
	if err := MaxLength("name", "toolongvalue", 5); err == nil {
		t.Error("MaxLength(toolongvalue, 5) = nil, want error")
	}
}

func TestOneOf(t *testing.T) {
	allowed := []string{"admin", "user", "guest"}
	if err := OneOf("role", "admin", allowed); err != nil {
		t.Errorf("OneOf(admin) = %v, want nil", err)
	}
	if err := OneOf("role", "superadmin", allowed); err == nil {
		t.Error("OneOf(superadmin) = nil, want error")
	}
}

func TestCollect_AllPass(t *testing.T) {
	errs := Collect(nil, nil, nil)
	if errs != nil {
		t.Errorf("Collect all nil = %v, want nil", errs)
	}
}

func TestCollect_SomeFail(t *testing.T) {
	errs := Collect(
		nil,
		Email("email", "bad"),
		nil,
		Required("name", ""),
	)
	if len(errs) != 2 {
		t.Errorf("Collect = %d errors, want 2", len(errs))
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{Field: "email", Message: "invalid email format"}
	got := ve.Error()
	want := "email: invalid email format"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestEmail_TrimmedInput(t *testing.T) {
	if err := Email("email", "  user@example.com  "); err != nil {
		t.Errorf("Email with whitespace should pass after trim: %v", err)
	}
}

func TestUUID_TrimmedInput(t *testing.T) {
	if err := UUID("id", "  550e8400-e29b-41d4-a716-446655440000  "); err != nil {
		t.Errorf("UUID with whitespace should pass after trim: %v", err)
	}
}

func TestPhone_TrimmedInput(t *testing.T) {
	if err := Phone("phone", "  +14155551234  "); err != nil {
		t.Errorf("Phone with whitespace should pass after trim: %v", err)
	}
}

func TestURL_TrimmedInput(t *testing.T) {
	if err := URL("url", "  https://example.com  "); err != nil {
		t.Errorf("URL with whitespace should pass after trim: %v", err)
	}
}

func TestURL_BadParse(t *testing.T) {
	err := URL("url", "://")
	if err == nil {
		t.Error("URL(://) should fail")
	}
}
