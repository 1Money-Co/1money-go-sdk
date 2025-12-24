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

package service

// ID type aliases improve code readability by making the purpose of string parameters clear.
// These are type aliases (not new types), so they are fully compatible with plain strings.
type (
	// CustomerID is a type alias for customer identifiers.
	CustomerID = string

	// ExternalAccountID is a type alias for external bank account identifiers.
	ExternalAccountID = string

	// RecipientID is a type alias for recipient identifiers.
	RecipientID = string

	// WalletAddressID is a type alias for wallet address identifiers.
	WalletAddressID = string

	// TransactionID is a type alias for transaction identifiers.
	TransactionID = string

	// AssociatedPersonID is a type alias for associated person identifiers.
	AssociatedPersonID = string

	// SignedAgreementID is a type alias for signed agreement identifiers.
	SignedAgreementID = string

	// AutoConversionRuleID is a type alias for auto conversion rule identifiers.
	AutoConversionRuleID = string

	// AutoConversionOrderID is a type alias for auto conversion order identifiers.
	AutoConversionOrderID = string

	// QuoteID is a type alias for conversion quote identifiers.
	QuoteID = string

	// OrderID is a type alias for conversion order identifiers.
	OrderID = string

	// SimulationID is a type alias for simulation identifiers.
	SimulationID = string

	// IdempotencyKey is a type alias for idempotency keys used in API requests.
	IdempotencyKey = string
)
