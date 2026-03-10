package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

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
