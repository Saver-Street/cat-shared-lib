package jsonutil

import (
"bytes"
"encoding/json"
"fmt"
"reflect"
)

// Pretty formats JSON data with indentation for human readability.
func Pretty(data []byte) ([]byte, error) {
var buf bytes.Buffer
if err := json.Indent(&buf, data, "", "  "); err != nil {
return nil, fmt.Errorf("jsonutil: indent: %w", err)
}
return buf.Bytes(), nil
}

// Compact removes insignificant whitespace from JSON data.
func Compact(data []byte) ([]byte, error) {
var buf bytes.Buffer
if err := json.Compact(&buf, data); err != nil {
return nil, fmt.Errorf("jsonutil: compact: %w", err)
}
return buf.Bytes(), nil
}

// Valid reports whether data is valid JSON.
func Valid(data []byte) bool {
return json.Valid(data)
}

// MustMarshal marshals v to JSON, panicking on error.
func MustMarshal(v any) []byte {
data, err := json.Marshal(v)
if err != nil {
panic(fmt.Sprintf("jsonutil: marshal: %v", err))
}
return data
}

// MustUnmarshal unmarshals data into v, panicking on error.
func MustUnmarshal(data []byte, v any) {
if err := json.Unmarshal(data, v); err != nil {
panic(fmt.Sprintf("jsonutil: unmarshal: %v", err))
}
}

// Map unmarshals JSON data into a map[string]any.
func Map(data []byte) (map[string]any, error) {
var m map[string]any
if err := json.Unmarshal(data, &m); err != nil {
return nil, fmt.Errorf("jsonutil: unmarshal map: %w", err)
}
return m, nil
}

// Merge deep-merges two JSON objects. Keys in b override keys in a.
// Both a and b must be valid JSON objects.
func Merge(a, b []byte) ([]byte, error) {
var ma, mb map[string]any
if err := json.Unmarshal(a, &ma); err != nil {
return nil, fmt.Errorf("jsonutil: unmarshal a: %w", err)
}
if err := json.Unmarshal(b, &mb); err != nil {
return nil, fmt.Errorf("jsonutil: unmarshal b: %w", err)
}
deepMerge(ma, mb)
return json.Marshal(ma)
}

func deepMerge(dst, src map[string]any) {
for k, sv := range src {
if dv, ok := dst[k]; ok {
dm, dOk := dv.(map[string]any)
sm, sOk := sv.(map[string]any)
if dOk && sOk {
deepMerge(dm, sm)
continue
}
}
dst[k] = sv
}
}

// GetPath extracts a value from JSON data using dot-notation path (e.g., "user.name").
// Returns nil if the path does not exist.
func GetPath(data []byte, path string) (any, error) {
var m any
if err := json.Unmarshal(data, &m); err != nil {
return nil, fmt.Errorf("jsonutil: unmarshal: %w", err)
}
parts := splitPath(path)
current := m
for _, part := range parts {
obj, ok := current.(map[string]any)
if !ok {
return nil, nil
}
current, ok = obj[part]
if !ok {
return nil, nil
}
}
return current, nil
}

func splitPath(path string) []string {
if path == "" {
return nil
}
var parts []string
start := 0
for i := 0; i < len(path); i++ {
if path[i] == '.' {
parts = append(parts, path[start:i])
start = i + 1
}
}
parts = append(parts, path[start:])
return parts
}

// Equal reports whether two JSON values are semantically equal,
// ignoring whitespace differences.
func Equal(a, b []byte) (bool, error) {
var va, vb any
if err := json.Unmarshal(a, &va); err != nil {
return false, fmt.Errorf("jsonutil: unmarshal a: %w", err)
}
if err := json.Unmarshal(b, &vb); err != nil {
return false, fmt.Errorf("jsonutil: unmarshal b: %w", err)
}
return reflect.DeepEqual(va, vb), nil
}
