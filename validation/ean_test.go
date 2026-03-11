package validation

import (
	"errors"
	"testing"
)

func TestEAN8Valid(t *testing.T) {
	t.Parallel()
	tests := []string{"96385074", "65833254"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			if err := EAN8("ean", v); err != nil {
				t.Errorf("EAN8(%q) = %v; want nil", v, err)
			}
		})
	}
}

func TestEAN8Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too short", "9638507"},
		{"too long", "963850741"},
		{"bad check", "96385075"},
		{"letters", "9638507a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := EAN8("ean", tt.value); err == nil {
				t.Errorf("EAN8(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestEAN13Valid(t *testing.T) {
	t.Parallel()
	tests := []string{"4006381333931", "5901234123457"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			if err := EAN13("ean", v); err != nil {
				t.Errorf("EAN13(%q) = %v; want nil", v, err)
			}
		})
	}
}

func TestEAN13Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too short", "400638133393"},
		{"too long", "40063813339311"},
		{"bad check", "4006381333932"},
		{"letters", "400638133393a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := EAN13("ean", tt.value); err == nil {
				t.Errorf("EAN13(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestUPCValid(t *testing.T) {
	t.Parallel()
	tests := []string{"036000291452", "012345678905"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			if err := UPC("upc", v); err != nil {
				t.Errorf("UPC(%q) = %v; want nil", v, err)
			}
		})
	}
}

func TestUPCInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too short", "03600029145"},
		{"too long", "0360002914523"},
		{"bad check", "036000291453"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := UPC("upc", tt.value); err == nil {
				t.Errorf("UPC(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestEAN8FieldName(t *testing.T) {
	t.Parallel()
	err := EAN8("barcode", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("type = %T; want *ValidationError", err)
	}
	if ve.Field != "barcode" {
		t.Errorf("Field = %q; want barcode", ve.Field)
	}
}

func BenchmarkEAN13(b *testing.B) {
	for range b.N {
		_ = EAN13("ean", "4006381333931")
	}
}

func BenchmarkUPC(b *testing.B) {
	for range b.N {
		_ = UPC("upc", "036000291452")
	}
}

func FuzzEAN13(f *testing.F) {
	f.Add("4006381333931")
	f.Add("")
	f.Add("bad")
	f.Fuzz(func(t *testing.T, value string) {
		_ = EAN13("f", value)
	})
}
