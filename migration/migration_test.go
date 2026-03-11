package migration

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ---- mock DB ----

type mockDB struct {
	execFn  func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	beginFn func(ctx context.Context) (pgx.Tx, error)
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFn != nil {
		return m.execFn(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, sql, args...)
	}
	return &mockRows{}, nil
}

func (m *mockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.beginFn != nil {
		return m.beginFn(ctx)
	}
	return &mockTx{}, nil
}

// mockTx implements pgx.Tx for testing.
type mockTx struct {
	execFn    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	commitErr error
	execCalls []string
}

func (t *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	t.execCalls = append(t.execCalls, sql)
	if t.execFn != nil {
		return t.execFn(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

func (t *mockTx) Commit(ctx context.Context) error   { return t.commitErr }
func (t *mockTx) Rollback(ctx context.Context) error { return nil }

// Satisfy remaining pgx.Tx interface methods with no-ops.
func (t *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return t, nil }
func (t *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return &mockRows{}, nil
}
func (t *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockRow{}
}
func (t *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (t *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (t *mockTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (t *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (t *mockTx) Conn() *pgx.Conn { return nil }

// mockRows implements pgx.Rows with configurable records.
type mockRows struct {
	records []Record
	idx     int
	err     error
	scanErr error
}

func (r *mockRows) Next() bool        { r.idx++; return r.idx <= len(r.records) }
func (r *mockRows) Err() error        { return r.err }
func (r *mockRows) Close()            {}
func (r *mockRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }
func (r *mockRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	rec := r.records[r.idx-1]
	if len(dest) >= 1 {
		*(dest[0].(*int)) = rec.ID
	}
	if len(dest) >= 2 {
		*(dest[1].(*string)) = rec.Name
	}
	if len(dest) >= 3 {
		*(dest[2].(*time.Time)) = rec.AppliedAt
	}
	return nil
}

type mockRow struct{}

func (r *mockRow) Scan(dest ...any) error { return nil }

// ---- helpers ----

func newRunner(db DB) *Runner {
	return NewWithDB(db, "schema_migrations")
}

var testMigrations = []Migration{
	{ID: 1, Name: "create_users", Up: "CREATE TABLE users(id INT)", Down: "DROP TABLE users"},
	{ID: 2, Name: "add_email", Up: "ALTER TABLE users ADD email TEXT", Down: "ALTER TABLE users DROP COLUMN email"},
}

// ---- Init tests ----

func TestInit_OK(t *testing.T) {
	db := &mockDB{}
	r := newRunner(db)
	if err := r.Init(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInit_ExecError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("db error")
	}}
	r := newRunner(db)
	err := r.Init(context.Background())
	if err == nil || !containsStr(err.Error(), "init table") {
		t.Fatalf("expected init table error, got %v", err)
	}
}

// ---- Migrate tests ----

func TestMigrate_Empty(t *testing.T) {
	db := &mockDB{}
	r := newRunner(db)
	if err := r.Migrate(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrate_AllNew(t *testing.T) {
	var txs []*mockTx
	db := &mockDB{
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			tx := &mockTx{}
			txs = append(txs, tx)
			return tx, nil
		},
	}
	r := newRunner(db)
	if err := r.Migrate(context.Background(), testMigrations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}
}

func TestMigrate_AlreadyApplied(t *testing.T) {
	now := time.Now()
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{
				{ID: 1, Name: "create_users", AppliedAt: now},
				{ID: 2, Name: "add_email", AppliedAt: now},
			}}, nil
		},
	}
	var txStarted bool
	db.beginFn = func(_ context.Context) (pgx.Tx, error) {
		txStarted = true
		return &mockTx{}, nil
	}
	r := newRunner(db)
	if err := r.Migrate(context.Background(), testMigrations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txStarted {
		t.Error("no transactions expected when all migrations are applied")
	}
}

func TestMigrate_DuplicateID(t *testing.T) {
	db := &mockDB{}
	r := newRunner(db)
	dups := []Migration{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}}
	err := r.Migrate(context.Background(), dups)
	if !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("expected ErrDuplicateID, got %v", err)
	}
}

func TestMigrate_InitError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("db down")
	}}
	r := newRunner(db)
	if err := r.Migrate(context.Background(), testMigrations); err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrate_QueryError(t *testing.T) {
	callCount := 0
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			callCount++
			return nil, errors.New("query failed")
		},
	}
	r := newRunner(db)
	err := r.Migrate(context.Background(), testMigrations)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrate_ApplyError(t *testing.T) {
	db := &mockDB{
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			return &mockTx{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("exec failed")
			}}, nil
		},
	}
	r := newRunner(db)
	err := r.Migrate(context.Background(), testMigrations)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrate_BeginError(t *testing.T) {
	db := &mockDB{
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			return nil, errors.New("cannot begin tx")
		},
	}
	r := newRunner(db)
	err := r.Migrate(context.Background(), testMigrations)
	if err == nil {
		t.Fatal("expected error from Begin")
	}
}

func TestMigrate_CommitError(t *testing.T) {
	db := &mockDB{
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			return &mockTx{commitErr: errors.New("commit failed")}, nil
		},
	}
	r := newRunner(db)
	err := r.Migrate(context.Background(), testMigrations)
	if err == nil {
		t.Fatal("expected commit error")
	}
}

// ---- Rollback tests ----

func TestRollback_NoApplied(t *testing.T) {
	db := &mockDB{}
	r := newRunner(db)
	if err := r.Rollback(context.Background(), testMigrations, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRollback_LastOne(t *testing.T) {
	now := time.Now()
	var txs []*mockTx
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{
				{ID: 1, Name: "create_users", AppliedAt: now},
				{ID: 2, Name: "add_email", AppliedAt: now},
			}}, nil
		},
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			tx := &mockTx{}
			txs = append(txs, tx)
			return tx, nil
		},
	}
	r := newRunner(db)
	if err := r.Rollback(context.Background(), testMigrations, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 tx, got %d", len(txs))
	}
}

func TestRollback_All(t *testing.T) {
	now := time.Now()
	var txCount int
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{
				{ID: 1, Name: "create_users", AppliedAt: now},
				{ID: 2, Name: "add_email", AppliedAt: now},
			}}, nil
		},
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			txCount++
			return &mockTx{}, nil
		},
	}
	r := newRunner(db)
	if err := r.Rollback(context.Background(), testMigrations, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txCount != 2 {
		t.Errorf("expected 2 rollback txs, got %d", txCount)
	}
}

func TestRollback_MissingMigration(t *testing.T) {
	now := time.Now()
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{{ID: 99, Name: "unknown", AppliedAt: now}}}, nil
		},
	}
	r := newRunner(db)
	err := r.Rollback(context.Background(), testMigrations, 0)
	if err == nil {
		t.Fatal("expected error for missing migration ID")
	}
}

func TestRollback_NoDownSQL(t *testing.T) {
	now := time.Now()
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{{ID: 1, Name: "m1", AppliedAt: now}}}, nil
		},
	}
	r := newRunner(db)
	noDown := []Migration{{ID: 1, Name: "m1", Up: "CREATE TABLE x(id INT)", Down: ""}}
	err := r.Rollback(context.Background(), noDown, 0)
	if err == nil || !containsStr(err.Error(), "no Down SQL") {
		t.Fatalf("expected 'no Down SQL' error, got %v", err)
	}
}

func TestRollback_InitError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("db down")
	}}
	r := newRunner(db)
	if err := r.Rollback(context.Background(), testMigrations, 1); err == nil {
		t.Fatal("expected error")
	}
}

// ---- Applied/Status tests ----

func TestApplied_OK(t *testing.T) {
	now := time.Now()
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{
				{ID: 1, Name: "a", AppliedAt: now},
			}}, nil
		},
	}
	r := newRunner(db)
	records, err := r.Applied(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 || records[0].ID != 1 {
		t.Errorf("unexpected records: %v", records)
	}
}

func TestApplied_ScanError(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{
				records: []Record{{ID: 1, Name: "a", AppliedAt: time.Now()}},
				scanErr: errors.New("scan failed"),
			}, nil
		},
	}
	r := newRunner(db)
	_, err := r.Applied(context.Background())
	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestApplied_RowsErr(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{err: errors.New("rows error")}, nil
		},
	}
	r := newRunner(db)
	_, err := r.Applied(context.Background())
	if err == nil {
		t.Fatal("expected rows error")
	}
}

func TestStatus_OK(t *testing.T) {
	now := time.Now()
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{{ID: 1, Name: "a", AppliedAt: now}}}, nil
		},
	}
	r := newRunner(db)
	status, err := r.Status(context.Background(), testMigrations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status[1] {
		t.Error("expected migration 1 to be applied")
	}
	if status[2] {
		t.Error("expected migration 2 to be unapplied")
	}
}

// ---- Pure unit tests (no DB) ----

func TestSortedCopy(t *testing.T) {
	in := []Migration{{ID: 3, Name: "c"}, {ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	out := sortedCopy(in)
	for i, want := range []int{1, 2, 3} {
		if out[i].ID != want {
			t.Errorf("position %d: got %d, want %d", i, out[i].ID, want)
		}
	}
	if in[0].ID != 3 {
		t.Error("sortedCopy must not modify original")
	}
}

func TestValidateIDs_OK(t *testing.T) {
	if err := validateIDs([]Migration{{ID: 1}, {ID: 2}, {ID: 3}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateIDs_Duplicate(t *testing.T) {
	err := validateIDs([]Migration{{ID: 1}, {ID: 2}, {ID: 2}})
	if !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("expected ErrDuplicateID, got %v", err)
	}
}

func TestValidateIDs_Empty(t *testing.T) {
	if err := validateIDs(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecord_Fields(t *testing.T) {
	now := time.Now()
	r := Record{ID: 1, Name: "init", AppliedAt: now}
	if r.ID != 1 || r.Name != "init" || !r.AppliedAt.Equal(now) {
		t.Error("Record field assignment failed")
	}
}

func TestSortedCopy_Empty(t *testing.T) {
	if len(sortedCopy(nil)) != 0 {
		t.Error("expected empty")
	}
}

func TestSortStability(t *testing.T) {
	migs := make([]Migration, 10)
	for i := range migs {
		migs[i] = Migration{ID: 10 - i}
	}
	sorted := sortedCopy(migs)
	if !sort.SliceIsSorted(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID }) {
		t.Error("sortedCopy result is not sorted")
	}
}

func TestDefaultTable(t *testing.T) {
	if DefaultTable != "schema_migrations" {
		t.Errorf("unexpected DefaultTable %q", DefaultTable)
	}
}

func TestNewWithTable(t *testing.T) {
	r := NewWithDB(&mockDB{}, "custom_migrations")
	if r.table != "custom_migrations" {
		t.Errorf("expected custom_migrations, got %q", r.table)
	}
}

func TestApplied_QueryError(t *testing.T) {
	execCalls := 0
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			execCalls++
			return pgconn.CommandTag{}, nil
		},
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("query failed")
		},
	}
	r := newRunner(db)
	_, err := r.Applied(context.Background())
	if err == nil {
		t.Fatal("expected query error")
	}
}

func TestRollback_QueryError(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("query failed")
		},
	}
	r := newRunner(db)
	err := r.Rollback(context.Background(), testMigrations, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRollback_RollbackExecError(t *testing.T) {
	now := time.Now()
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{{ID: 1, Name: "create_users", AppliedAt: now}}}, nil
		},
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			return &mockTx{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("exec failed")
			}}, nil
		},
	}
	r := newRunner(db)
	err := r.Rollback(context.Background(), testMigrations, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStatus_InitError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("down")
	}}
	r := newRunner(db)
	_, err := r.Status(context.Background(), testMigrations)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStatus_QueryError(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("query failed")
		},
	}
	r := newRunner(db)
	_, err := r.Status(context.Background(), testMigrations)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrate_PartiallyApplied(t *testing.T) {
	now := time.Now()
	var txCount int
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{records: []Record{{ID: 1, Name: "create_users", AppliedAt: now}}}, nil
		},
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			txCount++
			return &mockTx{}, nil
		},
	}
	r := newRunner(db)
	if err := r.Migrate(context.Background(), testMigrations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txCount != 1 {
		t.Errorf("expected 1 transaction (only migration 2), got %d", txCount)
	}
}

// helper
func containsStr(s, substr string) bool {
	return s != "" && len(s) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}

func TestNew(t *testing.T) {
	r := New(nil)
	if r.table != DefaultTable {
		t.Errorf("expected table %q, got %q", DefaultTable, r.table)
	}
}

func TestNewWithTable_Wrapper(t *testing.T) {
	r := NewWithTable(nil, "my_migrations")
	if r.table != "my_migrations" {
		t.Errorf("expected table %q, got %q", "my_migrations", r.table)
	}
}

func TestApplied_InitError(t *testing.T) {
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, errors.New("init failed")
		},
	}
	r := newRunner(db)
	_, err := r.Applied(context.Background())
	if err == nil {
		t.Fatal("expected init error to propagate")
	}
}

