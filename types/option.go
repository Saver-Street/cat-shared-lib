package types

import "encoding/json"

// Option represents an optional value. An Option either contains a
// value (Some) or is empty (None).
type Option[T any] struct {
	value T
	valid bool
}

// Some creates an Option containing v.
func Some[T any](v T) Option[T] {
	return Option[T]{value: v, valid: true}
}

// None returns an empty Option.
func None[T any]() Option[T] {
	var zero Option[T]
	return zero
}

// IsSome reports whether the Option contains a value.
func (o Option[T]) IsSome() bool { return o.valid }

// IsNone reports whether the Option is empty.
func (o Option[T]) IsNone() bool { return !o.valid }

// Unwrap returns the contained value. It panics if the Option is empty.
func (o Option[T]) Unwrap() T {
	if !o.valid {
		panic("option: unwrap called on None")
	}
	return o.value
}

// UnwrapOr returns the contained value or defaultVal if empty.
func (o Option[T]) UnwrapOr(defaultVal T) T {
	if !o.valid {
		return defaultVal
	}
	return o.value
}

// UnwrapOrZero returns the contained value or the zero value of T.
func (o Option[T]) UnwrapOrZero() T {
	return o.value
}

// Get returns the value and whether it is present.
func (o Option[T]) Get() (T, bool) {
	return o.value, o.valid
}

// MarshalJSON encodes the Option as JSON. None encodes as null,
// Some(v) encodes as the JSON representation of v.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.valid {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

// UnmarshalJSON decodes JSON into the Option. A JSON null sets the
// Option to None; any other value is decoded as Some.
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.valid = false
		var zero T
		o.value = zero
		return nil
	}
	if err := json.Unmarshal(data, &o.value); err != nil {
		return err
	}
	o.valid = true
	return nil
}
