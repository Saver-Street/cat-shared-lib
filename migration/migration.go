// Package migration provides a lightweight database migration runner built on
// top of pgx/v5. Migrations are plain SQL strings with an integer ID and a
// name. The runner tracks applied migrations in a dedicated table so each
// migration is applied exactly once.
//
// Usage:
//
//	pool, _ := pgxpool.New(ctx, dsn)
//	runner := migration.New(pool)
//	if err := runner.Migrate(ctx, []migration.Migration{
//	    {ID: 1, Name: "create_users", Up: "CREATE TABLE users (...)", Down: "DROP TABLE users"},
//	}); err != nil {
//	    log.Fatal(err)
//	}
package migration

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultTable is the name of the table used to track applied migrations.
const DefaultTable = "schema_migrations"

// Migration describes a single schema change with optional rollback SQL.
type Migration struct {
	// ID is a unique, monotonically increasing integer (e.g. 1, 2, 3).
	ID int
	// Name is a short human-readable description (e.g. "create_users").
	Name string
	// Up is the SQL to apply this migration.
	Up string
	// Down is the SQL to reverse this migration. Optional.
	Down string
}

// Record is a row from the migrations tracking table.
type Record struct {
	// ID matches Migration.ID.
	ID int
	// Name matches Migration.Name.
	Name string
	// AppliedAt is when the migration was applied.
	AppliedAt time.Time
}

// DB is the minimal interface required by Runner. *pgxpool.Pool satisfies it.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Runner runs and tracks database migrations.
type Runner struct {
	db    DB
	table string
}

// New creates a Runner using the given pool and the default migrations table.
func New(pool *pgxpool.Pool) *Runner {
	return NewWithDB(pool, DefaultTable)
}

// NewWithTable creates a Runner with a custom migrations table name.
func NewWithTable(pool *pgxpool.Pool, table string) *Runner {
	return NewWithDB(pool, table)
}

// NewWithDB creates a Runner using any DB implementation and a custom table name.
// This is useful for testing with mock implementations.
func NewWithDB(db DB, table string) *Runner {
	return &Runner{db: db, table: table}
}

// Init creates the migrations tracking table if it does not already exist.
// It is safe to call multiple times.
func (r *Runner) Init(ctx context.Context) error {
	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id         INTEGER     PRIMARY KEY,
			name       TEXT        NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, r.table)
	if _, err := r.db.Exec(ctx, sql); err != nil {
		return fmt.Errorf("migration: init table: %w", err)
	}
	return nil
}

// Migrate applies any migrations that have not yet been applied, in order of
// ascending ID. It initialises the tracking table automatically.
func (r *Runner) Migrate(ctx context.Context, migrations []Migration) error {
	if err := r.Init(ctx); err != nil {
		return err
	}
	sorted := sortedCopy(migrations)
	if err := validateIDs(sorted); err != nil {
		return err
	}

	applied, err := r.applied(ctx)
	if err != nil {
		return err
	}

	for _, m := range sorted {
		if applied[m.ID] {
			continue
		}
		if err := r.apply(ctx, m); err != nil {
			return fmt.Errorf("migration %d (%s): %w", m.ID, m.Name, err)
		}
	}
	return nil
}

// Rollback rolls back the last n applied migrations in reverse order.
// If n <= 0 all applied migrations are rolled back.
func (r *Runner) Rollback(ctx context.Context, migrations []Migration, n int) error {
	if err := r.Init(ctx); err != nil {
		return err
	}
	records, err := r.Applied(ctx)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	byID := make(map[int]Migration, len(migrations))
	for _, m := range migrations {
		byID[m.ID] = m
	}

	targets := records
	if n > 0 && n < len(targets) {
		targets = targets[len(targets)-n:]
	}
	for i := len(targets) - 1; i >= 0; i-- {
		rec := targets[i]
		m, ok := byID[rec.ID]
		if !ok {
			return fmt.Errorf("migration: no migration found for applied ID %d", rec.ID)
		}
		if m.Down == "" {
			return fmt.Errorf("migration %d (%s): no Down SQL provided", m.ID, m.Name)
		}
		if err := r.rollback(ctx, m); err != nil {
			return fmt.Errorf("migration %d (%s) rollback: %w", m.ID, m.Name, err)
		}
	}
	return nil
}

// Applied returns the list of applied migration records ordered by ID ascending.
func (r *Runner) Applied(ctx context.Context) ([]Record, error) {
	if err := r.Init(ctx); err != nil {
		return nil, err
	}
	return r.applied2(ctx)
}

// Status returns a map from migration ID to whether it has been applied.
func (r *Runner) Status(ctx context.Context, migrations []Migration) (map[int]bool, error) {
	if err := r.Init(ctx); err != nil {
		return nil, err
	}
	applied, err := r.applied(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[int]bool, len(migrations))
	for _, m := range migrations {
		result[m.ID] = applied[m.ID]
	}
	return result, nil
}

// apply runs a single migration inside a transaction and records it.
func (r *Runner) apply(ctx context.Context, m Migration) error {
	return r.inTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, m.Up); err != nil {
			return err
		}
		sql := fmt.Sprintf(
			`INSERT INTO %s (id, name) VALUES ($1, $2)`, r.table)
		_, err := tx.Exec(ctx, sql, m.ID, m.Name)
		return err
	})
}

// rollback undoes a single migration inside a transaction and removes its record.
func (r *Runner) rollback(ctx context.Context, m Migration) error {
	return r.inTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, m.Down); err != nil {
			return err
		}
		sql := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.table)
		_, err := tx.Exec(ctx, sql, m.ID)
		return err
	})
}

// applied returns a set of applied migration IDs.
func (r *Runner) applied(ctx context.Context) (map[int]bool, error) {
	records, err := r.applied2(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[int]bool, len(records))
	for _, rec := range records {
		m[rec.ID] = true
	}
	return m, nil
}

func (r *Runner) applied2(ctx context.Context) ([]Record, error) {
	sql := fmt.Sprintf(
		`SELECT id, name, applied_at FROM %s ORDER BY id ASC`, r.table)
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("migration: query applied: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.AppliedAt); err != nil {
			return nil, fmt.Errorf("migration: scan record: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *Runner) inTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("migration: begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func sortedCopy(migrations []Migration) []Migration {
	cp := make([]Migration, len(migrations))
	copy(cp, migrations)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ID < cp[j].ID })
	return cp
}

// ErrDuplicateID is returned when two migrations share the same ID.
var ErrDuplicateID = errors.New("migration: duplicate migration ID")

// ErrEmptyUp is returned when a migration has no Up SQL.
var ErrEmptyUp = errors.New("migration: empty Up SQL")

// ErrEmptyName is returned when a migration has no name.
var ErrEmptyName = errors.New("migration: empty name")

// ErrInvalidID is returned when a migration has a non-positive ID.
var ErrInvalidID = errors.New("migration: ID must be positive")

// ErrMissingDown is returned when a migration has no Down SQL.
// This is a warning rather than a hard error; callers may choose to ignore it.
var ErrMissingDown = errors.New("migration: missing Down SQL")

// ValidationError collects multiple validation issues found in a migration set.
type ValidationError struct {
	Errors []error
}

// Error returns a summary of all validation errors.
func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 1 {
		return ve.Errors[0].Error()
	}
	return fmt.Sprintf("migration: %d validation errors (first: %v)", len(ve.Errors), ve.Errors[0])
}

// Unwrap returns the first error for compatibility with errors.Is/As.
func (ve *ValidationError) Unwrap() error {
	if len(ve.Errors) == 0 {
		return nil
	}
	return ve.Errors[0]
}

// ValidateMigrations checks a migration set for common issues: non-positive IDs,
// duplicate IDs, empty names, empty Up SQL, and missing Down SQL. Returns nil
// if all migrations are valid. The returned error is a *ValidationError
// containing all issues found.
func ValidateMigrations(migrations []Migration) error {
	var errs []error
	seen := make(map[int]bool, len(migrations))

	for _, m := range migrations {
		if m.ID <= 0 {
			errs = append(errs, fmt.Errorf("%w: %d", ErrInvalidID, m.ID))
		}
		if m.Name == "" {
			errs = append(errs, fmt.Errorf("%w: migration ID %d", ErrEmptyName, m.ID))
		}
		if m.Up == "" {
			errs = append(errs, fmt.Errorf("%w: migration %d (%s)", ErrEmptyUp, m.ID, m.Name))
		}
		if m.Down == "" {
			errs = append(errs, fmt.Errorf("%w: migration %d (%s)", ErrMissingDown, m.ID, m.Name))
		}
		if seen[m.ID] {
			errs = append(errs, fmt.Errorf("%w: %d", ErrDuplicateID, m.ID))
		}
		seen[m.ID] = true
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

func validateIDs(sorted []Migration) error {
	for i := 1; i < len(sorted); i++ {
		if sorted[i].ID == sorted[i-1].ID {
			return fmt.Errorf("%w: %d", ErrDuplicateID, sorted[i].ID)
		}
	}
	return nil
}
