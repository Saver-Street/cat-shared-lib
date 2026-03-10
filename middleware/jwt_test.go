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

func TestSetUserID_Overwrite(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetUserID(r.Context(), "first")
	ctx = SetUserID(ctx, "second")
	r = r.WithContext(ctx)
	if id := GetUserID(r); id != "second" {
		t.Errorf("overwritten user ID = %q, want second", id)
	}
}

func TestSetExtCandidateID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtCandidateID(r.Context(), "cand-42"))
	if id := GetExtCandidateID(r); id != "cand-42" {
		t.Errorf("round trip SetExtCandidateID/GetExtCandidateID = %q, want cand-42", id)
	}
}

func TestSetExtCandidateID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtCandidateID(r.Context(), ""))
	if id := GetExtCandidateID(r); id != "" {
		t.Errorf("empty SetExtCandidateID = %q, want empty", id)
	}
}

func TestSetExtCandidateID_Overwrite(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetExtCandidateID(r.Context(), "first")
	ctx = SetExtCandidateID(ctx, "second")
	r = r.WithContext(ctx)
	if id := GetExtCandidateID(r); id != "second" {
		t.Errorf("overwritten ext candidate ID = %q, want second", id)
	}
}

func TestContextChain_MultipleSetters(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetUserID(r.Context(), "uid-1")
	ctx = SetUserRole(ctx, "admin")
	ctx = SetUserEmail(ctx, "admin@test.com")
	ctx = SetExtCandidateID(ctx, "cand-chain")
	r = r.WithContext(ctx)

	if id := GetUserID(r); id != "uid-1" {
		t.Errorf("chained user ID = %q, want uid-1", id)
	}
	if role := GetUserRole(r); role != "admin" {
		t.Errorf("chained role = %q, want admin", role)
	}
	if email := GetUserEmail(r); email != "admin@test.com" {
		t.Errorf("chained email = %q, want admin@test.com", email)
	}
	if cid := GetExtCandidateID(r); cid != "cand-chain" {
		t.Errorf("chained ext candidate ID = %q, want cand-chain", cid)
	}
}

func TestRequireAuth_EmptyStringUserID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler with empty user ID")
	})
	handler := RequireAuth(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(r.Context(), ""))
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("empty user ID: status = %d, want 401", w.Code)
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

func BenchmarkSetExtCandidateID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := r.Context()
	for b.Loop() {
		SetExtCandidateID(ctx, "cand-bench")
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

func BenchmarkGetExtCandidateID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtCandidateID(r.Context(), "cand-bench"))
	for b.Loop() {
		GetExtCandidateID(r)
	}
}

func BenchmarkGetExtTokenID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtTokenID(r.Context(), "tok-bench"))
	for b.Loop() {
		GetExtTokenID(r)
	}
}

func BenchmarkGetExtUserID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtUserID(r.Context(), "ext-user-bench"))
	for b.Loop() {
		GetExtUserID(r)
	}
}

func BenchmarkSetExtTokenID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := r.Context()
	for b.Loop() {
		SetExtTokenID(ctx, "tok-bench")
	}
}

func BenchmarkSetExtUserID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := r.Context()
	for b.Loop() {
		SetExtUserID(ctx, "ext-user-bench")
	}
}

func BenchmarkSetUserEmail(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := r.Context()
	for b.Loop() {
		SetUserEmail(ctx, "bench@example.com")
	}
}

func BenchmarkSetUserRole(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := r.Context()
	for b.Loop() {
		SetUserRole(ctx, "admin")
	}
}

func BenchmarkRequireRole(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := RequireRole("editor")(next)
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetUserID(r.Context(), "u1")
	ctx = SetUserRole(ctx, "editor")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()
	for b.Loop() {
		handler.ServeHTTP(w, r)
	}
}

func TestSetExtUserID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtUserID(r.Context(), "ext-user-99"))
	if id := GetExtUserID(r); id != "ext-user-99" {
		t.Errorf("SetExtUserID/GetExtUserID = %q, want ext-user-99", id)
	}
}

func TestGetExtUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if id := GetExtUserID(r); id != "" {
		t.Errorf("GetExtUserID without value = %q, want empty", id)
	}
}

func TestSetExtTokenID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtTokenID(r.Context(), "tok-abc"))
	if id := GetExtTokenID(r); id != "tok-abc" {
		t.Errorf("SetExtTokenID/GetExtTokenID = %q, want tok-abc", id)
	}
}

func TestGetExtTokenID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if id := GetExtTokenID(r); id != "" {
		t.Errorf("GetExtTokenID without value = %q, want empty", id)
	}
}

func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireRole("moderator")(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(SetUserRole(r.Context(), "moderator"), "user-1"))
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("matching role: status = %d, want 200", w.Code)
	}
}

func TestRequireRole_RejectsWrongRole(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler with wrong role")
	})
	handler := RequireRole("moderator")(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(SetUserRole(r.Context(), "user"), "user-1"))
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("wrong role: status = %d, want 403", w.Code)
	}
}

func TestRequireRole_RejectsUnauthenticated(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler without auth")
	})
	handler := RequireRole("admin")(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated: status = %d, want 401", w.Code)
	}
}
