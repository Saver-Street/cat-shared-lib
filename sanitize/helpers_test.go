package sanitize

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestSanitizePhone_WithPlus(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("+1 (555) 123-4567"), "+15551234567")
}

func TestSanitizePhone_DigitsOnly(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("5551234567"), "5551234567")
}

func TestSanitizePhone_WithDashes(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("555-123-4567"), "5551234567")
}

func TestSanitizePhone_WithParens(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("(555) 123-4567"), "5551234567")
}

func TestSanitizePhone_WithDots(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("555.123.4567"), "5551234567")
}

func TestSanitizePhone_Empty(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone(""), "")
}

func TestSanitizePhone_PlusInMiddle(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("1+23"), "123")
}

func TestSanitizePhone_International(t *testing.T) {
	testkit.AssertEqual(t, SanitizePhone("+44 20 7946 0958"), "+442079460958")
}

func TestSanitizeURL_Basic(t *testing.T) {
	testkit.AssertEqual(t, SanitizeURL("  HTTPS://Example.COM/path/  "), "https://example.com/path")
}

func TestSanitizeURL_TrailingSlash(t *testing.T) {
	testkit.AssertEqual(t, SanitizeURL("https://example.com/"), "https://example.com/")
}

func TestSanitizeURL_NoScheme(t *testing.T) {
	testkit.AssertEqual(t, SanitizeURL("example.com"), "example.com")
}

func TestSanitizeURL_Empty(t *testing.T) {
	testkit.AssertEqual(t, SanitizeURL(""), "")
}

func TestSanitizeURL_WhitespaceOnly(t *testing.T) {
	testkit.AssertEqual(t, SanitizeURL("   "), "")
}

func TestSanitizeURL_WithQuery(t *testing.T) {
	testkit.AssertEqual(t, SanitizeURL("HTTP://Example.COM/path?q=1"), "http://example.com/path?q=1")
}

func TestSanitizeName_Basic(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName("  john   doe  "), "John Doe")
}

func TestSanitizeName_AllCaps(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName("JANE DOE"), "Jane Doe")
}

func TestSanitizeName_AllLower(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName("alice smith"), "Alice Smith")
}

func TestSanitizeName_Mixed(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName("bOB jOnEs"), "Bob Jones")
}

func TestSanitizeName_Single(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName("alice"), "Alice")
}

func TestSanitizeName_Empty(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName(""), "")
}

func TestSanitizeName_Whitespace(t *testing.T) {
	testkit.AssertEqual(t, SanitizeName("   "), "")
}

func BenchmarkSanitizePhone(b *testing.B) {
	for b.Loop() {
		SanitizePhone("+1 (555) 123-4567")
	}
}

func BenchmarkSanitizeURL(b *testing.B) {
	for b.Loop() {
		SanitizeURL("HTTPS://Example.COM/path/")
	}
}

func BenchmarkSanitizeName(b *testing.B) {
	for b.Loop() {
		SanitizeName("  john   doe  ")
	}
}
