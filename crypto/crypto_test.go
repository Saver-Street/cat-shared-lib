package crypto

import (
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHashPassword_Basic(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	testkit.AssertTrue(t, err != nil)
}

func TestGenerateToken(t *testing.T) {
	tok, err := GenerateToken(32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tok) == 0 {
		t.Fatal("expected non-empty token")
	}
	// Two tokens should differ
	tok2, _ := GenerateToken(32)
	testkit.AssertNotEqual(t, tok, tok2)
}

func TestGenerateToken_InvalidLen(t *testing.T) {
	_, err := GenerateToken(0)
	testkit.AssertTrue(t, err != nil)
	_, err = GenerateToken(-1)
	testkit.AssertTrue(t, err != nil)
}

func TestGenerateToken_URLSafe(t *testing.T) {
	for range 20 {
		tok, err := GenerateToken(32)
		if err != nil {
			t.Fatal(err)
		}
		if strings.ContainsAny(tok, "+/=") {
			t.Errorf("token %q contains non-URL-safe chars", tok)
		}
	}
}

func TestGenerateHexToken(t *testing.T) {
	tok, err := GenerateHexToken(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertEqual(t, len(tok), 32)
}

func TestGenerateHexToken_InvalidLen(t *testing.T) {
	_, err := GenerateHexToken(0)
	testkit.AssertTrue(t, err != nil)
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertEqual(t, cost, 4)
}

func TestBcryptCost_Invalid(t *testing.T) {
	_, err := BcryptCost("notahash")
	testkit.AssertTrue(t, err != nil)
}

func TestNeedsRehash(t *testing.T) {
	hash, _ := HashPasswordWithCost("pw", 4)
	testkit.AssertTrue(t, NeedsRehash(hash, 10))
	testkit.AssertFalse(t, NeedsRehash(hash, 4))
	testkit.AssertFalse(t, NeedsRehash(hash, 3))
}

func TestHashPasswordWithCost_InvalidCost(t *testing.T) {
	_, err := HashPasswordWithCost("password", 100) // cost too high
	testkit.AssertTrue(t, err != nil)
	testkit.AssertErrorContains(t, err, "crypto: hash password")
}

func TestNeedsRehash_InvalidHash(t *testing.T) {
	testkit.AssertTrue(t, NeedsRehash("badhash", 10))
}
