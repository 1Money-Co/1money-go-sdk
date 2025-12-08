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

// Package transport provides HTTP transport functionality for the SDK.
package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// Sentinel errors for common error cases.
var (
	ErrAuthentication = errors.New("authentication failed")
	ErrForbidden      = errors.New("access forbidden")
	ErrNotFound       = errors.New("resource not found")
	ErrRateLimited    = errors.New("rate limit exceeded")
	ErrServerError    = errors.New("server error")
	ErrUnprocessable  = errors.New("unprocessable entity")
)

// APIError represents an API error response.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	Detail     string `json:"detail,omitempty"`   // Detailed error description from API
	Instance   string `json:"instance,omitempty"` // API endpoint that caused the error
	RequestID  string `json:"request_id,omitempty"`
	// Raw response body for debugging
	RawBody string `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	var b strings.Builder

	// Start with status code
	if e.Code != "" {
		fmt.Fprintf(&b, "API error %d (%s)", e.StatusCode, e.Code)
	} else {
		fmt.Fprintf(&b, "API error %d", e.StatusCode)
	}

	// Add the main error message
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	} else if e.Status != "" {
		fmt.Fprintf(&b, ": %s", e.Status)
	}

	// Add instance (API endpoint) if available
	if e.Instance != "" {
		fmt.Fprintf(&b, " [endpoint: %s]", e.Instance)
	}

	return b.String()
}

// Unwrap returns the underlying sentinel error for this API error.
// This allows errors.Is() to work with sentinel errors.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusUnauthorized:
		return ErrAuthentication
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusUnprocessableEntity:
		return ErrUnprocessable
	default:
		if e.IsServerError() {
			return ErrServerError
		}
		return nil
	}
}

// IsAuthError returns true if this is an authentication error (401).
func (e *APIError) IsAuthError() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbiddenError returns true if this is a forbidden error (403).
func (e *APIError) IsForbiddenError() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsNotFoundError returns true if this is a not found error (404).
func (e *APIError) IsNotFoundError() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnprocessableError returns true if this is an unprocessable entity error (422).
func (e *APIError) IsUnprocessableError() bool {
	return e.StatusCode == http.StatusUnprocessableEntity
}

// IsRateLimitError returns true if this is a rate limit error (429).
func (e *APIError) IsRateLimitError() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsClientError returns true if this is a client error (4xx).
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if this is a server error (5xx).
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsRetryable returns true if the error is potentially retryable.
// Retryable errors include rate limits and gateway errors.
// Note: 500 Internal Server Error is NOT retryable as it typically indicates
// permanent failures (business logic errors, invalid state) that won't be resolved by retrying.
func (e *APIError) IsRetryable() bool {
	switch e.StatusCode {
	case http.StatusTooManyRequests,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// errorResponse represents the API error response format.
// Example: {"code":"Unprocessable_Entity","status":422,"detail":"...","instance":"/v1/customers"}
type errorResponse struct {
	Code     string `json:"code,omitempty"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// parseErrorResponse attempts to parse the error response body.
func parseErrorResponse(statusCode int, status string, body []byte) *APIError {
	log := getLogger()

	apiErr := &APIError{
		StatusCode: statusCode,
		Status:     status,
		RawBody:    string(body),
	}

	// Try to parse the error response
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Detail != "" {
		apiErr.Code = errResp.Code
		apiErr.Detail = errResp.Detail
		apiErr.Instance = errResp.Instance
		apiErr.Message = errResp.Detail

		log.Debug("parsed API error response",
			zap.Int("status_code", statusCode),
			zap.String("code", apiErr.Code),
			zap.String("instance", apiErr.Instance),
			zap.String("detail", apiErr.Detail),
		)

		return apiErr
	}

	// If no message was parsed, use a default based on status code
	apiErr.Message = getDefaultErrorMessage(statusCode)

	log.Warn("failed to parse error response, using default message",
		zap.Int("status_code", statusCode),
		zap.String("status", status),
		zap.String("raw_body", string(body)),
	)

	return apiErr
}

// getDefaultErrorMessage returns a user-friendly error message based on status code.
func getDefaultErrorMessage(statusCode int) string {
	switch statusCode {
	case http.StatusUnauthorized:
		return "authentication failed, please verify your credentials (ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY)"
	case http.StatusForbidden:
		return "access forbidden, you don't have permission to access this resource"
	case http.StatusNotFound:
		return "resource not found, please check the API endpoint"
	case http.StatusUnprocessableEntity:
		return "unprocessable entity, request validation failed"
	case http.StatusTooManyRequests:
		return "too many requests, please reduce request rate and try again later"
	case http.StatusInternalServerError:
		return "internal server error, please try again later or contact support"
	case http.StatusBadGateway:
		return "bad gateway, the server is temporarily unavailable"
	case http.StatusServiceUnavailable:
		return "service unavailable, the server is under maintenance or overloaded"
	case http.StatusGatewayTimeout:
		return "gateway timeout, the server took too long to respond"
	default:
		if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
			return "client error occurred, please check your request parameters"
		} else if statusCode >= http.StatusInternalServerError {
			return "server error occurred, please try again later"
		}
		return "an unexpected error occurred"
	}
}

// IsAPIError checks if the error is an APIError and returns it.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// IsAuthError checks if the error is an authentication error (401).
func IsAuthError(err error) bool {
	if errors.Is(err, ErrAuthentication) {
		return true
	}
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsAuthError()
}

// IsForbiddenError checks if the error is a forbidden error (403).
func IsForbiddenError(err error) bool {
	if errors.Is(err, ErrForbidden) {
		return true
	}
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsForbiddenError()
}

// IsNotFoundError checks if the error is a not found error (404).
func IsNotFoundError(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsNotFoundError()
}

// IsUnprocessableError checks if the error is an unprocessable entity error (422).
func IsUnprocessableError(err error) bool {
	if errors.Is(err, ErrUnprocessable) {
		return true
	}
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsUnprocessableError()
}

// IsRateLimitError checks if the error is a rate limit error (429).
func IsRateLimitError(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsRateLimitError()
}

// IsClientError checks if the error is a client error (4xx).
func IsClientError(err error) bool {
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsClientError()
}

// IsServerError checks if the error is a server error (5xx).
func IsServerError(err error) bool {
	if errors.Is(err, ErrServerError) {
		return true
	}
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsServerError()
}

// IsRetryable checks if the error is potentially retryable.
func IsRetryable(err error) bool {
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsRetryable()
}

// checkEmbeddedRateLimitError checks if the response body contains an embedded rate limit error.
// Some APIs return HTTP 200 with rate limit info in the body:
// {"code":"Too_Many_Requests","status":429,"detail":"Rate limit exceeded. Retry after 4s."}
func checkEmbeddedRateLimitError(body []byte) *APIError {
	if len(body) == 0 || body[0] != '{' {
		return nil
	}

	var resp errorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil
	}

	// Check if this is a rate limit response
	if resp.Status == http.StatusTooManyRequests || resp.Code == "Too_Many_Requests" {
		return &APIError{
			StatusCode: http.StatusTooManyRequests,
			Status:     "429 Too Many Requests",
			Code:       resp.Code,
			Detail:     resp.Detail,
			Message:    resp.Detail,
			Instance:   resp.Instance,
			RawBody:    string(body),
		}
	}

	return nil
}
