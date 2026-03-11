package sanitize

import "testing"

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"heading h1", "# Hello", "Hello"},
		{"heading h3", "### Title", "Title"},
		{"bold asterisks", "**bold**", "bold"},
		{"bold underscores", "__bold__", "bold"},
		{"italic", "*italic*", "italic"},
		{"bold italic", "***both***", "both"},
		{"strikethrough", "~~removed~~", "removed"},
		{"inline code", "`code`", "code"},
		{"link", "[text](http://example.com)", "text"},
		{"image", "![alt](http://img.png)", "alt"},
		{"blockquote", "> quoted text", "quoted text"},
		{"horizontal rule", "---", ""},
		{"bullet list", "* item one", "item one"},
		{"numbered list", "1. first item", "first item"},
		{"plain text", "hello world", "hello world"},
		{"empty", "", ""},
		{"mixed", "# Title\n\n**bold** and *italic* with [link](url)", "Title\n\nbold and italic with link"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripMarkdown(tt.input)
			if got != tt.want {
				t.Fatalf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRedact(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"SSN", "My SSN is 123-45-6789", "My SSN is [REDACTED]"},
		{"credit card spaces", "Card: 4111 1111 1111 1111", "Card: [REDACTED]"},
		{"credit card dashes", "Card: 4111-1111-1111-1111", "Card: [REDACTED]"},
		{"email", "Contact user@example.com please", "Contact [REDACTED] please"},
		{"multiple", "SSN 123-45-6789 email a@b.com", "SSN [REDACTED] email [REDACTED]"},
		{"no match", "nothing sensitive here", "nothing sensitive here"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Redact(tt.input)
			if got != tt.want {
				t.Fatalf("Redact(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkStripMarkdown(b *testing.B) {
	input := "# Title\n\n**bold** text with [link](url) and `code`"
	for b.Loop() {
		StripMarkdown(input)
	}
}

func BenchmarkRedact(b *testing.B) {
	input := "SSN 123-45-6789 card 4111 1111 1111 1111 email user@test.com"
	for b.Loop() {
		Redact(input)
	}
}

func FuzzStripMarkdown(f *testing.F) {
	f.Add("# heading")
	f.Add("**bold**")
	f.Add("[link](url)")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		_ = StripMarkdown(s)
	})
}

func FuzzRedact(f *testing.F) {
	f.Add("123-45-6789")
	f.Add("user@example.com")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		_ = Redact(s)
	})
}
