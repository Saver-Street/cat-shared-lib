package servicetest_test

import (
	"fmt"
	"sort"

	"github.com/Saver-Street/cat-shared-lib/servicetest"
)

func ExampleNewFixtures() {
	f := servicetest.NewFixtures()

	f.Register("user", []byte(`{"name":"Alice","age":30}`))
	raw, _ := f.Load("user")
	fmt.Println(string(raw))
	// Output:
	// {"name":"Alice","age":30}
}

func ExampleFixtures_RegisterJSON() {
	f := servicetest.NewFixtures()

	_ = f.RegisterJSON("config", map[string]string{"env": "test"})
	raw, _ := f.Load("config")
	fmt.Println(string(raw))
	// Output:
	// {"env":"test"}
}

func ExampleFixtures_LoadInto() {
	f := servicetest.NewFixtures()
	f.Register("item", []byte(`{"id":1,"name":"widget"}`))

	var item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	_ = f.LoadInto("item", &item)
	fmt.Printf("%d %s\n", item.ID, item.Name)
	// Output:
	// 1 widget
}

func ExampleFixtures_Names() {
	f := servicetest.NewFixtures()
	f.Register("alpha", []byte(`{}`))
	f.Register("beta", []byte(`{}`))

	names := f.Names()
	sort.Strings(names)
	fmt.Println(names)
	// Output:
	// [alpha beta]
}

func ExampleMockRow_Scan() {
	row := &servicetest.MockRow{ScanValues: []any{"alice", 42}}

	var name string
	var age int
	_ = row.Scan(&name, &age)
	fmt.Println(name, age)
	// Output:
	// alice 42
}
