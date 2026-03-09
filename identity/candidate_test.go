package identity

import (
	"context"
	"net/http"
	"testing"
)

func TestGetUserID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if id := GetUserID(r); id != "" {
		t.Errorf("GetUserID empty context = %q, want empty", id)
	}
}

func TestGetUserID_FromContext(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), userIDKey, "user-123")
	r = r.WithContext(ctx)
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
	ctx := context.WithValue(r.Context(), extCandidateIDKey, "cand-456")
	r = r.WithContext(ctx)
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
	ctx := context.WithValue(r.Context(), extCandidateIDKey, "ext-cand-789")
	ctx = context.WithValue(ctx, userIDKey, "user-111")
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
