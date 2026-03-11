package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestDetailedLogging_BasicRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	for _, want := range []string{
		"method=GET",
		"path=/api/users",
		"status=200",
		"duration=",
		"response_bytes=0",
		"remote_addr=192.168.1.1:12345",
	} {
		testkit.AssertContains(t, logged, want)
	}
}

func TestDetailedLogging_QueryString(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello&page=2", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	testkit.AssertContains(t, logged, "query=")
	testkit.AssertContains(t, logged, "q=hello&page=2")
}

func TestDetailedLogging_NoQueryString(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertNotContains(t, buf.String(), "query=")
}

func TestDetailedLogging_UserAgent(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertContains(t, buf.String(), "user_agent=TestAgent/1.0")
}

func TestDetailedLogging_ContentType(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertContains(t, buf.String(), "content_type=application/json")
}

func TestDetailedLogging_RequestIDAndUserID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := SetRequestID(req.Context(), "req-abc-123")
	ctx = SetUserID(ctx, "user-42")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	testkit.AssertContains(t, logged, "request_id=req-abc-123")
	testkit.AssertContains(t, logged, "user_id=user-42")
}

func TestDetailedLogging_NilLogger(t *testing.T) {
	handler := DetailedLogging(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusOK)
}

func TestDetailedLogging_4xxWarnLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	testkit.AssertContains(t, logged, "status=404")
	testkit.AssertContains(t, logged, "level=WARN")
}

func TestDetailedLogging_5xxErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	testkit.AssertContains(t, logged, "status=500")
	testkit.AssertContains(t, logged, "level=ERROR")
}

func TestDetailedLogging_ResponseBodySize(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertContains(t, buf.String(), "response_bytes=11")
}

func TestDetailedLogging_DefaultStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := DetailedLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logged := buf.String()
	testkit.AssertContains(t, logged, "status=200")
	testkit.AssertContains(t, logged, "level=INFO")
}

func TestResponseCapture_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rc := &responseCapture{ResponseWriter: rr, code: http.StatusOK}
	rc.WriteHeader(http.StatusCreated)

	testkit.AssertEqual(t, rc.code, http.StatusCreated)
	testkit.AssertEqual(t, rr.Code, http.StatusCreated)
}

func TestResponseCapture_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	rc := &responseCapture{ResponseWriter: rr, code: http.StatusOK}
	n, err := rc.Write([]byte("test"))

	testkit.AssertEqual(t, n, 4)
	testkit.AssertEqual(t, rc.size, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
