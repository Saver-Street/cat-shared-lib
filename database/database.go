// Package database provides helpers for PostgreSQL pool setup,
// simple migration running, and transaction management.
package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds configuration for creating a connection pool.
type PoolConfig struct {
	// DSN is the PostgreSQL connection string (required).
	DSN string
	// MaxConns sets the maximum number of connections. Default: 10.
	MaxConns int32
	// MinConns sets the minimum number of idle connections. Default: 2.
	MinConns int32
	// MaxConnLifetime limits how long a connection can be reused. Default: 1 hour.
	MaxConnLifetime time.Duration
	// MaxConnIdleTime limits how long an idle connection is kept. Default: 30 minutes.
	MaxConnIdleTime time.Duration
}

// defaults fills in zero-value fields with sensible defaults.
func (c *PoolConfig) defaults() {
	if c.MaxConns == 0 {
		c.MaxConns = 10
	}
	if c.MinConns == 0 {
		c.MinConns = 2
	}
	if c.MaxConnLifetime == 0 {
		c.MaxConnLifetime = time.Hour
	}
	if c.MaxConnIdleTime == 0 {
		c.MaxConnIdleTime = 30 * time.Minute
	}
}

// NewPool creates a new pgxpool.Pool with the given configuration.
// It pings the database to verify connectivity.
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	cfg.defaults()

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("database: parse config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return pool, nil
}

// Querier abstracts query execution across pool, connection, and transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxFunc is a function that runs within a transaction.
type TxFunc func(tx pgx.Tx) error

// Pool wraps the pgxpool.Pool to add convenience methods.
type Pool interface {
	Querier
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
	Ping(ctx context.Context) error
}

// WithTx executes fn within a database transaction. If fn returns an error
// or panics, the transaction is rolled back; otherwise it is committed.
func WithTx(ctx context.Context, pool Pool, fn TxFunc) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("database: begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			slog.Error("database: rollback failed", "error", rbErr)
			return errors.Join(err, fmt.Errorf("database: rollback: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("database: commit: %w", err)
	}
	return nil
}

// Migration represents a single database migration.
type Migration struct {
	// Version is a unique identifier for ordering (e.g. 1, 2, 3...).
	Version int
	// Name describes the migration.
	Name string
	// SQL is the DDL/DML to execute.
	SQL string
}

// Migrate runs all pending migrations in order. It creates a schema_migrations
// table if it doesn't exist and skips already-applied versions.
func Migrate(ctx context.Context, q Querier, migrations []Migration) error {
	createTable := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INT PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
	if _, err := q.Exec(ctx, createTable); err != nil {
		return fmt.Errorf("database: create migrations table: %w", err)
	}

	// Sort by version
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})

	for _, m := range sorted {
		var exists bool
		err := q.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.Version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("database: check migration %d: %w", m.Version, err)
		}
		if exists {
			continue
		}

		slog.Info("database: applying migration", "version", m.Version, "name", m.Name)
		if _, err := q.Exec(ctx, m.SQL); err != nil {
			return fmt.Errorf("database: apply migration %d (%s): %w", m.Version, m.Name, err)
		}
		if _, err := q.Exec(ctx, "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", m.Version, m.Name); err != nil {
			return fmt.Errorf("database: record migration %d (%s): %w", m.Version, m.Name, err)
		}
	}

	return nil
}
