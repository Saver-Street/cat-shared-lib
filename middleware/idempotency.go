package middleware

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

// IdempotencyStore caches responses keyed by idempotency tokens.
type IdempotencyStore interface {
	// Get retrieves a cached response.  Returns the status code, headers,
	// body, and true if found; zero values and false otherwise.
	Get(key string) (int, http.Header, []byte, bool)
	// Set stores a response for the given key.
	Set(key string, status int, header http.Header, body []byte)
}

// IdempotencyOption configures the Idempotency middleware.
type IdempotencyOption func(*idempotencyConfig)

type idempotencyConfig struct {
	headerName string
	store      IdempotencyStore
}

// WithIdempotencyHeader sets the HTTP header name used for the idempotency
// key.  Defaults to "Idempotency-Key".
func WithIdempotencyHeader(name string) IdempotencyOption {
	return func(c *idempotencyConfig) { c.headerName = name }
}

// WithIdempotencyStore sets a custom store for caching responses.
// By default an in-memory store with 10-minute TTL is used.
func WithIdempotencyStore(s IdempotencyStore) IdempotencyOption {
	return func(c *idempotencyConfig) { c.store = s }
}

// Idempotency returns middleware that ensures non-idempotent requests
// (POST, PUT, PATCH) with the same idempotency key return the same
// response.  GET, DELETE, HEAD, and OPTIONS requests pass through
// unmodified.
//
// When a request includes the configured idempotency header, the
// middleware checks the store for a cached response.  If found, the
// cached response is replayed.  Otherwise the request is processed
// normally and the response is cached for future identical requests.
func Idempotency(opts ...IdempotencyOption) func(http.Handler) http.Handler {
	cfg := idempotencyConfig{
		headerName: "Idempotency-Key",
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.store == nil {
		cfg.store = NewMemoryIdempotencyStore(10 * time.Minute)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requiresIdempotency(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(cfg.headerName)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if status, header, body, ok := cfg.store.Get(key); ok {
				for k, vals := range header {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.WriteHeader(status)
				_, _ = w.Write(body)
				return
			}

			rec := &idempotencyRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(rec, r)

			cfg.store.Set(key, rec.statusCode, rec.Header().Clone(), rec.body.Bytes())
		})
	}
}

func requiresIdempotency(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	}
	return false
}

type idempotencyRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *idempotencyRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *idempotencyRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// MemoryIdempotencyStore is a simple in-memory implementation of
// [IdempotencyStore] with TTL-based expiration.
type MemoryIdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]*idempEntry
	ttl     time.Duration
}

type idempEntry struct {
	status  int
	header  http.Header
	body    []byte
	created time.Time
}

// NewMemoryIdempotencyStore creates a new in-memory store with the given TTL.
func NewMemoryIdempotencyStore(ttl time.Duration) *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{
		entries: make(map[string]*idempEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached response.
func (s *MemoryIdempotencyStore) Get(key string) (int, http.Header, []byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok {
		return 0, nil, nil, false
	}
	if time.Since(e.created) > s.ttl {
		return 0, nil, nil, false
	}
	return e.status, e.header, e.body, true
}

// Set stores a response.
func (s *MemoryIdempotencyStore) Set(key string, status int, header http.Header, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = &idempEntry{
		status:  status,
		header:  header,
		body:    body,
		created: time.Now(),
	}
}
