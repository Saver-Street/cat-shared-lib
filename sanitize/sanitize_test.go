package sanitize

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDocFilename_Normal(t *testing.T) {
	got := DocFilename("report.pdf")
	if got != "report.pdf" {
		t.Errorf("got %q, want report.pdf", got)
	}
}

func TestDocFilename_Empty(t *testing.T) {
	got := DocFilename("")
	if got != "unnamed" {
		t.Errorf("got %q, want unnamed", got)
	}
}

func TestDocFilename_PathTraversal(t *testing.T) {
	got := DocFilename("../../../etc/passwd")
	if got != "passwd" {
		t.Errorf("got %q, want passwd (base only)", got)
	}
}

func TestDocFilename_ControlChars(t *testing.T) {
	got := DocFilename("file\x00name\x01.txt")
	if got != "filename.txt" {
		t.Errorf("got %q, want filename.txt (control chars stripped)", got)
	}
}

func TestDocFilename_AllControlChars(t *testing.T) {
	got := DocFilename("\x01\x02\x03")
	if got != "unnamed" {
		t.Errorf("got %q, want unnamed (all chars stripped)", got)
	}
}

func TestDocFilename_DEL(t *testing.T) {
	got := DocFilename("file\x7fname.txt")
	if got != "filename.txt" {
		t.Errorf("got %q, want filename.txt (DEL stripped)", got)
	}
}

func TestNilIfEmpty_Empty(t *testing.T) {
	if NilIfEmpty("") != nil {
		t.Error("expected nil for empty string")
	}
}

func TestNilIfEmpty_NonEmpty(t *testing.T) {
	got := NilIfEmpty("hello")
	if got == nil || *got != "hello" {
		t.Errorf("got %v, want pointer to hello", got)
	}
}

func TestIsDuplicateKey_True(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"}
	if !IsDuplicateKey(err) {
		t.Error("expected true for 23505 error")
	}
}

func TestIsDuplicateKey_False(t *testing.T) {
	err := errors.New("some other error")
	if IsDuplicateKey(err) {
		t.Error("expected false for non-duplicate error")
	}
}

func TestIsDuplicateKey_Nil(t *testing.T) {
	if IsDuplicateKey(nil) {
		t.Error("expected false for nil error")
	}
}

func TestDocFilename_Unicode(t *testing.T) {
	got := DocFilename("résumé_日本語.pdf")
	if got != "résumé_日本語.pdf" {
		t.Errorf("got %q, want résumé_日本語.pdf", got)
	}
}

func TestDocFilename_LongFilename(t *testing.T) {
	long := strings.Repeat("a", 300) + ".pdf"
	got := DocFilename(long)
	if len(got) != len(long) {
		t.Errorf("long filename length = %d, want %d (no truncation)", len(got), len(long))
	}
}

func TestDocFilename_OnlySpaces(t *testing.T) {
	got := DocFilename("   ")
	if got != "   " {
		t.Errorf("got %q, want spaces preserved", got)
	}
}

func TestDocFilename_DotFile(t *testing.T) {
	got := DocFilename(".gitignore")
	if got != ".gitignore" {
		t.Errorf("got %q, want .gitignore", got)
	}
}

func TestDocFilename_Emoji(t *testing.T) {
	got := DocFilename("📄document.pdf")
	if got != "📄document.pdf" {
		t.Errorf("got %q, want 📄document.pdf", got)
	}
}

func TestIsDuplicateKey_WrappedError(t *testing.T) {
	inner := &pgconn.PgError{Code: "23505"}
	wrapped := fmt.Errorf("insert failed: %w", inner)
	if !IsDuplicateKey(wrapped) {
		t.Error("expected true for wrapped 23505 error")
	}
}

func TestNilIfEmpty_Whitespace(t *testing.T) {
	got := NilIfEmpty(" ")
	if got == nil {
		t.Fatal("whitespace-only string should not be nil")
	}
	if *got != " " {
		t.Errorf("got %q, want single space", *got)
	}
}

// --- Benchmarks ---

func BenchmarkDocFilename(b *testing.B) {
	for b.Loop() {
		DocFilename("../path/to/my\x00file\x7fname.pdf")
	}
}

func BenchmarkNilIfEmpty(b *testing.B) {
	for b.Loop() {
		NilIfEmpty("")
		NilIfEmpty("hello")
	}
}

func BenchmarkIsDuplicateKey(b *testing.B) {
	err := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
	for b.Loop() {
		IsDuplicateKey(err)
	}
}

func TestTrimAndNilIfEmpty_Empty(t *testing.T) {
	if TrimAndNilIfEmpty("") != nil {
		t.Error("expected nil for empty string")
	}
}

func TestTrimAndNilIfEmpty_WhitespaceOnly(t *testing.T) {
	if TrimAndNilIfEmpty("   ") != nil {
		t.Error("expected nil for whitespace-only string")
	}
}

func TestTrimAndNilIfEmpty_NonEmpty(t *testing.T) {
	got := TrimAndNilIfEmpty("  hello  ")
	if got == nil {
		t.Fatal("expected non-nil for non-empty string after trim")
	}
	if *got != "hello" {
		t.Errorf("got %q, want hello", *got)
	}
}

func TestTrimAndNilIfEmpty_NoTrimNeeded(t *testing.T) {
	got := TrimAndNilIfEmpty("world")
	if got == nil || *got != "world" {
		t.Errorf("got %v, want &world", got)
	}
}

// --- TruncateFilename table-driven tests ---

func TestTruncateFilename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"normal with extension", "report.pdf", 20, "report.pdf"},
		{"truncated preserves ext", "averylongfilename.txt", 10, "averyl.txt"},
		{"exact length", "file.txt", 8, "file.txt"},
		{"shorter than maxLen", "hi.go", 100, "hi.go"},
		{"maxLen zero", "file.txt", 0, ""},
		{"maxLen negative", "file.txt", -5, ""},
		{"empty name", "", 10, ""},
		{"empty name zero maxLen", "", 0, ""},
		{"extension longer than maxLen", "x.verylongext", 3, ".verylongext"},
		{"extension equal to maxLen", "x.tx", 3, ".tx"},
		{"no extension", "readme", 4, "read"},
		{"no extension within limit", "readme", 100, "readme"},
		{"unicode filename", "résumé.pdf", 7, "rés.pdf"},
		{"unicode ext preserved", "日本語ファイル.txt", 6, "日本.txt"},
		{"emoji filename", "📄📝.txt", 5, "📄.txt"},
		{"dot file", ".gitignore", 5, ".gitignore"},
		{"single char and ext", "a.go", 4, "a.go"},
		{"maxLen 1 with ext", "file.txt", 1, ".txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateFilename(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateFilename(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// --- MaxLength table-driven tests ---

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"normal truncation", "hello world", 5, "hello"},
		{"within limit", "hello", 10, "hello"},
		{"exact limit", "hello", 5, "hello"},
		{"maxLen zero", "hello", 0, ""},
		{"maxLen negative", "hello", -1, ""},
		{"empty string", "", 10, ""},
		{"empty string zero maxLen", "", 0, ""},
		{"unicode truncation", "héllo", 3, "hél"},
		{"unicode within limit", "héllo", 100, "héllo"},
		{"CJK characters", "日本語テスト", 3, "日本語"},
		{"emoji truncation", "👍🎉🚀💡", 2, "👍🎉"},
		{"single char", "x", 1, "x"},
		{"maxLen 1 long string", "abcdef", 1, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxLength(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("MaxLength(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// --- SanitizeEmail table-driven tests ---

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal email", "alice@example.com", "alice@example.com"},
		{"leading/trailing whitespace", "  alice@example.com  ", "alice@example.com"},
		{"uppercase", "Alice@Example.COM", "alice@example.com"},
		{"whitespace and uppercase", "  User@Example.COM  ", "user@example.com"},
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},
		{"tab whitespace", "\tuser@test.com\t", "user@test.com"},
		{"mixed whitespace", " \t user@test.com \n ", "user@test.com"},
		{"already lowercase no spaces", "test@test.io", "test@test.io"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeEmail(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- IsDatabaseError table-driven tests ---

func TestIsDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{"unique violation match", &pgconn.PgError{Code: "23505"}, "23505", true},
		{"foreign key violation", &pgconn.PgError{Code: "23503"}, "23503", true},
		{"not null violation", &pgconn.PgError{Code: "23502"}, "23502", true},
		{"check violation", &pgconn.PgError{Code: "23514"}, "23514", true},
		{"non-matching code", &pgconn.PgError{Code: "42000"}, "23505", false},
		{"nil error", nil, "23505", false},
		{"nil error different code", nil, "23503", false},
		{"plain error not pgconn", errors.New("some error"), "23505", false},
		{"wrapped pgconn error", fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: "23505"}), "23505", true},
		{"wrong code for unique", &pgconn.PgError{Code: "23505"}, "23503", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDatabaseError(tt.err, tt.code)
			if got != tt.want {
				t.Errorf("IsDatabaseError(%v, %q) = %v, want %v", tt.err, tt.code, got, tt.want)
			}
		})
	}
}

func BenchmarkTruncateFilename(b *testing.B) {
	for b.Loop() {
		TruncateFilename("long_document_name.pdf", 15)
	}
}

func BenchmarkMaxLength(b *testing.B) {
	s := "a string that is fairly long for benchmarking purposes"
	for b.Loop() {
		MaxLength(s, 20)
	}
}

func BenchmarkSanitizeEmail(b *testing.B) {
	for b.Loop() {
		SanitizeEmail("  User@Example.COM  ")
	}
}

func BenchmarkIsDatabaseError(b *testing.B) {
	err := &pgconn.PgError{Code: "23505"}
	for b.Loop() {
		IsDatabaseError(err, "23505")
	}
}

func BenchmarkTrimAndNilIfEmpty(b *testing.B) {
	for b.Loop() {
		TrimAndNilIfEmpty("  hello  ")
	}
}

func strPtr(s string) *string { return &s }
func int64Ptr(n int64) *int64 { return &n }
func boolPtr(b bool) *bool    { return &b }

func TestNullString_Nil(t *testing.T) {
	if got := NullString(nil, "default"); got != "default" {
		t.Errorf("NullString(nil) = %q, want default", got)
	}
}

func TestNullString_Value(t *testing.T) {
	if got := NullString(strPtr("hello"), "default"); got != "hello" {
		t.Errorf("NullString(ptr) = %q, want hello", got)
	}
}

func TestNullString_Empty(t *testing.T) {
	if got := NullString(strPtr(""), "default"); got != "" {
		t.Errorf("NullString(empty ptr) = %q, want empty", got)
	}
}

func TestNullInt64_Nil(t *testing.T) {
	if got := NullInt64(nil, 42); got != 42 {
		t.Errorf("NullInt64(nil) = %d, want 42", got)
	}
}

func TestNullInt64_Value(t *testing.T) {
	if got := NullInt64(int64Ptr(99), 0); got != 99 {
		t.Errorf("NullInt64(ptr) = %d, want 99", got)
	}
}

func TestNullInt64_Zero(t *testing.T) {
	if got := NullInt64(int64Ptr(0), 42); got != 0 {
		t.Errorf("NullInt64(zero ptr) = %d, want 0", got)
	}
}

func TestNullBool_Nil(t *testing.T) {
	if got := NullBool(nil, true); !got {
		t.Error("NullBool(nil) should return default true")
	}
}

func TestNullBool_False(t *testing.T) {
	if got := NullBool(boolPtr(false), true); got {
		t.Error("NullBool(false ptr) should return false")
	}
}

func TestNullBool_True(t *testing.T) {
	if got := NullBool(boolPtr(true), false); !got {
		t.Error("NullBool(true ptr) should return true")
	}
}

func TestDeref_Nil(t *testing.T) {
	var p *string
	if got := Deref(p, "default"); got != "default" { t.Errorf("got %q", got) }
}
func TestDeref_NonNil(t *testing.T) {
	s := "hello"
	if got := Deref(&s, "default"); got != "hello" { t.Errorf("got %q", got) }
}
func TestDeref_NilInt(t *testing.T) {
	var p *int
	if got := Deref(p, 42); got != 42 { t.Errorf("got %d", got) }
}
func TestDeref_NonNilInt(t *testing.T) {
	n := 7
	if got := Deref(&n, 42); got != 7 { t.Errorf("got %d", got) }
}
func TestDeref_NilBool(t *testing.T) {
	var p *bool
	if got := Deref(p, true); !got { t.Error("got false") }
}
func TestDeref_NonNilBool(t *testing.T) {
	b := false
	if got := Deref(&b, true); got { t.Error("got true") }
}

func TestStripHTML_Basic(t *testing.T) {
if got := StripHTML("<p>Hello</p>"); got != "Hello" {
t.Errorf("got %q, want %q", got, "Hello")
}
}

func TestStripHTML_NoTags(t *testing.T) {
if got := StripHTML("plain text"); got != "plain text" {
t.Errorf("got %q, want %q", got, "plain text")
}
}

func TestStripHTML_Empty(t *testing.T) {
if got := StripHTML(""); got != "" {
t.Errorf("got %q, want empty string", got)
}
}

func TestStripHTML_NestedTags(t *testing.T) {
if got := StripHTML("<div><b>bold</b> text</div>"); got != "bold text" {
t.Errorf("got %q, want %q", got, "bold text")
}
}

func TestStripHTML_Attributes(t *testing.T) {
	input := "<a href=\"https://example.com\">link</a>"
	if got := StripHTML(input); got != "link" {
		t.Errorf("got %q, want %q", got, "link")
	}
}

func TestStripHTML_SelfClosing(t *testing.T) {
if got := StripHTML("line1<br/>line2"); got != "line1line2" {
t.Errorf("got %q, want %q", got, "line1line2")
}
}

func TestStripHTML_Script(t *testing.T) {
if got := StripHTML("<script>alert(1)</script>safe"); got != "alert(1)safe" {
t.Errorf("StripHTML strips tags not content; got %q", got)
}
}

func TestStripHTML_OnlyTags(t *testing.T) {
if got := StripHTML("<p></p>"); got != "" {
t.Errorf("got %q, want empty", got)
}
}
