// Package transport provides HTTP transport functionality for the SDK.
package transport

import (
	"encoding/json"
	"errors"
	"fmt"
)

// APIError represents an API error response.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	// Raw response body for debugging
	RawBody string `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Status)
}

// IsAuthError returns true if this is an authentication error (401).
func (e *APIError) IsAuthError() bool {
	return e.StatusCode == 401
}

// IsForbiddenError returns true if this is a forbidden error (403).
func (e *APIError) IsForbiddenError() bool {
	return e.StatusCode == 403
}

// IsNotFoundError returns true if this is a not found error (404).
func (e *APIError) IsNotFoundError() bool {
	return e.StatusCode == 404
}

// IsServerError returns true if this is a server error (5xx).
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// errorResponse represents a standard API error response format.
type errorResponse struct {
	Message   string `json:"message"`
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
}

// parseErrorResponse attempts to parse the error response body.
func parseErrorResponse(statusCode int, status string, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Status:     status,
		RawBody:    string(body),
	}

	// Try to parse as JSON error response
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		// Use either "message" or "error" field
		if errResp.Message != "" {
			apiErr.Message = errResp.Message
		} else if errResp.Error != "" {
			apiErr.Message = errResp.Error
		}
		apiErr.Code = errResp.Code
		apiErr.RequestID = errResp.RequestID
	}

	// If no message was parsed, use a default based on status code
	if apiErr.Message == "" {
		apiErr.Message = getDefaultErrorMessage(statusCode)
	}

	return apiErr
}

// getDefaultErrorMessage returns a user-friendly error message based on status code.
func getDefaultErrorMessage(statusCode int) string {
	switch statusCode {
	case 401:
		return `Authentication failed. Please verify your credentials:
  1. Check your ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY in .env file
  2. Verify credentials in ~/.onemoney/credentials file
  3. Ensure access key and secret key are correct
  4. Check if credentials have expired or been revoked`
	case 403:
		return "Access forbidden. You don't have permission to access this resource."
	case 404:
		return "Resource not found. Please check the API endpoint."
	case 429:
		return "Too many requests. Please reduce request rate and try again later."
	case 500:
		return "Internal server error. Please try again later or contact support."
	case 502:
		return "Bad gateway. The server is temporarily unavailable."
	case 503:
		return "Service unavailable. The server is under maintenance or overloaded."
	case 504:
		return "Gateway timeout. The server took too long to respond."
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "Client error occurred. Please check your request parameters."
		} else if statusCode >= 500 {
			return "Server error occurred. Please try again later."
		}
		return "An unexpected error occurred."
	}
}

// IsAPIError checks if the error is an APIError.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// IsAuthError checks if the error is an authentication error (401).
func IsAuthError(err error) bool {
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsAuthError()
}

// IsForbiddenError checks if the error is a forbidden error (403).
func IsForbiddenError(err error) bool {
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsForbiddenError()
}

// IsNotFoundError checks if the error is a not found error (404).
func IsNotFoundError(err error) bool {
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsNotFoundError()
}

// IsServerError checks if the error is a server error (5xx).
func IsServerError(err error) bool {
	apiErr, ok := IsAPIError(err)
	return ok && apiErr.IsServerError()
}
