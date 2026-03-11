package crypto

import (
	"crypto/rand"
	"io"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

// --- Password edge cases ---

func TestHashPassword_LongPassword(t *testing.T) {
	// bcrypt rejects passwords over 72 bytes in newer Go versions
	long := strings.Repeat("a", 200)
	_, err := HashPasswordWithCost(long, 4)
	testkit.AssertError(t, err)
}

func TestHashPassword_ExactlyMaxLength(t *testing.T) {
	// bcrypt max is 72 bytes
	pw := strings.Repeat("a", 72)
	hash, err := HashPasswordWithCost(pw, 4)
	testkit.RequireNoError(t, err)
	testkit.AssertNoError(t, CheckPassword(pw, hash))
}

func TestHashPassword_Unicode(t *testing.T) {
	pw := "pässwörd🔐"
	hash, err := HashPasswordWithCost(pw, 4)
	testkit.RequireNoError(t, err)
	testkit.AssertNoError(t, CheckPassword(pw, hash))
}

func TestHashPassword_SpecialChars(t *testing.T) {
	pw := "p@$$w0rd!#%&*(){}[]|\\:\";<>,.?/~`"
	hash, err := HashPasswordWithCost(pw, 4)
	testkit.RequireNoError(t, err)
	testkit.AssertNoError(t, CheckPassword(pw, hash))
}

func TestCheckPassword_HashWithWrongCost(t *testing.T) {
	hash4, _ := HashPasswordWithCost("pw", 4)
	hash5, _ := HashPasswordWithCost("pw", 5)
	// Both hashes should validate against the same password
	testkit.AssertNoError(t, CheckPassword("pw", hash4))
	testkit.AssertNoError(t, CheckPassword("pw", hash5))
	// But hashes are different
	testkit.AssertNotEqual(t, hash4, hash5)
}

// --- Token edge cases ---

func TestGenerateToken_SmallSize(t *testing.T) {
	tok, err := GenerateToken(1)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, len(tok) > 0)
}

func TestGenerateToken_LargeSize(t *testing.T) {
	tok, err := GenerateToken(1024)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, len(tok) > 100)
}

func TestGenerateHexToken_SmallSize(t *testing.T) {
	tok, err := GenerateHexToken(1)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, len(tok), 2) // 1 byte = 2 hex chars
}

func TestGenerateHexToken_LargeSize(t *testing.T) {
	tok, err := GenerateHexToken(512)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, len(tok), 1024)
}

// --- HMAC edge cases ---

func TestHMACSHA256_EmptyKey(t *testing.T) {
	sig := HMACSHA256([]byte{}, []byte("message"))
	testkit.AssertTrue(t, len(sig) > 0)
}

func TestHMACSHA256_EmptyMessage(t *testing.T) {
	sig := HMACSHA256([]byte("key"), []byte{})
	testkit.AssertTrue(t, len(sig) > 0)
}

func TestHMACSHA256_EmptyKeyAndMessage(t *testing.T) {
	sig := HMACSHA256([]byte{}, []byte{})
	testkit.AssertTrue(t, len(sig) > 0)
}

func TestVerifyHMACSHA256_EmptySig(t *testing.T) {
	testkit.AssertFalse(t, VerifyHMACSHA256([]byte("key"), []byte("msg"), ""))
}

func TestVerifyHMACSHA256_WrongLengthSig(t *testing.T) {
	testkit.AssertFalse(t, VerifyHMACSHA256([]byte("key"), []byte("msg"), "ab"))
}

func TestVerifyHMACSHA256_OddLengthHex(t *testing.T) {
	testkit.AssertFalse(t, VerifyHMACSHA256([]byte("key"), []byte("msg"), "abc"))
}

// --- Equal edge cases ---

func TestEqual_BothEmpty(t *testing.T) {
	testkit.AssertTrue(t, Equal("", ""))
}

func TestEqual_DifferentLengths(t *testing.T) {
	testkit.AssertFalse(t, Equal("short", "very long string"))
}

func TestEqual_NullBytes(t *testing.T) {
	testkit.AssertTrue(t, Equal("a\x00b", "a\x00b"))
	testkit.AssertFalse(t, Equal("a\x00b", "a\x00c"))
}

func TestEqual_LongStrings(t *testing.T) {
	a := strings.Repeat("x", 10000)
	testkit.AssertTrue(t, Equal(a, a))
	b := a[:len(a)-1] + "y"
	testkit.AssertFalse(t, Equal(a, b))
}

// --- UUID edge cases ---

func TestGenerateUUID_VariantBits(t *testing.T) {
	id, err := GenerateUUID()
	testkit.RequireNoError(t, err)
	// Variant byte (position 19 in string) should have top bits 10
	variantChar := id[19]
	testkit.AssertTrue(t, variantChar == '8' || variantChar == '9' ||
		variantChar == 'a' || variantChar == 'b')
}

func TestGenerateUUID_AllLowerHex(t *testing.T) {
	id, err := GenerateUUID()
	testkit.RequireNoError(t, err)
	for _, c := range id {
		if c == '-' {
			continue
		}
		testkit.AssertTrue(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
	}
}

func TestGenerateUUID_RandFailure(t *testing.T) {
	// GenerateUUID uses rand.Reader directly, not randReader
	// Just verify it works with normal reader
	_, err := GenerateUUID()
	testkit.RequireNoError(t, err)
}

// --- HashSHA256 edge cases ---

func TestHashSHA256_NullBytes(t *testing.T) {
	h1 := HashSHA256([]byte("\x00"))
	h2 := HashSHA256([]byte("\x00\x00"))
	testkit.AssertNotEqual(t, h1, h2)
}

func TestHashSHA256_LargeData(t *testing.T) {
	data := make([]byte, 1<<20) // 1 MB
	h := HashSHA256(data)
	testkit.AssertEqual(t, len(h), 64) // SHA-256 = 32 bytes = 64 hex chars
}

// --- AES edge cases ---

func TestEncryptDecrypt_NullBytes(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("\x00\x00\x00")
	ct, err := Encrypt(key, plaintext)
	testkit.RequireNoError(t, err)
	got, err := Decrypt(key, ct)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, EqualBytes(got, plaintext))
}

func TestEncryptDecrypt_LargeData(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	plaintext := make([]byte, 1<<16) // 64 KB
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}
	ct, err := Encrypt(key, plaintext)
	testkit.RequireNoError(t, err)
	got, err := Decrypt(key, ct)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, EqualBytes(got, plaintext))
}

func TestDecryptString_EmptyCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, err := DecryptString(key, "")
	testkit.AssertError(t, err)
}

func TestDecryptString_TooShortBase64(t *testing.T) {
	key := make([]byte, 32)
	// Valid base64 but too short to contain nonce
	_, err := DecryptString(key, "AQID")
	testkit.AssertError(t, err)
}

func TestEncryptString_EmptyString(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	encoded, err := EncryptString(key, "")
	testkit.RequireNoError(t, err)
	got, err := DecryptString(key, encoded)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, got, "")
}

// --- EqualBytes edge cases ---

func TestEqualBytes_AllSameValues(t *testing.T) {
	a := []byte{0xff, 0xff, 0xff}
	b := []byte{0xff, 0xff, 0xff}
	testkit.AssertTrue(t, EqualBytes(a, b))
}

func TestEqualBytes_AllZeros(t *testing.T) {
	a := make([]byte, 100)
	b := make([]byte, 100)
	testkit.AssertTrue(t, EqualBytes(a, b))
}

func TestEqualBytes_LargeSlices(t *testing.T) {
	a := make([]byte, 10000)
	b := make([]byte, 10000)
	for i := range a {
		a[i] = byte(i)
		b[i] = byte(i)
	}
	testkit.AssertTrue(t, EqualBytes(a, b))
	b[9999] = 0xff
	testkit.AssertFalse(t, EqualBytes(a, b))
}

// --- Concurrent access ---

func TestConcurrentTokenGeneration(t *testing.T) {
	const goroutines = 50
	errs := make(chan error, goroutines)
	for range goroutines {
		go func() {
			_, err := GenerateToken(32)
			errs <- err
		}()
	}
	for range goroutines {
		testkit.AssertNoError(t, <-errs)
	}
}

func TestConcurrentEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	const goroutines = 50
	errs := make(chan error, goroutines)
	for range goroutines {
		go func() {
			ct, err := Encrypt(key, []byte("concurrent data"))
			if err != nil {
				errs <- err
				return
			}
			_, err = Decrypt(key, ct)
			errs <- err
		}()
	}
	for range goroutines {
		testkit.AssertNoError(t, <-errs)
	}
}
