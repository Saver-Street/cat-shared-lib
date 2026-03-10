package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// makeTestToken creates a valid HS256 JWT for benchmarking.
func makeTestToken(b *testing.B, secret []byte) string {
	b.Helper()
	claims := JWTClaims{
		Subject:   "bench-user-id",
		Email:     "bench@example.com",
		Role:      "user",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Issuer:    "bench",
	}
	tok, err := SignHS256(claims, secret)
	if err != nil {
		b.Fatalf("SignHS256: %v", err)
	}
	return tok
}

func BenchmarkJWTValidation(b *testing.B) {
	secret := []byte("bench-secret-key-32-bytes-long!!")
	tok := makeTestToken(b, secret)
	cfg := JWTAuthConfig{Secret: secret, Issuer: "bench"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		_, err := validateJWT(req, cfg.Secret, cfg.Issuer, time.Now())
		if err != nil {
			b.Fatalf("validateJWT: %v", err)
		}
	}
}

func BenchmarkJWTAuth_Middleware(b *testing.B) {
	secret := []byte("bench-secret-key-32-bytes-long!!")
	tok := makeTestToken(b, secret)

	handler := JWTAuth(JWTAuthConfig{Secret: secret, Issuer: "bench"})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkJWTAuth_SkipPath(b *testing.B) {
	secret := []byte("bench-secret-key-32-bytes-long!!")
	handler := JWTAuth(JWTAuthConfig{
		Secret:    secret,
		SkipPaths: []string{"/health"},
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkSignHS256(b *testing.B) {
	secret := []byte("bench-secret-key-32-bytes-long!!")
	claims := JWTClaims{
		Subject:   "user-id-123",
		Email:     "user@example.com",
		Role:      "admin",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Issuer:    "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SignHS256(claims, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogging_Middleware(b *testing.B) {
	handler := Logging(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/resource", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkRateLimiter_Middleware(b *testing.B) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerWindow: 1000000,
		WindowDuration:    time.Minute,
	})
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}
