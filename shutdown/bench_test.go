package shutdown

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkDrainer_AddDone(b *testing.B) {
	d := &Drainer{}
	for b.Loop() {
		d.Add()
		d.Done()
	}
}

func BenchmarkDrainer_Middleware(b *testing.B) {
	d := &Drainer{}
	handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
