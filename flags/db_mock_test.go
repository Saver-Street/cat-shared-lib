package flags

import (
	"context"
	"errors"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
	"github.com/jackc/pgx/v5"
)

// mockRow implements pgx.Row for testing.
type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error { return m.scanFunc(dest...) }

// mockQuerier implements Querier for testing.
type mockQuerier struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFunc(ctx, sql, args...)
}

func newValueRow(val string) *mockRow {
	return &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*string) = val
		return nil
	}}
}

func newNoRowsRow() *mockRow {
	return &mockRow{scanFunc: func(_ ...any) error { return pgx.ErrNoRows }}
}

func newErrorRow(msg string) *mockRow {
	return &mockRow{scanFunc: func(_ ...any) error { return errors.New(msg) }}
}

// --- IsFeatureEnabled tests ---

func TestIsFeatureEnabled_NilQuerier(t *testing.T) {
	testkit.AssertTrue(t, IsFeatureEnabled(context.Background(), nil, FlagAIScoring))
}

func TestIsFeatureEnabled_FlagTrue(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	testkit.AssertTrue(t, IsFeatureEnabled(context.Background(), db, FlagAIScoring))
}

func TestIsFeatureEnabled_FlagFalse(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	testkit.AssertFalse(t, IsFeatureEnabled(context.Background(), db, FlagAIScoring))
}

func TestIsFeatureEnabled_FlagNotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newNoRowsRow()
	}}
	testkit.AssertTrue(t, IsFeatureEnabled(context.Background(), db, "nonexistent"))
}

func TestIsFeatureEnabled_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newErrorRow("connection lost")
	}}
	testkit.AssertTrue(t, IsFeatureEnabled(context.Background(), db, FlagResumeParsing))
}

func TestIsFeatureEnabled_FlagKeyPrefix(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any
	db := &mockQuerier{queryRowFunc: func(_ context.Context, sql string, args ...any) pgx.Row {
		capturedSQL = sql
		capturedArgs = args
		return newValueRow("true")
	}}
	IsFeatureEnabled(context.Background(), db, "myFlag")
	testkit.AssertEqual(t, capturedSQL, "SELECT value FROM site_settings WHERE key = $1")
	testkit.AssertLen(t, capturedArgs, 1)
	testkit.AssertEqual(t, capturedArgs[0], "flag_myFlag")
}

func TestIsFeatureEnabled_EmptyValue(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("")
	}}
	testkit.AssertFalse(t, IsFeatureEnabled(context.Background(), db, FlagSIAFI))
}

// --- IsMaintenanceModeActive tests ---

func TestIsMaintenanceModeActive_NilQuerier(t *testing.T) {
	testkit.AssertFalse(t, IsMaintenanceModeActive(context.Background(), nil))
}

func TestIsMaintenanceModeActive_True(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	testkit.AssertTrue(t, IsMaintenanceModeActive(context.Background(), db))
}

func TestIsMaintenanceModeActive_False(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	testkit.AssertFalse(t, IsMaintenanceModeActive(context.Background(), db))
}

func TestIsMaintenanceModeActive_NotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newNoRowsRow()
	}}
	testkit.AssertFalse(t, IsMaintenanceModeActive(context.Background(), db))
}

func TestIsMaintenanceModeActive_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newErrorRow("timeout")
	}}
	testkit.AssertFalse(t, IsMaintenanceModeActive(context.Background(), db))
}

// --- IsGlobalAutomationPaused tests ---

func TestIsGlobalAutomationPaused_NilQuerier(t *testing.T) {
	testkit.AssertFalse(t, IsGlobalAutomationPaused(context.Background(), nil))
}

func TestIsGlobalAutomationPaused_True(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	testkit.AssertTrue(t, IsGlobalAutomationPaused(context.Background(), db))
}

func TestIsGlobalAutomationPaused_False(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	testkit.AssertFalse(t, IsGlobalAutomationPaused(context.Background(), db))
}

func TestIsGlobalAutomationPaused_NotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newNoRowsRow()
	}}
	testkit.AssertFalse(t, IsGlobalAutomationPaused(context.Background(), db))
}

func TestIsGlobalAutomationPaused_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newErrorRow("connection reset")
	}}
	testkit.AssertFalse(t, IsGlobalAutomationPaused(context.Background(), db))
}

func BenchmarkIsFeatureEnabled(b *testing.B) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	ctx := context.Background()
	for b.Loop() {
		IsFeatureEnabled(ctx, db, FlagAIScoring)
	}
}

func BenchmarkIsMaintenanceModeActive(b *testing.B) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	ctx := context.Background()
	for b.Loop() {
		IsMaintenanceModeActive(ctx, db)
	}
}

func BenchmarkIsGlobalAutomationPaused(b *testing.B) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	ctx := context.Background()
	for b.Loop() {
		IsGlobalAutomationPaused(ctx, db)
	}
}
