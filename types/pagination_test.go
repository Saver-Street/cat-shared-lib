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
