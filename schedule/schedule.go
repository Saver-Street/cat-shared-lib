package schedule

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Task is a function executed periodically by the scheduler.
type Task func(ctx context.Context)

type entry struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// Scheduler runs named tasks at fixed intervals. It is safe for concurrent
// use.
type Scheduler struct {
	mu      sync.Mutex
	entries map[string]*entry
	ctx     context.Context
	cancel  context.CancelFunc
}

// New creates a new Scheduler. Call [Scheduler.Stop] to stop all tasks.
func New() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		entries: make(map[string]*entry),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Every registers a named task that runs fn every interval. If a task with
// the same name already exists, it is replaced. The first execution happens
// after one interval elapses.
func (s *Scheduler) Every(name string, interval time.Duration, fn Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel existing task with the same name.
	if e, ok := s.entries[name]; ok {
		e.cancel()
		<-e.done
	}

	taskCtx, taskCancel := context.WithCancel(s.ctx)
	done := make(chan struct{})
	s.entries[name] = &entry{cancel: taskCancel, done: done}

	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-taskCtx.Done():
				return
			case <-ticker.C:
				s.runTask(taskCtx, name, fn)
			}
		}
	}()
}

// Remove cancels and removes a named task. It is a no-op if the task does
// not exist.
func (s *Scheduler) Remove(name string) {
	s.mu.Lock()
	e, ok := s.entries[name]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(s.entries, name)
	s.mu.Unlock()

	e.cancel()
	<-e.done
}

// Len returns the number of registered tasks.
func (s *Scheduler) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// Stop cancels all tasks and waits for them to finish.
func (s *Scheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	entries := make([]*entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.entries = make(map[string]*entry)
	s.mu.Unlock()

	for _, e := range entries {
		<-e.done
	}
}

func (s *Scheduler) runTask(ctx context.Context, name string, fn Task) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("schedule: task panicked", "task", name, "panic", r)
		}
	}()
	fn(ctx)
}
