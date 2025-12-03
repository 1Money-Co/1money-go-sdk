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

// Package instructions provides deposit instruction functionality.
//
// This package implements the deposit instructions service client for the 1Money platform,
// enabling retrieval of bank account information for fiat deposits and wallet addresses
// for crypto token deposits.
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
//	// Get deposit instructions
//	instruction, err := client.Instructions.GetDepositInstruction(ctx, "customer-id", assets.AssetNameUSD, assets.NetworkNameUSACH)
package instructions

import (
	"context"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// Service defines the instructions service interface for retrieving deposit instructions.
type Service interface {
	// GetDepositInstruction retrieves deposit instructions for a specific asset and network.
	GetDepositInstruction(
		ctx context.Context, customerID string, asset assets.AssetName, network assets.NetworkName,
	) (*InstructionResponse, error)
}

// AddressDetails represents the address details for bank instructions.
type AddressDetails struct {
	StreetLine1 string `json:"street_line_1,omitempty"`
	StreetLine2 string `json:"street_line_2,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
}

// BankInstruction represents bank account details for fiat deposits.
type BankInstruction struct {
	// BankName is the name of the bank that holds custody over the account.
	BankName string `json:"bank_name,omitempty"`
	// RoutingNumber is the routing number of the account.
	RoutingNumber string `json:"routing_number,omitempty"`
	// AccountHolder is the name of the account holder.
	AccountHolder string `json:"account_holder,omitempty"`
	// AccountNumber is the account number.
	AccountNumber string `json:"account_number,omitempty"`
	// AccountIdentifier is the brokerage account identifier.
	AccountIdentifier string `json:"account_identifier,omitempty"`
	// BICCode is the SWIFT/BIC code.
	BICCode string `json:"bic_code,omitempty"`
	// Address contains address details for the instruction.
	Address *AddressDetails `json:"address,omitempty"`
	// TransactionFee is the fee for the transaction.
	TransactionFee string `json:"transaction_fee"`
}

// WalletInstruction represents wallet address details for crypto deposits.
type WalletInstruction struct {
	// WalletAddress is the wallet address for deposits.
	WalletAddress string `json:"wallet_address,omitempty"`
	// TransactionFee is the fee for the transaction.
	TransactionFee string `json:"transaction_fee"`
}

// InstructionResponse represents the response for deposit instructions.
type InstructionResponse struct {
	// Asset is the asset name for the instruction.
	Asset string `json:"asset"`
	// Network is the network name for the instruction.
	Network string `json:"network"`
	// BankInstruction contains bank details for fiat deposits.
	BankInstruction *BankInstruction `json:"bank_instruction,omitempty"`
	// WalletInstruction contains wallet details for crypto deposits.
	WalletInstruction *WalletInstruction `json:"wallet_instruction,omitempty"`
	// TransactionAction is the transaction action type.
	TransactionAction string `json:"transaction_action"`
	// CreatedAt is the instruction creation timestamp.
	CreatedAt string `json:"created_at"`
	// ModifiedAt is the instruction last modification timestamp.
	ModifiedAt string `json:"modified_at"`
}

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new instructions service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// GetDepositInstruction retrieves deposit instructions for a specific asset and network.
func (s *serviceImpl) GetDepositInstruction(
	ctx context.Context,
	customerID string,
	asset assets.AssetName,
	network assets.NetworkName,
) (*InstructionResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/instruction", customerID)
	params := map[string]string{
		"asset":   string(asset),
		"network": string(network),
	}
	return svc.GetJSONWithParams[InstructionResponse](ctx, s.BaseService, path, params)
}
