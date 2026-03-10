// Package servicetest provides integration test helpers for microservices.
// It includes an HTTP test server with routing and request-recording capabilities,
// a database test helper that mocks the pgx Querier interface, and a fixture
// loader for managing test data.
//
// This package is intended to be imported only in _test.go files or test binaries.
package servicetest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// ---- HTTPTestServer ----

// RouteHandler holds a handler function and a method filter for a route.
type RouteHandler struct {
	method  string
	handler http.HandlerFunc
}

// HTTPTestServer is an HTTP test server with per-route registration, request
// recording, and response stubbing. It wraps httptest.Server and is safe for
// concurrent use.
type HTTPTestServer struct {
	*httptest.Server

	mu       sync.RWMutex
	routes   map[string][]RouteHandler // path → handlers
	fallback http.HandlerFunc

	reqMu    sync.Mutex
	requests []*RecordedRequest
}

// RecordedRequest captures all details of an incoming HTTP request.
type RecordedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
	Query   string
}

// NewHTTPTestServer creates a new HTTPTestServer. Call t.Cleanup(s.Close)
// automatically via the *testing.T registration.
func NewHTTPTestServer(t *testing.T) *HTTPTestServer {
	t.Helper()
	s := &HTTPTestServer{
		routes: make(map[string][]RouteHandler),
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.dispatch))
	t.Cleanup(s.Server.Close)
	return s
}

// Handle registers handler for the given HTTP method and path.
// Use "*" as method to match any method.
func (s *HTTPTestServer) Handle(method, path string, handler http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[path] = append(s.routes[path], RouteHandler{method: method, handler: handler})
}

// HandleJSON registers a handler that serves the given value as JSON with the
// provided status code.
func (s *HTTPTestServer) HandleJSON(method, path string, statusCode int, v any) {
	s.Handle(method, path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(v)
	})
}

// SetFallback sets a handler for requests that do not match any registered route.
// By default, unmatched requests receive 404 Not Found.
func (s *HTTPTestServer) SetFallback(fn http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fallback = fn
}

func (s *HTTPTestServer) dispatch(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	rec := &RecordedRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header.Clone(),
		Body:    body,
		Query:   r.URL.RawQuery,
	}
	s.reqMu.Lock()
	s.requests = append(s.requests, rec)
	s.reqMu.Unlock()

	s.mu.RLock()
	handlers, ok := s.routes[r.URL.Path]
	fallback := s.fallback
	s.mu.RUnlock()

	if ok {
		for _, rh := range handlers {
			if rh.method == "*" || rh.method == r.Method {
				r.Body = io.NopCloser(bytes.NewReader(body))
				rh.handler(w, r)
				return
			}
		}
		// Path matched but no method matched.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if fallback != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		fallback(w, r)
		return
	}

	http.NotFound(w, r)
}

// Requests returns a copy of all recorded requests.
func (s *HTTPTestServer) Requests() []*RecordedRequest {
	s.reqMu.Lock()
	defer s.reqMu.Unlock()
	out := make([]*RecordedRequest, len(s.requests))
	copy(out, s.requests)
	return out
}

// RequestCount returns the total number of requests received.
func (s *HTTPTestServer) RequestCount() int {
	s.reqMu.Lock()
	defer s.reqMu.Unlock()
	return len(s.requests)
}

// LastRequest returns the most recently recorded request, or nil.
func (s *HTTPTestServer) LastRequest() *RecordedRequest {
	s.reqMu.Lock()
	defer s.reqMu.Unlock()
	if len(s.requests) == 0 {
		return nil
	}
	return s.requests[len(s.requests)-1]
}

// Reset clears all recorded requests and registered routes.
func (s *HTTPTestServer) Reset() {
	s.mu.Lock()
	s.routes = make(map[string][]RouteHandler)
	s.fallback = nil
	s.mu.Unlock()

	s.reqMu.Lock()
	s.requests = nil
	s.reqMu.Unlock()
}

// ---- DBTestHelper ----

// MockRow is a pgx.Row implementation that returns pre-configured values.
// Assign ScanValues to the slice of values that Scan should populate.
// Set Err to simulate a scan error.
type MockRow struct {
	ScanValues []any
	Err        error
}

// Scan implements pgx.Row.
func (r *MockRow) Scan(dest ...any) error {
	if r.Err != nil {
		return r.Err
	}
	for i, d := range dest {
		if i >= len(r.ScanValues) {
			return fmt.Errorf("servicetest: MockRow: not enough values: have %d, want %d", len(r.ScanValues), len(dest))
		}
		if err := assignValue(d, r.ScanValues[i]); err != nil {
			return err
		}
	}
	return nil
}

// assignValue sets *dest to src using type assertion.
func assignValue(dest, src any) error {
	switch d := dest.(type) {
	case *string:
		v, ok := src.(string)
		if !ok {
			return fmt.Errorf("servicetest: cannot assign %T to *string", src)
		}
		*d = v
	case *int:
		v, ok := src.(int)
		if !ok {
			return fmt.Errorf("servicetest: cannot assign %T to *int", src)
		}
		*d = v
	case *int64:
		switch v := src.(type) {
		case int64:
			*d = v
		case int:
			*d = int64(v)
		default:
			return fmt.Errorf("servicetest: cannot assign %T to *int64", src)
		}
	case *bool:
		v, ok := src.(bool)
		if !ok {
			return fmt.Errorf("servicetest: cannot assign %T to *bool", src)
		}
		*d = v
	default:
		return fmt.Errorf("servicetest: unsupported destination type %T", dest)
	}
	return nil
}

// DBTestHelper is a mock Querier compatible with pgx-based packages.
// It records all queries for later assertion and returns pre-configured rows.
type DBTestHelper struct {
	mu      sync.Mutex
	queries []RecordedQuery
	rows    []*MockRow
}

// RecordedQuery holds a captured database query and its arguments.
type RecordedQuery struct {
	SQL  string
	Args []any
}

// QueueRow enqueues a row to be returned by the next QueryRow call.
func (d *DBTestHelper) QueueRow(row *MockRow) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rows = append(d.rows, row)
}

// QueryRow implements the Querier interface used by identity.LookupCandidateID
// and similar functions.
func (d *DBTestHelper) QueryRow(_ context.Context, sql string, args ...any) interface{ Scan(...any) error } {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.queries = append(d.queries, RecordedQuery{SQL: sql, Args: args})

	if len(d.rows) == 0 {
		return &MockRow{Err: fmt.Errorf("servicetest: no rows queued")}
	}
	row := d.rows[0]
	d.rows = d.rows[1:]
	return row
}

// QueryCount returns the number of queries executed.
func (d *DBTestHelper) QueryCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.queries)
}

// Queries returns a copy of all recorded queries.
func (d *DBTestHelper) Queries() []RecordedQuery {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]RecordedQuery, len(d.queries))
	copy(out, d.queries)
	return out
}

// Reset clears all recorded queries and queued rows.
func (d *DBTestHelper) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queries = nil
	d.rows = nil
}

// ---- Fixtures ----

// Fixtures manages loading test fixture data from JSON bytes or pre-registered
// payloads. Use Register to seed fixtures and Load to retrieve them.
type Fixtures struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewFixtures creates an empty Fixtures store.
func NewFixtures() *Fixtures {
	return &Fixtures{data: make(map[string][]byte)}
}

// Register adds a fixture with the given name and raw JSON bytes.
func (f *Fixtures) Register(name string, raw []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[name] = raw
}

// RegisterJSON marshals v to JSON and stores it as a named fixture.
func (f *Fixtures) RegisterJSON(name string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("servicetest: fixtures: marshal %q: %w", name, err)
	}
	f.Register(name, b)
	return nil
}

// Load retrieves the raw bytes of a named fixture.
// Returns an error if the fixture is not found.
func (f *Fixtures) Load(name string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b, ok := f.data[name]
	if !ok {
		return nil, fmt.Errorf("servicetest: fixture %q not found", name)
	}
	return b, nil
}

// LoadInto unmarshals a named fixture into dest.
func (f *Fixtures) LoadInto(name string, dest any) error {
	b, err := f.Load(name)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("servicetest: fixture %q: unmarshal: %w", name, err)
	}
	return nil
}

// MustLoad retrieves the raw bytes of a named fixture and panics if not found.
func (f *Fixtures) MustLoad(name string) []byte {
	b, err := f.Load(name)
	if err != nil {
		panic(err)
	}
	return b
}

// Names returns all registered fixture names.
func (f *Fixtures) Names() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	names := make([]string, 0, len(f.data))
	for k := range f.data {
		names = append(names, k)
	}
	return names
}
