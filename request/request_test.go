package request

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

// ── RequireURLParam Tests ───────────────────────────────────────────────────

func mockParamFn(params map[string]string) URLParamFunc {
	return func(r *http.Request, key string) string {
		return params[key]
	}
}

func TestRequireURLParam_Present(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	val, err := RequireURLParam(r, "id", mockParamFn(map[string]string{"id": "42"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "42" {
		t.Errorf("got %q, want %q", val, "42")
	}
}

func TestRequireURLParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParam(r, "id", mockParamFn(map[string]string{}))
	if err == nil {
		t.Fatal("expected error for missing param")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("expected 'missing' in error, got: %v", err)
	}
}

func TestRequireURLParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParam(r, "id", mockParamFn(map[string]string{"id": ""}))
	if err == nil {
		t.Fatal("expected error for empty param")
	}
}

func TestRequireURLParamInt_ValidInt(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	n, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "123"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 123 {
		t.Errorf("got %d, want 123", n)
	}
}

func TestRequireURLParamInt_NotAnInt(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "abc"}))
	if err == nil {
		t.Fatal("expected error for non-integer")
	}
	if !strings.Contains(err.Error(), "integer") {
		t.Errorf("expected 'integer' in error, got: %v", err)
	}
}

func TestRequireURLParamInt_Zero(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "0"}))
	if err == nil {
		t.Fatal("expected error for zero")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Errorf("expected 'positive' in error, got: %v", err)
	}
}

func TestRequireURLParamInt_Negative(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "-5"}))
	if err == nil {
		t.Fatal("expected error for negative")
	}
}

func TestRequireURLParamInt_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{}))
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}
