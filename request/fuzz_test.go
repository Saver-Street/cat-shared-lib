package request

import (
	"net/http"
	"net/url"
	"testing"
)

func FuzzParsePagination(f *testing.F) {
	f.Add("1", "20")
	f.Add("0", "0")
	f.Add("-1", "-5")
	f.Add("abc", "xyz")
	f.Add("999999", "999999")
	f.Add("", "")
	f.Add("1", "100")

	f.Fuzz(func(t *testing.T, page, limit string) {
		q := url.Values{}
		if page != "" {
			q.Set("page", page)
		}
		if limit != "" {
			q.Set("limit", limit)
		}
		p := ParsePagination(q, 20, 100)

		// Invariants that must always hold
		if p.Page < 1 {
			t.Errorf("Page = %d, must be >= 1", p.Page)
		}
		if p.Limit < 1 || p.Limit > 100 {
			t.Errorf("Limit = %d, must be 1-100", p.Limit)
		}
		if p.Offset != (p.Page-1)*p.Limit {
			t.Errorf("Offset = %d, want (page-1)*limit = %d", p.Offset, (p.Page-1)*p.Limit)
		}
	})
}

func FuzzRequireURLParam(f *testing.F) {
	f.Add("id", "123")
	f.Add("id", "")
	f.Add("name", "alice")
	f.Add("", "value")

	f.Fuzz(func(t *testing.T, key, value string) {
		paramFn := func(r *http.Request, k string) string {
			if k == key {
				return value
			}
			return ""
		}
		r, _ := http.NewRequest("GET", "/", nil)
		val, err := RequireURLParam(r, key, paramFn)

		if value == "" {
			if err == nil {
				t.Error("expected error for empty value")
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if val != value {
				t.Errorf("got %q, want %q", val, value)
			}
		}
	})
}
