package sanitize

import (
	"testing"
)

func TestTruncateWords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"basic", "the quick brown fox", 2, "the quick..."},
		{"exact", "the quick", 2, "the quick"},
		{"fewer words", "hello", 5, "hello"},
		{"empty", "", 3, ""},
		{"zero n", "hello world", 0, ""},
		{"negative n", "hello", -1, ""},
		{"single word", "hello world", 1, "hello..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TruncateWords(tt.s, tt.n)
			if got != tt.want {
				t.Errorf("TruncateWords(%q, %d) = %q; want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		s     string
		width int
		pad   rune
		want  string
	}{
		{"basic", "42", 5, '0', "00042"},
		{"already wide", "hello", 3, ' ', "hello"},
		{"exact", "hi", 2, ' ', "hi"},
		{"empty", "", 3, '*', "***"},
		{"unicode pad", "x", 3, '→', "→→x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := PadLeft(tt.s, tt.width, tt.pad)
			if got != tt.want {
				t.Errorf("PadLeft(%q, %d, %q) = %q; want %q", tt.s, tt.width, tt.pad, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		s     string
		width int
		pad   rune
		want  string
	}{
		{"basic", "hi", 5, '.', "hi..."},
		{"already wide", "hello", 3, ' ', "hello"},
		{"exact", "hi", 2, ' ', "hi"},
		{"empty", "", 3, '-', "---"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := PadRight(tt.s, tt.width, tt.pad)
			if got != tt.want {
				t.Errorf("PadRight(%q, %d, %q) = %q; want %q", tt.s, tt.width, tt.pad, got, tt.want)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"ascii", "hello", "olleh"},
		{"empty", "", ""},
		{"single", "a", "a"},
		{"unicode", "héllo", "olléh"},
		{"emoji", "ab🎉cd", "dc🎉ba"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReverseString(tt.s)
			if got != tt.want {
				t.Errorf("ReverseString(%q) = %q; want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestExcerpt(t *testing.T) {
	t.Parallel()
	text := "The quick brown fox jumps over the lazy dog and runs away"
	tests := []struct {
		name   string
		s      string
		phrase string
		maxLen int
		want   string
	}{
		{"found middle", text, "fox", 20, "...ick brown fox jumps ..."},
		{"found start", text, "The", 15, "The quick brown..."},
		{"found end", text, "away", 20, "...zy dog and runs away"},
		{"not found", text, "xyz", 10, "The quick ..."},
		{"short text", "hello", "hello", 20, "hello"},
		{"zero len", text, "fox", 0, ""},
		{"case insensitive", text, "FOX", 20, "...ick brown fox jumps ..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Excerpt(tt.s, tt.phrase, tt.maxLen)
			if got != tt.want {
				t.Errorf("Excerpt(%q, %q, %d) = %q; want %q", tt.s, tt.phrase, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestWordWrap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		s     string
		width int
		want  string
	}{
		{"basic", "the quick brown fox", 10, "the quick\nbrown fox"},
		{"long word", "superlongword short", 5, "superlongword\nshort"},
		{"exact fit", "hello world", 11, "hello world"},
		{"empty", "", 10, ""},
		{"zero width", "hello world", 0, "hello world"},
		{"single word", "hello", 10, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := WordWrap(tt.s, tt.width)
			if got != tt.want {
				t.Errorf("WordWrap(%q, %d) = %q; want %q", tt.s, tt.width, got, tt.want)
			}
		})
	}
}

func BenchmarkTruncateWords(b *testing.B) {
	s := "the quick brown fox jumps over the lazy dog"
	for range b.N {
		TruncateWords(s, 5)
	}
}

func BenchmarkReverseString(b *testing.B) {
	for range b.N {
		ReverseString("hello world 🎉")
	}
}

func BenchmarkWordWrap(b *testing.B) {
	s := "the quick brown fox jumps over the lazy dog and then goes home"
	for range b.N {
		WordWrap(s, 15)
	}
}

func FuzzTruncateWords(f *testing.F) {
	f.Add("hello world", 2)
	f.Add("", 0)
	f.Add("one", 1)
	f.Fuzz(func(t *testing.T, s string, n int) {
		_ = TruncateWords(s, n)
	})
}

func FuzzReverseString(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("🎉")
	f.Fuzz(func(t *testing.T, s string) {
		result := ReverseString(ReverseString(s))
		if result != s {
			t.Errorf("double reverse of %q = %q", s, result)
		}
	})
}
