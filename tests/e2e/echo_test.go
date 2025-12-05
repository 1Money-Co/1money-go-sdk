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
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
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

// TestRateLimiter tests that the rate limiter is working correctly for both API key and user limits.
// Each test case uses a different API key to ensure independent rate limit testing.
func (s *EchoTestSuite) TestRateLimiter() {
	s.T().Skip()
	// Create secondary client with different API key for independent rate limit testing
	secondaryClient, err := onemoney.NewClient(&onemoney.Config{
		BaseURL:   os.Getenv("ONEMONEY_BASE_URL"),
		AccessKey: "ALWUT3B11PGRKDDU05IR",
		SecretKey: "ewhWWIO7JuEHNHj3maIM2x-ghN1vCbxlXNcnANqoyL8",
	})
	s.Require().NoError(err, "failed to create secondary client")

	testCases := []struct {
		name         string
		burstSize    int
		ratePerSec   int
		extraRequest int
		client       *onemoney.Client
	}{
		{
			name:         "APIKey",
			burstSize:    2,
			ratePerSec:   1,
			extraRequest: 1,
			client:       s.Client,
		},
		{
			name:         "User",
			burstSize:    4,
			ratePerSec:   2,
			extraRequest: 1,
			client:       secondaryClient,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			client := tc.client
			totalRequest := tc.burstSize + tc.extraRequest*2

			s.T().Logf("Testing %s rate limiter (burst: %d, rate: %d/sec) with %d concurrent requests...",
				tc.name, tc.burstSize, tc.ratePerSec, totalRequest)

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
					resp, err := client.Echo.Get(s.Ctx)
					res := result{index: index + 1}
					if err != nil {
						res.err = err
						if errors.Is(err, transport.ErrRateLimited) {
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

			s.T().Logf("Rate limiter test results for %s:", tc.name)
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
			// Wait for rate limiter to fully reset (server typically needs ~10s)
			time.Sleep(10 * time.Second)

			resp, err := client.Echo.Post(s.Ctx, &echo.Request{Message: "After reset"})
			s.Require().NoError(err, "Request should succeed after rate limiter reset")
			s.Require().NotNil(resp, "Response should not be nil")
			s.T().Logf("After reset: Successfully sent request - %s", resp.Message)
		})
	}
}

// TestEchoTestSuite runs the echo test suite.
func TestEchoTestSuite(t *testing.T) {
	suite.Run(t, new(EchoTestSuite))
}
