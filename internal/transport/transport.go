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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"

	onemoney "github.com/1Money-Co/1money-go-sdk"
	"github.com/1Money-Co/1money-go-sdk/internal/auth"
)

// Request represents an HTTP request to be sent.
type Request struct {
	Method      string
	Path        string
	Body        []byte
	Headers     map[string]string
	QueryParams map[string]string
}

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Status     string
	Body       []byte
	Headers    http.Header
}

// GenericResponse represents the standard API response wrapper.
// It encapsulates the response code, message, and typed data.
type GenericResponse[T any] struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// Transport handles HTTP communication with the API.
type Transport struct {
	baseURL       string
	httpClient    *http.Client
	authenticator auth.Authenticator
}

// Config holds transport configuration.
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewTransport creates a new HTTP transport with the given configuration.
func NewTransport(cfg *Config, authenticator auth.Authenticator) *Transport {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				Proxy: nil, // Disable proxy for local testing
			},
		}
	}

	return &Transport{
		baseURL:       cfg.BaseURL,
		httpClient:    httpClient,
		authenticator: authenticator,
	}
}

// Do executes an HTTP request with automatic authentication.
func (t *Transport) Do(ctx context.Context, req *Request) (*Response, error) {
	log := getLogger()

	// Generate authentication headers
	sigResult, err := t.authenticator.Authenticate(req.Method, req.Path, req.Body)
	if err != nil {
		log.Error("failed to sign request",
			zap.String("method", req.Method),
			zap.String("path", req.Path),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Build HTTP request
	httpReq, err := t.buildHTTPRequest(ctx, req, sigResult)
	if err != nil {
		log.Error("failed to build HTTP request",
			zap.String("method", req.Method),
			zap.String("path", req.Path),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// Log request
	log.Debug("executing HTTP request",
		zap.String("method", req.Method),
		zap.String("url", httpReq.URL.String()),
		zap.Int("body_size", len(req.Body)),
	)

	// Print curl command separately for easy copy-paste
	if os.Getenv("ONEMONEY_DEBUG_CURL") == "1" {
		fmt.Fprintln(os.Stderr, buildCurlCommand(httpReq, req.Body))
	}

	// Execute request
	httpResp, err := t.httpClient.Do(httpReq)
	if err != nil {
		log.Error("failed to execute HTTP request",
			zap.String("method", req.Method),
			zap.String("path", req.Path),
			zap.String("url", httpReq.URL.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	log.Debug("received HTTP response",
		zap.Int("status_code", httpResp.StatusCode),
		zap.String("status", httpResp.Status),
	)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Error("failed to read response body",
			zap.Int("status_code", httpResp.StatusCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP error status codes
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Try to parse response as JSON for better logging
		logFields := []zap.Field{
			zap.Int("status_code", httpResp.StatusCode),
			zap.String("status", httpResp.Status),
			zap.String("method", req.Method),
			zap.String("path", req.Path),
		}

		// Attempt to parse and log response body as structured data
		if len(respBody) > 0 && respBody[0] == '{' {
			var responseData map[string]any
			if err := json.Unmarshal(respBody, &responseData); err == nil {
				// Successfully parsed as JSON, log as structured object
				logFields = append(logFields, zap.Any("response", responseData))
			} else {
				// Failed to parse, log as string
				logFields = append(logFields, zap.String("response_body", string(respBody)))
			}
		} else {
			// Not JSON, log as string
			logFields = append(logFields, zap.String("response_body", string(respBody)))
		}

		log.Warn("received error status code", logFields...)

		// Parse and return API error
		apiErr := parseErrorResponse(httpResp.StatusCode, httpResp.Status, respBody)
		return nil, apiErr
	}

	// Check for rate limit response embedded in HTTP 200
	// Some APIs return HTTP 200 with rate limit info in body:
	// {"code":"Too_Many_Requests","status":429,"detail":"..."}
	if apiErr := checkEmbeddedRateLimitError(respBody); apiErr != nil {
		log.Warn("detected embedded rate limit response",
			zap.Int("http_status", httpResp.StatusCode),
			zap.String("code", apiErr.Code),
			zap.String("detail", apiErr.Detail),
		)
		return nil, apiErr
	}

	log.Debug("request completed successfully",
		zap.Int("status_code", httpResp.StatusCode),
		zap.Int("response_size", len(respBody)),
		zap.String("resp", string(respBody)),
	)

	return &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Body:       respBody,
		Headers:    httpResp.Header,
	}, nil
}

// buildHTTPRequest constructs an http.Request from a transport.Request.
func (t *Transport) buildHTTPRequest(ctx context.Context, req *Request, sigResult *auth.SignatureResult) (*http.Request, error) {
	url := t.baseURL + req.Path

	// Add query parameters if any
	if len(req.QueryParams) > 0 {
		url += t.buildQueryString(req.QueryParams)
	}

	// Create request with body if present
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set User-Agent header with SDK version information
	userAgent := fmt.Sprintf("OneMoney-Go-SDK/%s (Go/%s; %s/%s)",
		onemoney.Version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
	httpReq.Header.Set("User-Agent", userAgent)

	// Set authentication headers
	httpReq.Header.Set(auth.HeaderAuthorization, sigResult.Authorization)
	httpReq.Header.Set(auth.HeaderDate, sigResult.Timestamp)

	// Set content type for requests with body
	if len(req.Body) > 0 {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add X-Forwarded-For header in debug mode for testing rate limiting
	if os.Getenv("ONEMONEY_DEBUG") == "1" {
		if localIP := getLocalIP(); localIP != "" {
			httpReq.Header.Set("X-Forwarded-For", localIP)
		}
	}

	return httpReq, nil
}

// getLocalIP retrieves the local IP address of the machine.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

// buildQueryString constructs a query string from parameters.
func (*Transport) buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	return "?" + joinStrings(parts, "&")
}

// joinStrings joins string slices with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// buildCurlCommand generates a curl command string from an HTTP request for debugging.
func buildCurlCommand(req *http.Request, body []byte) string {
	var lines []string
	lines = append(lines, "curl -v")

	// Add method
	if req.Method != http.MethodGet {
		lines = append(lines, fmt.Sprintf("  -X %s", req.Method))
	}

	// Add headers
	for key, values := range req.Header {
		for _, value := range values {
			lines = append(lines, fmt.Sprintf("  -H '%s: %s'", key, value))
		}
	}

	// Add body
	if len(body) > 0 {
		lines = append(lines, fmt.Sprintf("  -d '%s'", string(body)))
	}

	// Add URL
	lines = append(lines, fmt.Sprintf("  '%s'", req.URL.String()))

	return joinStrings(lines, " \\\n")
}
