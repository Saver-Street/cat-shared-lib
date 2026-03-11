package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestBasicAuth_ValidCredentials(t *testing.T) {
	handler := BasicAuth("admin", "secret", "test")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("admin", "secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertStatus(t, w, http.StatusOK)
}

func TestBasicAuth_InvalidCredentials(t *testing.T) {
	handler := BasicAuth("admin", "secret", "test")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("admin", "wrong")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertStatus(t, w, http.StatusUnauthorized)
	testkit.AssertHeader(t, w, "WWW-Authenticate", `Basic realm="test"`)
}

func TestBasicAuth_MissingCredentials(t *testing.T) {
	handler := BasicAuth("admin", "secret", "test")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertStatus(t, w, http.StatusUnauthorized)
	testkit.AssertHeader(t, w, "WWW-Authenticate", `Basic realm="test"`)
}

func TestBasicAuth_WrongUsername(t *testing.T) {
	handler := BasicAuth("admin", "secret", "test")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("hacker", "secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertStatus(t, w, http.StatusUnauthorized)
}
