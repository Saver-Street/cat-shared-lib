package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func apikeyOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAPIKey_ValidKey(t *testing.T) {
	handler := APIKey("X-API-Key", "secret123")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
}

func TestAPIKey_InvalidKey(t *testing.T) {
	handler := APIKey("X-API-Key", "secret123")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusUnauthorized)
}

func TestAPIKey_MissingHeader(t *testing.T) {
	handler := APIKey("X-API-Key", "secret123")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusUnauthorized)
}

func TestAPIKeyQuery_ValidKey(t *testing.T) {
	handler := APIKeyQuery("api_key", "mysecret")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/?api_key=mysecret", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
}

func TestAPIKeyQuery_InvalidKey(t *testing.T) {
	handler := APIKeyQuery("api_key", "mysecret")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/?api_key=wrong", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusUnauthorized)
}

func TestAPIKeyQuery_MissingParam(t *testing.T) {
	handler := APIKeyQuery("api_key", "mysecret")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusUnauthorized)
}

func TestAPIKeyMulti_FirstKey(t *testing.T) {
	handler := APIKeyMulti("X-API-Key", []string{"key1", "key2", "key3"})(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "key1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
}

func TestAPIKeyMulti_LastKey(t *testing.T) {
	handler := APIKeyMulti("X-API-Key", []string{"key1", "key2", "key3"})(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "key3")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
}

func TestAPIKeyMulti_InvalidKey(t *testing.T) {
	handler := APIKeyMulti("X-API-Key", []string{"key1", "key2"})(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusUnauthorized)
}

func TestAPIKeyMulti_EmptyKeys(t *testing.T) {
	handler := APIKeyMulti("X-API-Key", nil)(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "anything")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusUnauthorized)
}

func BenchmarkAPIKey(b *testing.B) {
	handler := APIKey("X-API-Key", "secret123")(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret123")
	for b.Loop() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkAPIKeyMulti(b *testing.B) {
	handler := APIKeyMulti("X-API-Key", []string{"key1", "key2", "key3"})(apikeyOKHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "key2")
	for b.Loop() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
