package validation

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestJWTFormatValid(t *testing.T) {
	t.Parallel()
	// Build a minimal JWT-like token
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"1"}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("signature"))
	token := header + "." + payload + "." + sig

	if err := JWTFormat("token", token); err != nil {
		t.Errorf("JWTFormat() = %v; want nil", err)
	}
}

func TestJWTFormatInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"no dots", "abcdef"},
		{"one dot", "abc.def"},
		{"four parts", "a.b.c.d"},
		{"empty header", ".payload.sig"},
		{"empty payload", "header..sig"},
		{"empty signature", "header.payload."},
		{"invalid base64 header", "not!valid.cGF5bG9hZA.c2ln"},
		{"invalid base64 payload", "aGVhZGVy.not!valid.c2ln"},
		{"invalid base64 sig", "aGVhZGVy.cGF5bG9hZA.not!valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := JWTFormat("token", tt.value); err == nil {
				t.Errorf("JWTFormat(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestJWTFormatFieldName(t *testing.T) {
	t.Parallel()
	err := JWTFormat("auth", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("error type = %T; want *ValidationError", err)
	}
	if ve.Field != "auth" {
		t.Errorf("Field = %q; want auth", ve.Field)
	}
}

func TestSemVerValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"simple", "1.2.3"},
		{"with v prefix", "v1.2.3"},
		{"zero", "0.0.0"},
		{"pre-release", "1.0.0-alpha"},
		{"pre-release dots", "1.0.0-alpha.1"},
		{"build meta", "1.0.0+build"},
		{"pre and build", "1.0.0-beta+build.123"},
		{"large numbers", "100.200.300"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := SemVer("version", tt.value); err != nil {
				t.Errorf("SemVer(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestSemVerInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"one part", "1"},
		{"two parts", "1.2"},
		{"four parts", "1.2.3.4"},
		{"leading zero", "01.2.3"},
		{"leading zero minor", "1.02.3"},
		{"leading zero patch", "1.2.03"},
		{"non-numeric", "a.b.c"},
		{"empty part", "1..3"},
		{"negative", "-1.2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := SemVer("version", tt.value); err == nil {
				t.Errorf("SemVer(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestSegmentName(t *testing.T) {
	t.Parallel()
	if segmentName(0) != "header" {
		t.Error("segmentName(0) != header")
	}
	if segmentName(1) != "payload" {
		t.Error("segmentName(1) != payload")
	}
	if segmentName(2) != "signature" {
		t.Error("segmentName(2) != signature")
	}
}

func TestIsDigits(t *testing.T) {
	t.Parallel()
	if !isDigits("123") {
		t.Error("isDigits(123) = false")
	}
	if isDigits("") {
		t.Error("isDigits('') = true")
	}
	if isDigits("12a") {
		t.Error("isDigits(12a) = true")
	}
}

func TestCutLast(t *testing.T) {
	t.Parallel()
	before, after := cutLast("a+b+c", "+")
	if before != "a+b" || after != "c" {
		t.Errorf("cutLast = (%q, %q); want (a+b, c)", before, after)
	}
	before, after = cutLast("abc", "+")
	if before != "abc" || after != "" {
		t.Errorf("cutLast no sep = (%q, %q); want (abc, '')", before, after)
	}
}

func BenchmarkJWTFormat(b *testing.B) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"1"}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("signature"))
	token := header + "." + payload + "." + sig
	for range b.N {
		_ = JWTFormat("t", token)
	}
}

func BenchmarkSemVer(b *testing.B) {
	for range b.N {
		_ = SemVer("v", "1.2.3-beta+build")
	}
}

func FuzzJWTFormat(f *testing.F) {
	f.Add("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln")
	f.Add("")
	f.Add("abc")
	f.Fuzz(func(t *testing.T, value string) {
		err := JWTFormat("f", value)
		if err == nil {
			parts := strings.Split(value, ".")
			if len(parts) != 3 {
				t.Error("nil error but not 3 parts")
			}
		}
	})
}

func FuzzSemVer(f *testing.F) {
	f.Add("1.2.3")
	f.Add("v0.0.0")
	f.Add("")
	f.Fuzz(func(t *testing.T, value string) {
		_ = SemVer("f", value)
	})
}
