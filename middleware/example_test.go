package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Saver-Street/cat-shared-lib/middleware"
)

func ExampleGetClientIP() {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	fmt.Println(middleware.GetClientIP(r))
	// Output:
	// 203.0.113.5
}

func ExampleIsExemptFromRateLimit() {
	fmt.Println(middleware.IsExemptFromRateLimit("/assets/main.js"))
	fmt.Println(middleware.IsExemptFromRateLimit("/health"))
	fmt.Println(middleware.IsExemptFromRateLimit("/api/users"))
	// Output:
	// true
	// true
	// false
}

func ExampleSetUserID() {
	r, _ := http.NewRequest("GET", "/", nil)
	ctx := middleware.SetUserID(r.Context(), "user-abc")
	r = r.WithContext(ctx)
	fmt.Println(middleware.GetUserID(r))
	// Output:
	// user-abc
}

func ExampleRequireAuth() {
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Without auth
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// With auth
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, "user-1"))
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)
	// Output:
	// 401
	// 200
}

func ExampleGetUserID() {
	r, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, "user-123")
	r = r.WithContext(ctx)
	fmt.Println(middleware.GetUserID(r))
	// Output:
	// user-123
}

func ExampleSetExtCandidateID() {
	ctx := middleware.SetExtCandidateID(context.Background(), "cand-42")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetExtCandidateID(r))
	// Output: cand-42
}

func ExampleGetExtCandidateID() {
	ctx := middleware.SetExtCandidateID(context.Background(), "cand-99")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetExtCandidateID(r))
	// Output: cand-99
}

func ExampleSetExtTokenID() {
	ctx := middleware.SetExtTokenID(context.Background(), "tok-abc")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetExtTokenID(r))
	// Output: tok-abc
}

func ExampleGetExtTokenID() {
	ctx := middleware.SetExtTokenID(context.Background(), "tok-xyz")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetExtTokenID(r))
	// Output: tok-xyz
}

func ExampleSetExtUserID() {
	ctx := middleware.SetExtUserID(context.Background(), "ext-user-1")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetExtUserID(r))
	// Output: ext-user-1
}

func ExampleGetExtUserID() {
	ctx := middleware.SetExtUserID(context.Background(), "ext-user-2")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetExtUserID(r))
	// Output: ext-user-2
}

func ExampleSetUserEmail() {
	ctx := middleware.SetUserEmail(context.Background(), "user@example.com")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetUserEmail(r))
	// Output: user@example.com
}

func ExampleGetUserEmail() {
	ctx := middleware.SetUserEmail(context.Background(), "admin@test.com")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetUserEmail(r))
	// Output: admin@test.com
}

func ExampleSetUserRole() {
	ctx := middleware.SetUserRole(context.Background(), "admin")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetUserRole(r))
	// Output: admin
}

func ExampleGetUserRole() {
	ctx := middleware.SetUserRole(context.Background(), "moderator")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetUserRole(r))
	// Output: moderator
}

func ExampleRequireAdmin() {
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No auth → 401
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin", nil)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// Non-admin → 403
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/admin", nil)
	ctx := middleware.SetUserID(r.Context(), "user-1")
	ctx = middleware.SetUserRole(ctx, "user")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// Admin → 200
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/admin", nil)
	ctx = middleware.SetUserID(r.Context(), "admin-1")
	ctx = middleware.SetUserRole(ctx, "admin")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)
	// Output:
	// 401
	// 403
	// 200
}

func ExampleChain() {
	// Chain composes middleware so the first argument is the outermost layer.
	addHeader := func(key, value string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(key, value)
				next.ServeHTTP(w, r)
			})
		}
	}

	handler := middleware.Chain(
		addHeader("X-First", "1"),
		addHeader("X-Second", "2"),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	fmt.Println(w.Header().Get("X-First"))
	fmt.Println(w.Header().Get("X-Second"))
	fmt.Println(w.Code)
	// Output:
	// 1
	// 2
	// 200
}

func ExampleRequireRole() {
	handler := middleware.RequireRole("editor")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No auth → 401
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/edit", nil)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// Wrong role → 403
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/edit", nil)
	ctx := middleware.SetUserID(r.Context(), "u1")
	ctx = middleware.SetUserRole(ctx, "viewer")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// Correct role → 200
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/edit", nil)
	ctx = middleware.SetUserID(r.Context(), "u2")
	ctx = middleware.SetUserRole(ctx, "editor")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)
	// Output:
	// 401
	// 403
	// 200
}

func ExampleSetSubscriptionTier() {
	ctx := middleware.SetSubscriptionTier(context.Background(), "pro")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetSubscriptionTier(r))
	// Output:
	// pro
}

func ExampleSetSubscriptionStatus() {
	ctx := middleware.SetSubscriptionStatus(context.Background(), "active")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	fmt.Println(middleware.GetSubscriptionStatus(r))
	// Output:
	// active
}

func ExampleRequireSubscriptionTier() {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.RequireSubscriptionTier("pro")(inner)

	// No subscription tier → 403
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/premium", nil)
	ctx := middleware.SetUserID(r.Context(), "u1")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// Correct tier → 200
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/premium", nil)
	ctx = middleware.SetUserID(r.Context(), "u1")
	ctx = middleware.SetSubscriptionTier(ctx, "pro")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)
	// Output:
	// 403
	// 200
}

func ExampleRecovery() {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)
	// Output: 500
}

func ExampleRequestID() {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r)
		fmt.Println("has id:", id != "")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	// Output: has id: true
}

func ExampleSecureHeaders() {
	handler := middleware.SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	fmt.Println(w.Header().Get("X-Content-Type-Options"))
	fmt.Println(w.Header().Get("X-Frame-Options"))
	// Output:
	// nosniff
	// DENY
}

func ExampleAPIKey() {
	handler := middleware.APIKey("X-API-Key", "secret123")(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	// Valid key.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Println(rec.Code)

	// Invalid key.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-API-Key", "wrong")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	fmt.Println(rec2.Code)
	// Output:
	// 200
	// 401
}

func ExampleAPIKeyMulti() {
	handler := middleware.APIKeyMulti("X-API-Key", []string{"key-a", "key-b"})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "key-b")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

func ExampleTimeout() {
handler := middleware.Timeout(5 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
fmt.Fprintln(w, "ok")
}))

w := httptest.NewRecorder()
r := httptest.NewRequest("GET", "/", nil)
handler.ServeHTTP(w, r)
fmt.Println(w.Code)
// Output: 200
}

func ExampleMaxBody() {
handler := middleware.MaxBody(16)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

w := httptest.NewRecorder()
r := httptest.NewRequest("POST", "/", strings.NewReader("small"))
handler.ServeHTTP(w, r)
fmt.Println(w.Code)
// Output: 200
}
