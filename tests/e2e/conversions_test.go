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
	"github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
)

// ConversionsTestSuite tests conversions service operations.
type ConversionsTestSuite struct {
	CustomerDependentTestSuite
}

// SetupSuite prepares balances for conversion tests.
func (s *ConversionsTestSuite) SetupSuite() {
	s.CustomerDependentTestSuite.SetupSuite()

	// Step 1: Simulate USD deposit
	s.T().Log("Simulating USD deposit...")
	_, err := s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, &simulations.SimulateDepositRequest{
		Asset:  assets.AssetNameUSD,
		Amount: "1000.00",
	})
	s.Require().NoError(err, "SimulateDeposit USD should succeed")

	// Step 2: Convert USD to USDC (Polygon)
	s.T().Log("Converting USD to USDC Polygon...")
	usdToUsdcQuote, err := s.Client.Conversions.CreateQuote(s.Ctx, s.CustomerID, &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:  assets.AssetNameUSD,
			Amount: "500.00",
		},
		ToAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDC,
			Network: conversions.WalletNetworkNamePOLYGON,
		},
	})
	s.Require().NoError(err, "CreateQuote USD->USDC should succeed")

	_, err = s.Client.Conversions.CreateHedge(s.Ctx, s.CustomerID, &conversions.CreateHedgeRequest{
		QuoteID: usdToUsdcQuote.QuoteID,
	})
	s.Require().NoError(err, "CreateHedge USD->USDC should succeed")

	// Step 3: Simulate USDC Ethereum deposit (for cross-chain tests)
	s.T().Log("Simulating USDC Ethereum deposit...")
	_, err = s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDC,
		Network: simulations.WalletNetworkNameETHEREUM,
		Amount:  "200.00",
	})
	s.Require().NoError(err, "SimulateDeposit USDC Ethereum should succeed")

	s.T().Log("Balance preparation completed")
}

type conversionTestCase struct {
	name             string
	fromAsset        assets.AssetName
	fromNetwork      conversions.WalletNetworkName
	fromAmount       string
	toAsset          assets.AssetName
	toNetwork        conversions.WalletNetworkName
	expectPayNetwork string // expected UserPayNetwork in response
	expectGetNetwork string // expected UserObtainNetwork in response
}

// TestConversions_Flow tests the complete conversion flow: CreateQuote -> CreateHedge -> GetOrder
func (s *ConversionsTestSuite) TestConversions_Flow() {
	testCases := []conversionTestCase{
		{
			name:             "CryptoToFiat_USDC_Polygon_to_USD",
			fromAsset:        assets.AssetNameUSDC,
			fromNetwork:      conversions.WalletNetworkNamePOLYGON,
			fromAmount:       "10.00",
			toAsset:          assets.AssetNameUSD,
			expectPayNetwork: "POLYGON",
			expectGetNetwork: "",
		},
		{
			name:             "FiatToCrypto_USD_to_USDC_Polygon",
			fromAsset:        assets.AssetNameUSD,
			fromAmount:       "10.00",
			toAsset:          assets.AssetNameUSDC,
			toNetwork:        conversions.WalletNetworkNamePOLYGON,
			expectPayNetwork: "",
			expectGetNetwork: "POLYGON",
		},
		{
			name:             "FiatToCrypto_USD_to_USDC_Ethereum",
			fromAsset:        assets.AssetNameUSD,
			fromAmount:       "10.00",
			toAsset:          assets.AssetNameUSDC,
			toNetwork:        conversions.WalletNetworkNameETHEREUM,
			expectPayNetwork: "",
			expectGetNetwork: "ETHEREUM",
		},
		// {
		// 	name:        "CrossChain_USDC_Ethereum_to_USDT_Solana",
		// 	fromAsset:   assets.AssetNameUSDC,
		// 	fromNetwork: conversions.WalletNetworkNameETHEREUM,
		// 	fromAmount:  "100.00", // minimum 100 for cross-chain
		// 	toAsset:     assets.AssetNameUSDT,
		// 	toNetwork:   conversions.WalletNetworkNameSOLANA,
		// },
	}

	for i := range testCases {
		tc := testCases[i]
		s.Run(tc.name, func() {
			// Step 1: Create Quote
			quoteResp, err := s.Client.Conversions.CreateQuote(s.Ctx, s.CustomerID, &conversions.CreateQuoteRequest{
				FromAsset: conversions.AssetInfo{
					Asset:   tc.fromAsset,
					Amount:  tc.fromAmount,
					Network: tc.fromNetwork,
				},
				ToAsset: conversions.AssetInfo{
					Asset:   tc.toAsset,
					Network: tc.toNetwork,
				},
			})
			s.Require().NoError(err, "CreateQuote should succeed")
			s.Require().NotNil(quoteResp)

			// Validate quote
			s.NotEmpty(quoteResp.QuoteID)
			s.NotEmpty(quoteResp.Rate)
			s.Positive(quoteResp.ExpireTime)
			s.Equal(string(tc.fromAsset), quoteResp.UserPayAsset)
			s.Equal(string(tc.toAsset), quoteResp.UserObtainAsset)
			s.Equal(tc.expectPayNetwork, quoteResp.UserPayNetwork, "Quote UserPayNetwork should match")
			s.Equal(tc.expectGetNetwork, quoteResp.UserObtainNetwork, "Quote UserObtainNetwork should match")

			s.T().Logf("Quote created: %s", quoteResp.QuoteID)

			// Step 2: Execute Hedge
			hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, s.CustomerID, &conversions.CreateHedgeRequest{
				QuoteID: quoteResp.QuoteID,
			})
			s.Require().NoError(err, "CreateHedge should succeed")
			s.Require().NotNil(hedgeResp)

			// Validate hedge
			s.NotEmpty(hedgeResp.OrderID)
			s.NotEmpty(hedgeResp.OrderStatus)
			s.Equal(quoteResp.QuoteID, hedgeResp.QuoteID)
			s.Equal(quoteResp.UserPayAsset, hedgeResp.UserPayAsset)
			s.Equal(quoteResp.UserObtainAsset, hedgeResp.UserObtainAsset)
			s.Equal(quoteResp.Rate, hedgeResp.Rate)
			s.Equal(tc.expectPayNetwork, hedgeResp.UserPayNetwork, "Hedge UserPayNetwork should match")
			s.Equal(tc.expectGetNetwork, hedgeResp.UserObtainNetwork, "Hedge UserObtainNetwork should match")

			s.T().Logf("Hedge executed: OrderID=%s, Status=%s", hedgeResp.OrderID, hedgeResp.OrderStatus)

			// Step 3: Get Order
			orderResp, err := s.Client.Conversions.GetOrder(s.Ctx, s.CustomerID, hedgeResp.OrderID)
			s.Require().NoError(err, "GetOrder should succeed")
			s.Require().NotNil(orderResp)

			// Validate order
			s.Equal(hedgeResp.OrderID, orderResp.OrderID)
			s.Equal(hedgeResp.QuoteID, orderResp.QuoteID)
			s.NotEmpty(orderResp.OrderStatus)
			s.Equal(hedgeResp.UserPayAsset, orderResp.UserPayAsset)
			s.Equal(hedgeResp.UserObtainAsset, orderResp.UserObtainAsset)
			s.Equal(tc.expectPayNetwork, orderResp.UserPayNetwork, "Order UserPayNetwork should match")
			s.Equal(tc.expectGetNetwork, orderResp.UserObtainNetwork, "Order UserObtainNetwork should match")

			s.T().Logf("Order verified: %s\n%s", orderResp.OrderID, PrettyJSON(orderResp))

			// Step 4: List Transactions
			txResp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
			s.Require().NoError(err, "ListTransactions should succeed")
			s.Require().NotNil(txResp)
			s.Positive(txResp.Total, "Should have at least one transaction")
			s.NotEmpty(txResp.List, "Transaction list should not be empty")
			s.T().Logf("Transactions: total=%d, returned=%d", txResp.Total, len(txResp.List))
		})
	}
}

// TestConversionsTestSuite runs the conversions test suite.
func TestConversionsTestSuite(t *testing.T) {
	suite.Run(t, new(ConversionsTestSuite))
}
