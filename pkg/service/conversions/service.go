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

// Package conversions provides asset conversion functionality.
//
// This package implements the conversions service client for the 1Money platform,
// enabling creation of quotes for converting between assets and executing hedges.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Create a quote for conversion
//	quote, err := client.Conversions.CreateQuote(ctx, "customer-id", &conversions.CreateQuoteRequest{
//	    FromAsset: conversions.AssetInfo{Asset: assets.AssetNameUSDT, Amount: "100.00", Network: conversions.WalletNetworkNameETHEREUM},
//	    ToAsset:   conversions.AssetInfo{Asset: assets.AssetNameUSD},
//	})
//
//	// Execute the hedge
//	order, err := client.Conversions.CreateHedge(ctx, "customer-id", &conversions.CreateHedgeRequest{
//	    QuoteID: quote.QuoteID,
//	})
package conversions

import (
	"context"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// Service defines the conversions service interface for managing asset conversions.
type Service interface {
	// CreateQuote creates a quote for converting between assets.
	CreateQuote(ctx context.Context, id svc.CustomerID, req *CreateQuoteRequest) (*QuoteResponse, error)
	// CreateHedge executes a hedge for a conversion quote.
	CreateHedge(ctx context.Context, id svc.CustomerID, req *CreateHedgeRequest) (*OrderResponse, error)
	// GetOrder retrieves a conversion order by ID.
	GetOrder(ctx context.Context, id svc.CustomerID, orderID string) (*OrderResponse, error)
}

// AssetInfo represents asset information for conversion quotes.
type AssetInfo struct {
	// Amount is the asset amount (optional, either from or to must have amount).
	Amount string `json:"amount,omitempty"`
	// Asset is the asset name.
	Asset assets.AssetName `json:"asset"`
	// Network is the network name (required for crypto assets).
	Network WalletNetworkName `json:"network,omitempty"`
}

// CreateQuote request and response types.
type (
	// CreateQuoteRequest represents the request body for creating a conversion quote.
	CreateQuoteRequest struct {
		// FromAsset is the source asset information.
		FromAsset AssetInfo `json:"from_asset"`
		// ToAsset is the destination asset information.
		ToAsset AssetInfo `json:"to_asset"`
	}

	// QuoteResponse represents the response for a conversion quote.
	QuoteResponse struct {
		// QuoteID is the unique quote identifier.
		QuoteID string `json:"quote_id"`
		// UserPayAmount is the amount the user will pay.
		UserPayAmount string `json:"user_pay_amount"`
		// UserPayAsset is the asset the user will pay.
		UserPayAsset string `json:"user_pay_asset"`
		// UserPayNetwork is the network for the payment asset.
		UserPayNetwork string `json:"user_pay_network"`
		// UserObtainAmount is the amount the user will receive.
		UserObtainAmount string `json:"user_obtain_amount"`
		// UserObtainAsset is the asset the user will receive.
		UserObtainAsset string `json:"user_obtain_asset"`
		// UserObtainNetwork is the network for the received asset.
		UserObtainNetwork string `json:"user_obtain_network"`
		// Rate is the conversion rate.
		Rate string `json:"rate"`
		// ExpireTime is the quote expiration time in seconds.
		ExpireTime int `json:"expire_time"`
		// ValidUntilTimestamp is the timestamp until which the quote is valid.
		ValidUntilTimestamp string `json:"valid_until_timestamp"`
	}
)

// CreateHedge request and response types.
type (
	// CreateHedgeRequest represents the request body for executing a conversion hedge.
	CreateHedgeRequest struct {
		// QuoteID is the quote ID to execute.
		QuoteID string `json:"quote_id"`
	}

	// OrderResponse represents the response for a conversion order.
	OrderResponse struct {
		// OrderID is the unique order identifier.
		OrderID string `json:"order_id"`
		// OrderStatus is the current order status.
		OrderStatus string `json:"order_status"`
		// QuoteID is the quote ID used for the order.
		QuoteID string `json:"quote_id"`
		// UserPayAmount is the amount the user paid.
		UserPayAmount string `json:"user_pay_amount"`
		// UserPayAsset is the asset the user paid.
		UserPayAsset string `json:"user_pay_asset"`
		// UserPayNetwork is the network for the payment asset.
		UserPayNetwork string `json:"user_pay_network"`
		// UserObtainAmount is the amount the user received.
		UserObtainAmount string `json:"user_obtain_amount"`
		// UserObtainAsset is the asset the user received.
		UserObtainAsset string `json:"user_obtain_asset"`
		// UserObtainNetwork is the network for the received asset.
		UserObtainNetwork string `json:"user_obtain_network"`
		// Rate is the conversion rate.
		Rate string `json:"rate"`
		// Fee is the fee amount.
		Fee string `json:"fee"`
		// FeeCurrency is the fee currency.
		FeeCurrency string `json:"fee_currency"`
	}
)

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new conversions service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateQuote creates a quote for converting between assets.
func (s *serviceImpl) CreateQuote(
	ctx context.Context,
	id svc.CustomerID,
	req *CreateQuoteRequest,
) (*QuoteResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/conversions/quote", id)
	return svc.PostJSON[CreateQuoteRequest, QuoteResponse](ctx, s.BaseService, path, *req)
}

// CreateHedge executes a hedge for a conversion quote.
func (s *serviceImpl) CreateHedge(
	ctx context.Context,
	id svc.CustomerID,
	req *CreateHedgeRequest,
) (*OrderResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/conversions/hedge", id)
	return svc.PostJSON[CreateHedgeRequest, OrderResponse](ctx, s.BaseService, path, *req)
}

// GetOrder retrieves a conversion order by ID.
func (s *serviceImpl) GetOrder(
	ctx context.Context,
	id svc.CustomerID,
	orderID string,
) (*OrderResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/conversions/order", id)
	params := map[string]string{
		"order_id": orderID,
	}
	return svc.GetJSONWithParams[OrderResponse](ctx, s.BaseService, path, params)
}
