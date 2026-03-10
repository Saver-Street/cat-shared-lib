package identity

import (
	"net/http"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/middleware"
)

func TestGetUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if id := GetUserID(r); id != "" {
		t.Errorf("GetUserID empty context = %q, want empty", id)
	}
}

func TestGetUserID_FromContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-123"))
	if id := GetUserID(r); id != "user-123" {
		t.Errorf("GetUserID = %q, want user-123", id)
	}
}

func TestGetExtCandidateID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if id := GetExtCandidateID(r); id != "" {
		t.Errorf("GetExtCandidateID empty = %q, want empty", id)
	}
}

func TestGetExtCandidateID_FromContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetExtCandidateID(r.Context(), "cand-456"))
	if id := GetExtCandidateID(r); id != "cand-456" {
		t.Errorf("GetExtCandidateID = %q, want cand-456", id)
	}
}

func TestResolveCandidate_NoContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	id, err := ResolveCandidate(r, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != "" {
		t.Errorf("empty context should return empty ID, got %q", id)
	}
}

func TestResolveCandidate_ExtCandidateIDWins(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.SetExtCandidateID(r.Context(), "ext-cand-789")
	ctx = middleware.SetUserID(ctx, "user-111")
	r = r.WithContext(ctx)

	// ext candidate ID takes priority over user lookup
	id, err := ResolveCandidate(r, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != "ext-cand-789" {
		t.Errorf("ResolveCandidate = %q, want ext-cand-789", id)
	}
}

func TestGetUserID_EmptyValue(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetUserID(r.Context(), ""))
	if id := GetUserID(r); id != "" {
		t.Errorf("empty user ID should return empty, got %q", id)
	}
}

func TestGetExtCandidateID_EmptyValue(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetExtCandidateID(r.Context(), ""))
	if id := GetExtCandidateID(r); id != "" {
		t.Errorf("empty ext candidate ID should return empty, got %q", id)
	}
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
