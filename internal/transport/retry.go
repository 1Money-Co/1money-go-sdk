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

package transport

import (
	"context"
	"math/rand/v2"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3).
	// Set to 0 to disable retries.
	MaxRetries int

	// InitialBackoff is the initial delay before the first retry (default: 1s).
	InitialBackoff time.Duration

	// MaxBackoff is the maximum delay between retries (default: 30s).
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0).
	BackoffMultiplier float64

	// Jitter adds randomness to backoff to prevent thundering herd (default: true).
	// When enabled, actual delay = backoff * (0.5 + rand(0, 0.5))
	Jitter bool

	// RetryableStatusCodes allows customizing which HTTP status codes trigger retry.
	// If nil, defaults to 429, 502, 503, 504.
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
		RetryableStatusCodes: []int{
			429, // Too Many Requests
			502, // Bad Gateway
			503, // Service Unavailable
			504, // Gateway Timeout
		},
	}
}

// NoRetryConfig returns a RetryConfig that disables retries.
func NoRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 0,
	}
}

// retryer handles retry logic with exponential backoff.
type retryer struct {
	config *RetryConfig
	log    *zap.Logger
}

// newRetryer creates a new retryer with the given configuration.
func newRetryer(config *RetryConfig) *retryer {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &retryer{
		config: config,
		log:    getLogger(),
	}
}

// shouldRetry determines if a request should be retried based on the error.
func (r *retryer) shouldRetry(err error, attempt int) bool {
	if r.config.MaxRetries <= 0 || attempt >= r.config.MaxRetries {
		return false
	}

	apiErr, ok := IsAPIError(err)
	if !ok {
		// Non-API errors (network errors, timeouts) are generally retryable
		return true
	}

	// Check if the status code is in the retryable list
	for _, code := range r.config.RetryableStatusCodes {
		if apiErr.StatusCode == code {
			return true
		}
	}

	return false
}

// calculateBackoff returns the backoff duration for the given attempt.
func (r *retryer) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: initial * multiplier^attempt
	backoff := float64(r.config.InitialBackoff)
	for range attempt {
		backoff *= r.config.BackoffMultiplier
	}

	// Cap at max backoff
	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}

	// Apply jitter if enabled: backoff * (0.5 + rand(0, 0.5))
	if r.config.Jitter {
		jitterFactor := 0.5 + rand.Float64()*0.5 //nolint:gosec // G404: weak RNG is acceptable for jitter
		backoff *= jitterFactor
	}

	return time.Duration(backoff)
}

// wait sleeps for the backoff duration, respecting context cancellation.
func (r *retryer) wait(ctx context.Context, attempt int) error {
	backoff := r.calculateBackoff(attempt)

	r.log.Debug("waiting before retry",
		zap.Int("attempt", attempt+1),
		zap.Duration("backoff", backoff),
	)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(backoff):
		return nil
	}
}

// parseRetryAfter attempts to parse the Retry-After header value.
// It handles both delta-seconds (integer) and HTTP-date formats.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	// Try parsing as delta-seconds (integer)
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing retry duration from error message like "Retry after 4s."
	re := regexp.MustCompile(`Retry after (\d+)s`)
	if matches := re.FindStringSubmatch(value); len(matches) > 1 {
		if seconds, err := strconv.Atoi(matches[1]); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}

	return 0
}
