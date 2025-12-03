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

	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
)

// TransactionsTestSuite tests transactions service operations.
type TransactionsTestSuite struct {
	E2ETestSuite
}

// TestTransactions_ListTransactions tests listing all transactions for a customer.
func (s *TransactionsTestSuite) TestTransactions_ListTransactions() {
	resp, err := s.Client.Transactions.ListTransactions(s.Ctx, testCustomerID, nil)
	s.Require().NoError(err, "ListTransactions should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Listed %d transactions (total: %d)", len(resp.List), resp.Total)
	if len(resp.List) > 0 {
		s.T().Logf("First transaction:\n%s", PrettyJSON(resp.List[0]))
	}
}

// TestTransactions_ListTransactions_WithPagination tests listing transactions with pagination.
func (s *TransactionsTestSuite) TestTransactions_ListTransactions_WithPagination() {
	req := &transactions.ListTransactionsRequest{
		Page: 1,
		Size: 5,
	}

	resp, err := s.Client.Transactions.ListTransactions(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "ListTransactions should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.LessOrEqual(len(resp.List), 5, "Should return at most 5 transactions")
	s.T().Logf("Listed %d transactions with pagination (total: %d)", len(resp.List), resp.Total)
}

// TestTransactions_ListTransactions_FilterByAsset tests filtering transactions by asset.
func (s *TransactionsTestSuite) TestTransactions_ListTransactions_FilterByAsset() {
	req := &transactions.ListTransactionsRequest{
		Asset: assets.AssetNameUSD,
	}

	resp, err := s.Client.Transactions.ListTransactions(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "ListTransactions should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Listed %d USD transactions", len(resp.List))
	for i := range resp.List {
		s.T().Logf("Transaction %s: %s %s", resp.List[i].TransactionID, resp.List[i].Amount, resp.List[i].Asset)
	}
}

// TestTransactions_GetTransaction tests retrieving a specific transaction.
func (s *TransactionsTestSuite) TestTransactions_GetTransaction() {
	// First list transactions to get an ID
	listResp, err := s.Client.Transactions.ListTransactions(s.Ctx, testCustomerID, nil)
	s.Require().NoError(err, "ListTransactions should succeed")

	if len(listResp.List) == 0 {
		s.T().Skip("No transactions available to retrieve")
	}

	// Get the first transaction
	transactionID := listResp.List[0].TransactionID

	resp, err := s.Client.Transactions.GetTransaction(s.Ctx, testCustomerID, transactionID)
	s.Require().NoError(err, "GetTransaction should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(transactionID, resp.TransactionID, "TransactionID should match")
	s.T().Logf("Retrieved transaction:\n%s", PrettyJSON(resp))
}

// TestTransactionsTestSuite runs the transactions test suite.
func TestTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionsTestSuite))
}
