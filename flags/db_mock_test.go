package flags

import (
	"context"
	"errors"
	"testing"

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
	if !IsFeatureEnabled(context.Background(), nil, FlagAIScoring) {
		t.Error("nil querier should return true (safe default)")
	}
}

func TestIsFeatureEnabled_FlagTrue(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	if !IsFeatureEnabled(context.Background(), db, FlagAIScoring) {
		t.Error("flag=true should return true")
	}
}

func TestIsFeatureEnabled_FlagFalse(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	if IsFeatureEnabled(context.Background(), db, FlagAIScoring) {
		t.Error("flag=false should return false")
	}
}

func TestIsFeatureEnabled_FlagNotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newNoRowsRow()
	}}
	if !IsFeatureEnabled(context.Background(), db, "nonexistent") {
		t.Error("missing flag should return true (safe default)")
	}
}

func TestIsFeatureEnabled_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newErrorRow("connection lost")
	}}
	if !IsFeatureEnabled(context.Background(), db, FlagResumeParsing) {
		t.Error("DB error should return true (safe default)")
	}
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
	if capturedSQL != "SELECT value FROM site_settings WHERE key = $1" {
		t.Errorf("unexpected SQL: %s", capturedSQL)
	}
	if len(capturedArgs) != 1 || capturedArgs[0] != "flag_myFlag" {
		t.Errorf("expected flag_myFlag arg, got %v", capturedArgs)
	}
}

func TestIsFeatureEnabled_EmptyValue(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("")
	}}
	if IsFeatureEnabled(context.Background(), db, FlagSIAFI) {
		t.Error("empty value should return false (not 'true')")
	}
}

// --- IsMaintenanceModeActive tests ---

func TestIsMaintenanceModeActive_NilQuerier(t *testing.T) {
	if IsMaintenanceModeActive(context.Background(), nil) {
		t.Error("nil querier should return false")
	}
}

func TestIsMaintenanceModeActive_True(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	if !IsMaintenanceModeActive(context.Background(), db) {
		t.Error("maintenance=true should return true")
	}
}

func TestIsMaintenanceModeActive_False(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	if IsMaintenanceModeActive(context.Background(), db) {
		t.Error("maintenance=false should return false")
	}
}

func TestIsMaintenanceModeActive_NotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newNoRowsRow()
	}}
	if IsMaintenanceModeActive(context.Background(), db) {
		t.Error("missing row should return false")
	}
}

func TestIsMaintenanceModeActive_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newErrorRow("timeout")
	}}
	if IsMaintenanceModeActive(context.Background(), db) {
		t.Error("DB error should return false")
	}
}

// --- IsGlobalAutomationPaused tests ---

func TestIsGlobalAutomationPaused_NilQuerier(t *testing.T) {
	if IsGlobalAutomationPaused(context.Background(), nil) {
		t.Error("nil querier should return false")
	}
}

func TestIsGlobalAutomationPaused_True(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("true")
	}}
	if !IsGlobalAutomationPaused(context.Background(), db) {
		t.Error("paused=true should return true")
	}
}

func TestIsGlobalAutomationPaused_False(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newValueRow("false")
	}}
	if IsGlobalAutomationPaused(context.Background(), db) {
		t.Error("paused=false should return false")
	}
}

func TestIsGlobalAutomationPaused_NotFound(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newNoRowsRow()
	}}
	if IsGlobalAutomationPaused(context.Background(), db) {
		t.Error("missing row should return false")
	}
}

func TestIsGlobalAutomationPaused_DBError(t *testing.T) {
	db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
		return newErrorRow("connection reset")
	}}
	if IsGlobalAutomationPaused(context.Background(), db) {
		t.Error("DB error should return false")
	}
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
