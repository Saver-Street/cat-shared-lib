package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPCheckerConfig configures an HTTP endpoint health checker.
type HTTPCheckerConfig struct {
	// Name is the check name shown in the health response. Required.
	Name string

	// URL is the health endpoint to probe (e.g., "http://billing:8080/health").
	URL string

	// Timeout for the HTTP request. Default: 5s.
	Timeout time.Duration

	// ExpectedStatus is the HTTP status code that indicates healthy. Default: 200.
	ExpectedStatus int

	// Client is the HTTP client to use. Default: a new client with the configured timeout.
	Client *http.Client
}

// HTTPChecker creates a Checker that probes a remote HTTP health endpoint.
// The endpoint is expected to return the configured status code. If the response
// body contains JSON with a "status" field, it is included in any error message.
func HTTPChecker(cfg HTTPCheckerConfig) Checker {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.ExpectedStatus == 0 {
		cfg.ExpectedStatus = http.StatusOK
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: cfg.Timeout}
	}

	return NewChecker(cfg.Name, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("User-Agent", "cat-shared-lib/health-checker")

		resp, err := cfg.Client.Do(req)
		if err != nil {
			return fmt.Errorf("request to %s failed: %w", cfg.URL, err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != cfg.ExpectedStatus {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return fmt.Errorf("%s returned %d: %s", cfg.URL, resp.StatusCode, string(body))
		}

		return nil
	})
}

// AggregateChecker creates a Checker that probes a remote health endpoint and
// verifies that its aggregate status is "ok". This is useful for checking
// downstream services that themselves run the Handler from this package.
func AggregateChecker(name, url string) Checker {
	client := &http.Client{Timeout: 5 * time.Second}

	return NewChecker(name, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("User-Agent", "cat-shared-lib/health-aggregator")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request to %s failed: %w", url, err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode == http.StatusServiceUnavailable {
			return fmt.Errorf("%s is degraded (503): %s", name, string(body))
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s returned %d: %s", name, resp.StatusCode, string(body))
		}

		// Parse the body to verify status field.
		var status struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(body, &status); err != nil {
			return fmt.Errorf("parsing %s response: %w", name, err)
		}
		if status.Status != "ok" {
			return fmt.Errorf("%s status is %q, want ok", name, status.Status)
		}

		return nil
	})
}
