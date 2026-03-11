package sorting

import (
	"net/url"
	"testing"
)

func TestParse_SingleField(t *testing.T) {
	q := url.Values{"sort": {"name"}, "order": {"desc"}}
	cfg := Config{Allowed: []string{"name", "email"}, DefaultField: "name", DefaultDirection: Asc}
	p := Parse(q, cfg)

	if len(p.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(p.Fields))
	}
	if p.Fields[0].Name != "name" || p.Fields[0].Direction != Desc {
		t.Errorf("got %v, want name desc", p.Fields[0])
	}
}

func TestParse_SingleField_DefaultOrder(t *testing.T) {
	q := url.Values{"sort": {"email"}}
	cfg := Config{Allowed: []string{"name", "email"}, DefaultField: "name", DefaultDirection: Asc}
	p := Parse(q, cfg)

	if len(p.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(p.Fields))
	}
	if p.Fields[0].Direction != Asc {
		t.Errorf("got %v, want asc", p.Fields[0].Direction)
	}
}

func TestParse_MultiField(t *testing.T) {
	q := url.Values{"sort": {"name:asc,created_at:desc"}}
	cfg := Config{
		Allowed:   []string{"name", "created_at", "email"},
		MaxFields: 3,
	}
	p := Parse(q, cfg)

	if len(p.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(p.Fields))
	}
	if p.Fields[0].Name != "name" || p.Fields[0].Direction != Asc {
		t.Errorf("field 0: got %v", p.Fields[0])
	}
	if p.Fields[1].Name != "created_at" || p.Fields[1].Direction != Desc {
		t.Errorf("field 1: got %v", p.Fields[1])
	}
}

func TestParse_MaxFields(t *testing.T) {
	q := url.Values{"sort": {"name:asc,email:desc,created_at:asc"}}
	cfg := Config{
		Allowed:   []string{"name", "email", "created_at"},
		MaxFields: 2,
	}
	p := Parse(q, cfg)
	if len(p.Fields) != 2 {
		t.Errorf("expected max 2 fields, got %d", len(p.Fields))
	}
}

func TestParse_InvalidFieldDropped(t *testing.T) {
	q := url.Values{"sort": {"hacked_col"}}
	cfg := Config{Allowed: []string{"name", "email"}, DefaultField: "name"}
	p := Parse(q, cfg)

	if len(p.Fields) != 1 {
		t.Fatalf("expected default field, got %d fields", len(p.Fields))
	}
	if p.Fields[0].Name != "name" {
		t.Errorf("expected default field name, got %s", p.Fields[0].Name)
	}
}

func TestParse_Empty(t *testing.T) {
	q := url.Values{}
	cfg := Config{Allowed: []string{"name"}, DefaultField: "name", DefaultDirection: Desc}
	p := Parse(q, cfg)

	if len(p.Fields) != 1 || p.Fields[0].Name != "name" || p.Fields[0].Direction != Desc {
		t.Errorf("expected default name desc, got %v", p.Fields)
	}
}

func TestParse_NoDefault(t *testing.T) {
	q := url.Values{}
	cfg := Config{Allowed: []string{"name"}}
	p := Parse(q, cfg)
	if len(p.Fields) != 0 {
		t.Errorf("expected no fields, got %d", len(p.Fields))
	}
}

func TestParse_CaseInsensitive(t *testing.T) {
	q := url.Values{"sort": {"NAME"}, "order": {"DESC"}}
	cfg := Config{Allowed: []string{"name"}, DefaultField: "name"}
	p := Parse(q, cfg)
	if p.Fields[0].Name != "name" {
		t.Errorf("expected canonical name 'name', got %s", p.Fields[0].Name)
	}
	if p.Fields[0].Direction != Desc {
		t.Errorf("expected desc, got %v", p.Fields[0].Direction)
	}
}

func TestOrderByClause(t *testing.T) {
	p := Params{Fields: []Field{
		{Name: "name", Direction: Asc},
		{Name: "created_at", Direction: Desc},
	}}
	got := p.OrderByClause()
	want := "name asc, created_at desc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOrderByClause_Empty(t *testing.T) {
	p := Params{}
	if got := p.OrderByClause(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestOrderBySQL(t *testing.T) {
	p := Params{Fields: []Field{{Name: "name", Direction: Asc}}}
	got := OrderBySQL(p)
	want := "ORDER BY name asc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOrderBySQL_Empty(t *testing.T) {
	p := Params{}
	if got := OrderBySQL(p); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestHasField(t *testing.T) {
	p := Params{Fields: []Field{
		{Name: "name", Direction: Asc},
		{Name: "email", Direction: Desc},
	}}
	if !p.HasField("name") {
		t.Error("expected HasField name = true")
	}
	if p.HasField("phone") {
		t.Error("expected HasField phone = false")
	}
}

func TestField_String(t *testing.T) {
	f := Field{Name: "created_at", Direction: Desc}
	if f.String() != "created_at desc" {
		t.Errorf("got %q", f.String())
	}
}
