package sanitize

import (
	"errors"
	"fmt"
	"strings"
	"testing"
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
	err := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
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
	inner := errors.New("SQLSTATE 23505")
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

// --- MaxLength tests ---

func TestMaxLength_WithinLimit(t *testing.T) {
	if got := MaxLength("hello", 10); got != "hello" {
		t.Errorf("got %q, want hello", got)
	}
}

func TestMaxLength_ExactLimit(t *testing.T) {
	if got := MaxLength("hello", 5); got != "hello" {
		t.Errorf("got %q, want hello", got)
	}
}

func TestMaxLength_Truncated(t *testing.T) {
	if got := MaxLength("hello world", 5); got != "hello" {
		t.Errorf("got %q, want hello", got)
	}
}

func TestMaxLength_ZeroLimit(t *testing.T) {
	if got := MaxLength("hello", 0); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestMaxLength_NegativeLimit(t *testing.T) {
	if got := MaxLength("hello", -1); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestMaxLength_Unicode(t *testing.T) {
	s := "héllo"
	if got := MaxLength(s, 3); got != "hél" {
		t.Errorf("got %q, want hél", got)
	}
}

// --- TruncateFilename tests ---

func TestTruncateFilename_WithinLimit(t *testing.T) {
	if got := TruncateFilename("file.txt", 20); got != "file.txt" {
		t.Errorf("got %q, want file.txt", got)
	}
}

func TestTruncateFilename_Truncated(t *testing.T) {
	got := TruncateFilename("averylongfilename.txt", 10)
	if len([]rune(got)) > 10 {
		t.Errorf("got %q with length %d, want at most 10 runes", got, len([]rune(got)))
	}
	if !strings.HasSuffix(got, ".txt") {
		t.Errorf("got %q, expected .txt extension preserved", got)
	}
}

func TestTruncateFilename_ZeroLimit(t *testing.T) {
	if got := TruncateFilename("file.txt", 0); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestTruncateFilename_NegativeLimit(t *testing.T) {
	if got := TruncateFilename("file.txt", -1); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestTruncateFilename_EmptyName(t *testing.T) {
	if got := TruncateFilename("", 10); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestTruncateFilename_LongExtension(t *testing.T) {
	// extension longer than maxLen: just return the extension
	got := TruncateFilename("x.verylongext", 3)
	if got != ".verylongext" {
		t.Errorf("got %q, want .verylongext", got)
	}
}

func TestSanitizeEmail_Lowercase(t *testing.T) {
got := SanitizeEmail("  User@Example.COM  ")
if got != "user@example.com" {
t.Errorf("SanitizeEmail = %q, want %q", got, "user@example.com")
}
}

func TestSanitizeEmail_AlreadyLower(t *testing.T) {
got := SanitizeEmail("alice@example.com")
if got != "alice@example.com" {
t.Errorf("SanitizeEmail = %q, want %q", got, "alice@example.com")
}
}

func TestSanitizeEmail_Empty(t *testing.T) {
if got := SanitizeEmail(""); got != "" {
t.Errorf("SanitizeEmail empty = %q, want empty string", got)
}
}

func TestIsDatabaseError_MatchingCode(t *testing.T) {
err := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
if !IsDatabaseError(err, "23505") {
t.Error("expected IsDatabaseError to return true for code 23505")
}
}

func TestIsDatabaseError_NonMatchingCode(t *testing.T) {
err := errors.New("ERROR: some other error")
if IsDatabaseError(err, "23505") {
t.Error("expected IsDatabaseError to return false for non-matching code")
}
}

func TestIsDatabaseError_NilError(t *testing.T) {
if IsDatabaseError(nil, "23505") {
t.Error("expected IsDatabaseError to return false for nil error")
}
}
