package response

import (
	"net/http"
	"strconv"
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
