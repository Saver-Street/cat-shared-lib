package apperror

import (
	"errors"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		New(400, CodeBadRequest, "invalid input")
	}
}

func BenchmarkWrap(b *testing.B) {
	inner := errors.New("root cause")
	for b.Loop() {
		Wrap(500, CodeInternal, "something broke", inner)
	}
}

func BenchmarkError_Error(b *testing.B) {
	e := New(404, CodeNotFound, "resource not found")
	for b.Loop() {
		_ = e.Error()
	}
}

func BenchmarkError_ErrorWrapped(b *testing.B) {
	e := Wrap(500, CodeInternal, "failed", errors.New("db error"))
	for b.Loop() {
		_ = e.Error()
	}
}

func BenchmarkHTTPStatus(b *testing.B) {
	e := New(404, CodeNotFound, "not found")
	for b.Loop() {
		HTTPStatus(e)
	}
}

func BenchmarkHTTPStatus_PlainError(b *testing.B) {
	e := errors.New("plain error")
	for b.Loop() {
		HTTPStatus(e)
	}
}

func BenchmarkIsCode(b *testing.B) {
	e := New(404, CodeNotFound, "not found")
	for b.Loop() {
		IsCode(e, CodeNotFound)
	}
}

func BenchmarkIsCode_Wrapped(b *testing.B) {
	inner := New(404, CodeNotFound, "not found")
	outer := Wrap(500, CodeInternal, "wrapped", inner)
	for b.Loop() {
		IsCode(outer, CodeNotFound)
	}
}
