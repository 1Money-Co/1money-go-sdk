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

// Package withdraws provides withdrawal transaction functionality.
//
// This package implements the withdrawals service client for the 1Money platform,
// enabling creation and retrieval of withdrawal transactions for both fiat
// (to external bank accounts) and crypto (to wallet addresses).
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Create a fiat withdrawal to external bank account
//	withdrawal, err := client.Withdrawals.CreateWithdrawal(ctx, "customer-id", &withdraws.CreateWithdrawalRequest{
//	    IdempotencyKey:    "unique-key",
//	    Amount:            "100.00",
//	    Asset:             assets.AssetNameUSD,
//	    Network:           assets.NetworkNameUSACH,
//	    ExternalAccountID: "external-account-id",
//	})
package withdraws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// Service defines the withdrawals service interface for managing withdrawal transactions.
type Service interface {
	// CreateWithdrawal creates a new withdrawal transaction.
	CreateWithdrawal(
		ctx context.Context, id svc.CustomerID, req *CreateWithdrawalRequest,
	) (*WithdrawalResponse, error)
	// GetWithdrawal retrieves a specific withdrawal by ID.
	GetWithdrawal(ctx context.Context, id svc.CustomerID, transactionID string) (*WithdrawalResponse, error)
	// GetWithdrawalByIdempotencyKey retrieves a withdrawal by its idempotency key.
	GetWithdrawalByIdempotencyKey(
		ctx context.Context, id svc.CustomerID, idempotencyKey string,
	) (*WithdrawalResponse, error)
}

// FeeMeta represents fee information for a transaction.
type FeeMeta struct {
	// Value is the fee amount.
	Value string `json:"value"`
	// Asset is the fee asset (fiat currency or crypto token).
	Asset string `json:"asset"`
}

// CreateWithdrawal request and response types.
type (
	// CreateWithdrawalRequest represents the request body for creating a withdrawal.
	CreateWithdrawalRequest struct {
		// IdempotencyKey is a unique key to ensure idempotent creation.
		// This is sent as a header, not in the body.
		IdempotencyKey string `json:"-"`
		// Amount is the amount to withdraw.
		Amount string `json:"amount"`
		// Asset is the asset to withdraw.
		Asset assets.AssetName `json:"asset"`
		// Network is the network for the withdrawal.
		Network assets.NetworkName `json:"network"`
		// WalletAddress is the wallet address for crypto withdrawals.
		// Required for crypto asset withdrawals (e.g., USDC, USDT).
		// Cannot be provided together with ExternalAccountID.
		WalletAddress string `json:"wallet_address,omitempty"`
		// ExternalAccountID is the external account ID for fiat withdrawals.
		// Required for fiat currency withdrawals (e.g., USD).
		// Cannot be provided together with WalletAddress.
		ExternalAccountID string `json:"external_account_id,omitempty"`
		// Code is the localized payment code.
		Code string `json:"code,omitempty"`
	}

	// WithdrawalResponse represents the response for a withdrawal transaction.
	WithdrawalResponse struct {
		// TransactionID is the unique transaction identifier.
		TransactionID string `json:"transaction_id"`
		// IdempotencyKey is the idempotency key used for creation.
		IdempotencyKey string `json:"idempotency_key"`
		// Amount is the withdrawal amount.
		Amount string `json:"amount"`
		// Asset is the asset being withdrawn.
		Asset string `json:"asset"`
		// Network is the network used for the withdrawal.
		Network string `json:"network"`
		// WalletAddress is the wallet address for crypto withdrawals.
		WalletAddress string `json:"wallet_address,omitempty"`
		// ExternalAccountID is the external account ID for fiat withdrawals.
		ExternalAccountID string `json:"external_account_id,omitempty"`
		// Code is the localized payment code.
		Code string `json:"code,omitempty"`
		// Status is the current status of the withdrawal.
		Status string `json:"status"`
		// TransactionFee contains the fee information.
		TransactionFee FeeMeta `json:"transaction_fee"`
		// TransactionAction is the transaction action (always "WITHDRAWAL").
		TransactionAction string `json:"transaction_action"`
		// CreatedAt is the withdrawal creation timestamp.
		CreatedAt string `json:"created_at"`
		// ModifiedAt is the withdrawal last modification timestamp.
		ModifiedAt string `json:"modified_at"`
	}
)

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new withdrawals service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateWithdrawal creates a new withdrawal transaction.
func (s *serviceImpl) CreateWithdrawal(
	ctx context.Context,
	id svc.CustomerID,
	req *CreateWithdrawalRequest,
) (*WithdrawalResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/withdrawals", id)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := make(map[string]string)
	if req.IdempotencyKey != "" {
		headers["Idempotency-Key"] = req.IdempotencyKey
	}

	resp, err := s.Do(ctx, &transport.Request{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
	if err != nil {
		return nil, err
	}

	var result WithdrawalResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetWithdrawal retrieves a specific withdrawal by ID.
func (s *serviceImpl) GetWithdrawal(
	ctx context.Context,
	id svc.CustomerID,
	withdrawalID string,
) (*WithdrawalResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/withdrawals/%s", id, withdrawalID)
	return svc.GetJSON[WithdrawalResponse](ctx, s.BaseService, path)
}

// GetWithdrawalByIdempotencyKey retrieves a withdrawal by its idempotency key.
func (s *serviceImpl) GetWithdrawalByIdempotencyKey(
	ctx context.Context,
	id svc.CustomerID,
	idempotencyKey string,
) (*WithdrawalResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/withdrawals", id)
	params := map[string]string{
		"idempotency_key": idempotencyKey,
	}
	return svc.GetJSONWithParams[WithdrawalResponse](ctx, s.BaseService, path, params)
}
