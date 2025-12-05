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

// Package scp provides the main SDK for OneMoney API.
package onemoney

import (
	"fmt"
	"net/http"
	"os"
	"time"

	onemoney "github.com/1Money-Co/1money-go-sdk"
	"github.com/1Money-Co/1money-go-sdk/internal/auth"
	"github.com/1Money-Co/1money-go-sdk/internal/credentials"
	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/echo"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/instructions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

// Client is the main OneMoney API client.
// It provides access to all service modules through a clean interface.
type Client struct {
	transport *transport.Transport

	// Service modules
	Assets              assets.Service
	AutoConversionRules auto_conversion_rules.Service
	Conversions         conversions.Service
	Customer            customer.Service
	Echo                echo.Service
	ExternalAccounts    external_accounts.Service
	Instructions        instructions.Service
	Simulations         simulations.Service
	Transactions        transactions.Service
	Withdrawals         withdraws.Service
}

// Config holds the client configuration.
// Credentials can be provided in multiple ways (similar to AWS SDK):
// 1. Directly via AccessKey/SecretKey fields (highest priority)
// 2. Environment variables: ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY
// 3. Config file: ~/.onemoney/credentials (with optional Profile)
type Config struct {
	// BaseURL is the API base URL (e.g., "http://localhost:9000")
	// Can also be set via ONEMONEY_BASE_URL environment variable or config file
	BaseURL string

	// AccessKey is the API access key (optional if using env vars or config file)
	AccessKey string

	// SecretKey is the API secret key (optional if using env vars or config file)
	SecretKey string

	// Profile specifies which profile to use from the credentials file
	// (default: "default")
	Profile string

	// Sandbox enables sandbox mode which uses simple Bearer token authentication
	// instead of HMAC signature. In sandbox mode, only AccessKey is required
	// and requests are sent with "Authorization: Bearer {AccessKey}" header.
	Sandbox bool

	// HTTPClient is an optional custom HTTP client
	HTTPClient *http.Client

	// Timeout is the request timeout (default: 30 seconds)
	Timeout time.Duration

	// Retry configures automatic retry behavior for rate limiting and transient errors.
	// If nil, default retry configuration is used (3 retries with exponential backoff).
	// Use NoRetryConfig() to disable retries.
	Retry *RetryConfig
}

// Option is a function that configures the client.
type Option func(*Config)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithBaseURL sets the API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithSandbox enables sandbox mode with simple Bearer token authentication.
func WithSandbox(sandbox bool) Option {
	return func(c *Config) {
		c.Sandbox = sandbox
	}
}

// WithRetry configures the retry behavior for rate limiting and transient errors.
// Pass nil to use default retry configuration, or use NoRetryConfig() to disable retries.
//
// Example with custom retry configuration:
//
//	client, err := onemoney.NewClient(&onemoney.Config{}, onemoney.WithRetry(&onemoney.RetryConfig{
//	    MaxRetries:        5,
//	    InitialBackoff:    500 * time.Millisecond,
//	    MaxBackoff:        60 * time.Second,
//	    BackoffMultiplier: 2.0,
//	    Jitter:            true,
//	}))
//
// Example to disable retries:
//
//	client, err := onemoney.NewClient(&onemoney.Config{}, onemoney.WithRetry(onemoney.NoRetryConfig()))
func WithRetry(retry *RetryConfig) Option {
	return func(c *Config) {
		c.Retry = retry
	}
}

// RetryConfig is an alias for transport.RetryConfig.
// It holds configuration for retry behavior.
type RetryConfig = transport.RetryConfig

// DefaultRetryConfig returns a RetryConfig with sensible defaults:
//   - MaxRetries: 3
//   - InitialBackoff: 1 second
//   - MaxBackoff: 30 seconds
//   - BackoffMultiplier: 2.0
//   - Jitter: true (to prevent thundering herd)
func DefaultRetryConfig() *RetryConfig {
	return transport.DefaultRetryConfig()
}

// NoRetryConfig returns a RetryConfig that disables retries.
func NoRetryConfig() *RetryConfig {
	return transport.NoRetryConfig()
}

// NewClient creates a new OneMoney API client with all services pre-initialized.
//
// Credentials are loaded using a chain of providers (similar to AWS SDK):
// 1. Config fields (AccessKey/SecretKey) - if provided
// 2. Environment variables (ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY)
// 3. Config file (~/.onemoney/credentials) - using the specified Profile
//
// Example with explicit credentials:
//
//	c := scp.NewClient(&scp.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	    BaseURL:   "http://localhost:9000",
//	})
//
// Example with environment variables:
//
//	// Set: export ONEMONEY_ACCESS_KEY=xxx ONEMONEY_SECRET_KEY=yyy
//	c := scp.NewClient(&scp.Config{})
//
// Example with config file:
//
//	// Create ~/.onemoney/credentials with [default] or [production] profile
//	c := scp.NewClient(&scp.Config{Profile: "production"})
//
// Usage:
//
//	resp, err := c.Echo.Get(ctx)
//	resp, err := c.Echo.Post(ctx, &echo.Request{Message: "hello"})
func NewClient(cfg *Config, opts ...Option) (*Client, error) {
	if cfg == nil {
		cfg = &Config{
			Sandbox: os.Getenv(credentials.EnvSandbox) == "1",
			BaseURL: os.Getenv(credentials.EnvBaseURL),
		}
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Load credentials using the provider chain
	provider := credentials.NewDefaultChainProvider(
		cfg.AccessKey,
		cfg.SecretKey,
		cfg.BaseURL,
		cfg.Profile,
		cfg.Sandbox,
	)

	creds, err := provider.Retrieve()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Use BaseURL from credentials if not explicitly set
	if cfg.BaseURL == "" && creds.BaseURL != "" {
		cfg.BaseURL = creds.BaseURL
	}

	// Set defaults
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:9000/openapi"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Create authenticator based on mode (use creds.Sandbox as it may come from env vars)
	var authenticator auth.Authenticator
	if creds.Sandbox {
		// Sandbox mode: use simple Bearer token authentication
		authenticator = auth.NewBearerAuth(creds.AccessKey)
	} else {
		// Production mode: use HMAC signature authentication
		authCreds := auth.NewCredentials(creds.AccessKey, creds.SecretKey)
		authenticator = auth.NewSigner(authCreds)
	}

	// Create transport
	transportCfg := &transport.Config{
		BaseURL:    cfg.BaseURL,
		HTTPClient: cfg.HTTPClient,
		Timeout:    cfg.Timeout,
		Retry:      cfg.Retry,
	}
	tr := transport.NewTransport(transportCfg, authenticator)

	// Initialize all service modules with base service
	base := svc.NewBaseService(tr)

	// Create client with pre-initialized services
	return &Client{
		transport:           tr,
		Assets:              assets.NewService(base),
		AutoConversionRules: auto_conversion_rules.NewService(base),
		Conversions:         conversions.NewService(base),
		Customer:            customer.NewService(base),
		Echo:                echo.NewService(base),
		ExternalAccounts:    external_accounts.NewService(base),
		Instructions:        instructions.NewService(base),
		Simulations:         simulations.NewService(base),
		Transactions:        transactions.NewService(base),
		Withdrawals:         withdraws.NewService(base),
	}, nil
}

// Version returns the SDK version.
// This can be used for logging, debugging, or telemetry purposes.
func (*Client) Version() string {
	return onemoney.Version
}
