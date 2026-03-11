package discovery

import (
	"sync"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() = nil")
	}
	testkit.AssertLen(t, r.Services(), 0)
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
		testkit.AssertEqual(t, tt.s.String(), tt.want)
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
		testkit.AssertEqual(t, inst.IsHealthy(), tt.want)
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
	if len(services) != 1 {
		t.Fatalf("len(Services()) = %d, want 1", len(services))
	}
	testkit.AssertEqual(t, services[0], "billing")
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
			testkit.AssertErrorIs(t, err, tt.want)
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
	testkit.AssertEqual(t, all[0].Addr, "http://new:8080")
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
	testkit.AssertEqual(t, all[0].ID, "2")
}

func TestDeregister_LastInstance(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a:8080"})
	_ = r.Deregister("svc", "1")

	testkit.AssertLen(t, r.Services(), 0)
}

func TestDeregister_Errors(t *testing.T) {
	r := NewRegistry()

	testkit.AssertErrorIs(t, r.Deregister("", "1"), ErrEmptyService)
	testkit.AssertErrorIs(t, r.Deregister("svc", ""), ErrEmptyInstanceID)
	testkit.AssertErrorIs(t, r.Deregister("nonexistent", "1"), ErrServiceNotFound)

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	testkit.AssertErrorIs(t, r.Deregister("svc", "nonexistent"), ErrServiceNotFound)
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
		testkit.AssertEqual(t, addrs[i], want)
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
		testkit.AssertNotEqual(t, inst.Addr, "http://b")
	}
}

func TestResolve_AllUnhealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusUnhealthy})

	_, err := r.Resolve("svc")
	testkit.AssertErrorIs(t, err, ErrNoHealthyInstances)
}

func TestResolve_ServiceNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Resolve("nonexistent")
	testkit.AssertErrorIs(t, err, ErrServiceNotFound)
}

func TestResolve_EmptyService(t *testing.T) {
	r := NewRegistry()
	_, err := r.Resolve("")
	testkit.AssertErrorIs(t, err, ErrEmptyService)
}

func TestResolveAll(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = r.Register(Instance{Service: "svc", ID: "2", Addr: "http://b", Status: StatusUnhealthy})

	all, err := r.ResolveAll("svc")
	if err != nil {
		t.Fatalf("ResolveAll() error = %v", err)
	}
	testkit.AssertLen(t, all, 2)
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
	testkit.AssertLen(t, healthy, 2)
}

func TestResolveHealthy_NoneHealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusDraining})

	_, err := r.ResolveHealthy("svc")
	testkit.AssertErrorIs(t, err, ErrNoHealthyInstances)
}

func TestSetStatus(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})

	err := r.SetStatus("svc", "1", StatusUnhealthy)
	if err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	all, _ := r.ResolveAll("svc")
	testkit.AssertEqual(t, all[0].Status, StatusUnhealthy)
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
	testkit.AssertEqual(t, gotFrom, StatusHealthy)
	testkit.AssertEqual(t, gotTo, StatusDraining)
}

func TestSetStatus_Errors(t *testing.T) {
	r := NewRegistry()

	testkit.AssertErrorIs(t, r.SetStatus("", "1", StatusHealthy), ErrEmptyService)
	testkit.AssertErrorIs(t, r.SetStatus("svc", "", StatusHealthy), ErrEmptyInstanceID)
	testkit.AssertErrorIs(t, r.SetStatus("nonexistent", "1", StatusHealthy), ErrServiceNotFound)

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	testkit.AssertErrorIs(t, r.SetStatus("svc", "bad-id", StatusHealthy), ErrServiceNotFound)
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
	testkit.AssertTrue(t, all[0].LastSeen.Equal(now))
}

func TestHeartbeat_Errors(t *testing.T) {
	r := NewRegistry()
	testkit.AssertErrorIs(t, r.Heartbeat("", "1"), ErrEmptyService)
	testkit.AssertErrorIs(t, r.Heartbeat("svc", ""), ErrEmptyInstanceID)
	testkit.AssertErrorIs(t, r.Heartbeat("svc", "1"), ErrServiceNotFound)
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
	testkit.AssertEqual(t, marked, 2)

	_, err := r.Resolve("svc")
	testkit.AssertErrorIs(t, err, ErrNoHealthyInstances)
}

func TestMarkStale_DoesNotAffectAlreadyUnhealthy(t *testing.T) {
	now := time.Now()
	r := NewRegistry()
	r.nowFunc = func() time.Time { return now }

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a", Status: StatusUnhealthy})

	now = now.Add(10 * time.Minute)
	marked := r.MarkStale(5 * time.Minute)
	testkit.AssertEqual(t, marked, 0)
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
	testkit.AssertLen(t, services, 2)

	billing, _ := r.ResolveAll("billing")
	testkit.AssertLen(t, billing, 2)
}

func TestRegisterStatic_Error(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterStatic([]Instance{
		{Service: "good", ID: "1", Addr: "http://a"},
		{Service: "", ID: "2", Addr: "http://b"}, // invalid
	})
	testkit.AssertError(t, err)
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
	testkit.AssertEqual(t, inst.Metadata["version"], "2.0")
	testkit.AssertEqual(t, inst.Metadata["region"], "us-east")
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

	testkit.AssertTrue(t, called)
}

func TestResolveAll_EmptyService(t *testing.T) {
	r := NewRegistry()
	_, err := r.ResolveAll("")
	testkit.AssertErrorIs(t, err, ErrEmptyService)
}

func TestResolveAll_ServiceNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.ResolveAll("nonexistent")
	testkit.AssertErrorIs(t, err, ErrServiceNotFound)
}

func TestResolveHealthy_ErrorPropagation(t *testing.T) {
	r := NewRegistry()
	_, err := r.ResolveHealthy("")
	testkit.AssertErrorIs(t, err, ErrEmptyService)
	_, err = r.ResolveHealthy("nonexistent")
	testkit.AssertErrorIs(t, err, ErrServiceNotFound)
}

func TestHeartbeat_InstanceNotFound(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	// Service exists but instance ID does not.
	err := r.Heartbeat("svc", "nonexistent")
	testkit.AssertErrorIs(t, err, ErrServiceNotFound)
}

func TestMarkStale_WithCallback(t *testing.T) {
	now := time.Now()
	var cbFrom, cbTo Status
	r := NewRegistry(WithOnInstanceStateChange(func(inst Instance, from, to Status) {
		cbFrom = from
		cbTo = to
	}))
	r.nowFunc = func() time.Time { return now }

	_ = r.Register(Instance{Service: "svc", ID: "1", Addr: "http://a"})
	// Advance time beyond TTL.
	r.nowFunc = func() time.Time { return now.Add(10 * time.Minute) }
	marked := r.MarkStale(5 * time.Minute)

	testkit.AssertEqual(t, marked, 1)
	testkit.AssertEqual(t, cbFrom, StatusHealthy)
	testkit.AssertEqual(t, cbTo, StatusUnhealthy)
}
