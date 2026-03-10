package server_test

import (
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/server"
)

func ExampleConfig_defaults() {
	cfg := server.Config{}
	// Zero-value Config gets sensible defaults when passed to ListenAndServe.
	fmt.Printf("Addr: %q\n", cfg.Addr)
	fmt.Println("ReadTimeout:", cfg.ReadTimeout)
	// Output:
	// Addr: ""
	// ReadTimeout: 0s
}

func ExampleConfig_custom() {
	cfg := server.Config{
		Addr:            ":8080",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
	fmt.Println("Addr:", cfg.Addr)
	fmt.Println("ReadTimeout:", cfg.ReadTimeout)
	fmt.Println("ShutdownTimeout:", cfg.ShutdownTimeout)
	// Output:
	// Addr: :8080
	// ReadTimeout: 5s
	// ShutdownTimeout: 5s
}
