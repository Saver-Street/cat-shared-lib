package types

import "encoding/json"

// Either represents a value of one of two types. It is either Left(L) or
// Right(R). By convention Left holds error/failure values and Right holds
// success values, but callers can use any semantics.
type Either[L, R any] struct {
	left    L
	right   R
	isRight bool
}

// Left creates an Either holding a left value.
func Left[L, R any](v L) Either[L, R] {
	return Either[L, R]{left: v}
}

// Right creates an Either holding a right value.
func Right[L, R any](v R) Either[L, R] {
	return Either[L, R]{right: v, isRight: true}
}

// IsLeft reports whether the Either holds a left value.
func (e Either[L, R]) IsLeft() bool { return !e.isRight }

// IsRight reports whether the Either holds a right value.
func (e Either[L, R]) IsRight() bool { return e.isRight }

// LeftVal returns the left value and true, or the zero value and false.
func (e Either[L, R]) LeftVal() (L, bool) {
	if e.isRight {
		var zero L
		return zero, false
	}
	return e.left, true
}

// RightVal returns the right value and true, or the zero value and false.
func (e Either[L, R]) RightVal() (R, bool) {
	if !e.isRight {
		var zero R
		return zero, false
	}
	return e.right, true
}

// LeftOr returns the left value or defaultVal if Right.
func (e Either[L, R]) LeftOr(defaultVal L) L {
	if e.isRight {
		return defaultVal
	}
	return e.left
}

// RightOr returns the right value or defaultVal if Left.
func (e Either[L, R]) RightOr(defaultVal R) R {
	if !e.isRight {
		return defaultVal
	}
	return e.right
}

// eitherJSON is the JSON representation for Either.
type eitherJSON[L, R any] struct {
	Left  *L `json:"left,omitempty"`
	Right *R `json:"right,omitempty"`
}

// MarshalJSON encodes the Either as {"left": v} or {"right": v}.
func (e Either[L, R]) MarshalJSON() ([]byte, error) {
	if e.isRight {
		return json.Marshal(eitherJSON[L, R]{Right: &e.right})
	}
	return json.Marshal(eitherJSON[L, R]{Left: &e.left})
}

// UnmarshalJSON decodes {"left": v} or {"right": v}.
func (e *Either[L, R]) UnmarshalJSON(data []byte) error {
	var ej eitherJSON[L, R]
	if err := json.Unmarshal(data, &ej); err != nil {
		return err
	}
	if ej.Right != nil {
		e.right = *ej.Right
		e.isRight = true
		return nil
	}
	if ej.Left != nil {
		e.left = *ej.Left
		e.isRight = false
		return nil
	}
	// Default to left zero value
	e.isRight = false
	return nil
}
