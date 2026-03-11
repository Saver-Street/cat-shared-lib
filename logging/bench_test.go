package logging

import (
	"bytes"
	"log/slog"
	"testing"
)

func BenchmarkNew_JSON(b *testing.B) {
	for b.Loop() {
		New(WithWriter(&bytes.Buffer{}), WithFormat("json"))
	}
}

func BenchmarkNew_Text(b *testing.B) {
	for b.Loop() {
		New(WithWriter(&bytes.Buffer{}), WithFormat("text"))
	}
}

func BenchmarkLogInfo(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithWriter(&buf), WithFormat("json"), WithLevel(slog.LevelInfo))
	for b.Loop() {
		l.Info("benchmark", "key", "value")
	}
}

func BenchmarkParseLevel(b *testing.B) {
	for b.Loop() {
		ParseLevel("info")
	}
}
