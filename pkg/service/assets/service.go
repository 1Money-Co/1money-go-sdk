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

// Package assets provides asset balance management for customer accounts.
//
// This package implements the assets service client for the 1Money platform,
// enabling retrieval of customer asset balances across different networks.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// List all assets for a customer
//	assets, err := client.Assets.ListAssets(ctx, "customer-id", nil)
//
//	// List assets with filters
//	assets, err := client.Assets.ListAssets(ctx, "customer-id", &assets.ListAssetsRequest{
//	    Asset:   assets.AssetNameUSD,
//	    Network: assets.NetworkNameEthereum,
//	})
package assets

import (
	"context"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// Service defines the assets service interface for managing customer asset balances.
type Service interface {
	// ListAssets retrieves all assets for a specific customer.
	// Supports optional filtering by asset name, network, and sort order.
	ListAssets(ctx context.Context, id svc.CustomerID, req *ListAssetsRequest) ([]AssetResponse, error)
}

// ListAssets request and response types.
type (
	// ListAssetsRequest represents the optional query parameters for listing assets.
	ListAssetsRequest struct {
		// Asset filters by specific asset name (e.g., "USD", "USDT").
		Asset AssetName `json:"asset,omitempty"`
		// Network filters by specific network name (e.g., "ETHEREUM", "SOLANA").
		Network NetworkName `json:"network,omitempty"`
		// SortOrder specifies the sort order for results ("ASC" or "DESC").
		SortOrder SortOrder `json:"sort_order,omitempty"`
	}

	// AssetResponse represents a customer's asset balance.
	AssetResponse struct {
		// CustomerID is the unique identifier of the customer.
		CustomerID string `json:"customer_id"`
		// Asset is the asset name/symbol (e.g., "USD", "USDT").
		// Uses string to handle any asset type returned by the API.
		Asset string `json:"asset"`
		// Network is the network name for the asset (optional, nil for fiat).
		// Uses string to handle any network type returned by the API.
		Network *string `json:"network,omitempty"`
		// AvailableAmount is the available balance amount.
		AvailableAmount string `json:"available_amount"`
		// UnavailableAmount is the unavailable/locked balance amount.
		UnavailableAmount string `json:"unavailable_amount"`
		// CreatedAt is the asset record creation timestamp (ISO 8601 format).
		CreatedAt string `json:"created_at"`
		// ModifiedAt is the asset record last modification timestamp (ISO 8601 format).
		ModifiedAt string `json:"modified_at"`
	}
)

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new assets service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// ListAssets retrieves all assets for a specific customer.
func (s *serviceImpl) ListAssets(ctx context.Context, id svc.CustomerID, req *ListAssetsRequest) ([]AssetResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/assets", id)

	params := make(map[string]string)
	if req != nil {
		if req.Asset != "" {
			params["asset"] = string(req.Asset)
		}
		if req.Network != "" {
			params["network"] = string(req.Network)
		}
		if req.SortOrder != "" {
			params["sort_order"] = string(req.SortOrder)
		}
	}

	result, err := svc.GetJSONWithParams[[]AssetResponse](ctx, s.BaseService, path, params)
	if err != nil {
		return nil, err
	}
	return *result, nil
}
