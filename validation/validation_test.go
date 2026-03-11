package validation

import (
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
		testkit.AssertNoError(t, Email("email", e))
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
		testkit.AssertErrorAs(t, err, &ve)
	}
}

func TestUUID_Valid(t *testing.T) {
	valid := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"F47AC10B-58CC-4372-A567-0E02B2C3D479",
	}
	for _, u := range valid {
		testkit.AssertNoError(t, UUID("id", u))
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
		testkit.AssertErrorAs(t, err, &ve)
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
		testkit.AssertNoError(t, Phone("phone", p))
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
		testkit.AssertErrorAs(t, err, &ve)
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
		testkit.AssertNoError(t, URL("website", u))
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
		testkit.AssertErrorAs(t, err, &ve)
	}
}

func TestRequired(t *testing.T) {
	testkit.AssertNoError(t, Required("name", "hello"))
	testkit.AssertTrue(t, Required("name", "") != nil)
	testkit.AssertTrue(t, Required("name", "   ") != nil)
}

func TestMinLength(t *testing.T) {
	testkit.AssertNoError(t, MinLength("name", "hello", 3))
	testkit.AssertTrue(t, MinLength("name", "hi", 3) != nil)
	testkit.AssertTrue(t, MinLength("name", "   hi   ", 3) != nil)
}

func TestMaxLength(t *testing.T) {
	testkit.AssertNoError(t, MaxLength("name", "hi", 5))
	testkit.AssertTrue(t, MaxLength("name", "toolongvalue", 5) != nil)
}

func TestOneOf(t *testing.T) {
	allowed := []string{"admin", "user", "guest"}
	testkit.AssertNoError(t, OneOf("role", "admin", allowed))
	testkit.AssertTrue(t, OneOf("role", "superadmin", allowed) != nil)
}

func TestCollect_AllPass(t *testing.T) {
	errs := Collect(nil, nil, nil)
	testkit.AssertNil(t, errs)
}

func TestCollect_SomeFail(t *testing.T) {
	errs := Collect(
		nil,
		Email("email", "bad"),
		nil,
		Required("name", ""),
	)
	testkit.AssertLen(t, errs, 2)
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{Field: "email", Message: "invalid email format"}
	got := ve.Error()
	want := "email: invalid email format"
	testkit.AssertEqual(t, got, want)
}

func TestEmail_TrimmedInput(t *testing.T) {
	testkit.AssertNoError(t, Email("email", "  user@example.com  "))
}

func TestUUID_TrimmedInput(t *testing.T) {
	testkit.AssertNoError(t, UUID("id", "  550e8400-e29b-41d4-a716-446655440000  "))
}

func TestPhone_TrimmedInput(t *testing.T) {
	testkit.AssertNoError(t, Phone("phone", "  +14155551234  "))
}

func TestURL_TrimmedInput(t *testing.T) {
	testkit.AssertNoError(t, URL("url", "  https://example.com  "))
}

func TestURL_BadParse(t *testing.T) {
	err := URL("url", "://")
	testkit.AssertTrue(t, err != nil)
}

func TestIntRange_Valid(t *testing.T) {
	testkit.AssertNoError(t, IntRange("age", 25, 18, 120))
	testkit.AssertNoError(t, IntRange("age", 18, 18, 120))
	testkit.AssertNoError(t, IntRange("age", 120, 18, 120))
}

func TestIntRange_BelowMin(t *testing.T) {
	testkit.AssertError(t, IntRange("age", 17, 18, 120))
}

func TestIntRange_AboveMax(t *testing.T) {
	testkit.AssertError(t, IntRange("age", 121, 18, 120))
}

func TestIntRange_ErrorMessage(t *testing.T) {
	err := IntRange("limit", 0, 1, 100)
	testkit.AssertErrorContains(t, err, "must be between 1 and 100")
}

func TestPositive_Valid(t *testing.T) {
	testkit.AssertNoError(t, Positive("count", 1))
	testkit.AssertNoError(t, Positive("count", 100))
}

func TestPositive_Zero(t *testing.T) {
	testkit.AssertError(t, Positive("count", 0))
}

func TestPositive_Negative(t *testing.T) {
	testkit.AssertError(t, Positive("count", -5))
}

func TestPositive_ErrorMessage(t *testing.T) {
	err := Positive("quantity", 0)
	testkit.AssertErrorContains(t, err, "must be positive")
}
