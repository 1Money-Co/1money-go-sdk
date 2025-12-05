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
	CustomerDependentTestSuite
}

// TestTransactions_List tests listing transactions with various scenarios.
func (s *TransactionsTestSuite) TestTransactions_List() {
	s.Run("Empty", func() {
		// For a fresh customer, there might be no transactions initially
		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListTransactions should succeed even with no transactions")
		s.Require().NotNil(resp, "Response should not be nil")
		s.T().Logf("Transactions list: %d transactions (total: %d)", len(resp.List), resp.Total)
	})

	s.Run("WithData", func() {
		// Ensure we have at least one transaction
		_, err := s.EnsureTransaction()
		s.Require().NoError(err, "EnsureTransaction should succeed")

		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListTransactions should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		s.Require().NotEmpty(resp.List, "Should have at least one transaction")
		s.T().Logf("Listed %d transactions (total: %d)", len(resp.List), resp.Total)
		s.T().Logf("First transaction:\n%s", PrettyJSON(resp.List[0]))
	})

	s.Run("WithPagination", func() {
		// Ensure we have at least one transaction
		_, err := s.EnsureTransaction()
		s.Require().NoError(err, "EnsureTransaction should succeed")

		req := &transactions.ListTransactionsRequest{
			Page: 1,
			Size: 5,
		}

		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, req)
		s.Require().NoError(err, "ListTransactions should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		s.LessOrEqual(len(resp.List), 5, "Should return at most 5 transactions")
		s.T().Logf("Listed %d transactions with pagination (total: %d)", len(resp.List), resp.Total)
	})

	s.Run("FilterByAsset", func() {
		// Ensure we have at least one USD transaction (EnsureTransaction creates a USD deposit)
		_, err := s.EnsureTransaction()
		s.Require().NoError(err, "EnsureTransaction should succeed")

		req := &transactions.ListTransactionsRequest{
			Asset: assets.AssetNameUSD,
		}

		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, req)
		s.Require().NoError(err, "ListTransactions should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		s.Require().NotEmpty(resp.List, "Should have at least one USD transaction")
		s.T().Logf("Listed %d USD transactions", len(resp.List))
		for i := range resp.List {
			s.T().Logf("Transaction %s: %s %s", resp.List[i].TransactionID, resp.List[i].Amount, resp.List[i].Asset)
		}
	})
}

// TestTransactions_GetTransaction tests retrieving a specific transaction.
func (s *TransactionsTestSuite) TestTransactions_GetTransaction() {
	// Ensure we have at least one transaction
	transactionID, err := s.EnsureTransaction()
	s.Require().NoError(err, "EnsureTransaction should succeed")

	// Get the transaction
	resp, err := s.Client.Transactions.GetTransaction(s.Ctx, s.CustomerID, transactionID)
	s.Require().NoError(err, "GetTransaction should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(transactionID, resp.TransactionID, "TransactionID should match")
	s.T().Logf("Retrieved transaction:\n%s", PrettyJSON(resp))
}

// TestTransactionsTestSuite runs the transactions test suite.
func TestTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionsTestSuite))
}
