package stringutil_test

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/stringutil"
)

// ── ToKebabCase ──

func TestToKebabCase_CamelCase(t *testing.T) {
	if got := stringutil.ToKebabCase("camelCase"); got != "camel-case" {
		t.Errorf("got %q, want %q", got, "camel-case")
	}
}

func TestToKebabCase_PascalCase(t *testing.T) {
	if got := stringutil.ToKebabCase("PascalCase"); got != "pascal-case" {
		t.Errorf("got %q, want %q", got, "pascal-case")
	}
}

func TestToKebabCase_SnakeCase(t *testing.T) {
	if got := stringutil.ToKebabCase("snake_case"); got != "snake-case" {
		t.Errorf("got %q, want %q", got, "snake-case")
	}
}

func TestToKebabCase_Empty(t *testing.T) {
	if got := stringutil.ToKebabCase(""); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// ── ToPascalCase ──

func TestToPascalCase_SnakeCase(t *testing.T) {
	if got := stringutil.ToPascalCase("hello_world"); got != "HelloWorld" {
		t.Errorf("got %q, want %q", got, "HelloWorld")
	}
}

func TestToPascalCase_KebabCase(t *testing.T) {
	if got := stringutil.ToPascalCase("hello-world"); got != "HelloWorld" {
		t.Errorf("got %q, want %q", got, "HelloWorld")
	}
}

func TestToPascalCase_CamelCase(t *testing.T) {
	if got := stringutil.ToPascalCase("helloWorld"); got != "HelloWorld" {
		t.Errorf("got %q, want %q", got, "HelloWorld")
	}
}

func TestToPascalCase_Empty(t *testing.T) {
	if got := stringutil.ToPascalCase(""); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// ── PadLeft ──

func TestPadLeft_Normal(t *testing.T) {
	if got := stringutil.PadLeft("42", 5, '0'); got != "00042" {
		t.Errorf("got %q, want %q", got, "00042")
	}
}

func TestPadLeft_AlreadyLong(t *testing.T) {
	if got := stringutil.PadLeft("hello", 3, '.'); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestPadLeft_Unicode(t *testing.T) {
	if got := stringutil.PadLeft("日", 3, '★'); got != "★★日" {
		t.Errorf("got %q, want %q", got, "★★日")
	}
}

// ── PadRight ──

func TestPadRight_Normal(t *testing.T) {
	if got := stringutil.PadRight("hi", 5, '.'); got != "hi..." {
		t.Errorf("got %q, want %q", got, "hi...")
	}
}

func TestPadRight_Empty(t *testing.T) {
	if got := stringutil.PadRight("", 3, 'x'); got != "xxx" {
		t.Errorf("got %q, want %q", got, "xxx")
	}
}

// ── IsBlank ──

func TestIsBlank_Empty(t *testing.T) {
	if !stringutil.IsBlank("") {
		t.Error("expected true for empty string")
	}
}

func TestIsBlank_Spaces(t *testing.T) {
	if !stringutil.IsBlank("   \t\n") {
		t.Error("expected true for whitespace-only string")
	}
}

func TestIsBlank_NotBlank(t *testing.T) {
	if stringutil.IsBlank("hello") {
		t.Error("expected false for non-blank string")
	}
}

// ── Reverse ──

func TestReverse_Normal(t *testing.T) {
	if got := stringutil.Reverse("hello"); got != "olleh" {
		t.Errorf("got %q, want %q", got, "olleh")
	}
}

func TestReverse_Empty(t *testing.T) {
	if got := stringutil.Reverse(""); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestReverse_Unicode(t *testing.T) {
	if got := stringutil.Reverse("héllo"); got != "olléh" {
		t.Errorf("got %q, want %q", got, "olléh")
	}
}

func TestReverse_SingleRune(t *testing.T) {
	if got := stringutil.Reverse("x"); got != "x" {
		t.Errorf("got %q, want %q", got, "x")
	}
}

// ── WordWrap ──

func TestWordWrap_Normal(t *testing.T) {
	got := stringutil.WordWrap("the quick brown fox", 10)
	want := "the quick\nbrown fox"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWordWrap_Empty(t *testing.T) {
	if got := stringutil.WordWrap("", 10); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestWordWrap_ZeroWidth(t *testing.T) {
	s := "hello world"
	if got := stringutil.WordWrap(s, 0); got != s {
		t.Errorf("got %q, want %q", got, s)
	}
}

// ── CountWords ──

func TestCountWords_Normal(t *testing.T) {
	if got := stringutil.CountWords("hello world foo"); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestCountWords_Empty(t *testing.T) {
	if got := stringutil.CountWords(""); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestCountWords_OnlySpaces(t *testing.T) {
	if got := stringutil.CountWords("   "); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

// ── Benchmarks (Go 1.25 b.Loop style) ──

func BenchmarkPadLeft(b *testing.B) {
	for b.Loop() {
		stringutil.PadLeft("42", 20, '0')
	}
}

func BenchmarkReverse(b *testing.B) {
	for b.Loop() {
		stringutil.Reverse("the quick brown fox jumps over the lazy dog")
	}
}

func BenchmarkWordWrap(b *testing.B) {
	for b.Loop() {
		stringutil.WordWrap("the quick brown fox jumps over the lazy dog near the river bank", 20)
	}
}

// ── Fuzz ──

func FuzzReverseRoundTrip(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("日本語")
	f.Add("a")
	f.Fuzz(func(t *testing.T, s string) {
		if got := stringutil.Reverse(stringutil.Reverse(s)); got != s {
			t.Errorf("Reverse(Reverse(%q)) = %q", s, got)
		}
	})
}
