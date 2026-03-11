package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/discovery"
	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestServiceDiscoveryChecker_SlowInstanceTimeout(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer slow.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "slow-svc", ID: "s-1", Addr: slow.URL})

	checker := ServiceDiscoveryChecker("slow-svc", ServiceDiscoveryCheckerConfig{
		Registry: reg,
		Timeout:  100 * time.Millisecond,
	})

	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "1/1 instances unhealthy")
}

func TestServiceDiscoveryChecker_AllInstancesUnhealthy(t *testing.T) {
	unhealthy1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(Status{Status: "degraded"})
	}))
	defer unhealthy1.Close()

	unhealthy2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(Status{Status: "degraded"})
	}))
	defer unhealthy2.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "bad-svc", ID: "b-1", Addr: unhealthy1.URL})
	_ = reg.Register(discovery.Instance{Service: "bad-svc", ID: "b-2", Addr: unhealthy2.URL})

	checker := ServiceDiscoveryChecker("bad-svc", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "2/2 instances unhealthy")

	// Verify both marked unhealthy
	all, _ := reg.ResolveAll("bad-svc")
	for _, inst := range all {
		testkit.AssertEqual(t, inst.Status, discovery.StatusUnhealthy)
	}
}

func TestServiceDiscoveryChecker_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "bad-json", ID: "j-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("bad-json", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	// Malformed JSON with 200 OK should still pass since json.Unmarshal
	// failure is silently ignored (the status check only triggers on success)
	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "empty-body", ID: "e-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("empty-body", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	// Empty body with 200 should pass (json unmarshal fails silently)
	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_TrailingSlashInAddr(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("Path = %q, want /health", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "trailing", ID: "t-1", Addr: srv.URL + "/"})

	checker := ServiceDiscoveryChecker("trailing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "cancel", ID: "c-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("cancel", ServiceDiscoveryCheckerConfig{
		Registry: reg,
		Timeout:  100 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := checker.Check(ctx)
	testkit.AssertError(t, err)
}

func TestServiceDiscoveryChecker_DNSFailure(t *testing.T) {
	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{
		Service: "dns-fail",
		ID:      "d-1",
		Addr:    "http://nonexistent.invalid.local",
	})

	checker := ServiceDiscoveryChecker("dns-fail", ServiceDiscoveryCheckerConfig{
		Registry: reg,
		Timeout:  1 * time.Second,
	})

	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "1/1 instances unhealthy")
}

func TestServiceDiscoveryChecker_CustomClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "custom", ID: "c-1", Addr: srv.URL})

	customClient := &http.Client{Timeout: 10 * time.Second}
	checker := ServiceDiscoveryChecker("custom", ServiceDiscoveryCheckerConfig{
		Registry: reg,
		Client:   customClient,
	})

	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_MixedHealthy(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer healthy.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "mixed", ID: "m-1", Addr: healthy.URL})
	_ = reg.Register(discovery.Instance{Service: "mixed", ID: "m-2", Addr: "http://127.0.0.1:1"}) // refused
	_ = reg.Register(discovery.Instance{Service: "mixed", ID: "m-3", Addr: healthy.URL})

	checker := ServiceDiscoveryChecker("mixed", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "1/3 instances unhealthy")
}

func TestServiceDiscoveryChecker_StatusNotOKInJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "warning"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "warning", ID: "w-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("warning", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, `status is "warning"`)
}

func TestServiceDiscoveryChecker_CheckerNameIsService(t *testing.T) {
	reg := discovery.NewRegistry()
	checker := ServiceDiscoveryChecker("my-service", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})
	testkit.AssertEqual(t, checker.Name(), "my-service")
}
