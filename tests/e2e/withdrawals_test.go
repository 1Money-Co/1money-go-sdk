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

	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

// WithdrawalsTestSuite tests withdrawals service operations.
type WithdrawalsTestSuite struct {
	CustomerDependentTestSuite
}

// TestWithdrawals_CreateFiatWithdrawal tests creating a fiat withdrawal to an external bank account.
// Validates all response fields and verifies the withdrawal can be retrieved.
func (s *WithdrawalsTestSuite) TestWithdrawals_CreateFiatWithdrawal() {
	// Ensure we have an external account to withdraw to
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	idempotencyKey := uuid.New().String()
	amount := "10.00"

	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    idempotencyKey,
		Amount:            amount,
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	resp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.TransactionID, "TransactionID should not be empty")
	s.Equal(idempotencyKey, resp.IdempotencyKey, "IdempotencyKey should match")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.Equal("WITHDRAWAL", resp.TransactionAction, "TransactionAction should be WITHDRAWAL")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate request fields are reflected in response
	s.Equal(amount, resp.Amount, "Amount should match request")
	s.Equal(string(assets.AssetNameUSD), resp.Asset, "Asset should match request")
	s.Equal(string(assets.NetworkNameUSACH), resp.Network, "Network should match request")
	s.Equal(externalAccountID, resp.ExternalAccountID, "ExternalAccountID should match request")

	// Validate fee info
	s.NotEmpty(resp.TransactionFee.Asset, "TransactionFee.Asset should not be empty")

	s.T().Logf("Created fiat withdrawal:\n%s", PrettyJSON(resp))

	// Verify withdrawal can be retrieved by ID
	getResp, err := s.Client.Withdrawals.GetWithdrawal(s.Ctx, s.CustomerID, resp.TransactionID)
	s.Require().NoError(err, "GetWithdrawal should succeed")
	s.Equal(resp.TransactionID, getResp.TransactionID, "Retrieved TransactionID should match")
}

// TestWithdrawals_CreateCryptoWithdrawal tests creating a crypto withdrawal to a wallet address.
// Validates all response fields and verifies the withdrawal can be retrieved.
func (s *WithdrawalsTestSuite) TestWithdrawals_CreateCryptoWithdrawal() {
	idempotencyKey := uuid.New().String()
	amount := "10.00"
	walletAddress := FakeEthereumAddress()

	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey: idempotencyKey,
		Amount:         amount,
		Asset:          assets.AssetNameUSDT,
		Network:        assets.NetworkNameETHEREUM,
		WalletAddress:  walletAddress,
	}

	resp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.TransactionID, "TransactionID should not be empty")
	s.Equal(idempotencyKey, resp.IdempotencyKey, "IdempotencyKey should match")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.Equal("WITHDRAWAL", resp.TransactionAction, "TransactionAction should be WITHDRAWAL")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate request fields are reflected in response
	s.Equal(amount, resp.Amount, "Amount should match request")
	s.Equal(string(assets.AssetNameUSDT), resp.Asset, "Asset should match request")
	s.Equal(string(assets.NetworkNameETHEREUM), resp.Network, "Network should match request")
	s.Equal(walletAddress, resp.WalletAddress, "WalletAddress should match request")

	// Validate fee info
	s.NotEmpty(resp.TransactionFee.Asset, "TransactionFee.Asset should not be empty")

	s.T().Logf("Created crypto withdrawal:\n%s", PrettyJSON(resp))

	// Verify withdrawal can be retrieved by ID
	getResp, err := s.Client.Withdrawals.GetWithdrawal(s.Ctx, s.CustomerID, resp.TransactionID)
	s.Require().NoError(err, "GetWithdrawal should succeed")
	s.Equal(resp.TransactionID, getResp.TransactionID, "Retrieved TransactionID should match")
}

// TestWithdrawals_GetByIdempotencyKey tests retrieving a withdrawal by idempotency key.
// Validates that retrieved withdrawal matches the created one exactly.
func (s *WithdrawalsTestSuite) TestWithdrawals_GetByIdempotencyKey() {
	// Ensure we have an external account
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	idempotencyKey := uuid.New().String()
	amount := "5.00"
	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    idempotencyKey,
		Amount:            amount,
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	createResp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	// Get by idempotency key
	getResp, err := s.Client.Withdrawals.GetWithdrawalByIdempotencyKey(s.Ctx, s.CustomerID, idempotencyKey)
	s.Require().NoError(err, "GetWithdrawalByIdempotencyKey should succeed")

	// Validate retrieved withdrawal matches created one
	s.Require().NotNil(getResp, "Response should not be nil")
	s.Equal(createResp.TransactionID, getResp.TransactionID, "TransactionID should match")
	s.Equal(createResp.IdempotencyKey, getResp.IdempotencyKey, "IdempotencyKey should match")
	s.Equal(createResp.Amount, getResp.Amount, "Amount should match")
	s.Equal(createResp.Asset, getResp.Asset, "Asset should match")
	s.Equal(createResp.Network, getResp.Network, "Network should match")
	s.Equal(createResp.ExternalAccountID, getResp.ExternalAccountID, "ExternalAccountID should match")
	s.Equal(createResp.Status, getResp.Status, "Status should match")
	s.Equal(createResp.TransactionAction, getResp.TransactionAction, "TransactionAction should match")

	s.T().Logf("Retrieved withdrawal by idempotency key:\n%s", PrettyJSON(getResp))
}

// TestWithdrawals_GetByID tests retrieving a withdrawal by ID.
// Validates that retrieved withdrawal matches the created one exactly.
func (s *WithdrawalsTestSuite) TestWithdrawals_GetByID() {
	// Ensure we have an external account
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	idempotencyKey := uuid.New().String()
	amount := "5.00"
	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    idempotencyKey,
		Amount:            amount,
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	createResp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	// Get by ID
	getResp, err := s.Client.Withdrawals.GetWithdrawal(s.Ctx, s.CustomerID, createResp.TransactionID)
	s.Require().NoError(err, "GetWithdrawal should succeed")

	// Validate retrieved withdrawal matches created one
	s.Require().NotNil(getResp, "Response should not be nil")
	s.Equal(createResp.TransactionID, getResp.TransactionID, "TransactionID should match")
	s.Equal(createResp.IdempotencyKey, getResp.IdempotencyKey, "IdempotencyKey should match")
	s.Equal(createResp.Amount, getResp.Amount, "Amount should match")
	s.Equal(createResp.Asset, getResp.Asset, "Asset should match")
	s.Equal(createResp.Network, getResp.Network, "Network should match")
	s.Equal(createResp.ExternalAccountID, getResp.ExternalAccountID, "ExternalAccountID should match")
	s.Equal(createResp.Status, getResp.Status, "Status should match")
	s.Equal(createResp.TransactionAction, getResp.TransactionAction, "TransactionAction should match")

	s.T().Logf("Retrieved withdrawal by ID:\n%s", PrettyJSON(getResp))
}

// TestWithdrawalsTestSuite runs the withdrawals test suite.
func TestWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalsTestSuite))
}
