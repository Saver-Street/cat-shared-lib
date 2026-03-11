package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func FuzzNew(f *testing.F) {
	f.Add(int64(5*time.Second), 0, int64(500*time.Millisecond), int64(30*time.Second), "my-agent")
	f.Add(int64(0), 3, int64(0), int64(0), "")
	f.Add(int64(-1), -1, int64(-1), int64(-1), "a]b[c")
	f.Add(int64(time.Hour), 100, int64(time.Minute), int64(time.Hour), "x/y z")

	f.Fuzz(func(t *testing.T, timeout int64, retries int, baseBO, maxBO int64, agent string) {
		// Clamp to avoid slow tests.
		if retries > 3 {
			retries = 3
		}

		c := New(
			WithTimeout(time.Duration(timeout)),
			WithRetries(retries),
			WithBaseBackoff(time.Duration(baseBO)),
			WithMaxBackoff(time.Duration(maxBO)),
			WithUserAgent(agent),
		)
		if c == nil {
			t.Fatal("New returned nil")
		}
	})
}

func FuzzDoRequest(f *testing.F) {
	f.Add("GET", "/test", "", 200)
	f.Add("POST", "/api/v1/users", `{"name":"fuzz"}`, 201)
	f.Add("DELETE", "/resource/123", "", 404)
	f.Add("PATCH", "/update", `bad json`, 500)
	f.Add("", "", "", 0)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	f.Cleanup(srv.Close)

	f.Fuzz(func(t *testing.T, method, path, body string, _ int) {
		c := New(WithTimeout(2 * time.Second))
		url := srv.URL + path

		ctx := context.Background()
		// Must not panic regardless of input.
		_, _ = c.Do(ctx, method, url, nil)
	})
}

func FuzzBackoff(f *testing.F) {
	f.Add(int64(500*time.Millisecond), int64(30*time.Second), 0)
	f.Add(int64(0), int64(0), 0)
	f.Add(int64(time.Nanosecond), int64(time.Hour), 10)
	f.Add(int64(-1), int64(-1), 100)

	f.Fuzz(func(t *testing.T, baseBO, maxBO int64, attempt int) {
		if attempt < 0 {
			attempt = 0
		}
		c := New(
			WithBaseBackoff(time.Duration(baseBO)),
			WithMaxBackoff(time.Duration(maxBO)),
		)
		d := c.backoff(attempt)
		if d < 0 {
			t.Errorf("backoff returned negative: %v", d)
		}
	})
}
