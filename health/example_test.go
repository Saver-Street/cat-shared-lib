package health_test

import (
	"context"
	"fmt"
	"net/http/httptest"

	"github.com/Saver-Street/cat-shared-lib/health"
)

func ExampleHandler() {
	h := health.Handler("my-service", "v1.0.0")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))
	fmt.Println(rr.Code)
	// Output:
	// 200
}

func ExampleNewChecker() {
	c := health.NewChecker("db", func(ctx context.Context) error {
		return nil // simulate healthy DB
	})
	fmt.Println(c.Name())
	fmt.Println(c.Check(context.Background()))
	// Output:
	// db
	// <nil>
}

func ExampleStatus_IsHealthy() {
	s := health.Status{Status: "ok"}
	fmt.Println(s.IsHealthy())
	s.Status = "degraded"
	fmt.Println(s.IsHealthy())
	// Output:
	// true
	// false
}

func ExampleStatus_HasErrors() {
	s := health.Status{Status: "ok"}
	fmt.Println(s.HasErrors())
	s.Status = "degraded"
	fmt.Println(s.HasErrors())
	// Output:
	// false
	// true
}
