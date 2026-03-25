// Package httpclient provides a shared HTTP client for service-to-service
// communication. It wraps [net/http.Client] with:
//
//   - Automatic retries with exponential back-off and jitter
//   - Optional circuit breaker integration via [circuitbreaker.Breaker]
//   - Structured logging with [log/slog]
//   - Request/response middleware hooks
//   - JSON convenience methods (GetJSON, PostJSON, PutJSON, DeleteJSON)
//   - Configurable timeouts and request-scoped headers
//
// Usage:
//
//	client := httpclient.New(
//	    httpclient.WithTimeout(10 * time.Second),
//	    httpclient.WithRetries(3),
//	    httpclient.WithCircuitBreaker(cb),
//	)
//
//	var result MyResponse
//	err := client.GetJSON(ctx, "https://api.example.com/data", &result)
package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Saver-Street/cat-shared-lib/circuitbreaker"
)

// maxResponseBody is the maximum response body size we will read (10 MB).
const maxResponseBody = 10 << 20

// Sentinel errors.
var (
	// ErrRequestFailed is returned when all retry attempts are exhausted.
	ErrRequestFailed = errors.New("httpclient: request failed after retries")
	// ErrNilContext is returned when a nil context is passed to a request method.
	ErrNilContext = errors.New("httpclient: nil context")
)

// RequestHook is called before each request attempt. It may modify the request.
type RequestHook func(req *http.Request) error

// ResponseHook is called after each successful (non-nil) response.
type ResponseHook func(resp *http.Response) error

// Options configures the HTTP client.
type Options struct {
	// Timeout for each individual request attempt. Default: 30s.
	Timeout time.Duration

	// MaxRetries is the number of retry attempts on transient errors.
	// Default: 0 (no retries).
	MaxRetries int

	// BaseBackoff is the initial back-off duration between retries. Default: 500ms.
	BaseBackoff time.Duration

	// MaxBackoff caps the exponential back-off. Default: 30s.
	MaxBackoff time.Duration

	// CircuitBreaker wraps each request attempt. If nil, no circuit breaker is used.
	CircuitBreaker *circuitbreaker.Breaker

	// Transport is the underlying HTTP round tripper. Default: http.DefaultTransport.
	Transport http.RoundTripper

	// RequestHooks are called in order before each request attempt.
	RequestHooks []RequestHook

	// ResponseHooks are called in order after each successful response.
	ResponseHooks []ResponseHook

	// Headers are added to every request.
	Headers map[string]string

	// UserAgent is set as the User-Agent header. Default: "cat-shared-lib/httpclient".
	UserAgent string

	// BaseURL is prepended to request URLs that start with "/".
	BaseURL string
}

// Option applies a configuration to the client.
type Option func(*Options)

// WithTimeout sets the per-request timeout.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) { o.Timeout = d }
}

// WithRetries sets the maximum number of retries.
func WithRetries(n int) Option {
	return func(o *Options) { o.MaxRetries = n }
}

// WithBaseBackoff sets the initial retry back-off duration.
func WithBaseBackoff(d time.Duration) Option {
	return func(o *Options) { o.BaseBackoff = d }
}

// WithMaxBackoff caps the exponential back-off.
func WithMaxBackoff(d time.Duration) Option {
	return func(o *Options) { o.MaxBackoff = d }
}

// WithCircuitBreaker wraps requests with the given circuit breaker.
func WithCircuitBreaker(cb *circuitbreaker.Breaker) Option {
	return func(o *Options) { o.CircuitBreaker = cb }
}

// WithTransport sets the underlying HTTP transport.
func WithTransport(t http.RoundTripper) Option {
	return func(o *Options) { o.Transport = t }
}

// WithRequestHook adds a pre-request hook.
func WithRequestHook(fn RequestHook) Option {
	return func(o *Options) { o.RequestHooks = append(o.RequestHooks, fn) }
}

// WithResponseHook adds a post-response hook.
func WithResponseHook(fn ResponseHook) Option {
	return func(o *Options) { o.ResponseHooks = append(o.ResponseHooks, fn) }
}

// WithHeader adds a default header to all requests.
func WithHeader(key, value string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(o *Options) { o.UserAgent = ua }
}

// WithBaseURL sets a base URL that is prepended to all request paths.
// For example, WithBaseURL("https://api.example.com/v1") turns a
// request to "/users" into "https://api.example.com/v1/users".
func WithBaseURL(base string) Option {
	return func(o *Options) { o.BaseURL = base }
}

// WithBearerToken sets a static Bearer token for Authorization headers.
func WithBearerToken(token string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers["Authorization"] = "Bearer " + token
	}
}

// WithBearerTokenFunc sets a dynamic Bearer token provider that is called
// before each request. Useful for tokens that expire and need refreshing.
func WithBearerTokenFunc(fn func() (string, error)) Option {
	return func(o *Options) {
		o.RequestHooks = append(o.RequestHooks, func(req *http.Request) error {
			token, err := fn()
			if err != nil {
				return fmt.Errorf("httpclient: bearer token provider: %w", err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		})
	}
}

// Client is a configured HTTP client with retry and circuit breaker support.
type Client struct {
	http *http.Client
	opts Options
}

// New creates a Client with the given options.
func New(opts ...Option) *Client {
	o := Options{
		Timeout:     30 * time.Second,
		BaseBackoff: 500 * time.Millisecond,
		MaxBackoff:  30 * time.Second,
		UserAgent:   "cat-shared-lib/httpclient",
	}
	for _, fn := range opts {
		fn(&o)
	}

	transport := o.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
			Timeout:   o.Timeout,
		},
		opts: o,
	}
}

// Response wraps an HTTP response with convenience methods.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Do sends an HTTP request with retries and optional circuit breaker protection.
// The request body is buffered so it can be replayed across retries.
// If a BaseURL is configured and url starts with "/", the BaseURL is prepended.
func (c *Client) Do(ctx context.Context, method, url string, body io.Reader) (*Response, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	url = c.resolveURL(url)

	// Buffer the body for retries.
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("httpclient: reading request body: %w", err)
		}
	}

	var lastErr error
	attempts := 1 + c.opts.MaxRetries

	for attempt := range attempts {
		resp, err := c.doAttempt(ctx, method, url, bodyBytes, attempt)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on circuit breaker errors or context cancellation.
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) ||
			errors.Is(err, circuitbreaker.ErrTooManyRequests) ||
			errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		if attempt < attempts-1 {
			backoff := c.backoff(attempt)
			slog.WarnContext(ctx, "httpclient: retrying request",
				"method", method,
				"url", url,
				"attempt", attempt+1,
				"backoff", backoff.String(),
				"error", err.Error(),
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return nil, fmt.Errorf("%w: %s %s: %w", ErrRequestFailed, method, url, lastErr)
}
