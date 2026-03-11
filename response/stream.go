package response

import (
	"encoding/json"
	"net/http"
)

// StreamWriter writes a JSON array to an http.ResponseWriter one
// element at a time, flushing after each write. This is useful for
// streaming large result sets without buffering them in memory.
type StreamWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	count   int
	err     error
}

// NewStreamWriter creates a StreamWriter that writes a JSON array to w.
// It immediately writes the opening bracket and sets appropriate headers.
func NewStreamWriter(w http.ResponseWriter) *StreamWriter {
	sw := &StreamWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		sw.flusher = f
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	_, sw.err = w.Write([]byte("["))
	return sw
}

// Write encodes v as JSON and appends it to the array. It flushes
// after each element if the ResponseWriter supports it.
func (sw *StreamWriter) Write(v any) error {
	if sw.err != nil {
		return sw.err
	}
	if sw.count > 0 {
		if _, err := sw.w.Write([]byte(",")); err != nil {
			sw.err = err
			return err
		}
	}
	if err := json.NewEncoder(sw.w).Encode(v); err != nil {
		sw.err = err
		return err
	}
	sw.count++
	if sw.flusher != nil {
		sw.flusher.Flush()
	}
	return nil
}

// Close writes the closing bracket to complete the JSON array.
func (sw *StreamWriter) Close() error {
	if sw.err != nil {
		return sw.err
	}
	_, err := sw.w.Write([]byte("]"))
	sw.err = err
	if sw.flusher != nil {
		sw.flusher.Flush()
	}
	return err
}

// Count returns the number of elements written so far.
func (sw *StreamWriter) Count() int { return sw.count }
