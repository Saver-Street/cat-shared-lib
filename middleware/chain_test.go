package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestChain_Order(t *testing.T) {
	var order []string
	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := Chain(mw("a"), mw("b"), mw("c"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	testkit.AssertEqual(t, order, []string{
		"a-before", "b-before", "c-before",
		"handler",
		"c-after", "b-after", "a-after",
	})
}

func TestChain_Empty(t *testing.T) {
	called := false
	handler := Chain()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	testkit.AssertTrue(t, called)
}

func TestChain_Single(t *testing.T) {
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "applied")
			next.ServeHTTP(w, r)
		})
	}

	handler := Chain(mw)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	testkit.AssertStatus(t, rr, http.StatusOK)
	testkit.AssertHeader(t, rr, "X-Test", "applied")
}

func BenchmarkChain(b *testing.B) {
	noop := func(next http.Handler) http.Handler { return next }
	handler := Chain(noop, noop, noop)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	for b.Loop() {
		handler.ServeHTTP(rr, req)
	}
}
