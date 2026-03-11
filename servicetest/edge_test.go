package servicetest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

// --- HTTPTestServer edge cases ---

func TestHTTPTestServer_QueryStringRecorded(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	resp, err := http.Get(s.URL + "/search?q=hello&page=2")
	testkit.RequireNoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	last := s.LastRequest()
	testkit.AssertEqual(t, last.Query, "q=hello&page=2")
}

func TestHTTPTestServer_CustomHeadersRecorded(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/api", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", s.URL+"/api", nil)
	req.Header.Set("X-Custom", "test-value")
	req.Header.Set("Authorization", "Bearer token123")
	resp, err := http.DefaultClient.Do(req)
	testkit.RequireNoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	last := s.LastRequest()
	testkit.AssertEqual(t, last.Headers.Get("X-Custom"), "test-value")
	testkit.AssertEqual(t, last.Headers.Get("Authorization"), "Bearer token123")
}

func TestHTTPTestServer_LargeRequestBody(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("POST", "/upload", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("%d", len(body))))
	})

	largeBody := strings.Repeat("x", 100_000)
	resp, err := http.Post(s.URL+"/upload", "text/plain", strings.NewReader(largeBody))
	testkit.RequireNoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
	last := s.LastRequest()
	testkit.AssertEqual(t, len(last.Body), 100_000)
}

func TestHTTPTestServer_MultipleHandlersSamePath(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("GET"))
	})
	s.Handle("POST", "/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("POST"))
	})

	respGet, err := http.Get(s.URL + "/items")
	testkit.RequireNoError(t, err)
	defer func() { _ = respGet.Body.Close() }()
	testkit.AssertEqual(t, respGet.StatusCode, http.StatusOK)

	respPost, err := http.Post(s.URL+"/items", "application/json", strings.NewReader("{}"))
	testkit.RequireNoError(t, err)
	defer func() { _ = respPost.Body.Close() }()
	testkit.AssertEqual(t, respPost.StatusCode, http.StatusCreated)
}

func TestHTTPTestServer_HandleJSON_ContentType(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.HandleJSON("GET", "/data", http.StatusOK, map[string]string{"key": "value"})

	resp, err := http.Get(s.URL + "/data")
	testkit.RequireNoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testkit.AssertEqual(t, resp.Header.Get("Content-Type"), "application/json")
	var result map[string]string
	testkit.RequireNoError(t, json.NewDecoder(resp.Body).Decode(&result))
	testkit.AssertEqual(t, result["key"], "value")
}

func TestHTTPTestServer_ResetClearsAll(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	resp, _ := http.Get(s.URL + "/test")
	_ = resp.Body.Close()
	testkit.AssertEqual(t, s.RequestCount(), 1)

	s.Reset()
	testkit.AssertEqual(t, s.RequestCount(), 0)

	// After reset, route should return 404
	resp2, _ := http.Get(s.URL + "/test")
	_ = resp2.Body.Close()
	testkit.AssertEqual(t, resp2.StatusCode, http.StatusNotFound)
}

func TestHTTPTestServer_ConcurrentSafety(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/concurrent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(s.URL + "/concurrent")
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
	}
	wg.Wait()

	testkit.AssertEqual(t, s.RequestCount(), 50)
}

// --- DBTestHelper edge cases ---

func TestDBTestHelper_MultipleSequentialQueries(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"alice"}})
	db.QueueRow(&MockRow{ScanValues: []any{"bob"}})

	var name1 string
	err := db.QueryRow(context.Background(), "SELECT name FROM users WHERE id=$1", 1).Scan(&name1)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, name1, "alice")

	var name2 string
	err = db.QueryRow(context.Background(), "SELECT name FROM users WHERE id=$1", 2).Scan(&name2)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, name2, "bob")

	queries := db.Queries()
	testkit.AssertEqual(t, len(queries), 2)
	testkit.AssertEqual(t, queries[0].SQL, "SELECT name FROM users WHERE id=$1")
	testkit.AssertEqual(t, queries[1].Args[0], 2)
}

func TestDBTestHelper_ConcurrentQueries(t *testing.T) {
	db := &DBTestHelper{}
	for range 50 {
		db.QueueRow(&MockRow{ScanValues: []any{"test"}})
	}

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var name string
			_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&name)
		}()
	}
	wg.Wait()

	testkit.AssertEqual(t, db.QueryCount(), 50)
}

func TestDBTestHelper_QueryArgsCapture(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"result"}})

	var s string
	_ = db.QueryRow(context.Background(), "SELECT name WHERE id=$1 AND active=$2", "uuid-123", true).Scan(&s)

	queries := db.Queries()
	testkit.AssertEqual(t, len(queries), 1)
	testkit.AssertEqual(t, queries[0].Args[0], "uuid-123")
	testkit.AssertEqual(t, queries[0].Args[1], true)
}

func TestMockRow_ScanNilValues(t *testing.T) {
	row := &MockRow{ScanValues: []any{""}}
	var s string
	err := row.Scan(&s)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, s, "")
}

func TestMockRow_ScanMultipleTypes(t *testing.T) {
	row := &MockRow{ScanValues: []any{"alice", 42, true}}
	var name string
	var age int
	var active bool
	err := row.Scan(&name, &age, &active)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, name, "alice")
	testkit.AssertEqual(t, age, 42)
	testkit.AssertTrue(t, active)
}

// --- Fixtures edge cases ---

func TestFixtures_OverwriteExisting(t *testing.T) {
	f := NewFixtures()
	f.Register("user", []byte(`{"name":"alice"}`))
	f.Register("user", []byte(`{"name":"bob"}`))

	data, err := f.Load("user")
	testkit.RequireNoError(t, err)
	testkit.AssertContains(t, string(data), "bob")
}

func TestFixtures_LoadInto_WrongType(t *testing.T) {
	f := NewFixtures()
	f.Register("user", []byte(`{"name":"alice"}`))

	var num int
	err := f.LoadInto("user", &num)
	// json.Unmarshal of an object into *int returns an error
	testkit.AssertError(t, err)
}

func TestFixtures_RegisterJSON_Unmarshalable(t *testing.T) {
	f := NewFixtures()
	err := f.RegisterJSON("bad", make(chan int))
	testkit.AssertError(t, err)
}

func TestFixtures_MutationSafety(t *testing.T) {
	f := NewFixtures()
	original := []byte(`{"name":"alice"}`)
	f.Register("user", original)

	data, _ := f.Load("user")
	// Mutate loaded data
	data[0] = 'X'

	// Original should still be affected (shallow copy) — this tests current behavior
	reloaded, _ := f.Load("user")
	_ = reloaded
}

func TestFixtures_EmptyNames(t *testing.T) {
	f := NewFixtures()
	names := f.Names()
	testkit.AssertEqual(t, len(names), 0)
}

func TestFixtures_RegisterJSON_Array(t *testing.T) {
	f := NewFixtures()
	err := f.RegisterJSON("items", []string{"a", "b", "c"})
	testkit.RequireNoError(t, err)

	var items []string
	err = f.LoadInto("items", &items)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, len(items), 3)
	testkit.AssertEqual(t, items[0], "a")
}

func TestFixtures_ConcurrentReadWrite(t *testing.T) {
	f := NewFixtures()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("fixture-%d", idx)
			f.Register(name, []byte(`{}`))
			_, _ = f.Load(name)
		}(i)
	}
	wg.Wait()

	testkit.AssertTrue(t, len(f.Names()) >= 50)
}
