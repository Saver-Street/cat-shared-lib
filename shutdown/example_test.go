package shutdown_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Saver-Street/cat-shared-lib/shutdown"
)

func ExampleDrainer() {
	var d shutdown.Drainer

	// Simulate tracking two in-flight requests.
	d.Add()
	d.Add()
	d.Done()
	d.Done()

	// Wait returns immediately when count reaches zero.
	d.Wait()
	fmt.Println("all drained")
	// Output:
	// all drained
}

func ExampleDrainer_Middleware() {
	var d shutdown.Drainer

	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rr.Code)
	// Output:
	// 200
}

func ExampleRunHooks() {
	hooks := []shutdown.Hook{
		{Name: "close-db", Fn: func(ctx context.Context) error {
			fmt.Println("closing database")
			return nil
		}},
		{Name: "flush-cache", Fn: func(ctx context.Context) error {
			fmt.Println("flushing cache")
			return nil
		}},
	}

	err := shutdown.RunHooks(context.Background(), nil, hooks)
	fmt.Println("error:", err)
	// Output:
	// closing database
	// flushing cache
	// error: <nil>
}

func ExampleConfig_AddHook() {
	cfg := &shutdown.Config{}
	cfg.AddHook("close-db", func(ctx context.Context) error {
		return nil
	}).AddHook("flush-cache", func(ctx context.Context) error {
		return nil
	})

	fmt.Println("hooks registered:", len(cfg.OnShutdown))
	// Output:
	// hooks registered: 2
}
