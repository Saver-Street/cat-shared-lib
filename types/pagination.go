package types

// PaginationParams holds limit/offset for paginated queries.
type PaginationParams struct {
	// Limit is the maximum number of rows to return (1–100).
	Limit int `json:"limit"`
	// Offset is the number of rows to skip before collecting results.
	Offset int `json:"offset"`
	// Page is the 1-based page number used to derive Offset.
	Page int `json:"page"`
}

// HasNextPage reports whether there is at least one more page after the current one.
// total is the total number of items in the full result set.
func (p PaginationParams) HasNextPage(total int) bool {
	return p.Offset+p.Limit < total
}

// IsLastPage reports whether the current page is the final page.
// total is the total number of items in the full result set.
func (p PaginationParams) IsLastPage(total int) bool {
	return !p.HasNextPage(total)
}

// NormalizePage ensures page >= 1 and derives a sane limit (1–100).
func NormalizePage(page, limit int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return PaginationParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
		Page:   page,
	}
}

// TotalPages returns the total number of pages required to display all items at the
// current limit. Returns 0 when total is zero or negative.
func (p PaginationParams) TotalPages(total int) int {
	if total <= 0 {
		return 0
	}
	return (total + p.Limit - 1) / p.Limit
}

// CursorParams holds parameters for cursor-based pagination.
type CursorParams struct {
	// Cursor is the opaque cursor from the previous page ("" for the first page).
	Cursor string
	// Limit is the maximum number of items to return (1–100).
	Limit int
}

// NormalizeCursor ensures limit is within [1, 100], defaulting to 20.
func NormalizeCursor(cursor string, limit int) CursorParams {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return CursorParams{Cursor: cursor, Limit: limit}
}

// CursorPage represents a page of cursor-paginated results.
type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// NewCursorPage constructs a CursorPage from items. If len(items) > limit, it
// trims to limit and uses cursorFn on the last kept item to derive NextCursor.
func NewCursorPage[T any](items []T, limit int, cursorFn func(T) string) CursorPage[T] {
	if len(items) <= limit {
		return CursorPage[T]{Items: items}
	}
	kept := items[:limit]
	return CursorPage[T]{
		Items:      kept,
		NextCursor: cursorFn(kept[limit-1]),
		HasMore:    true,
	}
}

// ApplyOffset returns the sub-slice of items for the given offset/limit pagination.
// Returns nil if offset >= len(items).
func ApplyOffset[T any](items []T, offset, limit int) []T {
	if offset >= len(items) {
		return nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
