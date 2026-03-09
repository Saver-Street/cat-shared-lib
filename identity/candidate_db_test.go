package identity

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
)

// mockRow implements pgx.Row for testing.
type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error { return m.scanFunc(dest...) }

// mockQuerier implements Querier for testing.
type mockQuerier struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFunc(ctx, sql, args...)
}

// --- LookupCandidateID tests ---

func TestLookupCandidateID_Found(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "cand-abc"
			return nil
		}}
	}}

	id, err := LookupCandidateID(context.Background(), db, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "cand-abc" {
		t.Errorf("LookupCandidateID = %q, want cand-abc", id)
	}
}

func TestLookupCandidateID_NotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(_ ...any) error { return pgx.ErrNoRows }}
	}}

	id, err := LookupCandidateID(context.Background(), db, "user-missing")
	if err == nil {
		t.Fatal("expected error for missing candidate")
	}
	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
	if !errors.Is(err, nil) {
		// The error wraps the message, not pgx.ErrNoRows
		expected := "candidate profile not found for user user-missing"
		if err.Error() != expected {
			t.Errorf("error = %q, want %q", err.Error(), expected)
		}
	}
}

func TestLookupCandidateID_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(_ ...any) error { return errors.New("connection failed") }}
	}}

	id, err := LookupCandidateID(context.Background(), db, "user-err")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "connection failed" {
		t.Errorf("error = %q, want connection failed", err.Error())
	}
	if id != "" {
		t.Errorf("id = %q, want empty on error", id)
	}
}

func TestLookupCandidateID_QuerySQL(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any
	db := &mockQuerier{queryRowFunc: func(_ context.Context, sql string, args ...any) pgx.Row {
		capturedSQL = sql
		capturedArgs = args
		return &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "cand-1"
			return nil
		}}
	}}

	LookupCandidateID(context.Background(), db, "user-42")
	if capturedSQL != "SELECT id FROM candidate_profiles WHERE user_id = $1" {
		t.Errorf("unexpected SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 1 || capturedArgs[0] != "user-42" {
		t.Errorf("unexpected args: %v", capturedArgs)
	}
}

// --- ResolveCandidate with DB tests ---

func TestResolveCandidate_UserID_LookupSuccess(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "cand-from-db"
			return nil
		}}
	}}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), userIDKey, "user-99")
	r = r.WithContext(ctx)

	id, err := ResolveCandidate(r, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "cand-from-db" {
		t.Errorf("ResolveCandidate = %q, want cand-from-db", id)
	}
}

func TestResolveCandidate_UserID_LookupNotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(_ ...any) error { return pgx.ErrNoRows }}
	}}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), userIDKey, "user-no-profile")
	r = r.WithContext(ctx)

	id, err := ResolveCandidate(r, db)
	if err == nil {
		t.Fatal("expected error for missing candidate")
	}
	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
}

func TestResolveCandidate_UserID_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(_ ...any) error { return errors.New("db down") }}
	}}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), userIDKey, "user-err")
	r = r.WithContext(ctx)

	_, err := ResolveCandidate(r, db)
	if err == nil {
		t.Fatal("expected error from DB")
	}
}

func TestResolveCandidate_ExtCandidateIDTakesPriorityOverDB(t *testing.T) {
	// DB should NOT be called when ext candidate ID is set
	dbCalled := false
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		dbCalled = true
		return &mockRow{scanFunc: func(_ ...any) error { return errors.New("should not be called") }}
	}}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), extCandidateIDKey, "ext-cand-1")
	ctx = context.WithValue(ctx, userIDKey, "user-1")
	r = r.WithContext(ctx)

	id, err := ResolveCandidate(r, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "ext-cand-1" {
		t.Errorf("id = %q, want ext-cand-1", id)
	}
	if dbCalled {
		t.Error("DB should not be called when ext candidate ID is present")
	}
}
