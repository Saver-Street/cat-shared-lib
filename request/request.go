// Package request provides HTTP request parsing helpers shared across microservices.
package request

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Saver-Street/cat-shared-lib/types"
)

// Pagination holds parsed page/limit/offset values from query parameters.
// For domain-layer pagination, see types.PaginationParams and types.NormalizePage.
type Pagination struct {
	// Page is the 1-based page number (minimum 1).
	Page int
	// Limit is the maximum number of rows to return per page.
	Limit int
	// Offset is the number of rows to skip, derived as (Page-1)*Limit.
	Offset int
}

// ParsePagination extracts page and limit from URL query parameters with defaults and bounds.
// Zero or negative defaultLimit defaults to 20; zero or negative maxLimit defaults to 100.
func ParsePagination(q url.Values, defaultLimit, maxLimit int) Pagination {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}
	page := 1
	if p := q.Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := defaultLimit
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

// URLParamFunc is a function that extracts a named URL parameter from a request.
// This allows callers to plug in any router (chi, gorilla/mux, etc.)
type URLParamFunc func(r *http.Request, key string) string

// RequireURLParam extracts a URL parameter using the provided paramFn.
// Returns an error if the parameter is empty or missing.
func RequireURLParam(r *http.Request, key string, paramFn URLParamFunc) (string, error) {
	val := paramFn(r, key)
	if val == "" {
		return "", fmt.Errorf("missing required URL parameter: %s", key)
	}
	return val, nil
}

// RequireURLParamInt extracts a URL parameter as an int64.
// Returns an error if the parameter is empty, missing, not a valid integer, or not positive (> 0).
// Use RequireQueryParamInt for parameters that may be zero or negative.
func RequireURLParamInt(r *http.Request, key string, paramFn URLParamFunc) (int64, error) {
	val, err := RequireURLParam(r, key, paramFn)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("URL parameter %q must be an integer, got %q", key, val)
	}
	if n <= 0 {
		return 0, fmt.Errorf("URL parameter %q must be positive, got %d", key, n)
	}
	return n, nil
}

// uuidRe matches standard UUID v1–v5 (8-4-4-4-12 hex digits).
var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// RequireUUIDParam extracts a URL parameter and validates it as a UUID.
// Returns an error if the parameter is empty, missing, or not a valid UUID.
func RequireUUIDParam(r *http.Request, key string, paramFn URLParamFunc) (string, error) {
	val, err := RequireURLParam(r, key, paramFn)
	if err != nil {
		return "", err
	}
	if !uuidRe.MatchString(val) {
		return "", fmt.Errorf("URL parameter %q must be a valid UUID, got %q", key, val)
	}
	return val, nil
}

// RequireQueryParam returns the query parameter value for key.
// Returns an error if the parameter is absent or empty.
func RequireQueryParam(q url.Values, key string) (string, error) {
	val := q.Get(key)
	if val == "" {
		return "", fmt.Errorf("missing required query parameter: %s", key)
	}
	return val, nil
}

// ParseBoolParam returns the boolean value of a query parameter.
// Accepts "true"/"1"/"yes" as true; "false"/"0"/"no" as false (case-insensitive).
// Returns defaultValue if the parameter is absent, empty, or unrecognised.
func ParseBoolParam(q url.Values, key string, defaultValue bool) bool {
	raw := strings.TrimSpace(strings.ToLower(q.Get(key)))
	switch raw {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

// OptionalQueryParam returns the query parameter value for key, or defaultValue if absent or empty.
func OptionalQueryParam(q url.Values, key string, defaultValue string) string {
	if val := q.Get(key); val != "" {
		return val
	}
	return defaultValue
}

// OptionalQueryInt returns the query parameter parsed as int64, or defaultValue if absent or invalid.
func OptionalQueryInt(q url.Values, key string, defaultValue int64) int64 {
	raw := q.Get(key)
	if raw == "" {
		return defaultValue
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return defaultValue
	}
	return n
}

// RequireQueryParamInt returns the query parameter parsed as int64.
// Returns an error if the parameter is absent, empty, or not a valid integer.
// Unlike RequireURLParamInt, zero and negative values are accepted.
func RequireQueryParamInt(q url.Values, key string) (int64, error) {
	val, err := RequireQueryParam(q, key)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("query parameter %q must be an integer, got %q", key, val)
	}
	return n, nil
}

// ParseCommaSeparated splits a comma-separated query parameter into a trimmed string slice.
// Returns nil if the key is absent or all tokens are blank after trimming.
func ParseCommaSeparated(q url.Values, key string) []string {
	raw := q.Get(key)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ParseCommaSeparatedInts splits a comma-separated query parameter and parses each token
// as an int64. Returns nil, nil if the key is absent. Returns an error if any token is
// not a valid integer.
func ParseCommaSeparatedInts(q url.Values, key string) ([]int64, error) {
	tokens := ParseCommaSeparated(q, key)
	if tokens == nil {
		return nil, nil
	}
	out := make([]int64, len(tokens))
	for i, t := range tokens {
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("query parameter %q: token %q is not a valid integer", key, t)
		}
		out[i] = n
	}
	return out, nil
}

// ParseSortOrder parses the "sort" and "order" query parameters, validating "sort" against
// allowed field names and "order" against "asc"/"desc". Matching is case-insensitive.
// If the "sort" value is absent or not in allowed, defaultField is returned.
// If the "order" value is absent or not "asc"/"desc", defaultDir is returned.
func ParseSortOrder(q url.Values, allowed []string, defaultField, defaultDir string) (field, dir string) {
	raw := strings.TrimSpace(q.Get("sort"))
	field = defaultField
	for _, a := range allowed {
		if strings.EqualFold(raw, a) {
			field = a
			break
		}
	}
	switch strings.ToLower(strings.TrimSpace(q.Get("order"))) {
	case "asc", "desc":
		dir = strings.ToLower(strings.TrimSpace(q.Get("order")))
	default:
		dir = defaultDir
	}
	return field, dir
}

// ParseDateParam parses a date query parameter. It accepts RFC 3339
// ("2024-01-15T10:30:00Z") and date-only ("2024-01-15") formats.
// Returns the zero time and nil if the parameter is absent or empty.
// Returns an error if the value cannot be parsed.
func ParseDateParam(q url.Values, key string) (time.Time, error) {
	raw := strings.TrimSpace(q.Get(key))
	if raw == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.DateOnly, raw); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("query parameter %q: invalid date format %q (expected RFC 3339 or YYYY-MM-DD)", key, raw)
}

// ParseEnumParam reads a query parameter and validates it against a set of
// allowed values (case-insensitive). Returns defaultValue if the parameter is
// absent, empty, or not in the allowed set.
func ParseEnumParam(q url.Values, key string, allowed []string, defaultValue string) string {
	raw := strings.TrimSpace(q.Get(key))
	if raw == "" {
		return defaultValue
	}
	for _, a := range allowed {
		if strings.EqualFold(raw, a) {
			return a
		}
	}
	return defaultValue
}

// ExtractBearerToken extracts the token from an Authorization header with
// the "Bearer" scheme. It returns the token and true if present, or an
// empty string and false otherwise.
func ExtractBearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(auth) > len(prefix) && strings.EqualFold(auth[:len(prefix)], prefix) {
		return auth[len(prefix):], true
	}
	return "", false
}

// ContentType returns the media type from the request's Content-Type header,
// stripping any parameters (charset, boundary, etc.). Returns an empty string
// if the header is missing or malformed.
func ContentType(r *http.Request) string {
ct := r.Header.Get("Content-Type")
if ct == "" {
return ""
}
mediaType, _, _ := mime.ParseMediaType(ct)
return mediaType
}

// IsJSON returns true if the request's Content-Type is application/json.
func IsJSON(r *http.Request) bool {
return ContentType(r) == "application/json"
}

// ClientIP extracts the client IP address from the request. It checks the
// X-Forwarded-For and X-Real-IP headers first (using the left-most entry
// in X-Forwarded-For), then falls back to the remote address from the
// connection. The returned value has any port stripped.
func ClientIP(r *http.Request) string {
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
// X-Forwarded-For may contain a comma-separated list; take the first.
if i := strings.IndexByte(xff, ','); i > 0 {
return strings.TrimSpace(xff[:i])
}
return strings.TrimSpace(xff)
}
if xri := r.Header.Get("X-Real-IP"); xri != "" {
return strings.TrimSpace(xri)
}
// Strip port from RemoteAddr (e.g. "192.168.1.1:12345").
addr := r.RemoteAddr
if i := strings.LastIndex(addr, ":"); i > 0 {
return addr[:i]
}
return addr
}

// ParseJSONBody reads the request body (limited to 1 MB) and decodes it
// into a value of type T. It returns the decoded value and any decoding
// error. The body is closed after reading.
func ParseJSONBody[T any](r *http.Request) (T, error) {
var v T
defer func() { _ = r.Body.Close() }()
limited := io.LimitReader(r.Body, 1<<20) // 1 MB
if err := json.NewDecoder(limited).Decode(&v); err != nil {
return v, fmt.Errorf("request: invalid JSON body: %w", err)
}
return v, nil
}

// OptionalQueryFloat reads a float64 query parameter, returning defaultValue
// if the parameter is missing or empty. Non-numeric values are silently
// treated as the default.
func OptionalQueryFloat(q url.Values, key string, defaultValue float64) float64 {
val := strings.TrimSpace(q.Get(key))
if val == "" {
return defaultValue
}
f, err := strconv.ParseFloat(val, 64)
if err != nil {
return defaultValue
}
return f
}

// OptionalQueryBool reads a boolean query parameter, returning nil when the
// parameter is absent or empty. Accepted true values: "true", "1", "yes";
// false values: "false", "0", "no". Unrecognised values return nil.
func OptionalQueryBool(q url.Values, key string) *bool {
val := strings.TrimSpace(strings.ToLower(q.Get(key)))
if val == "" {
return nil
}
switch val {
case "true", "1", "yes":
b := true
return &b
case "false", "0", "no":
b := false
return &b
default:
return nil
}
}

// RequireHeader extracts a required HTTP header value. It returns an error
// if the header is missing or empty.
func RequireHeader(r *http.Request, name string) (string, error) {
v := strings.TrimSpace(r.Header.Get(name))
if v == "" {
return "", fmt.Errorf("missing required header: %s", name)
}
return v, nil
}

// DateRange represents a start and end time window parsed from query params.
type DateRange struct {
Start time.Time
End   time.Time
}

// ParseDateRange parses start and end date query parameters. Both are
// optional — a zero-value time.Time indicates the parameter was absent.
// Returns an error if start is after end.
func ParseDateRange(q url.Values, startKey, endKey string) (DateRange, error) {
start, err := ParseDateParam(q, startKey)
if err != nil {
return DateRange{}, err
}
end, err := ParseDateParam(q, endKey)
if err != nil {
return DateRange{}, err
}
if !start.IsZero() && !end.IsZero() && start.After(end) {
return DateRange{}, fmt.Errorf("%s must be before %s", startKey, endKey)
}
return DateRange{Start: start, End: end}, nil
}

// ParseIDList parses a comma-separated list of UUIDs from the given query
// parameter. Returns nil if the key is absent. Returns an error if any
// token is not a valid UUID.
func ParseIDList(q url.Values, key string) ([]string, error) {
tokens := ParseCommaSeparated(q, key)
if tokens == nil {
return nil, nil
}
for _, t := range tokens {
if !uuidRe.MatchString(t) {
return nil, fmt.Errorf("query parameter %q: %q is not a valid UUID", key, t)
}
}
return tokens, nil
}

// ParseCursor extracts cursor and limit from URL query parameters.
// The cursor parameter name is "cursor" and limit is "limit".
func ParseCursor(q url.Values, defaultLimit, maxLimit int) types.CursorParams {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}
	cursor := q.Get("cursor")
	limit := defaultLimit
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return types.CursorParams{Cursor: cursor, Limit: limit}
}
