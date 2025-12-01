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

package assets

//go:generate go tool go-enum -f=$GOFILE --marshal --names --nocase

// AssetName represents the supported asset types for filtering.
// ENUM(USD, USDC, USDT, PYUSD, RLUSD, USDG, USDP, EURC, MXNB)
type AssetName string

// NetworkName represents the supported network types.
/* ENUM(
US_ACH
SWIFT
US_FEDWIRE
ARBITRUM
AVALANCHE
BASE
BNBCHAIN
ETHEREUM
POLYGON
SOLANA
)
*/
type NetworkName string

// SortOrder represents the sort order for results.
// ENUM(ASC, DESC)
type SortOrder string
