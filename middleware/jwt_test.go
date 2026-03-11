package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestGetUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetUserID(r), "")
}

func TestGetUserID_Set(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserIDKey, "user-abc")
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserID(r), "user-abc")
}

func TestGetUserRole_Set(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserRoleKey, "admin")
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserRole(r), "admin")
}

func TestSetUserID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserID(r.Context(), "u-999"))
	testkit.AssertEqual(t, GetUserID(r), "u-999")
}

func TestRequireAuth_Missing(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAuth(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	handler.ServeHTTP(w, r)
	testkit.AssertStatus(t, w, http.StatusUnauthorized)
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
	testkit.AssertStatus(t, w, http.StatusOK)
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
	testkit.AssertStatus(t, w, http.StatusForbidden)
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
	testkit.AssertStatus(t, w, http.StatusOK)
}

func TestGetUserID_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserIDKey, 12345)
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserID(r), "")
}

func TestGetUserRole_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserRoleKey, 42)
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserRole(r), "")
}

func TestGetUserEmail_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserEmailKey, true)
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserEmail(r), "")
}

func TestGetExtCandidateID_WrongType(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), ExtCandidateIDKey, []byte("bytes"))
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetExtCandidateID(r), "")
}

func TestGetUserRole_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetUserRole(r), "")
}

func TestSetUserID_Overwrite(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetUserID(r.Context(), "first")
	ctx = SetUserID(ctx, "second")
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserID(r), "second")
}

func TestSetExtCandidateID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtCandidateID(r.Context(), "cand-42"))
	testkit.AssertEqual(t, GetExtCandidateID(r), "cand-42")
}

func TestSetExtCandidateID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtCandidateID(r.Context(), ""))
	testkit.AssertEqual(t, GetExtCandidateID(r), "")
}

func TestSetExtCandidateID_Overwrite(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetExtCandidateID(r.Context(), "first")
	ctx = SetExtCandidateID(ctx, "second")
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetExtCandidateID(r), "second")
}

func TestContextChain_MultipleSetters(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := SetUserID(r.Context(), "uid-1")
	ctx = SetUserRole(ctx, "admin")
	ctx = SetUserEmail(ctx, "admin@test.com")
	ctx = SetExtCandidateID(ctx, "cand-chain")
	r = r.WithContext(ctx)

	testkit.AssertEqual(t, GetUserID(r), "uid-1")
	testkit.AssertEqual(t, GetUserRole(r), "admin")
	testkit.AssertEqual(t, GetUserEmail(r), "admin@test.com")
	testkit.AssertEqual(t, GetExtCandidateID(r), "cand-chain")
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
	testkit.AssertStatus(t, w, http.StatusUnauthorized)
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
	testkit.AssertEqual(t, GetExtUserID(r), "ext-user-99")
}

func TestGetExtUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetExtUserID(r), "")
}

func TestSetExtTokenID_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetExtTokenID(r.Context(), "tok-abc"))
	testkit.AssertEqual(t, GetExtTokenID(r), "tok-abc")
}

func TestGetExtTokenID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetExtTokenID(r), "")
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
	testkit.AssertStatus(t, w, http.StatusOK)
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
	testkit.AssertStatus(t, w, http.StatusForbidden)
}

func TestRequireRole_RejectsUnauthenticated(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler without auth")
	})
	handler := RequireRole("admin")(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)
	testkit.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestGetSetSubscriptionTier(t *testing.T) {
	ctx := SetSubscriptionTier(context.Background(), "pro")
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetSubscriptionTier(r), "pro")
}

func TestGetSubscriptionTier_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetSubscriptionTier(r), "")
}

func TestGetSetSubscriptionStatus(t *testing.T) {
	ctx := SetSubscriptionStatus(context.Background(), "active")
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetSubscriptionStatus(r), "active")
}

func TestGetSubscriptionStatus_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetSubscriptionStatus(r), "")
}

func TestSubscriptionTier_AllValues(t *testing.T) {
	tiers := []string{"free", "starter", "pro", "power", "concierge"}
	for _, tier := range tiers {
		ctx := SetSubscriptionTier(context.Background(), tier)
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(ctx)
		testkit.AssertEqual(t, GetSubscriptionTier(r), tier)
	}
}

func TestRequireSubscriptionTier_AllowsSufficientTier(t *testing.T) {
	handler := RequireSubscriptionTier("pro")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, tier := range []string{"pro", "power", "concierge"} {
		ctx := SetSubscriptionTier(context.Background(), tier)
		ctx = context.WithValue(ctx, UserIDKey, "user1")
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		testkit.AssertStatus(t, w, http.StatusOK)
	}
}

func TestRequireSubscriptionTier_RejectsInsufficientTier(t *testing.T) {
	handler := RequireSubscriptionTier("pro")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, tier := range []string{"free", "starter"} {
		ctx := SetSubscriptionTier(context.Background(), tier)
		ctx = context.WithValue(ctx, UserIDKey, "user1")
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		testkit.AssertStatus(t, w, http.StatusForbidden)
	}
}

func TestRequireSubscriptionTier_RejectsUnauthenticated(t *testing.T) {
	handler := RequireSubscriptionTier("free")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	testkit.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestRequireSubscriptionTier_RejectsUnknownTier(t *testing.T) {
	handler := RequireSubscriptionTier("pro")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ctx := SetSubscriptionTier(context.Background(), "enterprise")
	ctx = context.WithValue(ctx, UserIDKey, "user1")
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	testkit.AssertStatus(t, w, http.StatusForbidden)
}
