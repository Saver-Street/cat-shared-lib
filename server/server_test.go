package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestDefaults(t *testing.T) {
	c := Config{}
	c.defaults()

	testkit.AssertEqual(t, c.ReadTimeout, 15*time.Second)
	testkit.AssertEqual(t, c.WriteTimeout, 30*time.Second)
	testkit.AssertEqual(t, c.IdleTimeout, 60*time.Second)
	testkit.AssertEqual(t, c.ShutdownTimeout, 10*time.Second)
}

func TestDefaults_CustomValues(t *testing.T) {
	c := Config{
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     20 * time.Second,
		ShutdownTimeout: 3 * time.Second,
	}
	c.defaults()

	testkit.AssertEqual(t, c.ReadTimeout, 5*time.Second)
	testkit.AssertEqual(t, c.WriteTimeout, 10*time.Second)
	testkit.AssertEqual(t, c.IdleTimeout, 20*time.Second)
	testkit.AssertEqual(t, c.ShutdownTimeout, 3*time.Second)
}

func TestListenAndServe_GracefulShutdown(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// Pick a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	cleanupCalled := false
	done := make(chan error, 1)
	go func() {
		done <- ListenAndServe(Config{
			Addr:            addr,
			Handler:         mux,
			ShutdownTimeout: 2 * time.Second,
		}, func() {
			cleanupCalled = true
		})
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Verify server responds
	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("server not responding: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Send SIGTERM to self to trigger shutdown
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	proc.Signal(syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for shutdown")
	}

	testkit.AssertTrue(t, cleanupCalled)
}

func TestListenAndServe_BadAddr(t *testing.T) {
	// Bind to the same addr twice to cause an error
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cleanupCalled := false
	done := make(chan error, 1)
	go func() {
		done <- ListenAndServe(Config{
			Addr:    ln.Addr().String(),
			Handler: http.NewServeMux(),
		}, func() {
			cleanupCalled = true
		})
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error for occupied port")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for error")
	}

	testkit.AssertTrue(t, cleanupCalled)
}

func TestListenAndServe_ShutdownTimeout(t *testing.T) {
	// Handler that blocks until explicitly released, keeping a connection
	// alive so that Shutdown's context deadline is exceeded.
	release := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		<-release
		w.WriteHeader(200)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	cleanupCalled := false
	done := make(chan error, 1)
	go func() {
		done <- ListenAndServe(Config{
			Addr:            addr,
			Handler:         mux,
			ShutdownTimeout: 1 * time.Millisecond,
		}, func() {
			cleanupCalled = true
		})
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start a slow request that keeps a connection busy
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err == nil {
			resp.Body.Close()
		}
	}()

	// Wait for the request to be in-flight
	time.Sleep(50 * time.Millisecond)

	// Send SIGTERM to trigger shutdown
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	proc.Signal(syscall.SIGTERM)

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected shutdown timeout error, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for shutdown")
	}

	testkit.AssertTrue(t, cleanupCalled)

	// Release the blocked handler to avoid leaking goroutines
	close(release)
}

func TestConfig_Fields(t *testing.T) {
	mux := http.NewServeMux()
	cfg := Config{
		Addr:            ":0",
		Handler:         mux,
		ReadTimeout:     1 * time.Second,
		WriteTimeout:    2 * time.Second,
		IdleTimeout:     3 * time.Second,
		ShutdownTimeout: 4 * time.Second,
	}

	testkit.AssertEqual(t, cfg.Addr, ":0")
	testkit.AssertNotNil(t, cfg.Handler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = ctx // just verify context is usable
}

func TestDefaults_PartialOverride(t *testing.T) {
	c := Config{
		ReadTimeout: 5 * time.Second,
	}
	c.defaults()
	testkit.AssertEqual(t, c.ReadTimeout, 5*time.Second)
	testkit.AssertEqual(t, c.WriteTimeout, 30*time.Second)
}

func TestListenAndServe_NoCleanupFuncs(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	done := make(chan error, 1)
	go func() {
		done <- ListenAndServe(Config{
			Addr:            addr,
			Handler:         http.NewServeMux(),
			ShutdownTimeout: 2 * time.Second,
		})
	}()

	time.Sleep(100 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	proc.Signal(syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for shutdown")
	}
}

func BenchmarkDefaults(b *testing.B) {
	for b.Loop() {
		c := Config{}
		c.defaults()
	}
}

func TestListenAndServe_NilHandler(t *testing.T) {
	err := ListenAndServe(Config{Addr: ":0", Handler: nil})
	if err == nil {
		t.Fatal("expected error for nil Handler, got nil")
	}
	testkit.AssertEqual(t, err.Error(), "server: Handler must not be nil")
}
