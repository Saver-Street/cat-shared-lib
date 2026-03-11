package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testPayload struct {
	XMLName xml.Name `xml:"item" json:"-"`
	Name    string   `json:"name" xml:"name"`
	Value   int      `json:"value" xml:"value"`
}

func TestNegotiateJSON(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")

	data := testPayload{Name: "test", Value: 42}
	Negotiate(rec, req, http.StatusOK, data)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
	var got testPayload
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "test" || got.Value != 42 {
		t.Errorf("got = %+v; want {test 42}", got)
	}
}

func TestNegotiateXML(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/xml")

	data := testPayload{Name: "test", Value: 42}
	Negotiate(rec, req, http.StatusOK, data)

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/xml") {
		t.Errorf("Content-Type = %q; want application/xml", ct)
	}
	var got testPayload
	if err := xml.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "test" || got.Value != 42 {
		t.Errorf("got = %+v; want {test 42}", got)
	}
}

func TestNegotiateTextXML(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/xml")

	Negotiate(rec, req, http.StatusOK, testPayload{Name: "a", Value: 1})

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/xml") {
		t.Errorf("Content-Type = %q; want application/xml", ct)
	}
}

func TestNegotiateNoAccept(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Negotiate(rec, req, http.StatusOK, testPayload{Name: "b", Value: 2})

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json (default)", ct)
	}
}

func TestNegotiateWildcard(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "*/*")

	Negotiate(rec, req, http.StatusOK, testPayload{Name: "c", Value: 3})

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
}

func TestNegotiateMultipleAccept(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html, application/xml;q=0.9")

	Negotiate(rec, req, http.StatusCreated, testPayload{Name: "d", Value: 4})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d; want 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/xml") {
		t.Errorf("Content-Type = %q; want application/xml", ct)
	}
}

func TestNegotiateOK(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	NegotiateOK(rec, req, testPayload{Name: "e", Value: 5})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", rec.Code)
	}
}

func BenchmarkNegotiateJSON(b *testing.B) {
	data := testPayload{Name: "bench", Value: 99}
	for range b.N {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		Negotiate(rec, req, http.StatusOK, data)
	}
}

func BenchmarkNegotiateXML(b *testing.B) {
	data := testPayload{Name: "bench", Value: 99}
	for range b.N {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", "application/xml")
		Negotiate(rec, req, http.StatusOK, data)
	}
}

func FuzzNegotiate(f *testing.F) {
	f.Add("application/json")
	f.Add("application/xml")
	f.Add("text/xml")
	f.Add("*/*")
	f.Add("")
	f.Fuzz(func(t *testing.T, accept string) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", accept)
		Negotiate(rec, req, http.StatusOK, testPayload{Name: "fuzz"})
	})
}
