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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
)

// ExternalAccountsTestSuite tests external accounts service operations.
type ExternalAccountsTestSuite struct {
	CustomerDependentTestSuite
}

// TestExternalAccounts_List tests listing external accounts with various scenarios.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_List() {
	s.Run("Empty", func() {
		// For a fresh customer, listing should succeed even with no accounts
		resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListExternalAccounts should succeed even with no accounts")
		s.Require().NotNil(resp, "Response should not be nil")
		s.T().Logf("External accounts list: %d accounts", len(resp))
	})

	s.Run("WithData", func() {
		// Ensure we have at least one external account
		_, err := s.EnsureExternalAccount()
		if err != nil && strings.Contains(err.Error(), "verified fiat account") {
			s.T().Skip("Skipping: customer doesn't have a verified fiat account yet")
		}
		s.Require().NoError(err, "EnsureExternalAccount should succeed")

		resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListExternalAccounts should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		s.Require().NotEmpty(resp, "Should have at least one external account")
		s.T().Logf("External accounts list:\n%s", PrettyJSON(resp))

		for i := range resp {
			s.NotEmpty(resp[i].ExternalAccountID, "External account ID should not be empty")
			s.NotEmpty(resp[i].CustomerID, "Customer ID should not be empty")
			s.NotEmpty(resp[i].Network, "Network should not be empty")
			s.NotEmpty(resp[i].Currency, "Currency should not be empty")
			s.NotEmpty(resp[i].InstitutionName, "Institution name should not be empty")
			s.NotEmpty(resp[i].Status, "Status should not be empty")
		}
	})

	s.Run("FilterByNetwork", func() {
		networks := []external_accounts.BankNetworkName{
			external_accounts.BankNetworkNameUSACH,
			external_accounts.BankNetworkNameSWIFT,
			external_accounts.BankNetworkNameUSFEDWIRE,
		}

		for _, network := range networks {
			req := &external_accounts.ListReq{Network: network}
			resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, req)
			s.Require().NoError(err, "ListExternalAccounts with network %s should succeed", network)
			s.Require().NotNil(resp, "Response should not be nil")
			s.T().Logf("External accounts with network %s: %d accounts", network, len(resp))

			// Verify all returned accounts match the requested network
			for i := range resp {
				s.Equal(string(network), resp[i].Network, "Network should match filter")
			}
		}
	})
}

// TestExternalAccounts_CreateAndGet tests creating and retrieving an external account.
// Validates all response fields, verifies request fields are reflected in response,
// and polls until the account reaches APPROVED status.
// Note: This test may be skipped if the customer doesn't have a verified fiat account yet.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_CreateAndGet() {
	const (
		pollInterval = 2 * time.Second
		maxWaitTime  = 10 * time.Second
	)

	createReq := FakeExternalAccountRequest()

	// Create external account
	createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, createReq)

	// Skip test if fiat account is not yet verified (400 error)
	if err != nil {
		var apiErr *transport.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 400 &&
			strings.Contains(apiErr.Detail, "verified fiat account is required") {
			s.T().Skip("Skipping: customer doesn't have a verified fiat account yet")
		}
	}
	s.Require().NoError(err, "CreateExternalAccount should succeed")

	// Validate create response structure
	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.ExternalAccountID, "External account ID should not be empty")
	s.Equal(s.CustomerID, createResp.CustomerID, "CustomerID should match")
	s.NotEmpty(createResp.Status, "Status should not be empty")
	s.NotEmpty(createResp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(createResp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate request fields are reflected in response
	s.Equal(string(createReq.Network), createResp.Network, "Network should match request")
	s.Equal(string(createReq.Currency), createResp.Currency, "Currency should match request")
	s.Equal(createReq.InstitutionName, createResp.InstitutionName, "InstitutionName should match request")
	s.Equal(string(createReq.CountryCode), createResp.CountryCode, "CountryCode should match request")

	s.T().Logf("Created external account: %s (status: %s)", createResp.ExternalAccountID, createResp.Status)

	// Poll until approved or failed
	accountID := createResp.ExternalAccountID
	deadline := time.Now().Add(maxWaitTime)
	var finalStatus string

	for time.Now().Before(deadline) {
		acc, err := s.Client.ExternalAccounts.GetExternalAccount(s.Ctx, s.CustomerID, accountID)
		s.Require().NoError(err, "GetExternalAccount should succeed during polling")

		finalStatus = acc.Status
		s.T().Logf("Polling external account %s: status=%s", accountID, finalStatus)

		switch finalStatus {
		case string(external_accounts.BankAccountStatusAPPROVED):
			s.T().Logf("External account approved after polling")
			goto approved
		case string(external_accounts.BankAccountStatusFAILED):
			s.Require().Fail("External account approval failed")
		}

		time.Sleep(pollInterval)
	}
	s.Require().Fail("External account approval timed out after %v (final status: %s)", maxWaitTime, finalStatus)

approved:
	// List
	resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
	s.Require().NoError(err, "ListExternalAccounts should succeed")
	s.Require().NotNil(resp, "Response should not be nil")
	s.Require().NotEmpty(resp, "Should have at least one external account")
	s.T().Logf("External accounts list count: %d", len(resp))

	// Get external account by ID
	getResp, err := s.Client.ExternalAccounts.GetExternalAccount(s.Ctx, s.CustomerID, accountID)
	s.Require().NoError(err, "GetExternalAccount should succeed")

	// Validate retrieved account matches created one
	s.Require().NotNil(getResp, "Get response should not be nil")
	s.Equal(accountID, getResp.ExternalAccountID, "External account IDs should match")
	s.Equal(createResp.Network, getResp.Network, "Network should match")
	s.Equal(createResp.Currency, getResp.Currency, "Currency should match")
	s.Equal(createResp.InstitutionName, getResp.InstitutionName, "InstitutionName should match")
	s.Equal(string(external_accounts.BankAccountStatusAPPROVED), getResp.Status, "Status should be APPROVED")

	s.T().Logf("Retrieved external account:\n%s", PrettyJSON(getResp))

	// Get external account by idempotency key
	getByKeyResp, err := s.Client.ExternalAccounts.GetExternalAccountByIdempotencyKey(s.Ctx, s.CustomerID, createReq.IdempotencyKey)
	s.Require().NoError(err, "GetExternalAccountByIdempotencyKey should succeed")

	// Validate retrieved account matches created one
	s.Require().NotNil(getByKeyResp, "Get by key response should not be nil")
	s.Equal(accountID, getByKeyResp.ExternalAccountID, "External account IDs should match")
	s.Equal(createResp.Network, getByKeyResp.Network, "Network should match")
	s.Equal(createResp.Currency, getByKeyResp.Currency, "Currency should match")

	s.T().Logf("Retrieved external account by idempotency key:\n%s", PrettyJSON(getByKeyResp))
}

// TestExternalAccounts_Delete tests deleting an external account.
// Validates account is no longer retrievable after deletion.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_Delete() {
	// First create an account to delete
	createReq := FakeExternalAccountRequest()

	createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateExternalAccount should succeed")

	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.ExternalAccountID, "External account ID should not be empty")
	s.T().Logf("Created external account for deletion: %s", createResp.ExternalAccountID)

	// Delete the account
	err = s.Client.ExternalAccounts.RemoveExternalAccount(s.Ctx, s.CustomerID, createResp.ExternalAccountID)
	s.Require().NoError(err, "DeleteExternalAccount should succeed")

	s.T().Logf("Successfully deleted external account: %s", createResp.ExternalAccountID)

	// Verify account is no longer in the list
	listResp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
	s.Require().NoError(err, "ListExternalAccounts should succeed")

	// Verify deleted account is not in the list
	for i := range listResp {
		s.NotEqual(createResp.ExternalAccountID, listResp[i].ExternalAccountID,
			"Deleted account should not appear in list")
	}
}

// TestExternalAccountsTestSuite runs the external accounts test suite.
func TestExternalAccountsTestSuite(t *testing.T) {
	suite.Run(t, new(ExternalAccountsTestSuite))
}
