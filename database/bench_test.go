package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func BenchmarkPoolConfig_Defaults(b *testing.B) {
	for b.Loop() {
		cfg := PoolConfig{DSN: "postgres://localhost/bench"}
		cfg.defaults()
	}
}

func BenchmarkPoolConfig_Defaults_AllSet(b *testing.B) {
	for b.Loop() {
		cfg := PoolConfig{DSN: "postgres://localhost/bench", MaxConns: 20, MinConns: 5}
		cfg.defaults()
	}
}

func BenchmarkWithTx_Success(b *testing.B) {
	ctx := context.Background()
	tx := &mockTx{}
	pool := &mockPool{
		beginFn: func(ctx context.Context) (pgx.Tx, error) { return tx, nil },
	}
	noop := func(tx pgx.Tx) error { return nil }

	for b.Loop() {
		_ = WithTx(ctx, pool, noop)
	}
}

func BenchmarkWithTx_Rollback(b *testing.B) {
	ctx := context.Background()
	fnErr := pgx.ErrNoRows
	tx := &mockTx{}
	pool := &mockPool{
		beginFn: func(ctx context.Context) (pgx.Tx, error) { return tx, nil },
	}
	fail := func(tx pgx.Tx) error { return fnErr }

	for b.Loop() {
		_ = WithTx(ctx, pool, fail)
	}
}

func BenchmarkMigrate_SkipApplied(b *testing.B) {
	ctx := context.Background()
	q := &mockQuerier{
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					if p, ok := dest[0].(*bool); ok {
						*p = true
					}
					return nil
				},
			}
		},
	}
	migrations := []Migration{
		{Version: 1, Name: "init", SQL: "CREATE TABLE t1 (id INT)"},
		{Version: 2, Name: "add_col", SQL: "ALTER TABLE t1 ADD COLUMN name TEXT"},
		{Version: 3, Name: "add_idx", SQL: "CREATE INDEX idx_name ON t1(name)"},
	}

	for b.Loop() {
		_ = Migrate(ctx, q, migrations)
	}
}

func BenchmarkMigrate_ApplyNew(b *testing.B) {
	ctx := context.Background()
	q := &mockQuerier{
		execFn: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("OK 1"), nil
		},
		queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					if p, ok := dest[0].(*bool); ok {
						*p = false
					}
					return nil
				},
			}
		},
	}
	migrations := []Migration{
		{Version: 1, Name: "init", SQL: "CREATE TABLE t1 (id INT)"},
	}

	for b.Loop() {
		_ = Migrate(ctx, q, migrations)
	}
}
