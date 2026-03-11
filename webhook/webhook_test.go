package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSign(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"id":"123"}`)
	sig := Sign(secret, payload)

	if !strings.HasPrefix(sig, "sha256=") {
		t.Errorf("signature should start with sha256=, got %q", sig)
	}
	if len(sig) != 7+64 { // "sha256=" + 64 hex chars
		t.Errorf("unexpected signature length: %d", len(sig))
	}
}

func TestSign_Deterministic(t *testing.T) {
	secret := []byte("key")
	payload := []byte("data")
	if Sign(secret, payload) != Sign(secret, payload) {
		t.Error("Sign should be deterministic")
	}
}

func TestSign_DifferentSecrets(t *testing.T) {
	payload := []byte("data")
	s1 := Sign([]byte("secret1"), payload)
	s2 := Sign([]byte("secret2"), payload)
	if s1 == s2 {
		t.Error("different secrets should produce different signatures")
	}
}

func TestVerify_Valid(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"event":"test"}`)
	sig := Sign(secret, payload)

	if !Verify(secret, payload, sig) {
		t.Error("Verify should return true for valid signature")
	}
}

func TestVerify_Invalid(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"event":"test"}`)

	if Verify(secret, payload, "sha256=invalid") {
		t.Error("Verify should return false for invalid signature")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	payload := []byte(`{"event":"test"}`)
	sig := Sign([]byte("secret1"), payload)

	if Verify([]byte("secret2"), payload, sig) {
		t.Error("Verify should return false for wrong secret")
	}
}

func TestVerify_TamperedPayload(t *testing.T) {
	secret := []byte("test-secret")
	sig := Sign(secret, []byte("original"))

	if Verify(secret, []byte("tampered"), sig) {
		t.Error("Verify should return false for tampered payload")
	}
}

func TestVerifyRequest_Valid(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"id":"1"}`)
	sig := Sign(secret, payload)

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	r.Header.Set(SignatureHeader, sig)

	body, err := VerifyRequest(r, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != string(payload) {
		t.Errorf("body = %q, want %q", body, payload)
	}
}

func TestVerifyRequest_MissingHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	_, err := VerifyRequest(r, []byte("secret"))
	if err == nil {
		t.Fatal("expected error for missing header")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error should mention missing header: %v", err)
	}
}

func TestVerifyRequest_InvalidSignature(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	r.Header.Set(SignatureHeader, "sha256=bad")
	_, err := VerifyRequest(r, []byte("secret"))
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error should mention invalid: %v", err)
	}
}

func TestDeliver_Success(t *testing.T) {
	var gotBody []byte
	var gotSig string
	var gotEvent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotSig = r.Header.Get(SignatureHeader)
		gotEvent = r.Header.Get(EventHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	secret := []byte("deliver-secret")
	type Evt struct {
		ID string `json:"id"`
	}

	err := Deliver(context.Background(), srv.URL, secret, "order.placed", Evt{ID: "42"})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if gotEvent != "order.placed" {
		t.Errorf("event = %q, want %q", gotEvent, "order.placed")
	}

	if !Verify(secret, gotBody, gotSig) {
		t.Error("delivered payload signature is invalid")
	}

	var evt Evt
	if err := json.Unmarshal(gotBody, &evt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.ID != "42" {
		t.Errorf("event.ID = %q, want %q", evt.ID, "42")
	}
}

func TestDeliver_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := Deliver(context.Background(), srv.URL, []byte("s"), "test", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code: %v", err)
	}
}

func TestDeliver_EmptyEventType(t *testing.T) {
	var gotEventHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEventHeader = r.Header.Get(EventHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := Deliver(context.Background(), srv.URL, []byte("s"), "", "data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEventHeader != "" {
		t.Errorf("event header should be empty, got %q", gotEventHeader)
	}
}

func TestDeliver_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Deliver(ctx, srv.URL, []byte("s"), "test", "data")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func BenchmarkSign(b *testing.B) {
	secret := []byte("benchmark-secret")
	payload := []byte(`{"event":"test","data":{"id":"12345","name":"benchmark"}}`)

	for b.Loop() {
		Sign(secret, payload)
	}
}

func BenchmarkVerify(b *testing.B) {
	secret := []byte("benchmark-secret")
	payload := []byte(`{"event":"test","data":{"id":"12345","name":"benchmark"}}`)
	sig := Sign(secret, payload)

	for b.Loop() {
		Verify(secret, payload, sig)
	}
}

func FuzzVerify(f *testing.F) {
	f.Add([]byte("secret"), []byte("payload"), "sha256=abc123")
	f.Add([]byte("key"), []byte(`{"id":1}`), "sha256=0000000000000000000000000000000000000000000000000000000000000000")
	f.Add([]byte{}, []byte{}, "")

	f.Fuzz(func(t *testing.T, secret, payload []byte, sig string) {
		// Sign + Verify round-trip must always succeed.
		computed := Sign(secret, payload)
		if !Verify(secret, payload, computed) {
			t.Error("Sign/Verify round-trip failed")
		}

		// Verify with arbitrary sig must not panic.
		Verify(secret, payload, sig)
	})
}
