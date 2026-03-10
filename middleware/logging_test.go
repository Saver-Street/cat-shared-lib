package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging_BasicRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	for _, want := range []string{"method=GET", "path=/api/users", "status=200", "duration="} {
		if !strings.Contains(logged, want) {
			t.Errorf("log missing %q in: %s", want, logged)
		}
	}
}

func TestLogging_IncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := SetRequestID(req.Context(), "req-abc-123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "req-abc-123") {
		t.Errorf("log missing request_id: %s", buf.String())
	}
}

func TestLogging_IncludesUserID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := SetUserID(req.Context(), "user-42")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "user-42") {
		t.Errorf("log missing user_id: %s", buf.String())
	}
}

func TestLogging_NonOKStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "status=404") {
		t.Errorf("log missing status=404: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "method=POST") {
		t.Errorf("log missing method=POST: %s", buf.String())
	}
}

func TestLogging_NilLogger(t *testing.T) {
	// Should not panic with nil logger
	handler := Logging(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rr.Code)
	}
}

func TestLogging_DefaultStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	// Handler that writes body without explicit WriteHeader
	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "status=200") {
		t.Errorf("expected status=200 default: %s", buf.String())
	}
}

func TestStatusWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rr, code: http.StatusOK}
	sw.WriteHeader(http.StatusCreated)

	if sw.code != http.StatusCreated {
		t.Errorf("got code %d, want 201", sw.code)
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("underlying recorder got %d, want 201", rr.Code)
	}
}

func TestLogging_NoRequestIDOrUserID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/plain", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	if strings.Contains(logged, "request_id") {
		t.Errorf("should not contain request_id when not set: %s", logged)
	}
	if strings.Contains(logged, "user_id") {
		t.Errorf("should not contain user_id when not set: %s", logged)
	}
}

func TestLogging_ContextPropagation(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	type testKey string
	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val, _ := r.Context().Value(testKey("k")).(string)
		if val != "v" {
			t.Errorf("context value lost: got %q", val)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), testKey("k"), "v")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}
