package middleware

import (
"crypto/sha256"
"encoding/hex"
"net/http"
"net/http/httptest"
)

// ETag returns middleware that computes a weak ETag from the response body
// hash and returns 304 Not Modified when the client's If-None-Match header
// matches. Only applies to successful GET/HEAD responses.
func ETag(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet && r.Method != http.MethodHead {
next.ServeHTTP(w, r)
return
}

rec := httptest.NewRecorder()
next.ServeHTTP(rec, r)

if rec.Code < 200 || rec.Code >= 300 {
copyResponse(w, rec)
return
}

body := rec.Body.Bytes()
h := sha256.Sum256(body)
tag := `W/"` + hex.EncodeToString(h[:8]) + `"`

if r.Header.Get("If-None-Match") == tag {
w.Header().Set("ETag", tag)
w.WriteHeader(http.StatusNotModified)
return
}

w.Header().Set("ETag", tag)
copyResponse(w, rec)
})
}

func copyResponse(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
for k, vs := range rec.Header() {
for _, v := range vs {
w.Header().Add(k, v)
}
}
w.WriteHeader(rec.Code)
_, _ = w.Write(rec.Body.Bytes())
}
