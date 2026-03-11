// Package webhook provides HMAC-SHA256 signed webhook delivery and
// verification for both sending and receiving webhooks.
//
// # Sending webhooks
//
// Use [Sign] to compute the HMAC-SHA256 signature of a payload, and
// [Deliver] to POST a signed JSON payload to a URL.
//
//	sig := webhook.Sign(secret, payload)
//	err := webhook.Deliver(ctx, url, secret, event)
//
// # Receiving webhooks
//
// Use [Verify] to validate incoming webhook signatures, or
// [VerifyRequest] to validate an *http.Request directly.
//
//	ok := webhook.Verify(secret, payload, signatureHeader)
//	payload, err := webhook.VerifyRequest(r, secret)
package webhook
