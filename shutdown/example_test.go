package shutdown_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Saver-Street/cat-shared-lib/shutdown"
)

func ExampleDrainer() {
	var d shutdown.Drainer

	// Simulate tracking two in-flight requests.
	d.Add()
	d.Add()
	d.Done()
	d.Done()

	// Wait returns immediately when count reaches zero.
	d.Wait()
	fmt.Println("all drained")
	// Output:
	// all drained
}

func ExampleDrainer_Middleware() {
	var d shutdown.Drainer

	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rr.Code)
	// Output:
	// 200
}
