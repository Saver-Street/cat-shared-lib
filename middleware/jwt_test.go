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
