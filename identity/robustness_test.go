package identity

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestLookupCandidateID_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{scanFunc: func(_ ...any) error {
			return ctx.Err()
		}}
	}}

	id, err := LookupCandidateID(ctx, db, "user-cancel")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if id != "" {
		t.Errorf("id = %q, want empty on cancelled ctx", id)
	}
}

func TestResolveCandidate_ConcurrentLookups(t *testing.T) {
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func() {
			defer wg.Done()
			db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &mockRow{scanFunc: func(dest ...any) error {
					*dest[0].(*string) = "cand-concurrent"
					return nil
				}}
			}}
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(r.Context(), userIDKey, "user-"+string(rune('A'+i%26)))
			r = r.WithContext(ctx)

			id, err := ResolveCandidate(r, db)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if id != "cand-concurrent" {
				t.Errorf("id = %q, want cand-concurrent", id)
			}
		}()
	}
	wg.Wait()
}

func FuzzLookupCandidateID(f *testing.F) {
	f.Add("user-123")
	f.Add("")
	f.Add("a")
	f.Add("user' OR 1=1--")

	f.Fuzz(func(t *testing.T, userID string) {
		db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*string) = "cand-fuzz"
				return nil
			}}
		}}
		// Must not panic
		LookupCandidateID(context.Background(), db, userID)
	})
}

func FuzzGetUserID(f *testing.F) {
	f.Add("user-123")
	f.Add("")
	f.Add("a")

	f.Fuzz(func(t *testing.T, userID string) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		if userID != "" {
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			r = r.WithContext(ctx)
		}
		got := GetUserID(r)
		if userID != "" && got != userID {
			t.Errorf("GetUserID = %q, want %q", got, userID)
		}
	})
}
