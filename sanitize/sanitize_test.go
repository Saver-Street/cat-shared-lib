package sanitize

import (
	"errors"
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
