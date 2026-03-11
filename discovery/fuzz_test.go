package discovery

import (
	"testing"
)

// FuzzRegister exercises the registry with arbitrary service/ID/addr strings.
// It verifies that Register never panics and returns the expected sentinel
// errors for invalid inputs.
func FuzzRegister(f *testing.F) {
	f.Add("api-gateway", "inst-1", "localhost:8080")
	f.Add("", "inst-1", "localhost:8080")
	f.Add("api", "", "localhost:8080")
	f.Add("api", "inst-1", "")
	f.Add("", "", "")
	f.Add("svc", "id", "https://example.com:443/path")
	f.Add("svc-name-with-dashes", "id_with_underscores", "192.168.1.1:9090")
	f.Add("svc", "id", "\x00\xff\n\t")

	f.Fuzz(func(t *testing.T, service, id, addr string) {
		r := NewRegistry()
		err := r.Register(Instance{
			Service: service,
			ID:      id,
			Addr:    addr,
			Status:  StatusHealthy,
		})

		if service == "" && err != ErrEmptyService {
			t.Errorf("empty service: got %v, want ErrEmptyService", err)
		}
		if service != "" && id == "" && err != ErrEmptyInstanceID {
			t.Errorf("empty id: got %v, want ErrEmptyInstanceID", err)
		}
		if service != "" && id != "" && addr == "" && err != ErrEmptyAddr {
			t.Errorf("empty addr: got %v, want ErrEmptyAddr", err)
		}
	})
}

// FuzzResolve exercises the registry resolution with arbitrary service names
// after registering a known instance.
func FuzzResolve(f *testing.F) {
	f.Add("api-gateway")
	f.Add("")
	f.Add("nonexistent")
	f.Add("\x00\xff\n\t")

	f.Fuzz(func(t *testing.T, service string) {
		r := NewRegistry()
		_ = r.Register(Instance{
			Service: "api-gateway",
			ID:      "inst-1",
			Addr:    "localhost:8080",
			Status:  StatusHealthy,
		})

		inst, err := r.Resolve(service)
		if service == "" {
			if err != ErrEmptyService {
				t.Errorf("empty service: got %v, want ErrEmptyService", err)
			}
			return
		}
		if service == "api-gateway" {
			if err != nil {
				t.Errorf("known service: unexpected error: %v", err)
			}
			if inst.ID != "inst-1" {
				t.Errorf("ID = %q, want %q", inst.ID, "inst-1")
			}
		}
	})
}

// FuzzSetStatus exercises status transitions with arbitrary service/ID/status
// combinations.
func FuzzSetStatus(f *testing.F) {
	f.Add("api", "inst-1", 0)
	f.Add("api", "inst-1", 1)
	f.Add("api", "inst-1", 2)
	f.Add("", "inst-1", 0)
	f.Add("api", "", 0)
	f.Add("api", "inst-1", 99)

	f.Fuzz(func(t *testing.T, service, id string, status int) {
		r := NewRegistry()
		_ = r.Register(Instance{
			Service: "api",
			ID:      "inst-1",
			Addr:    "localhost:8080",
			Status:  StatusHealthy,
		})

		err := r.SetStatus(service, id, Status(status))
		if service == "" && err != ErrEmptyService {
			t.Errorf("empty service: got %v, want ErrEmptyService", err)
		}
		if service != "" && id == "" && err != ErrEmptyInstanceID {
			t.Errorf("empty id: got %v, want ErrEmptyInstanceID", err)
		}
	})
}

// FuzzHeartbeat exercises the heartbeat path with arbitrary service/ID pairs.
func FuzzHeartbeat(f *testing.F) {
	f.Add("api", "inst-1")
	f.Add("", "inst-1")
	f.Add("api", "")
	f.Add("", "")
	f.Add("nonexistent", "inst-99")

	f.Fuzz(func(t *testing.T, service, id string) {
		r := NewRegistry()
		_ = r.Register(Instance{
			Service: "api",
			ID:      "inst-1",
			Addr:    "localhost:8080",
			Status:  StatusHealthy,
		})

		err := r.Heartbeat(service, id)
		if service == "" && err != ErrEmptyService {
			t.Errorf("empty service: got %v, want ErrEmptyService", err)
		}
		if service != "" && id == "" && err != ErrEmptyInstanceID {
			t.Errorf("empty id: got %v, want ErrEmptyInstanceID", err)
		}
	})
}
