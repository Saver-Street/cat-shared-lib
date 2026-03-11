package crypto

import (
	"crypto/rand"
	"errors"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHashPassword_Basic(t *testing.T) {
	hash, err := HashPassword("secret123")
	testkit.RequireNoError(t, err)
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	testkit.AssertNotEqual(t, hash, "secret123")
}

func TestHashPassword_Empty(t *testing.T) {
	_, err := HashPassword("")
	testkit.AssertErrorIs(t, err, ErrEmptyPassword)
}

func TestHashPasswordWithCost(t *testing.T) {
	hash, err := HashPasswordWithCost("password", 4) // min cost for speed
	testkit.RequireNoError(t, err)
	cost, _ := BcryptCost(hash)
	testkit.AssertEqual(t, cost, 4)
}

func TestCheckPassword_Match(t *testing.T) {
	hash, _ := HashPasswordWithCost("correct", 4)
	if err := CheckPassword("correct", hash); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheckPassword_Mismatch(t *testing.T) {
	hash, _ := HashPasswordWithCost("correct", 4)
	err := CheckPassword("wrong", hash)
	testkit.AssertErrorIs(t, err, ErrInvalidToken)
}

func TestCheckPassword_Empty(t *testing.T) {
	hash, _ := HashPasswordWithCost("correct", 4)
	err := CheckPassword("", hash)
	testkit.AssertErrorIs(t, err, ErrEmptyPassword)
}

func TestCheckPassword_BadHash(t *testing.T) {
	err := CheckPassword("password", "notahash")
	testkit.AssertError(t, err)
}

func TestGenerateToken(t *testing.T) {
	tok, err := GenerateToken(32)
	testkit.RequireNoError(t, err)
	if len(tok) == 0 {
		t.Fatal("expected non-empty token")
	}
	// Two tokens should differ
	tok2, _ := GenerateToken(32)
	testkit.AssertNotEqual(t, tok, tok2)
}

func TestGenerateToken_InvalidLen(t *testing.T) {
	_, err := GenerateToken(0)
	testkit.AssertError(t, err)
	_, err = GenerateToken(-1)
	testkit.AssertError(t, err)
}

func TestGenerateToken_URLSafe(t *testing.T) {
	for range 20 {
		tok, err := GenerateToken(32)
		testkit.RequireNoError(t, err)
		if strings.ContainsAny(tok, "+/=") {
			t.Errorf("token %q contains non-URL-safe chars", tok)
		}
	}
}

func TestGenerateHexToken(t *testing.T) {
	tok, err := GenerateHexToken(16)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, len(tok), 32)
}

func TestGenerateHexToken_InvalidLen(t *testing.T) {
	_, err := GenerateHexToken(0)
	testkit.AssertError(t, err)
}

func TestHMACSHA256(t *testing.T) {
	key := []byte("secret-key")
	msg := []byte("hello world")
	sig := HMACSHA256(key, msg)
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	// Same inputs → same output
	sig2 := HMACSHA256(key, msg)
	testkit.AssertEqual(t, sig, sig2)
	// Different key → different sig
	sig3 := HMACSHA256([]byte("other"), msg)
	testkit.AssertNotEqual(t, sig, sig3)
}

func TestVerifyHMACSHA256_Valid(t *testing.T) {
	key := []byte("k")
	msg := []byte("m")
	sig := HMACSHA256(key, msg)
	testkit.AssertTrue(t, VerifyHMACSHA256(key, msg, sig))
}

func TestVerifyHMACSHA256_Invalid(t *testing.T) {
	key := []byte("k")
	msg := []byte("m")
	testkit.AssertFalse(t, VerifyHMACSHA256(key, msg, "badsig"))
}

func TestEqual(t *testing.T) {
	testkit.AssertTrue(t, Equal("abc", "abc"))
	testkit.AssertFalse(t, Equal("abc", "def"))
	testkit.AssertFalse(t, Equal("", "x"))
}

func TestBcryptCost(t *testing.T) {
	hash, _ := HashPasswordWithCost("pw", 4)
	cost, err := BcryptCost(hash)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, cost, 4)
}

func TestBcryptCost_Invalid(t *testing.T) {
	_, err := BcryptCost("notahash")
	testkit.AssertError(t, err)
}

func TestNeedsRehash(t *testing.T) {
	hash, _ := HashPasswordWithCost("pw", 4)
	testkit.AssertTrue(t, NeedsRehash(hash, 10))
	testkit.AssertFalse(t, NeedsRehash(hash, 4))
	testkit.AssertFalse(t, NeedsRehash(hash, 3))
}

func TestHashPasswordWithCost_InvalidCost(t *testing.T) {
	_, err := HashPasswordWithCost("password", 100) // cost too high
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "crypto: hash password")
}

func TestNeedsRehash_InvalidHash(t *testing.T) {
	testkit.AssertTrue(t, NeedsRehash("badhash", 10))
}

// failReader always returns an error from Read.
type failReader struct{}

func (failReader) Read(_ []byte) (int, error) {
	return 0, errors.New("entropy exhausted")
}

func TestGenerateToken_RandFailure(t *testing.T) {
	old := randReader
	randReader = failReader{}
	defer func() { randReader = old }()

	_, err := GenerateToken(32)
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "generate token")
}

func TestGenerateHexToken_RandFailure(t *testing.T) {
	old := randReader
	randReader = failReader{}
	defer func() { randReader = old }()

	_, err := GenerateHexToken(16)
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "generate hex token")
}

func TestGenerateToken_UsesDefaultReader(t *testing.T) {
	// Verify default reader works (crypto/rand).
	testkit.AssertEqual(t, randReader, rand.Reader)
}

func TestGenerateUUID_Format(t *testing.T) {
id, err := GenerateUUID()
testkit.RequireNoError(t, err)
// UUID v4 format: 8-4-4-4-12
testkit.AssertEqual(t, len(id), 36)
testkit.AssertEqual(t, id[8], byte('-'))
testkit.AssertEqual(t, id[13], byte('-'))
testkit.AssertEqual(t, id[18], byte('-'))
testkit.AssertEqual(t, id[23], byte('-'))
}

func TestGenerateUUID_Version4(t *testing.T) {
id, err := GenerateUUID()
testkit.RequireNoError(t, err)
// Version nibble should be '4'
testkit.AssertEqual(t, id[14], byte('4'))
}

func TestGenerateUUID_Unique(t *testing.T) {
a, err := GenerateUUID()
testkit.RequireNoError(t, err)
b, err := GenerateUUID()
testkit.RequireNoError(t, err)
testkit.AssertNotEqual(t, a, b)
}

func BenchmarkGenerateUUID(b *testing.B) {
for b.Loop() {
GenerateUUID()
}
}

func TestHashSHA256(t *testing.T) {
// SHA-256 of "hello" is well-known
got := HashSHA256([]byte("hello"))
testkit.AssertEqual(t, got, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
}

func TestHashSHA256_Empty(t *testing.T) {
got := HashSHA256([]byte(""))
testkit.AssertEqual(t, got, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
}

func TestHashSHA256_Deterministic(t *testing.T) {
data := []byte("test data")
testkit.AssertEqual(t, HashSHA256(data), HashSHA256(data))
}

func TestRandomString_Length(t *testing.T) {
for _, n := range []int{0, 1, 8, 16, 32, 64} {
s := RandomString(n)
testkit.AssertLen(t, []byte(s), n)
}
}

func TestRandomString_Alphanumeric(t *testing.T) {
s := RandomString(1000)
for _, c := range s {
if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
t.Fatalf("unexpected character %q in random string", c)
}
}
}

func TestRandomString_Unique(t *testing.T) {
a := RandomString(32)
b := RandomString(32)
testkit.AssertNotEqual(t, a, b)
}
