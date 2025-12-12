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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// InstructionsTestSuite tests instructions service operations.
type InstructionsTestSuite struct {
	CustomerDependentTestSuite
}

// skipIfVerifiedFiatAccountRequired checks if the error is about requiring a verified fiat account
// and skips the test if so. This is expected for new customers who haven't set up fiat accounts yet.
func (s *InstructionsTestSuite) skipIfVerifiedFiatAccountRequired(err error) bool {
	if err != nil {
		if apiErr, ok := transport.IsAPIError(err); ok {
			if strings.Contains(apiErr.Detail, "verified fiat account") {
				s.T().Skipf("Skipping test: %s", apiErr.Detail)
				return true
			}
		}
	}
	return false
}

type instructionType int

const (
	instructionTypeFiat instructionType = iota
	instructionTypeCrypto
)

type depositInstructionTestCase struct {
	name            string
	asset           assets.AssetName
	network         assets.NetworkName
	expectedAsset   string
	expectedNetwork string
	instrType       instructionType
	addressPrefix   string // for crypto: expected wallet address prefix
}

func (s *InstructionsTestSuite) TestInstructions_GetDepositInstruction() {
	testCases := []depositInstructionTestCase{
		{
			name:            "USD_ACH",
			asset:           assets.AssetNameUSD,
			network:         assets.NetworkNameUSACH,
			expectedAsset:   "USD",
			expectedNetwork: "US_ACH",
			instrType:       instructionTypeFiat,
		},
		{
			name:            "USD_Fedwire",
			asset:           assets.AssetNameUSD,
			network:         assets.NetworkNameUSFEDWIRE,
			expectedAsset:   "USD",
			expectedNetwork: "US_FEDWIRE",
			instrType:       instructionTypeFiat,
		},
		{
			name:            "USDT_Ethereum",
			asset:           assets.AssetNameUSDT,
			network:         assets.NetworkNameETHEREUM,
			expectedAsset:   "USDT",
			expectedNetwork: "ETHEREUM",
			instrType:       instructionTypeCrypto,
			addressPrefix:   "0x",
		},
		{
			name:            "USDC_Polygon",
			asset:           assets.AssetNameUSDC,
			network:         assets.NetworkNamePOLYGON,
			expectedAsset:   "USDC",
			expectedNetwork: "POLYGON",
			instrType:       instructionTypeCrypto,
			addressPrefix:   "0x",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.Client.Instructions.GetDepositInstruction(s.Ctx, s.CustomerID, tc.asset, tc.network)
			if s.skipIfVerifiedFiatAccountRequired(err) {
				return
			}
			s.Require().NoError(err, "GetDepositInstruction should succeed")
			s.Require().NotNil(resp, "Response should not be nil")

			// Common validations
			s.Equal(tc.expectedAsset, resp.Asset)
			s.Equal(tc.expectedNetwork, resp.Network)
			s.Equal("DEPOSIT", resp.TransactionAction)
			s.NotEmpty(resp.CreatedAt)
			s.NotEmpty(resp.ModifiedAt)

			// Type-specific validations
			switch tc.instrType {
			case instructionTypeFiat:
				s.Require().NotNil(resp.BankInstruction, "BankInstruction should be present for fiat")
				s.NotEmpty(resp.BankInstruction.BankName)
				s.NotEmpty(resp.BankInstruction.TransactionFee.Value)
				s.NotEmpty(resp.BankInstruction.TransactionFee.Asset)
				s.Nil(resp.WalletInstruction, "WalletInstruction should be nil for fiat")

			case instructionTypeCrypto:
				s.Require().NotNil(resp.WalletInstruction, "WalletInstruction should be present for crypto")
				s.NotEmpty(resp.WalletInstruction.WalletAddress)
				if tc.addressPrefix != "" {
					s.True(
						strings.HasPrefix(resp.WalletInstruction.WalletAddress, tc.addressPrefix),
						"Wallet address should start with %s", tc.addressPrefix,
					)
				}
				s.Nil(resp.BankInstruction, "BankInstruction should be nil for crypto")
			}

			s.T().Logf("%s deposit instruction:\n%s", tc.name, PrettyJSON(resp))
		})
	}
}

// TestInstructionsTestSuite runs the instructions test suite.
func TestInstructionsTestSuite(t *testing.T) {
	suite.Run(t, new(InstructionsTestSuite))
}
