package request

import (
	"net/url"
	"testing"
)

func TestParsePagination_Defaults(t *testing.T) {
	p := ParsePagination(url.Values{}, 20, 100)
	if p.Page != 1 || p.Limit != 20 || p.Offset != 0 {
		t.Errorf("got %+v, want {Page:1 Limit:20 Offset:0}", p)
	}
}

func TestParsePagination_CustomValues(t *testing.T) {
	q := url.Values{"page": {"3"}, "limit": {"15"}}
	p := ParsePagination(q, 20, 100)
	if p.Page != 3 || p.Limit != 15 || p.Offset != 30 {
		t.Errorf("got %+v, want {Page:3 Limit:15 Offset:30}", p)
	}
}

func TestParsePagination_MaxLimit(t *testing.T) {
	q := url.Values{"limit": {"200"}}
	p := ParsePagination(q, 20, 50)
	if p.Limit != 50 {
		t.Errorf("limit = %d, want 50 (capped at max)", p.Limit)
	}
}

func TestParsePagination_InvalidPage(t *testing.T) {
	q := url.Values{"page": {"abc"}}
	p := ParsePagination(q, 20, 100)
	if p.Page != 1 {
		t.Errorf("page = %d, want 1 (fallback for invalid)", p.Page)
	}
}

func TestParsePagination_ZeroPage(t *testing.T) {
	q := url.Values{"page": {"0"}}
	p := ParsePagination(q, 20, 100)
	if p.Page != 1 {
		t.Errorf("page = %d, want 1 (fallback for zero)", p.Page)
	}
}

func TestParsePagination_NegativePage(t *testing.T) {
	q := url.Values{"page": {"-1"}}
	p := ParsePagination(q, 20, 100)
	if p.Page != 1 {
		t.Errorf("page = %d, want 1 (fallback for negative)", p.Page)
	}
}

func TestParsePagination_InvalidLimit(t *testing.T) {
	q := url.Values{"limit": {"xyz"}}
	p := ParsePagination(q, 25, 100)
	if p.Limit != 25 {
		t.Errorf("limit = %d, want 25 (fallback for invalid)", p.Limit)
	}
}

func TestParsePagination_ZeroLimit(t *testing.T) {
	q := url.Values{"limit": {"0"}}
	p := ParsePagination(q, 25, 100)
	if p.Limit != 25 {
		t.Errorf("limit = %d, want 25 (fallback for zero)", p.Limit)
	}
}

func TestParsePagination_NegativeLimit(t *testing.T) {
	q := url.Values{"limit": {"-5"}}
	p := ParsePagination(q, 25, 100)
	if p.Limit != 25 {
		t.Errorf("limit = %d, want 25 (fallback for negative)", p.Limit)
	}
}

func TestParsePagination_Page2Offset(t *testing.T) {
	q := url.Values{"page": {"2"}, "limit": {"10"}}
	p := ParsePagination(q, 20, 100)
	if p.Offset != 10 {
		t.Errorf("offset = %d, want 10 for page 2 limit 10", p.Offset)
	}
}
