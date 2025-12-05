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
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USD_ACH() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSD, assets.NetworkNameUSACH)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("USD ACH deposit instruction:\n%s", PrettyJSON(resp))

	s.Equal("USD", resp.Asset, "Asset should be USD")
	s.Equal("US_ACH", resp.Network, "Network should be US_ACH")
	s.NotEmpty(resp.TransactionAction, "TransactionAction should not be empty")

	if resp.BankInstruction != nil {
		s.T().Logf("Bank instruction available")
		s.NotEmpty(resp.BankInstruction.TransactionFee, "TransactionFee should not be empty")
	}
}

// TestInstructions_GetDepositInstruction_USD_Fedwire tests getting USD deposit instructions via Fedwire.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USD_Fedwire() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSD, assets.NetworkNameUSFEDWIRE)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("USD Fedwire deposit instruction:\n%s", PrettyJSON(resp))

	s.Equal("USD", resp.Asset, "Asset should be USD")
	s.Equal("US_FEDWIRE", resp.Network, "Network should be US_FEDWIRE")
}

// TestInstructions_GetDepositInstruction_USDT_Ethereum tests getting USDT deposit instructions on Ethereum.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USDT_Ethereum() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSDT, assets.NetworkNameETHEREUM)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("USDT Ethereum deposit instruction:\n%s", PrettyJSON(resp))

	s.Equal("USDT", resp.Asset, "Asset should be USDT")
	s.Equal("ETHEREUM", resp.Network, "Network should be ETHEREUM")

	if resp.WalletInstruction != nil {
		s.T().Logf("Wallet instruction available")
		s.NotEmpty(resp.WalletInstruction.WalletAddress, "WalletAddress should not be empty")
	}
}

// TestInstructions_GetDepositInstruction_USDC_Polygon tests getting USDC deposit instructions on Polygon.
func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction_USDC_Polygon() {
	resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, assets.AssetNameUSDC, assets.NetworkNamePOLYGON)
	s.Require().NoError(err, "GetDepositInstruction should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("USDC Polygon deposit instruction:\n%s", PrettyJSON(resp))

	s.Equal("USDC", resp.Asset, "Asset should be USDC")
	s.Equal("POLYGON", resp.Network, "Network should be POLYGON")
}

// TestInstructionsTestSuite runs the instructions test suite.
func TestInstructionsTestSuite(t *testing.T) {
	suite.Run(t, new(InstructionsTestSuite))
}
