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

// Package auth provides authentication and signature generation functionality
// for the OneMoney API.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	// Algorithm is the signature algorithm identifier.
	Algorithm = "OneMoney-HMAC-SHA256"

	// HeaderDate is the custom date header name.
	HeaderDate = "X-OM-Date"

	// HeaderAuthorization is the authorization header name.
	HeaderAuthorization = "Authorization"

	// TimeFormat is the timestamp format used in signatures.
	TimeFormat = "20060102T150405Z"
)

// Credentials holds the API access credentials.
type Credentials struct {
	AccessKey string
	SecretKey string
}

// NewCredentials creates new API credentials.
func NewCredentials(accessKey, secretKey string) *Credentials {
	return &Credentials{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
}

// Signer handles request signature generation.
type Signer struct {
	credentials *Credentials
}

// NewSigner creates a new request signer with the given credentials.
func NewSigner(creds *Credentials) *Signer {
	return &Signer{
		credentials: creds,
	}
}

// SignatureResult contains the generated signature and related metadata.
type SignatureResult struct {
	Authorization string
	Timestamp     string
	BodyHash      string
}

// SignRequest generates a signature for an HTTP request.
//
// It takes the HTTP method, URI path, and request body, then computes
// the HMAC-SHA256 signature according to the OneMoney API specification.
func (s *Signer) SignRequest(method, path string, body []byte) (*SignatureResult, error) {
	// Generate timestamp
	timestamp := s.getTimestamp()

	// Calculate body hash
	bodyHash := s.hashBody(body)

	// Build string to sign
	stringToSign := s.buildStringToSign(method, path, timestamp, bodyHash)

	// Calculate signature
	signature, err := s.calculateSignature(stringToSign)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate signature: %w", err)
	}

	// Build authorization header
	authHeader := s.buildAuthorizationHeader(timestamp, signature)

	return &SignatureResult{
		Authorization: authHeader,
		Timestamp:     timestamp,
		BodyHash:      bodyHash,
	}, nil
}

// getTimestamp returns the current UTC timestamp in OneMoney format.
func (*Signer) getTimestamp() string {
	return time.Now().UTC().Format(TimeFormat)
}

// hashBody calculates the SHA256 hash of the request body.
func (*Signer) hashBody(body []byte) string {
	hasher := sha256.New()
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))
}

// buildStringToSign constructs the canonical string that will be signed.
func (s *Signer) buildStringToSign(method, path, timestamp, bodyHash string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		s.credentials.AccessKey,
		timestamp,
		strings.ToUpper(method),
		path,
		bodyHash,
	)
}

// calculateSignature computes the HMAC-SHA256 signature of the string to sign.
func (s *Signer) calculateSignature(stringToSign string) (string, error) {
	// Decode base64 URL-safe encoded secret key
	keyBytes, err := s.decodeSecretKey()
	if err != nil {
		return "", err
	}

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(stringToSign))
	signature := mac.Sum(nil)

	// Return hex-encoded signature
	return hex.EncodeToString(signature), nil
}

// decodeSecretKey decodes the base64 URL-safe encoded secret key.
// It automatically adds padding if needed.
func (s *Signer) decodeSecretKey() ([]byte, error) {
	secretKey := s.credentials.SecretKey

	// Add padding if needed for base64 decoding
	padding := (4 - len(secretKey)%4) % 4
	secretKeyWithPadding := secretKey + strings.Repeat("=", padding)

	keyBytes, err := base64.URLEncoding.DecodeString(secretKeyWithPadding)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret key: %w", err)
	}

	return keyBytes, nil
}

// buildAuthorizationHeader constructs the Authorization header value.
func (s *Signer) buildAuthorizationHeader(timestamp, signature string) string {
	return fmt.Sprintf("%s %s:%s:%s",
		Algorithm,
		s.credentials.AccessKey,
		timestamp,
		signature,
	)
}
