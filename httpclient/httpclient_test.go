package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/circuitbreaker"
	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNew_Defaults(t *testing.T) {
	c := New()
	testkit.AssertEqual(t, c.opts.Timeout, 30*time.Second)
	testkit.AssertEqual(t, c.opts.MaxRetries, 0)
	testkit.AssertEqual(t, c.opts.UserAgent, "cat-shared-lib/httpclient")
}

func TestNew_WithOptions(t *testing.T) {
	c := New(
		WithTimeout(5*time.Second),
		WithRetries(3),
		WithBaseBackoff(100*time.Millisecond),
		WithMaxBackoff(10*time.Second),
		WithUserAgent("test-agent"),
		WithHeader("X-Api-Key", "secret"),
	)
	testkit.AssertEqual(t, c.opts.Timeout, 5*time.Second)
	testkit.AssertEqual(t, c.opts.MaxRetries, 3)
	testkit.AssertEqual(t, c.opts.UserAgent, "test-agent")
	testkit.AssertEqual(t, c.opts.Headers["X-Api-Key"], "secret")
}

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		w.Header().Set("X-Test", "value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 200)
	testkit.AssertEqual(t, resp.Header.Get("X-Test"), "value")
	testkit.AssertEqual(t, string(resp.Body), "{\"ok\":true}")
}

func TestPost_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello" {
			t.Errorf("Body = %q, want hello", string(body))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Post(context.Background(), srv.URL, strings.NewReader("hello"))
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 201)
}

func TestPut_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %q, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Put(context.Background(), srv.URL, strings.NewReader("data"))
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 200)
}

func TestDelete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Delete(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 204)
}

func TestRetry_TransientFailure(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := New(
		WithRetries(3),
		WithBaseBackoff(1*time.Millisecond),
		WithMaxBackoff(5*time.Millisecond),
	)

	resp, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 200)
	testkit.AssertEqual(t, int(attempts.Load()), 3)
}

func TestRetry_AllFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(
		WithRetries(2),
		WithBaseBackoff(1*time.Millisecond),
		WithMaxBackoff(2*time.Millisecond),
	)

	_, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("Get() = nil, want error after all retries exhausted")
	}
	testkit.AssertErrorIs(t, err, ErrRequestFailed)
}

func TestDo_NilContext(t *testing.T) {
	c := New()
	//nolint:staticcheck // intentionally passing nil context
	_, err := c.Do(nil, "GET", "http://example.com", nil)
	testkit.AssertErrorIs(t, err, ErrNilContext)
}

func TestDo_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := New(WithRetries(2))
	_, err := c.Get(ctx, srv.URL)
	testkit.AssertError(t, err)
}

func TestHeaders_AppliedToRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "my-agent" {
			t.Errorf("User-Agent = %q, want my-agent", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("Authorization = %q, want Bearer token", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(
		WithUserAgent("my-agent"),
		WithHeader("Authorization", "Bearer token"),
	)
	_, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
}

func TestRequestHook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "from-hook" {
			t.Errorf("X-Custom = %q, want from-hook", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithRequestHook(func(req *http.Request) error {
		req.Header.Set("X-Custom", "from-hook")
		return nil
	}))
	_, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
}

func TestRequestHook_Error(t *testing.T) {
	hookErr := errors.New("hook failed")
	c := New(WithRequestHook(func(req *http.Request) error {
		return hookErr
	}))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	}))
	defer srv.Close()

	_, err := c.Get(context.Background(), srv.URL)
	testkit.AssertErrorContains(t, err, "hook failed")
}

func TestResponseHook(t *testing.T) {
	var hookedStatus int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithResponseHook(func(resp *http.Response) error {
		hookedStatus = resp.StatusCode
		return nil
	}))
	_, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, hookedStatus, 200)
}

func TestGetJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload{Name: "Alice", Age: 30})
	}))
	defer srv.Close()

	c := New()
	var result payload
	err := c.GetJSON(context.Background(), srv.URL, &result)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, result.Name, "Alice")
	testkit.AssertEqual(t, result.Age, 30)
}

func TestPostJSON(t *testing.T) {
	type req struct {
		Value string `json:"value"`
	}
	type resp struct {
		ID string `json:"id"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in req
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Value != "test" {
			t.Errorf("request value = %q, want test", in.Value)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp{ID: "123"})
	}))
	defer srv.Close()

	c := New()
	var result resp
	err := c.PostJSON(context.Background(), srv.URL, req{Value: "test"}, &result)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, result.ID, "123")
}

func TestPostJSON_NilTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New()
	err := c.PostJSON(context.Background(), srv.URL, map[string]string{"key": "val"}, nil)
	testkit.RequireNoError(t, err)
}

func TestPutJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer srv.Close()

	c := New()
	var result map[string]bool
	err := c.PutJSON(context.Background(), srv.URL, map[string]string{"key": "val"}, &result)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, result["updated"])
}

func TestDeleteJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	c := New()
	var result map[string]bool
	err := c.DeleteJSON(context.Background(), srv.URL, &result)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, result["deleted"])
}

func TestGetJSON_ClientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	c := New()
	var result map[string]any
	err := c.GetJSON(context.Background(), srv.URL, &result)
	if err == nil {
		t.Fatal("GetJSON() = nil, want error for 404")
	}
	testkit.AssertErrorContains(t, err, "404")
}

func TestCircuitBreaker_Integration(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cb := circuitbreaker.New("test",
		circuitbreaker.WithFailureThreshold(2),
	)

	c := New(
		WithCircuitBreaker(cb),
		WithRetries(0),
	)

	// First two failures trip the circuit
	_, _ = c.Get(context.Background(), srv.URL)
	_, _ = c.Get(context.Background(), srv.URL)

	// Third call should be rejected by circuit breaker without hitting server
	beforeAttempts := attempts.Load()
	_, err := c.Get(context.Background(), srv.URL)
	testkit.AssertErrorIs(t, err, circuitbreaker.ErrCircuitOpen)
	testkit.AssertEqual(t, attempts.Load(), beforeAttempts)
}

func TestRetry_BodyReplay(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "replay-me" {
			t.Errorf("Body on attempt %d = %q, want replay-me", attempts.Load(), string(body))
		}
		n := attempts.Add(1)
		if n <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(
		WithRetries(2),
		WithBaseBackoff(1*time.Millisecond),
	)

	resp, err := c.Post(context.Background(), srv.URL, strings.NewReader("replay-me"))
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 200)
}

func TestBackoff(t *testing.T) {
	c := New(
		WithBaseBackoff(100*time.Millisecond),
		WithMaxBackoff(1*time.Second),
	)
	// Just verify it doesn't panic and produces reasonable values
	for attempt := 0; attempt < 20; attempt++ {
		d := c.backoff(attempt)
		if d < time.Millisecond {
			t.Errorf("backoff(%d) = %v, want >= 1ms", attempt, d)
		}
		if d > 1*time.Second {
			t.Errorf("backoff(%d) = %v, want <= 1s (maxBackoff)", attempt, d)
		}
	}
}

func TestServerError_NotRetried_WhenNoRetries(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New() // no retries
	_, err := c.Get(context.Background(), srv.URL)
	testkit.AssertError(t, err)
	testkit.AssertEqual(t, attempts.Load(), int32(1))
}

func TestTransport_Custom(t *testing.T) {
	called := false
	c := New(WithTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("custom")),
		}, nil
	})))

	resp, err := c.Get(context.Background(), "http://fake.example")
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, called)
	testkit.AssertEqual(t, string(resp.Body), "custom")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// errReader is an io.Reader that always returns an error.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }
