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

package customer

//go:generate go tool go-enum -f=$GOFILE --marshal --names --nocase

// BusinessType represents the legal structure of a business entity.
// ENUM(cooperative, corporation, llc, partnership, sole_proprietorship)
type BusinessType string

// BusinessIndustry represents the NAICS (North American Industry Classification System) code
// representing the business industry. This is a string field that accepts NAICS codes
// (e.g., "541519" for Other Computer Related Services).
// The NAICS code will be converted to internal answer ID for database storage.
// Valid NAICS codes should be 1-10 characters in length.
type BusinessIndustry string

// IDType represents the type of identification document.
// ENUM(drivers_license, passport, national_id, state_id)
type IDType string

// AccountPurpose represents the primary purpose of the customer account.
/* ENUM(
charitable_donations
ecommerce_retail_payments
investment_purposes
other
payments_to_friends_or_family_abroad
payroll
personal_or_living_expenses
protect_wealth
purchase_goods_and_services
receive_payments_for_goods_and_services
tax_optimization
third_party_money_transmission
treasury_management
)
*/
type AccountPurpose string

// MoneyRange represents a range of monetary amounts in USD.
// ENUM(0_99999, 100000_499999, 500000_999999, 1000000_4999999, 5000000_plus)
type MoneyRange string

// DocumentType represents the type of business document.
/* ENUM(
aml_comfort_letter
constitutional_document
directors_registry
e_signature_certificate
evidence_of_good_standing
flow_of_funds
formation_document
marketing_materials
other
ownership_chart
ownership_information
proof_of_account_purpose
proof_of_address
proof_of_entity_name_change
proof_of_nature_of_business
proof_of_signatory_authority
proof_of_source_of_funds
proof_of_source_of_wealth
proof_of_tax_identification
registration_document
shareholder_register
)
*/
type DocumentType string

// TaxIDType represents the type of tax identification covering all supported countries.
/* ENUM(
SSN
EIN
TFN
ABN
ACN
UTR
NINO
NRIC
FIN
ASDG
ITR
NIF
TIN
VAT
CUIL
CUIT
DNI
BIN
UNP
RNPM
NIT
CPF
CNPJ
NIRE
UCN
UIC
SIN
BN
RUT
IIN
USCC
CNOC
USCN
ITIN
CPJ
OIB
DIC
CPR
CVR
CN
RNC
RUC
TN
HETU
YT
ALV
SIREN
IDNR
STNR
VTA
HKID
AJ
EN
KN
VSK
PAN
GSTN
NIK
NPWP
PPS
TRN
CRO
CHY
CF
IVA
IN
JCT
EDRPOU
EID
)
*/
type TaxIDType string

// SourceOfFunds represents the origin of funds for business operations.
// ENUM(
// business_loans,
// grants,
// inter_company_funds,
// investment_proceeds,
// legal_settlement,
// owners_capital,
// pension_retirement,
// sale_of_assets,
// sales_of_goods_and_services,
// tax_refund,
// third_party_funds,
// treasury_reserves)
type SourceOfFunds string

// SourceOfWealth represents the origin of the business's accumulated wealth.
// ENUM(
// business_dividends_or_profits,
// sale_of_business,
// inheritance,
// real_estate_investments,
// investment_returns,
// accumulated_revenue,
// other)
type SourceOfWealth string

// HighRiskActivity represents potentially high-risk business activities.
// ENUM(
// adult_entertainment,
// cannabis,
// cryptocurrency,
// gambling,
// money_services,
// precious_metals,
// weapons,
// none)
type HighRiskActivity string

// ImageFormat represents supported image formats for document uploads.
// ENUM(jpeg, jpg, png, heic, tif)
type ImageFormat string

// FileFormat represents all supported file formats for document uploads.
// This includes images, PDFs, and spreadsheet formats.
// ENUM(jpeg, jpg, png, heic, tif, pdf, csv, xls, xlsx)
type FileFormat string

// KybStatus represents the KYB (Know Your Business) verification status of a customer account.
// This status tracks the progress and state of the KYB verification process.
// ENUM(
// init,
// pending_review,
// under_review,
// pending_response,
// escalated,
// pending_approval,
// rejected,
// approved)
type KybStatus string
