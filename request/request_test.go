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

func TestRequireURLParamInt_MaxInt(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	n, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "9223372036854775807"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 9223372036854775807 {
		t.Errorf("got %d, want max int64", n)
	}
}

func TestRequireURLParamInt_Overflow(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "99999999999999999999"}))
	if err == nil {
		t.Fatal("expected error for overflow value")
	}
}

func TestParsePagination_LargePageOffset(t *testing.T) {
	q := url.Values{"page": {"10000"}, "limit": {"100"}}
	p := ParsePagination(q, 20, 100)
	expectedOffset := 9999 * 100
	if p.Offset != expectedOffset {
		t.Errorf("offset = %d, want %d", p.Offset, expectedOffset)
	}
}

// --- Benchmarks ---

func BenchmarkParsePagination(b *testing.B) {
	q := url.Values{"page": {"3"}, "limit": {"25"}}
	for b.Loop() {
		ParsePagination(q, 20, 100)
	}
}

func BenchmarkRequireURLParam(b *testing.B) {
	r := httptest.NewRequest("GET", "/test", nil)
	fn := mockParamFn(map[string]string{"id": "42"})
	for b.Loop() {
		RequireURLParam(r, "id", fn)
	}
}

func BenchmarkRequireURLParamInt(b *testing.B) {
	r := httptest.NewRequest("GET", "/test", nil)
	fn := mockParamFn(map[string]string{"id": "123"})
	for b.Loop() {
		RequireURLParamInt(r, "id", fn)
	}
}

func TestParsePagination_ZeroDefaults(t *testing.T) {
	// Zero defaultLimit and maxLimit should apply sane built-in defaults.
	q := url.Values{}
	p := ParsePagination(q, 0, 0)
	if p.Limit != 20 {
		t.Errorf("expected Limit=20, got %d", p.Limit)
	}
}

func TestParsePagination_NegativeDefaults(t *testing.T) {
	q := url.Values{}
	p := ParsePagination(q, -5, -1)
	if p.Limit != 20 {
		t.Errorf("expected Limit=20, got %d", p.Limit)
	}
}

func TestParsePagination_ZeroMaxLimitCapsCorrectly(t *testing.T) {
	// When maxLimit was 0 before the fix, any limit > 0 was capped to 0.
	// Now maxLimit defaults to 100, so explicit limit=50 should be honoured.
	q := url.Values{"limit": {"50"}}
	p := ParsePagination(q, 0, 0)
	if p.Limit != 50 {
		t.Errorf("expected Limit=50, got %d", p.Limit)
	}
}

func TestRequireQueryParam_Present(t *testing.T) {
	q := url.Values{"filter": {"active"}}
	got, err := RequireQueryParam(q, "filter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "active" {
		t.Errorf("got %q, want %q", got, "active")
	}
}

func TestRequireQueryParam_Missing(t *testing.T) {
	q := url.Values{}
	_, err := RequireQueryParam(q, "filter")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}

func TestRequireQueryParam_Empty(t *testing.T) {
	q := url.Values{"filter": {""}}
	_, err := RequireQueryParam(q, "filter")
	if err == nil {
		t.Fatal("expected error for empty param")
	}
}

func TestParseBoolParam_TrueValues(t *testing.T) {
	for _, v := range []string{"true", "1", "yes", "True", "YES"} {
		q := url.Values{"active": {v}}
		if !ParseBoolParam(q, "active", false) {
			t.Errorf("expected true for %q", v)
		}
	}
}

func TestParseBoolParam_FalseValues(t *testing.T) {
	for _, v := range []string{"false", "0", "no", "False", "NO"} {
		q := url.Values{"active": {v}}
		if ParseBoolParam(q, "active", true) {
			t.Errorf("expected false for %q", v)
		}
	}
}

func TestParseBoolParam_DefaultOnMissing(t *testing.T) {
	q := url.Values{}
	if !ParseBoolParam(q, "active", true) {
		t.Error("expected default true when param missing")
	}
	if ParseBoolParam(q, "active", false) {
		t.Error("expected default false when param missing")
	}
}

func TestParseBoolParam_DefaultOnUnknown(t *testing.T) {
	q := url.Values{"active": {"maybe"}}
	if !ParseBoolParam(q, "active", true) {
		t.Error("expected default true for unrecognised value")
	}
}

func BenchmarkRequireQueryParam(b *testing.B) {
	q := url.Values{"search": {"hello"}}
	for b.Loop() {
		_, _ = RequireQueryParam(q, "search")
	}
}

func BenchmarkParseBoolParam(b *testing.B) {
	q := url.Values{"active": {"true"}}
	for b.Loop() {
		ParseBoolParam(q, "active", false)
	}
}

func TestOptionalQueryParam_Present(t *testing.T) {
	q := url.Values{"name": {"alice"}}
	if got := OptionalQueryParam(q, "name", "default"); got != "alice" {
		t.Errorf("got %q, want alice", got)
	}
}

func TestOptionalQueryParam_Missing(t *testing.T) {
	q := url.Values{}
	if got := OptionalQueryParam(q, "name", "default"); got != "default" {
		t.Errorf("got %q, want default", got)
	}
}

func TestOptionalQueryParam_Empty(t *testing.T) {
	q := url.Values{"name": {""}}
	if got := OptionalQueryParam(q, "name", "fallback"); got != "fallback" {
		t.Errorf("got %q, want fallback", got)
	}
}

func TestOptionalQueryInt_Present(t *testing.T) {
	q := url.Values{"count": {"42"}}
	if got := OptionalQueryInt(q, "count", 0); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestOptionalQueryInt_Missing(t *testing.T) {
	q := url.Values{}
	if got := OptionalQueryInt(q, "count", 10); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

func TestOptionalQueryInt_Invalid(t *testing.T) {
	q := url.Values{"count": {"notanint"}}
	if got := OptionalQueryInt(q, "count", 99); got != 99 {
		t.Errorf("got %d, want 99 (default on invalid)", got)
	}
}

func TestOptionalQueryInt_Negative(t *testing.T) {
	q := url.Values{"page": {"-5"}}
	if got := OptionalQueryInt(q, "page", 1); got != -5 {
		t.Errorf("got %d, want -5 (negatives are valid)", got)
	}
}

func TestRequireQueryParamInt_Valid(t *testing.T) {
	q := url.Values{"id": {"123"}}
	n, err := RequireQueryParamInt(q, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 123 {
		t.Errorf("got %d, want 123", n)
	}
}

func TestRequireQueryParamInt_Missing(t *testing.T) {
	q := url.Values{}
	_, err := RequireQueryParamInt(q, "id")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}

func TestRequireQueryParamInt_NotAnInt(t *testing.T) {
	q := url.Values{"id": {"abc"}}
	_, err := RequireQueryParamInt(q, "id")
	if err == nil {
		t.Fatal("expected error for non-integer value")
	}
}

func TestParseCommaSeparated_Values(t *testing.T) {
	q := url.Values{"tags": {"go, python, rust"}}
	got := ParseCommaSeparated(q, "tags")
	if len(got) != 3 || got[0] != "go" || got[1] != "python" || got[2] != "rust" {
		t.Errorf("got %v, want [go python rust]", got)
	}
}
func TestParseCommaSeparated_Absent(t *testing.T) {
	if got := ParseCommaSeparated(url.Values{}, "tags"); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
func TestParseCommaSeparated_EmptyString(t *testing.T) {
	if got := ParseCommaSeparated(url.Values{"tags": {""}}, "tags"); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
func TestParseCommaSeparated_WhitespaceOnly(t *testing.T) {
	if got := ParseCommaSeparated(url.Values{"tags": {"  , ,  "}}, "tags"); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
func TestParseCommaSeparated_Single(t *testing.T) {
	got := ParseCommaSeparated(url.Values{"tags": {"go"}}, "tags")
	if len(got) != 1 || got[0] != "go" {
		t.Errorf("got %v, want [go]", got)
	}
}
func TestParseCommaSeparatedInts_Valid(t *testing.T) {
	got, err := ParseCommaSeparatedInts(url.Values{"ids": {"1,2,3"}}, "ids")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", got)
	}
}
func TestParseCommaSeparatedInts_Absent(t *testing.T) {
	got, err := ParseCommaSeparatedInts(url.Values{}, "ids")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got != nil { t.Errorf("got %v, want nil", got) }
}
func TestParseCommaSeparatedInts_Invalid(t *testing.T) {
	_, err := ParseCommaSeparatedInts(url.Values{"ids": {"1,abc,3"}}, "ids")
	if err == nil { t.Fatal("expected error") }
}
func TestParseCommaSeparatedInts_Negative(t *testing.T) {
	got, err := ParseCommaSeparatedInts(url.Values{"ids": {"-1,0,2"}}, "ids")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got[0] != -1 || got[1] != 0 || got[2] != 2 {
		t.Errorf("got %v, want [-1 0 2]", got)
	}
}
