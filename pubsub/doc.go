// Package pubsub provides a lightweight, typed, in-process publish/subscribe
// event bus for decoupling components within a service.
//
// The bus is generic over the event type. Create a bus with [New], subscribe
// with [Bus.Subscribe], and publish with [Bus.Publish].
//
//	type OrderPlaced struct{ ID string }
//
//	bus := pubsub.New[OrderPlaced]()
//	bus.Subscribe(func(ctx context.Context, e OrderPlaced) {
//	    log.Printf("order %s placed", e.ID)
//	})
//	bus.Publish(ctx, OrderPlaced{ID: "123"})
//
// Subscriptions are identified by an opaque token returned by Subscribe.
// Call [Bus.Unsubscribe] with the token to remove a handler.
//
// Publish is synchronous by default; all handlers run in the caller's
// goroutine. Use [Bus.PublishAsync] for non-blocking delivery.
package pubsub
