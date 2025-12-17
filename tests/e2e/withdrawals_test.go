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
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

// WithdrawalsTestSuite tests withdrawals service operations.
type WithdrawalsTestSuite struct {
	CustomerDependentTestSuite
	externalAccountID string
	testWalletAddress string
}

// SetupSuite prepares balances and external account for withdrawal tests.
func (s *WithdrawalsTestSuite) SetupSuite() {
	s.CustomerDependentTestSuite.SetupSuite()

	// Step 1: Simulate USD deposit for fiat withdrawals
	s.T().Log("Simulating USD deposit...")
	_, err := s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSD,
		Amount:  "500.00",
		Network: simulations.WalletNetworkNameUSACH,
	})
	s.Require().NoError(err, "SimulateDeposit USD should succeed")

	// Step 2: Simulate USDT deposit for crypto withdrawals
	s.T().Log("Simulating USDT deposit...")
	_, err = s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDT,
		Network: simulations.WalletNetworkNameETHEREUM,
		Amount:  "200.00",
	})
	s.Require().NoError(err, "SimulateDeposit USDT should succeed")

	// Step 3: Ensure external account for fiat withdrawals
	s.T().Log("Ensuring external account...")
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")
	s.externalAccountID = externalAccountID

	// Step 4: Load test wallet address for crypto withdrawals
	s.testWalletAddress = os.Getenv("ONEMONEY_TEST_WALLET_ADDRESS")
	s.Require().NotEmpty(s.testWalletAddress, "ONEMONEY_TEST_WALLET_ADDRESS must be set")

	s.T().Log("Withdrawal test setup completed")
}

type withdrawalTestCase struct {
	name    string
	asset   assets.AssetName
	network assets.NetworkName
	amount  string
	isFiat  bool
}

// TestWithdrawals_Flow tests the complete withdrawal flow: Create → GetByID → GetByIdempotencyKey
func (s *WithdrawalsTestSuite) TestWithdrawals_Flow() {
	testCases := []withdrawalTestCase{
		{
			name:    "Fiat_USD_ACH",
			asset:   assets.AssetNameUSD,
			network: assets.NetworkNameUSACH,
			amount:  "10.00",
			isFiat:  true,
		},
		{
			name:    "Crypto_USDT_Ethereum",
			asset:   assets.AssetNameUSDT,
			network: assets.NetworkNameETHEREUM,
			amount:  "10.00",
			isFiat:  false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			idempotencyKey := uuid.New().String()

			// Step 1: Create Withdrawal
			req := &withdraws.CreateWithdrawalRequest{
				IdempotencyKey: idempotencyKey,
				Amount:         tc.amount,
				Asset:          tc.asset,
				Network:        tc.network,
			}

			if tc.isFiat {
				req.ExternalAccountID = s.externalAccountID
			} else {
				req.WalletAddress = s.testWalletAddress
			}

			createResp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
			s.Require().NoError(err, "CreateWithdrawal should succeed")
			s.Require().NotNil(createResp)

			// Validate create response
			s.NotEmpty(createResp.TransactionID)
			s.Equal(idempotencyKey, createResp.IdempotencyKey)
			s.Equal("WITHDRAWAL", createResp.TransactionAction)
			s.NotEmpty(createResp.Status)
			s.Equal(tc.amount, createResp.Amount)
			s.Equal(string(tc.asset), createResp.Asset)
			s.Equal(string(tc.network), createResp.Network)

			s.T().Logf("Withdrawal created: %s", createResp.TransactionID)

			// Step 2: Get by ID
			getResp, err := s.Client.Withdrawals.GetWithdrawal(s.Ctx, s.CustomerID, createResp.TransactionID)
			s.Require().NoError(err, "GetWithdrawal should succeed")
			s.Require().NotNil(getResp)

			s.Equal(createResp.TransactionID, getResp.TransactionID)
			s.NotEmpty(getResp.Amount)
			s.Equal(createResp.Asset, getResp.Asset)
			s.Equal(createResp.Status, getResp.Status)

			s.T().Logf("GetWithdrawal verified: %s", getResp.TransactionID)

			// Step 3: Get by Idempotency Key
			getByKeyResp, err := s.Client.Withdrawals.GetWithdrawalByIdempotencyKey(s.Ctx, s.CustomerID, idempotencyKey)
			s.Require().NoError(err, "GetWithdrawalByIdempotencyKey should succeed")
			s.Require().NotNil(getByKeyResp)

			s.Equal(createResp.TransactionID, getByKeyResp.TransactionID)
			s.Equal(idempotencyKey, getByKeyResp.IdempotencyKey)

			s.T().Logf("GetByIdempotencyKey verified:\n%s", PrettyJSON(getByKeyResp))

			// Step 4: List Transactions
			txResp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
			s.Require().NoError(err, "ListTransactions should succeed")
			s.Require().NotNil(txResp)

			s.T().Logf("Transactions: total=%d, returned=%d", txResp.Total, len(txResp.List))
		})
	}
}

// TestWithdrawalsTestSuite runs the withdrawals test suite.
func TestWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalsTestSuite))
}
