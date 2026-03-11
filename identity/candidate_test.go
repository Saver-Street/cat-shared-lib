package identity

import (
	"net/http"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/middleware"
	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestGetUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetUserID(r), "")
}

func TestGetUserID_FromContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-123"))
	testkit.AssertEqual(t, GetUserID(r), "user-123")
}

func TestGetExtCandidateID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetExtCandidateID(r), "")
}

func TestGetExtCandidateID_FromContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetExtCandidateID(r.Context(), "cand-456"))
	testkit.AssertEqual(t, GetExtCandidateID(r), "cand-456")
}

func TestResolveCandidate_NoContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	id, err := ResolveCandidate(r, nil)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, id, "")
}

func TestResolveCandidate_ExtCandidateIDWins(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.SetExtCandidateID(r.Context(), "ext-cand-789")
	ctx = middleware.SetUserID(ctx, "user-111")
	r = r.WithContext(ctx)

	// ext candidate ID takes priority over user lookup
	id, err := ResolveCandidate(r, nil)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, id, "ext-cand-789")
}

func TestGetUserID_EmptyValue(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetUserID(r.Context(), ""))
	testkit.AssertEqual(t, GetUserID(r), "")
}

func TestGetExtCandidateID_EmptyValue(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetExtCandidateID(r.Context(), ""))
	testkit.AssertEqual(t, GetExtCandidateID(r), "")
}

func BenchmarkGetUserID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-123"))
	for b.Loop() {
		GetUserID(r)
	}
}

func BenchmarkGetExtCandidateID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetExtCandidateID(r.Context(), "cand-456"))
	for b.Loop() {
		GetExtCandidateID(r)
	}
}

func TestResolveCandidate_NilDB_WithUserID(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-abc"))
	_, err := ResolveCandidate(r, nil)
	testkit.AssertError(t, err)
}
