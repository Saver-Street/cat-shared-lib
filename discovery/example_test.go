package discovery_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/discovery"
)

func ExampleNewRegistry() {
	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{
		Service: "billing-service",
		ID:      "billing-1",
		Addr:    "http://billing-1:8080",
	})

	inst, err := reg.Resolve("billing-service")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(inst.Addr)
	// Output: http://billing-1:8080
}

func ExampleRegistry_RegisterStatic() {
	reg := discovery.NewRegistry()
	err := reg.RegisterStatic([]discovery.Instance{
		{Service: "auth", ID: "a-1", Addr: "http://auth-1:8080"},
		{Service: "auth", ID: "a-2", Addr: "http://auth-2:8080"},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	services := reg.Services()
	fmt.Println("services:", len(services))
	// Output: services: 1
}

func ExampleRegistry_Resolve() {
	reg := discovery.NewRegistry()
	_ = reg.Register(discovery.Instance{Service: "svc", ID: "1", Addr: "http://a"})
	_ = reg.Register(discovery.Instance{Service: "svc", ID: "2", Addr: "http://b"})

	// Round-robin across instances
	inst1, _ := reg.Resolve("svc")
	inst2, _ := reg.Resolve("svc")
	fmt.Println(inst1.Addr, inst2.Addr)
	// Output: http://a http://b
}
