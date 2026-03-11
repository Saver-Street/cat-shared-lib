package sorting

import (
	"net/url"
	"testing"
)

func BenchmarkParse_SingleField(b *testing.B) {
	q := url.Values{"sort": {"name"}, "order": {"asc"}}
	cfg := Config{Allowed: []string{"name", "email", "created_at"}, DefaultField: "name", DefaultDirection: Asc}
	for b.Loop() {
		Parse(q, cfg)
	}
}

func BenchmarkParse_MultiField(b *testing.B) {
	q := url.Values{"sort": {"name:asc,email:desc,created_at:asc"}}
	cfg := Config{
		Allowed:   []string{"name", "email", "created_at"},
		MaxFields: 3,
	}
	for b.Loop() {
		Parse(q, cfg)
	}
}

func BenchmarkParse_Default(b *testing.B) {
	q := url.Values{}
	cfg := Config{Allowed: []string{"name", "email"}, DefaultField: "name", DefaultDirection: Asc}
	for b.Loop() {
		Parse(q, cfg)
	}
}

func BenchmarkParse_InvalidField(b *testing.B) {
	q := url.Values{"sort": {"unknown_col"}}
	cfg := Config{Allowed: []string{"name", "email"}, DefaultField: "name", DefaultDirection: Asc}
	for b.Loop() {
		Parse(q, cfg)
	}
}

func BenchmarkOrderByClause(b *testing.B) {
	p := Params{Fields: []Field{
		{Name: "name", Direction: Asc},
		{Name: "created_at", Direction: Desc},
	}}
	for b.Loop() {
		p.OrderByClause()
	}
}

func BenchmarkOrderBySQL(b *testing.B) {
	p := Params{Fields: []Field{
		{Name: "name", Direction: Asc},
		{Name: "created_at", Direction: Desc},
	}}
	for b.Loop() {
		OrderBySQL(p)
	}
}

func BenchmarkHasField(b *testing.B) {
	p := Params{Fields: []Field{
		{Name: "name", Direction: Asc},
		{Name: "email", Direction: Desc},
		{Name: "created_at", Direction: Asc},
	}}
	for b.Loop() {
		p.HasField("created_at")
	}
}

func BenchmarkField_String(b *testing.B) {
	f := Field{Name: "created_at", Direction: Desc}
	for b.Loop() {
		_ = f.String()
	}
}
