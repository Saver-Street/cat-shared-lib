package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if id := GetUserID(r); id != "" {
		t.Errorf("empty context = %q, want empty", id)
	}
}

func TestGetUserID_Set(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserIDKey, "user-abc")
	r = r.WithContext(ctx)
	if id := GetUserID(r); id != "user-abc" {
		t.Errorf("GetUserID = %q, want user-abc", id)
	}
}

func TestGetUserRole_Set(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserRoleKey, "admin")
	r = r.WithContext(ctx)
	if role := GetUserRole(r); role != "admin" {
		t.Errorf("GetUserRole = %q, want admin", role)
	}
}

func TestSetUserID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(r.Context(), "u-999"))
	if id := GetUserID(r); id != "u-999" {
		t.Errorf("round trip SetUserID/GetUserID = %q, want u-999", id)
	}
}

func TestRequireAuth_Missing(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAuth(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuth_Present(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAuth(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r = r.WithContext(SetUserID(r.Context(), "user-1"))
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRequireAdmin_NotAdmin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAdmin(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	ctx := SetUserID(r.Context(), "u1")
	ctx = SetUserRole(ctx, "user")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestRequireAdmin_IsAdmin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAdmin(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	ctx := SetUserID(r.Context(), "admin-1")
	ctx = SetUserRole(ctx, "admin")
	r = r.WithContext(ctx)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserIDKey, 12345)
	r = r.WithContext(ctx)
	if id := GetUserID(r); id != "" {
		t.Errorf("non-string context value: GetUserID = %q, want empty", id)
	}
}

func TestGetUserRole_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserRoleKey, 42)
	r = r.WithContext(ctx)
	if role := GetUserRole(r); role != "" {
		t.Errorf("non-string context value: GetUserRole = %q, want empty", role)
	}
}

func TestGetUserEmail_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserEmailKey, true)
	r = r.WithContext(ctx)
	if email := GetUserEmail(r); email != "" {
		t.Errorf("non-string context value: GetUserEmail = %q, want empty", email)
	}
}

func TestGetExtCandidateID_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), ExtCandidateIDKey, []byte("bytes"))
	r = r.WithContext(ctx)
	if id := GetExtCandidateID(r); id != "" {
		t.Errorf("non-string context value: GetExtCandidateID = %q, want empty", id)
	}
}

func TestGetUserRole_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if role := GetUserRole(r); role != "" {
		t.Errorf("empty context role = %q, want empty", role)
	}
}

func BenchmarkGetUserID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(r.Context(), "user-bench"))
	for b.Loop() {
		GetUserID(r)
	}
}

func BenchmarkGetUserRole(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserRole(r.Context(), "admin"))
	for b.Loop() {
		GetUserRole(r)
	}
}

func BenchmarkGetUserEmail(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserEmail(r.Context(), "bench@example.com"))
	for b.Loop() {
		GetUserEmail(r)
	}
}

func BenchmarkSetUserID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := r.Context()
	for b.Loop() {
		SetUserID(ctx, "user-bench")
	}
}

func BenchmarkRequireAuth(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := RequireAuth(next)
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(r.Context(), "u1"))
	w := httptest.NewRecorder()
	for b.Loop() {
		handler.ServeHTTP(w, r)
	}
}

func BenchmarkRequireAdmin(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := RequireAdmin(next)
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetUserID(r.Context(), "u1")
	ctx = SetUserRole(ctx, "admin")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()
	for b.Loop() {
		handler.ServeHTTP(w, r)
	}
}
