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

// Package credentials provides credential management similar to AWS SDK.
// It supports multiple sources: static credentials, environment variables, and config files.
package credentials

import (
	"errors"
	"fmt"
)

var (
	// ErrNoCredentials is returned when no credentials are found.
	ErrNoCredentials = errors.New("no credentials found")

	// ErrInvalidCredentials is returned when credentials are found but invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Credentials represents API credentials.
type Credentials struct {
	AccessKey string
	SecretKey string
	BaseURL   string
}

// IsValid returns true if the credentials are valid (non-empty).
func (c *Credentials) IsValid() bool {
	return c.AccessKey != "" && c.SecretKey != ""
}

// Provider is the interface for credential providers.
// Each provider attempts to retrieve credentials from a specific source.
type Provider interface {
	// Retrieve attempts to retrieve credentials from the provider's source.
	// Returns ErrNoCredentials if credentials are not available from this source.
	Retrieve() (*Credentials, error)

	// Name returns the name of this provider for debugging purposes.
	Name() string
}

// ProviderError wraps provider errors with additional context.
type ProviderError struct {
	Provider string
	Err      error
	Message  string
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %v", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Provider, e.Err)
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	return e.Err
}
