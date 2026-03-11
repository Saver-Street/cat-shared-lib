package types

import "encoding/json"

// Option represents an optional value that distinguishes between "not set"
// (absent) and "set to the zero value". This is useful for PATCH-style API
// requests where a field being absent means "don't change" while a field
// set to null or zero means "clear/reset".
//
// The zero value of Option is an absent value.
//
//	var name types.Option[string]         // absent
//	name = types.Some("Alice")           // present
//	name = types.None[string]()          // absent
//	name = types.Some("")                // present, but empty
type Option[T any] struct {
	value   T
	present bool
}

// Some creates a present Option containing the given value.
func Some[T any](v T) Option[T] {
	return Option[T]{value: v, present: true}
}

// None creates an absent Option.
func None[T any]() Option[T] {
	return Option[T]{}
}

// IsPresent reports whether the Option contains a value.
func (o Option[T]) IsPresent() bool {
	return o.present
}

// IsAbsent reports whether the Option is empty.
func (o Option[T]) IsAbsent() bool {
	return !o.present
}

// Value returns the contained value and true if present, or the zero value
// and false if absent.
func (o Option[T]) Value() (T, bool) {
	return o.value, o.present
}

// ValueOr returns the contained value if present, or the given fallback.
func (o Option[T]) ValueOr(fallback T) T {
	if o.present {
		return o.value
	}
	return fallback
}

// ValueOrFunc returns the contained value if present, or calls fn to produce
// a fallback. Use this when computing the default is expensive.
func (o Option[T]) ValueOrFunc(fn func() T) T {
	if o.present {
		return o.value
	}
	return fn()
}

// MarshalJSON encodes a present Option as the JSON encoding of its value,
// and an absent Option as JSON null.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.present {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

// UnmarshalJSON decodes a JSON value into the Option. JSON null produces an
// absent Option. Any other JSON value produces a present Option.
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.present = false
		var zero T
		o.value = zero
		return nil
	}
	if err := json.Unmarshal(data, &o.value); err != nil {
		return err
	}
	o.present = true
	return nil
}
