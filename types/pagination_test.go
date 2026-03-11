package types

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNormalizePage_Defaults(t *testing.T) {
	p := NormalizePage(0, 0)
	testkit.AssertEqual(t, p.Page, 1)
	testkit.AssertEqual(t, p.Limit, 20)
	testkit.AssertEqual(t, p.Offset, 0)
}

func TestNormalizePage_CapLimit(t *testing.T) {
	p := NormalizePage(1, 500)
	testkit.AssertEqual(t, p.Limit, 100)
}

func TestNormalizePage_OffsetCalc(t *testing.T) {
	p := NormalizePage(3, 25)
	testkit.AssertEqual(t, p.Offset, 50)
}

func TestNormalizePage_Page1(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertEqual(t, p.Offset, 0)
}

func TestNormalizePage_NegativePage(t *testing.T) {
	p := NormalizePage(-5, 10)
	testkit.AssertEqual(t, p.Page, 1)
	testkit.AssertEqual(t, p.Offset, 0)
}

func TestNormalizePage_NegativeLimit(t *testing.T) {
	p := NormalizePage(1, -10)
	testkit.AssertEqual(t, p.Limit, 20)
}

func TestNormalizePage_ExactBoundary(t *testing.T) {
	p := NormalizePage(1, 100)
	testkit.AssertEqual(t, p.Limit, 100)
}

func TestNormalizePage_LimitOne(t *testing.T) {
	p := NormalizePage(1, 1)
	testkit.AssertEqual(t, p.Limit, 1)
}

func TestNormalizePage_HighPage(t *testing.T) {
	p := NormalizePage(1000, 25)
	testkit.AssertEqual(t, p.Offset, 999*25)
}

func BenchmarkNormalizePage(b *testing.B) {
	for b.Loop() {
		NormalizePage(5, 25)
	}
}

func TestPaginationParams_HasNextPage(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertTrue(t, p.HasNextPage(25))
	testkit.AssertFalse(t, p.HasNextPage(10))
	testkit.AssertFalse(t, p.HasNextPage(5))
}

func TestPaginationParams_HasNextPage_LastPage(t *testing.T) {
	p := NormalizePage(3, 10) // offset=20
	testkit.AssertFalse(t, p.HasNextPage(30))
	testkit.AssertTrue(t, p.HasNextPage(31))
}

func TestPaginationParams_IsLastPage(t *testing.T) {
	p := NormalizePage(2, 10) // offset=10
	testkit.AssertTrue(t, p.IsLastPage(15))
	testkit.AssertFalse(t, p.IsLastPage(25))
}

func TestPaginationParams_HasNextPage_ZeroTotal(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertFalse(t, p.HasNextPage(0))
}

func TestTotalPages_Normal(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertEqual(t, p.TotalPages(25), 3)
}
func TestTotalPages_Exact(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertEqual(t, p.TotalPages(30), 3)
}
func TestTotalPages_ZeroTotal(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertEqual(t, p.TotalPages(0), 0)
}
func TestTotalPages_OneItem(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertEqual(t, p.TotalPages(1), 1)
}
func TestTotalPages_Negative(t *testing.T) {
	p := NormalizePage(1, 10)
	testkit.AssertEqual(t, p.TotalPages(-5), 0)
}

// --- NormalizeCursor ---

func TestNormalizeCursor_Defaults(t *testing.T) {
	c := NormalizeCursor("", 0)
	testkit.AssertEqual(t, c.Cursor, "")
	testkit.AssertEqual(t, c.Limit, 20)
}

func TestNormalizeCursor_WithCursor(t *testing.T) {
	c := NormalizeCursor("abc123", 10)
	testkit.AssertEqual(t, c.Cursor, "abc123")
	testkit.AssertEqual(t, c.Limit, 10)
}

func TestNormalizeCursor_NegativeLimit(t *testing.T) {
	c := NormalizeCursor("", -5)
	testkit.AssertEqual(t, c.Limit, 20)
}

func TestNormalizeCursor_OverMaxLimit(t *testing.T) {
	c := NormalizeCursor("", 200)
	testkit.AssertEqual(t, c.Limit, 100)
}

func TestNormalizeCursor_ExactBoundary(t *testing.T) {
	c := NormalizeCursor("x", 100)
	testkit.AssertEqual(t, c.Limit, 100)
}

func TestNormalizeCursor_LimitOne(t *testing.T) {
	c := NormalizeCursor("x", 1)
	testkit.AssertEqual(t, c.Limit, 1)
}

func BenchmarkNormalizeCursor(b *testing.B) {
	for b.Loop() {
		NormalizeCursor("cursor_value", 25)
	}
}

// --- NewCursorPage ---

func TestNewCursorPage_FewerThanLimit(t *testing.T) {
	items := []string{"a", "b"}
	page := NewCursorPage(items, 5, func(s string) string { return s })
	testkit.AssertEqual(t, len(page.Items), 2)
	testkit.AssertEqual(t, page.NextCursor, "")
	testkit.AssertFalse(t, page.HasMore)
}

func TestNewCursorPage_ExactlyLimit(t *testing.T) {
	items := []string{"a", "b", "c"}
	page := NewCursorPage(items, 3, func(s string) string { return s })
	testkit.AssertEqual(t, len(page.Items), 3)
	testkit.AssertEqual(t, page.NextCursor, "")
	testkit.AssertFalse(t, page.HasMore)
}

func TestNewCursorPage_MoreThanLimit(t *testing.T) {
	items := []string{"a", "b", "c", "d"}
	page := NewCursorPage(items, 3, func(s string) string { return s })
	testkit.AssertEqual(t, len(page.Items), 3)
	testkit.AssertEqual(t, page.NextCursor, "c")
	testkit.AssertTrue(t, page.HasMore)
}

func TestNewCursorPage_Empty(t *testing.T) {
	var items []int
	page := NewCursorPage(items, 10, func(i int) string { return "" })
	testkit.AssertEqual(t, len(page.Items), 0)
	testkit.AssertFalse(t, page.HasMore)
}

func TestNewCursorPage_CustomCursorFn(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}
	items := []item{{1, "a"}, {2, "b"}, {3, "c"}, {4, "d"}}
	page := NewCursorPage(items, 2, func(i item) string { return i.Name })
	testkit.AssertEqual(t, len(page.Items), 2)
	testkit.AssertEqual(t, page.NextCursor, "b")
	testkit.AssertTrue(t, page.HasMore)
}

// --- ApplyOffset ---

func TestApplyOffset_Normal(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	got := ApplyOffset(items, 1, 2)
	testkit.AssertEqual(t, len(got), 2)
	testkit.AssertEqual(t, got[0], 2)
	testkit.AssertEqual(t, got[1], 3)
}

func TestApplyOffset_OffsetBeyondLen(t *testing.T) {
	items := []int{1, 2, 3}
	got := ApplyOffset(items, 5, 2)
	testkit.AssertTrue(t, got == nil)
}

func TestApplyOffset_OffsetEqualsLen(t *testing.T) {
	items := []int{1, 2, 3}
	got := ApplyOffset(items, 3, 2)
	testkit.AssertTrue(t, got == nil)
}

func TestApplyOffset_LimitExceedsRemaining(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	got := ApplyOffset(items, 3, 10)
	testkit.AssertEqual(t, len(got), 2)
	testkit.AssertEqual(t, got[0], 4)
	testkit.AssertEqual(t, got[1], 5)
}

func TestApplyOffset_ZeroOffset(t *testing.T) {
	items := []string{"a", "b", "c"}
	got := ApplyOffset(items, 0, 2)
	testkit.AssertEqual(t, len(got), 2)
	testkit.AssertEqual(t, got[0], "a")
}

func TestApplyOffset_EmptySlice(t *testing.T) {
	var items []int
	got := ApplyOffset(items, 0, 10)
	testkit.AssertTrue(t, got == nil)
}

func BenchmarkApplyOffset(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		ApplyOffset(items, 100, 50)
	}
}
