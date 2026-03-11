package response

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// CursorPage holds a page of results with opaque cursors for forward/backward
// navigation. It is typically serialised as JSON.
type CursorPage[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// EncodeCursor encodes an arbitrary value into a URL-safe, opaque cursor
// string using base64-encoded JSON.
func EncodeCursor(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor decodes a cursor string produced by EncodeCursor back into
// the target value.
func DecodeCursor(cursor string, target any) error {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return fmt.Errorf("decode cursor: %w", err)
	}
	if err := json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("decode cursor: %w", err)
	}
	return nil
}

// WriteCursorPage serialises a CursorPage as JSON and writes it to the
// response with a 200 OK status.
func WriteCursorPage[T any](w http.ResponseWriter, page CursorPage[T]) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(page)
}
