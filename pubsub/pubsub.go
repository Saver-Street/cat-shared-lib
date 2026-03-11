package pubsub

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
)

// Handler is a function that processes an event of type T.
type Handler[T any] func(ctx context.Context, event T)

// Token identifies a subscription. Use it with [Bus.Unsubscribe].
type Token uint64

var nextToken atomic.Uint64

// Bus is a typed, in-process publish/subscribe event bus. It is safe for
// concurrent use. The zero value is not usable; create with [New].
type Bus[T any] struct {
	mu   sync.RWMutex
	subs map[Token]Handler[T]
}

// New creates a new event bus for events of type T.
func New[T any]() *Bus[T] {
	return &Bus[T]{subs: make(map[Token]Handler[T])}
}

// Subscribe registers a handler and returns a token that can be used to
// unsubscribe later.
func (b *Bus[T]) Subscribe(h Handler[T]) Token {
	tok := Token(nextToken.Add(1))
	b.mu.Lock()
	b.subs[tok] = h
	b.mu.Unlock()
	return tok
}

// Unsubscribe removes the handler associated with the given token.
// It is a no-op if the token is not found.
func (b *Bus[T]) Unsubscribe(tok Token) {
	b.mu.Lock()
	delete(b.subs, tok)
	b.mu.Unlock()
}

// Publish delivers event to all subscribers synchronously in the caller's
// goroutine. Handlers that panic are recovered and logged via slog.
func (b *Bus[T]) Publish(ctx context.Context, event T) {
	b.mu.RLock()
	handlers := make([]Handler[T], 0, len(b.subs))
	for _, h := range b.subs {
		handlers = append(handlers, h)
	}
	b.mu.RUnlock()

	for _, h := range handlers {
		b.safeCall(ctx, event, h)
	}
}

// PublishAsync delivers event to all subscribers concurrently, each in its
// own goroutine. It returns immediately. Handlers that panic are recovered
// and logged via slog.
func (b *Bus[T]) PublishAsync(ctx context.Context, event T) {
	b.mu.RLock()
	handlers := make([]Handler[T], 0, len(b.subs))
	for _, h := range b.subs {
		handlers = append(handlers, h)
	}
	b.mu.RUnlock()

	for _, h := range handlers {
		go b.safeCall(ctx, event, h)
	}
}

// Len returns the number of active subscriptions.
func (b *Bus[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}

func (b *Bus[T]) safeCall(ctx context.Context, event T, h Handler[T]) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("pubsub: handler panicked", "panic", r)
		}
	}()
	h(ctx, event)
}
