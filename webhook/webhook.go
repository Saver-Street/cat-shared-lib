package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// SignatureHeader is the default HTTP header used for the HMAC signature.
	SignatureHeader = "X-Signature-256"

	// EventHeader is the default HTTP header used for the event type.
	EventHeader = "X-Event-Type"
)

// Sign computes the HMAC-SHA256 signature of payload using the given secret
// and returns the hex-encoded result prefixed with "sha256=".
func Sign(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Verify checks that the given signature matches the HMAC-SHA256 of payload
// using secret. The signature should be in the format "sha256=<hex>".
func Verify(secret, payload []byte, signature string) bool {
	expected := Sign(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifyRequest reads the body of r, verifies the signature in the
// X-Signature-256 header, and returns the raw body on success.
func VerifyRequest(r *http.Request, secret []byte) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("webhook: read body: %w", err)
	}

	sig := r.Header.Get(SignatureHeader)
	if sig == "" {
		return nil, fmt.Errorf("webhook: missing %s header", SignatureHeader)
	}

	if !Verify(secret, body, sig) {
		return nil, fmt.Errorf("webhook: invalid signature")
	}
	return body, nil
}

// Deliver marshals event as JSON, signs it with secret, and POSTs it to url.
// The request includes X-Signature-256 and X-Event-Type headers. The
// eventType parameter is set as the X-Event-Type header value.
func Deliver(ctx context.Context, url string, secret []byte, eventType string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	sig := Sign(secret, payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(SignatureHeader, sig)
	if eventType != "" {
		req.Header.Set(EventHeader, eventType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: deliver: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
