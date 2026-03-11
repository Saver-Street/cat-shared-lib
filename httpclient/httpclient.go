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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
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

// doAttempt executes a single request attempt, optionally via the circuit breaker.
func (c *Client) doAttempt(ctx context.Context, method, url string, body []byte, attempt int) (*Response, error) {
	do := func() (*Response, error) {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("httpclient: creating request: %w", err)
		}

		// Apply default headers.
		if c.opts.UserAgent != "" {
			req.Header.Set("User-Agent", c.opts.UserAgent)
		}
		for k, v := range c.opts.Headers {
			req.Header.Set(k, v)
		}

		// Apply request hooks.
		for _, hook := range c.opts.RequestHooks {
			if err := hook(req); err != nil {
				return nil, fmt.Errorf("httpclient: request hook: %w", err)
			}
		}

		start := time.Now()
		resp, err := c.http.Do(req)
		duration := time.Since(start)

		if err != nil {
			slog.ErrorContext(ctx, "httpclient: request error",
				"method", method,
				"url", url,
				"attempt", attempt+1,
				"duration", duration.String(),
				"error", err.Error(),
			)
			return nil, fmt.Errorf("httpclient: %s %s: %w", method, url, err)
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		if err != nil {
			return nil, fmt.Errorf("httpclient: reading response body: %w", err)
		}

		slog.DebugContext(ctx, "httpclient: request complete",
			"method", method,
			"url", url,
			"status", resp.StatusCode,
			"attempt", attempt+1,
			"duration", duration.String(),
		)

		// Apply response hooks.
		for _, hook := range c.opts.ResponseHooks {
			if err := hook(resp); err != nil {
				return nil, fmt.Errorf("httpclient: response hook: %w", err)
			}
		}

		result := &Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Header,
			Body:       respBody,
		}

		// Treat server errors as retryable failures.
		if resp.StatusCode >= 500 {
			return result, fmt.Errorf("httpclient: server error: %d", resp.StatusCode)
		}

		return result, nil
	}

	if c.opts.CircuitBreaker != nil {
		var result *Response
		err := c.opts.CircuitBreaker.Execute(func() error {
			var execErr error
			result, execErr = do()
			return execErr
		})
		return result, err
	}

	return do()
}

// backoff returns the back-off duration for the given attempt using exponential
// back-off with full jitter.
func (c *Client) backoff(attempt int) time.Duration {
	base := float64(c.opts.BaseBackoff)
	max := float64(c.opts.MaxBackoff)
	delay := math.Min(base*math.Pow(2, float64(attempt)), max)
	jittered := time.Duration(rand.Float64() * delay) //nolint:gosec
	if jittered < time.Millisecond {
		jittered = time.Millisecond
	}
	return jittered
}

// resolveURL prepends BaseURL to path-only URLs (starting with "/").
func (c *Client) resolveURL(u string) string {
	if c.opts.BaseURL != "" && len(u) > 0 && u[0] == '/' {
		return strings.TrimRight(c.opts.BaseURL, "/") + u
	}
	return u
}

// Get sends a GET request and returns the response.
func (c *Client) Get(ctx context.Context, url string) (*Response, error) {
	return c.Do(ctx, http.MethodGet, url, nil)
}

// Post sends a POST request with the given body.
func (c *Client) Post(ctx context.Context, url string, body io.Reader) (*Response, error) {
	return c.Do(ctx, http.MethodPost, url, body)
}

// Put sends a PUT request with the given body.
func (c *Client) Put(ctx context.Context, url string, body io.Reader) (*Response, error) {
	return c.Do(ctx, http.MethodPut, url, body)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, url string) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, url, nil)
}

// Patch sends a PATCH request with the given body.
func (c *Client) Patch(ctx context.Context, url string, body io.Reader) (*Response, error) {
	return c.Do(ctx, http.MethodPatch, url, body)
}

// Head sends a HEAD request.
func (c *Client) Head(ctx context.Context, url string) (*Response, error) {
	return c.Do(ctx, http.MethodHead, url, nil)
}

// GetJSON sends a GET request and decodes the JSON response into target.
func (c *Client) GetJSON(ctx context.Context, url string, target any) error {
	resp, err := c.Get(ctx, url)
	if err != nil {
		return err
	}
	return decodeResponse(resp, target)
}

// PostJSON sends a POST request with a JSON body and decodes the response into target.
// If target is nil, the response body is discarded.
func (c *Client) PostJSON(ctx context.Context, url string, payload, target any) error {
	body, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	resp, err := c.Do(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	if target != nil {
		return decodeResponse(resp, target)
	}
	return nil
}

// PutJSON sends a PUT request with a JSON body and decodes the response into target.
func (c *Client) PutJSON(ctx context.Context, url string, payload, target any) error {
	body, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	resp, err := c.Do(ctx, http.MethodPut, url, body)
	if err != nil {
		return err
	}
	if target != nil {
		return decodeResponse(resp, target)
	}
	return nil
}

// DeleteJSON sends a DELETE request and decodes the response into target.
func (c *Client) DeleteJSON(ctx context.Context, url string, target any) error {
	resp, err := c.Delete(ctx, url)
	if err != nil {
		return err
	}
	if target != nil {
		return decodeResponse(resp, target)
	}
	return nil
}

// PatchJSON sends a PATCH request with a JSON payload and decodes the response.
func (c *Client) PatchJSON(ctx context.Context, url string, payload, target any) error {
	body, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	resp, err := c.Do(ctx, http.MethodPatch, url, body)
	if err != nil {
		return err
	}
	if target != nil {
		return decodeResponse(resp, target)
	}
	return nil
}

func marshalJSON(v any) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("httpclient: marshaling JSON: %w", err)
	}
	return bytes.NewReader(b), nil
}

func decodeResponse(resp *Response, target any) error {
	if resp.StatusCode >= 400 {
		return fmt.Errorf("httpclient: unexpected status %d: %s", resp.StatusCode, string(resp.Body))
	}
	if err := json.Unmarshal(resp.Body, target); err != nil {
		return fmt.Errorf("httpclient: decoding response JSON: %w", err)
	}
	return nil
}
