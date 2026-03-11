package urlutil_test

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/urlutil"
)

// ─── Join ───

func TestJoin_BaseWithSegments(t *testing.T) {
	got := urlutil.Join("https://example.com/api", "v1", "users")
	want := "https://example.com/api/v1/users"
	if got != want {
		t.Errorf("Join = %q, want %q", got, want)
	}
}

func TestJoin_TrailingSlash(t *testing.T) {
	got := urlutil.Join("https://example.com/api/", "v1")
	want := "https://example.com/api/v1"
	if got != want {
		t.Errorf("Join = %q, want %q", got, want)
	}
}

func TestJoin_EmptySegments(t *testing.T) {
	got := urlutil.Join("https://example.com/api")
	want := "https://example.com/api"
	if got != want {
		t.Errorf("Join = %q, want %q", got, want)
	}
}

func TestJoin_SingleSegment(t *testing.T) {
	got := urlutil.Join("https://example.com", "path")
	want := "https://example.com/path"
	if got != want {
		t.Errorf("Join = %q, want %q", got, want)
	}
}

// ─── SetQuery ───

func TestSetQuery_NewKey(t *testing.T) {
	got := urlutil.SetQuery("https://example.com", "page", "1")
	want := "https://example.com?page=1"
	if got != want {
		t.Errorf("SetQuery = %q, want %q", got, want)
	}
}

func TestSetQuery_OverwriteExisting(t *testing.T) {
	got := urlutil.SetQuery("https://example.com?page=1", "page", "2")
	want := "https://example.com?page=2"
	if got != want {
		t.Errorf("SetQuery = %q, want %q", got, want)
	}
}

func TestSetQuery_PreservesOtherParams(t *testing.T) {
	got := urlutil.SetQuery("https://example.com?foo=bar", "page", "1")
	if got != "https://example.com?foo=bar&page=1" {
		t.Errorf("SetQuery = %q, unexpected", got)
	}
}

// ─── AddQuery ───

func TestAddQuery_AddsToExisting(t *testing.T) {
	got := urlutil.AddQuery("https://example.com?tag=go", "tag", "url")
	// Should contain both tag=go and tag=url
	if got != "https://example.com?tag=go&tag=url" {
		t.Errorf("AddQuery = %q, unexpected", got)
	}
}

func TestAddQuery_NewParam(t *testing.T) {
	got := urlutil.AddQuery("https://example.com", "key", "val")
	want := "https://example.com?key=val"
	if got != want {
		t.Errorf("AddQuery = %q, want %q", got, want)
	}
}

// ─── RemoveQuery ───

func TestRemoveQuery_ExistingKey(t *testing.T) {
	got := urlutil.RemoveQuery("https://example.com?page=1&sort=asc", "page")
	want := "https://example.com?sort=asc"
	if got != want {
		t.Errorf("RemoveQuery = %q, want %q", got, want)
	}
}

func TestRemoveQuery_MissingKey(t *testing.T) {
	got := urlutil.RemoveQuery("https://example.com?sort=asc", "page")
	want := "https://example.com?sort=asc"
	if got != want {
		t.Errorf("RemoveQuery = %q, want %q", got, want)
	}
}

func TestRemoveQuery_LastKey(t *testing.T) {
	got := urlutil.RemoveQuery("https://example.com?page=1", "page")
	want := "https://example.com"
	if got != want {
		t.Errorf("RemoveQuery = %q, want %q", got, want)
	}
}

// ─── StripQuery ───

func TestStripQuery_WithQueryAndFragment(t *testing.T) {
	got := urlutil.StripQuery("https://example.com/path?foo=bar#section")
	want := "https://example.com/path"
	if got != want {
		t.Errorf("StripQuery = %q, want %q", got, want)
	}
}

func TestStripQuery_NoQuery(t *testing.T) {
	got := urlutil.StripQuery("https://example.com/path")
	want := "https://example.com/path"
	if got != want {
		t.Errorf("StripQuery = %q, want %q", got, want)
	}
}

// ─── Domain ───

func TestDomain_FullURL(t *testing.T) {
	got := urlutil.Domain("https://www.example.com/path?q=1")
	want := "www.example.com"
	if got != want {
		t.Errorf("Domain = %q, want %q", got, want)
	}
}

func TestDomain_WithPort(t *testing.T) {
	got := urlutil.Domain("https://example.com:8080/path")
	want := "example.com"
	if got != want {
		t.Errorf("Domain = %q, want %q", got, want)
	}
}

func TestDomain_IPAddress(t *testing.T) {
	got := urlutil.Domain("http://192.168.1.1:3000/api")
	want := "192.168.1.1"
	if got != want {
		t.Errorf("Domain = %q, want %q", got, want)
	}
}

// ─── IsAbsolute ───

func TestIsAbsolute_HTTP(t *testing.T) {
	if !urlutil.IsAbsolute("http://example.com") {
		t.Error("expected true for http URL")
	}
}

func TestIsAbsolute_HTTPS(t *testing.T) {
	if !urlutil.IsAbsolute("https://example.com") {
		t.Error("expected true for https URL")
	}
}

func TestIsAbsolute_Relative(t *testing.T) {
	if urlutil.IsAbsolute("/path/to/resource") {
		t.Error("expected false for relative URL")
	}
}

// ─── IsHTTPS ───

func TestIsHTTPS_True(t *testing.T) {
	if !urlutil.IsHTTPS("https://example.com") {
		t.Error("expected true for https URL")
	}
}

func TestIsHTTPS_HTTP(t *testing.T) {
	if urlutil.IsHTTPS("http://example.com") {
		t.Error("expected false for http URL")
	}
}

func TestIsHTTPS_NoScheme(t *testing.T) {
	if urlutil.IsHTTPS("/path") {
		t.Error("expected false for URL without scheme")
	}
}

// ─── HasQuery ───

func TestHasQuery_Present(t *testing.T) {
	if !urlutil.HasQuery("https://example.com?page=1", "page") {
		t.Error("expected true for existing key")
	}
}

func TestHasQuery_Absent(t *testing.T) {
	if urlutil.HasQuery("https://example.com?page=1", "sort") {
		t.Error("expected false for missing key")
	}
}

// ─── QueryValue ───

func TestQueryValue_Present(t *testing.T) {
	got := urlutil.QueryValue("https://example.com?page=42", "page")
	if got != "42" {
		t.Errorf("QueryValue = %q, want %q", got, "42")
	}
}

func TestQueryValue_Absent(t *testing.T) {
	got := urlutil.QueryValue("https://example.com", "page")
	if got != "" {
		t.Errorf("QueryValue = %q, want empty", got)
	}
}

// ─── Benchmarks ───

func BenchmarkJoin(b *testing.B) {
	for b.Loop() {
		urlutil.Join("https://example.com/api", "v1", "users")
	}
}

func BenchmarkSetQuery(b *testing.B) {
	for b.Loop() {
		urlutil.SetQuery("https://example.com?page=1", "page", "2")
	}
}

// ─── Fuzz ───

func FuzzJoin(f *testing.F) {
	f.Add("https://example.com", "path")
	f.Add("", "")
	f.Add("http://localhost:8080/api", "v1/users")
	f.Fuzz(func(t *testing.T, base, segment string) {
		// Must not panic
		urlutil.Join(base, segment)
	})
}
