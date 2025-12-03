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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
)

// ExternalAccountsTestSuite tests external accounts service operations.
type ExternalAccountsTestSuite struct {
	E2ETestSuite
}

// TestExternalAccounts_ListAll tests listing all external accounts for a customer.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_ListAll() {
	resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, testCustomerID, nil)
	s.Require().NoError(err, "ListExternalAccounts should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("External accounts list:\n%s", PrettyJSON(resp))

	for i := range resp {
		s.NotEmpty(resp[i].ExternalAccountID, "External account ID should not be empty")
		s.NotEmpty(resp[i].CustomerID, "Customer ID should not be empty")
		s.NotEmpty(resp[i].BankNetworkName, "Bank network name should not be empty")
		s.NotEmpty(resp[i].Currency, "Currency should not be empty")
		s.NotEmpty(resp[i].BankName, "Bank name should not be empty")
		s.NotEmpty(resp[i].Status, "Status should not be empty")
	}
}

// TestExternalAccounts_ListByStatus tests listing external accounts filtered by status.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_ListByStatus() {
	req := &external_accounts.ListExternalAccountsRequest{
		Status: external_accounts.BankAccountStatusAPPROVED,
	}

	resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "ListExternalAccounts should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Approved external accounts:\n%s", PrettyJSON(resp))

	for i := range resp {
		s.Equal("APPROVED", resp[i].Status, "Status should be APPROVED")
	}
}

// TestExternalAccounts_CreateAndGet tests creating and retrieving an external account.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_CreateAndGet() {
	idempotencyKey := uuid.New().String()

	createReq := &external_accounts.CreateExternalAccountRequest{
		IdempotencyKey:       idempotencyKey,
		BankNetworkName:      external_accounts.BankNetworkNameUSACH,
		Currency:             external_accounts.CurrencyUSD,
		BankName:             "Test Bank",
		BankAccountOwnerName: "Test Account Owner",
		BankAccountNumber:    "123456789",
		BankRoutingNumber:    "021000021",
	}

	// Create external account
	createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, testCustomerID, createReq)
	s.Require().NoError(err, "CreateExternalAccount should succeed")

	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.ExternalAccountID, "External account ID should not be empty")
	s.T().Logf("Created external account:\n%s", PrettyJSON(createResp))

	// Get external account by ID
	getResp, err := s.Client.ExternalAccounts.GetExternalAccount(s.Ctx, testCustomerID, createResp.ExternalAccountID)
	s.Require().NoError(err, "GetExternalAccount should succeed")

	s.Require().NotNil(getResp, "Get response should not be nil")
	s.Equal(createResp.ExternalAccountID, getResp.ExternalAccountID, "External account IDs should match")
	s.T().Logf("Retrieved external account:\n%s", PrettyJSON(getResp))

	// Get external account by idempotency key
	getByKeyResp, err := s.Client.ExternalAccounts.GetExternalAccountByIdempotencyKey(s.Ctx, testCustomerID, idempotencyKey)
	s.Require().NoError(err, "GetExternalAccountByIdempotencyKey should succeed")

	s.Require().NotNil(getByKeyResp, "Get by key response should not be nil")
	s.Equal(createResp.ExternalAccountID, getByKeyResp.ExternalAccountID, "External account IDs should match")
	s.T().Logf("Retrieved external account by idempotency key:\n%s", PrettyJSON(getByKeyResp))
}

// TestExternalAccounts_Delete tests deleting an external account.
func (s *ExternalAccountsTestSuite) TestExternalAccounts_Delete() {
	// First create an account to delete
	idempotencyKey := uuid.New().String()

	createReq := &external_accounts.CreateExternalAccountRequest{
		IdempotencyKey:       idempotencyKey,
		BankNetworkName:      external_accounts.BankNetworkNameUSACH,
		Currency:             external_accounts.CurrencyUSD,
		BankName:             "Test Bank to Delete",
		BankAccountOwnerName: "Test Account Owner",
		BankAccountNumber:    "987654321",
		BankRoutingNumber:    "021000021",
	}

	createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, testCustomerID, createReq)
	s.Require().NoError(err, "CreateExternalAccount should succeed")

	s.Require().NotNil(createResp, "Create response should not be nil")
	s.T().Logf("Created external account for deletion: %s", createResp.ExternalAccountID)

	// Delete the account
	err = s.Client.ExternalAccounts.DeleteExternalAccount(s.Ctx, testCustomerID, createResp.ExternalAccountID)
	s.Require().NoError(err, "DeleteExternalAccount should succeed")

	s.T().Logf("Successfully deleted external account: %s", createResp.ExternalAccountID)
}

// TestExternalAccountsTestSuite runs the external accounts test suite.
func TestExternalAccountsTestSuite(t *testing.T) {
	suite.Run(t, new(ExternalAccountsTestSuite))
}
