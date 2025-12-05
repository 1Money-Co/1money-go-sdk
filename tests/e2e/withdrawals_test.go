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
func (s *WithdrawalsTestSuite) TestWithdrawals_CreateFiatWithdrawal() {
	// Ensure we have an external account to withdraw to
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	idempotencyKey := uuid.New().String()

	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    idempotencyKey,
		Amount:            "10.00",
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	resp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.TransactionID, "TransactionID should not be empty")
	s.Equal(idempotencyKey, resp.IdempotencyKey, "IdempotencyKey should match")
	s.T().Logf("Created fiat withdrawal:\n%s", PrettyJSON(resp))
}

// TestWithdrawals_CreateCryptoWithdrawal tests creating a crypto withdrawal to a wallet address.
func (s *WithdrawalsTestSuite) TestWithdrawals_CreateCryptoWithdrawal() {
	idempotencyKey := uuid.New().String()

	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey: idempotencyKey,
		Amount:         "10.00",
		Asset:          assets.AssetNameUSDT,
		Network:        assets.NetworkNameETHEREUM,
		WalletAddress:  FakeEthereumAddress(),
	}

	resp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.TransactionID, "TransactionID should not be empty")
	s.T().Logf("Created crypto withdrawal:\n%s", PrettyJSON(resp))
}

// TestWithdrawals_GetByIdempotencyKey tests retrieving a withdrawal by idempotency key.
func (s *WithdrawalsTestSuite) TestWithdrawals_GetByIdempotencyKey() {
	// Ensure we have an external account
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	idempotencyKey := uuid.New().String()
	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    idempotencyKey,
		Amount:            "5.00",
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	createResp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	// Get by idempotency key
	getResp, err := s.Client.Withdrawals.GetWithdrawalByIdempotencyKey(s.Ctx, s.CustomerID, idempotencyKey)
	s.Require().NoError(err, "GetWithdrawalByIdempotencyKey should succeed")

	s.Require().NotNil(getResp, "Response should not be nil")
	s.Equal(createResp.TransactionID, getResp.TransactionID, "TransactionID should match")
	s.T().Logf("Retrieved withdrawal by idempotency key:\n%s", PrettyJSON(getResp))
}

// TestWithdrawals_GetByID tests retrieving a withdrawal by ID.
func (s *WithdrawalsTestSuite) TestWithdrawals_GetByID() {
	// Ensure we have an external account
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	idempotencyKey := uuid.New().String()
	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    idempotencyKey,
		Amount:            "5.00",
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	createResp, err := s.Client.Withdrawals.CreateWithdrawal(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateWithdrawal should succeed")

	// Get by ID
	getResp, err := s.Client.Withdrawals.GetWithdrawal(s.Ctx, s.CustomerID, createResp.TransactionID)
	s.Require().NoError(err, "GetWithdrawal should succeed")

	s.Require().NotNil(getResp, "Response should not be nil")
	s.Equal(createResp.TransactionID, getResp.TransactionID, "TransactionID should match")
	s.T().Logf("Retrieved withdrawal by ID:\n%s", PrettyJSON(getResp))
}

// TestWithdrawalsTestSuite runs the withdrawals test suite.
func TestWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalsTestSuite))
}
