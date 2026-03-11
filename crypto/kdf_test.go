package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"testing"
)

func TestDeriveKeyScrypt(t *testing.T) {
	t.Parallel()
	salt := bytes.Repeat([]byte{0xAB}, DefaultSaltSize)
	key, err := DeriveKeyScrypt([]byte("password"), salt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key) != DefaultKeyLen {
		t.Errorf("key len = %d; want %d", len(key), DefaultKeyLen)
	}

	// Deterministic: same inputs produce same key
	key2, err := DeriveKeyScrypt([]byte("password"), salt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(key, key2) {
		t.Error("same inputs should produce same key")
	}

	// Different password produces different key
	key3, err := DeriveKeyScrypt([]byte("other"), salt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Equal(key, key3) {
		t.Error("different password should produce different key")
	}
}

func TestDeriveKeyScryptEmptyPassword(t *testing.T) {
	t.Parallel()
	salt := bytes.Repeat([]byte{0xAB}, DefaultSaltSize)
	_, err := DeriveKeyScrypt(nil, salt)
	if !errors.Is(err, ErrEmptyPassword) {
		t.Errorf("err = %v; want ErrEmptyPassword", err)
	}
}

func TestDeriveKeyScryptShortSalt(t *testing.T) {
	t.Parallel()
	_, err := DeriveKeyScrypt([]byte("pw"), []byte("short"))
	if err == nil {
		t.Error("expected error for short salt")
	}
}

func TestGenerateSalt(t *testing.T) {
	t.Parallel()
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(salt) != DefaultSaltSize {
		t.Errorf("salt len = %d; want %d", len(salt), DefaultSaltSize)
	}
	// Two salts should differ
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Equal(salt, salt2) {
		t.Error("two salts should differ")
	}
}

func TestGenerateSaltError(t *testing.T) {
	t.Parallel()
	orig := randReader
	randReader = &errorReader{}
	t.Cleanup(func() { randReader = orig })

	_, err := GenerateSalt()
	if err == nil {
		t.Error("expected error when reader fails")
	}
}

type errorReader struct{}

func (r *errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestDeriveKeyHKDF(t *testing.T) {
	t.Parallel()
	ikm := []byte("secret input keying material")
	salt := []byte("optional salt")
	info := []byte("context info")

	key, err := DeriveKeyHKDF(ikm, salt, info, 32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("key len = %d; want 32", len(key))
	}

	// Deterministic
	key2, err := DeriveKeyHKDF(ikm, salt, info, 32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(key, key2) {
		t.Error("same inputs should produce same key")
	}

	// Nil salt/info allowed
	key3, err := DeriveKeyHKDF(ikm, nil, nil, 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key3) != 16 {
		t.Errorf("key len = %d; want 16", len(key3))
	}
}

func TestDeriveKeyHKDFEmptyIKM(t *testing.T) {
	t.Parallel()
	_, err := DeriveKeyHKDF(nil, nil, nil, 32)
	if err == nil {
		t.Error("expected error for empty ikm")
	}
}

func TestDeriveKeyHKDFBadKeyLen(t *testing.T) {
	t.Parallel()
	_, err := DeriveKeyHKDF([]byte("ikm"), nil, nil, 0)
	if err == nil {
		t.Error("expected error for keyLen=0")
	}
}

func TestDeriveKeyHex(t *testing.T) {
	t.Parallel()
	salt := bytes.Repeat([]byte{0xCD}, DefaultSaltSize)
	h, err := DeriveKeyHex([]byte("password"), salt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := hex.DecodeString(h)
	if err != nil {
		t.Fatalf("invalid hex: %v", err)
	}
	if len(b) != DefaultKeyLen {
		t.Errorf("decoded len = %d; want %d", len(b), DefaultKeyLen)
	}
}

func TestDeriveKeyHexError(t *testing.T) {
	t.Parallel()
	_, err := DeriveKeyHex(nil, bytes.Repeat([]byte{0xCD}, DefaultSaltSize))
	if err == nil {
		t.Error("expected error for empty password")
	}
}

func BenchmarkDeriveKeyScrypt(b *testing.B) {
	salt := bytes.Repeat([]byte{0xAB}, DefaultSaltSize)
	pw := []byte("benchmark-password")
	for range b.N {
		_, _ = DeriveKeyScrypt(pw, salt)
	}
}

func BenchmarkDeriveKeyHKDF(b *testing.B) {
	ikm := []byte("benchmark-ikm")
	salt := []byte("benchmark-salt")
	for range b.N {
		_, _ = DeriveKeyHKDF(ikm, salt, nil, 32)
	}
}

func FuzzDeriveKeyHKDF(f *testing.F) {
	f.Add([]byte("ikm"), []byte("salt"), 16)
	f.Fuzz(func(t *testing.T, ikm, salt []byte, keyLen int) {
		if len(ikm) == 0 || keyLen < 1 || keyLen > 255*32 {
			return
		}
		key, err := DeriveKeyHKDF(ikm, salt, nil, keyLen)
		if err != nil {
			return
		}
		if len(key) != keyLen {
			t.Errorf("key len = %d; want %d", len(key), keyLen)
		}
	})
}

func TestDeriveKeyHKDFReadError(t *testing.T) {
t.Parallel()
// HKDF-SHA256 max output is 255*32 = 8160 bytes; requesting more
// triggers an io.ReadFull error.
_, err := DeriveKeyHKDF([]byte("ikm"), nil, nil, 255*32+1)
if err == nil {
t.Error("expected error for keyLen > max HKDF output")
}
}
