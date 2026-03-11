package sorting_test

import (
	"fmt"
	"net/url"

	"github.com/Saver-Street/cat-shared-lib/sorting"
)

func ExampleParse() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
	}
	q := url.Values{"sort": {"name"}, "order": {"asc"}}
	p := sorting.Parse(q, cfg)
	fmt.Println(p.OrderByClause())
	// Output:
	// name asc
}

func ExampleParse_defaults() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
	}
	q := url.Values{}
	p := sorting.Parse(q, cfg)
	fmt.Println(p.OrderByClause())
	// Output:
	// created_at desc
}

func ExampleParse_multiField() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at", "updated_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
		MaxFields:        3,
	}
	q := url.Values{"sort": {"name:asc,created_at:desc"}}
	p := sorting.Parse(q, cfg)
	fmt.Println(p.OrderByClause())
	// Output:
	// name asc, created_at desc
}

func ExampleParams_HasField() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
	}
	q := url.Values{"sort": {"name"}}
	p := sorting.Parse(q, cfg)
	fmt.Println(p.HasField("name"))
	fmt.Println(p.HasField("email"))
	// Output:
	// true
	// false
}

func ExampleOrderBySQL() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
	}
	q := url.Values{"sort": {"name"}}
	p := sorting.Parse(q, cfg)
	fmt.Println(sorting.OrderBySQL(p))
	// Output:
	// ORDER BY name desc
}

func ExampleField_String() {
	f := sorting.Field{Name: "created_at", Direction: sorting.Desc}
	fmt.Println(f.String())
	// Output:
	// created_at desc
}
