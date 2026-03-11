package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHTTPChecker_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "cat-shared-lib/health-checker" {
			t.Errorf("User-Agent = %q, want cat-shared-lib/health-checker", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "remote-api",
		URL:  srv.URL,
	})

	testkit.AssertEqual(t, checker.Name(), "remote-api")

	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestHTTPChecker_Unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"degraded"}`))
	}))
	defer srv.Close()

	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "remote-api",
		URL:  srv.URL,
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestHTTPChecker_CustomStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	checker := HTTPChecker(HTTPCheckerConfig{
		Name:           "custom",
		URL:            srv.URL,
		ExpectedStatus: http.StatusNoContent,
	})

	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestHTTPChecker_ConnectionRefused(t *testing.T) {
	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "down-service",
		URL:  "http://127.0.0.1:1", // unlikely to be listening
	})

	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestAggregateChecker_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "cat-shared-lib/health-aggregator" {
			t.Errorf("User-Agent = %q, want cat-shared-lib/health-aggregator", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok", Service: "downstream"})
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	testkit.AssertNoError(t, checker.Check(t.Context()))
}

func TestAggregateChecker_Degraded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(Status{Status: "degraded"})
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestAggregateChecker_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 200 but status field is not "ok"
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestAggregateChecker_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestAggregateChecker_ConnectionRefused(t *testing.T) {
	checker := AggregateChecker("down", "http://127.0.0.1:1")
	testkit.AssertError(t, checker.Check(t.Context()))
}

func TestAggregateChecker_UnexpectedStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "returned 500")
}

func TestHTTPChecker_InHandler(t *testing.T) {
	// Test that HTTPChecker integrates with the existing Handler.
	remoteSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer remoteSrv.Close()

	handler := Handler("gateway", "1.0.0",
		HTTPChecker(HTTPCheckerConfig{Name: "billing", URL: remoteSrv.URL}),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	testkit.AssertEqual(t, rec.Code, http.StatusOK)

	var status Status
	_ = json.NewDecoder(rec.Body).Decode(&status)
	testkit.AssertEqual(t, status.Status, "ok")
	testkit.AssertEqual(t, status.Checks["billing"], "ok")
}

func TestHTTPChecker_DegradedInHandler(t *testing.T) {
	remoteSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer remoteSrv.Close()

	handler := Handler("gateway", "1.0.0",
		HTTPChecker(HTTPCheckerConfig{Name: "billing", URL: remoteSrv.URL}),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	testkit.AssertEqual(t, rec.Code, http.StatusServiceUnavailable)

	var status Status
	_ = json.NewDecoder(rec.Body).Decode(&status)
	testkit.AssertEqual(t, status.Status, "degraded")
}

func TestHTTPChecker_InvalidURL(t *testing.T) {
	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "bad-url",
		URL:  "http://\x7f",
	})

	err := checker.Check(t.Context())
	if err == nil {
		t.Fatal("Check() = nil, want error for invalid URL")
	}
	testkit.AssertErrorContains(t, err, "creating request")
}

func TestAggregateChecker_InvalidURL(t *testing.T) {
	checker := AggregateChecker("bad", "http://\x7f")

	err := checker.Check(t.Context())
	if err == nil {
		t.Fatal("Check() = nil, want error for invalid URL")
	}
	testkit.AssertErrorContains(t, err, "creating request")
}

func TestAggregateChecker_ReadBodyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Advertise a large body then close the connection to trigger read error.
		w.Header().Set("Content-Length", "9999")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
		}
	}))
	defer srv.Close()

	checker := AggregateChecker("bad-body", srv.URL)
	err := checker.Check(t.Context())
	testkit.AssertError(t, err)
}
