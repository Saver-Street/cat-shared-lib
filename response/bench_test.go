package response

import (
	"net/http/httptest"
	"testing"
)

func BenchmarkPaginated(b *testing.B) {
	items := make([]int, 25)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		w := httptest.NewRecorder()
		Paginated(w, items, 100, 1, 25)
	}
}

func BenchmarkSetPaginationHeaders(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		SetPaginationHeaders(w, 100, 25, 0)
	}
}

func BenchmarkPaginatedWithHeaders(b *testing.B) {
	items := make([]int, 25)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		w := httptest.NewRecorder()
		PaginatedWithHeaders(w, items, 100, 1, 25)
	}
}
