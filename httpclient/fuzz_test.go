package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
		testkit.RequireNotNil(t, c)
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

func FuzzWithHeader(f *testing.F) {
	f.Add("X-Custom", "value")
	f.Add("", "")
	f.Add("Authorization", "Bearer token123")
	f.Add("Content-Type", "application/json")
	f.Add("X-Evil\r\nInjection", "value")
	f.Add("Key", "val\x00ue")
	f.Add("X-Unicode-日本語", "テスト値")
	f.Add("Host", "evil.com")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	f.Cleanup(srv.Close)

	f.Fuzz(func(t *testing.T, key, value string) {
		c := New(
			WithTimeout(2*time.Second),
			WithHeader(key, value),
		)
		testkit.RequireNotNil(t, c)
		// Must not panic when making a request with arbitrary headers.
		ctx := context.Background()
		_, _ = c.Get(ctx, srv.URL+"/test")
	})
}

func FuzzURLConstruction(f *testing.F) {
	f.Add("http://localhost", "/api/v1/users")
	f.Add("http://localhost", "")
	f.Add("http://localhost", "/../../../etc/passwd")
	f.Add("http://localhost", "/path?q=<script>alert(1)</script>")
	f.Add("http://localhost", "/path\x00null")
	f.Add("http://localhost", "/unicode/日本語/パス")
	f.Add("http://localhost", "/"+string(make([]byte, 1000)))
	f.Add("://invalid", "/path")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	f.Cleanup(srv.Close)

	f.Fuzz(func(t *testing.T, base, path string) {
		c := New(WithTimeout(2 * time.Second))
		ctx := context.Background()
		// Must not panic regardless of URL shape.
		_, _ = c.Get(ctx, base+path)
	})
}

func FuzzDoMethodVariants(f *testing.F) {
	f.Add("GET")
	f.Add("POST")
	f.Add("PUT")
	f.Add("DELETE")
	f.Add("PATCH")
	f.Add("OPTIONS")
	f.Add("HEAD")
	f.Add("")
	f.Add("INVALID\x00METHOD")
	f.Add("G\r\nET")
	f.Add(string(make([]byte, 500)))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	f.Cleanup(srv.Close)

	f.Fuzz(func(t *testing.T, method string) {
		c := New(WithTimeout(2 * time.Second))
		ctx := context.Background()
		// Must not panic regardless of method string.
		_, _ = c.Do(ctx, method, srv.URL+"/test", nil)
	})
}

func FuzzPostJSONPayload(f *testing.F) {
	f.Add(`{"name":"test"}`, 200)
	f.Add(``, 200)
	f.Add(`invalid json`, 200)
	f.Add(`{"nested":{"deep":{"value":true}}}`, 200)
	f.Add("\x00\x01\x02", 500)
	f.Add(`{"key":"`+string(make([]byte, 10000))+`"}`, 200)

	f.Fuzz(func(t *testing.T, body string, statusCode int) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			sc := statusCode
			if sc < 100 || sc > 599 {
				sc = 200
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(sc)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		defer srv.Close()

		c := New(WithTimeout(2 * time.Second))
		ctx := context.Background()
		var result map[string]any
		// Must not panic regardless of payload.
		_ = c.PostJSON(ctx, srv.URL+"/api", body, &result)
	})
}
