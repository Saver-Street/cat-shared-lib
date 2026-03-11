package webhook_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Saver-Street/cat-shared-lib/webhook"
)

func ExampleSign() {
	sig := webhook.Sign([]byte("my-secret"), []byte(`{"event":"test"}`))
	fmt.Println(strings.HasPrefix(sig, "sha256="))
	// Output:
	// true
}

func ExampleVerify() {
	secret := []byte("my-secret")
	payload := []byte(`{"event":"test"}`)
	sig := webhook.Sign(secret, payload)

	fmt.Println(webhook.Verify(secret, payload, sig))
	fmt.Println(webhook.Verify(secret, payload, "sha256=wrong"))
	// Output:
	// true
	// false
}

func ExampleDeliver() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	type Event struct {
		Action string `json:"action"`
	}

	err := webhook.Deliver(context.Background(), srv.URL, []byte("secret"), "user.created", Event{Action: "signup"})
	fmt.Println(err)
	// Output:
	// <nil>
}
