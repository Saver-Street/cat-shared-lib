package validation

import (
	"regexp"
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
	testkit.AssertError(t, Required("name", ""))
	testkit.AssertError(t, Required("name", "   "))
}

func TestMinLength(t *testing.T) {
	testkit.AssertNoError(t, MinLength("name", "hello", 3))
	testkit.AssertError(t, MinLength("name", "hi", 3))
	testkit.AssertError(t, MinLength("name", "   hi   ", 3))
}

func TestMaxLength(t *testing.T) {
	testkit.AssertNoError(t, MaxLength("name", "hi", 5))
	testkit.AssertError(t, MaxLength("name", "toolongvalue", 5))
}

func TestOneOf(t *testing.T) {
	allowed := []string{"admin", "user", "guest"}
	testkit.AssertNoError(t, OneOf("role", "admin", allowed))
	testkit.AssertError(t, OneOf("role", "superadmin", allowed))
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
	testkit.AssertError(t, err)
}

func TestSlug_Valid(t *testing.T) {
	valid := []string{
		"hello",
		"hello-world",
		"my-cool-project",
		"abc123",
		"a1-b2-c3",
		"x",
	}
	for _, s := range valid {
		testkit.AssertNoError(t, Slug("slug", s))
	}
}

func TestSlug_Invalid(t *testing.T) {
	invalid := []string{
		"Hello",
		"hello_world",
		"hello world",
		"-leading-hyphen",
		"trailing-hyphen-",
		"double--hyphen",
		"UPPER",
		"has.dot",
		"has/slash",
		"has@symbol",
	}
	for _, s := range invalid {
		testkit.AssertError(t, Slug("slug", s))
	}
}

func TestSlug_Empty(t *testing.T) {
	testkit.AssertError(t, Slug("slug", ""))
}

func TestSlug_Trimmed(t *testing.T) {
	testkit.AssertNoError(t, Slug("slug", "  hello-world  "))
}

var alphanumRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func TestMatch_Valid(t *testing.T) {
	testkit.AssertNoError(t, Match("code", "ABC123", alphanumRe, "alphanumeric characters"))
}

func TestMatch_Invalid(t *testing.T) {
	err := Match("code", "ABC-123", alphanumRe, "alphanumeric characters")
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "must match alphanumeric characters")
}

func TestMatch_Empty(t *testing.T) {
	err := Match("code", "", alphanumRe, "alphanumeric characters")
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "is required")
}

func TestMatch_Trimmed(t *testing.T) {
	testkit.AssertNoError(t, Match("code", "  ABC123  ", alphanumRe, "alphanumeric characters"))
}

func TestExactLength_Valid(t *testing.T) {
testkit.AssertNoError(t, ExactLength("country", "US", 2))
}

func TestExactLength_TooShort(t *testing.T) {
err := ExactLength("country", "U", 2)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "exactly 2 characters")
}

func TestExactLength_TooLong(t *testing.T) {
err := ExactLength("country", "USA", 2)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "exactly 2 characters")
}

func TestExactLength_Trimmed(t *testing.T) {
testkit.AssertNoError(t, ExactLength("pin", "  1234  ", 4))
}

func TestNoWhitespace_Valid(t *testing.T) {
testkit.AssertNoError(t, NoWhitespace("username", "alice_123"))
}

func TestNoWhitespace_Space(t *testing.T) {
err := NoWhitespace("username", "alice bob")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "must not contain whitespace")
}

func TestNoWhitespace_Tab(t *testing.T) {
testkit.AssertError(t, NoWhitespace("key", "abc\tdef"))
}

func TestNoWhitespace_Empty(t *testing.T) {
err := NoWhitespace("username", "")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "is required")
}

func TestNoWhitespace_Newline(t *testing.T) {
testkit.AssertError(t, NoWhitespace("token", "abc\ndef"))
}

func TestHex_Valid(t *testing.T) {
testkit.AssertNoError(t, Hex("token", "deadBEEF09"))
}

func TestHex_Invalid(t *testing.T) {
err := Hex("token", "GHIJK")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "hexadecimal")
}

func TestHex_Empty(t *testing.T) {
err := Hex("token", "")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "is required")
}

func TestHex_Trimmed(t *testing.T) {
testkit.AssertNoError(t, Hex("token", "  abc123  "))
}
