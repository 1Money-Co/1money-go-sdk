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

	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
)

// AutoConversionRulesTestSuite tests auto conversion rules service operations.
type AutoConversionRulesTestSuite struct {
	CustomerDependentTestSuite
}

// TestAutoConversionRules_List tests listing auto conversion rules with various scenarios.
func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_List() {
	s.Run("Empty", func() {
		// For a fresh customer, listing should succeed even with no rules
		resp, err := s.Client.AutoConversionRules.ListRules(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListRules should succeed even with no rules")
		s.Require().NotNil(resp, "Response should not be nil")
		s.T().Logf("Auto conversion rules list: %d rules", len(resp.Items))
	})

	s.Run("WithData", func() {
		// Ensure we have at least one auto conversion rule
		_, err := s.EnsureAutoConversionRule()
		s.Require().NoError(err, "EnsureAutoConversionRule should succeed")

		resp, err := s.Client.AutoConversionRules.ListRules(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListRules should succeed")

		s.Require().NotNil(resp, "Response should not be nil")
		s.Require().NotEmpty(resp.Items, "Should have at least one auto conversion rule")
		s.T().Logf("Auto conversion rules list:\n%s", PrettyJSON(resp))

		for i := range resp.Items {
			s.NotEmpty(resp.Items[i].AutoConversionRuleID, "Rule ID should not be empty")
			s.NotEmpty(resp.Items[i].IdempotencyKey, "Idempotency key should not be empty")
			s.NotEmpty(resp.Items[i].Status, "Status should not be empty")
			s.NotEmpty(resp.Items[i].Source.Asset, "Source asset should not be empty")
			s.NotEmpty(resp.Items[i].Destination.Asset, "Destination asset should not be empty")
		}
	})

	s.Run("WithPagination", func() {
		req := &auto_conversion_rules.ListRulesRequest{
			Page: 1,
			Size: 5,
		}

		resp, err := s.Client.AutoConversionRules.ListRules(s.Ctx, s.CustomerID, req)
		s.Require().NoError(err, "ListRules with pagination should succeed")
		s.Require().NotNil(resp, "Response should not be nil")
		s.LessOrEqual(len(resp.Items), 5, "Should return at most 5 items")
		s.T().Logf("Auto conversion rules (page 1, size 5): %d items, total: %d", len(resp.Items), resp.Total)
	})
}

// TestAutoConversionRules_CreateAndGet tests creating and retrieving an auto conversion rule.
func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_CreateAndGet() {
	createReq := FakeAutoConversionRuleRequest()

	// Create auto conversion rule
	createResp, err := s.Client.AutoConversionRules.CreateRule(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRule should succeed")

	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.AutoConversionRuleID, "Rule ID should not be empty")
	s.Equal(createReq.IdempotencyKey, createResp.IdempotencyKey, "Idempotency key should match")
	s.Equal("ACTIVE", createResp.Status, "Status should be ACTIVE")
	s.Equal(createReq.Source.Asset, createResp.Source.Asset, "Source asset should match")
	s.Equal(createReq.Destination.Asset, createResp.Destination.Asset, "Destination asset should match")
	s.T().Logf("Created auto conversion rule:\n%s", PrettyJSON(createResp))

	// Get auto conversion rule by ID
	getResp, err := s.Client.AutoConversionRules.GetRule(s.Ctx, s.CustomerID, createResp.AutoConversionRuleID)
	s.Require().NoError(err, "GetRule should succeed")

	s.Require().NotNil(getResp, "Get response should not be nil")
	s.Equal(createResp.AutoConversionRuleID, getResp.AutoConversionRuleID, "Rule IDs should match")
	s.NotNil(getResp.SourceDepositInfo, "SourceDepositInfo should be present in get response")
	s.T().Logf("Retrieved auto conversion rule:\n%s", PrettyJSON(getResp))

	// Verify source deposit info based on source asset type
	if createReq.Source.Asset == "USD" {
		s.NotNil(getResp.SourceDepositInfo.Bank, "Bank deposit info should be present for USD source")
		s.NotEmpty(getResp.SourceDepositInfo.Bank.ReferenceCode, "Reference code should not be empty")
	} else {
		s.NotNil(getResp.SourceDepositInfo.Crypto, "Crypto deposit info should be present for crypto source")
		s.NotEmpty(getResp.SourceDepositInfo.Crypto.WalletAddress, "Wallet address should not be empty")
	}

	// Get auto conversion rule by idempotency key
	getByKeyResp, err := s.Client.AutoConversionRules.GetRuleByIdempotencyKey(s.Ctx, s.CustomerID, createReq.IdempotencyKey)
	s.Require().NoError(err, "GetRuleByIdempotencyKey should succeed")

	s.Require().NotNil(getByKeyResp, "Get by key response should not be nil")
	s.Equal(createResp.AutoConversionRuleID, getByKeyResp.AutoConversionRuleID, "Rule IDs should match")
	s.T().Logf("Retrieved auto conversion rule by idempotency key:\n%s", PrettyJSON(getByKeyResp))
}

// TestAutoConversionRules_CreateCryptoToFiat tests creating an auto conversion rule for crypto to fiat.
func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_CreateCryptoToFiat() {
	// First ensure we have an external account for fiat withdrawal
	externalAccountID, err := s.EnsureExternalAccount()
	s.Require().NoError(err, "EnsureExternalAccount should succeed")

	network := "POLYGON"
	createReq := &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: uuid.New().String(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USDC",
			Network: "POLYGON",
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:             "USD",
			Network:           &network,
			ExternalAccountID: &externalAccountID,
		},
	}

	createResp, err := s.Client.AutoConversionRules.CreateRule(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRule (crypto to fiat) should succeed")

	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.AutoConversionRuleID, "Rule ID should not be empty")
	s.Equal("USDC", createResp.Source.Asset, "Source asset should be USDC")
	s.Equal("USD", createResp.Destination.Asset, "Destination asset should be USD")
	s.T().Logf("Created crypto-to-fiat auto conversion rule:\n%s", PrettyJSON(createResp))

	// Verify source deposit info is crypto wallet
	getResp, err := s.Client.AutoConversionRules.GetRule(s.Ctx, s.CustomerID, createResp.AutoConversionRuleID)
	s.Require().NoError(err, "GetRule should succeed")
	s.NotNil(getResp.SourceDepositInfo, "SourceDepositInfo should be present")
	s.NotNil(getResp.SourceDepositInfo.Crypto, "Crypto deposit info should be present")
	s.NotEmpty(getResp.SourceDepositInfo.Crypto.WalletAddress, "Wallet address should not be empty")
	s.T().Logf("Crypto deposit wallet address: %s", getResp.SourceDepositInfo.Crypto.WalletAddress)
}

// TestAutoConversionRules_Delete tests deleting an auto conversion rule.
func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_Delete() {
	// First create a rule to delete
	createReq := FakeAutoConversionRuleRequest()

	createResp, err := s.Client.AutoConversionRules.CreateRule(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRule should succeed")

	s.Require().NotNil(createResp, "Create response should not be nil")
	s.T().Logf("Created auto conversion rule for deletion: %s", createResp.AutoConversionRuleID)

	// Delete the rule
	err = s.Client.AutoConversionRules.DeleteRule(s.Ctx, s.CustomerID, createResp.AutoConversionRuleID)
	s.Require().NoError(err, "DeleteRule should succeed")

	s.T().Logf("Successfully deleted auto conversion rule: %s", createResp.AutoConversionRuleID)

	// Verify the rule is now inactive
	getResp, err := s.Client.AutoConversionRules.GetRule(s.Ctx, s.CustomerID, createResp.AutoConversionRuleID)
	s.Require().NoError(err, "GetRule should succeed after deletion")
	s.Equal("INACTIVE", getResp.Status, "Status should be INACTIVE after deletion")
}

// TestAutoConversionRules_ListOrders tests listing orders for an auto conversion rule.
func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_ListOrders() {
	// Ensure we have a rule
	ruleID, err := s.EnsureAutoConversionRule()
	s.Require().NoError(err, "EnsureAutoConversionRule should succeed")

	s.Run("Empty", func() {
		// List orders (may be empty for a new rule)
		resp, err := s.Client.AutoConversionRules.ListOrders(s.Ctx, s.CustomerID, ruleID, nil)
		s.Require().NoError(err, "ListOrders should succeed")
		s.Require().NotNil(resp, "Response should not be nil")
		s.T().Logf("Auto conversion orders: %d orders, total: %d", len(resp.Items), resp.Total)
	})

	s.Run("WithPagination", func() {
		req := &auto_conversion_rules.ListOrdersRequest{
			Page: 1,
			Size: 10,
		}

		resp, err := s.Client.AutoConversionRules.ListOrders(s.Ctx, s.CustomerID, ruleID, req)
		s.Require().NoError(err, "ListOrders with pagination should succeed")
		s.Require().NotNil(resp, "Response should not be nil")
		s.LessOrEqual(len(resp.Items), 10, "Should return at most 10 items")
	})

	s.Run("FilterByStatus", func() {
		req := &auto_conversion_rules.ListOrdersRequest{
			Status: "Completed",
		}

		resp, err := s.Client.AutoConversionRules.ListOrders(s.Ctx, s.CustomerID, ruleID, req)
		s.Require().NoError(err, "ListOrders with status filter should succeed")
		s.Require().NotNil(resp, "Response should not be nil")

		// Verify all returned orders have the expected status
		for i := range resp.Items {
			s.Equal("Completed", resp.Items[i].Status, "Status should match filter")
		}
	})
}

// TestAutoConversionRulesTestSuite runs the auto conversion rules test suite.
func TestAutoConversionRulesTestSuite(t *testing.T) {
	suite.Run(t, new(AutoConversionRulesTestSuite))
}
