// Package discovery provides service discovery and registration for
// microservice communication. It supports static configuration and runtime
// registration of service instances, with health-check-aware routing and
// round-robin load balancing across healthy instances.
//
// Usage:
//
//	reg := discovery.NewRegistry()
//
//	// Register services from static config or at runtime.
//	reg.Register(discovery.Instance{
//	    Service: "billing-service",
//	    ID:      "billing-1",
//	    Addr:    "http://billing-1:8080",
//	})
//
//	// Resolve a healthy instance.
//	inst, err := reg.Resolve("billing-service")
package discovery

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Sentinel errors.
var (
	// ErrServiceNotFound is returned when no instances are registered for a service.
	ErrServiceNotFound = errors.New("discovery: service not found")
	// ErrNoHealthyInstances is returned when all instances of a service are unhealthy.
	ErrNoHealthyInstances = errors.New("discovery: no healthy instances available")
	// ErrEmptyService is returned when a service name is empty.
	ErrEmptyService = errors.New("discovery: service name must not be empty")
	// ErrEmptyInstanceID is returned when an instance ID is empty.
	ErrEmptyInstanceID = errors.New("discovery: instance ID must not be empty")
	// ErrEmptyAddr is returned when an instance address is empty.
	ErrEmptyAddr = errors.New("discovery: instance address must not be empty")
)

// Status represents the health status of a service instance.
type Status int

const (
	// StatusHealthy indicates the instance is healthy and can receive traffic.
	StatusHealthy Status = iota
	// StatusUnhealthy indicates the instance is unhealthy and should not receive traffic.
	StatusUnhealthy
	// StatusDraining indicates the instance is draining and should not receive new traffic.
	StatusDraining
)

// String returns the human-readable name of the status.
func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	case StatusDraining:
		return "draining"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// Instance represents a single service instance in the registry.
type Instance struct {
	// Service is the logical service name (e.g., "billing-service").
	Service string

	// ID is the unique identifier for this instance (e.g., "billing-1").
	ID string

	// Addr is the network address including scheme (e.g., "http://billing-1:8080").
	Addr string

	// Metadata holds arbitrary key-value pairs (e.g., version, region).
	Metadata map[string]string

	// Status is the current health status. Default: StatusHealthy.
	Status Status

	// RegisteredAt is when the instance was registered.
	RegisteredAt time.Time

	// LastSeen is the last time the instance was updated or health-checked.
	LastSeen time.Time
}

// IsHealthy returns true if the instance can receive traffic.
func (i Instance) IsHealthy() bool {
	return i.Status == StatusHealthy
}

// serviceEntry holds the instances and round-robin counter for a service.
type serviceEntry struct {
	instances []Instance
	counter   uint64
}

// StateChangeFunc is called when an instance's status changes.
type StateChangeFunc func(inst Instance, from, to Status)

// Registry is a thread-safe in-memory service registry.
type Registry struct {
	mu            sync.RWMutex
	services      map[string]*serviceEntry
	onStateChange StateChangeFunc
	nowFunc       func() time.Time
}

// RegistryOption configures the registry.
type RegistryOption func(*Registry)

// WithOnInstanceStateChange registers a callback for instance status changes.
func WithOnInstanceStateChange(fn StateChangeFunc) RegistryOption {
	return func(r *Registry) { r.onStateChange = fn }
}

// NewRegistry creates an empty service registry.
func NewRegistry(opts ...RegistryOption) *Registry {
	r := &Registry{
		services: make(map[string]*serviceEntry),
		nowFunc:  time.Now,
	}
	for _, fn := range opts {
		fn(r)
	}
	return r
}

// Register adds or updates an instance in the registry.
// If an instance with the same service+ID already exists, it is updated.
func (r *Registry) Register(inst Instance) error {
	if inst.Service == "" {
		return ErrEmptyService
	}
	if inst.ID == "" {
		return ErrEmptyInstanceID
	}
	if inst.Addr == "" {
		return ErrEmptyAddr
	}

	now := r.nowFunc()
	inst.RegisteredAt = now
	inst.LastSeen = now

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.services[inst.Service]
	if !ok {
		entry = &serviceEntry{}
		r.services[inst.Service] = entry
	}

	// Update existing instance if found.
	for i, existing := range entry.instances {
		if existing.ID == inst.ID {
			oldStatus := existing.Status
			entry.instances[i] = inst
			if oldStatus != inst.Status && r.onStateChange != nil {
				r.onStateChange(inst, oldStatus, inst.Status)
			}
			slog.Info("discovery: instance updated",
				"service", inst.Service,
				"id", inst.ID,
				"addr", inst.Addr,
				"status", inst.Status.String(),
			)
			return nil
		}
	}

	entry.instances = append(entry.instances, inst)
	slog.Info("discovery: instance registered",
		"service", inst.Service,
		"id", inst.ID,
		"addr", inst.Addr,
	)
	return nil
}

// Deregister removes an instance from the registry.
func (r *Registry) Deregister(service, id string) error {
	if service == "" {
		return ErrEmptyService
	}
	if id == "" {
		return ErrEmptyInstanceID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.services[service]
	if !ok {
		return ErrServiceNotFound
	}

	for i, inst := range entry.instances {
		if inst.ID == id {
			entry.instances = append(entry.instances[:i], entry.instances[i+1:]...)
			slog.Info("discovery: instance deregistered",
				"service", service,
				"id", id,
			)
			if len(entry.instances) == 0 {
				delete(r.services, service)
			}
			return nil
		}
	}

	return ErrServiceNotFound
}

// Resolve returns a healthy instance of the given service using round-robin
// load balancing. Only instances with [StatusHealthy] are considered.
func (r *Registry) Resolve(service string) (Instance, error) {
	if service == "" {
		return Instance{}, ErrEmptyService
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.services[service]
	if !ok || len(entry.instances) == 0 {
		return Instance{}, fmt.Errorf("%w: %s", ErrServiceNotFound, service)
	}

	// Round-robin across healthy instances.
	n := uint64(len(entry.instances))
	start := entry.counter
	for i := uint64(0); i < n; i++ {
		idx := (start + i) % n
		inst := entry.instances[idx]
		if inst.IsHealthy() {
			entry.counter = (start + i + 1) % n
			return inst, nil
		}
	}
	entry.counter = (start + 1) % n

	return Instance{}, fmt.Errorf("%w: %s", ErrNoHealthyInstances, service)
}

// ResolveAll returns all instances of the given service regardless of status.
func (r *Registry) ResolveAll(service string) ([]Instance, error) {
	if service == "" {
		return nil, ErrEmptyService
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.services[service]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrServiceNotFound, service)
	}

	result := make([]Instance, len(entry.instances))
	copy(result, entry.instances)
	return result, nil
}

// ResolveHealthy returns only healthy instances of the given service.
func (r *Registry) ResolveHealthy(service string) ([]Instance, error) {
	all, err := r.ResolveAll(service)
	if err != nil {
		return nil, err
	}

	healthy := make([]Instance, 0, len(all))
	for _, inst := range all {
		if inst.IsHealthy() {
			healthy = append(healthy, inst)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoHealthyInstances, service)
	}
	return healthy, nil
}

// SetStatus updates the status of a specific instance.
func (r *Registry) SetStatus(service, id string, status Status) error {
	if service == "" {
		return ErrEmptyService
	}
	if id == "" {
		return ErrEmptyInstanceID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.services[service]
	if !ok {
		return ErrServiceNotFound
	}

	for i, inst := range entry.instances {
		if inst.ID == id {
			oldStatus := inst.Status
			entry.instances[i].Status = status
			entry.instances[i].LastSeen = r.nowFunc()
			if oldStatus != status {
				slog.Info("discovery: instance status changed",
					"service", service,
					"id", id,
					"from", oldStatus.String(),
					"to", status.String(),
				)
				if r.onStateChange != nil {
					r.onStateChange(entry.instances[i], oldStatus, status)
				}
			}
			return nil
		}
	}

	return ErrServiceNotFound
}

// Services returns the names of all registered services.
func (r *Registry) Services() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}

// Heartbeat updates the LastSeen timestamp for an instance.
func (r *Registry) Heartbeat(service, id string) error {
	if service == "" {
		return ErrEmptyService
	}
	if id == "" {
		return ErrEmptyInstanceID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.services[service]
	if !ok {
		return ErrServiceNotFound
	}

	for i, inst := range entry.instances {
		if inst.ID == id {
			entry.instances[i].LastSeen = r.nowFunc()
			return nil
		}
	}

	return ErrServiceNotFound
}

// MarkStale marks instances that haven't been seen within the given TTL as
// unhealthy. Returns the number of instances marked.
func (r *Registry) MarkStale(ttl time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.nowFunc()
	marked := 0

	for _, entry := range r.services {
		for i, inst := range entry.instances {
			if inst.Status == StatusHealthy && now.Sub(inst.LastSeen) > ttl {
				entry.instances[i].Status = StatusUnhealthy
				marked++
				slog.Warn("discovery: marking stale instance as unhealthy",
					"service", inst.Service,
					"id", inst.ID,
					"lastSeen", inst.LastSeen.Format(time.RFC3339),
				)
				if r.onStateChange != nil {
					r.onStateChange(entry.instances[i], StatusHealthy, StatusUnhealthy)
				}
			}
		}
	}

	return marked
}

// RegisterStatic is a convenience method to register multiple instances from
// static configuration.
func (r *Registry) RegisterStatic(instances []Instance) error {
	for _, inst := range instances {
		if err := r.Register(inst); err != nil {
			return fmt.Errorf("discovery: registering %s/%s: %w", inst.Service, inst.ID, err)
		}
	}
	return nil
}
