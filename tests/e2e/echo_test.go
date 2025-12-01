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

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/echo"
)

// EchoTestSuite tests echo service operations.
type EchoTestSuite struct {
	E2ETestSuite
}

func (s *EchoTestSuite) TestEchoService_Get() {
	resp, err := s.Client.Echo.Get(s.Ctx)
	s.Require().NoError(err, "Echo Get should succeed")
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Echo response:\n%s", PrettyJSON(resp))
}

func (s *EchoTestSuite) TestEchoService_Post() {
	resp, err := s.Client.Echo.Post(s.Ctx, &echo.Request{Message: "Hello, World!"})
	s.Require().NoError(err, "Echo Post should succeed")
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Echo response:\n%s", PrettyJSON(resp))
}

// TestRateLimiter_IPBasedLimiting tests that the IP-based rate limiter is working correctly.
func (s *EchoTestSuite) TestRateLimiter_IPBasedLimiting() {
	const (
		burstSize    = 10
		extraRequest = 5
		totalRequest = burstSize + extraRequest
	)

	s.T().Log("Testing rate limiter with concurrent requests...")

	type result struct {
		index       int
		success     bool
		rateLimited bool
		err         error
		responseMsg string
	}

	resultChan := make(chan result, totalRequest)

	for i := range totalRequest {
		go func(index int) {
			resp, err := s.Client.Echo.Post(s.Ctx, &echo.Request{
				Message: fmt.Sprintf("Rate limit test message #%d", index+1),
			})

			res := result{index: index + 1}
			if err != nil {
				res.err = err
				if containsRateLimitError(err.Error()) {
					res.rateLimited = true
				}
			} else {
				res.success = true
				res.responseMsg = resp.Message
			}
			resultChan <- res
		}(i)
	}

	successCount := 0
	rateLimitedCount := 0
	unexpectedErrors := 0

	for range totalRequest {
		res := <-resultChan
		if res.success {
			successCount++
			s.T().Logf("Request #%d: Success - %s", res.index, res.responseMsg)
		} else if res.rateLimited {
			rateLimitedCount++
			s.T().Logf("Request #%d: Rate limited (expected after burst)", res.index)
		} else {
			unexpectedErrors++
			s.T().Logf("Request #%d: Unexpected error: %v", res.index, res.err)
		}
	}
	close(resultChan)

	s.T().Logf("Rate limiter test results:")
	s.T().Logf("  Total requests: %d", totalRequest)
	s.T().Logf("  Successful: %d", successCount)
	s.T().Logf("  Rate limited: %d", rateLimitedCount)
	s.T().Logf("  Unexpected errors: %d", unexpectedErrors)

	s.Positive(successCount, "Should have at least some successful requests")
	s.Positive(rateLimitedCount, "Should have at least one rate-limited request when exceeding burst size")
	s.Equal(totalRequest, successCount+rateLimitedCount+unexpectedErrors,
		"Total requests should equal successful + rate limited + unexpected errors")
	s.Equal(0, unexpectedErrors, "Should not have unexpected errors")

	s.T().Log("Waiting for rate limiter to reset...")
	time.Sleep(1500 * time.Millisecond)

	resp, err := s.Client.Echo.Post(s.Ctx, &echo.Request{Message: "After reset"})
	s.Require().NoError(err, "Request should succeed after rate limiter reset")
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("After reset: Successfully sent request - %s", resp.Message)
}

// containsRateLimitError checks if an error message indicates a rate limit error.
func containsRateLimitError(errMsg string) bool {
	indicators := []string{
		"429",
		"Too Many Requests",
		"rate limit",
		"too many requests",
		"throttle",
	}

	errMsgLower := toLower(errMsg)
	for _, indicator := range indicators {
		if contains(errMsgLower, toLower(indicator)) {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestEchoTestSuite runs the echo test suite.
func TestEchoTestSuite(t *testing.T) {
	suite.Run(t, new(EchoTestSuite))
}
