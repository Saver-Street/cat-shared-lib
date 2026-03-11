package entitlements

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
	calls   []mockCall
	callIdx int
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, tier, "pro")
	testkit.AssertEqual(t, count, 42)
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, tier, "free")
	testkit.AssertEqual(t, count, 5)
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, tier, "starter")
	testkit.AssertEqual(t, count, 10)
}

func TestGetUserTierAndUsage_UserNotFound(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return pgx.ErrNoRows
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "nonexistent")
	testkit.AssertError(t, err)
	testkit.AssertEqual(t, tier, "free")
	testkit.AssertEqual(t, count, 0)
}

func TestGetUserTierAndUsage_DBError(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return errors.New("connection refused")
		}}},
	}}

	tier, _, err := GetUserTierAndUsage(context.Background(), db, "user-x")
	testkit.AssertError(t, err)
	testkit.AssertEqual(t, tier, "free")
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, tier, "pro")
	// count defaults to 0 when the app count query fails (error is swallowed)
	testkit.AssertEqual(t, count, 0)
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, tier, "concierge")
}

func TestGetUserTier_PropagatesError(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return errors.New("timeout")
		}}},
	}}

	tier, err := GetUserTier(context.Background(), db, "user-6")
	testkit.AssertError(t, err)
	testkit.AssertEqual(t, tier, "free")
}

func TestGetUserTierAndUsage_EmptyUserID(t *testing.T) {
	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return pgx.ErrNoRows
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(context.Background(), db, "")
	testkit.AssertError(t, err)
	testkit.AssertEqual(t, tier, "free")
	testkit.AssertEqual(t, count, 0)
}

func BenchmarkGetUserTierAndUsage(b *testing.B) {
	active := "active"
	db := &mockQuerier{calls: nil}
	ctx := context.Background()
	for b.Loop() {
		db.calls = []mockCall{
			{row: &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*string) = "pro"
				*dest[1].(**string) = &active
				return nil
			}}},
			{row: &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*int) = 42
				return nil
			}}},
		}
		db.callIdx = 0
		GetUserTierAndUsage(ctx, db, "user-bench")
	}
}

func BenchmarkGetUserTier(b *testing.B) {
	active := "active"
	db := &mockQuerier{calls: nil}
	ctx := context.Background()
	for b.Loop() {
		db.calls = []mockCall{
			{row: &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*string) = "pro"
				*dest[1].(**string) = &active
				return nil
			}}},
			{row: &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*int) = 10
				return nil
			}}},
		}
		db.callIdx = 0
		GetUserTier(ctx, db, "user-bench")
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, tier, "pro")
}
