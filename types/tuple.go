package types

// Tuple3 holds three typed values.
type Tuple3[A, B, C any] struct {
	First  A
	Second B
	Third  C
}

// NewTuple3 creates a Tuple3.
func NewTuple3[A, B, C any](a A, b B, c C) Tuple3[A, B, C] {
	return Tuple3[A, B, C]{First: a, Second: b, Third: c}
}

// Unpack returns all three values.
func (t Tuple3[A, B, C]) Unpack() (A, B, C) {
	return t.First, t.Second, t.Third
}

// Tuple4 holds four typed values.
type Tuple4[A, B, C, D any] struct {
	First  A
	Second B
	Third  C
	Fourth D
}

// NewTuple4 creates a Tuple4.
func NewTuple4[A, B, C, D any](a A, b B, c C, d D) Tuple4[A, B, C, D] {
	return Tuple4[A, B, C, D]{First: a, Second: b, Third: c, Fourth: d}
}

// Unpack returns all four values.
func (t Tuple4[A, B, C, D]) Unpack() (A, B, C, D) {
	return t.First, t.Second, t.Third, t.Fourth
}

// Tuple5 holds five typed values.
type Tuple5[A, B, C, D, E any] struct {
	First  A
	Second B
	Third  C
	Fourth D
	Fifth  E
}

// NewTuple5 creates a Tuple5.
func NewTuple5[A, B, C, D, E any](a A, b B, c C, d D, e E) Tuple5[A, B, C, D, E] {
	return Tuple5[A, B, C, D, E]{First: a, Second: b, Third: c, Fourth: d, Fifth: e}
}

// Unpack returns all five values.
func (t Tuple5[A, B, C, D, E]) Unpack() (A, B, C, D, E) {
	return t.First, t.Second, t.Third, t.Fourth, t.Fifth
}
