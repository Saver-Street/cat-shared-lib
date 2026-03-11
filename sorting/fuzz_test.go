package sorting

import (
	"net/url"
	"testing"
)

func FuzzParse(f *testing.F) {
	f.Add("name", "asc")
	f.Add("name:asc,email:desc", "")
	f.Add("", "")
	f.Add("NAME", "DESC")
	f.Add("unknown", "asc")
	f.Add("name:asc,name:desc,email:asc", "")
	f.Add(",,,", "")
	f.Add("name:", "")
	f.Add(":asc", "")
	f.Add("  name  ", "  desc  ")

	f.Fuzz(func(t *testing.T, sort, order string) {
		q := url.Values{}
		if sort != "" {
			q.Set("sort", sort)
		}
		if order != "" {
			q.Set("order", order)
		}
		cfg := Config{
			Allowed:          []string{"name", "email", "created_at"},
			DefaultField:     "name",
			DefaultDirection: Asc,
			MaxFields:        3,
		}
		p := Parse(q, cfg)

		// Must not panic; results must be internally consistent.
		_ = p.OrderByClause()
		_ = OrderBySQL(p)
		for _, field := range p.Fields {
			_ = field.String()
			if field.Direction != Asc && field.Direction != Desc {
				t.Errorf("unexpected direction %q", field.Direction)
			}
			if !p.HasField(field.Name) {
				t.Errorf("HasField(%q) returned false for present field", field.Name)
			}
		}
	})
}
