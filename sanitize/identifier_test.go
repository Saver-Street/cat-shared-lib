package sanitize

import "testing"

func TestSQLIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "users", "users"},
		{"with digits", "table1", "table1"},
		{"leading digit", "1table", "table"},
		{"all digits", "123", "_"},
		{"special chars", "my-table!@#", "mytable"},
		{"underscores", "my_table", "my_table"},
		{"empty", "", "_"},
		{"spaces", "my table", "mytable"},
		{"unicode", "tëst", "tëst"},
		{"mixed", "123_abc_456", "_abc_456"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SQLIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("SQLIdentifier(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCSVEscape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "hello", "hello"},
		{"comma", "a,b", `"a,b"`},
		{"quote", `say "hi"`, `"say ""hi"""`},
		{"newline", "line1\nline2", "\"line1\nline2\""},
		{"cr", "a\rb", "\"a\rb\""},
		{"empty", "", ""},
		{"no special", "abc123", "abc123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CSVEscape(tt.input)
			if got != tt.want {
				t.Errorf("CSVEscape(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHeaderName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "content-type", "Content-Type"},
		{"uppercase", "CONTENT-TYPE", "Content-Type"},
		{"mixed", "x-Custom-Header", "X-Custom-Header"},
		{"double hyphen", "x--header", "X-Header"},
		{"leading hyphen", "-type", "Type"},
		{"trailing hyphen", "type-", "Type"},
		{"no hyphen", "host", "Host"},
		{"empty", "", ""},
		{"non-printable", "con\x00tent", "Content"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HeaderName(tt.input)
			if got != tt.want {
				t.Errorf("HeaderName(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnvVarName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "path", "PATH"},
		{"with hyphen", "my-var", "MY_VAR"},
		{"with dot", "app.config", "APP_CONFIG"},
		{"leading digit", "1var", "VAR"},
		{"all digits", "123", ""},
		{"multiple specials", "a--b..c", "A_B_C"},
		{"trailing special", "var-", "VAR"},
		{"empty", "", ""},
		{"spaces", "my var", "MY_VAR"},
		{"with digits", "port8080", "PORT8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EnvVarName(tt.input)
			if got != tt.want {
				t.Errorf("EnvVarName(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkSQLIdentifier(b *testing.B) {
	for range b.N {
		SQLIdentifier("my-table!@#123")
	}
}

func BenchmarkCSVEscape(b *testing.B) {
	for range b.N {
		CSVEscape(`value with "quotes" and, commas`)
	}
}

func BenchmarkHeaderName(b *testing.B) {
	for range b.N {
		HeaderName("content-type")
	}
}

func BenchmarkEnvVarName(b *testing.B) {
	for range b.N {
		EnvVarName("my-app.config")
	}
}

func FuzzSQLIdentifier(f *testing.F) {
	f.Add("users")
	f.Add("")
	f.Add("123")
	f.Fuzz(func(t *testing.T, s string) {
		result := SQLIdentifier(s)
		if len(result) == 0 {
			t.Error("SQLIdentifier should never return empty string")
		}
	})
}

func FuzzCSVEscape(f *testing.F) {
	f.Add("hello")
	f.Add(`"hi"`)
	f.Add("a,b")
	f.Fuzz(func(t *testing.T, s string) {
		_ = CSVEscape(s)
	})
}
