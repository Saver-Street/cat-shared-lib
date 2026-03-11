package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func FuzzGetClientIP(f *testing.F) {
	f.Add("203.0.113.5, 10.0.0.1", "192.168.1.1:8080")
	f.Add("", "10.0.0.5:8080")
	f.Add("  192.168.1.1  , 10.0.0.1", "127.0.0.1:80")
	f.Add("2001:db8::1, 10.0.0.1", "[::1]:8080")
	f.Add("", "192.168.1.1")

	f.Fuzz(func(t *testing.T, xff, remoteAddr string) {
		r, _ := http.NewRequest("GET", "/", nil)
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		r.RemoteAddr = remoteAddr

		ip := GetClientIP(r)
		// Should never return empty
		if ip == "" {
			t.Errorf("GetClientIP returned empty for xff=%q remote=%q", xff, remoteAddr)
		}
	})
}

func FuzzIsExemptFromRateLimit(f *testing.F) {
	f.Add("/assets/main.js")
	f.Add("/health")
	f.Add("/api/health")
	f.Add("/api/users")
	f.Add("/icons/logo.png")
	f.Add("/static/bundle.js")
	f.Add("")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic
		_ = IsExemptFromRateLimit(path)
	})
}

func FuzzSignHS256(f *testing.F) {
	f.Add("user-123", "user@example.com", "admin", "my-issuer", int64(1700000000), int64(1700003600))
	f.Add("", "", "", "", int64(0), int64(0))
	f.Add("sub\x00ject", "e@e.com", "role\nnewline", "iss", int64(-1), int64(-1))
	f.Add("日本語ユーザー", "用户@例え.jp", "管理者", "発行者", int64(1700000000), int64(1800000000))
	f.Add("a]b[c", "a<b>c@d.com", "role\"quote", "iss'quote", int64(1), int64(2))

	f.Fuzz(func(t *testing.T, sub, email, role, issuer string, iat, exp int64) {
		secret := []byte("test-secret-key-32bytes-long!!")
		claims := JWTClaims{
			Subject:   sub,
			Email:     email,
			Role:      role,
			Issuer:    issuer,
			IssuedAt:  iat,
			ExpiresAt: exp,
		}
		token, err := SignHS256(claims, secret)
		if err != nil {
			// JSON marshal errors are acceptable for unusual inputs.
			return
		}
		if token == "" {
			t.Error("SignHS256 returned empty token")
		}
		// Token must have exactly 3 dot-separated parts.
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Errorf("token has %d parts, want 3", len(parts))
		}
	})
}

func FuzzJWTAuthValidation(f *testing.F) {
	f.Add("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.invalid")
	f.Add("")
	f.Add("Bearer ")
	f.Add("Bearer a.b.c")
	f.Add("Basic dXNlcjpwYXNz")
	f.Add("Bearer " + strings.Repeat("x", 10000))
	f.Add("Bearer \x00.\x00.\x00")
	f.Add("Bearer eyJhbGciOiJub25lIn0.eyJzdWIiOiIxMjMifQ.")

	f.Fuzz(func(t *testing.T, authHeader string) {
		secret := []byte("test-secret-key-32bytes-long!!")
		handler := JWTAuth(JWTAuthConfig{
			Secret: secret,
			Issuer: "test-issuer",
		})

		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := handler(inner)

		req := httptest.NewRequest("GET", "/protected", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		rec := httptest.NewRecorder()
		// Must not panic regardless of auth header content.
		wrapped.ServeHTTP(rec, req)

		// Without valid token, should not return 200.
		if authHeader == "" && rec.Code == http.StatusOK {
			t.Error("empty auth header should not produce 200")
		}
	})
}

func FuzzJWTRoundTrip(f *testing.F) {
	f.Add("user-1", "user@test.com", "admin", int64(1700000000))
	f.Add("", "a@b.c", "user", int64(0))
	f.Add("sub", "", "", int64(9999999999))

	f.Fuzz(func(t *testing.T, sub, email, role string, iat int64) {
		if sub == "" {
			return // Subject is required for valid JWTs.
		}
		// Skip inputs with invalid UTF-8 since JSON encoding normalises them.
		if !utf8.ValidString(sub) || !utf8.ValidString(email) || !utf8.ValidString(role) {
			return
		}
		secret := []byte("fuzz-secret-key-for-testing!!!!!")
		exp := iat + 3600
		claims := JWTClaims{
			Subject:   sub,
			Email:     email,
			Role:      role,
			Issuer:    "fuzz-issuer",
			IssuedAt:  iat,
			ExpiresAt: exp,
		}

		token, err := SignHS256(claims, secret)
		if err != nil {
			return
		}

		// Validate by sending through middleware.
		nowFunc := func() time.Time { return time.Unix(iat, 0) }
		handler := JWTAuth(JWTAuthConfig{
			Secret:  secret,
			Issuer:  "fuzz-issuer",
			NowFunc: nowFunc,
		})

		var gotID, gotEmail, gotRole string
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID = GetUserID(r)
			gotEmail = GetUserEmail(r)
			gotRole = GetUserRole(r)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler(inner).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("round-trip failed: status %d for sub=%q", rec.Code, sub)
			return
		}
		if gotID != sub {
			t.Errorf("user ID = %q, want %q", gotID, sub)
		}
		if email != "" && gotEmail != email {
			t.Errorf("email = %q, want %q", gotEmail, email)
		}
		if role != "" && gotRole != role {
			t.Errorf("role = %q, want %q", gotRole, role)
		}
	})
}

func FuzzBruteForceGuard(f *testing.F) {
	f.Add("192.168.1.1", 3)
	f.Add("", 1)
	f.Add("10.0.0.1", 10)
	f.Add("::1", 0)
	f.Add("key\x00null", -1)

	f.Fuzz(func(t *testing.T, ip string, maxFailures int) {
		if maxFailures <= 0 {
			maxFailures = 1
		}
		if maxFailures > 100 {
			maxFailures = 100
		}
		g := NewBruteForceGuard(BruteForceConfig{
			MaxFailures:   maxFailures,
			BlockDuration: time.Minute,
			Window:        time.Minute,
		})
		defer g.Stop()

		// Initially not blocked.
		if g.IsBlocked(ip) {
			t.Error("IP should not be blocked initially")
		}

		// Record failures up to threshold.
		blocked := false
		for range maxFailures {
			if g.RecordFailure(ip) {
				blocked = true
			}
		}
		if !blocked {
			t.Errorf("IP should be blocked after %d failures", maxFailures)
		}
		if !g.IsBlocked(ip) {
			t.Error("IP should be blocked after max failures")
		}

		// Reset should unblock.
		g.Reset(ip)
		if g.IsBlocked(ip) {
			t.Error("IP should not be blocked after Reset")
		}
	})
}

func FuzzRequireRole(f *testing.F) {
	f.Add("admin", "admin", "user-1")
	f.Add("user", "admin", "user-1")
	f.Add("", "admin", "")
	f.Add("admin", "", "user-1")
	f.Add("特殊な役割", "特殊な役割", "user-1")

	f.Fuzz(func(t *testing.T, required, actual, userID string) {
		mw := RequireRole(required)
		var called bool
		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		wrapped := mw(inner)

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := req.Context()
		if userID != "" {
			ctx = SetUserID(ctx, userID)
		}
		if actual != "" {
			ctx = SetUserRole(ctx, actual)
		}
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if userID == "" {
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("no userID: got %d, want 401", rec.Code)
			}
		} else if required != actual {
			if rec.Code != http.StatusForbidden {
				t.Errorf("role mismatch: got %d, want 403", rec.Code)
			}
		} else {
			if !called {
				t.Error("handler not called for matching role")
			}
		}
	})
}

func FuzzRequireSubscriptionTier(f *testing.F) {
	f.Add("free", "free", "user-1")
	f.Add("pro", "starter", "user-1")
	f.Add("concierge", "power", "user-1")
	f.Add("pro", "pro", "user-1")
	f.Add("invalid", "pro", "user-1")
	f.Add("pro", "invalid", "user-1")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, minTier, userTier, userID string) {
		mw := RequireSubscriptionTier(minTier)
		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := mw(inner)

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := req.Context()
		if userID != "" {
			ctx = SetUserID(ctx, userID)
		}
		if userTier != "" {
			ctx = SetSubscriptionTier(ctx, userTier)
		}
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		// Must not panic.
		wrapped.ServeHTTP(rec, req)

		if userID == "" {
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("no userID: got %d, want 401", rec.Code)
			}
		}
	})
}
