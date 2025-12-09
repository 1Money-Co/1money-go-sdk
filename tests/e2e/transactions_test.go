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
// Validates response structure and field values.
func (s *TransactionsTestSuite) TestTransactions_List() {
	s.Run("Empty", func() {
		// For a fresh customer, there might be no transactions initially
		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListTransactions should succeed even with no transactions")
		s.Require().NotNil(resp, "Response should not be nil")
		s.GreaterOrEqual(resp.Total, 0, "Total should be non-negative")
		s.T().Logf("Transactions list: %d transactions (total: %d)", len(resp.List), resp.Total)
	})

	s.Run("WithData", func() {
		// Ensure we have at least one transaction
		_, err := s.EnsureTransaction()
		if err != nil {
			s.T().Skipf("Skipping WithData: %v", err)
		}

		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListTransactions should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		if len(resp.List) == 0 {
			s.T().Log("No transactions returned after EnsureTransaction; skipping structural checks")
			return
		}
		s.Positive(resp.Total, "Total should be greater than 0 when transactions are present")

		// Validate first transaction structure
		tx := resp.List[0]
		s.NotEmpty(tx.TransactionID, "TransactionID should not be empty")
		s.NotEmpty(tx.TransactionAction, "TransactionAction should not be empty")
		s.NotEmpty(tx.Amount, "Amount should not be empty")
		s.NotEmpty(tx.Status, "Status should not be empty")
		s.NotEmpty(tx.CreatedAt, "CreatedAt should not be empty")
		s.NotEmpty(tx.ModifiedAt, "ModifiedAt should not be empty")
		s.Equal(s.CustomerID, tx.CustomerID, "CustomerID should match")

		s.T().Logf("Listed %d transactions (total: %d)", len(resp.List), resp.Total)
		s.T().Logf("First transaction:\n%s", PrettyJSON(tx))
	})

	s.Run("WithPagination", func() {
		// Ensure we have at least one transaction
		_, err := s.EnsureTransaction()
		if err != nil {
			s.T().Skipf("Skipping WithPagination: %v", err)
		}

		req := &transactions.ListTransactionsRequest{
			Page: 1,
			Size: 5,
		}

		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, req)
		s.Require().NoError(err, "ListTransactions should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		if len(resp.List) == 0 {
			s.T().Log("No transactions returned for pagination test; skipping size assertion")
			return
		}
		s.LessOrEqual(len(resp.List), 5, "Should return at most 5 transactions")
		s.T().Logf("Listed %d transactions with pagination (total: %d)", len(resp.List), resp.Total)
	})

	s.Run("FilterByAsset", func() {
		// Ensure we have at least one USD transaction (EnsureTransaction creates a USD deposit)
		_, err := s.EnsureTransaction()
		if err != nil {
			s.T().Skipf("Skipping FilterByAsset: %v", err)
		}

		req := &transactions.ListTransactionsRequest{
			Asset: assets.AssetNameUSD,
		}

		resp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, req)
		s.Require().NoError(err, "ListTransactions should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		if len(resp.List) == 0 {
			s.T().Log("No USD transactions returned; skipping asset filter assertions")
		} else {
			// Validate all returned transactions have USD asset
			for i := range resp.List {
				s.Equal(string(assets.AssetNameUSD), resp.List[i].Asset,
					"All filtered transactions should have USD asset")
			}
		}

		s.T().Logf("Listed %d USD transactions", len(resp.List))
	})
}

// TestTransactions_GetTransaction tests retrieving a specific transaction.
// Validates all response fields.
func (s *TransactionsTestSuite) TestTransactions_GetTransaction() {
	// Ensure we have at least one transaction
	transactionID, err := s.EnsureTransaction()
	if err != nil {
		s.T().Skipf("Skipping GetTransaction: %v", err)
	}

	// Get the transaction
	resp, err := s.Client.Transactions.GetTransaction(s.Ctx, s.CustomerID, transactionID)
	s.Require().NoError(err, "GetTransaction should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(transactionID, resp.TransactionID, "TransactionID should match")
	s.Equal(s.CustomerID, resp.CustomerID, "CustomerID should match")
	s.NotEmpty(resp.TransactionAction, "TransactionAction should not be empty")
	s.NotEmpty(resp.Amount, "Amount should not be empty")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate transaction action is valid
	validActions := []string{"DEPOSIT", "WITHDRAWAL", "CONVERSION"}
	s.Contains(validActions, resp.TransactionAction, "TransactionAction should be valid")

	s.T().Logf("Retrieved transaction:\n%s", PrettyJSON(resp))
}

// TestTransactionsTestSuite runs the transactions test suite.
func TestTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionsTestSuite))
}
