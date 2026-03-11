package crypto

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // SHA1 is required by RFC 4226 HOTP
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// TOTPConfig holds configuration for TOTP generation and validation.
type TOTPConfig struct {
	// Digits is the number of digits in the code (default: 6).
	Digits int
	// Period is the time step in seconds (default: 30).
	Period int
	// Skew is the number of periods to check before/after current (default: 1).
	Skew int
}

// DefaultTOTPConfig returns a TOTPConfig with standard defaults.
func DefaultTOTPConfig() TOTPConfig {
	return TOTPConfig{Digits: 6, Period: 30, Skew: 1}
}

// GenerateTOTPSecret creates a random 20-byte base32-encoded secret
// suitable for use as a TOTP shared secret.
func GenerateTOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := randReader.Read(b); err != nil {
		return "", fmt.Errorf("totp: failed to generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// HOTP computes an HMAC-based One-Time Password (RFC 4226).
func HOTP(secret []byte, counter uint64, digits int) string {
	mac := hmac.New(sha1.New, secret)
	_ = binary.Write(mac, binary.BigEndian, counter)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	mod := uint32(math.Pow10(digits))
	return fmt.Sprintf("%0*d", digits, code%mod)
}

// TOTP computes a Time-based One-Time Password (RFC 6238) for the given time.
func TOTP(secret []byte, t time.Time, cfg TOTPConfig) string {
	counter := uint64(t.Unix()) / uint64(cfg.Period)
	return HOTP(secret, counter, cfg.Digits)
}

// ValidateTOTP checks whether code is valid for the secret at the given time,
// allowing for clock skew.
func ValidateTOTP(secret []byte, code string, t time.Time, cfg TOTPConfig) bool {
	counter := uint64(t.Unix()) / uint64(cfg.Period)
	for i := -cfg.Skew; i <= cfg.Skew; i++ {
		c := int64(counter) + int64(i)
		if c < 0 {
			continue
		}
		if HOTP(secret, uint64(c), cfg.Digits) == code {
			return true
		}
	}
	return false
}

// ParseTOTPSecret decodes a base32-encoded TOTP secret.
func ParseTOTPSecret(encoded string) ([]byte, error) {
	encoded = strings.ToUpper(strings.TrimSpace(encoded))
	// Add padding if missing.
	if m := len(encoded) % 8; m != 0 {
		encoded += strings.Repeat("=", 8-m)
	}
	return base32.StdEncoding.DecodeString(encoded)
}

// TOTPKeyURI builds an otpauth:// URI for QR code generation.
func TOTPKeyURI(issuer, account, secret string, cfg TOTPConfig) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		issuer, account, secret, issuer, cfg.Digits, cfg.Period,
	)
}
