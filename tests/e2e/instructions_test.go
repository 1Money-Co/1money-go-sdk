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
)

// InstructionsTestSuite tests instructions service operations.
type InstructionsTestSuite struct {
	CustomerDependentTestSuite
}

// TestInstructions_GetDepositInstruction_USD_ACH tests getting USD deposit instructions via ACH.
// Validates all response fields including bank instruction details.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USD_ACH() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSD, assets.NetworkNameUSACH)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal("USD", resp.Asset, "Asset should be USD")
	s.Equal("US_ACH", resp.Network, "Network should be US_ACH")
	s.Equal("DEPOSIT", resp.TransactionAction, "TransactionAction should be DEPOSIT")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate bank instruction is present for fiat
	s.Require().NotNil(resp.BankInstruction, "BankInstruction should be present for USD")
	s.NotEmpty(resp.BankInstruction.BankName, "BankName should not be empty")
	s.NotEmpty(resp.BankInstruction.RoutingNumber, "RoutingNumber should not be empty")
	s.NotEmpty(resp.BankInstruction.AccountNumber, "AccountNumber should not be empty")
	s.NotEmpty(resp.BankInstruction.AccountHolder, "AccountHolder should not be empty")
	s.NotEmpty(resp.BankInstruction.TransactionFee, "TransactionFee should not be empty")

	// WalletInstruction should be nil for fiat
	s.Nil(resp.WalletInstruction, "WalletInstruction should be nil for USD")

	s.T().Logf("USD ACH deposit instruction:\n%s", PrettyJSON(resp))
}

// TestInstructions_GetDepositInstruction_USD_Fedwire tests getting USD deposit instructions via Fedwire.
// Validates all response fields including bank instruction details.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USD_Fedwire() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSD, assets.NetworkNameUSFEDWIRE)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal("USD", resp.Asset, "Asset should be USD")
	s.Equal("US_FEDWIRE", resp.Network, "Network should be US_FEDWIRE")
	s.Equal("DEPOSIT", resp.TransactionAction, "TransactionAction should be DEPOSIT")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate bank instruction is present for fiat
	s.Require().NotNil(resp.BankInstruction, "BankInstruction should be present for USD")
	s.NotEmpty(resp.BankInstruction.BankName, "BankName should not be empty")
	s.NotEmpty(resp.BankInstruction.TransactionFee, "TransactionFee should not be empty")

	s.T().Logf("USD Fedwire deposit instruction:\n%s", PrettyJSON(resp))
}

// TestInstructions_GetDepositInstruction_USDT_Ethereum tests getting USDT deposit instructions on Ethereum.
// Validates all response fields including wallet instruction details.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USDT_Ethereum() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSDT, assets.NetworkNameETHEREUM)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal("USDT", resp.Asset, "Asset should be USDT")
	s.Equal("ETHEREUM", resp.Network, "Network should be ETHEREUM")
	s.Equal("DEPOSIT", resp.TransactionAction, "TransactionAction should be DEPOSIT")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate wallet instruction is present for crypto
	s.Require().NotNil(resp.WalletInstruction, "WalletInstruction should be present for USDT")
	s.NotEmpty(resp.WalletInstruction.WalletAddress, "WalletAddress should not be empty")
	s.NotEmpty(resp.WalletInstruction.TransactionFee, "TransactionFee should not be empty")

	// Validate wallet address format (Ethereum addresses start with 0x)
	s.Greater(len(resp.WalletInstruction.WalletAddress), 2, "WalletAddress should have valid length")
	s.Equal("0x", resp.WalletInstruction.WalletAddress[:2], "Ethereum address should start with 0x")

	// BankInstruction should be nil for crypto
	s.Nil(resp.BankInstruction, "BankInstruction should be nil for crypto")

	s.T().Logf("USDT Ethereum deposit instruction:\n%s", PrettyJSON(resp))
}

// TestInstructions_GetDepositInstruction_USDC_Polygon tests getting USDC deposit instructions on Polygon.
// Validates all response fields including wallet instruction details.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USDC_Polygon() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSDC, assets.NetworkNamePOLYGON)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal("USDC", resp.Asset, "Asset should be USDC")
	s.Equal("POLYGON", resp.Network, "Network should be POLYGON")
	s.Equal("DEPOSIT", resp.TransactionAction, "TransactionAction should be DEPOSIT")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.ModifiedAt, "ModifiedAt should not be empty")

	// Validate wallet instruction is present for crypto
	s.Require().NotNil(resp.WalletInstruction, "WalletInstruction should be present for USDC")
	s.NotEmpty(resp.WalletInstruction.WalletAddress, "WalletAddress should not be empty")
	s.NotEmpty(resp.WalletInstruction.TransactionFee, "TransactionFee should not be empty")

	// Validate wallet address format (Polygon uses Ethereum-compatible addresses)
	s.Greater(len(resp.WalletInstruction.WalletAddress), 2, "WalletAddress should have valid length")
	s.Equal("0x", resp.WalletInstruction.WalletAddress[:2], "Polygon address should start with 0x")

	s.T().Logf("USDC Polygon deposit instruction:\n%s", PrettyJSON(resp))
}

// TestInstructionsTestSuite runs the instructions test suite.
func TestInstructionsTestSuite(t *testing.T) {
	suite.Run(t, new(InstructionsTestSuite))
}
