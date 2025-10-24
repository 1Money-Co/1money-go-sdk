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
	"time"

	onemoney "github.com/1Money-Co/1money-go-sdk"
	"github.com/1Money-Co/1money-go-sdk/internal/auth"
	"github.com/1Money-Co/1money-go-sdk/internal/credentials"
	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/echo"
)

// Client is the main OneMoney API client.
// It provides access to all service modules through a clean interface.
type Client struct {
	transport *transport.Transport

	// Service modules
	Echo     echo.Service
	Customer customer.Service
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

	// HTTPClient is an optional custom HTTP client
	HTTPClient *http.Client

	// Timeout is the request timeout (default: 30 seconds)
	Timeout time.Duration
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
		cfg = &Config{}
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
		cfg.BaseURL = "http://localhost:9000"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Create auth credentials and signer
	authCreds := auth.NewCredentials(creds.AccessKey, creds.SecretKey)
	signer := auth.NewSigner(authCreds)

	// Create transport
	transportCfg := &transport.Config{
		BaseURL:    cfg.BaseURL,
		HTTPClient: cfg.HTTPClient,
		Timeout:    cfg.Timeout,
	}
	tr := transport.NewTransport(transportCfg, signer)

	// Initialize all service modules with base service
	base := svc.NewBaseService(tr)
	echoSvc := echo.NewService(base)
	customerSvc := customer.NewService(base)

	// Create client with pre-initialized services
	return &Client{
		transport: tr,
		Echo:      echoSvc,
		Customer:  customerSvc,
	}, nil
}

// Version returns the SDK version.
// This can be used for logging, debugging, or telemetry purposes.
func (c *Client) Version() string {
	return onemoney.Version
}
