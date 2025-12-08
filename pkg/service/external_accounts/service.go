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

// Package external_accounts provides external bank account management for customer accounts.
//
// This package implements the external accounts service client for the 1Money platform,
// enabling management of external bank accounts for fiat transfers (deposits and withdrawals).
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Create an external bank account
//	account, err := client.ExternalAccounts.CreateExternalAccount(ctx, "customer-id", &external_accounts.CreateReq{
//	    IdempotencyKey:  "unique-key",
//	    Network:         external_accounts.BankNetworkNameUSACH,
//	    Currency:        external_accounts.CurrencyUSD,
//	    CountryCode:     external_accounts.CountryCodeUSA,
//	    AccountNumber:   "123456789",
//	    InstitutionID:   "021000021",
//	    InstitutionName: "Bank of America",
//	})
package external_accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// Service defines the external accounts service interface for managing customer external bank accounts.
type Service interface {
	// CreateExternalAccount creates a new external bank account for a customer.
	// The IdempotencyKey in the request is used to ensure idempotent creation.
	CreateExternalAccount(ctx context.Context, id svc.CustomerID, req *CreateReq) (*Resp, error)
	// GetExternalAccount retrieves a specific external account by ID.
	GetExternalAccount(ctx context.Context, id svc.CustomerID, externalAccountID string) (*Resp, error)
	// GetExternalAccountByIdempotencyKey retrieves an external account by its idempotency key.
	GetExternalAccountByIdempotencyKey(ctx context.Context, id svc.CustomerID, idempotencyKey string) (*Resp, error)
	// ListExternalAccounts retrieves all external accounts for a customer.
	ListExternalAccounts(ctx context.Context, id svc.CustomerID, req *ListReq) ([]Resp, error)
	// RemoveExternalAccount deletes an external bank account.
	RemoveExternalAccount(ctx context.Context, id svc.CustomerID, externalAccountID string) error
}

// IntermediaryBank represents intermediary bank details for international wire transfers.
type IntermediaryBank struct {
	// InstitutionID is the intermediary institution identifier (SWIFT code or ABA routing number).
	InstitutionID string `json:"institution_id"`
	// InstitutionName is the full legal name of the intermediary bank (optional).
	InstitutionName *string `json:"institution_name,omitempty"`
}

// CreateExternalAccount request and response types.
type (
	// CreateReq represents the request body for creating an external bank account.
	CreateReq struct {
		// IdempotencyKey is a unique key to ensure idempotent creation.
		// This is sent as a header, not in the body.
		IdempotencyKey string `json:"-"`
		// Network is the bank network type (US_ACH, SWIFT, US_FEDWIRE).
		Network BankNetworkName `json:"network"`
		// Currency is the currency of the account (USD).
		Currency Currency `json:"currency"`
		// CountryCode is the ISO 3166-1 alpha-3 country code where the bank account is held.
		CountryCode CountryCode `json:"country_code"`
		// AccountNumber is the bank account number or IBAN.
		AccountNumber string `json:"account_number"`
		// InstitutionID is the routing identifier (ABA routing number or SWIFT/BIC code).
		InstitutionID string `json:"institution_id"`
		// InstitutionName is the full legal name of the bank.
		InstitutionName string `json:"institution_name"`
		// Nickname is a user-defined label for the account (optional).
		Nickname *string `json:"nickname,omitempty"`
		// InstitutionClearingCode is additional local routing code (optional).
		InstitutionClearingCode *string `json:"institution_clearing_code,omitempty"`
		// IntermediaryBank contains intermediary bank details for international transfers (optional).
		IntermediaryBank *IntermediaryBank `json:"intermediary_bank,omitempty"`
	}

	// Resp represents the response data for an external bank account.
	Resp struct {
		// ExternalAccountID is the unique identifier for the external account.
		ExternalAccountID string `json:"external_account_id"`
		// IdempotencyKey is the idempotency key associated with the account creation.
		IdempotencyKey string `json:"idempotency_key"`
		// CustomerID is the ID of the customer who owns this account.
		CustomerID string `json:"customer_id"`
		// Status is the current status of the external account.
		Status string `json:"status"`
		// Network is the bank network type.
		Network string `json:"network"`
		// Nickname is a user-defined label for the account (optional).
		Nickname *string `json:"nickname,omitempty"`
		// AccountHolderName is the full legal name of the account holder.
		AccountHolderName string `json:"account_holder_name"`
		// Currency is the currency of the account.
		Currency string `json:"currency"`
		// CountryCode is the ISO 3166-1 alpha-3 country code.
		CountryCode string `json:"country_code"`
		// AccountNumber is the bank account number.
		AccountNumber string `json:"account_number"`
		// InstitutionID is the routing identifier (ABA or SWIFT/BIC).
		InstitutionID string `json:"institution_id"`
		// InstitutionName is the full legal name of the bank.
		InstitutionName string `json:"institution_name"`
		// InstitutionClearingCode is additional local routing code (optional).
		InstitutionClearingCode *string `json:"institution_clearing_code,omitempty"`
		// IntermediaryBank contains intermediary bank details (optional).
		IntermediaryBank *IntermediaryBank `json:"intermediary_bank,omitempty"`
		// ReferenceCode is a reference code for wire transfers (optional).
		ReferenceCode *string `json:"reference_code,omitempty"`
		// CreatedAt is the timestamp when the account was created (ISO 8601 format).
		CreatedAt string `json:"created_at"`
		// ModifiedAt is the timestamp when the account was last modified (ISO 8601 format).
		ModifiedAt string `json:"modified_at"`
	}
)

// ListReq represents optional query parameters for listing external accounts.
type ListReq struct {
	// Currency filters by currency code (e.g., USD).
	Currency Currency `json:"currency,omitempty"`
	// Network filters by bank network type (US_ACH, SWIFT, US_FEDWIRE).
	Network BankNetworkName `json:"network,omitempty"`
}

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new external accounts service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateExternalAccount creates a new external bank account for a customer.
func (s *serviceImpl) CreateExternalAccount(
	ctx context.Context,
	id svc.CustomerID,
	req *CreateReq,
) (*Resp, error) {
	path := fmt.Sprintf("/v1/customers/%s/external-accounts", id)

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

	var result Resp
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetExternalAccount retrieves a specific external account by ID.
func (s *serviceImpl) GetExternalAccount(
	ctx context.Context,
	id svc.CustomerID,
	externalAccountID string,
) (*Resp, error) {
	path := fmt.Sprintf("/v1/customers/%s/external-accounts/%s", id, externalAccountID)
	return svc.GetJSON[Resp](ctx, s.BaseService, path)
}

// GetExternalAccountByIdempotencyKey retrieves an external account by its idempotency key.
func (s *serviceImpl) GetExternalAccountByIdempotencyKey(
	ctx context.Context,
	id svc.CustomerID,
	idempotencyKey string,
) (*Resp, error) {
	path := fmt.Sprintf("/v1/customers/%s/external-accounts", id)
	params := map[string]string{
		"idempotency_key": idempotencyKey,
	}
	return svc.GetJSONWithParams[Resp](ctx, s.BaseService, path, params)
}

// ListExternalAccounts retrieves all external accounts for a customer.
func (s *serviceImpl) ListExternalAccounts(
	ctx context.Context,
	id svc.CustomerID,
	req *ListReq,
) ([]Resp, error) {
	path := fmt.Sprintf("/v1/customers/%s/external-accounts/list", id)

	params := make(map[string]string)
	if req != nil {
		if req.Currency != "" {
			params["currency"] = string(req.Currency)
		}
		if req.Network != "" {
			params["network"] = string(req.Network)
		}
	}

	result, err := svc.GetJSONWithParams[[]Resp](ctx, s.BaseService, path, params)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// RemoveExternalAccount deletes an external bank account.
func (s *serviceImpl) RemoveExternalAccount(
	ctx context.Context,
	id svc.CustomerID,
	externalAccountID string,
) error {
	path := fmt.Sprintf("/v1/customers/%s/external-accounts/%s", id, externalAccountID)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}
