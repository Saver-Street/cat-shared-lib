package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JSONLWriter writes newline-delimited JSON (JSON Lines) to an HTTP response.
// Each call to Write serialises one value as a single line followed by a
// newline character. The response is flushed after each write if the
// underlying ResponseWriter supports http.Flusher.
type JSONLWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	started bool
}

// NewJSONLWriter creates a JSONLWriter that writes to w. It sets the
// Content-Type header to application/x-ndjson on first write.
func NewJSONLWriter(w http.ResponseWriter) *JSONLWriter {
	f, _ := w.(http.Flusher)
	return &JSONLWriter{w: w, flusher: f}
}

// Write serialises v as JSON and writes it as a single line.
func (jw *JSONLWriter) Write(v any) error {
	if !jw.started {
		jw.w.Header().Set("Content-Type", "application/x-ndjson")
		jw.w.WriteHeader(http.StatusOK)
		jw.started = true
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("jsonl marshal: %w", err)
	}
	b = append(b, '\n')
	if _, err := jw.w.Write(b); err != nil {
		return fmt.Errorf("jsonl write: %w", err)
	}
	if jw.flusher != nil {
		jw.flusher.Flush()
	}
	return nil
}

// WriteSlice writes each element of items as a separate JSONL line.
func WriteSlice[T any](jw *JSONLWriter, items []T) error {
	for _, item := range items {
		if err := jw.Write(item); err != nil {
			return err
		}
	}
	return nil
}
