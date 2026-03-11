package contracts_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/contracts"
)

func ExampleHealthStatus_IsHealthy() {
	healthy := contracts.HealthStatus{State: contracts.HealthStateOK, Service: "api"}
	degraded := contracts.HealthStatus{State: contracts.HealthStateDegraded, Service: "api"}

	fmt.Println(healthy.IsHealthy())
	fmt.Println(degraded.IsHealthy())
	// Output:
	// true
	// false
}

func ExampleNewStandardError() {
	err := contracts.NewStandardError("NOT_FOUND", "resource does not exist")
	fmt.Println(err)
	// Output:
	// NOT_FOUND: resource does not exist
}

func ExampleNewStandardErrorWithDetails() {
	err := contracts.NewStandardErrorWithDetails(
		"VALIDATION_ERROR",
		"invalid input",
		map[string]any{"field": "email", "reason": "required"},
	)
	fmt.Println(err.Code)
	fmt.Println(err.Details["field"])
	// Output:
	// VALIDATION_ERROR
	// email
}

func ExampleStandardError_Error() {
	withMsg := contracts.StandardError{Code: "FORBIDDEN", Message: "access denied"}
	codeOnly := contracts.StandardError{Code: "UNKNOWN"}

	fmt.Println(withMsg.Error())
	fmt.Println(codeOnly.Error())
	// Output:
	// FORBIDDEN: access denied
	// UNKNOWN
}
