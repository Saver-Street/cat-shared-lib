package migration

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func FuzzValidateIDs(f *testing.F) {
	f.Add(1, 2, 3)
	f.Add(1, 1, 2)
	f.Add(0, 0, 0)
	f.Add(-1, 0, 1)
	f.Add(100, 50, 100)
	f.Add(1, 2, 2)
	f.Add(3, 2, 1)
	f.Add(1<<30, 1<<30, 1<<30+1)

	f.Fuzz(func(t *testing.T, id1, id2, id3 int) {
		migrations := []Migration{
			{ID: id1, Name: "m1"},
			{ID: id2, Name: "m2"},
			{ID: id3, Name: "m3"},
		}
		sorted := sortedCopy(migrations)
		err := validateIDs(sorted)

		// Check if there are actual duplicates.
		hasDup := (id1 == id2 || id2 == id3 || id1 == id3)
		if hasDup && err == nil {
			t.Errorf("expected ErrDuplicateID for ids %d,%d,%d", id1, id2, id3)
		}
		if hasDup && err != nil && !errors.Is(err, ErrDuplicateID) {
			t.Errorf("expected ErrDuplicateID, got %v", err)
		}
		if !hasDup && err != nil {
			t.Errorf("unexpected error for unique ids %d,%d,%d: %v", id1, id2, id3, err)
		}
	})
}

func FuzzSortedCopy(f *testing.F) {
	f.Add(3, 1, 2, "c", "a", "b")
	f.Add(1, 1, 1, "same", "same", "same")
	f.Add(-100, 0, 100, "neg", "zero", "pos")
	f.Add(1<<30, -(1 << 30), 0, "big", "small", "zero")

	f.Fuzz(func(t *testing.T, id1, id2, id3 int, n1, n2, n3 string) {
		original := []Migration{
			{ID: id1, Name: n1, Up: "SELECT 1", Down: "SELECT 2"},
			{ID: id2, Name: n2, Up: "SELECT 3", Down: "SELECT 4"},
			{ID: id3, Name: n3, Up: "SELECT 5", Down: "SELECT 6"},
		}
		sorted := sortedCopy(original)

		// Must not modify original.
		if original[0].ID != id1 || original[1].ID != id2 || original[2].ID != id3 {
			t.Error("sortedCopy modified original slice")
		}
		// Result must be sorted by ID ascending.
		if !sort.SliceIsSorted(sorted, func(i, j int) bool {
			return sorted[i].ID < sorted[j].ID
		}) {
			t.Errorf("sortedCopy result not sorted: %d, %d, %d",
				sorted[0].ID, sorted[1].ID, sorted[2].ID)
		}
		// Must preserve all elements.
		if len(sorted) != 3 {
			t.Errorf("sortedCopy length = %d, want 3", len(sorted))
		}
	})
}

func FuzzNewWithTable(f *testing.F) {
	f.Add("schema_migrations")
	f.Add("custom_table")
	f.Add("")
	f.Add("table_with_underscores")
	f.Add("日本語テーブル")
	f.Add("table\x00null")
	f.Add("very_long_" + string(make([]byte, 200)))

	f.Fuzz(func(t *testing.T, table string) {
		// Must not panic.
		r := NewWithDB(&mockDB{}, table)
		testkit.RequireNotNil(t, r)
		if r.table != table {
			t.Errorf("table = %q, want %q", r.table, table)
		}
	})
}

// fuzzDB tracks migration application order for verification.
type fuzzDB struct {
	applied  map[int]bool
	order    []int
	mu       sync.Mutex
	initErr  error
	applyErr error
}

func newFuzzDB() *fuzzDB {
	return &fuzzDB{applied: make(map[int]bool)}
}

func (d *fuzzDB) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	if d.initErr != nil {
		return pgconn.CommandTag{}, d.initErr
	}
	return pgconn.CommandTag{}, nil
}

func (d *fuzzDB) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	records := make([]Record, 0, len(d.applied))
	for id := range d.applied {
		records = append(records, Record{ID: id, Name: "m", AppliedAt: time.Now()})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	return &mockRows{records: records}, nil
}

func (d *fuzzDB) Begin(_ context.Context) (pgx.Tx, error) {
	return &fuzzTx{db: d}, nil
}

type fuzzTx struct {
	db *fuzzDB
}

func (tx *fuzzTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx.db.applyErr != nil {
		return pgconn.CommandTag{}, tx.db.applyErr
	}
	// Track INSERT into migration table.
	if len(args) >= 2 {
		if id, ok := args[0].(int); ok {
			tx.db.mu.Lock()
			tx.db.applied[id] = true
			tx.db.order = append(tx.db.order, id)
			tx.db.mu.Unlock()
		}
	}
	return pgconn.CommandTag{}, nil
}

func (tx *fuzzTx) Commit(_ context.Context) error          { return nil }
func (tx *fuzzTx) Rollback(_ context.Context) error        { return nil }
func (tx *fuzzTx) Begin(_ context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *fuzzTx) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return &mockRows{}, nil
}
func (tx *fuzzTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return &mockRow{}
}
func (tx *fuzzTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *fuzzTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults { return nil }
func (tx *fuzzTx) LargeObjects() pgx.LargeObjects                             { return pgx.LargeObjects{} }
func (tx *fuzzTx) Prepare(_ context.Context, _, _ string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *fuzzTx) Conn() *pgx.Conn { return nil }

func FuzzMigrateOrder(f *testing.F) {
	f.Add(5, 1, 3, "five", "one", "three")
	f.Add(1, 2, 3, "a", "b", "c")
	f.Add(100, 10, 50, "hundred", "ten", "fifty")
	f.Add(-1, 0, 1, "neg", "zero", "pos")

	f.Fuzz(func(t *testing.T, id1, id2, id3 int, n1, n2, n3 string) {
		if id1 == id2 || id2 == id3 || id1 == id3 {
			return // skip duplicates
		}

		db := newFuzzDB()
		r := NewWithDB(db, "test_migrations")

		migrations := []Migration{
			{ID: id1, Name: n1, Up: "SELECT 1"},
			{ID: id2, Name: n2, Up: "SELECT 2"},
			{ID: id3, Name: n3, Up: "SELECT 3"},
		}

		err := r.Migrate(context.Background(), migrations)
		if err != nil {
			return
		}

		// Migrations must be applied in ascending ID order.
		db.mu.Lock()
		defer db.mu.Unlock()
		for i := 1; i < len(db.order); i++ {
			if db.order[i] <= db.order[i-1] {
				t.Errorf("migrations applied out of order: %v", db.order)
				return
			}
		}
	})
}

func FuzzMigrateDuplicateDetection(f *testing.F) {
	f.Add(1, 1, "a", "b")
	f.Add(0, 0, "", "")
	f.Add(-1, -1, "neg", "neg")
	f.Add(100, 100, "cent", "century")

	f.Fuzz(func(t *testing.T, id1, id2 int, n1, n2 string) {
		db := newFuzzDB()
		r := NewWithDB(db, "test_migrations")

		migrations := []Migration{
			{ID: id1, Name: n1, Up: "SELECT 1"},
			{ID: id2, Name: n2, Up: "SELECT 2"},
		}

		err := r.Migrate(context.Background(), migrations)
		if id1 == id2 {
			if !errors.Is(err, ErrDuplicateID) {
				t.Errorf("expected ErrDuplicateID for ids %d,%d, got %v", id1, id2, err)
			}
		}
	})
}

func FuzzRollbackSequence(f *testing.F) {
	f.Add(1, 2, 3, 1)
	f.Add(1, 2, 3, 0)
	f.Add(1, 2, 3, 3)
	f.Add(1, 2, 3, -1)
	f.Add(10, 20, 30, 2)

	f.Fuzz(func(t *testing.T, id1, id2, id3, n int) {
		if id1 == id2 || id2 == id3 || id1 == id3 {
			return
		}
		if n > 100 {
			n = 100
		}

		// First, apply migrations.
		db := newFuzzDB()
		r := NewWithDB(db, "test_migrations")

		migrations := []Migration{
			{ID: id1, Name: "m1", Up: "CREATE TABLE m1(id INT)", Down: "DROP TABLE m1"},
			{ID: id2, Name: "m2", Up: "CREATE TABLE m2(id INT)", Down: "DROP TABLE m2"},
			{ID: id3, Name: "m3", Up: "CREATE TABLE m3(id INT)", Down: "DROP TABLE m3"},
		}

		if err := r.Migrate(context.Background(), migrations); err != nil {
			return
		}

		// Then rollback n migrations — must not panic.
		_ = r.Rollback(context.Background(), migrations, n)
	})
}

func FuzzValidateMigrations(f *testing.F) {
	f.Add(1, "create_users", "CREATE TABLE users()", "DROP TABLE users", 2, "create_posts", "CREATE TABLE posts()", "DROP TABLE posts")
	f.Add(0, "", "", "", -1, "", "", "")
	f.Add(1, "a", "SELECT 1", "", 1, "a", "SELECT 1", "")
	f.Add(1<<30, "big", "UP", "DOWN", -(1 << 30), "neg", "UP2", "DOWN2")

	f.Fuzz(func(t *testing.T, id1 int, n1, u1, d1 string, id2 int, n2, u2, d2 string) {
		migrations := []Migration{
			{ID: id1, Name: n1, Up: u1, Down: d1},
			{ID: id2, Name: n2, Up: u2, Down: d2},
		}
		err := ValidateMigrations(migrations)

		// If all fields are valid and IDs are unique, should pass.
		allValid := id1 > 0 && id2 > 0 && n1 != "" && n2 != "" &&
			u1 != "" && u2 != "" && d1 != "" && d2 != "" && id1 != id2
		if allValid && err != nil {
			t.Errorf("expected nil error for valid migrations, got %v", err)
		}

		// If error, must be a *ValidationError.
		if err != nil {
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Errorf("expected *ValidationError, got %T", err)
			}
		}
	})
}
