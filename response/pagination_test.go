package response

import (
	"encoding/json"
	"net/http/httptest"
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
