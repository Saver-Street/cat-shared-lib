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

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

// ---- HTTPTestServer tests ----

func TestHTTPTestServer_BasicRoute(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	resp, err := http.Get(s.URL + "/ping") //nolint:noctx
	testkit.RequireNoError(t, err)
	defer resp.Body.Close()

	testkit.AssertEqual(t, resp.StatusCode, http.StatusOK)
	body, _ := io.ReadAll(resp.Body)
	testkit.AssertEqual(t, string(body), "pong")
}

func TestHTTPTestServer_MethodNotAllowed(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("GET", "/only-get", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	resp, err := http.Post(s.URL+"/only-get", "application/json", nil) //nolint:noctx
	testkit.RequireNoError(t, err)
	defer resp.Body.Close()

	testkit.AssertEqual(t, resp.StatusCode, http.StatusMethodNotAllowed)
}

func TestHTTPTestServer_NotFound(t *testing.T) {
	s := NewHTTPTestServer(t)

	resp, err := http.Get(s.URL + "/nonexistent") //nolint:noctx
	testkit.RequireNoError(t, err)
	defer resp.Body.Close()

	testkit.AssertEqual(t, resp.StatusCode, http.StatusNotFound)
}

func TestHTTPTestServer_Fallback(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.SetFallback(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	resp, err := http.Get(s.URL + "/anything") //nolint:noctx
	testkit.RequireNoError(t, err)
	defer resp.Body.Close()

	testkit.AssertEqual(t, resp.StatusCode, http.StatusTeapot)
}

func TestHTTPTestServer_WildcardMethod(t *testing.T) {
	s := NewHTTPTestServer(t)
	s.Handle("*", "/any", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	for _, method := range []string{"GET", "POST", "DELETE"} {
		req, _ := http.NewRequestWithContext(context.Background(), method, s.URL+"/any", nil)
		resp, err := http.DefaultClient.Do(req)
		testkit.RequireNoError(t, err)
		resp.Body.Close()
		testkit.AssertEqual(t, resp.StatusCode, http.StatusAccepted)
	}
}

func TestHTTPTestServer_HandleJSON(t *testing.T) {
	s := NewHTTPTestServer(t)
	payload := map[string]string{"key": "value"}
	s.HandleJSON("GET", "/json", http.StatusOK, payload)

	resp, err := http.Get(s.URL + "/json") //nolint:noctx
	testkit.RequireNoError(t, err)
	defer resp.Body.Close()

	testkit.AssertEqual(t, resp.Header.Get("Content-Type"), "application/json")

	var got map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	testkit.AssertEqual(t, got["key"], "value")
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
	testkit.RequireNoError(t, err)
	resp.Body.Close()

	if s.RequestCount() != 1 {
		t.Fatalf("RequestCount = %d, want 1", s.RequestCount())
	}
	last := s.LastRequest()
	if last == nil {
		t.Fatal("LastRequest is nil")
	}
	testkit.AssertEqual(t, last.Method, "POST")
	testkit.AssertEqual(t, last.Path, "/data")
	testkit.AssertEqual(t, string(last.Body), `{"x":1}`)
}

func TestHTTPTestServer_LastRequest_Nil(t *testing.T) {
	s := NewHTTPTestServer(t)
	testkit.AssertNil(t, s.LastRequest())
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
	testkit.AssertLen(t, reqs, 1)
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

	testkit.AssertEqual(t, s.RequestCount(), 0)

	// Route should be gone after reset.
	resp2, err := http.Get(s.URL + "/before") //nolint:noctx
	testkit.RequireNoError(t, err)
	resp2.Body.Close()
	testkit.AssertEqual(t, resp2.StatusCode, http.StatusNotFound)
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

	testkit.AssertEqual(t, s.RequestCount(), n)
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
	testkit.AssertEqual(t, id, "candidate-123")
}

func TestDBTestHelper_QueryRow_RecordsQuery(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"x"}})

	_ = db.QueryRow(context.Background(), "SELECT 1", "arg1")
	queries := db.Queries()
	testkit.RequireLen(t, queries, 1)
	testkit.AssertEqual(t, queries[0].SQL, "SELECT 1")
}

func TestDBTestHelper_QueryRow_NoRowsQueued(t *testing.T) {
	db := &DBTestHelper{}

	row := db.QueryRow(context.Background(), "SELECT 1")
	var v string
	err := row.Scan(&v)
	testkit.AssertError(t, err)
}

func TestDBTestHelper_QueryRow_ScanError(t *testing.T) {
	db := &DBTestHelper{}
	scanErr := errors.New("scan failed")
	db.QueueRow(&MockRow{Err: scanErr})

	row := db.QueryRow(context.Background(), "SELECT 1")
	var v string
	err := row.Scan(&v)
	testkit.AssertErrorIs(t, err, scanErr)
}

func TestDBTestHelper_QueryCount(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"a"}})
	db.QueueRow(&MockRow{ScanValues: []any{"b"}})

	_ = db.QueryRow(context.Background(), "Q1")
	_ = db.QueryRow(context.Background(), "Q2")

	testkit.AssertEqual(t, db.QueryCount(), 2)
}

func TestDBTestHelper_Reset(t *testing.T) {
	db := &DBTestHelper{}
	db.QueueRow(&MockRow{ScanValues: []any{"x"}})
	_ = db.QueryRow(context.Background(), "SELECT 1")
	db.Reset()

	testkit.AssertEqual(t, db.QueryCount(), 0)
}

func TestMockRow_ScanInt(t *testing.T) {
	row := &MockRow{ScanValues: []any{42}}
	var v int
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	testkit.AssertEqual(t, v, 42)
}

func TestMockRow_ScanInt64(t *testing.T) {
	row := &MockRow{ScanValues: []any{int64(99)}}
	var v int64
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	testkit.AssertEqual(t, v, int64(99))
}

func TestMockRow_ScanInt64_FromInt(t *testing.T) {
	row := &MockRow{ScanValues: []any{int(7)}}
	var v int64
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	testkit.AssertEqual(t, v, int64(7))
}

func TestMockRow_ScanBool(t *testing.T) {
	row := &MockRow{ScanValues: []any{true}}
	var v bool
	if err := row.Scan(&v); err != nil {
		t.Fatal(err)
	}
	testkit.AssertTrue(t, v)
}

func TestMockRow_ScanTypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{"not-an-int"}}
	var v int
	err := row.Scan(&v)
	testkit.AssertError(t, err)
}

func TestMockRow_ScanTooFewValues(t *testing.T) {
	row := &MockRow{ScanValues: []any{"one"}}
	var a, b string
	err := row.Scan(&a, &b)
	testkit.AssertError(t, err)
}

func TestMockRow_ScanUnsupportedType(t *testing.T) {
	row := &MockRow{ScanValues: []any{[]byte("blob")}}
	var v []byte
	err := row.Scan(&v)
	testkit.AssertError(t, err)
}

func TestMockRow_ScanStringTypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{123}}
	var v string
	err := row.Scan(&v)
	testkit.AssertError(t, err)
}

func TestMockRow_ScanBoolTypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{"not-bool"}}
	var v bool
	err := row.Scan(&v)
	testkit.AssertError(t, err)
}

func TestMockRow_ScanInt64TypeMismatch(t *testing.T) {
	row := &MockRow{ScanValues: []any{"not-int64"}}
	var v int64
	err := row.Scan(&v)
	testkit.AssertError(t, err)
}

// ---- Fixtures tests ----

func TestFixtures_RegisterAndLoad(t *testing.T) {
	f := NewFixtures()
	f.Register("user", []byte(`{"id":"u1","name":"Alice"}`))

	b, err := f.Load("user")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, string(b), `{"id":"u1","name":"Alice"}`)
}

func TestFixtures_Load_NotFound(t *testing.T) {
	f := NewFixtures()
	_, err := f.Load("missing")
	testkit.AssertError(t, err)
}

func TestFixtures_MustLoad(t *testing.T) {
	f := NewFixtures()
	f.Register("ping", []byte(`"pong"`))
	b := f.MustLoad("ping")
	testkit.AssertEqual(t, string(b), `"pong"`)
}

func TestFixtures_MustLoad_Panics(t *testing.T) {
	f := NewFixtures()
	testkit.AssertPanics(t, func() {
		f.MustLoad("nonexistent")
	})
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
	testkit.AssertEqual(t, got.ID, "1")
	testkit.AssertEqual(t, got.Name, "Widget")
}

func TestFixtures_RegisterJSON_Error(t *testing.T) {
	f := NewFixtures()
	err := f.RegisterJSON("bad", make(chan int))
	testkit.AssertError(t, err)
}

func TestFixtures_LoadInto_NotFound(t *testing.T) {
	f := NewFixtures()
	var v any
	err := f.LoadInto("missing", &v)
	testkit.AssertError(t, err)
}

func TestFixtures_LoadInto_BadJSON(t *testing.T) {
	f := NewFixtures()
	f.Register("bad", []byte(`not json`))
	var v map[string]any
	err := f.LoadInto("bad", &v)
	testkit.AssertError(t, err)
}

func TestFixtures_Names(t *testing.T) {
	f := NewFixtures()
	f.Register("a", []byte(`1`))
	f.Register("b", []byte(`2`))

	names := f.Names()
	testkit.AssertLen(t, names, 2)
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
