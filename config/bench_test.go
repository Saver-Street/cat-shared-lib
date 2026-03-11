package config

import (
	"testing"
	"time"
)

func BenchmarkBool(b *testing.B) {
	b.Setenv("BENCH_BOOL", "true")
	for b.Loop() {
		Bool("BENCH_BOOL", false)
	}
}

func BenchmarkBool_Default(b *testing.B) {
	for b.Loop() {
		Bool("BENCH_BOOL_MISS", false)
	}
}

func BenchmarkDuration(b *testing.B) {
	b.Setenv("BENCH_DUR", "5s")
	for b.Loop() {
		Duration("BENCH_DUR", time.Second)
	}
}

func BenchmarkDuration_Default(b *testing.B) {
	for b.Loop() {
		Duration("BENCH_DUR_MISS", time.Second)
	}
}

func BenchmarkStringSlice(b *testing.B) {
	b.Setenv("BENCH_SLICE", "a,b,c,d,e")
	for b.Loop() {
		StringSlice("BENCH_SLICE", nil)
	}
}

func BenchmarkStringRequired(b *testing.B) {
	b.Setenv("BENCH_REQ", "value")
	for b.Loop() {
		StringRequired("BENCH_REQ")
	}
}

func BenchmarkValidate(b *testing.B) {
	b.Setenv("V1", "a")
	b.Setenv("V2", "b")
	b.Setenv("V3", "c")
	for b.Loop() {
		Validate("V1", "V2", "V3")
	}
}

func BenchmarkMustString(b *testing.B) {
	b.Setenv("BENCH_MUST", "value")
	for b.Loop() {
		MustString("BENCH_MUST")
	}
}
