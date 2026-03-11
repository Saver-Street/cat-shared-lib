// Package crypto provides common cryptographic utilities including password
// hashing with bcrypt, secure random token generation, HMAC-SHA256 signing,
// and constant-time comparison helpers.
package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

// randReader is the source of cryptographic randomness. It defaults to
// crypto/rand.Reader and can be overridden in tests to simulate failures.
var randReader io.Reader = rand.Reader

// DefaultCost is the default bcrypt work factor. Increase for more security
// at the cost of hashing speed.
const DefaultCost = bcrypt.DefaultCost

// ErrInvalidToken is returned when a token fails validation.
var ErrInvalidToken = errors.New("crypto: invalid token")

// ErrEmptyPassword is returned when an empty password is provided.
var ErrEmptyPassword = errors.New("crypto: password must not be empty")

// HashPassword hashes the given password using bcrypt with DefaultCost.
// Returns the hashed string suitable for storage.
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultCost)
}

// HashPasswordWithCost hashes the given password using bcrypt with the
// specified cost factor. Use DefaultCost unless you have a specific reason.
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("crypto: hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil if they match, or an error otherwise.
func CheckPassword(password, hash string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidToken
		}
		return fmt.Errorf("crypto: check password: %w", err)
	}
	return nil
}

// GenerateToken generates a cryptographically secure random token of n bytes,
// returned as a URL-safe base64 string (no padding).
func GenerateToken(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("crypto: token length must be positive, got %d", n)
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(randReader, b); err != nil {
		return "", fmt.Errorf("crypto: generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateHexToken generates a cryptographically secure random token of n
// bytes, returned as a lowercase hex string of length 2*n.
func GenerateHexToken(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("crypto: token length must be positive, got %d", n)
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(randReader, b); err != nil {
		return "", fmt.Errorf("crypto: generate hex token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HMACSHA256 returns the HMAC-SHA256 signature of message using key,
// encoded as a lowercase hex string.
func HMACSHA256(key, message []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSHA256 reports whether the given hex-encoded signature is the
// correct HMAC-SHA256 of message under key. Uses constant-time comparison.
func VerifyHMACSHA256(key, message []byte, signature string) bool {
	expected := HMACSHA256(key, message)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}

// Equal reports whether a == b using constant-time comparison to prevent
// timing side-channels.
func Equal(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// BcryptCost returns the cost factor used to generate the given bcrypt hash.
// Useful for detecting whether a hash needs to be rehashed with a higher cost.
func BcryptCost(hash string) (int, error) {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return 0, fmt.Errorf("crypto: read bcrypt cost: %w", err)
	}
	return cost, nil
}

// NeedsRehash reports whether the given bcrypt hash was generated with a cost
// lower than the desired cost, indicating it should be rehashed.
func NeedsRehash(hash string, desiredCost int) bool {
	cost, err := BcryptCost(hash)
	if err != nil {
		return true
	}
	return cost < desiredCost
}

// GenerateUUID returns a new random UUID v4 string. It uses crypto/rand
// for cryptographically secure random bytes.
func GenerateUUID() (string, error) {
var uuid [16]byte
if _, err := io.ReadFull(rand.Reader, uuid[:]); err != nil {
return "", fmt.Errorf("crypto: generate uuid: %w", err)
}
// Set version 4 (bits 12-15 of time_hi_and_version)
uuid[6] = (uuid[6] & 0x0f) | 0x40
// Set variant (bits 6-7 of clock_seq_hi)
uuid[8] = (uuid[8] & 0x3f) | 0x80

return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}
