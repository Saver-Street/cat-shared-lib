package identity

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/middleware"
	"github.com/Saver-Street/cat-shared-lib/testkit"
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
	testkit.AssertErrorContains(t, err, "candidate profile not found for user user-missing")
}

func TestLookupCandidateID_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(_ ...any) error { return errors.New("connection failed") }}
	}}

	id, err := LookupCandidateID(context.Background(), db, "user-err")
	if err == nil {
		t.Fatal("expected error")
	}
	testkit.AssertErrorContains(t, err, "connection failed")
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
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-99"))

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
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-no-profile"))

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
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-err"))

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
	ctx := middleware.SetExtCandidateID(r.Context(), "ext-cand-1")
	ctx = middleware.SetUserID(ctx, "user-1")
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

func BenchmarkLookupCandidateID(b *testing.B) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "cand-bench"
			return nil
		}}
	}}
	ctx := context.Background()
	for b.Loop() {
		LookupCandidateID(ctx, db, "user-bench")
	}
}

func BenchmarkResolveCandidate_ExtID(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(middleware.SetExtCandidateID(r.Context(), "ext-cand-1"))
	for b.Loop() {
		ResolveCandidate(r, nil)
	}
}
