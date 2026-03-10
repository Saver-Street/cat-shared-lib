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
