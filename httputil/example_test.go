package httputil_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/Saver-Street/cat-shared-lib/httputil"
)

func ExampleBearerToken() {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer my-token")
	fmt.Println(httputil.BearerToken(r))
	// Output: my-token
}

func ExampleWriteJSON() {
	w := httptest.NewRecorder()
	_ = httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	fmt.Fprint(os.Stdout, w.Body.String())
	// Output: {"status":"ok"}
}

func ExampleIsJSON() {
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "application/json")
	fmt.Println(httputil.IsJSON(r))
	// Output: true
}
