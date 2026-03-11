package shutdown

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func FuzzDrainer_AddDone(f *testing.F) {
	f.Add(1, 1)
	f.Add(10, 10)
	f.Add(0, 0)
	f.Add(100, 50)

	f.Fuzz(func(t *testing.T, adds, dones int) {
		if adds < 0 || adds > 200 {
			t.Skip()
		}
		if dones < 0 || dones > adds {
			t.Skip()
		}

		d := &Drainer{}
		var wg sync.WaitGroup

		for i := 0; i < adds; i++ {
			wg.Add(1)
			d.Add()
			go func() {
				defer wg.Done()
				d.Done()
			}()
		}

		wg.Wait()
		d.Wait() // must not block
	})
}

func FuzzDrainer_Middleware(f *testing.F) {
	f.Add(1)
	f.Add(5)
	f.Add(20)

	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 || n > 100 {
			t.Skip()
		}

		d := &Drainer{}
		handler := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rr := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/", nil)
				handler.ServeHTTP(rr, req)
			}()
		}
		wg.Wait()
		d.Wait() // must not block after all requests complete
	})
}
