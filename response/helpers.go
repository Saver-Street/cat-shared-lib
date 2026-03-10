// Package response provides HTTP response helpers for JSON APIs.
package response

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
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

// InternalError logs the error and sends a 500 response.
func InternalError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
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
