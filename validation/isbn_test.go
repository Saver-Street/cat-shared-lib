package validation

import (
	"errors"
	"testing"
)

func TestISBN10Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"plain", "0306406152"},
		{"with hyphens", "0-306-40615-2"},
		{"check X", "080442957X"},
		{"check x lowercase", "080442957x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ISBN10("isbn", tt.value); err != nil {
				t.Errorf("ISBN10(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestISBN10Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too short", "123456789"},
		{"too long", "12345678901"},
		{"bad check digit", "0306406151"},
		{"X not at end", "0X06406152"},
		{"non-digit", "030640615a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ISBN10("isbn", tt.value); err == nil {
				t.Errorf("ISBN10(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestISBN13Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"plain", "9780306406157"},
		{"with hyphens", "978-0-306-40615-7"},
		{"another", "9783161484100"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ISBN13("isbn", tt.value); err != nil {
				t.Errorf("ISBN13(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestISBN13Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too short", "978030640615"},
		{"too long", "97803064061571"},
		{"bad check digit", "9780306406158"},
		{"letters", "978030640615a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ISBN13("isbn", tt.value); err == nil {
				t.Errorf("ISBN13(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestISBN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"isbn10", "0306406152", false},
		{"isbn13", "9780306406157", false},
		{"bad length", "12345", true},
		{"isbn10 bad check", "0306406151", true},
		{"isbn13 bad check", "9780306406158", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ISBN("isbn", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ISBN(%q) error = %v; wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestISBN10FieldName(t *testing.T) {
	t.Parallel()
	err := ISBN10("book_isbn", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("type = %T; want *ValidationError", err)
	}
	if ve.Field != "book_isbn" {
		t.Errorf("Field = %q; want book_isbn", ve.Field)
	}
}

func BenchmarkISBN10(b *testing.B) {
	for range b.N {
		_ = ISBN10("isbn", "0306406152")
	}
}

func BenchmarkISBN13(b *testing.B) {
	for range b.N {
		_ = ISBN13("isbn", "9780306406157")
	}
}

func FuzzISBN(f *testing.F) {
	f.Add("0306406152")
	f.Add("9780306406157")
	f.Add("")
	f.Add("invalid")
	f.Fuzz(func(t *testing.T, value string) {
		_ = ISBN("f", value)
	})
}
