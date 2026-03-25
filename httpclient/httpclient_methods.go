package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

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
