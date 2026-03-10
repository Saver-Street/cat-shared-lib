package types

import "testing"

func TestNormalizePage_Defaults(t *testing.T) {
	p := NormalizePage(0, 0)
	if p.Page != 1 {
		t.Errorf("page = %d, want 1", p.Page)
	}
	if p.Limit != 20 {
		t.Errorf("limit = %d, want 20", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("offset = %d, want 0", p.Offset)
	}
}

func TestNormalizePage_CapLimit(t *testing.T) {
	p := NormalizePage(1, 500)
	if p.Limit != 100 {
		t.Errorf("limit capped at 100, got %d", p.Limit)
	}
}

func TestNormalizePage_OffsetCalc(t *testing.T) {
	p := NormalizePage(3, 25)
	if p.Offset != 50 {
		t.Errorf("offset = %d, want 50 (page 3 * limit 25)", p.Offset)
	}
}

func TestNormalizePage_Page1(t *testing.T) {
	p := NormalizePage(1, 10)
	if p.Offset != 0 {
		t.Errorf("page 1 offset should be 0, got %d", p.Offset)
	}
}

func TestNormalizePage_NegativePage(t *testing.T) {
	p := NormalizePage(-5, 10)
	if p.Page != 1 {
		t.Errorf("negative page should be clamped to 1, got %d", p.Page)
	}
	if p.Offset != 0 {
		t.Errorf("offset for page 1 should be 0, got %d", p.Offset)
	}
}

func TestNormalizePage_NegativeLimit(t *testing.T) {
	p := NormalizePage(1, -10)
	if p.Limit != 20 {
		t.Errorf("negative limit should default to 20, got %d", p.Limit)
	}
}

func TestNormalizePage_ExactBoundary(t *testing.T) {
	p := NormalizePage(1, 100)
	if p.Limit != 100 {
		t.Errorf("limit=100 should stay at 100, got %d", p.Limit)
	}
}

func TestNormalizePage_LimitOne(t *testing.T) {
	p := NormalizePage(1, 1)
	if p.Limit != 1 {
		t.Errorf("limit=1 should stay at 1, got %d", p.Limit)
	}
}

func TestNormalizePage_HighPage(t *testing.T) {
	p := NormalizePage(1000, 25)
	if p.Offset != 999*25 {
		t.Errorf("offset = %d, want %d", p.Offset, 999*25)
	}
}

func BenchmarkNormalizePage(b *testing.B) {
	for b.Loop() {
		NormalizePage(5, 25)
	}
}

func TestPaginationParams_HasNextPage(t *testing.T) {
	p := NormalizePage(1, 10)
	if !p.HasNextPage(25) {
		t.Error("page 1 of 25 items at limit 10 should have next page")
	}
	if p.HasNextPage(10) {
		t.Error("page 1 of 10 items at limit 10 should NOT have next page")
	}
	if p.HasNextPage(5) {
		t.Error("page 1 of 5 items at limit 10 should NOT have next page (fewer than limit)")
	}
}

func TestPaginationParams_HasNextPage_LastPage(t *testing.T) {
	p := NormalizePage(3, 10) // offset=20
	if p.HasNextPage(30) {
		t.Error("page 3 of 30 items at limit 10 should NOT have next page (exact)")
	}
	if !p.HasNextPage(31) {
		t.Error("page 3 of 31 items at limit 10 should have next page")
	}
}

func TestPaginationParams_IsLastPage(t *testing.T) {
	p := NormalizePage(2, 10) // offset=10
	if !p.IsLastPage(15) {
		t.Error("page 2 with 15 total should be last page")
	}
	if p.IsLastPage(25) {
		t.Error("page 2 with 25 total should NOT be last page")
	}
}

func TestPaginationParams_HasNextPage_ZeroTotal(t *testing.T) {
	p := NormalizePage(1, 10)
	if p.HasNextPage(0) {
		t.Error("no items should not have next page")
	}
}

func TestTotalPages_Normal(t *testing.T) {
	p := NormalizePage(1, 10)
	if got := p.TotalPages(25); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}
func TestTotalPages_Exact(t *testing.T) {
	p := NormalizePage(1, 10)
	if got := p.TotalPages(30); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}
func TestTotalPages_ZeroTotal(t *testing.T) {
	p := NormalizePage(1, 10)
	if got := p.TotalPages(0); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}
func TestTotalPages_OneItem(t *testing.T) {
	p := NormalizePage(1, 10)
	if got := p.TotalPages(1); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}
func TestTotalPages_Negative(t *testing.T) {
	p := NormalizePage(1, 10)
	if got := p.TotalPages(-5); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}
