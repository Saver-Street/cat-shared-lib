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
