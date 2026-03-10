package types

import "testing"

func FuzzNormalizePage(f *testing.F) {
	f.Add(0, 0)
	f.Add(1, 20)
	f.Add(-5, -10)
	f.Add(1000, 500)
	f.Add(1, 1)
	f.Add(1, 100)
	f.Add(1, 101)

	f.Fuzz(func(t *testing.T, page, limit int) {
		p := NormalizePage(page, limit)

		if p.Page < 1 {
			t.Errorf("Page = %d, must be >= 1", p.Page)
		}
		if p.Limit < 1 || p.Limit > 100 {
			t.Errorf("Limit = %d, must be 1-100", p.Limit)
		}
		if p.Offset != (p.Page-1)*p.Limit {
			t.Errorf("Offset = %d, want %d", p.Offset, (p.Page-1)*p.Limit)
		}
	})
}
