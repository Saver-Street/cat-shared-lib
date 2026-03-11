package urlutil_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/urlutil"
)

func ExampleJoin() {
	fmt.Println(urlutil.Join("https://api.example.com", "v1", "users"))
	// Output: https://api.example.com/v1/users
}

func ExampleSetQuery() {
	fmt.Println(urlutil.SetQuery("https://example.com/search", "q", "golang"))
	// Output: https://example.com/search?q=golang
}

func ExampleDomain() {
	fmt.Println(urlutil.Domain("https://www.example.com:8080/path"))
	// Output: www.example.com
}
