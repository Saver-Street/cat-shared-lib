package httpclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Saver-Street/cat-shared-lib/circuitbreaker"
	"github.com/Saver-Street/cat-shared-lib/httpclient"
)

func ExampleNew() {
	client := httpclient.New(
		httpclient.WithTimeout(10*time.Second),
		httpclient.WithRetries(3),
		httpclient.WithUserAgent("my-service/1.0"),
		httpclient.WithHeader("Authorization", "Bearer token"),
	)
	_ = client // use client for HTTP calls
	fmt.Println("client created")
	// Output: client created
}

func ExampleClient_GetJSON() {
	type User struct {
		Name string `json:"name"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(User{Name: "Alice"})
	}))
	defer srv.Close()

	client := httpclient.New()
	var user User
	if err := client.GetJSON(context.Background(), srv.URL, &user); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(user.Name)
	// Output: Alice
}

func ExampleWithCircuitBreaker() {
	cb := circuitbreaker.New("api",
		circuitbreaker.WithFailureThreshold(5),
		circuitbreaker.WithResetTimeout(30*time.Second),
	)

	client := httpclient.New(
		httpclient.WithCircuitBreaker(cb),
		httpclient.WithRetries(2),
	)
	_ = client
	fmt.Println("client with circuit breaker")
	// Output: client with circuit breaker
}
