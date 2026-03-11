package crypto

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	plaintext := []byte("hello, world!")
	ct, err := Encrypt(key, plaintext)
	testkit.AssertNoError(t, err)
	testkit.AssertTrue(t, len(ct) > len(plaintext))

	got, err := Decrypt(key, ct)
	testkit.AssertNoError(t, err)
	testkit.AssertTrue(t, bytes.Equal(got, plaintext))
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ct, err := Encrypt(key, []byte{})
	testkit.AssertNoError(t, err)

	got, err := Decrypt(key, ct)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, len(got), 0)
}

func TestEncrypt_KeySizes(t *testing.T) {
	tests := []struct {
		name    string
		keyLen  int
		wantErr bool
	}{
		{"AES-128", 16, false},
		{"AES-192", 24, false},
		{"AES-256", 32, false},
		{"too short", 10, true},
		{"too long", 33, true},
		{"zero", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keyLen)
			_, err := Encrypt(key, []byte("data"))
			if tc.wantErr {
				testkit.AssertError(t, err)
				testkit.AssertTrue(t, errors.Is(err, ErrInvalidKeySize))
			} else {
				testkit.AssertNoError(t, err)
			}
		})
	}
}

func TestDecrypt_InvalidKey(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ct, err := Encrypt(key, []byte("secret"))
	testkit.AssertNoError(t, err)

	wrongKey := make([]byte, 32)
	wrongKey[0] = key[0] ^ 0xff // flip a bit
	_, err = Decrypt(wrongKey, ct)
	testkit.AssertError(t, err)
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	key := make([]byte, 32)
	_, err := Decrypt(key, []byte{1, 2, 3})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrCiphertextTooShort))
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ct, err := Encrypt(key, []byte("sensitive data"))
	testkit.AssertNoError(t, err)

	ct[len(ct)-1] ^= 0xff // tamper with last byte
	_, err = Decrypt(key, ct)
	testkit.AssertError(t, err)
}

func TestDecrypt_InvalidKeySize(t *testing.T) {
	_, err := Decrypt([]byte("short"), []byte("doesn't matter"))
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrInvalidKeySize))
}

func TestEncrypt_UniqueNonces(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ct1, err := Encrypt(key, []byte("same plaintext"))
	testkit.AssertNoError(t, err)
	ct2, err := Encrypt(key, []byte("same plaintext"))
	testkit.AssertNoError(t, err)

	testkit.AssertFalse(t, bytes.Equal(ct1, ct2))
}

func TestEncryptString_DecryptString(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	original := "hello, world! 🌍"
	encoded, err := EncryptString(key, original)
	testkit.AssertNoError(t, err)
	testkit.AssertTrue(t, len(encoded) > 0)
	testkit.AssertFalse(t, strings.Contains(encoded, original))

	got, err := DecryptString(key, encoded)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, original)
}

func TestEncryptString_InvalidKey(t *testing.T) {
	_, err := EncryptString([]byte("bad"), "data")
	testkit.AssertError(t, err)
}

func TestDecryptString_InvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	_, err := DecryptString(key, "not!valid!base64!!!")
	testkit.AssertError(t, err)
}

func TestDecryptString_InvalidKey(t *testing.T) {
	key := make([]byte, 32)
	_, err := DecryptString([]byte("bad"), "dGVzdA")
	testkit.AssertError(t, err)
	_ = key
}

func TestEqualBytes(t *testing.T) {
	tests := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"equal", []byte("hello"), []byte("hello"), true},
		{"different", []byte("hello"), []byte("world"), false},
		{"different length", []byte("hi"), []byte("hello"), false},
		{"both empty", []byte{}, []byte{}, true},
		{"both nil", nil, nil, true},
		{"one nil", nil, []byte{}, true},
		{"single byte equal", []byte{42}, []byte{42}, true},
		{"single byte different", []byte{42}, []byte{43}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testkit.AssertEqual(t, EqualBytes(tc.a, tc.b), tc.want)
		})
	}
}

func TestEncrypt_RandReaderFailure(t *testing.T) {
	orig := randReader
	randReader = failReader{}
	defer func() { randReader = orig }()

	key := make([]byte, 32)
	_, err := Encrypt(key, []byte("data"))
	testkit.AssertError(t, err)
}
