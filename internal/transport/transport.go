// Package transport provides HTTP transport functionality for the SDK.
package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

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

// Transport handles HTTP communication with the API.
type Transport struct {
	baseURL    string
	httpClient *http.Client
	signer     *auth.Signer
}

// Config holds transport configuration.
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewTransport creates a new HTTP transport with the given configuration.
func NewTransport(cfg *Config, signer *auth.Signer) *Transport {
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
		baseURL:    cfg.BaseURL,
		httpClient: httpClient,
		signer:     signer,
	}
}

// Do executes an HTTP request with automatic signature generation.
func (t *Transport) Do(ctx context.Context, req *Request) (*Response, error) {
	// Generate signature
	sigResult, err := t.signer.SignRequest(req.Method, req.Path, req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Build HTTP request
	httpReq, err := t.buildHTTPRequest(ctx, req, sigResult)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// Execute request
	httpResp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP error status codes
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Parse and return API error
		apiErr := parseErrorResponse(httpResp.StatusCode, httpResp.Status, respBody)
		return nil, apiErr
	}

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

	return httpReq, nil
}

// buildQueryString constructs a query string from parameters.
func (t *Transport) buildQueryString(params map[string]string) string {
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
