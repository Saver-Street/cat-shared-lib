package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestSetPaginationHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	SetPaginationHeaders(rr, 100, 20, 40)

	if got := rr.Header().Get("X-Total-Count"); got != "100" {
		t.Errorf("X-Total-Count = %q, want %q", got, "100")
	}
	if got := rr.Header().Get("X-Limit"); got != "20" {
		t.Errorf("X-Limit = %q, want %q", got, "20")
	}
	if got := rr.Header().Get("X-Offset"); got != "40" {
		t.Errorf("X-Offset = %q, want %q", got, "40")
	}
}

func TestSetPaginationHeaders_Zero(t *testing.T) {
	rr := httptest.NewRecorder()
	SetPaginationHeaders(rr, 0, 10, 0)

	if got := rr.Header().Get("X-Total-Count"); got != "0" {
		t.Errorf("X-Total-Count = %q, want %q", got, "0")
	}
}

func TestPaginatedWithHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"a", "b", "c"}
	PaginatedWithHeaders(rr, data, 50, 3, 10)

	if got := rr.Header().Get("X-Total-Count"); got != "50" {
		t.Errorf("X-Total-Count = %q, want %q", got, "50")
	}
	if got := rr.Header().Get("X-Limit"); got != "10" {
		t.Errorf("X-Limit = %q, want %q", got, "10")
	}
	if got := rr.Header().Get("X-Offset"); got != "20" {
		t.Errorf("X-Offset = %q, want %q", got, "20")
	}

	var result PagedResult[string]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 50 || result.Page != 3 || result.Limit != 10 {
		t.Errorf("body mismatch: %+v", result)
	}
	if !result.HasMore {
		t.Error("expected HasMore=true")
	}
}

func TestPaginatedWithHeaders_Page1(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []int{1, 2, 3}
	PaginatedWithHeaders(rr, data, 3, 1, 10)

	if got := rr.Header().Get("X-Offset"); got != "0" {
		t.Errorf("X-Offset = %q, want %q for page 1", got, "0")
	}
}

func TestPaginatedWithHeaders_LastPage(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"z"}
	PaginatedWithHeaders(rr, data, 21, 3, 10)

	var result PagedResult[string]
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.HasMore {
		t.Error("expected HasMore=false on last page")
	}
}
