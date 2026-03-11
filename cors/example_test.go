package cors_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Saver-Street/cat-shared-lib/cors"
)

func ExampleMiddleware() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := cors.Middleware(cors.Config{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
	})(handler)

	// Simulate a preflight request.
	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("Access-Control-Allow-Origin"))
	// Output:
	// https://example.com
}
