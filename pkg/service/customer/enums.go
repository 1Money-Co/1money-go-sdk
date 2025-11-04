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

// BusinessIndustry represents the industry classification of a business.
/* ENUM(
bank_credit_unions_regulated_financial_institution
professional_services
technology_e_commerce_platforms
general_manufacturing
general_wholesalers
healthcare_and_social_assistance
educational_services
scientific_and_technical_services
non_bank_financial_institution
investment_fund
real_estate
retail_trade
arts_entertainment_recreation
accommodation_food_services
other
)
*/
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
certificate_of_incorporation
certificate_of_formation_registration
certificate_of_incorporation_and_articles_of_organization
constitutional_or_formation_documents
partnership_agreement
articles_of_organization
articles_of_incorporation
operating_lp_agreement_if_applicable
prospectus_offering_memorandum_or_private_placement_memorandum
aml_attestation_letter
fund_structure_chart
articles_of_incorporation_by_laws_or_equivalent_document
irs_determination_letter
annual_reports
business_license
trade_name_registration_doing_business_as_dba_filing
tax_filings
list_manager_or_similar_persons_that_has_have_the_ability_to_legally_bind_the_dao_and_carry_out_the_daos_instructions
voting_records
trust_agreement
certificate_of_good_standing
ownership_and_formation_documents
ownership_structure_llc
ownership_structure_corp
ownership_structure_part
ownership_structure_dao
ownership_structure_gov
authorized_representative_list
proof_of_source_of_funds
proof_of_business_entity_address
proof_of_business_entity_address_dao
w9_form
state_local_money_transmission_licensing_evidence_or_equivalent_regulatory_authorization_non_us
aml_policy
certificate_of_incumbency_or_register_of_directors
tax_exemption_or_charity_registration_letter
memorandum_of_association_or_article_of_association_or_equivalent_document
supporting_documents
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

// CustomerStatus represents the current status of a customer account.
// This status represents the overall state of the customer account,
// including KYB verification progress and account operational status.
// ENUM(
// active,
// awaiting_questionnaire,
// awaiting_ubo,
// incomplete,
// not_started,
// offboarded,
// paused,
// rejected,
// under_review)
type CustomerStatus string
