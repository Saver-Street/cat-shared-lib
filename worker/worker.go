package worker

import (
	"context"
	"log/slog"
	"sync"
)

// Job is a unit of work submitted to the pool. If it returns a non-nil
// error, that error is collected and returned by [Pool.Shutdown].
type Job func(ctx context.Context) error

// Pool is a bounded worker pool that processes jobs concurrently.
// The zero value is not usable; create with [New].
type Pool struct {
	jobs chan Job
	wg   sync.WaitGroup
	ctx  context.Context
	stop context.CancelFunc

	mu   sync.Mutex
	errs []error
}

// New creates a pool with n concurrent workers. Workers start immediately
// and process jobs as they are submitted.
func New(n int) *Pool {
	if n < 1 {
		n = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		jobs: make(chan Job, n*2),
		ctx:  ctx,
		stop: cancel,
	}
	for range n {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

// Submit adds a job to the pool. It blocks if all workers are busy and the
// internal buffer is full. Submit panics if called after Shutdown.
func (p *Pool) Submit(j Job) {
	p.jobs <- j
}

// Shutdown signals workers to finish remaining jobs and waits for them.
// It returns all errors collected from failed jobs.
func (p *Pool) Shutdown(ctx context.Context) []error {
	close(p.jobs)
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		p.stop()
		<-done
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	return p.errs
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for j := range p.jobs {
		p.run(j)
	}
}

func (p *Pool) run(j Job) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("worker: job panicked", "panic", r)
		}
	}()
	if err := j(p.ctx); err != nil {
		p.mu.Lock()
		p.errs = append(p.errs, err)
		p.mu.Unlock()
	}
}
