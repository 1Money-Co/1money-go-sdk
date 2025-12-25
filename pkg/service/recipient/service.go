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

// Package recipient provides Recipient management for customer accounts.
//
// This package implements the Recipient service client for the 1Money platform,
// enabling management of third-party recipients (counterparties) for payments.
// Recipients can be individuals or companies and can have bank accounts and
// wallet addresses associated with them for fiat and crypto transfers.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/recipient"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Create a recipient
//	resp, err := client.Recipient.CreateRecipient(ctx, "customer-id", &recipient.CreateRecipientRequest{
//	    IdempotencyKey:  "unique-key",
//	    RecipientType:   recipient.RecipientTypeIndividual,
//	    FirstName:       ptr("John"),
//	    LastName:        ptr("Doe"),
//	    Email:           ptr("john.doe@example.com"),
//	})
//
//	// List recipients
//	list, err := client.Recipient.ListRecipients(ctx, "customer-id", &recipient.ListRecipientsRequest{
//	    Page: 1,
//	    Size: 10,
//	})
package recipient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	"github.com/1Money-Co/1money-go-sdk/pkg/common"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// Service defines the recipient service interface for managing recipient operations.
type Service interface {
	// CreateRecipient creates a new recipient for a customer.
	// The IdempotencyKey in the request is required for idempotent creation.
	CreateRecipient(ctx context.Context, cid svc.CustomerID, req *CreateRecipientRequest) (*RecipientResponse, error)
	// GetRecipient retrieves a specific recipient by ID.
	GetRecipient(ctx context.Context, cid svc.CustomerID, rid svc.RecipientID) (*RecipientResponse, error)
	// GetRecipientByIdempotencyKey retrieves a recipient by its idempotency key.
	GetRecipientByIdempotencyKey(ctx context.Context, cid svc.CustomerID, key svc.IdempotencyKey) (*RecipientResponse, error)
	// ListRecipients retrieves all recipients for a customer with optional filtering and pagination.
	ListRecipients(ctx context.Context, cid svc.CustomerID, req *ListRecipientsRequest) (*ListRecipientsResponse, error)
	// UpdateRecipient updates an existing recipient (full replacement).
	UpdateRecipient(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, req *UpdateRecipientRequest,
	) (*RecipientResponse, error)
	// DeleteRecipient soft-deletes a recipient.
	DeleteRecipient(ctx context.Context, cid svc.CustomerID, rid svc.RecipientID) error

	// AddBankAccount adds a new bank account to an existing recipient.
	// The IdempotencyKey in the request is required for idempotent creation.
	AddBankAccount(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, req *BankAccountRequest,
	) (*BankAccountResponse, error)
	// GetBankAccountByIdempotencyKey retrieves a bank account by its idempotency key.
	GetBankAccountByIdempotencyKey(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, key svc.IdempotencyKey,
	) (*BankAccountResponse, error)
	// ListBankAccounts retrieves all bank accounts for a recipient.
	ListBankAccounts(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, req *ListBankAccountsRequest,
	) (*ListBankAccountsResponse, error)
	// DeleteBankAccount removes a bank account from a recipient.
	DeleteBankAccount(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, eid svc.ExternalAccountID,
	) error

	// AddWalletAddress adds a new wallet address to an existing recipient.
	AddWalletAddress(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, req *WalletAddressRequest,
	) (*WalletAddressResponse, error)
	// ListWalletAddresses retrieves all wallet addresses for a recipient.
	ListWalletAddresses(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, req *ListWalletAddressesRequest,
	) (*ListWalletAddressesResponse, error)
	// DeleteWalletAddress removes a wallet address from a recipient.
	DeleteWalletAddress(
		ctx context.Context, cid svc.CustomerID, rid svc.RecipientID, wid svc.WalletAddressID,
	) error
}

// Address represents recipient address information.
type Address struct {
	// CountryCode is the ISO 3166-1 alpha-3 country code in uppercase (e.g., USA, GBR, DEU).
	CountryCode common.CountryCode `json:"country_code"`
	// AddressLine1 is the primary street address.
	AddressLine1 string `json:"address_line1"`
	// AddressLine2 is the secondary address line (optional).
	AddressLine2 *string `json:"address_line2,omitempty"`
	// City is the city name.
	City string `json:"city"`
	// Region is the state/region/province (optional).
	Region *string `json:"region,omitempty"`
	// PostalCode is the postal code.
	PostalCode string `json:"postal_code"`
}

// AddressDetails represents recipient address information in responses.
type AddressDetails struct {
	// FullAddress is the formatted full address.
	FullAddress string `json:"full_address"`
	// CountryCode is the ISO 3166-1 alpha-3 country code.
	CountryCode string `json:"country_code"`
	// AddressLine1 is the primary street address.
	AddressLine1 string `json:"address_line1"`
	// AddressLine2 is the secondary address line (optional).
	AddressLine2 *string `json:"address_line2,omitempty"`
	// City is the city name.
	City string `json:"city"`
	// Region is the state/region/province (optional).
	Region *string `json:"region,omitempty"`
	// PostalCode is the postal code.
	PostalCode string `json:"postal_code"`
}

// IntermediaryBank represents intermediary bank details for international wire transfers.
type IntermediaryBank struct {
	// InstitutionID is the intermediary institution identifier (SWIFT code or ABA routing number).
	InstitutionID string `json:"institution_id"`
	// InstitutionName is the full legal name of the intermediary bank (optional).
	InstitutionName *string `json:"institution_name,omitempty"`
}

// CreateRecipientRequest represents the request body for creating a new recipient.
type CreateRecipientRequest struct {
	// IdempotencyKey is a unique key to ensure idempotent creation (sent as header).
	IdempotencyKey string `json:"-"`
	// RecipientType is the type of recipient (individual or company).
	RecipientType RecipientType `json:"recipient_type"`
	// FirstName is the first name (required for individual type).
	FirstName *string `json:"first_name,omitempty"`
	// LastName is the last name (required for individual type).
	LastName *string `json:"last_name,omitempty"`
	// CompanyName is the company name (required for company type).
	CompanyName *string `json:"company_name,omitempty"`
	// Nickname is a display nickname for the recipient.
	Nickname *string `json:"nickname,omitempty"`
	// Email is the email address of the recipient.
	Email *string `json:"email,omitempty"`
	// Relationship is the relationship with the recipient.
	Relationship *RecipientRelationship `json:"relationship,omitempty"`
	// Address is the address information.
	Address *Address `json:"address,omitempty"`
	// BankAccounts are bank accounts to add (optional, for one-shot creation).
	BankAccounts []BankAccountRequest `json:"bank_accounts,omitempty"`
	// WalletAddresses are wallet addresses to add (optional, for one-shot creation).
	WalletAddresses []WalletAddressRequest `json:"wallet_addresses,omitempty"`
}

// RecipientResponse represents the response object for recipient information.
type RecipientResponse struct {
	// RecipientID is the unique identifier for the recipient.
	RecipientID string `json:"recipient_id"`
	// CustomerID is the customer ID that owns this recipient.
	CustomerID string `json:"customer_id"`
	// RecipientType is the type of recipient (individual or company).
	RecipientType RecipientType `json:"recipient_type"`
	// FullName is the full display name.
	FullName string `json:"full_name"`
	// Nickname is the display nickname.
	Nickname *string `json:"nickname,omitempty"`
	// Email is the email address.
	Email *string `json:"email,omitempty"`
	// Relationship is the relationship with the recipient.
	Relationship *RecipientRelationship `json:"relationship,omitempty"`
	// Status is the current status.
	Status RecipientStatus `json:"status"`
	// Address is the address information.
	Address *AddressDetails `json:"address,omitempty"`
	// CreatedAt is the creation timestamp (ISO 8601 format).
	CreatedAt string `json:"created_at"`
	// ModifiedAt is the last modification timestamp (ISO 8601 format).
	ModifiedAt string `json:"modified_at"`
}

// UpdateRecipientRequest represents the request body for updating a recipient (PUT - full replacement).
type UpdateRecipientRequest struct {
	// RecipientType is the type of recipient (individual or company).
	RecipientType RecipientType `json:"recipient_type"`
	// FirstName is the first name (required for individual type).
	FirstName *string `json:"first_name,omitempty"`
	// LastName is the last name (required for individual type).
	LastName *string `json:"last_name,omitempty"`
	// CompanyName is the company name (required for company type).
	CompanyName *string `json:"company_name,omitempty"`
	// Nickname is a display nickname for the recipient.
	Nickname *string `json:"nickname,omitempty"`
	// Email is the email address of the recipient.
	Email *string `json:"email,omitempty"`
	// Relationship is the relationship with the recipient.
	Relationship *RecipientRelationship `json:"relationship,omitempty"`
	// Address is the address information.
	Address *Address `json:"address,omitempty"`
}

// ListRecipientsRequest represents the request parameters for listing recipients.
type ListRecipientsRequest struct {
	// Search is a search string to filter by name, nickname, or email.
	Search *string `json:"search,omitempty"`
	// Page is the page number (starts from 1, default: 1).
	Page int `json:"page,omitempty"`
	// Size is the number of items per page (1-100, default: 10).
	Size int `json:"size,omitempty"`
}

// ListRecipientsResponse represents the response for listing recipients.
type ListRecipientsResponse struct {
	// List is the list of recipients.
	List []RecipientResponse `json:"list"`
	// Total is the total number of recipients.
	Total *int64 `json:"total,omitempty"`
}

// BankAccountRequest represents the request body for adding a bank account to a recipient.
type BankAccountRequest struct {
	// IdempotencyKey is a unique key to ensure idempotent creation (sent as header).
	IdempotencyKey string `json:"-"`
	// Network is the bank network type (US_FEDWIRE, US_ACH, SWIFT).
	Network common.BankNetworkName `json:"network"`
	// Currency is the currency of the bank account (USD).
	Currency string `json:"currency"`
	// CountryCode is the country where the bank is located (ISO 3166-1 alpha-3).
	CountryCode common.CountryCode `json:"country_code"`
	// AccountNumber is the bank account number.
	AccountNumber string `json:"account_number"`
	// InstitutionID is the institution routing ID (ABA or SWIFT code).
	InstitutionID string `json:"institution_id"`
	// InstitutionName is the bank institution name.
	InstitutionName string `json:"institution_name"`
	// InstitutionClearingCode is an additional clearing code (optional).
	InstitutionClearingCode *string `json:"institution_clearing_code,omitempty"`
	// IntermediaryBank is the intermediary bank info for international wires (optional).
	IntermediaryBank *IntermediaryBank `json:"intermediary_bank,omitempty"`
}

// BankAccountResponse represents the response for recipient bank account.
type BankAccountResponse struct {
	// ExternalAccountID is the unique identifier for the bank account.
	ExternalAccountID string `json:"external_account_id"`
	// RecipientID is the associated recipient ID.
	RecipientID string `json:"recipient_id"`
	// CustomerID is the customer ID.
	CustomerID string `json:"customer_id"`
	// IdempotencyKey is the idempotency key if provided.
	IdempotencyKey string `json:"idempotency_key"`
	// Status is the current status.
	Status common.BankAccountStatus `json:"status"`
	// Network is the bank network type.
	Network common.BankNetworkName `json:"network"`
	// AccountHolderName is the account holder name.
	AccountHolderName string `json:"account_holder_name"`
	// Currency is the currency code.
	Currency string `json:"currency"`
	// CountryCode is the country code (ISO 3166-1 alpha-3).
	CountryCode common.CountryCode `json:"country_code"`
	// AccountNumber is the masked account number.
	AccountNumber string `json:"account_number"`
	// InstitutionID is the institution routing ID.
	InstitutionID string `json:"institution_id"`
	// InstitutionName is the bank name.
	InstitutionName string `json:"institution_name"`
	// InstitutionClearingCode is an additional clearing code (optional).
	InstitutionClearingCode *string `json:"institution_clearing_code,omitempty"`
	// IntermediaryBank is the intermediary bank info (optional).
	IntermediaryBank *IntermediaryBank `json:"intermediary_bank,omitempty"`
	// ReferenceCode is a reference code (optional).
	ReferenceCode *string `json:"reference_code,omitempty"`
	// CreatedAt is the creation timestamp (ISO 8601 format).
	CreatedAt string `json:"created_at"`
	// ModifiedAt is the last modification timestamp (ISO 8601 format).
	ModifiedAt string `json:"modified_at"`
}

// ListBankAccountsRequest represents optional query parameters for listing bank accounts.
type ListBankAccountsRequest struct {
	// Currency filters by currency code (e.g., USD).
	Currency *string `json:"currency,omitempty"`
	// Network filters by bank network type (US_ACH, SWIFT, US_FEDWIRE).
	Network *common.BankNetworkName `json:"network,omitempty"`
}

// ListBankAccountsResponse represents the response for listing bank accounts.
type ListBankAccountsResponse struct {
	// List is the list of bank accounts.
	List []BankAccountResponse `json:"list"`
	// Total is the total number of bank accounts.
	Total *int64 `json:"total,omitempty"`
}

// WalletAddressRequest represents the request body for adding a wallet address to a recipient.
type WalletAddressRequest struct {
	// Blockchain is the blockchain network (ETHEREUM, POLYGON, SOLANA, ARBITRUM, AVALANCHE, BASE, BNBCHAIN, TRON).
	Blockchain string `json:"blockchain"`
	// Token is the token symbol (USDC, USDT, etc.).
	Token string `json:"token"`
	// Address is the wallet address.
	Address string `json:"address"`
	// Nickname is an optional nickname for the address.
	Nickname *string `json:"nickname,omitempty"`
}

// WalletAddressResponse represents the response for recipient wallet address.
type WalletAddressResponse struct {
	// WalletAddressID is the unique identifier.
	WalletAddressID string `json:"wallet_address_id"`
	// RecipientID is the associated recipient ID.
	RecipientID string `json:"recipient_id"`
	// CustomerID is the customer ID.
	CustomerID string `json:"customer_id"`
	// Blockchain is the blockchain network.
	Blockchain string `json:"blockchain"`
	// Token is the token symbol.
	Token string `json:"token"`
	// Address is the full wallet address.
	Address string `json:"address"`
	// Nickname is the display nickname.
	Nickname string `json:"nickname"`
	// CreatedAt is the creation timestamp (ISO 8601 format).
	CreatedAt string `json:"created_at"`
	// ModifiedAt is the last modification timestamp (ISO 8601 format).
	ModifiedAt string `json:"modified_at"`
}

// ListWalletAddressesRequest represents optional query parameters for listing wallet addresses.
type ListWalletAddressesRequest struct {
	// Blockchain filters by blockchain network.
	Blockchain *string `json:"blockchain,omitempty"`
	// Token filters by token symbol.
	Token *string `json:"token,omitempty"`
}

// ListWalletAddressesResponse represents the response for listing wallet addresses.
type ListWalletAddressesResponse struct {
	// List is the list of wallet addresses.
	List []WalletAddressResponse `json:"list"`
	// Total is the total number of wallet addresses.
	Total *int64 `json:"total,omitempty"`
}

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new recipient service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateRecipient creates a new recipient for a customer.
func (s *serviceImpl) CreateRecipient(
	ctx context.Context,
	cid svc.CustomerID,
	req *CreateRecipientRequest,
) (*RecipientResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients", cid)

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

	var result RecipientResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetRecipient retrieves a specific recipient by ID.
func (s *serviceImpl) GetRecipient(ctx context.Context, cid svc.CustomerID, rid svc.RecipientID) (*RecipientResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s", cid, rid)
	return svc.GetJSON[RecipientResponse](ctx, s.BaseService, path)
}

// GetRecipientByIdempotencyKey retrieves a recipient by its idempotency key.
func (s *serviceImpl) GetRecipientByIdempotencyKey(
	ctx context.Context,
	cid svc.CustomerID,
	key svc.IdempotencyKey,
) (*RecipientResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients", cid)
	params := map[string]string{
		"idempotency_key": key,
	}
	return svc.GetJSONWithParams[RecipientResponse](ctx, s.BaseService, path, params)
}

// ListRecipients retrieves all recipients for a customer with optional filtering and pagination.
func (s *serviceImpl) ListRecipients(
	ctx context.Context,
	cid svc.CustomerID,
	req *ListRecipientsRequest,
) (*ListRecipientsResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/list", cid)

	params := make(map[string]string)
	if req != nil {
		if req.Search != nil && *req.Search != "" {
			params["search"] = *req.Search
		}
		if req.Page > 0 {
			params["page"] = fmt.Sprintf("%d", req.Page)
		}
		if req.Size > 0 {
			params["size"] = fmt.Sprintf("%d", req.Size)
		}
	}

	return svc.GetJSONWithParams[ListRecipientsResponse](ctx, s.BaseService, path, params)
}

// UpdateRecipient updates an existing recipient (full replacement).
func (s *serviceImpl) UpdateRecipient(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	req *UpdateRecipientRequest,
) (*RecipientResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s", cid, rid)
	return svc.PutJSON[*UpdateRecipientRequest, RecipientResponse](ctx, s.BaseService, path, req)
}

// DeleteRecipient soft-deletes a recipient.
func (s *serviceImpl) DeleteRecipient(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
) error {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s", cid, rid)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}

// AddBankAccount adds a new bank account to an existing recipient.
func (s *serviceImpl) AddBankAccount(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	req *BankAccountRequest,
) (*BankAccountResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/external-accounts", cid, rid)

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

	var result BankAccountResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetBankAccountByIdempotencyKey retrieves a bank account by its idempotency key.
func (s *serviceImpl) GetBankAccountByIdempotencyKey(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	key svc.IdempotencyKey,
) (*BankAccountResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/external-accounts", cid, rid)
	params := map[string]string{
		"idempotency_key": key,
	}
	return svc.GetJSONWithParams[BankAccountResponse](ctx, s.BaseService, path, params)
}

// ListBankAccounts retrieves all bank accounts for a recipient.
func (s *serviceImpl) ListBankAccounts(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	req *ListBankAccountsRequest,
) (*ListBankAccountsResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/external-accounts/list", cid, rid)

	params := make(map[string]string)
	if req != nil {
		if req.Currency != nil && *req.Currency != "" {
			params["currency"] = *req.Currency
		}
		if req.Network != nil && *req.Network != "" {
			params["network"] = string(*req.Network)
		}
	}

	return svc.GetJSONWithParams[ListBankAccountsResponse](ctx, s.BaseService, path, params)
}

// DeleteBankAccount removes a bank account from a recipient.
func (s *serviceImpl) DeleteBankAccount(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	eid svc.ExternalAccountID,
) error {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/external-accounts/%s", cid, rid, eid)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}

// AddWalletAddress adds a new wallet address to an existing recipient.
func (s *serviceImpl) AddWalletAddress(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	req *WalletAddressRequest,
) (*WalletAddressResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/wallet-addresses", cid, rid)
	return svc.PostJSON[*WalletAddressRequest, WalletAddressResponse](ctx, s.BaseService, path, req)
}

// ListWalletAddresses retrieves all wallet addresses for a recipient.
func (s *serviceImpl) ListWalletAddresses(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	req *ListWalletAddressesRequest,
) (*ListWalletAddressesResponse, error) {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/wallet-addresses/list", cid, rid)

	params := make(map[string]string)
	if req != nil {
		if req.Blockchain != nil && *req.Blockchain != "" {
			params["blockchain"] = *req.Blockchain
		}
		if req.Token != nil && *req.Token != "" {
			params["token"] = *req.Token
		}
	}

	return svc.GetJSONWithParams[ListWalletAddressesResponse](ctx, s.BaseService, path, params)
}

// DeleteWalletAddress removes a wallet address from a recipient.
func (s *serviceImpl) DeleteWalletAddress(
	ctx context.Context,
	cid svc.CustomerID,
	rid svc.RecipientID,
	wid svc.WalletAddressID,
) error {
	path := fmt.Sprintf("/v1/customers/%s/recipients/%s/wallet-addresses/%s", cid, rid, wid)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}
