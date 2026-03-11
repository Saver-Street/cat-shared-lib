package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// ErrCiphertextTooShort is returned when the ciphertext is shorter than the
// GCM nonce size, indicating it was corrupted or not produced by Encrypt.
var ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")

// ErrInvalidKeySize is returned when the key is not a valid AES key size
// (16, 24, or 32 bytes).
var ErrInvalidKeySize = errors.New("crypto: invalid key size (must be 16, 24, or 32 bytes)")

// Encrypt encrypts plaintext using AES-256-GCM with authenticated encryption.
// The key must be exactly 16, 24, or 32 bytes (for AES-128, AES-192, AES-256).
// Returns the nonce prepended to the ciphertext.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	if err := validateKeySize(len(key)); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext produced by Encrypt using AES-GCM.
// The key must match the one used during encryption.
// The ciphertext must have the nonce prepended (as produced by Encrypt).
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if err := validateKeySize(len(key)); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrCiphertextTooShort
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ct := ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decrypt: %w", err)
	}
	return plaintext, nil
}

// EncryptString is a convenience wrapper that encrypts a plaintext string and
// returns the result as a base64-encoded string (URL-safe, no padding).
func EncryptString(key []byte, plaintext string) (string, error) {
	ct, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(ct), nil
}

// DecryptString decrypts a base64-encoded ciphertext string produced by
// EncryptString and returns the original plaintext string.
func DecryptString(key []byte, ciphertext string) (string, error) {
	ct, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("crypto: decode base64: %w", err)
	}
	pt, err := Decrypt(key, ct)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// EqualBytes reports whether a and b have equal contents using constant-time
// comparison to prevent timing side-channels. Unlike Equal, this operates
// on byte slices directly.
func EqualBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func validateKeySize(n int) error {
	switch n {
	case 16, 24, 32:
		return nil
	default:
		return ErrInvalidKeySize
	}
}
