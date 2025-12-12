/*
 * Copyright 2025 1Money Co.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package webhook provides helpers for building webhook receivers for the
// OneMoney platform.
//
// This package is intentionally minimal and focused on example usage. It
// exposes an HTTP handler that:
//   - validates request method
//   - verifies webhook signatures and timestamp freshness
//   - extracts standard webhook headers
//   - hands the payload to a user-provided callback
//
// The goal is to make it easy to bootstrap a local webhook server while still
// keeping application-specific logic in the caller.
package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	// HeaderSignature is the header that carries the webhook signature.
	HeaderSignature = "X-Webhook-Signature"

	// HeaderTimestamp is the header that carries the webhook timestamp as a
	// Unix timestamp in seconds.
	HeaderTimestamp = "X-Webhook-Timestamp"

	// HeaderEventID is the unique identifier for the event.
	HeaderEventID = "X-Webhook-Event-Id"

	// HeaderEventType identifies the type of event.
	HeaderEventType = "X-Webhook-Event-Type"

	// HeaderDeliveryID identifies the delivery attempt for the event.
	HeaderDeliveryID = "X-Webhook-Delivery-Id"

	// DefaultSecretKey is a convenience key for local development and examples.
	// In production you must configure WEBHOOK_SECRET_KEY explicitly and never
	// rely on this value.
	DefaultSecretKey = "LTCs85bMOvmLxjKmfMete2FsH-nfa3qP1PdVSOSbeLo"

	// DefaultTolerance is the default maximum allowed clock skew between the
	// webhook timestamp and the local time.
	DefaultTolerance = 5 * time.Minute
)

// Metadata contains the standard webhook metadata extracted from request headers.
type Metadata struct {
	EventID    string
	EventType  string
	DeliveryID string
	Timestamp  time.Time
	RawUnixTs  int64
}

// HandlerFunc is the user callback that processes a verified webhook event.
//
// The payload is provided as json.RawMessage so callers can either log it
// directly or unmarshal into a typed struct.
type HandlerFunc func(ctx context.Context, meta Metadata, payload json.RawMessage)

// HandlerOptions configures the webhook HTTP handler.
type HandlerOptions struct {
	// Tolerance defines the maximum allowed difference between the webhook
	// timestamp and the local clock. If zero, DefaultTolerance is used.
	Tolerance time.Duration

	// OnEvent is invoked after the webhook signature and timestamp have been
	// validated. If nil, events are simply ignored after verification.
	OnEvent HandlerFunc

	// Sandbox, when true, disables signature and timestamp verification.
	// This is intended for local development and testing only.
	Sandbox bool
}

// NewHandler returns an http.Handler that verifies and dispatches webhook events.
//
// The handler assumes:
//   - HTTP method is POST
//   - Body is JSON-encoded
//   - Signature and timestamp are provided via the standard headers
//
// Signature verification is intentionally kept simple and self-contained:
//
//	expected_signature = hex(hmac_sha256(secret_key, "<timestamp>.<body>"))
//
// This matches the strategy used in the Rust example and is suitable for
// local development. For production, ensure it matches the server-side
// specification exactly.
func NewHandler(secretKey string, opts *HandlerOptions) http.Handler {
	if secretKey == "" {
		secretKey = DefaultSecretKey
	}

	h := &handler{
		secretKey: secretKey,
	}

	if opts != nil {
		h.tolerance = opts.Tolerance
		h.onEvent = opts.OnEvent
		h.sandbox = opts.Sandbox
	}

	if h.tolerance <= 0 {
		h.tolerance = DefaultTolerance
	}

	return h
}

type handler struct {
	secretKey string
	tolerance time.Duration
	onEvent   HandlerFunc
	sandbox   bool
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var unixTs int64

	if h.sandbox {
		// In sandbox mode we do not enforce signature or timestamp validation.
		// If a timestamp header is present and valid, prefer it; otherwise use
		// the current time for metadata purposes.
		tsHeader := r.Header.Get(HeaderTimestamp)
		if tsHeader != "" {
			if ts, err := strconv.ParseInt(tsHeader, 10, 64); err == nil {
				unixTs = ts
			}
		}
		if unixTs == 0 {
			unixTs = time.Now().UTC().Unix()
		}
	} else {
		signature := r.Header.Get(HeaderSignature)
		if signature == "" {
			http.Error(w, "missing signature header", http.StatusUnauthorized)
			return
		}

		tsHeader := r.Header.Get(HeaderTimestamp)
		if tsHeader == "" {
			http.Error(w, "missing timestamp header", http.StatusUnauthorized)
			return
		}

		var err error
		unixTs, err = strconv.ParseInt(tsHeader, 10, 64)
		if err != nil {
			http.Error(w, "invalid timestamp header", http.StatusBadRequest)
			return
		}

		if err := h.verifyTimestamp(unixTs); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ok, err := verifySignature(body, signature, h.secretKey, unixTs)
		if err != nil {
			http.Error(w, "failed to verify signature", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
			return
		}
	}

	// At this point the request is verified. Build metadata and dispatch.
	meta := Metadata{
		EventID:    r.Header.Get(HeaderEventID),
		EventType:  r.Header.Get(HeaderEventType),
		DeliveryID: r.Header.Get(HeaderDeliveryID),
		RawUnixTs:  unixTs,
		Timestamp:  time.Unix(unixTs, 0).UTC(),
	}

	var raw json.RawMessage
	if len(body) > 0 {
		raw = append(raw, body...)
	}

	if h.onEvent != nil {
		h.onEvent(r.Context(), meta, raw)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(fmt.Sprintf(`{"status":"success","message":"Webhook received and processed","event_id":"%s","delivery_id":"%s"}`, meta.EventID, meta.DeliveryID)))
}

func (h *handler) verifyTimestamp(unixTs int64) error {
	now := time.Now().UTC().Unix()
	diff := time.Duration(abs64(now-unixTs)) * time.Second
	if diff > h.tolerance {
		return fmt.Errorf("timestamp outside allowed tolerance")
	}
	return nil
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// computeSignature calculates the webhook signature for the given payload.
//
// The secretKey is treated as raw bytes. Callers should ensure it matches
// the format used by the OneMoney backend for webhook signing.
func computeSignature(body []byte, secretKey string, unixTs int64) (string, error) {
	mac := hmac.New(sha256.New, []byte(secretKey))
	if _, err := mac.Write([]byte(fmt.Sprintf("%d.", unixTs))); err != nil {
		return "", err
	}
	if _, err := mac.Write(body); err != nil {
		return "", err
	}
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum), nil
}

// verifySignature verifies the provided signature for the given payload and timestamp.
func verifySignature(body []byte, signature string, secretKey string, unixTs int64) (bool, error) {
	expected, err := computeSignature(body, secretKey, unixTs)
	if err != nil {
		return false, err
	}
	// Constant time compare to avoid timing attacks.
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) != 1 {
		return false, nil
	}
	return true, nil
}
