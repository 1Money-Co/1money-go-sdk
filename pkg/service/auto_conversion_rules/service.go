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

// Package auto_conversion_rules provides auto conversion rule management for automatic fiat/crypto conversions.
//
// This package implements the auto conversion rules service client for the 1Money platform,
// enabling automatic conversion of deposits between fiat and crypto currencies.
//
// # Overview
//
// Auto conversion rules allow customers to automatically convert incoming deposits
// from one asset to another. For example:
//   - USD deposits can be automatically converted to USDC on Polygon
//   - USDC deposits can be automatically converted to USD for bank withdrawal
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Create an auto conversion rule (USD -> USDC)
//	rule, err := client.AutoConversionRules.CreateRule(ctx, "customer-id", &auto_conversion_rules.CreateRuleRequest{
//	    IdempotencyKey: "unique-key",
//	    Source: auto_conversion_rules.SourceAssetInfo{
//	        Asset:   "USD",
//	        Network: "US_ACH",
//	    },
//	    Destination: auto_conversion_rules.DestinationAssetInfo{
//	        Asset:   "USDC",
//	        Network: ptr("POLYGON"),
//	    },
//	})
//
//	// List auto conversion rules
//	rules, err := client.AutoConversionRules.ListRules(ctx, "customer-id", &auto_conversion_rules.ListRulesRequest{
//	    Page: 1,
//	    Size: 10,
//	})
package auto_conversion_rules

import (
	"context"
	"encoding/json"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// Service defines the auto conversion rules service interface for managing automatic conversions.
type Service interface {
	// CreateRule creates a new auto conversion rule for a customer.
	// The IdempotencyKey in the request is used to ensure idempotent creation.
	CreateRule(ctx context.Context, customerID string, req *CreateRuleRequest) (*RuleResponse, error)

	// GetRule retrieves a specific auto conversion rule by ID.
	GetRule(ctx context.Context, customerID, ruleID string) (*RuleResponse, error)

	// GetRuleByIdempotencyKey retrieves an auto conversion rule by its idempotency key.
	GetRuleByIdempotencyKey(ctx context.Context, customerID, idempotencyKey string) (*RuleResponse, error)

	// ListRules retrieves all auto conversion rules for a customer with pagination.
	ListRules(ctx context.Context, customerID string, req *ListRulesRequest) (*ListRulesResponse, error)

	// DeleteRule soft-deletes an auto conversion rule (marks as inactive).
	DeleteRule(ctx context.Context, customerID, ruleID string) error

	// ListOrders retrieves the execution history (orders) for a specific auto conversion rule.
	ListOrders(ctx context.Context, customerID, ruleID string, req *ListOrdersRequest) (*ListOrdersResponse, error)

	// GetOrder retrieves detailed information about a specific auto conversion order.
	GetOrder(ctx context.Context, customerID, ruleID, orderID string) (*OrderResponse, error)
}

// Common types for asset and amount information.
type (
	// SourceAssetInfo represents the source asset configuration for an auto conversion rule.
	SourceAssetInfo struct {
		// Asset is the source asset name: USD (fiat), USDC, USDT (crypto).
		Asset string `json:"asset"`
		// Network is the source network: US_ACH, US_FEDWIRE, SWIFT for fiat;
		// ETHEREUM, POLYGON, BASE, etc. for crypto.
		Network string `json:"network"`
	}

	// DestinationAssetInfo represents the destination asset configuration for an auto conversion rule.
	DestinationAssetInfo struct {
		// Asset is the destination asset name: USD (fiat), USDC, USDT (crypto).
		Asset string `json:"asset"`
		// Network is the destination network (required for crypto, omit for fiat).
		Network *string `json:"network,omitempty"`
		// WalletAddress is the external wallet address for automatic crypto withdrawal (fiat->crypto only).
		WalletAddress *string `json:"wallet_address,omitempty"`
		// ExternalAccountID is the external account ID for automatic fiat withdrawal (crypto->fiat only).
		ExternalAccountID *string `json:"external_account_id,omitempty"`
	}

	// AmountInfo represents an amount with asset information.
	AmountInfo struct {
		// Amount is the amount value as string (preserves precision).
		Amount string `json:"amount"`
		// Asset is the asset code: USD, USDT, USDC.
		Asset string `json:"asset"`
	}
)

// Deposit information types.
type (
	// BankDepositInfo contains bank deposit information for fiat source.
	BankDepositInfo struct {
		// Network is the bank network type: ach, wire, or swift.
		Network string `json:"network"`
		// ReferenceCode is the reference code (memo) - must be included in wire transfer for proper routing.
		ReferenceCode string `json:"reference_code"`
		// MinimumDepositAmount is the minimum deposit amount required.
		MinimumDepositAmount string `json:"minimum_deposit_amount"`
		// RecipientName is the recipient name on the bank account.
		RecipientName *string `json:"recipient_name,omitempty"`
		// BankName is the receiving bank name.
		BankName *string `json:"bank_name,omitempty"`
		// RoutingNumber is the bank routing number (for US domestic transfers).
		RoutingNumber *string `json:"routing_number,omitempty"`
		// AccountHolderName is the account holder name.
		AccountHolderName *string `json:"account_holder_name,omitempty"`
		// AccountNumber is the bank account number.
		AccountNumber *string `json:"account_number,omitempty"`
		// CountryCode is the country code (ISO 3166-1 alpha-3).
		CountryCode *string `json:"country_code,omitempty"`
		// Street is the bank street address.
		Street *string `json:"street,omitempty"`
		// Additional is additional address information.
		Additional *string `json:"additional,omitempty"`
		// City is the city.
		City *string `json:"city,omitempty"`
		// Region is the state/region.
		Region *string `json:"region,omitempty"`
		// PostalCode is the postal/ZIP code.
		PostalCode *string `json:"postal_code,omitempty"`
		// BICCode is the BIC/SWIFT code for international wire transfers.
		BICCode *string `json:"bic_code,omitempty"`
	}

	// CryptoDepositInfo contains crypto wallet deposit information for crypto source.
	CryptoDepositInfo struct {
		// WalletAddress is the wallet address for receiving crypto deposits.
		WalletAddress string `json:"wallet_address"`
		// MinimumDepositAmount is the minimum deposit amount required.
		MinimumDepositAmount string `json:"minimum_deposit_amount"`
		// ContractAddress is the token contract address (ERC-20). Empty string for native tokens.
		ContractAddress string `json:"contract_address"`
	}

	// SourceDepositInfo represents either bank or crypto deposit information.
	// It can be unmarshaled from JSON that contains either type.
	SourceDepositInfo struct {
		// Bank contains bank deposit info (for fiat source).
		Bank *BankDepositInfo
		// Crypto contains crypto deposit info (for crypto source).
		Crypto *CryptoDepositInfo
	}
)

// UnmarshalJSON implements custom JSON unmarshaling for SourceDepositInfo.
func (s *SourceDepositInfo) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as crypto first (has wallet_address as required field)
	var crypto CryptoDepositInfo
	if err := json.Unmarshal(data, &crypto); err == nil && crypto.WalletAddress != "" {
		s.Crypto = &crypto
		return nil
	}

	// Try to unmarshal as bank (has reference_code as required field)
	var bank BankDepositInfo
	if err := json.Unmarshal(data, &bank); err == nil && bank.ReferenceCode != "" {
		s.Bank = &bank
		return nil
	}

	return fmt.Errorf("unable to unmarshal SourceDepositInfo: unknown type")
}

// MarshalJSON implements custom JSON marshaling for SourceDepositInfo.
func (s SourceDepositInfo) MarshalJSON() ([]byte, error) {
	if s.Bank != nil {
		return json.Marshal(s.Bank)
	}
	if s.Crypto != nil {
		return json.Marshal(s.Crypto)
	}
	return []byte("null"), nil
}

// CreateRule request and response types.
type (
	// createRuleBody is the request body for creating an auto conversion rule (without idempotency key).
	createRuleBody struct {
		Source      SourceAssetInfo      `json:"source"`
		Destination DestinationAssetInfo `json:"destination"`
	}

	// CreateRuleRequest represents the request for creating an auto conversion rule.
	CreateRuleRequest struct {
		// IdempotencyKey is a unique key to ensure idempotent creation.
		// This is sent as a header, not in the body.
		IdempotencyKey string `json:"-"`
		// Source is the source asset and network configuration.
		Source SourceAssetInfo `json:"source"`
		// Destination is the destination asset and optional withdrawal configuration.
		Destination DestinationAssetInfo `json:"destination"`
	}

	// RuleResponse represents the response data for an auto conversion rule.
	RuleResponse struct {
		// AutoConversionRuleID is the unique auto conversion rule identifier (UUID).
		AutoConversionRuleID string `json:"auto_conversion_rule_id"`
		// IdempotencyKey is the idempotency key provided during creation.
		IdempotencyKey string `json:"idempotency_key"`
		// Nickname is the auto-generated nickname based on source/destination.
		Nickname string `json:"nickname"`
		// Status is the rule status: ACTIVE or INACTIVE.
		Status string `json:"status"`
		// Source is the source asset and network configuration.
		Source SourceAssetInfo `json:"source"`
		// Destination is the destination asset, network, and withdrawal configuration.
		Destination DestinationAssetInfo `json:"destination"`
		// SourceDepositInfo contains deposit info (bank or wallet). Only included in retrieve responses.
		SourceDepositInfo *SourceDepositInfo `json:"source_deposit_info,omitempty"`
		// CreatedAt is the rule creation timestamp (ISO 8601).
		CreatedAt string `json:"created_at"`
		// ModifiedAt is the last modification timestamp (ISO 8601).
		ModifiedAt string `json:"modified_at"`
	}
)

// ListRules request and response types.
type (
	// ListRulesRequest represents the pagination parameters for listing auto conversion rules.
	ListRulesRequest struct {
		// Page is the page number (starts from 1, default: 1).
		Page int `json:"page,omitempty"`
		// Size is the number of items per page (1-100, default: 10).
		Size int `json:"size,omitempty"`
	}

	// ListRulesResponse represents the paginated response for listing auto conversion rules.
	ListRulesResponse struct {
		// Total is the total number of auto conversion rules matching the query.
		Total int64 `json:"total"`
		// Items is the list of auto conversion rules.
		Items []RuleResponse `json:"items"`
	}
)

// Order types for auto conversion order management.
type (
	// OrderReceipt contains fee breakdown for an auto conversion order.
	OrderReceipt struct {
		// Initial is the initial deposit amount received.
		Initial AmountInfo `json:"initial"`
		// DeveloperFee is the developer fee (reserved for future use, currently 0).
		DeveloperFee AmountInfo `json:"developer_fee"`
		// DepositFee is the fee charged for the deposit operation.
		DepositFee AmountInfo `json:"deposit_fee"`
		// ConversionFee is the fee charged for currency conversion.
		ConversionFee AmountInfo `json:"conversion_fee"`
		// WithdrawalFee is the fee charged for withdrawal (only present if rule includes withdrawal step).
		WithdrawalFee *AmountInfo `json:"withdrawal_fee,omitempty"`
	}

	// OrderResponse represents the response data for an auto conversion order.
	OrderResponse struct {
		// AutoConversionOrderID is the unique order identifier (UUID).
		AutoConversionOrderID string `json:"auto_conversion_order_id"`
		// AutoConversionRuleID is the parent auto conversion rule ID (UUID).
		AutoConversionRuleID string `json:"auto_conversion_rule_id"`
		// Status is the order status: Init, Deposit Completed, Conversion Completed, Completed,
		// Deposit Failed, Conversion Failed, Withdrawal Failed.
		Status string `json:"status"`
		// Source is the source asset and network.
		Source SourceAssetInfo `json:"source"`
		// Destination is the destination asset and network.
		Destination DestinationAssetInfo `json:"destination"`
		// Receipt is the fee breakdown for this order.
		Receipt OrderReceipt `json:"receipt"`
		// CreatedAt is the order creation timestamp (ISO 8601).
		CreatedAt string `json:"created_at"`
		// UpdatedAt is the last update timestamp (ISO 8601).
		UpdatedAt string `json:"updated_at"`
	}

	// ListOrdersRequest represents the parameters for listing auto conversion orders.
	ListOrdersRequest struct {
		// Status filters by order status (optional).
		Status string `json:"status,omitempty"`
		// Page is the page number (starts from 1, default: 1).
		Page int `json:"page,omitempty"`
		// Size is the number of items per page (1-100, default: 10).
		Size int `json:"size,omitempty"`
	}

	// ListOrdersResponse represents the paginated response for listing auto conversion orders.
	ListOrdersResponse struct {
		// Total is the total number of orders matching the query.
		Total int64 `json:"total"`
		// Items is the list of auto conversion orders.
		Items []OrderResponse `json:"items"`
	}
)

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new auto conversion rules service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateRule creates a new auto conversion rule for a customer.
func (s *serviceImpl) CreateRule(
	ctx context.Context,
	customerID string,
	req *CreateRuleRequest,
) (*RuleResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules", customerID)

	headers := make(map[string]string)
	if req.IdempotencyKey != "" {
		headers["Idempotency-Key"] = req.IdempotencyKey
	}

	body := createRuleBody{
		Source:      req.Source,
		Destination: req.Destination,
	}

	return svc.PostJSONWithHeaders[createRuleBody, RuleResponse](ctx, s.BaseService, path, body, headers)
}

// GetRule retrieves a specific auto conversion rule by ID.
func (s *serviceImpl) GetRule(
	ctx context.Context,
	customerID, ruleID string,
) (*RuleResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules/%s", customerID, ruleID)
	return svc.GetJSON[RuleResponse](ctx, s.BaseService, path)
}

// GetRuleByIdempotencyKey retrieves an auto conversion rule by its idempotency key.
func (s *serviceImpl) GetRuleByIdempotencyKey(
	ctx context.Context,
	customerID, idempotencyKey string,
) (*RuleResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules", customerID)
	params := map[string]string{
		"idempotency_key": idempotencyKey,
	}
	return svc.GetJSONWithParams[RuleResponse](ctx, s.BaseService, path, params)
}

// ListRules retrieves all auto conversion rules for a customer with pagination.
func (s *serviceImpl) ListRules(
	ctx context.Context,
	customerID string,
	req *ListRulesRequest,
) (*ListRulesResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules/list", customerID)

	params := make(map[string]string)
	if req != nil {
		if req.Page > 0 {
			params["pagination[page]"] = fmt.Sprintf("%d", req.Page)
		}
		if req.Size > 0 {
			params["pagination[size]"] = fmt.Sprintf("%d", req.Size)
		}
	}

	return svc.GetJSONWithParams[ListRulesResponse](ctx, s.BaseService, path, params)
}

// DeleteRule soft-deletes an auto conversion rule (marks as inactive).
func (s *serviceImpl) DeleteRule(
	ctx context.Context,
	customerID, ruleID string,
) error {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules/%s", customerID, ruleID)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}

// ListOrders retrieves the execution history (orders) for a specific auto conversion rule.
func (s *serviceImpl) ListOrders(
	ctx context.Context,
	customerID, ruleID string,
	req *ListOrdersRequest,
) (*ListOrdersResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules/%s/orders", customerID, ruleID)

	params := make(map[string]string)
	if req != nil {
		if req.Status != "" {
			params["status"] = req.Status
		}
		if req.Page > 0 {
			params["pagination[page]"] = fmt.Sprintf("%d", req.Page)
		}
		if req.Size > 0 {
			params["pagination[size]"] = fmt.Sprintf("%d", req.Size)
		}
	}

	return svc.GetJSONWithParams[ListOrdersResponse](ctx, s.BaseService, path, params)
}

// GetOrder retrieves detailed information about a specific auto conversion order.
func (s *serviceImpl) GetOrder(
	ctx context.Context,
	customerID, ruleID, orderID string,
) (*OrderResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/auto-conversion-rules/%s/orders/%s", customerID, ruleID, orderID)
	return svc.GetJSON[OrderResponse](ctx, s.BaseService, path)
}
