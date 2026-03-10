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
)

func TestNew_Defaults(t *testing.T) {
	c := New()
	if c.opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", c.opts.Timeout)
	}
	if c.opts.MaxRetries != 0 {
		t.Errorf("MaxRetries = %d, want 0", c.opts.MaxRetries)
	}
	if c.opts.UserAgent != "cat-shared-lib/httpclient" {
		t.Errorf("UserAgent = %q, want default", c.opts.UserAgent)
	}
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
	if c.opts.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", c.opts.Timeout)
	}
	if c.opts.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", c.opts.MaxRetries)
	}
	if c.opts.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %q, want test-agent", c.opts.UserAgent)
	}
	if c.opts.Headers["X-Api-Key"] != "secret" {
		t.Errorf("Headers[X-Api-Key] = %q, want secret", c.opts.Headers["X-Api-Key"])
	}
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
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if resp.Header.Get("X-Test") != "value" {
		t.Errorf("Header X-Test = %q, want value", resp.Header.Get("X-Test"))
	}
	if string(resp.Body) != `{"ok":true}` {
		t.Errorf("Body = %q, want {\"ok\":true}", string(resp.Body))
	}
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
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}
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
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
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
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", resp.StatusCode)
	}
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
	if err != nil {
		t.Fatalf("Get() error = %v, expected success after retries", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if int(attempts.Load()) != 3 {
		t.Errorf("attempts = %d, want 3", attempts.Load())
	}
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
	if !errors.Is(err, ErrRequestFailed) {
		t.Errorf("error = %v, want %v", err, ErrRequestFailed)
	}
}

func TestDo_NilContext(t *testing.T) {
	c := New()
	//nolint:staticcheck // intentionally passing nil context
	_, err := c.Do(nil, "GET", "http://example.com", nil)
	if !errors.Is(err, ErrNilContext) {
		t.Errorf("Do(nil ctx) = %v, want %v", err, ErrNilContext)
	}
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
	if err == nil {
		t.Fatal("Get() = nil, want error on cancelled context")
	}
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
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
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
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
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
	if err == nil || !strings.Contains(err.Error(), "hook failed") {
		t.Errorf("Get() error = %v, want hook error", err)
	}
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
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if hookedStatus != 200 {
		t.Errorf("hooked status = %d, want 200", hookedStatus)
	}
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
	if err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if result.Name != "Alice" || result.Age != 30 {
		t.Errorf("GetJSON() result = %+v, want {Alice 30}", result)
	}
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
	if err != nil {
		t.Fatalf("PostJSON() error = %v", err)
	}
	if result.ID != "123" {
		t.Errorf("PostJSON() id = %q, want 123", result.ID)
	}
}

func TestPostJSON_NilTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New()
	err := c.PostJSON(context.Background(), srv.URL, map[string]string{"key": "val"}, nil)
	if err != nil {
		t.Fatalf("PostJSON() error = %v", err)
	}
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
	if err != nil {
		t.Fatalf("PutJSON() error = %v", err)
	}
	if !result["updated"] {
		t.Errorf("PutJSON() updated = false, want true")
	}
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
	if err != nil {
		t.Fatalf("DeleteJSON() error = %v", err)
	}
	if !result["deleted"] {
		t.Errorf("DeleteJSON() deleted = false, want true")
	}
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
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %v, want to contain 404", err)
	}
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
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("Get() = %v, want %v", err, circuitbreaker.ErrCircuitOpen)
	}
	if attempts.Load() != beforeAttempts {
		t.Error("server was called despite open circuit")
	}
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
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
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
	if err == nil {
		t.Fatal("Get() = nil, want error for 500")
	}
	if attempts.Load() != 1 {
		t.Errorf("attempts = %d, want 1", attempts.Load())
	}
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
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !called {
		t.Error("custom transport was not called")
	}
	if string(resp.Body) != "custom" {
		t.Errorf("Body = %q, want custom", string(resp.Body))
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
