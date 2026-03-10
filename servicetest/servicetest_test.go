package servicetest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// ---- HTTPTestServer tests ----

func TestHTTPTestServer_BasicRoute(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	resp, err := http.Get(s.URL + "/ping") //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Errorf("body = %q, want pong", body)
	}
}

func TestHTTPTestServer_MethodNotAllowed(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/only-get", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	resp, err := http.Post(s.URL+"/only-get", "application/json", nil) //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestHTTPTestServer_NotFound(t *testing.T) {
	s := NewHTTPTestServer(t)

	resp, err := http.Get(s.URL + "/nonexistent") //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHTTPTestServer_Fallback(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.SetFallback(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	resp, err := http.Get(s.URL + "/anything") //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("status = %d, want 418", resp.StatusCode)
	}
}

func TestHTTPTestServer_WildcardMethod(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("*", "/any", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	for _, method := range []string{"GET", "POST", "DELETE"} {
		req, _ := http.NewRequestWithContext(context.Background(), method, s.URL+"/any", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("method=%s: status = %d, want 202", method, resp.StatusCode)
		}
	}
}

func TestHTTPTestServer_HandleJSON(t *testing.T) {
	s := NewHTTPTestServer(t)
	payload := map[string]string{"key": "value"}
	s.HandleJSON("GET", "/json", http.StatusOK, payload)

	resp, err := http.Get(s.URL + "/json") //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["key"] != "value" {
		t.Errorf("body key = %q, want value", got["key"])
	}
}

func TestHTTPTestServer_RecordsRequests(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("POST", "/data", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	body := strings.NewReader(`{"x":1}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, s.URL+"/data", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if s.RequestCount() != 1 {
		t.Fatalf("RequestCount = %d, want 1", s.RequestCount())
	}
	last := s.LastRequest()
	if last == nil {
		t.Fatal("LastRequest is nil")
	}
	if last.Method != "POST" {
		t.Errorf("Method = %q, want POST", last.Method)
	}
	if last.Path != "/data" {
		t.Errorf("Path = %q, want /data", last.Path)
	}
	if string(last.Body) != `{"x":1}` {
		t.Errorf("Body = %q, want {\"x\":1}", last.Body)
	}
}

func TestHTTPTestServer_LastRequest_Nil(t *testing.T) {
	s := NewHTTPTestServer(t)
	if s.LastRequest() != nil {
		t.Error("expected nil before any requests")
	}
}

func TestHTTPTestServer_Requests_Copy(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/a", func(w http.ResponseWriter, _ *http.Request) {})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, s.URL+"/a", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	reqs := s.Requests()
	if len(reqs) != 1 {
		t.Errorf("expected 1 request, got %d", len(reqs))
	}
}

func TestHTTPTestServer_Reset(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/before", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, s.URL+"/before", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	s.Reset()

	if s.RequestCount() != 0 {
		t.Errorf("expected 0 requests after reset, got %d", s.RequestCount())
	}

	// Route should be gone after reset.
	resp2, err := http.Get(s.URL + "/before") //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after reset, got %d", resp2.StatusCode)
	}
}

func TestHTTPTestServer_ConcurrentRequests(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/concurrent", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	const n = 20
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, s.URL+"/concurrent", nil)
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		}()
	}
	wg.Wait()

	if s.RequestCount() != n {
		t.Errorf("RequestCount = %d, want %d", s.RequestCount(), n)
	}
}

// ---- DBTestHelper tests ----

func TestDBTestHelper_QueryRow_Success(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"candidate-123"}})

	row := db.QueryRow(context.Background(), "SELECT id FROM candidate_profiles WHERE user_id = $1", "user-456")
	var id string
	if err := row.Scan(&id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "candidate-123" {
		t.Errorf("id = %q, want candidate-123", id)
	}
}

func TestDBTestHelper_QueryRow_RecordsQuery(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"x"}})

	_ = db.QueryRow(context.Background(), "SELECT 1", "arg1")
	queries := db.Queries()
	if len(queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(queries))
	}
	if queries[0].SQL != "SELECT 1" {
		t.Errorf("SQL = %q, want SELECT 1", queries[0].SQL)
	}
}

func TestDBTestHelper_QueryRow_NoRowsQueued(t *testing.T) {
	db := &DBTestHelper{}

	row := db.QueryRow(context.Background(), "SELECT 1")
	var v string
	err := row.Scan(&v)
	if err == nil {
		t.Error("expected error when no rows queued")
	}
}

func TestDBTestHelper_QueryRow_ScanError(t *testing.T) {
	db := &DBTestHelper{}
	scanErr := errors.New("scan failed")
	db.QueueRow(&MockRow{Err: scanErr})

	row := db.QueryRow(context.Background(), "SELECT 1")
	var v string
	err := row.Scan(&v)
	if !errors.Is(err, scanErr) {
		t.Errorf("expected scan error, got %v", err)
	}
}

func TestDBTestHelper_QueryCount(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"a"}})
	db.QueueRow(&MockRow{ScanValues: []any{"b"}})

	_ = db.QueryRow(context.Background(), "Q1")
	_ = db.QueryRow(context.Background(), "Q2")

	if db.QueryCount() != 2 {
		t.Errorf("QueryCount = %d, want 2", db.QueryCount())
	}
}

func TestDBTestHelper_Reset(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"x"}})
	_ = db.QueryRow(context.Background(), "SELECT 1")
	db.Reset()

	if db.QueryCount() != 0 {
		t.Errorf("expected 0 queries after reset, got %d", db.QueryCount())
	}
}

func TestMockRow_ScanInt(t *testing.T) {
	row := &MockRow{ScanValues: []any{42}}
	var v int
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("v = %d, want 42", v)
	}
}

func TestMockRow_ScanInt64(t *testing.T) {
	row := &MockRow{ScanValues: []any{int64(99)}}
	var v int64
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != 99 {
		t.Errorf("v = %d, want 99", v)
	}
}

func TestMockRow_ScanInt64_FromInt(t *testing.T) {
	row := &MockRow{ScanValues: []any{int(7)}}
	var v int64
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != 7 {
		t.Errorf("v = %d, want 7", v)
	}
}

func TestMockRow_ScanBool(t *testing.T) {
	row := &MockRow{ScanValues: []any{true}}
	var v bool
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Error("expected true")
	}
}

func TestMockRow_ScanTypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{"not-an-int"}}
	var v int
	if err := row.Scan(&v); err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestMockRow_ScanTooFewValues(t *testing.T) {
	row := &MockRow{ScanValues: []any{"one"}}
	var a, b string
	if err := row.Scan(&a, &b); err == nil {
		t.Error("expected error when too few values")
	}
}

func TestMockRow_ScanUnsupportedType(t *testing.T) {
	row := &MockRow{ScanValues: []any{[]byte("blob")}}
	var v []byte
	if err := row.Scan(&v); err == nil {
		t.Error("expected error for unsupported dest type")
	}
}

func TestMockRow_ScanStringTypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{123}}
	var v string
	if err := row.Scan(&v); err == nil {
		t.Error("expected error for type mismatch string")
	}
}

func TestMockRow_ScanBoolTypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{"not-bool"}}
	var v bool
	if err := row.Scan(&v); err == nil {
		t.Error("expected error for type mismatch bool")
	}
}

func TestMockRow_ScanInt64TypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{"not-int64"}}
	var v int64
	if err := row.Scan(&v); err == nil {
		t.Error("expected error for type mismatch int64")
	}
}

// ---- Fixtures tests ----

func TestFixtures_RegisterAndLoad(t *testing.T) {
	f := NewFixtures()
	f.Register("user", []byte(`{"id":"u1","name":"Alice"}`))

	b, err := f.Load("user")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"id":"u1","name":"Alice"}` {
		t.Errorf("loaded = %q", b)
	}
}

func TestFixtures_Load_NotFound(t *testing.T) {
	f := NewFixtures()
	_, err := f.Load("missing")
	if err == nil {
		t.Error("expected error for missing fixture")
	}
}

func TestFixtures_MustLoad(t *testing.T) {
	f := NewFixtures()
	f.Register("ping", []byte(`"pong"`))
	b := f.MustLoad("ping")
	if string(b) != `"pong"` {
		t.Errorf("got %q", b)
	}
}

func TestFixtures_MustLoad_Panics(t *testing.T) {
	f := NewFixtures()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing fixture")
		}
	}()
	f.MustLoad("nonexistent")
}

func TestFixtures_RegisterJSON(t *testing.T) {
	f := NewFixtures()
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := f.RegisterJSON("item", Item{ID: "1", Name: "Widget"}); err != nil {
		t.Fatal(err)
	}

	var got Item
	if err := f.LoadInto("item", &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "1" || got.Name != "Widget" {
		t.Errorf("got %+v", got)
	}
}

func TestFixtures_RegisterJSON_Error(t *testing.T) {
	f := NewFixtures()
	if err := f.RegisterJSON("bad", make(chan int)); err == nil {
		t.Error("expected error for un-marshalable value")
	}
}

func TestFixtures_LoadInto_NotFound(t *testing.T) {
	f := NewFixtures()
	var v any
	if err := f.LoadInto("missing", &v); err == nil {
		t.Error("expected error for missing fixture")
	}
}

func TestFixtures_LoadInto_BadJSON(t *testing.T) {
	f := NewFixtures()
	f.Register("bad", []byte(`not json`))
	var v map[string]any
	if err := f.LoadInto("bad", &v); err == nil {
		t.Error("expected error for bad JSON")
	}
}

func TestFixtures_Names(t *testing.T) {
	f := NewFixtures()
	f.Register("a", []byte(`1`))
	f.Register("b", []byte(`2`))

	names := f.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestFixtures_ConcurrentAccess(t *testing.T) {
	f := NewFixtures()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			if err := f.RegisterJSON(key, map[string]int{"n": i}); err != nil {
				t.Errorf("RegisterJSON: %v", err)
			}
			_, _ = f.Load(key)
		}(i)
	}
	wg.Wait()
}
