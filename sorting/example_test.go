package sorting_test

import (
	"fmt"
	"net/url"

	"github.com/Saver-Street/cat-shared-lib/sorting"
)

func ExampleParse() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at", "email"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
		MaxFields:        3,
	}

	q := make(url.Values)
	q.Set("sort", "name:asc,created_at:desc")

	params := sorting.Parse(q, cfg)
	fmt.Println(params.OrderByClause())
	// Output: name asc, created_at desc
}

func ExampleParse_singleField() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Asc,
	}

	q := make(url.Values)
	q.Set("sort", "name")
	q.Set("order", "desc")

	params := sorting.Parse(q, cfg)
	fmt.Println(params.OrderByClause())
	// Output: name desc
}

func ExampleParse_default() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "created_at"},
		DefaultField:     "created_at",
		DefaultDirection: sorting.Desc,
	}

	params := sorting.Parse(make(url.Values), cfg)
	fmt.Println(params.OrderByClause())
	// Output: created_at desc
}

func ExampleOrderBySQL() {
	cfg := sorting.Config{
		Allowed:          []string{"name", "age"},
		DefaultField:     "name",
		DefaultDirection: sorting.Asc,
	}

	params := sorting.Parse(make(url.Values), cfg)
	fmt.Println(sorting.OrderBySQL(params))
	// Output: ORDER BY name asc
}

func ExampleParams_HasField() {
	cfg := sorting.Config{
		Allowed:   []string{"name", "age"},
		MaxFields: 2,
	}

	q := make(url.Values)
	q.Set("sort", "name:asc")

	params := sorting.Parse(q, cfg)
	fmt.Println(params.HasField("name"))
	fmt.Println(params.HasField("age"))
	// Output:
	// true
	// false
}

func ExampleField_String() {
	f := sorting.Field{Name: "created_at", Direction: sorting.Desc}
	fmt.Println(f.String())
	// Output: created_at desc
}
