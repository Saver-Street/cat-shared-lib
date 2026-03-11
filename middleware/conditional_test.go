package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// testMW is a simple middleware that sets a header to mark it was applied.
func testMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Applied", "true")
		next.ServeHTTP(w, r)
	})
}

var baseHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestIfPathMatch(t *testing.T) {
	mw := IfPath(testMW, "/api/", "/v1/")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") != "true" {
		t.Fatal("middleware should be applied for /api/ prefix")
	}
}

func TestIfPathNoMatch(t *testing.T) {
	mw := IfPath(testMW, "/api/")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") == "true" {
		t.Fatal("middleware should NOT be applied for /health")
	}
}

func TestIfPathSecondPrefix(t *testing.T) {
	mw := IfPath(testMW, "/api/", "/v1/")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/v1/items", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") != "true" {
		t.Fatal("middleware should be applied for /v1/ prefix")
	}
}

func TestExceptPathMatch(t *testing.T) {
	mw := ExceptPath(testMW, "/health", "/ready")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") == "true" {
		t.Fatal("middleware should be skipped for /health")
	}
}

func TestExceptPathNoMatch(t *testing.T) {
	mw := ExceptPath(testMW, "/health")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") != "true" {
		t.Fatal("middleware should be applied for /api/users")
	}
}

func TestExceptPathSecondPrefix(t *testing.T) {
	mw := ExceptPath(testMW, "/health", "/ready")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") == "true" {
		t.Fatal("middleware should be skipped for /ready")
	}
}

func TestIfMethodMatch(t *testing.T) {
	mw := IfMethod(testMW, "POST", "PUT")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") != "true" {
		t.Fatal("middleware should be applied for POST")
	}
}

func TestIfMethodNoMatch(t *testing.T) {
	mw := IfMethod(testMW, "POST")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") == "true" {
		t.Fatal("middleware should NOT be applied for GET")
	}
}

func TestIfMethodCaseInsensitive(t *testing.T) {
	mw := IfMethod(testMW, "post")
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") != "true" {
		t.Fatal("lowercase method should match")
	}
}

func TestIfPredicateTrue(t *testing.T) {
	pred := func(r *http.Request) bool {
		return r.Header.Get("X-Custom") == "yes"
	}
	mw := If(testMW, pred)
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Custom", "yes")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") != "true" {
		t.Fatal("middleware should be applied when predicate is true")
	}
}

func TestIfPredicateFalse(t *testing.T) {
	pred := func(r *http.Request) bool {
		return r.Header.Get("X-Custom") == "yes"
	}
	mw := If(testMW, pred)
	h := mw(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Header().Get("X-Applied") == "true" {
		t.Fatal("middleware should NOT be applied when predicate is false")
	}
}

func BenchmarkIfPath(b *testing.B) {
	mw := IfPath(testMW, "/api/")
	h := mw(baseHandler)
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	for b.Loop() {
		h.ServeHTTP(w, r)
	}
}

func BenchmarkExceptPath(b *testing.B) {
	mw := ExceptPath(testMW, "/health")
	h := mw(baseHandler)
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	for b.Loop() {
		h.ServeHTTP(w, r)
	}
}
