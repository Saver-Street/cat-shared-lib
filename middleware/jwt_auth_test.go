package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

// helper to create a valid HS256 token
func makeToken(t *testing.T, claims JWTClaims, secret []byte) string {
	t.Helper()
	tok, err := SignHS256(claims, secret)
	if err != nil {
		t.Fatalf("SignHS256: %v", err)
	}
	return tok
}

func TestJWTAuth_ValidToken(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt")
	now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	claims := JWTClaims{
		Subject:   "user-123",
		Email:     "user@example.com",
		Role:      "admin",
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Hour).Unix(),
		Issuer:    "cat-service",
	}
	token := makeToken(t, claims, secret)

	mw := JWTAuth(JWTAuthConfig{
		Secret:  secret,
		Issuer:  "cat-service",
		NowFunc: func() time.Time { return now },
	})

	var gotUserID, gotEmail, gotRole string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = GetUserID(r)
		gotEmail = GetUserEmail(r)
		gotRole = GetUserRole(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusOK)
	testkit.AssertEqual(t, gotUserID, "user-123")
	testkit.AssertEqual(t, gotEmail, "user@example.com")
	testkit.AssertEqual(t, gotRole, "admin")
}

func TestJWTAuth_MissingToken(t *testing.T) {
	mw := JWTAuth(JWTAuthConfig{Secret: []byte("secret")})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_InvalidBearerPrefix(t *testing.T) {
	mw := JWTAuth(JWTAuthConfig{Secret: []byte("secret")})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_MalformedToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"no dots", "notavalidtoken"},
		{"one dot", "header.payload"},
		{"four dots", "a.b.c.d"},
		{"bad base64 header", "!!!.cGF5bG9hZA.c2ln"},
		{"bad base64 payload", "eyJhbGciOiJIUzI1NiJ9.!!!.c2ln"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := JWTAuth(JWTAuthConfig{Secret: []byte("secret")})
			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			testkit.AssertStatus(t, rr, http.StatusUnauthorized)
		})
	}
}

func TestJWTAuth_WrongAlgorithm(t *testing.T) {
	// Build a token with RS256 header
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user-1","exp":9999999999}`))
	token := header + "." + payload + "." + "fakesig"

	mw := JWTAuth(JWTAuthConfig{Secret: []byte("secret")})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_BadSignature(t *testing.T) {
	secret := []byte("correct-secret")
	claims := JWTClaims{Subject: "user-1", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	token := makeToken(t, claims, secret)

	// Use a different secret for validation
	mw := JWTAuth(JWTAuthConfig{Secret: []byte("wrong-secret")})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	secret := []byte("secret")
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	claims := JWTClaims{Subject: "user-1", ExpiresAt: past.Unix()}
	token := makeToken(t, claims, secret)

	mw := JWTAuth(JWTAuthConfig{Secret: secret})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_MissingSubject(t *testing.T) {
	secret := []byte("secret")
	claims := JWTClaims{ExpiresAt: time.Now().Add(time.Hour).Unix()}
	token := makeToken(t, claims, secret)

	mw := JWTAuth(JWTAuthConfig{Secret: secret})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_WrongIssuer(t *testing.T) {
	secret := []byte("secret")
	claims := JWTClaims{Subject: "user-1", ExpiresAt: time.Now().Add(time.Hour).Unix(), Issuer: "other"}
	token := makeToken(t, claims, secret)

	mw := JWTAuth(JWTAuthConfig{Secret: secret, Issuer: "expected"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusUnauthorized)
}

func TestJWTAuth_SkipPaths(t *testing.T) {
	mw := JWTAuth(JWTAuthConfig{
		Secret:    []byte("secret"),
		SkipPaths: []string{"/health", "/ready"},
	})

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusOK)
	testkit.AssertTrue(t, called)
}

func TestJWTAuth_NoExpiration(t *testing.T) {
	secret := []byte("secret")
	claims := JWTClaims{Subject: "user-1"}
	token := makeToken(t, claims, secret)

	mw := JWTAuth(JWTAuthConfig{Secret: secret})
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusOK)
	testkit.AssertTrue(t, called)
}

func TestJWTAuth_EmptyEmailAndRole(t *testing.T) {
	secret := []byte("secret")
	claims := JWTClaims{Subject: "user-1", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	token := makeToken(t, claims, secret)

	mw := JWTAuth(JWTAuthConfig{Secret: secret})
	var gotEmail, gotRole string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEmail = GetUserEmail(r)
		gotRole = GetUserRole(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rr.Code)
	}
	testkit.AssertEqual(t, gotEmail, "")
	testkit.AssertEqual(t, gotRole, "")
}

func TestSignHS256_RoundTrip(t *testing.T) {
	secret := []byte("round-trip-secret")
	now := time.Now()
	claims := JWTClaims{
		Subject:   "user-42",
		Email:     "alice@example.com",
		Role:      "editor",
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
		Issuer:    "test",
	}

	token, err := SignHS256(claims, secret)
	if err != nil {
		t.Fatalf("SignHS256: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	got, err := validateJWT(req, secret, "test", now)
	if err != nil {
		t.Fatalf("validateJWT: %v", err)
	}
	testkit.AssertEqual(t, got.Subject, "user-42")
	testkit.AssertEqual(t, got.Email, "alice@example.com")
	testkit.AssertEqual(t, got.Role, "editor")
}

func TestValidateJWT_BadSignatureBase64(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user-1"}`))
	token := header + "." + payload + ".!!invalid!!base64!!"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := validateJWT(req, []byte("secret"), "", time.Now())
	testkit.AssertErrorIs(t, err, ErrSignatureFail)
}

func TestValidateJWT_InvalidHeaderJSON(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`not json`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u"}`))
	token := header + "." + payload + ".sig"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := validateJWT(req, []byte("secret"), "", time.Now())
	testkit.AssertErrorIs(t, err, ErrInvalidToken)
}

func TestValidateJWT_InvalidPayloadJSON(t *testing.T) {
	secret := []byte("secret")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`not json`))

	// Need valid signature
	sigInput := header + "." + payload
	tok, _ := signPayload(sigInput, secret)
	token := sigInput + "." + tok

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := validateJWT(req, secret, "", time.Now())
	testkit.AssertErrorIs(t, err, ErrInvalidToken)
}

func signPayload(input string, secret []byte) (string, error) {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}

func TestJWTClaims_JSON(t *testing.T) {
	c := JWTClaims{Subject: "u1", Email: "a@b.com", Role: "admin"}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var got JWTClaims
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	testkit.AssertEqual(t, got.Subject, c.Subject)
	testkit.AssertEqual(t, got.Email, c.Email)
	testkit.AssertEqual(t, got.Role, c.Role)
}

func TestJWTAuthConfig_NowFunc(t *testing.T) {
	fixed := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := JWTAuthConfig{NowFunc: func() time.Time { return fixed }}
	testkit.AssertTrue(t, cfg.now().Equal(fixed))

	cfg2 := JWTAuthConfig{}
	before := time.Now()
	got := cfg2.now()
	after := time.Now()
	testkit.AssertFalse(t, got.Before(before) || got.After(after))
}

func TestValidateJWT_InvalidPayloadBase64(t *testing.T) {
	secret := []byte("secret")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	// "!!!" is invalid base64, but HMAC is computed on raw bytes so it still signs fine.
	payload := "!!!"
	sigInput := header + "." + payload
	sig, _ := signPayload(sigInput, secret)
	token := sigInput + "." + sig

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := validateJWT(req, secret, "", time.Now())
	testkit.AssertErrorIs(t, err, ErrInvalidToken)
}
