package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestMaxBody_Allowed(t *testing.T) {
	handler := MaxBody(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		testkit.RequireNoError(t, err)
		w.Write(body)
	}))

	body := strings.NewReader("hello")
	r := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Body.String(), "hello")
}

func TestMaxBody_Exceeded(t *testing.T) {
	handler := MaxBody(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader("this is way too long for a 5 byte limit")
	r := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	testkit.AssertEqual(t, w.Code, http.StatusRequestEntityTooLarge)
}

func TestMaxBody_NoBody(t *testing.T) {
	handler := MaxBody(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	testkit.AssertEqual(t, w.Code, http.StatusOK)
}
