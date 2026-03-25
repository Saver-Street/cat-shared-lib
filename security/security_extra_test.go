package security

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestRedactPII_DeepNesting_Truncated(t *testing.T) {
	// Build a map nested deeper than maxRedactDepth to exercise the truncation path.
	var build func(depth int) map[string]any
	build = func(depth int) map[string]any {
		if depth == 0 {
			return map[string]any{"leaf": "value"}
		}
		return map[string]any{"nested": build(depth - 1)}
	}
	deep := build(maxRedactDepth + 5)
	result := RedactPII(deep)
	// Traverse until we hit the truncation sentinel.
	var traverse func(v any, depth int) bool
	traverse = func(v any, depth int) bool {
		if s, ok := v.(string); ok && s == "[TRUNCATED]" {
			return true
		}
		if m, ok := v.(map[string]any); ok {
			for _, child := range m {
				if traverse(child, depth+1) {
					return true
				}
			}
		}
		return false
	}
	testkit.AssertTrue(t, traverse(result, 0))
}

func TestTruncateForLog_Basic(t *testing.T) {
	got := TruncateForLog("hello world", 5)
	testkit.AssertEqual(t, got, "hello")
}

func TestTruncateForLog_UnderLimit(t *testing.T) {
	got := TruncateForLog("hi", 100)
	testkit.AssertEqual(t, got, "hi")
}

func TestTruncateForLog_StripControlChars(t *testing.T) {
	// \n, \r, \t are control characters and should be stripped
	got := TruncateForLog("a\nb\rc", 10)
	testkit.AssertEqual(t, got, "abc")
}

func TestTruncateForLog_ZeroMaxLen(t *testing.T) {
	testkit.AssertEqual(t, TruncateForLog("hello", 0), "")
}

func TestTruncateForLog_NegativeMaxLen(t *testing.T) {
	testkit.AssertEqual(t, TruncateForLog("hello", -1), "")
}

func TestTruncateForLog_Empty(t *testing.T) {
	testkit.AssertEqual(t, TruncateForLog("", 10), "")
}

func TestTruncateForLog_Unicode(t *testing.T) {
	got := TruncateForLog("héllo", 3)
	testkit.AssertEqual(t, got, "hél")
}

func BenchmarkTruncateForLog(b *testing.B) {
	s := "This is a long string that needs to be truncated for safe logging"
	for b.Loop() {
		TruncateForLog(s, 20)
	}
}

func TestSanitizeHeader_Clean(t *testing.T) {
	testkit.AssertEqual(t, SanitizeHeader("application/json"), "application/json")
}

func TestSanitizeHeader_CRLF(t *testing.T) {
	testkit.AssertEqual(t, SanitizeHeader("value\r\nInjected-Header: evil"), "valueInjected-Header: evil")
}

func TestSanitizeHeader_CR(t *testing.T) {
	testkit.AssertEqual(t, SanitizeHeader("value\rInjected"), "valueInjected")
}

func TestSanitizeHeader_LF(t *testing.T) {
	testkit.AssertEqual(t, SanitizeHeader("value\nInjected"), "valueInjected")
}

func TestSanitizeHeader_Empty(t *testing.T) {
	testkit.AssertEqual(t, SanitizeHeader(""), "")
}

func TestSanitizeHeader_MultipleCRLF(t *testing.T) {
	testkit.AssertEqual(t, SanitizeHeader("a\r\nb\r\nc"), "abc")
}

func BenchmarkSanitizeHeader(b *testing.B) {
	s := "X-Custom-Value: safe-header-value"
	for b.Loop() {
		SanitizeHeader(s)
	}
}

func TestIsRelativeURL_Valid(t *testing.T) {
	testkit.AssertTrue(t, IsRelativeURL("/dashboard"))
	testkit.AssertTrue(t, IsRelativeURL("/users?page=2"))
	testkit.AssertTrue(t, IsRelativeURL("/path/to/resource"))
	testkit.AssertTrue(t, IsRelativeURL("/"))
}

func TestIsRelativeURL_Invalid(t *testing.T) {
	testkit.AssertFalse(t, IsRelativeURL(""))
	testkit.AssertFalse(t, IsRelativeURL("https://evil.com"))
	testkit.AssertFalse(t, IsRelativeURL("//evil.com"))
	testkit.AssertFalse(t, IsRelativeURL("http://evil.com/path"))
	testkit.AssertFalse(t, IsRelativeURL("javascript:alert(1)"))
	testkit.AssertFalse(t, IsRelativeURL("data:text/html,<h1>hi</h1>"))
	testkit.AssertFalse(t, IsRelativeURL("ftp://files.example.com"))
}

func BenchmarkIsRelativeURL(b *testing.B) {
	for b.Loop() {
		IsRelativeURL("/dashboard?tab=overview")
	}
}

func TestMaskEmail(t *testing.T) {
	testkit.AssertEqual(t, MaskEmail("alice@example.com"), "a****@example.com")
}

func TestMaskEmail_Short(t *testing.T) {
	testkit.AssertEqual(t, MaskEmail("a@x.com"), "a@x.com")
}

func TestMaskEmail_NoAt(t *testing.T) {
	testkit.AssertEqual(t, MaskEmail("notanemail"), "notanemail")
}

func TestMaskEmail_LongLocal(t *testing.T) {
	testkit.AssertEqual(t, MaskEmail("john.doe@company.com"), "j*******@company.com")
}

func TestRedactURL_WithCredentials(t *testing.T) {
	got := RedactURL("https://admin:s3cret@db.example.com:5432/mydb")
	testkit.AssertEqual(t, got, "https://REDACTED@db.example.com:5432/mydb")
}

func TestRedactURL_UsernameOnly(t *testing.T) {
	got := RedactURL("https://admin@db.example.com/mydb")
	testkit.AssertEqual(t, got, "https://REDACTED@db.example.com/mydb")
}

func TestRedactURL_NoCredentials(t *testing.T) {
	got := RedactURL("https://example.com/path?q=1")
	testkit.AssertEqual(t, got, "https://example.com/path?q=1")
}

func TestRedactURL_Invalid(t *testing.T) {
	got := RedactURL("://not-a-url")
	testkit.AssertEqual(t, got, "://not-a-url")
}

func TestRedactURL_Empty(t *testing.T) {
	got := RedactURL("")
	testkit.AssertEqual(t, got, "")
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "report.pdf", "report.pdf"},
		{"path traversal", "../../etc/passwd", "passwd"},
		{"windows path", `C:\Users\evil\doc.txt`, "doc.txt"},
		{"unix path", "/var/uploads/file.txt", "file.txt"},
		{"dotdot only", "..", "unnamed"},
		{"dot only", ".", "unnamed"},
		{"empty", "", "unnamed"},
		{"null bytes", "file\x00.txt", "file.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testkit.AssertEqual(t, SanitizeFilename(tt.input), tt.expect)
		})
	}
}

func TestCSPHeader(t *testing.T) {
	csp := CSPHeader(map[string]string{
		"default-src": "'self'",
		"script-src":  "'self' 'unsafe-inline'",
		"img-src":     "'self' data:",
	})
	testkit.AssertEqual(t, csp, "default-src 'self'; img-src 'self' data:; script-src 'self' 'unsafe-inline'")
}

func TestCSPHeader_Single(t *testing.T) {
	csp := CSPHeader(map[string]string{"default-src": "'none'"})
	testkit.AssertEqual(t, csp, "default-src 'none'")
}

func TestCSPHeader_Empty(t *testing.T) {
	csp := CSPHeader(map[string]string{})
	testkit.AssertEqual(t, csp, "")
}

func TestIsStrongPassword_Valid(t *testing.T) {
	testkit.AssertNoError(t, IsStrongPassword("P@ssw0rd!", 8))
}

func TestIsStrongPassword_TooShort(t *testing.T) {
	err := IsStrongPassword("P@1a", 8)
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "at least 8 characters")
}

func TestIsStrongPassword_NoUpper(t *testing.T) {
	err := IsStrongPassword("p@ssw0rd!", 8)
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "uppercase")
}

func TestIsStrongPassword_NoLower(t *testing.T) {
	err := IsStrongPassword("P@SSW0RD!", 8)
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "lowercase")
}

func TestIsStrongPassword_NoDigit(t *testing.T) {
	err := IsStrongPassword("P@ssword!", 8)
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "digit")
}

func TestIsStrongPassword_NoSpecial(t *testing.T) {
	err := IsStrongPassword("Passw0rdd", 8)
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "special")
}
