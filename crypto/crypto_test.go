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
	if hash == "secret123" {
		t.Fatal("hash must not equal plaintext")
	}
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
	if cost != 4 {
		t.Errorf("expected cost 4, got %d", cost)
	}
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
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
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
	if tok == tok2 {
		t.Error("expected tokens to differ")
	}
}

func TestGenerateToken_InvalidLen(t *testing.T) {
	_, err := GenerateToken(0)
	if err == nil {
		t.Fatal("expected error for zero length")
	}
	_, err = GenerateToken(-1)
	if err == nil {
		t.Fatal("expected error for negative length")
	}
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
	if len(tok) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(tok))
	}
}

func TestGenerateHexToken_InvalidLen(t *testing.T) {
	_, err := GenerateHexToken(0)
	if err == nil {
		t.Fatal("expected error for zero length")
	}
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
	if sig != sig2 {
		t.Error("HMAC is not deterministic")
	}
	// Different key → different sig
	sig3 := HMACSHA256([]byte("other"), msg)
	if sig == sig3 {
		t.Error("different keys should produce different MACs")
	}
}

func TestVerifyHMACSHA256_Valid(t *testing.T) {
	key := []byte("k")
	msg := []byte("m")
	sig := HMACSHA256(key, msg)
	if !VerifyHMACSHA256(key, msg, sig) {
		t.Error("expected valid signature to verify")
	}
}

func TestVerifyHMACSHA256_Invalid(t *testing.T) {
	key := []byte("k")
	msg := []byte("m")
	if VerifyHMACSHA256(key, msg, "badsig") {
		t.Error("expected invalid signature to fail")
	}
}

func TestEqual(t *testing.T) {
	if !Equal("abc", "abc") {
		t.Error("expected equal strings to be equal")
	}
	if Equal("abc", "def") {
		t.Error("expected different strings to be not equal")
	}
	if Equal("", "x") {
		t.Error("expected empty vs non-empty to be not equal")
	}
}

func TestBcryptCost(t *testing.T) {
	hash, _ := HashPasswordWithCost("pw", 4)
	cost, err := BcryptCost(hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost != 4 {
		t.Errorf("expected 4, got %d", cost)
	}
}

func TestBcryptCost_Invalid(t *testing.T) {
	_, err := BcryptCost("notahash")
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestNeedsRehash(t *testing.T) {
	hash, _ := HashPasswordWithCost("pw", 4)
	if !NeedsRehash(hash, 10) {
		t.Error("cost 4 should need rehash to 10")
	}
	if NeedsRehash(hash, 4) {
		t.Error("cost 4 should not need rehash to 4")
	}
	if NeedsRehash(hash, 3) {
		t.Error("cost 4 should not need rehash to 3")
	}
}

func TestHashPasswordWithCost_InvalidCost(t *testing.T) {
	_, err := HashPasswordWithCost("password", 100) // cost too high
	if err == nil {
		t.Fatal("expected error for invalid bcrypt cost")
	}
	testkit.AssertErrorContains(t, err, "crypto: hash password")
}

func TestNeedsRehash_InvalidHash(t *testing.T) {
	if !NeedsRehash("badhash", 10) {
		t.Error("invalid hash should trigger rehash")
	}
}
