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
		t.Error("whitespace-only string should not be nil")
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
