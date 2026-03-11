package request

import (
	"net/url"
	"testing"
)

func BenchmarkParsePagination_Defaults(b *testing.B) {
	q := url.Values{}
	for b.Loop() {
		ParsePagination(q, 20, 100)
	}
}

func BenchmarkOptionalQueryParam(b *testing.B) {
	q := url.Values{"search": {"hello world"}}
	for b.Loop() {
		OptionalQueryParam(q, "search", "")
	}
}

func BenchmarkOptionalQueryInt(b *testing.B) {
	q := url.Values{"count": {"42"}}
	for b.Loop() {
		OptionalQueryInt(q, "count", 0)
	}
}

func BenchmarkParseCommaSeparated(b *testing.B) {
	q := url.Values{"tags": {"go,rust,python,typescript,java"}}
	for b.Loop() {
		ParseCommaSeparated(q, "tags")
	}
}

func BenchmarkParseCommaSeparatedInts(b *testing.B) {
	q := url.Values{"ids": {"1,2,3,4,5,6,7,8,9,10"}}
	for b.Loop() {
		ParseCommaSeparatedInts(q, "ids")
	}
}

func BenchmarkParseSortOrder(b *testing.B) {
	q := url.Values{"sort": {"name"}, "dir": {"desc"}}
	allowed := []string{"name", "created_at", "updated_at", "email"}
	for b.Loop() {
		ParseSortOrder(q, allowed, "created_at", "asc")
	}
}

func BenchmarkRequireQueryParamInt(b *testing.B) {
	q := url.Values{"id": {"12345"}}
	for b.Loop() {
		RequireQueryParamInt(q, "id")
	}
}
