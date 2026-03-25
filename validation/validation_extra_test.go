package validation

import (
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

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

func TestCIDR_Valid(t *testing.T) {
	testkit.AssertNoError(t, CIDR("192.168.1.0/24"))
	testkit.AssertNoError(t, CIDR("10.0.0.0/8"))
	testkit.AssertNoError(t, CIDR("::1/128"))
}

func TestCIDR_Invalid(t *testing.T) {
	testkit.AssertError(t, CIDR("192.168.1.1"))
	testkit.AssertError(t, CIDR("not-cidr"))
	testkit.AssertErrorContains(t, CIDR(""), "invalid CIDR")
}

func TestHostname_Valid(t *testing.T) {
	testkit.AssertNoError(t, Hostname("host", "example.com"))
	testkit.AssertNoError(t, Hostname("host", "sub.example.com"))
	testkit.AssertNoError(t, Hostname("host", "my-host"))
	testkit.AssertNoError(t, Hostname("host", "a"))
}

func TestHostname_Invalid(t *testing.T) {
	testkit.AssertError(t, Hostname("host", ""))
	testkit.AssertError(t, Hostname("host", "-invalid.com"))
	testkit.AssertError(t, Hostname("host", "invalid-.com"))
	testkit.AssertError(t, Hostname("host", strings.Repeat("a", 254)))
	testkit.AssertError(t, Hostname("host", "under_score.com"))
}

func TestHexColor_Valid(t *testing.T) {
	testkit.AssertNoError(t, HexColor("color", "#fff"))
	testkit.AssertNoError(t, HexColor("color", "#FFF"))
	testkit.AssertNoError(t, HexColor("color", "#ff00ff"))
	testkit.AssertNoError(t, HexColor("color", "#123456"))
}

func TestHexColor_Invalid(t *testing.T) {
	testkit.AssertError(t, HexColor("color", ""))
	testkit.AssertError(t, HexColor("color", "fff"))
	testkit.AssertError(t, HexColor("color", "#ff"))
	testkit.AssertError(t, HexColor("color", "#gggggg"))
	testkit.AssertError(t, HexColor("color", "#12345"))
}

func TestValidator_Chain(t *testing.T) {
	v := NewValidator().
		Check(Required("name", "Alice")).
		Check(Email("email", "alice@example.com")).
		Check(MinLength("name", "Alice", 2))

	testkit.AssertTrue(t, v.Valid())
	testkit.AssertNil(t, v.Errors())
}

func TestValidator_WithErrors(t *testing.T) {
	v := NewValidator().
		Check(Required("name", "")).
		Check(Email("email", "bad"))

	testkit.AssertFalse(t, v.Valid())
	testkit.AssertEqual(t, len(v.Errors()), 2)
}

func TestValidator_CheckIf(t *testing.T) {
	website := ""
	v := NewValidator().
		Check(Required("name", "Alice")).
		CheckIf(website != "", URL("website", website))

	testkit.AssertTrue(t, v.Valid())
}

func TestValidator_CheckIf_Active(t *testing.T) {
	website := "not-a-url"
	v := NewValidator().
		CheckIf(website != "", URL("website", website))

	testkit.AssertFalse(t, v.Valid())
	testkit.AssertEqual(t, len(v.Errors()), 1)
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
