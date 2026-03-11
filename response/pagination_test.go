package response

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestSetPaginationHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	SetPaginationHeaders(rr, 100, 20, 40)

	testkit.AssertHeader(t, rr, "X-Total-Count", "100")
	testkit.AssertHeader(t, rr, "X-Limit", "20")
	testkit.AssertHeader(t, rr, "X-Offset", "40")
}

func TestSetPaginationHeaders_Zero(t *testing.T) {
	rr := httptest.NewRecorder()
	SetPaginationHeaders(rr, 0, 10, 0)

	testkit.AssertHeader(t, rr, "X-Total-Count", "0")
}

func TestPaginatedWithHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"a", "b", "c"}
	PaginatedWithHeaders(rr, data, 50, 3, 10)

	testkit.AssertHeader(t, rr, "X-Total-Count", "50")
	testkit.AssertHeader(t, rr, "X-Limit", "10")
	testkit.AssertHeader(t, rr, "X-Offset", "20")

	var result PagedResult[string]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	testkit.AssertEqual(t, result.Total, 50)
	testkit.AssertEqual(t, result.Page, 3)
	testkit.AssertEqual(t, result.Limit, 10)
	testkit.AssertTrue(t, result.HasMore)
}

func TestPaginatedWithHeaders_Page1(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []int{1, 2, 3}
	PaginatedWithHeaders(rr, data, 3, 1, 10)

	testkit.AssertHeader(t, rr, "X-Offset", "0")
}

func TestPaginatedWithHeaders_LastPage(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"z"}
	PaginatedWithHeaders(rr, data, 21, 3, 10)

	var result PagedResult[string]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	testkit.AssertFalse(t, result.HasMore)
}

func TestSetLinkHeader_MiddlePage(t *testing.T) {
	rr := httptest.NewRecorder()
	SetLinkHeader(rr, "https://api.example.com/items", 3, 10, 50)
	link := rr.Header().Get("Link")

	testkit.AssertContains(t, link, `rel="first"`)
	testkit.AssertContains(t, link, `rel="prev"`)
	testkit.AssertContains(t, link, `rel="next"`)
	testkit.AssertContains(t, link, `rel="last"`)
	testkit.AssertContains(t, link, "page=2") // prev
	testkit.AssertContains(t, link, "page=4") // next
	testkit.AssertContains(t, link, "page=5") // last
	testkit.AssertContains(t, link, "limit=10")
}

func TestSetLinkHeader_FirstPage(t *testing.T) {
	rr := httptest.NewRecorder()
	SetLinkHeader(rr, "https://api.example.com/items", 1, 10, 30)
	link := rr.Header().Get("Link")

	testkit.AssertContains(t, link, `rel="first"`)
	testkit.AssertContains(t, link, `rel="next"`)
	testkit.AssertContains(t, link, `rel="last"`)
	// No prev on first page
	testkit.AssertFalse(t, strings.Contains(link, `rel="prev"`))
}

func TestSetLinkHeader_LastPage(t *testing.T) {
	rr := httptest.NewRecorder()
	SetLinkHeader(rr, "https://api.example.com/items", 3, 10, 30)
	link := rr.Header().Get("Link")

	testkit.AssertContains(t, link, `rel="prev"`)
	// No next on last page
	testkit.AssertFalse(t, strings.Contains(link, `rel="next"`))
}

func TestSetLinkHeader_ZeroTotal(t *testing.T) {
	rr := httptest.NewRecorder()
	SetLinkHeader(rr, "https://api.example.com/items", 1, 10, 0)
	testkit.AssertEqual(t, rr.Header().Get("Link"), "")
}

func TestCursorPaginated_WithMore(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"a", "b", "c"}
	CursorPaginated(rr, data, "eyJpZCI6MTB9")

	var result CursorResult[string]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	testkit.AssertEqual(t, len(result.Data), 3)
	testkit.AssertEqual(t, result.NextCursor, "eyJpZCI6MTB9")
	testkit.AssertTrue(t, result.HasMore)
}

func TestCursorPaginated_LastPage(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"z"}
	CursorPaginated(rr, data, "")

	var result CursorResult[string]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	testkit.AssertEqual(t, result.NextCursor, "")
	testkit.AssertFalse(t, result.HasMore)
}

func TestCursorPaginated_EmptyData(t *testing.T) {
	rr := httptest.NewRecorder()
	CursorPaginated(rr, []int{}, "")

	var result CursorResult[int]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	testkit.AssertEqual(t, len(result.Data), 0)
	testkit.AssertFalse(t, result.HasMore)
}
