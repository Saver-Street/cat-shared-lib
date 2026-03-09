package entitlements

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
	calls    []mockCall
	callIdx  int
}

type mockCall struct {
	row *mockRow
}

func (m *mockQuerier) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	if m.callIdx >= len(m.calls) {
		return &mockRow{scanFunc: func(_ ...any) error { return errors.New("unexpected query") }}
	}
	row := m.calls[m.callIdx].row
	m.callIdx++
	return row
}

func TestGetUserTierAndUsage_ActiveProUser(t *testing.T) {
	active := "active"
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "pro"
			*dest[1].(**string) = &active
			return nil
		}}},
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*int) = 42
			return nil
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != "pro" {
		t.Errorf("tier = %q, want pro", tier)
	}
	if count != 42 {
		t.Errorf("count = %d, want 42", count)
	}
}

func TestGetUserTierAndUsage_PastDueDowngradesToFree(t *testing.T) {
	pastDue := "past_due"
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "power"
			*dest[1].(**string) = &pastDue
			return nil
		}}},
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*int) = 5
			return nil
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "user-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != "free" {
		t.Errorf("past_due user tier = %q, want free", tier)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestGetUserTierAndUsage_NilStatus(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "starter"
			*dest[1].(**string) = nil
			return nil
		}}},
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*int) = 10
			return nil
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "user-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != "starter" {
		t.Errorf("tier = %q, want starter", tier)
	}
	if count != 10 {
		t.Errorf("count = %d, want 10", count)
	}
}

func TestGetUserTierAndUsage_UserNotFound(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return pgx.ErrNoRows
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if tier != "free" {
		t.Errorf("tier = %q, want free (default)", tier)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestGetUserTierAndUsage_DBError(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return errors.New("connection refused")
		}}},
	}}

	tier, _, err := GetUserTierAndUsage(context.Background(), db, "user-x")
	if err == nil {
		t.Fatal("expected error")
	}
	if tier != "free" {
		t.Errorf("tier = %q, want free on error", tier)
	}
}

func TestGetUserTierAndUsage_AppCountQueryFails(t *testing.T) {
	active := "active"
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "pro"
			*dest[1].(**string) = &active
			return nil
		}}},
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return errors.New("count query failed")
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "user-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != "pro" {
		t.Errorf("tier = %q, want pro", tier)
	}
	// count defaults to 0 when the app count query fails (error is swallowed)
	if count != 0 {
		t.Errorf("count = %d, want 0 on count query failure", count)
	}
}

func TestGetUserTier_ReturnsOnlyTier(t *testing.T) {
	active := "active"
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "concierge"
			*dest[1].(**string) = &active
			return nil
		}}},
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*int) = 100
			return nil
		}}},
	}}

	tier, err := GetUserTier(context.Background(), db, "user-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != "concierge" {
		t.Errorf("GetUserTier = %q, want concierge", tier)
	}
}

func TestGetUserTier_PropagatesError(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return errors.New("timeout")
		}}},
	}}

	tier, err := GetUserTier(context.Background(), db, "user-6")
	if err == nil {
		t.Fatal("expected error")
	}
	if tier != "free" {
		t.Errorf("tier = %q, want free on error", tier)
	}
}

func TestGetUserTierAndUsage_NonPastDueStatus(t *testing.T) {
	canceled := "canceled"
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*string) = "pro"
			*dest[1].(**string) = &canceled
			return nil
		}}},
		{row: &mockRow{scanFunc: func(dest ...any) error {
			*dest[0].(*int) = 0
			return nil
		}}},
	}}

	tier, _, err := GetUserTierAndUsage(context.Background(), db, "user-7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "canceled" status should NOT downgrade to free (only "past_due" does)
	if tier != "pro" {
		t.Errorf("canceled status tier = %q, want pro", tier)
	}
}
