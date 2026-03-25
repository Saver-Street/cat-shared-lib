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
