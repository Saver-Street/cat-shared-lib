package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Saver-Street/cat-shared-lib/discovery"
)

// ServiceDiscoveryCheckerConfig configures health checks across discovered services.
type ServiceDiscoveryCheckerConfig struct {
	// Registry is the service discovery registry to query.
	Registry *discovery.Registry

	// HealthPath is the path appended to each instance's address.
	// Default: "/health".
	HealthPath string

	// Timeout for each individual health check request. Default: 5s.
	Timeout time.Duration

	// Client is the HTTP client. Default: new client with configured timeout.
	Client *http.Client
}

// ServiceDiscoveryChecker creates a Checker that probes the health endpoint of
// every instance of the given service in the registry. It reports "ok" when all
// instances are healthy, and "degraded" with details about failures.
func ServiceDiscoveryChecker(service string, cfg ServiceDiscoveryCheckerConfig) Checker {
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: cfg.Timeout}
	}

	return NewChecker(service, func(ctx context.Context) error {
		instances, err := cfg.Registry.ResolveAll(service)
		if err != nil {
			return fmt.Errorf("resolving %s: %w", service, err)
		}

		if len(instances) == 0 {
			return fmt.Errorf("no instances registered for %s", service)
		}

		var errs []string
		for _, inst := range instances {
			if err := checkInstance(ctx, cfg.Client, inst, cfg.HealthPath, cfg.Timeout); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", inst.ID, err))
				// Mark instance as unhealthy in the registry.
				_ = cfg.Registry.SetStatus(inst.Service, inst.ID, discovery.StatusUnhealthy)
			} else if inst.Status != discovery.StatusHealthy {
				// Instance recovered; mark healthy again.
				_ = cfg.Registry.SetStatus(inst.Service, inst.ID, discovery.StatusHealthy)
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("%d/%d instances unhealthy: %s",
				len(errs), len(instances), strings.Join(errs, "; "))
		}

		return nil
	})
}

func checkInstance(ctx context.Context, client *http.Client, inst discovery.Instance, path string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := strings.TrimRight(inst.Addr, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "cat-shared-lib/service-health-checker")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("degraded (503): %s", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Verify the status field from the JSON response.
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &status); err == nil && status.Status != "ok" {
		return fmt.Errorf("status is %q, want ok", status.Status)
	}

	return nil
}
