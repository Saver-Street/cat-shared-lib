package crypto

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"
)

func TestDefaultTOTPConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultTOTPConfig()
	if cfg.Digits != 6 {
		t.Errorf("Digits = %d; want 6", cfg.Digits)
	}
	if cfg.Period != 30 {
		t.Errorf("Period = %d; want 30", cfg.Period)
	}
	if cfg.Skew != 1 {
		t.Errorf("Skew = %d; want 1", cfg.Skew)
	}
}

func TestHOTPRFC4226(t *testing.T) {
	t.Parallel()
	// Test vectors from RFC 4226 Appendix D.
	secret := []byte("12345678901234567890")
	expected := []string{
		"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489",
	}
	for i, want := range expected {
		t.Run(want, func(t *testing.T) {
			t.Parallel()
			got := HOTP(secret, uint64(i), 6)
			if got != want {
				t.Errorf("HOTP(counter=%d) = %q; want %q", i, got, want)
			}
		})
	}
}

func TestTOTP(t *testing.T) {
	t.Parallel()
	// RFC 6238 test vector: secret = "12345678901234567890", time = 59, step=30
	// Expected: 287082
	secret := []byte("12345678901234567890")
	ts := time.Unix(59, 0)
	cfg := TOTPConfig{Digits: 6, Period: 30, Skew: 0}
	code := TOTP(secret, ts, cfg)
	if code != "287082" {
		t.Errorf("TOTP(t=59) = %q; want 287082", code)
	}
}

func TestTOTPT1111111109(t *testing.T) {
	t.Parallel()
	secret := []byte("12345678901234567890")
	ts := time.Unix(1111111109, 0)
	cfg := TOTPConfig{Digits: 8, Period: 30, Skew: 0}
	code := TOTP(secret, ts, cfg)
	if code != "07081804" {
		t.Errorf("TOTP(t=1111111109) = %q; want 07081804", code)
	}
}

func TestValidateTOTP(t *testing.T) {
	t.Parallel()
	secret := []byte("12345678901234567890")
	ts := time.Unix(59, 0)
	cfg := TOTPConfig{Digits: 6, Period: 30, Skew: 1}
	code := TOTP(secret, ts, cfg)
	if !ValidateTOTP(secret, code, ts, cfg) {
		t.Error("ValidateTOTP = false for correct code")
	}
}

func TestValidateTOTPSkew(t *testing.T) {
	t.Parallel()
	secret := []byte("12345678901234567890")
	cfg := TOTPConfig{Digits: 6, Period: 30, Skew: 1}
	ts := time.Unix(59, 0)
	code := TOTP(secret, ts, cfg)
	// Validate one period later should still work with skew=1
	later := time.Unix(89, 0) // next period
	if !ValidateTOTP(secret, code, later, cfg) {
		t.Error("ValidateTOTP with skew = false for adjacent period")
	}
}

func TestValidateTOTPInvalid(t *testing.T) {
	t.Parallel()
	secret := []byte("12345678901234567890")
	cfg := TOTPConfig{Digits: 6, Period: 30, Skew: 0}
	ts := time.Unix(59, 0)
	if ValidateTOTP(secret, "000000", ts, cfg) {
		t.Error("ValidateTOTP = true for wrong code")
	}
}

func TestValidateTOTPNegativeCounter(t *testing.T) {
	t.Parallel()
	secret := []byte("12345678901234567890")
	cfg := TOTPConfig{Digits: 6, Period: 30, Skew: 1}
	ts := time.Unix(0, 0) // counter=0, skew=-1 would be negative
	code := TOTP(secret, ts, cfg)
	if !ValidateTOTP(secret, code, ts, cfg) {
		t.Error("ValidateTOTP at t=0 with skew=1 failed")
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	t.Parallel()
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret error: %v", err)
	}
	if len(secret) == 0 {
		t.Error("secret is empty")
	}
	// Should be valid base32.
	_, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		t.Errorf("not valid base32: %v", err)
	}
}

func TestGenerateTOTPSecretError(t *testing.T) {
	t.Parallel()
	orig := randReader
	randReader = failReader{}
	defer func() { randReader = orig }()

	_, err := GenerateTOTPSecret()
	if err == nil {
		t.Error("expected error")
	}
}

func TestParseTOTPSecret(t *testing.T) {
	t.Parallel()
	secret := "JBSWY3DPEHPK3PXP"
	b, err := ParseTOTPSecret(secret)
	if err != nil {
		t.Fatalf("ParseTOTPSecret error: %v", err)
	}
	if len(b) == 0 {
		t.Error("parsed secret is empty")
	}
}

func TestParseTOTPSecretLowercase(t *testing.T) {
	t.Parallel()
	secret := "jbswy3dpehpk3pxp"
	b, err := ParseTOTPSecret(secret)
	if err != nil {
		t.Fatalf("ParseTOTPSecret lowercase error: %v", err)
	}
	if len(b) == 0 {
		t.Error("parsed secret is empty")
	}
}

func TestParseTOTPSecretInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParseTOTPSecret("!@#$%^&*")
	if err == nil {
		t.Error("expected error for invalid base32")
	}
}

func TestTOTPKeyURI(t *testing.T) {
	t.Parallel()
	cfg := DefaultTOTPConfig()
	uri := TOTPKeyURI("MyApp", "user@example.com", "SECRET", cfg)
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Errorf("URI doesn't start with otpauth://totp/: %s", uri)
	}
	if !strings.Contains(uri, "secret=SECRET") {
		t.Errorf("URI missing secret: %s", uri)
	}
	if !strings.Contains(uri, "issuer=MyApp") {
		t.Errorf("URI missing issuer: %s", uri)
	}
}

func BenchmarkHOTP(b *testing.B) {
	secret := []byte("12345678901234567890")
	for i := range b.N {
		HOTP(secret, uint64(i), 6)
	}
}

func BenchmarkTOTP(b *testing.B) {
	secret := []byte("12345678901234567890")
	ts := time.Now()
	cfg := DefaultTOTPConfig()
	for range b.N {
		TOTP(secret, ts, cfg)
	}
}

func FuzzHOTP(f *testing.F) {
	f.Add(uint64(0), 6)
	f.Add(uint64(999999), 8)
	f.Fuzz(func(t *testing.T, counter uint64, digits int) {
		if digits < 1 || digits > 10 {
			return
		}
		secret := []byte("12345678901234567890")
		code := HOTP(secret, counter, digits)
		if len(code) != digits {
			t.Errorf("HOTP len = %d; want %d", len(code), digits)
		}
	})
}

func TestParseTOTPSecretNeedsPadding(t *testing.T) {
	t.Parallel()
	// JBSWY3D = 7 chars, needs padding to 8
	_, err := ParseTOTPSecret("JBSWY3D")
	if err != nil {
		t.Fatalf("ParseTOTPSecret with padding needed: %v", err)
	}
}
