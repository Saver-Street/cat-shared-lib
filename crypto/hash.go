package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
)

// HashSHA512 returns the hex-encoded SHA-512 hash of data.
func HashSHA512(data []byte) string {
	h := sha512.Sum512(data)
	return hex.EncodeToString(h[:])
}

// HashReader returns the hex-encoded SHA-512 hash of r's content.
// Useful for hashing files and streams without loading them into memory.
func HashReader(r io.Reader) (string, error) {
	h := sha512.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HMACSHA512 returns the hex-encoded HMAC-SHA512 of message using key.
func HMACSHA512(key, message []byte) string {
	mac := hmac.New(sha512.New, key)
	_, _ = mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSHA512 reports whether signature matches the HMAC-SHA512 of
// message using key. Comparison is constant-time.
func VerifyHMACSHA512(key, message []byte, signature string) bool {
	expected := HMACSHA512(key, message)
	return hmac.Equal([]byte(expected), []byte(signature))
}
