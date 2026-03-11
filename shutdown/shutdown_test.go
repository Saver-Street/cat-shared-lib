package shutdown

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	cfg.defaults()

	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}
	if len(cfg.Signals) != 2 {
		t.Errorf("expected 2 default signals, got %d", len(cfg.Signals))
	}
	if cfg.Logger == nil {
		t.Error("expected non-nil default logger")
	}
}

func TestConfigCustom(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	cfg := Config{
		Timeout: 10 * time.Second,
		Signals: []os.Signal{syscall.SIGUSR1},
		Logger:  logger,
	}
	cfg.defaults()

	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected custom timeout 10s, got %v", cfg.Timeout)
	}
	if len(cfg.Signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(cfg.Signals))
	}
}

func TestDrainer_AddDoneWait(t *testing.T) {
	d := &Drainer{}
	done := make(chan struct{})

	d.Add()
	go func() {
		d.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Wait should block while there are active connections")
	case <-time.After(50 * time.Millisecond):
	}

	d.Done()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Wait should unblock after Done")
	}
}

func TestDrainer_Middleware(t *testing.T) {
	d := &Drainer{}
	var tracked atomic.Bool

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracked.Store(true)
		w.WriteHeader(http.StatusOK)
	})

	handler := d.Middleware(inner)

	// Track requests through middleware.
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := &fakeResponseWriter{}
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			handler.ServeHTTP(w, r)
		}()
	}
	wg.Wait()

	if !tracked.Load() {
		t.Error("inner handler was never called")
	}
}

func TestListenAndServe_GracefulShutdown(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Find a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	hookCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	cfg := Config{
		Timeout: 5 * time.Second,
		Signals: []os.Signal{syscall.SIGUSR1},
		Logger:  logger,
		OnShutdown: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				hookCalled = true
				return nil
			},
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- ListenAndServe(srv, cfg)
	}()

	// Wait for server to start.
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Send shutdown signal.
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGUSR1)

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown timed out")
	}

	if !hookCalled {
		t.Error("OnShutdown hook was not called")
	}
	if !bytes.Contains(buf.Bytes(), []byte("server stopped gracefully")) {
		t.Error("expected graceful stop log message")
	}
}

func TestListenAndServe_OnShutdownError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	srv := &http.Server{Addr: addr, Handler: http.NewServeMux()}
	cfg := Config{
		Timeout: 2 * time.Second,
		Signals: []os.Signal{syscall.SIGUSR2},
		Logger:  logger,
		OnShutdown: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				return errors.New("cleanup failed")
			},
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- ListenAndServe(srv, cfg)
	}()

	// Wait for server to start.
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGUSR2)

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error from shutdown, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown timed out")
	}

	if !bytes.Contains(buf.Bytes(), []byte("cleanup failed")) {
		t.Error("expected hook error in logs")
	}
}

func TestListenAndServe_BindError(t *testing.T) {
	// Occupy a port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	srv := &http.Server{Addr: addr, Handler: http.NewServeMux()}
	cfg := Config{Timeout: time.Second}

	err = ListenAndServe(srv, cfg)
	if err == nil {
		t.Error("expected bind error")
	}
}

// fakeResponseWriter is a minimal http.ResponseWriter for unit tests.
type fakeResponseWriter struct {
	code int
}

func (f *fakeResponseWriter) Header() http.Header         { return http.Header{} }
func (f *fakeResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (f *fakeResponseWriter) WriteHeader(code int)        { f.code = code }

func TestDrainer_ConcurrentMiddleware(t *testing.T) {
	d := &Drainer{}
	var count atomic.Int64

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := d.Middleware(inner)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := &fakeResponseWriter{}
			r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/%d", i), nil)
			handler.ServeHTTP(w, r)
		}()
	}

	wg.Wait()
	d.Wait() // All should be drained.

	if count.Load() != 50 {
		t.Errorf("expected 50 requests, got %d", count.Load())
	}
}

func TestWaitForSignal(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := Config{
		Timeout: 5 * time.Second,
		Signals: []os.Signal{syscall.SIGUSR1},
		Logger:  logger,
	}

	done := make(chan struct{})
	var ctx context.Context
	var cancel context.CancelFunc

	go func() {
		ctx, cancel = WaitForSignal(cfg)
		close(done)
	}()

	// Give the goroutine time to register the signal handler.
	time.Sleep(50 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGUSR1)

	select {
	case <-done:
		defer cancel()
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected context with deadline")
		}
		remaining := time.Until(deadline)
		if remaining < 4*time.Second || remaining > 6*time.Second {
			t.Errorf("expected ~5s remaining, got %v", remaining)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WaitForSignal did not return")
	}

	if !bytes.Contains(buf.Bytes(), []byte("shutdown signal received")) {
		t.Error("expected signal log message")
	}
}

func TestListenAndServe_ShutdownTimeout(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	// Handler that blocks longer than the shutdown timeout.
	reqStarted := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(reqStarted)
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	cfg := Config{
		Timeout: 1 * time.Millisecond, // very short timeout to force shutdown error
		Signals: []os.Signal{syscall.SIGUSR1},
		Logger:  logger,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- ListenAndServe(srv, cfg)
	}()

	// Wait for server to start.
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Start a long-running request to keep the server busy.
	go http.Get("http://" + addr + "/slow") //nolint:errcheck

	// Wait for the request to be in-flight.
	<-reqStarted

	// Send shutdown signal.
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGUSR1)

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("expected shutdown timeout error, got nil")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("test timed out")
	}

	if !bytes.Contains(buf.Bytes(), []byte("server shutdown error")) {
		t.Error("expected shutdown error in logs")
	}
}
