package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func makeTestToken(t *testing.T, secret string, claims map[string]any, exp time.Time) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	if !exp.IsZero() {
		claims["exp"] = exp.Unix()
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	sigInput := header + "." + payloadB64

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return sigInput + "." + sig
}

const testSecret = "this-is-a-test-secret-that-is-at-least-32-bytes-long"

func TestNewLocalValidator_ShortSecret(t *testing.T) {
	_, err := NewLocalValidator("short")
	if !errors.Is(err, ErrSecretTooShort) {
		t.Fatalf("expected ErrSecretTooShort, got %v", err)
	}
}

func TestNewLocalValidator_ValidSecret(t *testing.T) {
	v, err := NewLocalValidator(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	if v == nil {
		t.Fatal("expected non-nil validator")
	}
}

func TestValidate_ValidToken(t *testing.T) {
	v, _ := NewLocalValidator(testSecret)
	token := makeTestToken(t, testSecret, map[string]any{
		"uid":            "user-123",
		"email":          "test@example.com",
		"name":           "Test User",
		"role":           "admin",
		"status":         "active",
		"is_tester":      true,
		"email_verified": true,
	}, time.Now().Add(time.Hour))

	claims, err := v.Validate(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected uid user-123, got %s", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}
	if !claims.IsTester {
		t.Error("expected is_tester=true")
	}
	if !claims.EmailVerified {
		t.Error("expected email_verified=true")
	}
	if claims.ExpiresIn <= 0 {
		t.Errorf("expected positive ExpiresIn, got %d", claims.ExpiresIn)
	}
}

func TestValidate_ExpiredToken(t *testing.T) {
	v, _ := NewLocalValidator(testSecret)
	token := makeTestToken(t, testSecret, map[string]any{
		"uid": "user-123",
	}, time.Now().Add(-time.Hour))

	_, err := v.Validate(token)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestValidate_WrongSecret(t *testing.T) {
	v, _ := NewLocalValidator(testSecret)
	wrongSecret := "wrong-secret-that-is-also-at-least-32-bytes-long!!"
	token := makeTestToken(t, wrongSecret, map[string]any{
		"uid": "user-123",
	}, time.Now().Add(time.Hour))

	_, err := v.Validate(token)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestValidate_MissingUID(t *testing.T) {
	v, _ := NewLocalValidator(testSecret)
	token := makeTestToken(t, testSecret, map[string]any{
		"email": "no-uid@example.com",
	}, time.Now().Add(time.Hour))

	_, err := v.Validate(token)
	if err == nil {
		t.Fatal("expected error for missing uid")
	}
	if !strings.Contains(err.Error(), "missing uid") {
		t.Fatalf("expected 'missing uid' in error, got: %v", err)
	}
}

func TestValidate_MalformedToken(t *testing.T) {
	v, _ := NewLocalValidator(testSecret)
	_, err := v.Validate("not-a-jwt")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestValidate_WrongAlgorithm(t *testing.T) {
	v, _ := NewLocalValidator(testSecret)
	// Craft a token with alg:none
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"uid":"hacker"}`))
	token := fmt.Sprintf("%s.%s.", header, payload)

	_, err := v.Validate(token)
	if err == nil {
		t.Fatal("expected error for alg:none")
	}
}
