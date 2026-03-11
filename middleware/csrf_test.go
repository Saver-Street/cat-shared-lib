package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestCSRF_SafeMethodSetsTokenCookie(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusOK)
	cookies := rr.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "_csrf" {
			found = true
			testkit.AssertTrue(t, c.HttpOnly)
			testkit.AssertTrue(t, c.Secure)
			testkit.AssertEqual(t, c.SameSite, http.SameSiteStrictMode)
			testkit.AssertEqual(t, c.Path, "/")
			testkit.AssertEqual(t, len(c.Value), 64)
		}
	}
	testkit.AssertTrue(t, found)
}

func TestCSRF_SafeMethodsAreSkipped(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace}
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(m, "/", nil)
			h.ServeHTTP(rr, req)
			testkit.AssertEqual(t, rr.Code, http.StatusOK)
		})
	}
}

func TestCSRF_UnsafeMethodWithoutTokenRejects(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(m, "/", nil)
			h.ServeHTTP(rr, req)
			testkit.AssertEqual(t, rr.Code, http.StatusForbidden)
		})
	}
}

func TestCSRF_UnsafeMethodWithValidHeaderToken(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	var token string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "_csrf" {
			token = c.Value
		}
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusOK)
}

func TestCSRF_UnsafeMethodWithValidFormToken(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	var token string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "_csrf" {
			token = c.Value
		}
	}
	form := url.Values{"csrf_token": {token}}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: token})
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusOK)
}

func TestCSRF_WrongTokenRejects(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "real-token"})
	req.Header.Set("X-CSRF-Token", "wrong-token")
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusForbidden)
}

func TestCSRF_EmptySubmittedToken(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "some-token"})
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusForbidden)
}

func TestCSRF_CustomConfig(t *testing.T) {
	secure := false
	h := CSRF(CSRFConfig{
		TokenLength: 16,
		TokenHeader: "X-Custom-CSRF",
		FormField:   "my_csrf",
		CookieName:  "my_csrf_cookie",
		CookiePath:  "/api",
		Secure:      &secure,
		SameSite:    http.SameSiteLaxMode,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	h.ServeHTTP(rr, req)
	var token string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "my_csrf_cookie" {
			token = c.Value
			testkit.AssertEqual(t, c.Path, "/api")
			testkit.AssertTrue(t, !c.Secure)
			testkit.AssertEqual(t, c.SameSite, http.SameSiteLaxMode)
			testkit.AssertEqual(t, len(c.Value), 32)
		}
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "my_csrf_cookie", Value: token})
	req.Header.Set("X-Custom-CSRF", token)
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusOK)
}

func TestCSRF_SkipCheck(t *testing.T) {
	h := CSRF(CSRFConfig{
		SkipCheck: func(r *http.Request) bool {
			return r.URL.Path == "/webhook"
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusOK)
}

func TestCSRF_ErrorHandler(t *testing.T) {
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "custom error")
	})
	h := CSRF(CSRFConfig{
		ErrorHandler: customHandler,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "tok"})
	h.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusTeapot)
	testkit.AssertEqual(t, strings.TrimSpace(rr.Body.String()), "custom error")
}

func TestCSRF_GetCSRFToken(t *testing.T) {
	var got string
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetCSRFToken(r)
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	testkit.AssertTrue(t, got != "")
	testkit.AssertEqual(t, len(got), 64)
}

func TestCSRF_ExistingCookieIsReused(t *testing.T) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	var firstToken string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "_csrf" {
			firstToken = c.Value
		}
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: firstToken})
	h.ServeHTTP(rr, req)
	var secondToken string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "_csrf" {
			secondToken = c.Value
		}
	}
	testkit.AssertEqual(t, firstToken, secondToken)
}

func TestCSRFTokensMatch(t *testing.T) {
	testkit.AssertTrue(t, csrfTokensMatch("abc123", "abc123"))
	testkit.AssertTrue(t, !csrfTokensMatch("abc", "def"))
	testkit.AssertTrue(t, !csrfTokensMatch("", "abc"))
	testkit.AssertTrue(t, !csrfTokensMatch("abc", ""))
	testkit.AssertTrue(t, !csrfTokensMatch("", ""))
}

func TestCSRFIsSafeMethod(t *testing.T) {
	testkit.AssertTrue(t, csrfIsSafeMethod("GET"))
	testkit.AssertTrue(t, csrfIsSafeMethod("HEAD"))
	testkit.AssertTrue(t, csrfIsSafeMethod("OPTIONS"))
	testkit.AssertTrue(t, csrfIsSafeMethod("TRACE"))
	testkit.AssertTrue(t, !csrfIsSafeMethod("POST"))
	testkit.AssertTrue(t, !csrfIsSafeMethod("PUT"))
	testkit.AssertTrue(t, !csrfIsSafeMethod("PATCH"))
	testkit.AssertTrue(t, !csrfIsSafeMethod("DELETE"))
}

func TestCSRFConfig_Defaults(t *testing.T) {
	cfg := CSRFConfig{}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.TokenLength, 32)
	testkit.AssertEqual(t, cfg.TokenHeader, "X-CSRF-Token")
	testkit.AssertEqual(t, cfg.FormField, "csrf_token")
	testkit.AssertEqual(t, cfg.CookieName, "_csrf")
	testkit.AssertEqual(t, cfg.CookiePath, "/")
	testkit.AssertTrue(t, *cfg.Secure)
	testkit.AssertEqual(t, cfg.SameSite, http.SameSiteStrictMode)
}

func TestCSRFConfig_DefaultsPreserveCustom(t *testing.T) {
	secure := false
	cfg := CSRFConfig{
		TokenLength: 16,
		TokenHeader: "X-My-Token",
		FormField:   "my_token",
		CookieName:  "my_cookie",
		CookiePath:  "/api",
		Secure:      &secure,
		SameSite:    http.SameSiteLaxMode,
	}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.TokenLength, 16)
	testkit.AssertEqual(t, cfg.TokenHeader, "X-My-Token")
	testkit.AssertEqual(t, cfg.FormField, "my_token")
	testkit.AssertEqual(t, cfg.CookieName, "my_cookie")
	testkit.AssertEqual(t, cfg.CookiePath, "/api")
	testkit.AssertTrue(t, !*cfg.Secure)
	testkit.AssertEqual(t, cfg.SameSite, http.SameSiteLaxMode)
}

func BenchmarkCSRF_SafeMethod(b *testing.B) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for b.Loop() {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rr, req)
	}
}

func BenchmarkCSRF_UnsafeMethod(b *testing.B) {
	h := CSRF(CSRFConfig{})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	var tok string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "_csrf" {
			tok = c.Value
		}
	}
	for b.Loop() {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: "_csrf", Value: tok})
		req.Header.Set("X-CSRF-Token", tok)
		h.ServeHTTP(rr, req)
	}
}

func TestCSRFGenerateToken(t *testing.T) {
	tok := csrfGenerateToken(32)
	testkit.AssertEqual(t, len(tok), 64)
	tok2 := csrfGenerateToken(16)
	testkit.AssertEqual(t, len(tok2), 32)
}

func TestCSRFTokenFromCookie_NoCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, csrfTokenFromCookie(req, "_csrf"), "")
}

func TestCSRFTokenFromCookie_WithCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "mytoken"})
	testkit.AssertEqual(t, csrfTokenFromCookie(req, "_csrf"), "mytoken")
}

func TestGetCSRFToken_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetCSRFToken(req), "")
}

func ExampleCSRF() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := GetCSRFToken(r)
		fmt.Fprintf(w, `<input type="hidden" name="csrf_token" value="%s">`, token)
	})
	mux := http.NewServeMux()
	mux.Handle("/", CSRF(CSRFConfig{})(handler))
	_ = mux // use with http.ListenAndServe
}
