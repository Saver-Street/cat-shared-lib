package security_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/security"
)

func ExampleContainsSuspiciousInput() {
	fmt.Println(security.ContainsSuspiciousInput("normal text"))
	fmt.Println(security.ContainsSuspiciousInput("DROP TABLE users"))
	fmt.Println(security.ContainsSuspiciousInput("<script>alert(1)</script>"))
	// Output:
	// false
	// true
	// true
}

func ExampleRedactPII() {
	data := map[string]any{
		"email":   "user@example.com",
		"name":    "Jane Doe",
		"message": "Call me at 555-123-4567",
	}
	result := security.RedactPII(data)
	fmt.Println(result["email"])
	fmt.Println(result["name"])
	// Output:
	// [REDACTED]
	// Jane Doe
}

func ExampleTruncateForLog() {
	short := security.TruncateForLog("hello", 10)
	fmt.Println(short)

	long := security.TruncateForLog("a very long string", 10)
	fmt.Println(long)
	// Output:
	// hello
	// a very lon
}

func ExampleSanitizeHeader() {
	clean := security.SanitizeHeader("Bearer token\r\nX-Injected: evil")
	fmt.Println(clean)
	// Output: Bearer tokenX-Injected: evil
}

func ExampleIsRelativeURL() {
	fmt.Println(security.IsRelativeURL("/dashboard"))
	fmt.Println(security.IsRelativeURL("//evil.com"))
	fmt.Println(security.IsRelativeURL("https://evil.com"))
	// Output:
	// true
	// false
	// false
}
