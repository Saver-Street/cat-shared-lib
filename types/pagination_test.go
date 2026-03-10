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
