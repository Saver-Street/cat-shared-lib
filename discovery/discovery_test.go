package discovery

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() = nil")
	}
	if len(r.Services()) != 0 {
		t.Errorf("Services() = %v, want empty", r.Services())
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusHealthy, "healthy"},
		{StatusUnhealthy, "unhealthy"},
		{StatusDraining, "draining"},
		{Status(99), "unknown(99)"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestInstance_IsHealthy(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusHealthy, true},
		{StatusUnhealthy, false},
		{StatusDraining, false},
	}
	for _, tt := range tests {
		inst := Instance{Status: tt.status}
		if got := inst.IsHealthy(); got != tt.want {
			t.Errorf("Instance{Status: %v}.IsHealthy() = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestRegister_Success(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Instance{
		Service: "billing",
		ID:      "b-1",
		Addr:    "http://billing-1:8080",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	services := r.Services()
	if len(services) != 1 || services[0] != "billing" {
		t.Errorf("Services() = %v, want [billing]", services)
	}
}

func TestRegister_Validation(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		name string
		inst Instance
		want error
	}{
		{"empty service", Instance{ID: "1", Addr: "http://a"}, ErrEmptyService},
		{"empty ID", Instance{Service: "s", Addr: "http://a"}, ErrEmptyInstanceID},
		{"empty addr", Instance{Service: "s", ID: "1"}, ErrEmptyAddr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Register(tt.inst)
			if !errors.Is(err, tt.want) {
				t.Errorf("Register() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestRegister_UpdateExisting(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://old:8080"})
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://new:8080"})

	all, _ := r.ResolveAll("svc")
	if len(all) != 1 {
		t.Fatalf("len(instances) = %d, want 1 (update, not duplicate)", len(all))
	}
	if all[0].Addr != "http://new:8080" {
		t.Errorf("Addr = %q, want http://new:8080", all[0].Addr)
	}
}

func TestDeregister(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a:8080"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b:8080"})

	err := r.Deregister("svc", "1")
	if err != nil {
		t.Fatalf("Deregister() error = %v", err)
	}

	all, _ := r.ResolveAll("svc")
	if len(all) != 1 {
		t.Fatalf("len(instances) = %d, want 1", len(all))
	}
	if all[0].ID != "2" {
		t.Errorf("remaining ID = %q, want 2", all[0].ID)
	}
}

func TestDeregister_LastInstance(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a:8080"})
	_ = r.Deregister("svc", "1")

	if len(r.Services()) != 0 {
		t.Errorf("Services() = %v, want empty after last instance removed", r.Services())
	}
}

func TestDeregister_Errors(t *testing.T) {
	r := NewRegistry()

	if err := r.Deregister("", "1"); !errors.Is(err, ErrEmptyService) {
		t.Errorf("Deregister('', '1') = %v, want %v", err, ErrEmptyService)
	}
	if err := r.Deregister("svc", ""); !errors.Is(err, ErrEmptyInstanceID) {
		t.Errorf("Deregister('svc', '') = %v, want %v", err, ErrEmptyInstanceID)
	}
	if err := r.Deregister("nonexistent", "1"); !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Deregister(nonexistent) = %v, want %v", err, ErrServiceNotFound)
	}

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	if err := r.Deregister("svc", "nonexistent"); !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Deregister(bad id) = %v, want %v", err, ErrServiceNotFound)
	}
}

func TestResolve_RoundRobin(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b"})
	_ = r.Register(Instance{Service: "svc", ID: "3", Addr: "http://c"})

	addrs := make([]string, 6)
	for i := range 6 {
		inst, err := r.Resolve("svc")
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		addrs[i] = inst.Addr
	}

	// Should cycle through all 3 instances twice.
	expected := []string{"http://a", "http://b", "http://c", "http://a", "http://b", "http://c"}
	for i, want := range expected {
		if addrs[i] != want {
			t.Errorf("Resolve()[%d] = %q, want %q", i, addrs[i], want)
		}
	}
}

func TestResolve_SkipsUnhealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b", Status: StatusUnhealthy})
	_ = r.Register(Instance{Service: "svc", ID: "3", Addr: "http://c"})

	// Should only resolve 1 and 3
	for range 4 {
		inst, err := r.Resolve("svc")
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if inst.Addr == "http://b" {
			t.Error("Resolve() returned unhealthy instance")
		}
	}
}

func TestResolve_AllUnhealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusUnhealthy})

	_, err := r.Resolve("svc")
	if !errors.Is(err, ErrNoHealthyInstances) {
		t.Errorf("Resolve() = %v, want %v", err, ErrNoHealthyInstances)
	}
}

func TestResolve_ServiceNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Resolve("nonexistent")
	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Resolve() = %v, want %v", err, ErrServiceNotFound)
	}
}

func TestResolve_EmptyService(t *testing.T) {
	r := NewRegistry()
	_, err := r.Resolve("")
	if !errors.Is(err, ErrEmptyService) {
		t.Errorf("Resolve('') = %v, want %v", err, ErrEmptyService)
	}
}

func TestResolveAll(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b", Status: StatusUnhealthy})

	all, err := r.ResolveAll("svc")
	if err != nil {
		t.Fatalf("ResolveAll() error = %v", err)
	}
	if len(all) != 2 {
		t.Errorf("ResolveAll() len = %d, want 2 (includes unhealthy)", len(all))
	}
}

func TestResolveHealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b", Status: StatusUnhealthy})
	_ = r.Register(Instance{Service: "svc", ID: "3", Addr: "http://c"})

	healthy, err := r.ResolveHealthy("svc")
	if err != nil {
		t.Fatalf("ResolveHealthy() error = %v", err)
	}
	if len(healthy) != 2 {
		t.Errorf("ResolveHealthy() len = %d, want 2", len(healthy))
	}
}

func TestResolveHealthy_NoneHealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusDraining})

	_, err := r.ResolveHealthy("svc")
	if !errors.Is(err, ErrNoHealthyInstances) {
		t.Errorf("ResolveHealthy() = %v, want %v", err, ErrNoHealthyInstances)
	}
}

func TestSetStatus(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})

	err := r.SetStatus("svc", "1", StatusUnhealthy)
	if err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	all, _ := r.ResolveAll("svc")
	if all[0].Status != StatusUnhealthy {
		t.Errorf("Status = %v, want unhealthy", all[0].Status)
	}
}

func TestSetStatus_Callback(t *testing.T) {
	var called bool
	var gotFrom, gotTo Status
	r := NewRegistry(WithOnInstanceStateChange(func(inst Instance, from, to Status) {
		called = true
		gotFrom = from
		gotTo = to
	}))

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.SetStatus("svc", "1", StatusDraining)

	if !called {
		t.Fatal("OnStateChange callback not called")
	}
	if gotFrom != StatusHealthy || gotTo != StatusDraining {
		t.Errorf("callback got %v→%v, want healthy→draining", gotFrom, gotTo)
	}
}

func TestSetStatus_Errors(t *testing.T) {
	r := NewRegistry()

	if err := r.SetStatus("", "1", StatusHealthy); !errors.Is(err, ErrEmptyService) {
		t.Errorf("SetStatus('', ...) = %v, want %v", err, ErrEmptyService)
	}
	if err := r.SetStatus("svc", "", StatusHealthy); !errors.Is(err, ErrEmptyInstanceID) {
		t.Errorf("SetStatus(..., '', ...) = %v, want %v", err, ErrEmptyInstanceID)
	}
	if err := r.SetStatus("nonexistent", "1", StatusHealthy); !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("SetStatus(nonexistent) = %v, want %v", err, ErrServiceNotFound)
	}

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	if err := r.SetStatus("svc", "bad-id", StatusHealthy); !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("SetStatus(bad id) = %v, want %v", err, ErrServiceNotFound)
	}
}

func TestHeartbeat(t *testing.T) {
	now := time.Now()
	r := NewRegistry()
	r.nowFunc = func() time.Time { return now }

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})

	now = now.Add(5 * time.Minute)
	err := r.Heartbeat("svc", "1")
	if err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}

	all, _ := r.ResolveAll("svc")
	if !all[0].LastSeen.Equal(now) {
		t.Errorf("LastSeen = %v, want %v", all[0].LastSeen, now)
	}
}

func TestHeartbeat_Errors(t *testing.T) {
	r := NewRegistry()
	if err := r.Heartbeat("", "1"); !errors.Is(err, ErrEmptyService) {
		t.Errorf("Heartbeat('') = %v, want %v", err, ErrEmptyService)
	}
	if err := r.Heartbeat("svc", ""); !errors.Is(err, ErrEmptyInstanceID) {
		t.Errorf("Heartbeat('svc', '') = %v, want %v", err, ErrEmptyInstanceID)
	}
	if err := r.Heartbeat("svc", "1"); !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("Heartbeat(nonexistent) = %v, want %v", err, ErrServiceNotFound)
	}
}

func TestMarkStale(t *testing.T) {
	now := time.Now()
	r := NewRegistry()
	r.nowFunc = func() time.Time { return now }

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b"})

	// Advance time past TTL
	now = now.Add(10 * time.Minute)
	marked := r.MarkStale(5 * time.Minute)
	if marked != 2 {
		t.Errorf("MarkStale() = %d, want 2", marked)
	}

	_, err := r.Resolve("svc")
	if !errors.Is(err, ErrNoHealthyInstances) {
		t.Errorf("Resolve() = %v, want %v after MarkStale", err, ErrNoHealthyInstances)
	}
}

func TestMarkStale_DoesNotAffectAlreadyUnhealthy(t *testing.T) {
	now := time.Now()
	r := NewRegistry()
	r.nowFunc = func() time.Time { return now }

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusUnhealthy})

	now = now.Add(10 * time.Minute)
	marked := r.MarkStale(5 * time.Minute)
	if marked != 0 {
		t.Errorf("MarkStale() = %d, want 0 (already unhealthy)", marked)
	}
}

func TestRegisterStatic(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterStatic([]Instance{
		{Service: "billing", ID: "b-1", Addr: "http://billing-1:8080"},
		{Service: "billing", ID: "b-2", Addr: "http://billing-2:8080"},
		{Service: "auth", ID: "a-1", Addr: "http://auth-1:8080"},
	})
	if err != nil {
		t.Fatalf("RegisterStatic() error = %v", err)
	}

	services := r.Services()
	if len(services) != 2 {
		t.Errorf("len(Services()) = %d, want 2", len(services))
	}

	billing, _ := r.ResolveAll("billing")
	if len(billing) != 2 {
		t.Errorf("billing instances = %d, want 2", len(billing))
	}
}

func TestRegisterStatic_Error(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterStatic([]Instance{
		{Service: "good", ID: "1", Addr: "http://a"},
		{Service: "", ID: "2", Addr: "http://b"}, // invalid
	})
	if err == nil {
		t.Fatal("RegisterStatic() = nil, want error on invalid instance")
	}
}

func TestMetadata(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{
		Service:  "svc",
		ID:       "1",
		Addr:     "http://a",
		Metadata: map[string]string{"version": "2.0", "region": "us-east"},
	})

	inst, _ := r.Resolve("svc")
	if inst.Metadata["version"] != "2.0" {
		t.Errorf("Metadata[version] = %q, want 2.0", inst.Metadata["version"])
	}
	if inst.Metadata["region"] != "us-east" {
		t.Errorf("Metadata[region] = %q, want us-east", inst.Metadata["region"])
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = r.Register(Instance{
				Service: "svc",
				ID:      string(rune('a' + n)),
				Addr:    "http://addr",
			})
		}(i)
	}

	// Concurrent resolves
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = r.Resolve("svc")
		}()
	}

	wg.Wait()
}

func TestRegister_StatusChangeCallback(t *testing.T) {
	var called bool
	r := NewRegistry(WithOnInstanceStateChange(func(inst Instance, from, to Status) {
		called = true
	}))

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusUnhealthy})

	if !called {
		t.Error("OnStateChange not called when re-registering with different status")
	}
}
