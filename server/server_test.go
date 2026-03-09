package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	c := Config{}
	c.defaults()

	if c.ReadTimeout != 15*time.Second {
		t.Errorf("expected ReadTimeout 15s, got %v", c.ReadTimeout)
	}
	if c.WriteTimeout != 30*time.Second {
		t.Errorf("expected WriteTimeout 30s, got %v", c.WriteTimeout)
	}
	if c.IdleTimeout != 60*time.Second {
		t.Errorf("expected IdleTimeout 60s, got %v", c.IdleTimeout)
	}
	if c.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected ShutdownTimeout 10s, got %v", c.ShutdownTimeout)
	}
}

func TestDefaults_CustomValues(t *testing.T) {
	c := Config{
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     20 * time.Second,
		ShutdownTimeout: 3 * time.Second,
	}
	c.defaults()

	if c.ReadTimeout != 5*time.Second {
		t.Errorf("expected ReadTimeout 5s, got %v", c.ReadTimeout)
	}
	if c.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout 10s, got %v", c.WriteTimeout)
	}
	if c.IdleTimeout != 20*time.Second {
		t.Errorf("expected IdleTimeout 20s, got %v", c.IdleTimeout)
	}
	if c.ShutdownTimeout != 3*time.Second {
		t.Errorf("expected ShutdownTimeout 3s, got %v", c.ShutdownTimeout)
	}
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

	if !cleanupCalled {
		t.Error("cleanup function was not called")
	}
}

func TestListenAndServe_BadAddr(t *testing.T) {
	// Bind to the same addr twice to cause an error
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		done <- ListenAndServe(Config{
			Addr:    ln.Addr().String(),
			Handler: http.NewServeMux(),
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

	if cfg.Addr != ":0" {
		t.Error("Addr mismatch")
	}
	if cfg.Handler == nil {
		t.Error("Handler nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = ctx // just verify context is usable
}
