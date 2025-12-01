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
)

// ConversionsTestSuite tests conversions service operations.
type ConversionsTestSuite struct {
	E2ETestSuite
}

// TestConversions_CreateQuote_CryptoToFiat tests creating a quote to convert crypto to fiat.
func (s *ConversionsTestSuite) TestConversions_CreateQuote_CryptoToFiat() {
	req := &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDT,
			Amount:  "100.00",
			Network: conversions.WalletNetworkNameETHEREUM,
		},
		ToAsset: conversions.AssetInfo{
			Asset: assets.AssetNameUSD,
		},
	}

	resp, err := s.Client.Conversions.CreateQuote(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "CreateQuote should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.QuoteID, "QuoteID should not be empty")
	s.NotEmpty(resp.Rate, "Rate should not be empty")
	s.T().Logf("Created crypto-to-fiat quote:\n%s", PrettyJSON(resp))
}

// TestConversions_CreateQuote_FiatToCrypto tests creating a quote to convert fiat to crypto.
func (s *ConversionsTestSuite) TestConversions_CreateQuote_FiatToCrypto() {
	req := &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:  assets.AssetNameUSD,
			Amount: "100.00",
		},
		ToAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDT,
			Network: conversions.WalletNetworkNameETHEREUM,
		},
	}

	resp, err := s.Client.Conversions.CreateQuote(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "CreateQuote should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.QuoteID, "QuoteID should not be empty")
	s.NotEmpty(resp.Rate, "Rate should not be empty")
	s.T().Logf("Created fiat-to-crypto quote:\n%s", PrettyJSON(resp))
}

// TestConversions_CreateHedge tests executing a conversion hedge.
func (s *ConversionsTestSuite) TestConversions_CreateHedge() {
	// First create a quote
	quoteReq := &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDT,
			Amount:  "10.00",
			Network: conversions.WalletNetworkNameETHEREUM,
		},
		ToAsset: conversions.AssetInfo{
			Asset: assets.AssetNameUSD,
		},
	}

	quoteResp, err := s.Client.Conversions.CreateQuote(s.Ctx, testCustomerID, quoteReq)
	s.Require().NoError(err, "CreateQuote should succeed")

	// Execute the hedge
	hedgeReq := &conversions.CreateHedgeRequest{
		QuoteID: quoteResp.QuoteID,
	}

	hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, testCustomerID, hedgeReq)
	s.Require().NoError(err, "CreateHedge should succeed")

	s.Require().NotNil(hedgeResp, "Response should not be nil")
	s.NotEmpty(hedgeResp.OrderID, "OrderID should not be empty")
	s.Equal(quoteResp.QuoteID, hedgeResp.QuoteID, "QuoteID should match")
	s.T().Logf("Created hedge order:\n%s", PrettyJSON(hedgeResp))
}

// TestConversions_GetOrder tests retrieving a conversion order.
func (s *ConversionsTestSuite) TestConversions_GetOrder() {
	// First create a quote and execute hedge
	quoteReq := &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDT,
			Amount:  "10.00",
			Network: conversions.WalletNetworkNameETHEREUM,
		},
		ToAsset: conversions.AssetInfo{
			Asset: assets.AssetNameUSD,
		},
	}

	quoteResp, err := s.Client.Conversions.CreateQuote(s.Ctx, testCustomerID, quoteReq)
	s.Require().NoError(err, "CreateQuote should succeed")

	hedgeReq := &conversions.CreateHedgeRequest{
		QuoteID: quoteResp.QuoteID,
	}

	hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, testCustomerID, hedgeReq)
	s.Require().NoError(err, "CreateHedge should succeed")

	// Get the order
	orderResp, err := s.Client.Conversions.GetOrder(s.Ctx, testCustomerID, hedgeResp.OrderID)
	s.Require().NoError(err, "GetOrder should succeed")

	s.Require().NotNil(orderResp, "Response should not be nil")
	s.Equal(hedgeResp.OrderID, orderResp.OrderID, "OrderID should match")
	s.T().Logf("Retrieved order:\n%s", PrettyJSON(orderResp))
}

// TestConversionsTestSuite runs the conversions test suite.
func TestConversionsTestSuite(t *testing.T) {
	suite.Run(t, new(ConversionsTestSuite))
}
