package circuitbreaker

import (
	"errors"
	"testing"
)

func BenchmarkExecute_Success(b *testing.B) {
	cb := New("bench")
	for b.Loop() {
		cb.Execute(func() error { return nil })
	}
}

func BenchmarkExecute_Failure(b *testing.B) {
	errFail := errors.New("fail")
	cb := New("bench", WithFailureThreshold(1_000_000))
	for b.Loop() {
		cb.Execute(func() error { return errFail })
	}
}

func BenchmarkExecute_OpenCircuit(b *testing.B) {
	errFail := errors.New("fail")
	cb := New("bench", WithFailureThreshold(1))
	cb.Execute(func() error { return errFail })
	b.ResetTimer()
	for b.Loop() {
		cb.Execute(func() error { return nil })
	}
}

func BenchmarkState(b *testing.B) {
	cb := New("bench")
	for b.Loop() {
		cb.State()
	}
}

func BenchmarkCounts(b *testing.B) {
	cb := New("bench")
	for b.Loop() {
		cb.Counts()
	}
}

func BenchmarkExecute_Parallel(b *testing.B) {
	cb := New("bench")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() error { return nil })
		}
	})
}
