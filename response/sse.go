package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter writes Server-Sent Events to an [http.ResponseWriter].
// Create one with [NewSSEWriter] and send events via [SSEWriter.Send],
// [SSEWriter.SendJSON], or [SSEWriter.SendComment].
//
// The writer sets the required headers on the first call and flushes
// after each event.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	started bool
}

// NewSSEWriter creates a new SSE writer that wraps the given
// [http.ResponseWriter].  The underlying writer should support
// [http.Flusher] for real-time delivery.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	f, _ := w.(http.Flusher)
	return &SSEWriter{w: w, flusher: f}
}

func (s *SSEWriter) init() {
	if s.started {
		return
	}
	s.started = true
	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
	s.w.WriteHeader(http.StatusOK)
}

// Send writes a single SSE event.  The event name and id are optional;
// pass empty strings to omit them.
func (s *SSEWriter) Send(event, id, data string) error {
	s.init()
	if id != "" {
		if _, err := fmt.Fprintf(s.w, "id: %s\n", id); err != nil {
			return err
		}
	}
	if event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}

// SendJSON marshals v as JSON and sends it as the data field.
func (s *SSEWriter) SendJSON(event, id string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Send(event, id, string(b))
}

// SendComment writes an SSE comment line (prefixed with ":").
// Comments are typically used as keep-alive pings.
func (s *SSEWriter) SendComment(text string) error {
	s.init()
	if _, err := fmt.Fprintf(s.w, ": %s\n\n", text); err != nil {
		return err
	}
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}

// SendRetry tells the client to set its reconnection interval to ms
// milliseconds.
func (s *SSEWriter) SendRetry(ms int) error {
	s.init()
	if _, err := fmt.Fprintf(s.w, "retry: %d\n\n", ms); err != nil {
		return err
	}
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}
