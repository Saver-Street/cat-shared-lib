package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_DefaultJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf))
	logger.Info("hello", "key", "val")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if entry["msg"] != "hello" {
		t.Errorf("got msg=%v, want hello", entry["msg"])
	}
	if entry["key"] != "val" {
		t.Errorf("got key=%v, want val", entry["key"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf), WithFormat("text"))
	logger.Info("ping")

	out := buf.String()
	if !strings.Contains(out, "ping") {
		t.Errorf("expected text output containing 'ping', got: %s", out)
	}
	// Text format should not be valid JSON
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err == nil {
		t.Error("expected non-JSON output for text format")
	}
}

func TestNew_WithLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf), WithLevel(slog.LevelWarn))
	logger.Info("should be filtered")
	if buf.Len() != 0 {
		t.Error("expected info message to be filtered at warn level")
	}
	logger.Warn("should appear")
	if buf.Len() == 0 {
		t.Error("expected warn message to appear at warn level")
	}
}

func TestNew_WithSource(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf), WithSource())
	logger.Info("src")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if _, ok := entry["source"]; !ok {
		t.Error("expected source field in output")
	}
}

func TestNew_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger := New(
		WithWriter(&buf),
		WithAttrs(Service("test-svc"), Version("1.0")),
	)
	logger.Info("boot")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if entry["service"] != "test-svc" {
		t.Errorf("got service=%v, want test-svc", entry["service"])
	}
	if entry["version"] != "1.0" {
		t.Errorf("got version=%v, want 1.0", entry["version"])
	}
}

func TestNew_WithReplaceAttr(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf), WithReplaceAttr(func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "time" {
			return slog.Attr{}
		}
		return a
	}))
	logger.Info("stripped")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if _, ok := entry["time"]; ok {
		t.Error("expected time field to be stripped")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"  debug  ", slog.LevelDebug},
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestWithContext_FromContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf))

	ctx := WithContext(context.Background(), logger)
	got := FromContext(ctx)
	if got != logger {
		t.Error("expected same logger from context")
	}

	got.Info("from-ctx")
	if buf.Len() == 0 {
		t.Error("expected logger from context to write output")
	}
}

func TestFromContext_Default(t *testing.T) {
	got := FromContext(context.Background())
	if got != slog.Default() {
		t.Error("expected slog.Default() when no logger in context")
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf))
	ctx := WithContext(context.Background(), logger)

	child := With(ctx, "extra", "data")
	child.Info("enriched")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if entry["extra"] != "data" {
		t.Errorf("got extra=%v, want data", entry["extra"])
	}
}

func TestAttributeHelpers(t *testing.T) {
	tests := []struct {
		name string
		attr slog.Attr
		key  string
		val  string
	}{
		{"RequestID", RequestID("abc"), "request_id", "abc"},
		{"UserID", UserID("u1"), "user_id", "u1"},
		{"Service", Service("svc"), "service", "svc"},
		{"Version", Version("v2"), "version", "v2"},
		{"Component", Component("auth"), "component", "auth"},
		{"TraceID", TraceID("t1"), "trace_id", "t1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attr.Key != tt.key {
				t.Errorf("key = %s, want %s", tt.attr.Key, tt.key)
			}
			if tt.attr.Value.String() != tt.val {
				t.Errorf("value = %s, want %s", tt.attr.Value.String(), tt.val)
			}
		})
	}
}

func TestErr_NilError(t *testing.T) {
	attr := Err(nil)
	if attr.Key != "" {
		t.Errorf("expected empty attr for nil error, got key=%s", attr.Key)
	}
}

func TestErr_WithError(t *testing.T) {
	attr := Err(context.DeadlineExceeded)
	if attr.Key != "error" {
		t.Errorf("key = %s, want error", attr.Key)
	}
	if attr.Value.String() != "context deadline exceeded" {
		t.Errorf("value = %s, want context deadline exceeded", attr.Value.String())
	}
}
