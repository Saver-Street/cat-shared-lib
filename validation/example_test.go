package validation_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/validation"
)

func ExampleEmail() {
	fmt.Println(validation.Email("email", "user@example.com"))
	fmt.Println(validation.Email("email", "invalid"))
	// Output:
	// <nil>
	// email: invalid email format
}

func ExampleRequired() {
	fmt.Println(validation.Required("name", "Alice"))
	fmt.Println(validation.Required("name", ""))
	// Output:
	// <nil>
	// name: name is required
}

func ExampleMinLength() {
	fmt.Println(validation.MinLength("password", "abc", 8))
	// Output:
	// password: password must be at least 8 characters
}

func ExampleURL() {
	fmt.Println(validation.URL("website", "https://example.com"))
	fmt.Println(validation.URL("website", "not-a-url"))
	// Output:
	// <nil>
	// website: URL must use http or https scheme
}
