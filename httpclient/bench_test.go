package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGet_Success(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithRetries(0))
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		c.Get(ctx, srv.URL)
	}
}

func BenchmarkGetJSON_Success(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name":"test","value":42}`))
	}))
	defer srv.Close()

	c := New(WithRetries(0))
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		var result struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		c.GetJSON(ctx, srv.URL, &result)
	}
}

func BenchmarkPostJSON_Success(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer srv.Close()

	c := New(WithRetries(0))
	ctx := context.Background()
	payload := map[string]string{"name": "test"}
	b.ResetTimer()
	for b.Loop() {
		var result struct{ ID string }
		c.PostJSON(ctx, srv.URL, payload, &result)
	}
}

func BenchmarkBackoff(b *testing.B) {
	c := New()
	for b.Loop() {
		c.backoff(3)
	}
}

func BenchmarkGet_Parallel(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(WithRetries(0))
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get(ctx, srv.URL)
		}
	})
}
