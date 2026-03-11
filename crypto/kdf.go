package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/scrypt"
)

const (
	// DefaultSaltSize is the default salt length in bytes.
	DefaultSaltSize = 16
	// DefaultKeyLen is the default derived key length in bytes.
	DefaultKeyLen = 32
	// scrypt parameters (N=32768, r=8, p=1) balance security and speed.
	scryptN = 32768
	scryptR = 8
	scryptP = 1
)

// DeriveKeyScrypt derives a key from password and salt using scrypt.
// The returned key is DefaultKeyLen bytes. Salt must be at least
// DefaultSaltSize bytes.
func DeriveKeyScrypt(password, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}
	if len(salt) < DefaultSaltSize {
		return nil, errors.New("crypto: salt must be at least 16 bytes")
	}
	return scrypt.Key(password, salt, scryptN, scryptR, scryptP, DefaultKeyLen)
}

// GenerateSalt returns a cryptographically random salt of DefaultSaltSize
// bytes.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, DefaultSaltSize)
	if _, err := io.ReadFull(randReader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// DeriveKeyHKDF derives a key using HKDF-SHA256. The ikm (input keying
// material) must not be empty. Salt and info can be nil.
func DeriveKeyHKDF(ikm, salt, info []byte, keyLen int) ([]byte, error) {
	if len(ikm) == 0 {
		return nil, errors.New("crypto: ikm must not be empty")
	}
	if keyLen < 1 {
		return nil, errors.New("crypto: keyLen must be at least 1")
	}
	r := hkdf.New(sha256.New, ikm, salt, info)
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, err
	}
	return key, nil
}

// DeriveKeyHex is like DeriveKeyScrypt but returns the key as a hex string.
func DeriveKeyHex(password, salt []byte) (string, error) {
	key, err := DeriveKeyScrypt(password, salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}
