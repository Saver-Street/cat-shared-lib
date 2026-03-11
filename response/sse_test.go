package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSSEWriter_Send(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	if err := sw.Send("update", "1", "hello"); err != nil {
		t.Fatal(err)
	}

	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
	if rec.Header().Get("Connection") != "keep-alive" {
		t.Fatalf("Connection = %q", rec.Header().Get("Connection"))
	}

	body := rec.Body.String()
	if !strings.Contains(body, "id: 1\n") {
		t.Fatalf("missing id line in %q", body)
	}
	if !strings.Contains(body, "event: update\n") {
		t.Fatalf("missing event line in %q", body)
	}
	if !strings.Contains(body, "data: hello\n\n") {
		t.Fatalf("missing data line in %q", body)
	}
}

func TestSSEWriter_Send_NoEventNoID(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	if err := sw.Send("", "", "plain data"); err != nil {
		t.Fatal(err)
	}

	body := rec.Body.String()
	if strings.Contains(body, "id:") {
		t.Fatal("should not contain id line")
	}
	if strings.Contains(body, "event:") {
		t.Fatal("should not contain event line")
	}
	if !strings.Contains(body, "data: plain data\n\n") {
		t.Fatalf("body = %q", body)
	}
}

func TestSSEWriter_SendJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	payload := map[string]string{"key": "value"}
	if err := sw.SendJSON("msg", "", payload); err != nil {
		t.Fatal(err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `data: {"key":"value"}`) {
		t.Fatalf("body = %q", body)
	}
}

func TestSSEWriter_SendJSON_MarshalError(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	err := sw.SendJSON("msg", "", make(chan int))
	if err == nil {
		t.Fatal("expected error for unmarshallable type")
	}
}

func TestSSEWriter_SendComment(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	if err := sw.SendComment("keepalive"); err != nil {
		t.Fatal(err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, ": keepalive\n\n") {
		t.Fatalf("body = %q", body)
	}
}

func TestSSEWriter_SendRetry(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	if err := sw.SendRetry(5000); err != nil {
		t.Fatal(err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "retry: 5000\n\n") {
		t.Fatalf("body = %q", body)
	}
}

func TestSSEWriter_MultipleEvents(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	sw.Send("a", "1", "first")
	sw.Send("b", "2", "second")

	body := rec.Body.String()
	if strings.Count(body, "data: ") != 2 {
		t.Fatalf("expected 2 data lines, got body = %q", body)
	}
}

func TestSSEWriter_InitOnlyOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)

	sw.Send("", "", "first")
	sw.Send("", "", "second")

	// StatusOK is the default for httptest.NewRecorder so just verify
	// headers are set (init runs only once, second call should not overwrite).
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatal("Content-Type missing")
	}
}

type noFlushWriter struct {
	http.ResponseWriter
}

func TestSSEWriter_NoFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	nf := &noFlushWriter{ResponseWriter: rec}
	sw := NewSSEWriter(nf)

	if err := sw.Send("", "", "data"); err != nil {
		t.Fatal(err)
	}
	if err := sw.SendComment("ping"); err != nil {
		t.Fatal(err)
	}
	if err := sw.SendRetry(1000); err != nil {
		t.Fatal(err)
	}
}

type errWriter struct {
	http.ResponseWriter
}

func (e *errWriter) Write([]byte) (int, error) {
	return 0, http.ErrAbortHandler
}

func TestSSEWriter_Send_WriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	ew := &errWriter{ResponseWriter: rec}
	sw := NewSSEWriter(ew)

	err := sw.Send("event", "1", "data")
	if err == nil {
		t.Fatal("expected write error")
	}
}

func TestSSEWriter_Send_EventWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	// First call initializes; now swap writer to fail on event line.
	sw.init()
	sw.w = &errWriter{ResponseWriter: rec}

	err := sw.Send("event", "", "data")
	if err == nil {
		t.Fatal("expected write error on event line")
	}
}

func TestSSEWriter_Send_IDWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	sw.init()
	sw.w = &errWriter{ResponseWriter: rec}

	err := sw.Send("", "1", "data")
	if err == nil {
		t.Fatal("expected write error on id line")
	}
}

func TestSSEWriter_SendComment_WriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	sw.init()
	sw.w = &errWriter{ResponseWriter: rec}

	err := sw.SendComment("ping")
	if err == nil {
		t.Fatal("expected write error")
	}
}

func TestSSEWriter_SendRetry_WriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	sw.init()
	sw.w = &errWriter{ResponseWriter: rec}

	err := sw.SendRetry(5000)
	if err == nil {
		t.Fatal("expected write error")
	}
}

func TestSSEWriter_Send_DataWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	sw.init()
	sw.w = &errWriter{ResponseWriter: rec}

	err := sw.Send("", "", "data")
	if err == nil {
		t.Fatal("expected write error on data line")
	}
}

func BenchmarkSSEWriter_Send(b *testing.B) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	for b.Loop() {
		sw.Send("event", "1", "benchmark data")
	}
}

func BenchmarkSSEWriter_SendJSON(b *testing.B) {
	rec := httptest.NewRecorder()
	sw := NewSSEWriter(rec)
	payload := map[string]string{"key": "value"}
	for b.Loop() {
		sw.SendJSON("event", "", payload)
	}
}
