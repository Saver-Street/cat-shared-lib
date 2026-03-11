package httputil_test

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/httputil"
)

func newRequest(method, url string, headers map[string]string) *http.Request {
	r := httptest.NewRequest(method, url, nil)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		name   string
		ct     string
		want   bool
	}{
		{"json", "application/json", true},
		{"json with charset", "application/json; charset=utf-8", true},
		{"text", "text/plain", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRequest("GET", "/", map[string]string{"Content-Type": tt.ct})
			if got := httputil.IsJSON(r); got != tt.want {
				t.Errorf("IsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsForm(t *testing.T) {
	tests := []struct {
		name string
		ct   string
		want bool
	}{
		{"form", "application/x-www-form-urlencoded", true},
		{"json", "application/json", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRequest("GET", "/", map[string]string{"Content-Type": tt.ct})
			if got := httputil.IsForm(r); got != tt.want {
				t.Errorf("IsForm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMultipart(t *testing.T) {
	tests := []struct {
		name string
		ct   string
		want bool
	}{
		{"multipart form", "multipart/form-data", true},
		{"multipart mixed", "multipart/mixed", true},
		{"json", "application/json", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRequest("GET", "/", map[string]string{"Content-Type": tt.ct})
			if got := httputil.IsMultipart(r); got != tt.want {
				t.Errorf("IsMultipart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAJAX(t *testing.T) {
	t.Run("with header", func(t *testing.T) {
		r := newRequest("GET", "/", map[string]string{"X-Requested-With": "XMLHttpRequest"})
		if !httputil.IsAJAX(r) {
			t.Error("expected IsAJAX to return true")
		}
	})
	t.Run("without header", func(t *testing.T) {
		r := newRequest("GET", "/", nil)
		if httputil.IsAJAX(r) {
			t.Error("expected IsAJAX to return false")
		}
	})
}

func TestIsTLS(t *testing.T) {
	t.Run("with TLS", func(t *testing.T) {
		r := newRequest("GET", "/", nil)
		r.TLS = &tls.ConnectionState{}
		if !httputil.IsTLS(r) {
			t.Error("expected IsTLS to return true")
		}
	})
	t.Run("with X-Forwarded-Proto", func(t *testing.T) {
		r := newRequest("GET", "/", map[string]string{"X-Forwarded-Proto": "https"})
		if !httputil.IsTLS(r) {
			t.Error("expected IsTLS to return true")
		}
	})
	t.Run("plain HTTP", func(t *testing.T) {
		r := newRequest("GET", "/", nil)
		if httputil.IsTLS(r) {
			t.Error("expected IsTLS to return false")
		}
	})
}

func TestAccepts(t *testing.T) {
	tests := []struct {
		name        string
		accept      string
		contentType string
		want        bool
	}{
		{"wildcard", "*/*", "application/json", true},
		{"empty accept", "", "text/html", true},
		{"exact match", "application/json", "application/json", true},
		{"contains match", "text/html, application/json", "application/json", true},
		{"no match", "text/html", "application/json", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRequest("GET", "/", map[string]string{"Accept": tt.accept})
			if got := httputil.Accepts(r, tt.contentType); got != tt.want {
				t.Errorf("Accepts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name string
		auth string
		want string
	}{
		{"valid", "Bearer abc123", "abc123"},
		{"basic auth", "Basic dXNlcjpwYXNz", ""},
		{"empty", "", ""},
		{"bearer lowercase", "bearer abc123", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRequest("GET", "/", map[string]string{"Authorization": tt.auth})
			if got := httputil.BearerToken(r); got != tt.want {
				t.Errorf("BearerToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBasicAuth(t *testing.T) {
	t.Run("valid basic auth", func(t *testing.T) {
		r := newRequest("GET", "/", nil)
		r.SetBasicAuth("user", "pass")
		u, p, ok := httputil.BasicAuth(r)
		if !ok || u != "user" || p != "pass" {
			t.Errorf("BasicAuth() = (%q, %q, %v), want (user, pass, true)", u, p, ok)
		}
	})
	t.Run("no auth", func(t *testing.T) {
		r := newRequest("GET", "/", nil)
		_, _, ok := httputil.BasicAuth(r)
		if ok {
			t.Error("expected BasicAuth to return ok=false")
		}
	})
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}
	err := httputil.WriteJSON(w, http.StatusCreated, data)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got["hello"] != "world" {
		t.Errorf("body = %v, want {hello: world}", got)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	err := httputil.WriteError(w, http.StatusBadRequest, "invalid input")
	if err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got["error"] != "invalid input" {
		t.Errorf("error = %q, want %q", got["error"], "invalid input")
	}
}

func TestFullURL(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		r := newRequest("GET", "/api/v1?q=test", nil)
		r.Host = "example.com"
		r.RequestURI = "/api/v1?q=test"
		got := httputil.FullURL(r)
		want := "http://example.com/api/v1?q=test"
		if got != want {
			t.Errorf("FullURL() = %q, want %q", got, want)
		}
	})
	t.Run("https via header", func(t *testing.T) {
		r := newRequest("GET", "/secure", map[string]string{"X-Forwarded-Proto": "https"})
		r.Host = "example.com"
		r.RequestURI = "/secure"
		got := httputil.FullURL(r)
		want := "https://example.com/secure"
		if got != want {
			t.Errorf("FullURL() = %q, want %q", got, want)
		}
	})
}

func BenchmarkBearerToken(b *testing.B) {
	r := newRequest("GET", "/", map[string]string{"Authorization": "Bearer my-secret-token-12345"})
	for b.Loop() {
		httputil.BearerToken(r)
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	data := map[string]string{"status": "ok", "message": "hello"}
	for b.Loop() {
		w := httptest.NewRecorder()
		_ = httputil.WriteJSON(w, http.StatusOK, data)
	}
}
