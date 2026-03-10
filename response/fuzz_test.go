package response

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func FuzzDecodeJSON(f *testing.F) {
	f.Add([]byte(`{"name":"test"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`{"a":1,"b":"hello","c":true}`))
	f.Add([]byte(`invalid json`))
	f.Add([]byte(`{"nested":{"deep":{"value":42}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
		var result map[string]any
		// Must not panic regardless of input
		DecodeJSON(r, &result)
	})
}

func FuzzJSON(f *testing.F) {
	f.Add("test message", 200)
	f.Add("", 404)
	f.Add("error occurred", 500)
	f.Add(strings.Repeat("x", 10000), 200)

	f.Fuzz(func(t *testing.T, msg string, status int) {
		if status < 100 || status > 999 {
			return // skip invalid HTTP status codes
		}
		w := httptest.NewRecorder()
		// Must not panic
		JSON(w, status, map[string]string{"message": msg})
		if w.Code != status {
			t.Errorf("status = %d, want %d", w.Code, status)
		}
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
	})
}
