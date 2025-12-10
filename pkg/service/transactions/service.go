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

// Package transactions provides transaction history functionality.
//
// This package implements the transactions service client for the 1Money platform,
// enabling retrieval of transaction history including deposits, withdrawals, and conversions.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// List transactions
//	txns, err := client.Transactions.ListTransactions(ctx, "customer-id", nil)
//
//	// Get a specific transaction
//	txn, err := client.Transactions.GetTransaction(ctx, "customer-id", "transaction-id")
package transactions

import (
	"context"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// Service defines the transactions service interface for retrieving transaction history.
type Service interface {
	// ListTransactions retrieves a list of transactions for a customer.
	ListTransactions(ctx context.Context, id svc.CustomerID, req *ListTransactionsRequest) (*ListTransactionsResponse, error)
	// GetTransaction retrieves a specific transaction by ID.
	GetTransaction(ctx context.Context, id svc.CustomerID, transactionID string) (*TransactionResponse, error)
}

// Common types for transaction operations.
type (
	// TransactionEndpoint represents the source or destination of a transaction.
	TransactionEndpoint struct {
		// Amount is the amount at this endpoint.
		Amount string `json:"amount,omitempty"`
		// Asset is the asset at this endpoint.
		Asset string `json:"asset,omitempty"`
		// Network is the network at this endpoint.
		Network string `json:"network,omitempty"`
		// AddressID is the address identifier (Platform, External Account ID, Wallet Address ID, or Wallet Address).
		AddressID string `json:"address_id"`
	}

	// TransactionResponse represents a transaction.
	TransactionResponse struct {
		// CustomerID is the customer ID.
		CustomerID string `json:"customer_id"`
		// TransactionID is the unique transaction identifier.
		TransactionID string `json:"transaction_id"`
		// IdempotencyKey is the external transaction identifier.
		IdempotencyKey string `json:"idempotency_key"`
		// TransactionAction is the transaction type (DEPOSIT, WITHDRAWAL, CONVERSION).
		TransactionAction string `json:"transaction_action"`
		// Amount is the transaction amount.
		Amount string `json:"amount"`
		// Asset is the transaction asset.
		Asset string `json:"asset,omitempty"`
		// Network is the transaction network.
		Network string `json:"network,omitempty"`
		// TransactionFee is the transaction fee amount.
		TransactionFee string `json:"transaction_fee"`
		// Source contains the transaction source details.
		Source TransactionEndpoint `json:"source"`
		// Destination contains the transaction destination details.
		Destination TransactionEndpoint `json:"destination"`
		// Status is the current transaction status.
		Status string `json:"status"`
		// CreatedAt is the transaction creation timestamp.
		CreatedAt string `json:"created_at"`
		// ModifiedAt is the transaction last modification timestamp.
		ModifiedAt string `json:"modified_at"`
	}
)

// ListTransactions request and response types.
type (
	// ListTransactionsRequest represents optional query parameters for listing transactions.
	ListTransactionsRequest struct {
		// TransactionID filters by specific transaction ID.
		TransactionID string `json:"transaction_id,omitempty"`
		// Asset filters by asset name.
		Asset assets.AssetName `json:"asset,omitempty"`
		// CreatedAfter filters transactions created after this timestamp (RFC3339/ISO 8601 format).
		CreatedAfter string `json:"created_after,omitempty"`
		// CreatedBefore filters transactions created before this timestamp (RFC3339/ISO 8601 format).
		CreatedBefore string `json:"created_before,omitempty"`
		// Page is the page number (starts from 1).
		Page int `json:"page,omitempty"`
		// Size is the number of items per page (1-100).
		Size int `json:"size,omitempty"`
	}

	// ListTransactionsResponse represents the response for listing transactions.
	ListTransactionsResponse struct {
		// List contains the list of transactions.
		List []TransactionResponse `json:"list"`
		// Total is the total number of transactions.
		Total int `json:"total,omitempty"`
	}
)

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new transactions service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// ListTransactions retrieves a list of transactions for a customer.
func (s *serviceImpl) ListTransactions(
	ctx context.Context,
	id svc.CustomerID,
	req *ListTransactionsRequest,
) (*ListTransactionsResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/transactions", id)

	params := make(map[string]string)
	if req != nil {
		if req.TransactionID != "" {
			params["transaction_id"] = req.TransactionID
		}
		if req.Asset != "" {
			params["asset"] = string(req.Asset)
		}
		if req.CreatedAfter != "" {
			params["created_after"] = req.CreatedAfter
		}
		if req.CreatedBefore != "" {
			params["created_before"] = req.CreatedBefore
		}
		if req.Page > 0 {
			params["pagination[page]"] = fmt.Sprintf("%d", req.Page)
		}
		if req.Size > 0 {
			params["pagination[size]"] = fmt.Sprintf("%d", req.Size)
		}
	}

	return svc.GetJSONWithParams[ListTransactionsResponse](ctx, s.BaseService, path, params)
}

// GetTransaction retrieves a specific transaction by ID.
func (s *serviceImpl) GetTransaction(
	ctx context.Context,
	id svc.CustomerID,
	transactionID string,
) (*TransactionResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/transactions/%s", id, transactionID)
	return svc.GetJSON[TransactionResponse](ctx, s.BaseService, path)
}
