package request

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

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

func TestParseDateParam_RFC3339(t *testing.T) {
	q := url.Values{"start": {"2024-06-15T10:30:00Z"}}
	got, err := ParseDateParam(q, "start")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC))
}

func TestParseDateParam_DateOnly(t *testing.T) {
	q := url.Values{"start": {"2024-06-15"}}
	got, err := ParseDateParam(q, "start")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC))
}

func TestParseDateParam_Missing(t *testing.T) {
	q := url.Values{}
	got, err := ParseDateParam(q, "start")
	testkit.AssertNoError(t, err)
	testkit.AssertTrue(t, got.IsZero())
}

func TestParseDateParam_Empty(t *testing.T) {
	q := url.Values{"start": {""}}
	got, err := ParseDateParam(q, "start")
	testkit.AssertNoError(t, err)
	testkit.AssertTrue(t, got.IsZero())
}

func TestParseDateParam_Invalid(t *testing.T) {
	q := url.Values{"start": {"not-a-date"}}
	_, err := ParseDateParam(q, "start")
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "invalid date format")
}

func TestParseDateParam_WhitespaceTrimmed(t *testing.T) {
	q := url.Values{"start": {"  2024-06-15  "}}
	got, err := ParseDateParam(q, "start")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC))
}

func TestParseEnumParam_Match(t *testing.T) {
	q := url.Values{"status": {"active"}}
	got := ParseEnumParam(q, "status", []string{"active", "inactive", "pending"}, "active")
	testkit.AssertEqual(t, got, "active")
}

func TestParseEnumParam_CaseInsensitive(t *testing.T) {
	q := url.Values{"status": {"ACTIVE"}}
	got := ParseEnumParam(q, "status", []string{"active", "inactive"}, "inactive")
	testkit.AssertEqual(t, got, "active")
}

func TestParseEnumParam_NotAllowed(t *testing.T) {
	q := url.Values{"status": {"deleted"}}
	got := ParseEnumParam(q, "status", []string{"active", "inactive"}, "active")
	testkit.AssertEqual(t, got, "active")
}

func TestParseEnumParam_Missing(t *testing.T) {
	q := url.Values{}
	got := ParseEnumParam(q, "status", []string{"active", "inactive"}, "inactive")
	testkit.AssertEqual(t, got, "inactive")
}

func TestParseEnumParam_Empty(t *testing.T) {
	q := url.Values{"status": {""}}
	got := ParseEnumParam(q, "status", []string{"active", "inactive"}, "active")
	testkit.AssertEqual(t, got, "active")
}

func TestExtractBearerToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer abc123")
	token, ok := ExtractBearerToken(r)
	testkit.AssertTrue(t, ok)
	testkit.AssertEqual(t, token, "abc123")
}

func TestExtractBearerToken_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	token, ok := ExtractBearerToken(r)
	testkit.AssertFalse(t, ok)
	testkit.AssertEqual(t, token, "")
}

func TestExtractBearerToken_WrongScheme(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	token, ok := ExtractBearerToken(r)
	testkit.AssertFalse(t, ok)
	testkit.AssertEqual(t, token, "")
}

func TestExtractBearerToken_CaseInsensitive(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "bearer xyz789")
	token, ok := ExtractBearerToken(r)
	testkit.AssertTrue(t, ok)
	testkit.AssertEqual(t, token, "xyz789")
}

func TestContentType(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	testkit.AssertEqual(t, ContentType(r), "application/json")
}

func TestContentType_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	testkit.AssertEqual(t, ContentType(r), "")
}

func TestContentType_Plain(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "text/plain")
	testkit.AssertEqual(t, ContentType(r), "text/plain")
}

func TestIsJSON_True(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "application/json")
	testkit.AssertTrue(t, IsJSON(r))
}

func TestIsJSON_False(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "text/html")
	testkit.AssertFalse(t, IsJSON(r))
}

func TestClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.1, 70.41.3.18, 150.172.238.178")
	testkit.AssertEqual(t, ClientIP(r), "203.0.113.1")
}

func TestClientIP_XForwardedFor_Single(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "10.0.0.1")
	testkit.AssertEqual(t, ClientIP(r), "10.0.0.1")
}

func TestClientIP_XRealIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "192.168.1.100")
	testkit.AssertEqual(t, ClientIP(r), "192.168.1.100")
}

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:54321"
	testkit.AssertEqual(t, ClientIP(r), "10.0.0.5")
}

func TestClientIP_XForwardedFor_Priority(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "1.1.1.1")
	r.Header.Set("X-Real-IP", "2.2.2.2")
	r.RemoteAddr = "3.3.3.3:9999"
	testkit.AssertEqual(t, ClientIP(r), "1.1.1.1")
}

func TestParseJSONBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	body := strings.NewReader(`{"name":"Alice","age":30}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)
	got, err := ParseJSONBody[payload](r)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got.Name, "Alice")
	testkit.AssertEqual(t, got.Age, 30)
}

func TestParseJSONBody_Invalid(t *testing.T) {
	body := strings.NewReader(`{invalid`)
	r := httptest.NewRequest(http.MethodPost, "/", body)
	_, err := ParseJSONBody[map[string]string](r)
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "invalid JSON body")
}

func TestOptionalQueryFloat_Present(t *testing.T) {
	q := url.Values{"price": {"19.99"}}
	testkit.AssertEqual(t, OptionalQueryFloat(q, "price", 0), 19.99)
}

func TestOptionalQueryFloat_Missing(t *testing.T) {
	q := url.Values{}
	testkit.AssertEqual(t, OptionalQueryFloat(q, "price", 9.99), 9.99)
}

func TestOptionalQueryFloat_Invalid(t *testing.T) {
	q := url.Values{"price": {"abc"}}
	testkit.AssertEqual(t, OptionalQueryFloat(q, "price", 5.0), 5.0)
}

func TestOptionalQueryFloat_Zero(t *testing.T) {
	q := url.Values{"price": {"0"}}
	testkit.AssertEqual(t, OptionalQueryFloat(q, "price", 10.0), 0.0)
}

func TestOptionalQueryBool_True(t *testing.T) {
	for _, v := range []string{"true", "1", "yes", "TRUE", "Yes"} {
		q := url.Values{"active": {v}}
		got := OptionalQueryBool(q, "active")
		testkit.RequireNotNil(t, got)
		testkit.AssertTrue(t, *got)
	}
}

func TestOptionalQueryBool_False(t *testing.T) {
	for _, v := range []string{"false", "0", "no", "FALSE", "No"} {
		q := url.Values{"active": {v}}
		got := OptionalQueryBool(q, "active")
		testkit.RequireNotNil(t, got)
		testkit.AssertFalse(t, *got)
	}
}

func TestOptionalQueryBool_Absent(t *testing.T) {
	q := url.Values{}
	got := OptionalQueryBool(q, "active")
	testkit.AssertNil(t, got)
}

func TestOptionalQueryBool_Unrecognised(t *testing.T) {
	q := url.Values{"active": {"maybe"}}
	got := OptionalQueryBool(q, "active")
	testkit.AssertNil(t, got)
}

func TestRequireHeader_Present(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "acme")
	v, err := RequireHeader(req, "X-Tenant-ID")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, v, "acme")
}

func TestRequireHeader_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := RequireHeader(req, "X-Tenant-ID")
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "X-Tenant-ID")
}

func TestRequireHeader_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "  ")
	_, err := RequireHeader(req, "X-Tenant-ID")
	testkit.AssertError(t, err)
}

func TestParseDateRange_Valid(t *testing.T) {
	q := url.Values{
		"start": {"2024-01-01"},
		"end":   {"2024-12-31"},
	}
	dr, err := ParseDateRange(q, "start", "end")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, dr.Start.Year(), 2024)
	testkit.AssertEqual(t, dr.End.Month(), time.December)
}

func TestParseDateRange_StartAfterEnd(t *testing.T) {
	q := url.Values{
		"start": {"2024-12-31"},
		"end":   {"2024-01-01"},
	}
	_, err := ParseDateRange(q, "start", "end")
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "before")
}

func TestParseDateRange_BothAbsent(t *testing.T) {
	dr, err := ParseDateRange(url.Values{}, "start", "end")
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, dr.Start.IsZero())
	testkit.AssertTrue(t, dr.End.IsZero())
}

func TestParseDateRange_OnePresent(t *testing.T) {
	q := url.Values{"start": {"2024-06-15"}}
	dr, err := ParseDateRange(q, "start", "end")
	testkit.RequireNoError(t, err)
	testkit.AssertFalse(t, dr.Start.IsZero())
	testkit.AssertTrue(t, dr.End.IsZero())
}

func TestParseIDList_Valid(t *testing.T) {
	q := url.Values{"ids": {"550e8400-e29b-41d4-a716-446655440000,6ba7b810-9dad-11d1-80b4-00c04fd430c8"}}
	got, err := ParseIDList(q, "ids")
	testkit.AssertNoError(t, err)
	testkit.AssertLen(t, got, 2)
	testkit.AssertEqual(t, got[0], "550e8400-e29b-41d4-a716-446655440000")
}

func TestParseIDList_Missing(t *testing.T) {
	got, err := ParseIDList(url.Values{}, "ids")
	testkit.AssertNoError(t, err)
	testkit.AssertNil(t, got)
}

func TestParseIDList_InvalidUUID(t *testing.T) {
	q := url.Values{"ids": {"550e8400-e29b-41d4-a716-446655440000,not-a-uuid"}}
	_, err := ParseIDList(q, "ids")
	testkit.AssertErrorContains(t, err, "not a valid UUID")
}

// --- ParseCursor ---

func TestParseCursor_Defaults(t *testing.T) {
	c := ParseCursor(url.Values{}, 20, 100)
	testkit.AssertEqual(t, c.Cursor, "")
	testkit.AssertEqual(t, c.Limit, 20)
}

func TestParseCursor_WithValues(t *testing.T) {
	q := url.Values{"cursor": {"abc"}, "limit": {"15"}}
	c := ParseCursor(q, 20, 100)
	testkit.AssertEqual(t, c.Cursor, "abc")
	testkit.AssertEqual(t, c.Limit, 15)
}

func TestParseCursor_MaxLimit(t *testing.T) {
	q := url.Values{"limit": {"200"}}
	c := ParseCursor(q, 20, 50)
	testkit.AssertEqual(t, c.Limit, 50)
}

func TestParseCursor_InvalidLimit(t *testing.T) {
	q := url.Values{"limit": {"abc"}}
	c := ParseCursor(q, 25, 100)
	testkit.AssertEqual(t, c.Limit, 25)
}

func TestParseCursor_ZeroLimit(t *testing.T) {
	q := url.Values{"limit": {"0"}}
	c := ParseCursor(q, 20, 100)
	testkit.AssertEqual(t, c.Limit, 20)
}

func TestParseCursor_NegativeLimit(t *testing.T) {
	q := url.Values{"limit": {"-5"}}
	c := ParseCursor(q, 20, 100)
	testkit.AssertEqual(t, c.Limit, 20)
}

func TestParseCursor_NegativeDefaults(t *testing.T) {
	c := ParseCursor(url.Values{}, -1, -1)
	testkit.AssertEqual(t, c.Limit, 20)
}

func TestParseCursor_CursorOnly(t *testing.T) {
	q := url.Values{"cursor": {"tok_xyz"}}
	c := ParseCursor(q, 10, 50)
	testkit.AssertEqual(t, c.Cursor, "tok_xyz")
	testkit.AssertEqual(t, c.Limit, 10)
}
