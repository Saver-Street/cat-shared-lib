package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestCompress_GzipAccepted(t *testing.T) {
	handler := Compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Content-Encoding", "gzip")
	testkit.AssertHeader(t, rr, "Vary", "Accept-Encoding")

	// Verify body is valid gzip
	gr, err := gzip.NewReader(rr.Body)
	testkit.RequireNoError(t, err)
	defer gr.Close()
	body, err := io.ReadAll(gr)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, string(body), "hello world")
}

func TestCompress_NoGzip(t *testing.T) {
	handler := Compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, rr.Header().Get("Content-Encoding"), "")
	testkit.AssertEqual(t, rr.Body.String(), "hello world")
}

func TestCompress_PoolReuse(t *testing.T) {
	handler := Compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		testkit.AssertHeader(t, rr, "Content-Encoding", "gzip")
	}
}
