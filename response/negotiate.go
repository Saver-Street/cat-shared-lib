package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

// Negotiate writes data to w in the format requested by the Accept header.
// It supports JSON (default) and XML. If the Accept header is absent or
// contains */* or application/json, JSON is used. If it contains
// application/xml or text/xml, XML is used.
func Negotiate(w http.ResponseWriter, r *http.Request, status int, data any) {
	accept := r.Header.Get("Accept")
	if wantsXML(accept) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(status)
		_ = xml.NewEncoder(w).Encode(data)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func wantsXML(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		mt := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if mt == "application/xml" || mt == "text/xml" {
			return true
		}
	}
	return false
}

// NegotiateOK is a convenience wrapper that writes with 200 OK.
func NegotiateOK(w http.ResponseWriter, r *http.Request, data any) {
	Negotiate(w, r, http.StatusOK, data)
}
