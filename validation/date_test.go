package validation

import (
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestDate_Valid(t *testing.T) {
	testkit.AssertNil(t, Date("dob", "2024-01-15", time.DateOnly))
}

func TestDate_ValidRFC3339(t *testing.T) {
	testkit.AssertNil(t, Date("ts", "2024-01-15T10:30:00Z", time.RFC3339))
}

func TestDate_Invalid(t *testing.T) {
	err := Date("dob", "not-a-date", time.DateOnly)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "dob")
	testkit.AssertContains(t, err.Error(), "valid date")
}

func TestDate_WrongFormat(t *testing.T) {
	err := Date("dob", "01/15/2024", time.DateOnly)
	testkit.AssertNotNil(t, err)
}

func TestDateBefore_Valid(t *testing.T) {
	boundary := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	testkit.AssertNil(t, DateBefore("expiry", "2024-06-15", time.DateOnly, boundary))
}

func TestDateBefore_ExactBoundary(t *testing.T) {
	boundary := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	err := DateBefore("expiry", "2024-06-15", time.DateOnly, boundary)
	testkit.AssertNotNil(t, err) // Equal is not before.
}

func TestDateBefore_After(t *testing.T) {
	boundary := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	err := DateBefore("expiry", "2024-06-15", time.DateOnly, boundary)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "before")
}

func TestDateBefore_InvalidDate(t *testing.T) {
	boundary := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	err := DateBefore("expiry", "bad", time.DateOnly, boundary)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "valid date")
}

func TestDateAfter_Valid(t *testing.T) {
	boundary := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testkit.AssertNil(t, DateAfter("start", "2024-06-15", time.DateOnly, boundary))
}

func TestDateAfter_ExactBoundary(t *testing.T) {
	boundary := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	err := DateAfter("start", "2024-06-15", time.DateOnly, boundary)
	testkit.AssertNotNil(t, err) // Equal is not after.
}

func TestDateAfter_Before(t *testing.T) {
	boundary := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	err := DateAfter("start", "2024-06-15", time.DateOnly, boundary)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "after")
}

func TestDateAfter_InvalidDate(t *testing.T) {
	boundary := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	err := DateAfter("start", "nope", time.DateOnly, boundary)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "valid date")
}

func TestDateRange_InRange(t *testing.T) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	testkit.AssertNil(t, DateRange("event", "2024-06-15", time.DateOnly, earliest, latest))
}

func TestDateRange_AtEarliest(t *testing.T) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	testkit.AssertNil(t, DateRange("event", "2024-01-01", time.DateOnly, earliest, latest))
}

func TestDateRange_AtLatest(t *testing.T) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	testkit.AssertNil(t, DateRange("event", "2024-12-31", time.DateOnly, earliest, latest))
}

func TestDateRange_BeforeEarliest(t *testing.T) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	err := DateRange("event", "2023-06-15", time.DateOnly, earliest, latest)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "between")
}

func TestDateRange_AfterLatest(t *testing.T) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	err := DateRange("event", "2025-01-01", time.DateOnly, earliest, latest)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "between")
}

func TestDateRange_InvalidDate(t *testing.T) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	err := DateRange("event", "xyz", time.DateOnly, earliest, latest)
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "valid date")
}

func TestFutureDate_Valid(t *testing.T) {
	future := time.Now().AddDate(1, 0, 0).Format(time.DateOnly)
	testkit.AssertNil(t, FutureDate("expiry", future, time.DateOnly))
}

func TestFutureDate_PastDate(t *testing.T) {
	past := time.Now().AddDate(-1, 0, 0).Format(time.DateOnly)
	err := FutureDate("expiry", past, time.DateOnly)
	testkit.AssertNotNil(t, err)
}

func TestPastDate_Valid(t *testing.T) {
	past := time.Now().AddDate(-1, 0, 0).Format(time.DateOnly)
	testkit.AssertNil(t, PastDate("dob", past, time.DateOnly))
}

func TestPastDate_FutureDate(t *testing.T) {
	future := time.Now().AddDate(1, 0, 0).Format(time.DateOnly)
	err := PastDate("dob", future, time.DateOnly)
	testkit.AssertNotNil(t, err)
}

func BenchmarkDate(b *testing.B) {
	for b.Loop() {
		_ = Date("dob", "2024-01-15", time.DateOnly)
	}
}

func BenchmarkDateRange(b *testing.B) {
	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		_ = DateRange("event", "2024-06-15", time.DateOnly, earliest, latest)
	}
}
