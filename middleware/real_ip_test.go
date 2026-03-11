package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestRealIP_XForwardedFor(t *testing.T) {
	var got string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Real-IP")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	testkit.AssertEqual(t, got, "1.2.3.4")
}

func TestRealIP_XRealIP(t *testing.T) {
	var got string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Real-IP")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "5.6.7.8")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	testkit.AssertEqual(t, got, "5.6.7.8")
}

func TestRealIP_RemoteAddr(t *testing.T) {
	var got string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Real-IP")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "9.8.7.6:12345"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	testkit.AssertEqual(t, got, "9.8.7.6")
}

func TestRealIP_SingleXFF(t *testing.T) {
	var got string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Real-IP")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "11.22.33.44")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	testkit.AssertEqual(t, got, "11.22.33.44")
}
