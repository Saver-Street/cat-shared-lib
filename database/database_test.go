package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockQuerier implements Querier for testing.
type mockQuerier struct {
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFn != nil {
		return m.execFn(ctx, sql, args...)
	}
	return pgconn.NewCommandTag(""), nil
}

func (m *mockQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFn != nil {
		return m.queryRowFn(ctx, sql, args...)
	}
	return &mockRow{}
}

// mockRow implements pgx.Row for testing.
type mockRow struct {
	scanFn func(dest ...any) error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.scanFn != nil {
		return r.scanFn(dest...)
	}
	return nil
}

// mockTx implements pgx.Tx for testing.
type mockTx struct {
	commitFn   func(ctx context.Context) error
	rollbackFn func(ctx context.Context) error
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (t *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return t, nil }
func (t *mockTx) Conn() *pgx.Conn                           { return nil }
func (t *mockTx) CopyFrom(ctx context.Context, tn pgx.Identifier, cn []string, rs pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (t *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (t *mockTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (t *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (t *mockTx) Commit(ctx context.Context) error {
	if t.commitFn != nil {
		return t.commitFn(ctx)
	}
	return nil
}
func (t *mockTx) Rollback(ctx context.Context) error {
	if t.rollbackFn != nil {
		return t.rollbackFn(ctx)
	}
	return nil
}
func (t *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if t.execFn != nil {
		return t.execFn(ctx, sql, args...)
	}
	return pgconn.NewCommandTag(""), nil
}
func (t *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if t.queryFn != nil {
		return t.queryFn(ctx, sql, args...)
	}
	return nil, nil
}
func (t *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if t.queryRowFn != nil {
		return t.queryRowFn(ctx, sql, args...)
	}
	return &mockRow{}
}

// mockPool implements Pool for testing.
type mockPool struct {
	beginFn func(ctx context.Context) (pgx.Tx, error)
	closeFn func()
	pingFn  func(ctx context.Context) error
	mockQuerier
}

func (p *mockPool) Begin(ctx context.Context) (pgx.Tx, error) {
	if p.beginFn != nil {
		return p.beginFn(ctx)
	}
	return &mockTx{}, nil
}
func (p *mockPool) Close() {
	if p.closeFn != nil {
		p.closeFn()
	}
}
func (p *mockPool) Ping(ctx context.Context) error {
	if p.pingFn != nil {
		return p.pingFn(ctx)
	}
	return nil
}

func TestPoolConfig_Defaults(t *testing.T) {
	cfg := PoolConfig{DSN: "postgres://localhost/test"}
	cfg.defaults()

	if cfg.MaxConns != 10 {
		t.Errorf("MaxConns = %d, want 10", cfg.MaxConns)
	}
	if cfg.MinConns != 2 {
		t.Errorf("MinConns = %d, want 2", cfg.MinConns)
	}
}

func TestPoolConfig_NoOverrideNonZero(t *testing.T) {
	cfg := PoolConfig{DSN: "x", MaxConns: 20, MinConns: 5}
	cfg.defaults()

	if cfg.MaxConns != 20 {
		t.Errorf("MaxConns = %d, want 20", cfg.MaxConns)
	}
	if cfg.MinConns != 5 {
		t.Errorf("MinConns = %d, want 5", cfg.MinConns)
	}
}

func TestWithTx_Success(t *testing.T) {
	committed := false
	tx := &mockTx{
		commitFn: func(ctx context.Context) error {
			committed = true
			return nil
		},
	}
	pool := &mockPool{
		beginFn: func(ctx context.Context) (pgx.Tx, error) {
			return tx, nil
		},
	}

	err := WithTx(context.Background(), pool, func(tx pgx.Tx) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !committed {
		t.Error("transaction should have been committed")
	}
}

func TestWithTx_FnError(t *testing.T) {
	rolledBack := false
	tx := &mockTx{
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
		return fmt.Errorf("operation failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !rolledBack {
		t.Error("transaction should have been rolled back")
	}
}

func TestWithTx_Panic(t *testing.T) {
	rolledBack := false
	tx := &mockTx{
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

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to propagate")
		}
		if !rolledBack {
			t.Error("transaction should have been rolled back on panic")
		}
	}()

	_ = WithTx(context.Background(), pool, func(tx pgx.Tx) error {
		panic("test panic")
	})
}

func TestWithTx_BeginError(t *testing.T) {
	pool := &mockPool{
		beginFn: func(ctx context.Context) (pgx.Tx, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	err := WithTx(context.Background(), pool, func(tx pgx.Tx) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error from Begin")
	}
}

func TestWithTx_CommitError(t *testing.T) {
	tx := &mockTx{
		commitFn: func(ctx context.Context) error {
			return fmt.Errorf("commit failed")
		},
	}
	pool := &mockPool{
		beginFn: func(ctx context.Context) (pgx.Tx, error) {
			return tx, nil
		},
	}

	err := WithTx(context.Background(), pool, func(tx pgx.Tx) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error from Commit")
	}
}

func TestWithTx_RollbackError(t *testing.T) {
	tx := &mockTx{
		rollbackFn: func(ctx context.Context) error {
			return fmt.Errorf("rollback failed")
		},
	}
	pool := &mockPool{
		beginFn: func(ctx context.Context) (pgx.Tx, error) {
			return tx, nil
		},
	}

	err := WithTx(context.Background(), pool, func(tx pgx.Tx) error {
		return fmt.Errorf("fn error")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrate_CreateTableError(t *testing.T) {
	q := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag(""), fmt.Errorf("exec error")
		},
	}

	err := Migrate(context.Background(), q, nil)
	if err == nil {
		t.Fatal("expected error from create table")
	}
}

func TestMigrate_EmptyList(t *testing.T) {
	q := &mockQuerier{}
	err := Migrate(context.Background(), q, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrate_AppliesMigrations(t *testing.T) {
	applied := map[int]bool{}
	insertedVersions := map[int]bool{}

	q := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if len(args) == 2 {
				// INSERT INTO schema_migrations
				v, ok := args[0].(int)
				if ok {
					insertedVersions[v] = true
				}
			}
			return pgconn.NewCommandTag(""), nil
		},
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			v := args[0].(int)
			return &mockRow{
				scanFn: func(dest ...any) error {
					b, ok := dest[0].(*bool)
					if ok {
						*b = applied[v]
					}
					return nil
				},
			}
		},
	}

	migrations := []Migration{
		{Version: 2, Name: "add_users", SQL: "CREATE TABLE users (id INT)"},
		{Version: 1, Name: "init", SQL: "CREATE TABLE config (key TEXT)"},
	}

	err := Migrate(context.Background(), q, migrations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !insertedVersions[1] || !insertedVersions[2] {
		t.Errorf("expected both migrations applied, got %v", insertedVersions)
	}
}

func TestMigrate_SkipsApplied(t *testing.T) {
	q := &mockQuerier{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					b, ok := dest[0].(*bool)
					if ok {
						*b = true // already applied
					}
					return nil
				},
			}
		},
	}

	migrations := []Migration{
		{Version: 1, Name: "init", SQL: "CREATE TABLE config (key TEXT)"},
	}

	err := Migrate(context.Background(), q, migrations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrate_CheckError(t *testing.T) {
	q := &mockQuerier{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					return fmt.Errorf("query error")
				},
			}
		},
	}

	err := Migrate(context.Background(), q, []Migration{{Version: 1, Name: "init", SQL: "SELECT 1"}})
	if err == nil {
		t.Fatal("expected error from check query")
	}
}

func TestMigrate_ExecError(t *testing.T) {
	callCount := 0
	q := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			callCount++
			if callCount > 1 {
				return pgconn.NewCommandTag(""), fmt.Errorf("migration exec error")
			}
			return pgconn.NewCommandTag(""), nil
		},
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					b, ok := dest[0].(*bool)
					if ok {
						*b = false
					}
					return nil
				},
			}
		},
	}

	err := Migrate(context.Background(), q, []Migration{{Version: 1, Name: "init", SQL: "BAD SQL"}})
	if err == nil {
		t.Fatal("expected error from exec")
	}
}

func TestMigrate_RecordError(t *testing.T) {
	callCount := 0
	q := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			callCount++
			if callCount == 3 {
				// Third exec is the INSERT into schema_migrations
				return pgconn.NewCommandTag(""), fmt.Errorf("insert error")
			}
			return pgconn.NewCommandTag(""), nil
		},
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					b, ok := dest[0].(*bool)
					if ok {
						*b = false
					}
					return nil
				},
			}
		},
	}

	err := Migrate(context.Background(), q, []Migration{{Version: 1, Name: "init", SQL: "CREATE TABLE x (id INT)"}})
	if err == nil {
		t.Fatal("expected error from record")
	}
}

func TestNewPool_BadDSN(t *testing.T) {
	// An empty DSN should fail to parse
	_, err := NewPool(context.Background(), PoolConfig{DSN: "://invalid"})
	if err == nil {
		t.Fatal("expected error for bad DSN")
	}
}

func TestNewPool_ConnectionRefused(t *testing.T) {
	// Valid DSN but unreachable host - tests the Ping failure path
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPool(ctx, PoolConfig{DSN: "postgres://localhost:1/nonexistent?connect_timeout=1"})
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}
