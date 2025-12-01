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

package external_accounts

//go:generate go tool go-enum -f=$GOFILE --marshal --names --nocase

// BankNetworkName represents the bank network type for external accounts.
// ENUM(US_ACH, SWIFT, US_FEDWIRE)
type BankNetworkName string

// Currency represents the supported currencies for external accounts.
// ENUM(USD)
type Currency string

// BankAccountStatus represents the status of an external bank account.
// ENUM(PENDING, APPROVED, FAILED)
type BankAccountStatus string
