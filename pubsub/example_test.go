package pubsub_test

import (
	"context"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/pubsub"
)

func ExampleBus() {
	type OrderPlaced struct{ ID string }

	bus := pubsub.New[OrderPlaced]()

	bus.Subscribe(func(_ context.Context, e OrderPlaced) {
		fmt.Printf("order placed: %s\n", e.ID)
	})

	bus.Publish(context.Background(), OrderPlaced{ID: "ABC-123"})
	// Output:
	// order placed: ABC-123
}

func ExampleBus_Unsubscribe() {
	bus := pubsub.New[string]()

	tok := bus.Subscribe(func(_ context.Context, msg string) {
		fmt.Println(msg)
	})

	bus.Publish(context.Background(), "first")
	bus.Unsubscribe(tok)
	bus.Publish(context.Background(), "second") // not received
	// Output:
	// first
}
