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
	CustomerDependentTestSuite
}

// TestConversions_CreateQuote_CryptoToFiat tests creating a quote to convert crypto to fiat.
// Validates the quote by verifying all response fields and executing a hedge.
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

	resp, err := s.Client.Conversions.CreateQuote(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateQuote should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.QuoteID, "QuoteID should not be empty")
	s.NotEmpty(resp.Rate, "Rate should not be empty")
	s.Positive(resp.ExpireTime, "ExpireTime should be positive")
	s.NotEmpty(resp.ValidUntilTimestamp, "ValidUntilTimestamp should not be empty")

	// Validate asset mapping matches request
	s.Equal(string(assets.AssetNameUSDT), resp.UserPayAsset, "UserPayAsset should match FromAsset")
	s.Equal("ETHEREUM", resp.UserPayNetwork, "UserPayNetwork should match FromAsset network")
	s.Equal(string(assets.AssetNameUSD), resp.UserObtainAsset, "UserObtainAsset should match ToAsset")
	s.NotEmpty(resp.UserPayAmount, "UserPayAmount should not be empty")
	s.NotEmpty(resp.UserObtainAmount, "UserObtainAmount should not be empty")

	s.T().Logf("Created crypto-to-fiat quote:\n%s", PrettyJSON(resp))

	// Validate quote is usable by executing hedge
	hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, s.CustomerID, &conversions.CreateHedgeRequest{
		QuoteID: resp.QuoteID,
	})
	s.Require().NoError(err, "CreateHedge should succeed - validates quote was truly created")
	s.Equal(resp.QuoteID, hedgeResp.QuoteID, "Hedge should reference the created quote")
	s.T().Logf("Quote validated via hedge execution: OrderID=%s", hedgeResp.OrderID)
}

// TestConversions_CreateQuote_FiatToCrypto tests creating a quote to convert fiat to crypto.
// Validates the quote by verifying all response fields and executing a hedge.
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

	resp, err := s.Client.Conversions.CreateQuote(s.Ctx, s.CustomerID, req)
	s.Require().NoError(err, "CreateQuote should succeed")

	// Validate response structure
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.QuoteID, "QuoteID should not be empty")
	s.NotEmpty(resp.Rate, "Rate should not be empty")
	s.Positive(resp.ExpireTime, "ExpireTime should be positive")
	s.NotEmpty(resp.ValidUntilTimestamp, "ValidUntilTimestamp should not be empty")

	// Validate asset mapping matches request
	s.Equal(string(assets.AssetNameUSD), resp.UserPayAsset, "UserPayAsset should match FromAsset")
	s.Equal(string(assets.AssetNameUSDT), resp.UserObtainAsset, "UserObtainAsset should match ToAsset")
	s.Equal("ETHEREUM", resp.UserObtainNetwork, "UserObtainNetwork should match ToAsset network")
	s.NotEmpty(resp.UserPayAmount, "UserPayAmount should not be empty")
	s.NotEmpty(resp.UserObtainAmount, "UserObtainAmount should not be empty")

	s.T().Logf("Created fiat-to-crypto quote:\n%s", PrettyJSON(resp))

	// Validate quote is usable by executing hedge
	hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, s.CustomerID, &conversions.CreateHedgeRequest{
		QuoteID: resp.QuoteID,
	})
	s.Require().NoError(err, "CreateHedge should succeed - validates quote was truly created")
	s.Equal(resp.QuoteID, hedgeResp.QuoteID, "Hedge should reference the created quote")
	s.T().Logf("Quote validated via hedge execution: OrderID=%s", hedgeResp.OrderID)
}

// TestConversions_CreateHedge tests executing a conversion hedge.
// Validates the order response fields match the original quote.
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

	quoteResp, err := s.Client.Conversions.CreateQuote(s.Ctx, s.CustomerID, quoteReq)
	s.Require().NoError(err, "CreateQuote should succeed")

	// Execute the hedge
	hedgeReq := &conversions.CreateHedgeRequest{
		QuoteID: quoteResp.QuoteID,
	}

	hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, s.CustomerID, hedgeReq)
	s.Require().NoError(err, "CreateHedge should succeed")

	// Validate response structure
	s.Require().NotNil(hedgeResp, "Response should not be nil")
	s.NotEmpty(hedgeResp.OrderID, "OrderID should not be empty")
	s.NotEmpty(hedgeResp.OrderStatus, "OrderStatus should not be empty")
	s.Equal(quoteResp.QuoteID, hedgeResp.QuoteID, "QuoteID should match")

	// Validate order inherits quote details
	s.Equal(quoteResp.UserPayAsset, hedgeResp.UserPayAsset, "UserPayAsset should match quote")
	s.Equal(quoteResp.UserPayNetwork, hedgeResp.UserPayNetwork, "UserPayNetwork should match quote")
	s.Equal(quoteResp.UserObtainAsset, hedgeResp.UserObtainAsset, "UserObtainAsset should match quote")
	s.Equal(quoteResp.Rate, hedgeResp.Rate, "Rate should match quote")
	s.NotEmpty(hedgeResp.UserPayAmount, "UserPayAmount should not be empty")
	s.NotEmpty(hedgeResp.UserObtainAmount, "UserObtainAmount should not be empty")

	s.T().Logf("Created hedge order:\n%s", PrettyJSON(hedgeResp))
}

// TestConversions_GetOrder tests retrieving a conversion order.
// Validates retrieved order matches the created hedge order.
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

	quoteResp, err := s.Client.Conversions.CreateQuote(s.Ctx, s.CustomerID, quoteReq)
	s.Require().NoError(err, "CreateQuote should succeed")

	hedgeReq := &conversions.CreateHedgeRequest{
		QuoteID: quoteResp.QuoteID,
	}

	hedgeResp, err := s.Client.Conversions.CreateHedge(s.Ctx, s.CustomerID, hedgeReq)
	s.Require().NoError(err, "CreateHedge should succeed")

	// Get the order
	orderResp, err := s.Client.Conversions.GetOrder(s.Ctx, s.CustomerID, hedgeResp.OrderID)
	s.Require().NoError(err, "GetOrder should succeed")

	// Validate response structure
	s.Require().NotNil(orderResp, "Response should not be nil")
	s.Equal(hedgeResp.OrderID, orderResp.OrderID, "OrderID should match")
	s.Equal(hedgeResp.QuoteID, orderResp.QuoteID, "QuoteID should match")
	s.NotEmpty(orderResp.OrderStatus, "OrderStatus should not be empty")

	// Validate order details match hedge response
	s.Equal(hedgeResp.UserPayAsset, orderResp.UserPayAsset, "UserPayAsset should match hedge")
	s.Equal(hedgeResp.UserPayNetwork, orderResp.UserPayNetwork, "UserPayNetwork should match hedge")
	s.Equal(hedgeResp.UserObtainAsset, orderResp.UserObtainAsset, "UserObtainAsset should match hedge")
	s.Equal(hedgeResp.Rate, orderResp.Rate, "Rate should match hedge")
	s.NotEmpty(orderResp.UserPayAmount, "UserPayAmount should not be empty")
	s.NotEmpty(orderResp.UserObtainAmount, "UserObtainAmount should not be empty")

	s.T().Logf("Retrieved order:\n%s", PrettyJSON(orderResp))
}

// TestConversionsTestSuite runs the conversions test suite.
func TestConversionsTestSuite(t *testing.T) {
	suite.Run(t, new(ConversionsTestSuite))
}
