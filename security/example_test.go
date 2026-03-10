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
