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
// ENUM(CERT_OF_INC, OPERATING_AGREEMENT, BYLAWS, PARTNERSHIP_AGREEMENT, BUSINESS_LICENSE, TAX_DOCUMENT, BANK_STATEMENT, UTILITY_BILL, OTHER)
type DocumentType string

// Gender represents the gender of an individual.
// ENUM(M, F, O)
type Gender string

// TaxIDType represents the type of tax identification.
// ENUM(EIN, SSN, TIN, ITIN)
type TaxIDType string

// SourceOfFunds represents the origin of funds for business operations.
// ENUM(business_loans, grants, inter_company_funds, investment_proceeds, legal_settlement, owners_capital, pension_retirement, sale_of_assets, sales_of_goods_and_services, tax_refund, third_party_funds, treasury_reserves)
type SourceOfFunds string

// SourceOfWealth represents the origin of the business's accumulated wealth.
// ENUM(business_dividends_or_profits, sale_of_business, inheritance, real_estate_investments, investment_returns, accumulated_revenue, other)
type SourceOfWealth string

// HighRiskActivity represents potentially high-risk business activities.
// ENUM(adult_entertainment, cannabis, cryptocurrency, gambling, money_services, precious_metals, weapons, none)
type HighRiskActivity string
