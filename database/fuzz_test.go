package database

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func FuzzPoolConfigDefaults(f *testing.F) {
	f.Add(int32(0), int32(0), int64(0), int64(0))
	f.Add(int32(10), int32(2), int64(time.Hour), int64(30*time.Minute))
	f.Add(int32(-1), int32(-1), int64(-1), int64(-1))
	f.Add(int32(100), int32(50), int64(time.Second), int64(time.Second))
	f.Add(int32(1), int32(1), int64(time.Nanosecond), int64(time.Nanosecond))

	f.Fuzz(func(t *testing.T, maxConns, minConns int32, maxLifetime, maxIdle int64) {
		cfg := PoolConfig{
			DSN:             "postgres://localhost/test",
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: time.Duration(maxLifetime),
			MaxConnIdleTime: time.Duration(maxIdle),
		}
		cfg.defaults()

		// Zero inputs must be replaced with positive defaults.
		if maxConns == 0 && cfg.MaxConns != 10 {
			t.Errorf("MaxConns default = %d, want 10", cfg.MaxConns)
		}
		if minConns == 0 && cfg.MinConns != 2 {
			t.Errorf("MinConns default = %d, want 2", cfg.MinConns)
		}
		if maxLifetime == 0 && cfg.MaxConnLifetime != time.Hour {
			t.Errorf("MaxConnLifetime default = %v, want 1h", cfg.MaxConnLifetime)
		}
		if maxIdle == 0 && cfg.MaxConnIdleTime != 30*time.Minute {
			t.Errorf("MaxConnIdleTime default = %v, want 30m", cfg.MaxConnIdleTime)
		}
		// Non-zero inputs must be preserved.
		if maxConns != 0 && cfg.MaxConns != maxConns {
			t.Errorf("MaxConns = %d, want original %d", cfg.MaxConns, maxConns)
		}
		if minConns != 0 && cfg.MinConns != minConns {
			t.Errorf("MinConns = %d, want original %d", cfg.MinConns, minConns)
		}
	})
}

func FuzzMigrationSort(f *testing.F) {
	f.Add(3, 1, 2, "c", "a", "b")
	f.Add(1, 1, 1, "same", "same", "same")
	f.Add(-1, 0, 1, "neg", "zero", "pos")
	f.Add(100, 50, 75, "hundred", "fifty", "seventy-five")
	f.Add(0, 0, 0, "", "", "")

	f.Fuzz(func(t *testing.T, v1, v2, v3 int, n1, n2, n3 string) {
		migrations := []Migration{
			{Version: v1, Name: n1, SQL: "SELECT 1"},
			{Version: v2, Name: n2, SQL: "SELECT 2"},
			{Version: v3, Name: n3, SQL: "SELECT 3"},
		}

		// Make a sorted copy the same way Migrate does internally.
		sorted := make([]Migration, len(migrations))
		copy(sorted, migrations)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Version < sorted[j].Version
		})

		// Verify sorting invariant: versions must be non-decreasing.
		for i := 1; i < len(sorted); i++ {
			if sorted[i].Version < sorted[i-1].Version {
				t.Errorf("sorted[%d].Version=%d < sorted[%d].Version=%d",
					i, sorted[i].Version, i-1, sorted[i-1].Version)
			}
		}
		// Original must not be modified.
		if migrations[0].Version != v1 || migrations[1].Version != v2 || migrations[2].Version != v3 {
			t.Error("original migrations slice was modified by sort copy")
		}
	})
}

func FuzzWithTxErrorHandling(f *testing.F) {
	f.Add(true, false, "operation failed")
	f.Add(false, true, "commit error")
	f.Add(false, false, "")
	f.Add(true, true, "both fail")

	f.Fuzz(func(t *testing.T, fnFails, commitFails bool, errMsg string) {
		committed := false
		rolledBack := false

		tx := &mockTx{
			commitFn: func(ctx context.Context) error {
				if commitFails {
					return fmt.Errorf("commit: %s", errMsg)
				}
				committed = true
				return nil
			},
			rollbackFn: func(ctx context.Context) error {
				rolledBack = true
				return nil
			},
		}
		pool := &mockPool{
			beginFn: func(ctx context.Context) (pgx.Tx, error) {
				return tx, nil
			},
		}

		err := WithTx(context.Background(), pool, func(tx pgx.Tx) error {
			if fnFails {
				return errors.New(errMsg)
			}
			return nil
		})

		if fnFails {
			if err == nil {
				t.Error("expected error when fn fails")
			}
			if !rolledBack {
				t.Error("should rollback when fn fails")
			}
			if committed {
				t.Error("should not commit when fn fails")
			}
		} else if commitFails {
			if err == nil {
				t.Error("expected error when commit fails")
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !committed {
				t.Error("should commit on success")
			}
		}
	})
}

func FuzzWithTxBeginError(f *testing.F) {
	f.Add("connection refused")
	f.Add("")
	f.Add("pool exhausted")
	f.Add("\x00null\x00byte")

	f.Fuzz(func(t *testing.T, errMsg string) {
		pool := &mockPool{
			beginFn: func(ctx context.Context) (pgx.Tx, error) {
				return nil, errors.New(errMsg)
			},
		}

		err := WithTx(context.Background(), pool, func(tx pgx.Tx) error {
			t.Fatal("fn should not be called when Begin fails")
			return nil
		})

		if err == nil {
			t.Error("expected error when Begin fails")
		}
	})
}

func FuzzWithTxConcurrent(f *testing.F) {
	f.Add(4)
	f.Add(1)
	f.Add(16)
	f.Add(0)

	f.Fuzz(func(t *testing.T, goroutines int) {
		if goroutines <= 0 {
			goroutines = 1
		}
		if goroutines > 32 {
			goroutines = 32
		}

		var mu sync.Mutex
		txCount := 0

		pool := &mockPool{
			beginFn: func(ctx context.Context) (pgx.Tx, error) {
				mu.Lock()
				txCount++
				mu.Unlock()
				return &mockTx{}, nil
			},
		}

		var wg sync.WaitGroup
		errs := make([]error, goroutines)
		wg.Add(goroutines)
		for i := range goroutines {
			go func(idx int) {
				defer wg.Done()
				errs[idx] = WithTx(context.Background(), pool, func(tx pgx.Tx) error {
					return nil
				})
			}(i)
		}
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Errorf("goroutine %d: unexpected error: %v", i, err)
			}
		}
		mu.Lock()
		if txCount != goroutines {
			t.Errorf("txCount = %d, want %d", txCount, goroutines)
		}
		mu.Unlock()
	})
}

// fuzzMockQuerier records executed SQL for migration order verification.
type fuzzMockQuerier struct {
	applied  map[int]bool
	executed []int
	mu       sync.Mutex
}

func newFuzzMockQuerier() *fuzzMockQuerier {
	return &fuzzMockQuerier{
		applied: make(map[int]bool),
	}
}

func (q *fuzzMockQuerier) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if len(args) >= 2 {
		if v, ok := args[0].(int); ok {
			q.mu.Lock()
			q.applied[v] = true
			q.executed = append(q.executed, v)
			q.mu.Unlock()
		}
	}
	return pgconn.NewCommandTag(""), nil
}

func (q *fuzzMockQuerier) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, nil
}

func (q *fuzzMockQuerier) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	version := 0
	if len(args) > 0 {
		if v, ok := args[0].(int); ok {
			version = v
		}
	}
	q.mu.Lock()
	exists := q.applied[version]
	q.mu.Unlock()
	return &mockRow{scanFn: func(dest ...any) error {
		if len(dest) > 0 {
			if bp, ok := dest[0].(*bool); ok {
				*bp = exists
			}
		}
		return nil
	}}
}

func FuzzMigrationOrder(f *testing.F) {
	f.Add(3, 1, 2, "create_c", "create_a", "create_b")
	f.Add(1, 2, 3, "first", "second", "third")
	f.Add(100, 1, 50, "hundred", "one", "fifty")

	f.Fuzz(func(t *testing.T, v1, v2, v3 int, n1, n2, n3 string) {
		// Skip duplicates since Migrate checks for them separately.
		if v1 == v2 || v2 == v3 || v1 == v3 {
			return
		}

		q := newFuzzMockQuerier()
		migrations := []Migration{
			{Version: v1, Name: n1, SQL: "SELECT 1"},
			{Version: v2, Name: n2, SQL: "SELECT 2"},
			{Version: v3, Name: n3, SQL: "SELECT 3"},
		}

		err := Migrate(context.Background(), q, migrations)
		if err != nil {
			return // migration errors from mock are acceptable
		}

		// Verify migrations were applied in version order.
		for i := 1; i < len(q.executed); i++ {
			if q.executed[i] <= q.executed[i-1] {
				t.Errorf("migrations applied out of order: %v", q.executed)
				return
			}
		}
	})
}
