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

func TestAlphanumeric_Valid(t *testing.T) {
testkit.AssertNoError(t, Alphanumeric("code", "ABC123"))
}

func TestAlphanumeric_Invalid(t *testing.T) {
err := Alphanumeric("code", "ABC-123")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "only letters and digits")
}

func TestAlphanumeric_Empty(t *testing.T) {
err := Alphanumeric("code", "")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "is required")
}

func TestAlphanumeric_Trimmed(t *testing.T) {
testkit.AssertNoError(t, Alphanumeric("code", "  abc123  "))
}

func TestNumeric_Valid(t *testing.T) {
testkit.AssertNoError(t, Numeric("zip", "90210"))
}

func TestNumeric_Invalid(t *testing.T) {
err := Numeric("zip", "90-210")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "only digits")
}

func TestNumeric_Empty(t *testing.T) {
err := Numeric("zip", "")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "is required")
}

func TestNumeric_Letters(t *testing.T) {
err := Numeric("zip", "ABC")
testkit.AssertError(t, err)
}

func TestNumeric_Trimmed(t *testing.T) {
testkit.AssertNoError(t, Numeric("zip", "  12345  "))
}

func TestBetween_Valid(t *testing.T) {
testkit.AssertNoError(t, Between("age", 25, 18, 65))
}

func TestBetween_AtMin(t *testing.T) {
testkit.AssertNoError(t, Between("age", 18, 18, 65))
}

func TestBetween_AtMax(t *testing.T) {
testkit.AssertNoError(t, Between("age", 65, 18, 65))
}

func TestBetween_BelowMin(t *testing.T) {
err := Between("age", 10, 18, 65)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "must be between 18 and 65")
}

func TestBetween_AboveMax(t *testing.T) {
err := Between("age", 70, 18, 65)
testkit.AssertError(t, err)
}

func TestBetween_Float(t *testing.T) {
testkit.AssertNoError(t, Between("rate", 0.5, 0.0, 1.0))
}

func TestBetween_String(t *testing.T) {
testkit.AssertNoError(t, Between("grade", "B", "A", "F"))
}

func TestEachString_Valid(t *testing.T) {
tags := []string{"abc", "def", "ghi"}
err := EachString("tags", tags, Alphanumeric)
testkit.AssertNoError(t, err)
}

func TestEachString_Invalid(t *testing.T) {
tags := []string{"abc", "d-e-f", "ghi"}
err := EachString("tags", tags, Alphanumeric)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "tags[1]")
}

func TestEachString_Empty(t *testing.T) {
err := EachString("tags", []string{}, Alphanumeric)
testkit.AssertNoError(t, err)
}

func TestEachString_Email(t *testing.T) {
emails := []string{"a@b.com", "c@d.org"}
err := EachString("emails", emails, Email)
testkit.AssertNoError(t, err)
}

func TestLowercase_Valid(t *testing.T) {
testkit.AssertNoError(t, Lowercase("code", "abc-123"))
}

func TestLowercase_Invalid(t *testing.T) {
err := Lowercase("code", "abcDef")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "lowercase")
}

func TestLowercase_Empty(t *testing.T) {
testkit.AssertNoError(t, Lowercase("code", ""))
}

func TestUppercase_Valid(t *testing.T) {
testkit.AssertNoError(t, Uppercase("code", "ABC-123"))
}

func TestUppercase_Invalid(t *testing.T) {
err := Uppercase("code", "ABCdef")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "uppercase")
}

func TestUppercase_Empty(t *testing.T) {
testkit.AssertNoError(t, Uppercase("code", ""))
}

func TestStartsWith_Valid(t *testing.T) {
testkit.AssertNoError(t, StartsWith("path", "/api/v1/users", "/api/"))
}

func TestStartsWith_Invalid(t *testing.T) {
err := StartsWith("path", "/web/index", "/api/")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "start with")
}

func TestEndsWith_Valid(t *testing.T) {
testkit.AssertNoError(t, EndsWith("file", "report.pdf", ".pdf"))
}

func TestEndsWith_Invalid(t *testing.T) {
err := EndsWith("file", "report.txt", ".pdf")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "end with")
}

func TestNotOneOf_Valid(t *testing.T) {
testkit.AssertNoError(t, NotOneOf("username", "alice", []string{"admin", "root", "system"}))
}

func TestNotOneOf_Invalid(t *testing.T) {
err := NotOneOf("username", "admin", []string{"admin", "root", "system"})
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "admin")
}

func TestNotOneOf_Empty(t *testing.T) {
testkit.AssertNoError(t, NotOneOf("username", "anything", []string{}))
}

func TestJSON_Valid(t *testing.T) {
testkit.AssertNoError(t, JSON(`{"key":"val"}`))
testkit.AssertNoError(t, JSON(`[1,2,3]`))
testkit.AssertNoError(t, JSON(`"hello"`))
testkit.AssertNoError(t, JSON(`42`))
testkit.AssertNoError(t, JSON(`true`))
testkit.AssertNoError(t, JSON(`null`))
}

func TestJSON_Invalid(t *testing.T) {
testkit.AssertError(t, JSON(`{bad}`))
testkit.AssertError(t, JSON(`not json`))
testkit.AssertError(t, JSON(``))
testkit.AssertErrorContains(t, JSON(`{`), "invalid JSON")
}

func TestBase64_Valid(t *testing.T) {
testkit.AssertNoError(t, Base64("aGVsbG8="))
testkit.AssertNoError(t, Base64(""))
testkit.AssertNoError(t, Base64("dGVzdA=="))
}

func TestBase64_Invalid(t *testing.T) {
testkit.AssertError(t, Base64("not!valid!base64"))
testkit.AssertErrorContains(t, Base64("abc"), "invalid base64")
}

func TestIP_Valid(t *testing.T) {
testkit.AssertNoError(t, IP("192.168.1.1"))
testkit.AssertNoError(t, IP("::1"))
testkit.AssertNoError(t, IP("2001:db8::1"))
}

func TestIP_Invalid(t *testing.T) {
testkit.AssertError(t, IP("not-an-ip"))
testkit.AssertError(t, IP(""))
testkit.AssertErrorContains(t, IP("999.999.999.999"), "invalid IP")
}

func TestIPv4_Valid(t *testing.T) {
testkit.AssertNoError(t, IPv4("192.168.1.1"))
testkit.AssertNoError(t, IPv4("10.0.0.1"))
}

func TestIPv4_RejectsIPv6(t *testing.T) {
testkit.AssertError(t, IPv4("::1"))
testkit.AssertError(t, IPv4("2001:db8::1"))
}

func TestIPv4_Invalid(t *testing.T) {
testkit.AssertError(t, IPv4("not-an-ip"))
testkit.AssertErrorContains(t, IPv4(""), "invalid IPv4")
}
