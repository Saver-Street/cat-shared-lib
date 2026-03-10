package scan_test

import (
	"errors"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/scan"
)

// exampleRow implements scan.SingleRowScanner for examples.
type exampleRow struct {
	data []any
	err  error
}

func (r *exampleRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i, d := range dest {
		if i < len(r.data) {
			switch ptr := d.(type) {
			case *string:
				*ptr = r.data[i].(string)
			case *int:
				*ptr = r.data[i].(int)
			}
		}
	}
	return nil
}

// exampleRows implements scan.RowScanner for examples.
type exampleRows struct {
	data [][]any
	idx  int
}

func (r *exampleRows) Next() bool {
	if r.idx < len(r.data) {
		r.idx++
		return true
	}
	return false
}

func (r *exampleRows) Scan(dest ...any) error {
	row := r.data[r.idx-1]
	for i, d := range dest {
		if i < len(row) {
			switch ptr := d.(type) {
			case *string:
				*ptr = row[i].(string)
			case *int:
				*ptr = row[i].(int)
			}
		}
	}
	return nil
}

func (r *exampleRows) Close()     {}
func (r *exampleRows) Err() error { return nil }

type person struct {
	Name string
	City string
}

func ExampleRow() {
	row := &exampleRow{data: []any{"Alice", "Portland"}}
	p, err := scan.Row[person](row, func(p *person) []any {
		return []any{&p.Name, &p.City}
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s from %s\n", p.Name, p.City)
	// Output:
	// Alice from Portland
}

func ExampleRow_error() {
	row := &exampleRow{err: errors.New("no rows")}
	_, err := scan.Row[person](row, func(p *person) []any {
		return []any{&p.Name, &p.City}
	})
	fmt.Println(err)
	// Output:
	// no rows
}

func ExampleRows() {
	rows := &exampleRows{data: [][]any{
		{"Alice", "Portland"},
		{"Bob", "Seattle"},
	}}
	people, err := scan.Rows[person](rows, func(p *person) []any {
		return []any{&p.Name, &p.City}
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, p := range people {
		fmt.Printf("%s from %s\n", p.Name, p.City)
	}
	// Output:
	// Alice from Portland
	// Bob from Seattle
}

func ExampleRows_empty() {
	rows := &exampleRows{data: [][]any{}}
	people, err := scan.Rows[person](rows, func(p *person) []any {
		return []any{&p.Name, &p.City}
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("found %d people\n", len(people))
	// Output:
	// found 0 people
}

func ExampleFirst() {
	rows := &exampleRows{data: [][]any{
		{"Alice", "Portland"},
		{"Bob", "Seattle"},
	}}
	p, err := scan.First[person](rows, func(p *person) []any {
		return []any{&p.Name, &p.City}
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s from %s\n", p.Name, p.City)
	// Output:
	// Alice from Portland
}

func ExampleRowsLimit() {
	rows := &exampleRows{data: [][]any{
		{"Alice", "Portland"},
		{"Bob", "Seattle"},
		{"Charlie", "Denver"},
	}}
	people, err := scan.RowsLimit[person](rows, func(p *person) []any {
		return []any{&p.Name, &p.City}
	}, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, p := range people {
		fmt.Printf("%s from %s\n", p.Name, p.City)
	}
	// Output:
	// Alice from Portland
	// Bob from Seattle
}
