package migration

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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

func (r *mockRows) Next() bool                                   { r.idx++; return r.idx <= len(r.records) }
func (r *mockRows) Err() error                                   { return r.err }
func (r *mockRows) Close()                                       {}
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
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
	testkit.AssertErrorContains(t, err, "init table")
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
	testkit.AssertLen(t, txs, 2)
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
	testkit.AssertFalse(t, txStarted)
}

func TestMigrate_DuplicateID(t *testing.T) {
	db := &mockDB{}
	r := newRunner(db)
	dups := []Migration{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}}
	err := r.Migrate(context.Background(), dups)
	testkit.AssertErrorIs(t, err, ErrDuplicateID)
}

func TestMigrate_InitError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("db down")
	}}
	r := newRunner(db)
	testkit.AssertError(t, r.Migrate(context.Background(), testMigrations))
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
	testkit.AssertError(t, err)
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
	testkit.AssertError(t, err)
}

func TestMigrate_BeginError(t *testing.T) {
	db := &mockDB{
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			return nil, errors.New("cannot begin tx")
		},
	}
	r := newRunner(db)
	err := r.Migrate(context.Background(), testMigrations)
	testkit.AssertError(t, err)
}

func TestMigrate_CommitError(t *testing.T) {
	db := &mockDB{
		beginFn: func(_ context.Context) (pgx.Tx, error) {
			return &mockTx{commitErr: errors.New("commit failed")}, nil
		},
	}
	r := newRunner(db)
	err := r.Migrate(context.Background(), testMigrations)
	testkit.AssertError(t, err)
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
	testkit.AssertLen(t, txs, 1)
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
	testkit.AssertEqual(t, txCount, 2)
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
	testkit.AssertError(t, err)
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
	testkit.AssertErrorContains(t, err, "no Down SQL")
}

func TestRollback_InitError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("db down")
	}}
	r := newRunner(db)
	testkit.AssertError(t, r.Rollback(context.Background(), testMigrations, 1))
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
	testkit.RequireNoError(t, err)
	testkit.RequireLen(t, records, 1)
	testkit.AssertEqual(t, records[0].ID, 1)
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
	testkit.AssertError(t, err)
}

func TestApplied_RowsErr(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{err: errors.New("rows error")}, nil
		},
	}
	r := newRunner(db)
	_, err := r.Applied(context.Background())
	testkit.AssertError(t, err)
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
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, status[1])
	testkit.AssertFalse(t, status[2])
}

// ---- Pure unit tests (no DB) ----

func TestSortedCopy(t *testing.T) {
	in := []Migration{{ID: 3, Name: "c"}, {ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	out := sortedCopy(in)
	for i, want := range []int{1, 2, 3} {
		testkit.AssertEqual(t, out[i].ID, want)
	}
	testkit.AssertEqual(t, in[0].ID, 3)
}

func TestValidateIDs_OK(t *testing.T) {
	if err := validateIDs([]Migration{{ID: 1}, {ID: 2}, {ID: 3}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateIDs_Duplicate(t *testing.T) {
	err := validateIDs([]Migration{{ID: 1}, {ID: 2}, {ID: 2}})
	testkit.AssertErrorIs(t, err, ErrDuplicateID)
}

func TestValidateIDs_Empty(t *testing.T) {
	if err := validateIDs(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecord_Fields(t *testing.T) {
	now := time.Now()
	r := Record{ID: 1, Name: "init", AppliedAt: now}
	testkit.AssertEqual(t, r.ID, 1)
	testkit.AssertEqual(t, r.Name, "init")
	testkit.AssertTrue(t, r.AppliedAt.Equal(now))
}

func TestSortedCopy_Empty(t *testing.T) {
	testkit.AssertLen(t, sortedCopy(nil), 0)
}

func TestSortStability(t *testing.T) {
	migs := make([]Migration, 10)
	for i := range migs {
		migs[i] = Migration{ID: 10 - i}
	}
	sorted := sortedCopy(migs)
	testkit.AssertTrue(t, sort.SliceIsSorted(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID }))
}

func TestDefaultTable(t *testing.T) {
	testkit.AssertEqual(t, DefaultTable, "schema_migrations")
}

func TestNewWithTable(t *testing.T) {
	r := NewWithDB(&mockDB{}, "custom_migrations")
	testkit.AssertEqual(t, r.table, "custom_migrations")
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
	testkit.AssertError(t, err)
}

func TestRollback_QueryError(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("query failed")
		},
	}
	r := newRunner(db)
	err := r.Rollback(context.Background(), testMigrations, 1)
	testkit.AssertError(t, err)
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
	testkit.AssertError(t, err)
}

func TestStatus_InitError(t *testing.T) {
	db := &mockDB{execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		return pgconn.CommandTag{}, errors.New("down")
	}}
	r := newRunner(db)
	_, err := r.Status(context.Background(), testMigrations)
	testkit.AssertError(t, err)
}

func TestStatus_QueryError(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("query failed")
		},
	}
	r := newRunner(db)
	_, err := r.Status(context.Background(), testMigrations)
	testkit.AssertError(t, err)
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
	testkit.AssertEqual(t, txCount, 1)
}

func TestNew(t *testing.T) {
	r := New(nil)
	testkit.AssertEqual(t, r.table, DefaultTable)
}

func TestNewWithTable_Wrapper(t *testing.T) {
	r := NewWithTable(nil, "my_migrations")
	testkit.AssertEqual(t, r.table, "my_migrations")
}

func TestApplied_InitError(t *testing.T) {
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, errors.New("init failed")
		},
	}
	r := newRunner(db)
	_, err := r.Applied(context.Background())
	testkit.AssertError(t, err)
}

// ---- ValidateMigrations tests ----

func TestValidateMigrations_Valid(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 1, Name: "create_users", Up: "CREATE TABLE users ()", Down: "DROP TABLE users"},
		{ID: 2, Name: "create_posts", Up: "CREATE TABLE posts ()", Down: "DROP TABLE posts"},
	})
	testkit.AssertNoError(t, err)
}

func TestValidateMigrations_Empty(t *testing.T) {
	err := ValidateMigrations(nil)
	testkit.AssertNoError(t, err)

	err = ValidateMigrations([]Migration{})
	testkit.AssertNoError(t, err)
}

func TestValidateMigrations_DuplicateID(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 1, Name: "first", Up: "SELECT 1", Down: "SELECT 1"},
		{ID: 1, Name: "second", Up: "SELECT 2", Down: "SELECT 2"},
	})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrDuplicateID))
}

func TestValidateMigrations_InvalidID(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 0, Name: "zero", Up: "SELECT 1", Down: "SELECT 1"},
	})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrInvalidID))

	err = ValidateMigrations([]Migration{
		{ID: -1, Name: "negative", Up: "SELECT 1", Down: "SELECT 1"},
	})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrInvalidID))
}

func TestValidateMigrations_EmptyName(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 1, Name: "", Up: "SELECT 1", Down: "SELECT 1"},
	})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrEmptyName))
}

func TestValidateMigrations_EmptyUp(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 1, Name: "no_up", Up: "", Down: "SELECT 1"},
	})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrEmptyUp))
}

func TestValidateMigrations_MissingDown(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 1, Name: "no_down", Up: "SELECT 1", Down: ""},
	})
	testkit.AssertError(t, err)
	testkit.AssertTrue(t, errors.Is(err, ErrMissingDown))
}

func TestValidateMigrations_MultipleErrors(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 0, Name: "", Up: "", Down: ""},
	})
	testkit.AssertError(t, err)
	var ve *ValidationError
	testkit.AssertTrue(t, errors.As(err, &ve))
	testkit.AssertTrue(t, len(ve.Errors) >= 3)
}

func TestValidationError_SingleError(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 1, Name: "ok", Up: "SELECT 1", Down: ""},
	})
	testkit.AssertError(t, err)
	var ve *ValidationError
	testkit.AssertTrue(t, errors.As(err, &ve))
	testkit.AssertEqual(t, len(ve.Errors), 1)
	testkit.AssertContains(t, ve.Error(), "missing Down SQL")
}

func TestValidationError_MultipleErrorMessage(t *testing.T) {
	err := ValidateMigrations([]Migration{
		{ID: 0, Name: "", Up: "", Down: ""},
	})
	var ve *ValidationError
	testkit.AssertTrue(t, errors.As(err, &ve))
	testkit.AssertContains(t, ve.Error(), "validation errors")
}

func TestValidationError_Unwrap(t *testing.T) {
	ve := &ValidationError{Errors: []error{ErrInvalidID}}
	testkit.AssertTrue(t, errors.Is(ve, ErrInvalidID))

	empty := &ValidationError{}
	testkit.AssertNil(t, empty.Unwrap())
}
