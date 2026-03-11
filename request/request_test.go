package request

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestParsePagination_Defaults(t *testing.T) {
	p := ParsePagination(url.Values{}, 20, 100)
	testkit.AssertEqual(t, p.Page, 1)
	testkit.AssertEqual(t, p.Limit, 20)
	testkit.AssertEqual(t, p.Offset, 0)
}

func TestParsePagination_CustomValues(t *testing.T) {
	q := url.Values{"page": {"3"}, "limit": {"15"}}
	p := ParsePagination(q, 20, 100)
	testkit.AssertEqual(t, p.Page, 3)
	testkit.AssertEqual(t, p.Limit, 15)
	testkit.AssertEqual(t, p.Offset, 30)
}

func TestParsePagination_MaxLimit(t *testing.T) {
	q := url.Values{"limit": {"200"}}
	p := ParsePagination(q, 20, 50)
	testkit.AssertEqual(t, p.Limit, 50)
}

func TestParsePagination_InvalidPage(t *testing.T) {
	q := url.Values{"page": {"abc"}}
	p := ParsePagination(q, 20, 100)
	testkit.AssertEqual(t, p.Page, 1)
}

func TestParsePagination_ZeroPage(t *testing.T) {
	q := url.Values{"page": {"0"}}
	p := ParsePagination(q, 20, 100)
	testkit.AssertEqual(t, p.Page, 1)
}

func TestParsePagination_NegativePage(t *testing.T) {
	q := url.Values{"page": {"-1"}}
	p := ParsePagination(q, 20, 100)
	testkit.AssertEqual(t, p.Page, 1)
}

func TestParsePagination_InvalidLimit(t *testing.T) {
	q := url.Values{"limit": {"xyz"}}
	p := ParsePagination(q, 25, 100)
	testkit.AssertEqual(t, p.Limit, 25)
}

func TestParsePagination_ZeroLimit(t *testing.T) {
	q := url.Values{"limit": {"0"}}
	p := ParsePagination(q, 25, 100)
	testkit.AssertEqual(t, p.Limit, 25)
}

func TestParsePagination_NegativeLimit(t *testing.T) {
	q := url.Values{"limit": {"-5"}}
	p := ParsePagination(q, 25, 100)
	testkit.AssertEqual(t, p.Limit, 25)
}

func TestParsePagination_Page2Offset(t *testing.T) {
	q := url.Values{"page": {"2"}, "limit": {"10"}}
	p := ParsePagination(q, 20, 100)
	testkit.AssertEqual(t, p.Offset, 10)
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
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, val, "42")
}

func TestRequireURLParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParam(r, "id", mockParamFn(map[string]string{}))
	if err == nil {
		t.Fatal("expected error for missing param")
	}
	testkit.AssertErrorContains(t, err, "missing")
}

func TestRequireURLParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParam(r, "id", mockParamFn(map[string]string{"id": ""}))
	testkit.AssertError(t, err)
}

func TestRequireURLParamInt_ValidInt(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	n, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "123"}))
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, n, int64(123))
}

func TestRequireURLParamInt_NotAnInt(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "abc"}))
	if err == nil {
		t.Fatal("expected error for non-integer")
	}
	testkit.AssertErrorContains(t, err, "integer")
}

func TestRequireURLParamInt_Zero(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "0"}))
	if err == nil {
		t.Fatal("expected error for zero")
	}
	testkit.AssertErrorContains(t, err, "positive")
}

func TestRequireURLParamInt_Negative(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "-5"}))
	testkit.AssertError(t, err)
}

func TestRequireURLParamInt_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{}))
	testkit.AssertError(t, err)
}

func TestRequireURLParamInt_MaxInt(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	n, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "9223372036854775807"}))
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, n, int64(9223372036854775807))
}

func TestRequireURLParamInt_Overflow(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	_, err := RequireURLParamInt(r, "id", mockParamFn(map[string]string{"id": "99999999999999999999"}))
	testkit.AssertError(t, err)
}

func TestParsePagination_LargePageOffset(t *testing.T) {
	q := url.Values{"page": {"10000"}, "limit": {"100"}}
	p := ParsePagination(q, 20, 100)
	expectedOffset := 9999 * 100
	testkit.AssertEqual(t, p.Offset, expectedOffset)
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
	testkit.AssertEqual(t, p.Limit, 20)
}

func TestParsePagination_NegativeDefaults(t *testing.T) {
	q := url.Values{}
	p := ParsePagination(q, -5, -1)
	testkit.AssertEqual(t, p.Limit, 20)
}

func TestParsePagination_ZeroMaxLimitCapsCorrectly(t *testing.T) {
	// When maxLimit was 0 before the fix, any limit > 0 was capped to 0.
	// Now maxLimit defaults to 100, so explicit limit=50 should be honoured.
	q := url.Values{"limit": {"50"}}
	p := ParsePagination(q, 0, 0)
	testkit.AssertEqual(t, p.Limit, 50)
}

func TestRequireQueryParam_Present(t *testing.T) {
	q := url.Values{"filter": {"active"}}
	got, err := RequireQueryParam(q, "filter")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, got, "active")
}

func TestRequireQueryParam_Missing(t *testing.T) {
	q := url.Values{}
	_, err := RequireQueryParam(q, "filter")
	testkit.AssertError(t, err)
}

func TestRequireQueryParam_Empty(t *testing.T) {
	q := url.Values{"filter": {""}}
	_, err := RequireQueryParam(q, "filter")
	testkit.AssertError(t, err)
}

func TestParseBoolParam_TrueValues(t *testing.T) {
	for _, v := range []string{"true", "1", "yes", "True", "YES"} {
		q := url.Values{"active": {v}}
		testkit.AssertTrue(t, ParseBoolParam(q, "active", false))
	}
}

func TestParseBoolParam_FalseValues(t *testing.T) {
	for _, v := range []string{"false", "0", "no", "False", "NO"} {
		q := url.Values{"active": {v}}
		testkit.AssertFalse(t, ParseBoolParam(q, "active", true))
	}
}

func TestParseBoolParam_DefaultOnMissing(t *testing.T) {
	q := url.Values{}
	testkit.AssertTrue(t, ParseBoolParam(q, "active", true))
	testkit.AssertFalse(t, ParseBoolParam(q, "active", false))
}

func TestParseBoolParam_DefaultOnUnknown(t *testing.T) {
	q := url.Values{"active": {"maybe"}}
	testkit.AssertTrue(t, ParseBoolParam(q, "active", true))
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
	testkit.AssertEqual(t, OptionalQueryParam(q, "name", "default"), "alice")
}

func TestOptionalQueryParam_Missing(t *testing.T) {
	q := url.Values{}
	testkit.AssertEqual(t, OptionalQueryParam(q, "name", "default"), "default")
}

func TestOptionalQueryParam_Empty(t *testing.T) {
	q := url.Values{"name": {""}}
	testkit.AssertEqual(t, OptionalQueryParam(q, "name", "fallback"), "fallback")
}

func TestOptionalQueryInt_Present(t *testing.T) {
	q := url.Values{"count": {"42"}}
	testkit.AssertEqual(t, OptionalQueryInt(q, "count", 0), int64(42))
}

func TestOptionalQueryInt_Missing(t *testing.T) {
	q := url.Values{}
	testkit.AssertEqual(t, OptionalQueryInt(q, "count", 10), int64(10))
}

func TestOptionalQueryInt_Invalid(t *testing.T) {
	q := url.Values{"count": {"notanint"}}
	testkit.AssertEqual(t, OptionalQueryInt(q, "count", 99), int64(99))
}

func TestOptionalQueryInt_Negative(t *testing.T) {
	q := url.Values{"page": {"-5"}}
	testkit.AssertEqual(t, OptionalQueryInt(q, "page", 1), int64(-5))
}

func TestRequireQueryParamInt_Valid(t *testing.T) {
	q := url.Values{"id": {"123"}}
	n, err := RequireQueryParamInt(q, "id")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, n, int64(123))
}

func TestRequireQueryParamInt_Missing(t *testing.T) {
	q := url.Values{}
	_, err := RequireQueryParamInt(q, "id")
	testkit.AssertError(t, err)
}

func TestRequireQueryParamInt_NotAnInt(t *testing.T) {
	q := url.Values{"id": {"abc"}}
	_, err := RequireQueryParamInt(q, "id")
	testkit.AssertError(t, err)
}

func TestParseCommaSeparated_Values(t *testing.T) {
	q := url.Values{"tags": {"go, python, rust"}}
	got := ParseCommaSeparated(q, "tags")
	testkit.AssertLen(t, got, 3)
	testkit.AssertEqual(t, got[0], "go")
	testkit.AssertEqual(t, got[1], "python")
	testkit.AssertEqual(t, got[2], "rust")
}
func TestParseCommaSeparated_Absent(t *testing.T) {
	testkit.AssertNil(t, ParseCommaSeparated(url.Values{}, "tags"))
}
func TestParseCommaSeparated_EmptyString(t *testing.T) {
	testkit.AssertNil(t, ParseCommaSeparated(url.Values{"tags": {""}}, "tags"))
}
func TestParseCommaSeparated_WhitespaceOnly(t *testing.T) {
	testkit.AssertNil(t, ParseCommaSeparated(url.Values{"tags": {"  , ,  "}}, "tags"))
}
func TestParseCommaSeparated_Single(t *testing.T) {
	got := ParseCommaSeparated(url.Values{"tags": {"go"}}, "tags")
	testkit.AssertLen(t, got, 1)
	testkit.AssertEqual(t, got[0], "go")
}
func TestParseCommaSeparatedInts_Valid(t *testing.T) {
	got, err := ParseCommaSeparatedInts(url.Values{"ids": {"1,2,3"}}, "ids")
	testkit.RequireNoError(t, err)
	testkit.AssertLen(t, got, 3)
	testkit.AssertEqual(t, got[0], int64(1))
	testkit.AssertEqual(t, got[1], int64(2))
	testkit.AssertEqual(t, got[2], int64(3))
}
func TestParseCommaSeparatedInts_Absent(t *testing.T) {
	got, err := ParseCommaSeparatedInts(url.Values{}, "ids")
	testkit.RequireNoError(t, err)
	testkit.AssertNil(t, got)
}
func TestParseCommaSeparatedInts_Invalid(t *testing.T) {
	_, err := ParseCommaSeparatedInts(url.Values{"ids": {"1,abc,3"}}, "ids")
	testkit.AssertError(t, err)
}
func TestParseCommaSeparatedInts_Negative(t *testing.T) {
	got, err := ParseCommaSeparatedInts(url.Values{"ids": {"-1,0,2"}}, "ids")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, got[0], int64(-1))
	testkit.AssertEqual(t, got[1], int64(0))
	testkit.AssertEqual(t, got[2], int64(2))
}

func TestParseSortOrder_Default(t *testing.T) {
	field, dir := ParseSortOrder(url.Values{}, []string{"name", "created_at"}, "created_at", "desc")
	testkit.AssertEqual(t, field, "created_at")
	testkit.AssertEqual(t, dir, "desc")
}

func TestParseSortOrder_Valid(t *testing.T) {
	q := url.Values{"sort": {"name"}, "order": {"asc"}}
	field, dir := ParseSortOrder(q, []string{"name", "created_at"}, "created_at", "desc")
	testkit.AssertEqual(t, field, "name")
	testkit.AssertEqual(t, dir, "asc")
}

func TestParseSortOrder_InvalidSort(t *testing.T) {
	q := url.Values{"sort": {"unknown"}, "order": {"asc"}}
	field, dir := ParseSortOrder(q, []string{"name", "created_at"}, "created_at", "desc")
	testkit.AssertEqual(t, field, "created_at")
	testkit.AssertEqual(t, dir, "asc")
}

func TestParseSortOrder_InvalidOrder(t *testing.T) {
	q := url.Values{"sort": {"name"}, "order": {"random"}}
	field, dir := ParseSortOrder(q, []string{"name"}, "created_at", "desc")
	testkit.AssertEqual(t, field, "name")
	testkit.AssertEqual(t, dir, "desc")
}

func TestParseSortOrder_CaseInsensitive(t *testing.T) {
	q := url.Values{"sort": {"NAME"}, "order": {"ASC"}}
	field, dir := ParseSortOrder(q, []string{"name", "created_at"}, "created_at", "desc")
	testkit.AssertEqual(t, field, "name")
	testkit.AssertEqual(t, dir, "asc")
}

func TestParseSortOrder_EmptyAllowed(t *testing.T) {
	q := url.Values{"sort": {"name"}, "order": {"asc"}}
	field, dir := ParseSortOrder(q, []string{}, "created_at", "desc")
	testkit.AssertEqual(t, field, "created_at")
	testkit.AssertEqual(t, dir, "asc")
}
