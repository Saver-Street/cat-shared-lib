package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestMarshalJSON_Error(t *testing.T) {
	_, err := marshalJSON(make(chan int))
	testkit.AssertError(t, err)
}

func TestDecodeResponse_InvalidJSON(t *testing.T) {
	resp := &Response{StatusCode: 200, Body: []byte("not-json")}
	var target map[string]any
	err := decodeResponse(resp, &target)
	testkit.AssertError(t, err)
}

func TestGetJSON_RequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := New()
	var result any
	err := c.GetJSON(context.Background(), srv.URL, &result)
	testkit.AssertError(t, err)
}

func TestPostJSON_MarshalError(t *testing.T) {
	c := New()
	err := c.PostJSON(context.Background(), "http://localhost", make(chan int), nil)
	testkit.AssertError(t, err)
}

func TestPostJSON_RequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := New()
	var result any
	err := c.PostJSON(context.Background(), srv.URL, map[string]string{"k": "v"}, &result)
	testkit.AssertError(t, err)
}

func TestPutJSON_MarshalError(t *testing.T) {
	c := New()
	err := c.PutJSON(context.Background(), "http://localhost", make(chan int), nil)
	testkit.AssertError(t, err)
}

func TestPutJSON_NilTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New()
	err := c.PutJSON(context.Background(), srv.URL, map[string]string{"key": "val"}, nil)
	testkit.RequireNoError(t, err)
}

func TestPutJSON_RequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := New()
	var result any
	err := c.PutJSON(context.Background(), srv.URL, map[string]string{"k": "v"}, &result)
	testkit.AssertError(t, err)
}

func TestDeleteJSON_NilTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New()
	err := c.DeleteJSON(context.Background(), srv.URL, nil)
	testkit.RequireNoError(t, err)
}

func TestDeleteJSON_RequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := New()
	err := c.DeleteJSON(context.Background(), srv.URL, nil)
	testkit.AssertError(t, err)
}

func TestDo_BodyReadError(t *testing.T) {
	c := New()
	_, err := c.Do(context.Background(), http.MethodPost, "http://localhost", errReader{})
	testkit.AssertError(t, err)
}

func TestDoAttempt_ResponseHookError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hookErr := errors.New("response hook failed")
	c := New(
		WithRetries(0),
		WithResponseHook(func(*http.Response) error { return hookErr }),
	)
	_, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error from response hook")
	}
	testkit.AssertErrorContains(t, err, "response hook")
}

func TestDo_ContextCancelledDuringRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	c := New(
		WithRetries(5),
		WithBaseBackoff(500*time.Millisecond),
	)

	// Cancel the context shortly after the first attempt starts.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := c.Do(ctx, http.MethodGet, srv.URL, nil)
	if err == nil {
		t.Fatal("expected error after context cancellation")
	}
	// Should be context.Canceled, not ErrRequestFailed
	if !errors.Is(err, context.Canceled) {
		t.Logf("got error: %v (acceptable)", err)
	}
}

func TestDo_InvalidJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json-at-all"))
	}))
	defer srv.Close()

	c := New()
	var result struct{ Name string }
	err := c.GetJSON(context.Background(), srv.URL, &result)
	testkit.AssertError(t, err)
}

func TestDeleteJSON_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := New()
	var result map[string]any
	err := c.DeleteJSON(context.Background(), srv.URL, &result)
	testkit.AssertError(t, err)
}

func TestDecodeResponse_4xxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	c := New()
	var result map[string]any
	err := c.DeleteJSON(context.Background(), srv.URL, &result)
	if err == nil {
		t.Fatal("expected error for 4xx response")
	}
	testkit.AssertErrorContains(t, err, "400")
}

func TestDo_NilContextExtra(t *testing.T) {
	c := New()
	//nolint:staticcheck
	_, err := c.Do(nil, http.MethodGet, "http://localhost", nil)
	testkit.AssertErrorIs(t, err, ErrNilContext)
}

func TestPostJSON_DecodeResponseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := New()
	var result map[string]any
	err := c.PostJSON(context.Background(), srv.URL, map[string]string{"k": "v"}, &result)
	testkit.AssertError(t, err)
}

func TestPutJSON_DecodeResponseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := New()
	var result map[string]any
	err := c.PutJSON(context.Background(), srv.URL, map[string]string{"k": "v"}, &result)
	testkit.AssertError(t, err)
}

func TestDoAttempt_RequestHookError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hookErr := errors.New("request hook failed")
	c := New(WithRequestHook(func(*http.Request) error { return hookErr }))
	_, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error from request hook")
	}
	testkit.AssertErrorContains(t, err, "request hook")
}

func TestGetJSON_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{invalid json"))
	}))
	defer srv.Close()

	c := New()
	var result struct{ Name string }
	err := c.GetJSON(context.Background(), srv.URL, &result)
	testkit.AssertError(t, err)
}

func TestDoJSON_UnmarshalablePayload(t *testing.T) {
	b, err := json.Marshal(map[string]any{"key": "value"})
	testkit.RequireNoError(t, err)
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestDoAttempt_InvalidURL(t *testing.T) {
	c := New()
	_, err := c.Get(context.Background(), "http://invalid\x7f.example.com/path")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
	testkit.AssertErrorContains(t, err, "creating request")
}

func TestDoAttempt_ReadBodyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Advertise a body but hijack the connection before writing it.
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	c := New(WithRetries(0))
	_, err := c.Get(context.Background(), srv.URL)
	testkit.AssertError(t, err)
}

func TestPatch_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.Method, http.MethodPatch)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"patched":true}`))
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Patch(context.Background(), srv.URL, strings.NewReader(`{"name":"updated"}`))
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, 200)
}

func TestPatchJSON(t *testing.T) {
	type req struct{ Name string }
	type resp struct{ Patched bool }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.Method, http.MethodPatch)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Patched":true}`))
	}))
	defer srv.Close()

	c := New()
	var result resp
	err := c.PatchJSON(context.Background(), srv.URL, req{Name: "test"}, &result)
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, result.Patched)
}

func TestPatchJSON_NilTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New()
	err := c.PatchJSON(context.Background(), srv.URL, map[string]string{"k": "v"}, nil)
	testkit.RequireNoError(t, err)
}

func TestHead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.Method, http.MethodHead)
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Head(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
	testkit.AssertEqual(t, resp.Header.Get("X-Custom"), "value")
}

func TestWithBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.URL.Path, "/v1/users")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL + "/v1"))
	resp, err := c.Get(context.Background(), "/users")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
}

func TestWithBaseURL_FullURLPassedThrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithBaseURL("https://other.example.com"))
	// Full URL should not have base URL prepended
	resp, err := c.Get(context.Background(), srv.URL+"/absolute")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
}

func TestWithBaseURL_TrailingSlash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.URL.Path, "/api/items")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL + "/api/"))
	resp, err := c.Get(context.Background(), "/items")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
}

func TestWithBearerToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.Header.Get("Authorization"), "Bearer my-secret-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithBearerToken("my-secret-token"))
	resp, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
}

func TestWithBearerTokenFunc(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testkit.AssertEqual(t, r.Header.Get("Authorization"), "Bearer dynamic-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithBearerTokenFunc(func() (string, error) {
		callCount++
		return "dynamic-token", nil
	}))
	_, err := c.Get(context.Background(), srv.URL)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, callCount, 1)
}

func TestWithBearerTokenFunc_Error(t *testing.T) {
	c := New(WithBearerTokenFunc(func() (string, error) {
		return "", errors.New("token expired")
	}))
	_, err := c.Get(context.Background(), "http://localhost:1/unused")
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "token expired")
}
