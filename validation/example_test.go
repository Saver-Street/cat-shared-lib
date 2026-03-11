package validation_test

import (
	"fmt"
	"regexp"

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

func ExampleSlug() {
	fmt.Println(validation.Slug("handle", "my-cool-page"))
	fmt.Println(validation.Slug("handle", "Not A Slug!"))
	// Output:
	// <nil>
	// handle: handle must contain only lowercase letters, digits, and hyphens
}

func ExampleMatch() {
	re := regexp.MustCompile(`^[A-Z]{3}-[0-9]{4}$`)
	fmt.Println(validation.Match("code", "ABC-1234", re, "format XXX-0000"))
	fmt.Println(validation.Match("code", "invalid", re, "format XXX-0000"))
	// Output:
	// <nil>
	// code: code must match format XXX-0000
}

func ExampleUUID() {
	fmt.Println(validation.UUID("id", "550e8400-e29b-41d4-a716-446655440000"))
	fmt.Println(validation.UUID("id", "not-a-uuid"))
	// Output:
	// <nil>
	// id: invalid UUID format
}

func ExamplePhone() {
	fmt.Println(validation.Phone("phone", "+1-555-123-4567"))
	fmt.Println(validation.Phone("phone", "123"))
	// Output:
	// <nil>
	// phone: phone number must contain 7-15 digits
}

func ExampleMaxLength() {
	fmt.Println(validation.MaxLength("bio", "short", 100))
	fmt.Println(validation.MaxLength("bio", "this is way too long", 5))
	// Output:
	// <nil>
	// bio: bio must be at most 5 characters
}

func ExampleOneOf() {
	fmt.Println(validation.OneOf("status", "active", []string{"active", "inactive", "pending"}))
	fmt.Println(validation.OneOf("status", "deleted", []string{"active", "inactive", "pending"}))
	// Output:
	// <nil>
	// status: status must be one of: active, inactive, pending
}

func ExampleCollect() {
	errs := validation.Collect(
		validation.Required("name", "Alice"),
		validation.Email("email", "bad-email"),
		validation.MinLength("password", "abc", 8),
	)
	for _, e := range errs {
		fmt.Println(e)
	}
	// Output:
	// email: invalid email format
	// password: password must be at least 8 characters
}

func ExampleDate() {
	fmt.Println(validation.Date("dob", "2024-01-15", "2006-01-02"))
	fmt.Println(validation.Date("dob", "not-a-date", "2006-01-02"))
	// Output:
	// <nil>
	// dob: dob must be a valid date (2006-01-02)
}
