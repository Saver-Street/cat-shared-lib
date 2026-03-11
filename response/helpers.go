// Package response provides HTTP response helpers for JSON APIs.
package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/apperror"
)

// JSON sends a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("response: failed to encode JSON", "error", err)
	}
}

// OK sends a 200 JSON response.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// Created sends a 201 JSON response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// Error sends a JSON error response.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// Accepted sends a 202 JSON response.
func Accepted(w http.ResponseWriter, data any) {
	JSON(w, http.StatusAccepted, data)
}

// NoContent sends a 204 No Content response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 JSON error response.
func BadRequest(w http.ResponseWriter, msg string) {
	Error(w, http.StatusBadRequest, msg)
}

// Unauthorized sends a 401 JSON error response.
func Unauthorized(w http.ResponseWriter, msg string) {
	Error(w, http.StatusUnauthorized, msg)
}

// Forbidden sends a 403 JSON error response.
func Forbidden(w http.ResponseWriter, msg string) {
	Error(w, http.StatusForbidden, msg)
}

// NotFound sends a 404 JSON error response.
func NotFound(w http.ResponseWriter, msg string) {
	Error(w, http.StatusNotFound, msg)
}

// Conflict sends a 409 JSON error response.
func Conflict(w http.ResponseWriter, msg string) {
	Error(w, http.StatusConflict, msg)
}

// UnprocessableEntity sends a 422 JSON error response.
func UnprocessableEntity(w http.ResponseWriter, msg string) {
	Error(w, http.StatusUnprocessableEntity, msg)
}

// TooManyRequests sends a 429 JSON error response.
// Use this when a client exceeds a rate limit or brute-force threshold.
func TooManyRequests(w http.ResponseWriter, msg string) {
	Error(w, http.StatusTooManyRequests, msg)
}

// ServiceUnavailable sends a 503 JSON error response.
// Use this during maintenance windows or when a critical dependency is down.
func ServiceUnavailable(w http.ResponseWriter, msg string) {
	Error(w, http.StatusServiceUnavailable, msg)
}

// MethodNotAllowed sends a 405 JSON error response.
// Use this when the HTTP method is not supported for the requested resource.
func MethodNotAllowed(w http.ResponseWriter, msg string) {
	Error(w, http.StatusMethodNotAllowed, msg)
}

// Gone sends a 410 JSON error response.
// Use this for permanently removed resources that will never return.
func Gone(w http.ResponseWriter, msg string) {
	Error(w, http.StatusGone, msg)
}

// GatewayTimeout sends a 504 JSON error response.
// Use this when an upstream service takes too long to respond.
func GatewayTimeout(w http.ResponseWriter, msg string) {
	Error(w, http.StatusGatewayTimeout, msg)
}

// Redirect sends an HTTP redirect response. Use http.StatusFound (302) for
// temporary redirects or http.StatusMovedPermanently (301) for permanent ones.
func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

// PagedResult is a standard paginated JSON response envelope.
// Data holds the current page of items; Total is the count across all pages.
type PagedResult[T any] struct {
	// Data is the slice of items for the current page.
	Data []T `json:"data"`
	// Total is the count of all matching items across all pages.
	Total int `json:"total"`
	// Page is the 1-based page number of this result set.
	Page int `json:"page"`
	// Limit is the maximum number of items per page.
	Limit int `json:"limit"`
	// HasMore is true when at least one further page exists.
	HasMore bool `json:"has_more"`
}

// Paginated sends a 200 JSON response wrapped in a PagedResult envelope.
// total is the full count of matching items; page and limit reflect the current
// pagination parameters. HasMore is true when further pages exist.
func Paginated[T any](w http.ResponseWriter, data []T, total, page, limit int) {
	hasMore := (page * limit) < total
	JSON(w, http.StatusOK, PagedResult[T]{
		Data:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	})
}

// InternalError logs the error and sends a 500 response.
func InternalError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	Error(w, http.StatusInternalServerError, "Internal server error")
}

// AppError writes an appropriate JSON error response for the given error.
// If err is an *apperror.Error, it uses its HTTPStatus and Message.
// Otherwise it falls back to a generic 500 Internal Server Error.
func AppError(w http.ResponseWriter, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		JSON(w, appErr.HTTPStatus, appErr)
		return
	}
	Error(w, http.StatusInternalServerError, "Internal server error")
}

// DecodeJSON decodes a JSON request body into the given struct.
func DecodeJSON(r *http.Request, v any) error {
	defer func() { _ = r.Body.Close() }()
	limited := io.LimitReader(r.Body, 1<<20) // 1MB limit
	return json.NewDecoder(limited).Decode(v)
}

// DecodeOrFail decodes JSON and writes a 400 error on failure.
// Returns true if decoding succeeded, false if the handler should return.
func DecodeOrFail(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := DecodeJSON(r, v); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}

// Text writes a plain-text response with the given status code.
func Text(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

// HTML writes an HTML response with the given status code.
func HTML(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

// Stream writes an io.Reader to the response with the given content type and
// status code. It is useful for proxying file downloads, serving generated
// content, or streaming large payloads without buffering in memory.
func Stream(w http.ResponseWriter, code int, contentType string, r io.Reader) {
w.Header().Set("Content-Type", contentType)
w.WriteHeader(code)
if _, err := io.Copy(w, r); err != nil {
slog.Error("response: stream copy failed", "error", err)
}
}

// Download writes an io.Reader as a file download response. It sets
// Content-Disposition to "attachment" with the given filename, and
// streams the content using the specified content type.
func Download(w http.ResponseWriter, contentType, filename string, r io.Reader) {
w.Header().Set("Content-Type", contentType)
w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
w.WriteHeader(http.StatusOK)
if _, err := io.Copy(w, r); err != nil {
slog.Error("response: download copy failed", "error", err)
}
}

// SSEvent writes a single Server-Sent Event to w. The event field is optional;
// pass an empty string to omit it. Data is written as-is (no JSON encoding).
// Call Flush() on the ResponseWriter afterward if it implements http.Flusher.
func SSEvent(w http.ResponseWriter, event, data string) {
if event != "" {
fmt.Fprintf(w, "event: %s\n", event)
}
fmt.Fprintf(w, "data: %s\n\n", data)
}

// SSEHeaders sets the standard headers for a Server-Sent Events stream.
func SSEHeaders(w http.ResponseWriter) {
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")
}
