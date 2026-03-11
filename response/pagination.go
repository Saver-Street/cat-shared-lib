package response

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// SetPaginationHeaders writes standard pagination headers to the response:
//   - X-Total-Count: total number of matching items
//   - X-Limit: maximum items per page
//   - X-Offset: zero-based offset into the result set
func SetPaginationHeaders(w http.ResponseWriter, total, limit, offset int) {
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Offset", strconv.Itoa(offset))
}

// PaginatedWithHeaders sends a paginated JSON response and sets the
// X-Total-Count, X-Limit, and X-Offset headers. The offset is derived
// from page and limit: offset = (page - 1) * limit.
func PaginatedWithHeaders[T any](w http.ResponseWriter, data []T, total, page, limit int) {
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	SetPaginationHeaders(w, total, limit, offset)
	Paginated(w, data, total, page, limit)
}

// SetLinkHeader writes a Link header (RFC 5988) with first, prev, next, and
// last pagination links derived from the given base URL, current page, limit,
// and total item count. Only applicable links are included (e.g. prev is
// omitted on page 1).
func SetLinkHeader(w http.ResponseWriter, baseURL string, page, limit, total int) {
	if total <= 0 || limit <= 0 {
		return
	}
	lastPage := (total + limit - 1) / limit
	if lastPage < 1 {
		lastPage = 1
	}

	var links []string
	add := func(p int, rel string) {
		links = append(links, fmt.Sprintf(`<%s?page=%d&limit=%d>; rel="%s"`, baseURL, p, limit, rel))
	}

	add(1, "first")
	if page > 1 {
		add(page-1, "prev")
	}
	if page < lastPage {
		add(page+1, "next")
	}
	add(lastPage, "last")

	w.Header().Set("Link", strings.Join(links, ", "))
}

// CursorResult is a response envelope for cursor-based pagination.
type CursorResult[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// CursorPaginated sends a 200 JSON response wrapped in a CursorResult envelope.
// nextCursor is the opaque cursor for the next page; pass "" on the last page.
func CursorPaginated[T any](w http.ResponseWriter, data []T, nextCursor string) {
	JSON(w, http.StatusOK, CursorResult[T]{
		Data:       data,
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}
