package metrics_test

import (
	"fmt"
	"strings"

	"github.com/Saver-Street/cat-shared-lib/metrics"
)

func ExampleCounter() {
	c := metrics.NewCounter("requests_total", "Total HTTP requests")
	c.Inc()
	c.Inc()
	c.Add(3)
	fmt.Println(c.Value())
	// Output:
	// 5
}

func ExampleGauge() {
	g := metrics.NewGauge("active_connections", "Current active connections")
	g.Set(10)
	g.Inc()
	g.Dec()
	fmt.Println(g.Value())
	// Output:
	// 10
}

func ExampleRegistry_Expose() {
	reg := metrics.NewRegistry()
	c := metrics.NewCounter("http_requests", "Total requests")
	reg.Register(c)
	c.Inc()

	output := reg.Expose()
	fmt.Println(strings.Contains(output, "http_requests"))
	// Output:
	// true
}
