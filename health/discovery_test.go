package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/discovery"
	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestServiceDiscoveryChecker_AllHealthy(t *testing.T) {
	// Create two healthy service instances.
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv2.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-1", Addr: srv1.URL})
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-2", Addr: srv2.URL})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_OneUnhealthy(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(Status{Status: "degraded"})
	}))
	defer unhealthy.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-1", Addr: healthy.URL})
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-2", Addr: unhealthy.URL})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	err := checker.Check(t.Context())
	if err == nil {
		t.Fatal("Check() = nil, want error when one instance is unhealthy")
	}
	testkit.AssertErrorContains(t, err, "1/2 instances unhealthy")

	// Verify that the unhealthy instance was marked in registry.
	all, _ := reg.ResolveAll("billing")
	for _, inst := range all {
		if inst.ID == "b-2" {
			testkit.AssertEqual(t, inst.Status, discovery.StatusUnhealthy)
		}
	}
}

func TestServiceDiscoveryChecker_InstanceRecovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{
		Service: "billing",
		ID:      "b-1",
		Addr:    srv.URL,
		Status:  discovery.StatusUnhealthy, // was unhealthy
	})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertNoError(t, checker.Check(t.Context()))

	// Verify instance marked healthy again.
	all, _ := reg.ResolveAll("billing")
	testkit.AssertEqual(t, all[0].Status, discovery.StatusHealthy)
}

func TestServiceDiscoveryChecker_NoInstances(t *testing.T) {
	reg := discovery.NewRegistry()

	checker := ServiceDiscoveryChecker("nonexistent", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_InvalidInstanceAddr(t *testing.T) {
	reg := discovery.NewRegistry()
	// Control character in the address makes http.NewRequestWithContext fail.
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-1", Addr: "http://\x7f"})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_ConnectionRefused(t *testing.T) {
	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{
		Service: "billing",
		ID:      "b-1",
		Addr:    "http://127.0.0.1:1", // connection refused
	})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_CustomPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/healthz" {
			t.Errorf("Path = %q, want /api/healthz", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry:   reg,
		HealthPath: "/api/healthz",
	})

	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_UnexpectedStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "billing", ID: "b-1", Addr: srv.URL})

	checker := ServiceDiscoveryChecker("billing", ServiceDiscoveryCheckerConfig{
		Registry: reg,
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestServiceDiscoveryChecker_InHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok"})
	}))
	defer srv.Close()

	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "auth", ID: "a-1", Addr: srv.URL})

	handler := Handler("gateway", "1.0.0",
		ServiceDiscoveryChecker("auth", ServiceDiscoveryCheckerConfig{Registry: reg}),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	testkit.AssertEqual(t, rec.Code, http.StatusOK)

	var status Status
	_ = json.NewDecoder(rec.Body).Decode(&status)
	testkit.AssertEqual(t, status.Checks["auth"], "ok")
}
