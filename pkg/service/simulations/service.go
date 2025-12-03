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

// Package simulations provides transaction simulation functionality.
//
// This package implements the simulations service client for the 1Money platform,
// enabling simulation of deposit transactions for testing purposes.
// NOTE: This service is only available in non-production environments.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Simulate a USD deposit
//	result, err := client.Simulations.SimulateDeposit(ctx, "customer-id", &simulations.SimulateDepositRequest{
//	    Asset:  assets.AssetNameUSD,
//	    Amount: "100.00",
//	})
//
//	// Simulate a crypto token deposit (requires network)
//	result, err := client.Simulations.SimulateDeposit(ctx, "customer-id", &simulations.SimulateDepositRequest{
//	    Asset:   assets.AssetNameUSDT,
//	    Network: simulations.WalletNetworkNameETHEREUM,
//	    Amount:  "100.00",
//	})
package simulations

import (
	"context"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// Service defines the simulations service interface for simulating transactions.
type Service interface {
	// SimulateDeposit simulates a deposit transaction for testing purposes.
	// Only available in non-production environments.
	SimulateDeposit(ctx context.Context, customerID string, req *SimulateDepositRequest) (*SimulateDepositResponse, error)
}

// SimulateDepositRequest represents the request body for simulating a deposit.
type SimulateDepositRequest struct {
	// Asset is the asset to deposit.
	Asset assets.AssetName `json:"asset"`
	// Network is the network for the deposit.
	// Required for token assets (USDT, USDC, MXNB), must be a wallet network (e.g., ETHEREUM).
	// For currency assets (USD), network is optional and will be ignored if provided.
	Network WalletNetworkName `json:"network,omitempty"`
	// Amount is the deposit amount.
	Amount string `json:"amount"`
}

// SimulateDepositResponse represents the response for a simulated deposit.
type SimulateDepositResponse struct {
	// SimulationID is the unique identifier for the simulation.
	SimulationID string `json:"simulation_id"`
	// Status is the transaction status (SUCCESS or REVERSED for simulated deposits).
	Status string `json:"status"`
	// CreatedAt is the transaction creation timestamp.
	CreatedAt string `json:"created_at"`
	// ModifiedAt is the transaction last modification timestamp.
	ModifiedAt string `json:"modified_at"`
}

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new simulations service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// SimulateDeposit simulates a deposit transaction for testing purposes.
func (s *serviceImpl) SimulateDeposit(
	ctx context.Context,
	customerID string,
	req *SimulateDepositRequest,
) (*SimulateDepositResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/simulate-transactions", customerID)
	return svc.PostJSON[SimulateDepositRequest, SimulateDepositResponse](ctx, s.BaseService, path, *req)
}
