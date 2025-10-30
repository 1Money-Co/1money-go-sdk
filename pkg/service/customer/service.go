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

// Package customer provides customer account management and KYB (Know Your Business) services.
//
// This package implements the customer service client for the 1Money platform, enabling
// creation and management of business customer accounts with comprehensive KYB compliance.
//
// # Overview
//
// The customer service handles:
//   - Business customer account creation with full KYB information
//   - Beneficial ownership tracking and verification
//   - Business address management (registered and physical)
//   - Risk assessment and compliance documentation
//   - Associated person management (owners, directors, signers)
//
// # Basic Usage
//
// Create a customer service instance and use it to create and manage customer accounts:
//
//	import (
//	    "context"
//	    "github.com/1Money-Co/1money-go-sdk/internal/transport"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
//	)
//
//	// Initialize transport with credentials
//	t := transport.New(accessKey, secretKey)
//
//	// Create customer service
//	svc := customer.NewService(t)
//
//	// Prepare customer creation request
//	req := &customer.CreateCustomerRequest{
//	    BusinessLegalName:          "Acme Corporation",
//	    BusinessDescription:        "Technology services provider",
//	    BusinessRegistrationNumber: "1234567890",
//	    Email:                      "contact@acme.com",
//	    BusinessType:               "corporation",
//	    BusinessIndustry:           "technology",
//	    RegisteredAddress: &customer.Address{
//	        StreetLine1: "123 Main Street",
//	        City:        "San Francisco",
//	        State:       "CA",
//	        Country:     "USA",
//	        PostalCode:  "94102",
//	    },
//	    DateOfIncorporation: "2020-01-15",
//	    // ... additional fields
//	}
//
//	// Create the customer
//	resp, err := svc.CreateCustomer(context.Background(), req)
//	if err != nil {
//	    // Handle error
//	}
//
//	// List customers with pagination
//	listReq := &customer.ListCustomersRequest{
//	    Page:     1,
//	    PageSize: 10,
//	    Status:   "active", // Optional filter
//	}
//	listResp, err := svc.ListCustomers(context.Background(), listReq)
//	if err != nil {
//	    // Handle error
//	}
//	// Process customer list
//	for _, customer := range listResp.Data {
//	    fmt.Printf("Customer: %s (%s)\n", customer.BusinessLegalName, customer.Email)
//	}
//
// # KYB Compliance
//
// The customer creation process requires comprehensive KYB information including:
//   - Business registration and incorporation details
//   - Beneficial ownership information (persons with >25% ownership)
//   - Ultimate beneficial owners (UBOs)
//   - Business purpose and source of funds/wealth
//   - Risk assessment data (high-risk activities, money services)
//   - Supporting documentation (certificates, agreements, IDs)
//
// # Associated Persons
//
// All business customers must provide information about associated persons who have:
//   - Ownership stake (typically >25%)
//   - Control over business operations
//   - Signing authority
//   - Director or officer positions
//
// Each associated person requires:
//   - Full legal name and contact information
//   - Date of birth and nationality
//   - Residential address
//   - Government-issued ID documents
//   - Tax identification information
//
// # Data Privacy
//
// This package handles sensitive personal and business information. All data should be:
//   - Transmitted over secure HTTPS connections
//   - Encrypted at rest when stored
//   - Handled in compliance with applicable privacy regulations (GDPR, CCPA, etc.)
//   - Accessed only by authorized personnel
//
// # Error Handling
//
// All service methods return errors that can be inspected for specific failure conditions.
// Common error scenarios include:
//   - Invalid or incomplete KYB information
//   - Duplicate business registration numbers
//   - Failed document verification
//   - Network or authentication errors
//
// Always check returned errors and handle them appropriately.
package customer

import (
	"context"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

const ROUTE_PREFIX = "/openapi/v1/customers"

// Service defines the customer service interface for managing customer accounts.
type Service interface {
	// CreateCustomer creates a new business customer account with KYB information.
	CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error)
	// ListCustomers retrieves a list of customer accounts with pagination support.
	ListCustomers(ctx context.Context, req *ListCustomersRequest) (*ListCustomersResponse, error)
	// GetCustomer retrieves a specific customer by ID.
	GetCustomer(ctx context.Context, customerID string) (*CustomerResponse, error)
	// UpdateCustomer updates an existing business customer account with partial KYB information.
	UpdateCustomer(ctx context.Context, customerID string, req *UpdateCustomerRequest) (*UpdateCustomerResponse, error)
	// CreateAssociatedPerson creates a new associated person (beneficial owner, controller, signer) for a customer.
	CreateAssociatedPerson(ctx context.Context, customerID string, req *CreateAssociatedPersonRequest) (*AssociatedPersonResponse, error)
	// ListAssociatedPersons retrieves all associated persons for a specific customer.
	ListAssociatedPersons(ctx context.Context, customerID string) (*ListAssociatedPersonsResponse, error)
	// GetAssociatedPerson retrieves a specific associated person by ID.
	GetAssociatedPerson(
		ctx context.Context,
		customerID string,
		associatedPersonID string,
	) (*AssociatedPersonResponse, error)
	// UpdateAssociatedPerson updates an existing associated person with partial data.
	UpdateAssociatedPerson(
		ctx context.Context,
		customerID string,
		associatedPersonID string,
		req *UpdateAssociatedPersonRequest,
	) (*AssociatedPersonResponse, error)
	// DeleteAssociatedPerson soft-deletes a specific associated person.
	DeleteAssociatedPerson(
		ctx context.Context,
		customerID string,
		associatedPersonID string,
	) error
}

// Address represents a physical or registered address for business or personal use.
// This structure is used for both registered business addresses and residential addresses.
type Address struct {
	// StreetLine1 is the primary street address (e.g., "123 Main Street").
	StreetLine1 string `json:"street_line_1"`
	// City is the city name (e.g., "San Francisco").
	City string `json:"city"`
	// Country is the country code or name (e.g., "USA").
	Country string `json:"country"`
	// State is the state or province (e.g., "CA").
	State string `json:"state"`
	// StreetLine2 is the secondary address line, such as suite or apartment number (optional).
	StreetLine2 string `json:"street_line_2,omitempty"`
	// Subdivision is the administrative subdivision within the country (optional).
	Subdivision string `json:"subdivision,omitempty"`
	// PostalCode is the postal or ZIP code.
	PostalCode string `json:"postal_code"`
}

// IdentifyingInformation represents identification documents for associated persons.
// This includes government-issued IDs such as driver's licenses, passports, or national ID cards.
type IdentifyingInformation struct {
	// Type is the type of identification document (e.g., "drivers_license", "passport").
	Type IDType `json:"type"`
	// IssuingCountry is the country that issued the identification document.
	IssuingCountry string `json:"issuing_country"`
	// ImageFront is the front image of the ID document in data-uri format (e.g., "data:image/jpeg;base64,/9j/4AAQ...").
	// Supported formats: jpeg, jpg, png, heic, tif.
	ImageFront string `json:"image_front"`
	// ImageBack is the back image of the ID document in data-uri format (optional for some ID types).
	// Supported formats: jpeg, jpg, png, heic, tif.
	ImageBack string `json:"image_back,omitempty"`
}

// AssociatedPerson represents a person associated with the business entity.
// This includes beneficial owners, directors, signers, and persons with control over the business.
// All associated persons must provide complete KYC information for compliance purposes.
type AssociatedPerson struct {
	// FirstName is the person's legal first name.
	FirstName string `json:"first_name"`
	// MiddleName is the person's middle name (optional).
	MiddleName string `json:"middle_name,omitempty"`
	// LastName is the person's legal last name.
	LastName string `json:"last_name"`
	// Email is the person's contact email address.
	Email string `json:"email"`
	// ResidentialAddress is the person's current residential address.
	ResidentialAddress *Address `json:"residential_address"`
	// BirthDate is the person's date of birth in ISO format (e.g., "1980-01-15").
	BirthDate string `json:"birth_date"`
	// CountryOfBirth is the country where the person was born.
	CountryOfBirth string `json:"country_of_birth"`
	// Gender is the person's gender (e.g., "M", "F").
	Gender Gender `json:"gender"`
	// PrimaryNationality is the person's primary nationality or citizenship.
	PrimaryNationality string `json:"primary_nationality"`
	// HasOwnership indicates whether the person has ownership stake in the business.
	HasOwnership bool `json:"has_ownership"`
	// OwnershipPercentage is the percentage of ownership (required if HasOwnership is true).
	OwnershipPercentage int `json:"ownership_percentage,omitempty"`
	// HasControl indicates whether the person has control over the business operations.
	HasControl bool `json:"has_control"`
	// IsSigner indicates whether the person is authorized to sign on behalf of the business.
	IsSigner bool `json:"is_signer"`
	// IsDirector indicates whether the person serves as a director of the business.
	IsDirector bool `json:"is_director"`
	// IdentifyingInformation is a list of identification documents for this person.
	IdentifyingInformation []IdentifyingInformation `json:"identifying_information,omitempty"`
	// DualNationality is the person's second nationality if they hold dual citizenship (optional).
	DualNationality string `json:"dual_nationality,omitempty"`
	// CountryOfTax is the country where the person is subject to taxation.
	CountryOfTax string `json:"country_of_tax"`
	// TaxType is the type of tax identification (e.g., "SSN", "EIN").
	TaxType TaxIDType `json:"tax_type"`
	// TaxIDNumber is the person's tax identification number.
	TaxIDNumber string `json:"tax_id_number"`
	// POA is the Power of Attorney document in data-uri format (optional).
	// Format: "data:image/[type];base64,[base64_data]" where type is jpeg, jpg, png, heic, or tif.
	POA string `json:"poa,omitempty"`
}

// Document represents a business document attachment for KYB verification.
// Documents may include certificates of incorporation, operating agreements, or other legal documents.
type Document struct {
	// DocType is the type of document (e.g., "certificate_of_incorporation" for Certificate of Incorporation).
	DocType DocumentType `json:"doc_type"`
	// File is the document file in data-uri format.
	// Format: "data:image/[type];base64,[base64_data]" where type is jpeg, jpg, png, heic, or tif.
	File string `json:"file"`
	// Description is an optional description of the document.
	Description string `json:"description,omitempty"`
}

// ValidationError represents a validation error detail in response.
// This is returned when the API processes a request but finds validation issues.
type ValidationError struct {
	// ErrorType is the type of error (e.g., "missing_field", "invalid_value").
	ErrorType string `json:"error_type"`
	// Location is the location of the error (field path or index).
	Location string `json:"location"`
	// Message is the detailed error message.
	Message string `json:"message"`
}

// CreateCustomerRequest represents the request body for creating a business customer.
// This request includes comprehensive KYB (Know Your Business) information required for
// regulatory compliance, including business details, beneficial ownership, and risk assessment.
//
// All monetary ranges use string enumerations (e.g., "0_99999", "100000_499999").
// Business types include: "cooperative", "corporation", "llc", "partnership", "sole_proprietorship".
// Industries follow standard classification codes for financial services, retail, technology, etc.
type CreateCustomerRequest struct {
	// BusinessLegalName is the official registered legal name of the business entity.
	BusinessLegalName string `json:"business_legal_name"`
	// BusinessDescription provides a detailed description of the business operations and activities.
	BusinessDescription string `json:"business_description"`
	// BusinessRegistrationNumber is the official business registration or incorporation number.
	BusinessRegistrationNumber string `json:"business_registration_number"`
	// Email is the primary contact email address for the business.
	Email string `json:"email"`
	// BusinessType specifies the legal structure (e.g., "cooperative", "corporation", "llc").
	BusinessType BusinessType `json:"business_type"`
	// BusinessIndustry specifies the industry classification (e.g., "bank_credit_unions_regulated_financial_institution").
	BusinessIndustry BusinessIndustry `json:"business_industry"`
	// RegisteredAddress is the official registered address of the business.
	RegisteredAddress *Address `json:"registered_address"`
	// DateOfIncorporation is the date when the business was officially incorporated (ISO format).
	DateOfIncorporation string `json:"date_of_incorporation"`
	// PhysicalAddress is the actual operating address if different from registered address (optional).
	PhysicalAddress *Address `json:"physical_address,omitempty"`
	// SignedAgreementID is the identifier of the signed service agreement.
	SignedAgreementID string `json:"signed_agreement_id"`
	// IsDAO indicates whether this is a Decentralized Autonomous Organization.
	IsDAO bool `json:"is_dao"`
	// AssociatedPersons is a list of all persons associated with the business (owners, directors, signers).
	// At least one associated person with ownership or control is typically required.
	AssociatedPersons []AssociatedPerson `json:"associated_persons"`
	// AccountPurpose describes the primary purpose of the account (e.g., "charitable_donations").
	AccountPurpose AccountPurpose `json:"account_purpose"`
	// SourceOfFunds is a list of sources for the funds being used (e.g., ["business_loans"]).
	SourceOfFunds []SourceOfFunds `json:"source_of_funds"`
	// SourceOfWealth is a list of sources for the business's wealth (e.g., ["business_dividends_or_profits"]).
	SourceOfWealth []SourceOfWealth `json:"source_of_wealth"`
	// Documents is a list of supporting documents for KYB verification (optional).
	Documents []Document `json:"documents,omitempty"`
	// PrimaryWebsite is the business's primary website URL (optional).
	PrimaryWebsite string `json:"primary_website,omitempty"`
	// PubliclyTraded indicates whether the business is publicly traded on a stock exchange.
	PubliclyTraded bool `json:"publicly_traded"`
	// EstimatedAnnualRevenueUSD is the estimated annual revenue range (e.g., "0_99999").
	EstimatedAnnualRevenueUSD MoneyRange `json:"estimated_annual_revenue_usd"`
	// ExpectedMonthlyFiatDeposits is the expected monthly fiat deposit range (e.g., "0_99999").
	ExpectedMonthlyFiatDeposits MoneyRange `json:"expected_monthly_fiat_deposits"`
	// ExpectedMonthlyFiatWithdrawals is the expected monthly fiat withdrawal range (e.g., "0_99999").
	ExpectedMonthlyFiatWithdrawals MoneyRange `json:"expected_monthly_fiat_withdrawals"`
	// AccountPurposeOther provides additional details if AccountPurpose is "other" (optional).
	AccountPurposeOther string `json:"account_purpose_other,omitempty"`
	// HighRiskActivities is a list of high-risk business activities (e.g., ["adult_entertainment"]).
	HighRiskActivities []HighRiskActivity `json:"high_risk_activities,omitempty"`
	// HighRiskActivitiesExplanation provides additional context for high-risk activities (optional).
	HighRiskActivitiesExplanation string `json:"high_risk_activities_explanation,omitempty"`
	// ConductsMoneyServices indicates whether the business conducts money service business activities.
	ConductsMoneyServices bool `json:"conducts_money_services"`
	// TaxID is the business tax identification number.
	TaxID string `json:"tax_id"`
	// TaxType is the type of tax ID (e.g., "EIN", "TIN").
	TaxType TaxIDType `json:"tax_type"`
}

// CustomerResponse represents the standard customer response data.
// This structure is used for customer creation, retrieval, and update operations.
type CustomerResponse struct {
	// ID is the unique identifier of the customer.
	ID string `json:"id"`
	// Email is the primary contact email for the customer.
	Email string `json:"email"`
	// BusinessLegalName is the legal business name.
	BusinessLegalName string `json:"business_legal_name"`
	// BusinessDescription provides a detailed description of the business operations and activities.
	BusinessDescription string `json:"business_description,omitempty"`
	// BusinessType is the type of business entity.
	BusinessType BusinessType `json:"business_type"`
	// BusinessIndustry specifies the industry classification.
	BusinessIndustry BusinessIndustry `json:"business_industry,omitempty"`
	// BusinessRegistrationNumber is the official business registration or incorporation number.
	BusinessRegistrationNumber string `json:"business_registration_number,omitempty"`
	// DateOfIncorporation is the date when the business was officially incorporated (ISO format).
	DateOfIncorporation string `json:"date_of_incorporation,omitempty"`
	// IncorporationCountry is the country where the business was incorporated.
	IncorporationCountry string `json:"incorporation_country,omitempty"`
	// IncorporationState is the state or province where the business was incorporated.
	IncorporationState string `json:"incorporation_state,omitempty"`
	// RegisteredAddress is the official registered address of the business.
	RegisteredAddress *Address `json:"registered_address,omitempty"`
	// PhysicalAddress is the actual operating address if different from registered address.
	PhysicalAddress *Address `json:"physical_address,omitempty"`
	// PrimaryWebsite is the business's primary website URL.
	PrimaryWebsite string `json:"primary_website,omitempty"`
	// PubliclyTraded indicates whether the business is publicly traded on a stock exchange.
	PubliclyTraded bool `json:"publicly_traded,omitempty"`
	// TaxID is the business tax identification number.
	TaxID string `json:"tax_id,omitempty"`
	// TaxType is the type of tax ID (e.g., "EIN", "TIN").
	TaxType TaxIDType `json:"tax_type,omitempty"`
	// TaxCountry is the country where the business is subject to taxation.
	TaxCountry string `json:"tax_country,omitempty"`
	// Status is the current customer account status.
	Status CustomerStatus `json:"status"`
	// RiskScore is the calculated risk score for the customer (0-100).
	RiskScore float64 `json:"risk_score,omitempty"`
	// SubmittedAt is the timestamp when the customer application was submitted (ISO 8601 format).
	SubmittedAt string `json:"submitted_at,omitempty"`
	// CreatedAt is the timestamp when the customer account was created (ISO 8601 format).
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the customer account was last updated (ISO 8601 format).
	UpdatedAt string `json:"updated_at"`
	// ValidationErrors contains validation errors if any were found during processing.
	// This field is present when the request was processed but validation issues were found.
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

// CreateCustomerResponse is an alias for CustomerResponse.
// This type is used for clarity in the CreateCustomer method signature.
type CreateCustomerResponse = CustomerResponse

// UpdateCustomerRequest represents the request body for updating an existing business customer.
// This request supports partial updates - only the fields that are provided will be updated.
// All fields are optional (using pointers for primitive types and omitempty for complex types).
//
// The structure follows the generic update pattern where each field can be independently updated
// without affecting other fields. Fields set to nil or omitted will not modify the existing values.
type UpdateCustomerRequest struct {
	// BusinessLegalName is the official registered legal name of the business entity.
	BusinessLegalName *string `json:"business_legal_name,omitempty"`
	// BusinessDescription provides a detailed description of the business operations and activities.
	BusinessDescription *string `json:"business_description,omitempty"`
	// BusinessRegistrationNumber is the official business registration or incorporation number.
	BusinessRegistrationNumber *string `json:"business_registration_number,omitempty"`
	// Email is the primary contact email address for the business.
	Email *string `json:"email,omitempty"`
	// BusinessType specifies the legal structure (e.g., "cooperative", "corporation", "llc").
	BusinessType *BusinessType `json:"business_type,omitempty"`
	// BusinessIndustry specifies the industry classification.
	BusinessIndustry *BusinessIndustry `json:"business_industry,omitempty"`
	// RegisteredAddress is the official registered address of the business.
	RegisteredAddress *Address `json:"registered_address,omitempty"`
	// DateOfIncorporation is the date when the business was officially incorporated (ISO format).
	DateOfIncorporation *string `json:"date_of_incorporation,omitempty"`
	// PhysicalAddress is the actual operating address if different from registered address.
	PhysicalAddress *Address `json:"physical_address,omitempty"`
	// SignedAgreementID is the identifier of the signed service agreement.
	SignedAgreementID *string `json:"signed_agreement_id,omitempty"`
	// IsDAO indicates whether this is a Decentralized Autonomous Organization.
	IsDAO *bool `json:"is_dao,omitempty"`
	// AssociatedPersons is a list of all persons associated with the business (owners, directors, signers).
	// This can be used to add new associated persons to the customer account.
	AssociatedPersons []AssociatedPerson `json:"associated_persons,omitempty"`
	// AccountPurpose describes the primary purpose of the account.
	AccountPurpose *AccountPurpose `json:"account_purpose,omitempty"`
	// SourceOfFunds is a list of sources for the funds being used.
	SourceOfFunds []SourceOfFunds `json:"source_of_funds,omitempty"`
	// SourceOfWealth is a list of sources for the business's wealth.
	SourceOfWealth []SourceOfWealth `json:"source_of_wealth,omitempty"`
	// Documents is a list of supporting documents for KYB verification.
	Documents []Document `json:"documents,omitempty"`
	// PrimaryWebsite is the business's primary website URL.
	PrimaryWebsite *string `json:"primary_website,omitempty"`
	// PubliclyTraded indicates whether the business is publicly traded on a stock exchange.
	PubliclyTraded *bool `json:"publicly_traded,omitempty"`
	// EstimatedAnnualRevenueUSD is the estimated annual revenue range.
	EstimatedAnnualRevenueUSD *MoneyRange `json:"estimated_annual_revenue_usd,omitempty"`
	// ExpectedMonthlyFiatDeposits is the expected monthly fiat deposit range.
	ExpectedMonthlyFiatDeposits *MoneyRange `json:"expected_monthly_fiat_deposits,omitempty"`
	// ExpectedMonthlyFiatWithdrawals is the expected monthly fiat withdrawal range.
	ExpectedMonthlyFiatWithdrawals *MoneyRange `json:"expected_monthly_fiat_withdrawals,omitempty"`
	// AccountPurposeOther provides additional details if AccountPurpose is "other".
	AccountPurposeOther *string `json:"account_purpose_other,omitempty"`
	// HighRiskActivities is a list of high-risk business activities.
	HighRiskActivities []HighRiskActivity `json:"high_risk_activities,omitempty"`
	// HighRiskActivitiesExplanation provides additional context for high-risk activities.
	HighRiskActivitiesExplanation *string `json:"high_risk_activities_explanation,omitempty"`
	// ConductsMoneyServices indicates whether the business conducts money service business activities.
	ConductsMoneyServices *bool `json:"conducts_money_services,omitempty"`
	// TaxID is the business tax identification number.
	TaxID *string `json:"tax_id,omitempty"`
	// TaxType is the type of tax ID (e.g., "EIN", "TIN").
	TaxType *TaxIDType `json:"tax_type,omitempty"`
}

// UpdateCustomerResponse is an alias for CustomerResponse.
// This type is used for clarity in the UpdateCustomer method signature.
type UpdateCustomerResponse = CustomerResponse

// ListCustomersRequest represents the request parameters for listing customers.
// This supports pagination and filtering of customer accounts.
type ListCustomersRequest struct {
	// Page is the page number for pagination (1-indexed).
	Page int `json:"page,omitempty"`
	// PageSize is the number of items per page.
	PageSize int `json:"page_size,omitempty"`
	// Status filters customers by their account status (e.g., "active", "pending", "suspended").
	Status string `json:"status,omitempty"`
	// Email filters customers by email address.
	Email string `json:"email,omitempty"`
	// Name filters customers by business name (partial match supported).
	Name string `json:"name,omitempty"`
}

// CustomerSummary represents a summary of a customer account in list responses.
// This contains a subset of customer information for efficient listing.
type CustomerSummary struct {
	// ID is the unique identifier of the customer.
	ID string `json:"id"`
	// Email is the primary contact email.
	Email string `json:"email"`
	// BusinessLegalName is the legal business name.
	BusinessLegalName string `json:"business_legal_name"`
	// BusinessType is the type of business entity.
	BusinessType BusinessType `json:"business_type"`
	// Status is the current account status.
	Status CustomerStatus `json:"status"`
	// CreatedAt is the timestamp when the customer was created (ISO 8601 format).
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the customer was last updated (ISO 8601 format).
	UpdatedAt string `json:"updated_at"`
}

// ListCustomersResponse represents the response data for listing customers.
// This includes the customer list and pagination information.
// Note: The API returns only the customer array. Pagination fields (Total, Page, PageSize, TotalPages)
// are computed on the client side based on the returned data and request parameters.
type ListCustomersResponse struct {
	// Data is the list of customer summaries returned by the API.
	Data []CustomerSummary `json:"data"`
	// Total is the number of customers in the current response (computed as len(Data)).
	Total int `json:"total"`
}

// CreateAssociatedPersonRequest represents the request body for creating an associated person.
// This wraps the AssociatedPerson structure for the creation endpoint.
type CreateAssociatedPersonRequest struct {
	AssociatedPerson
}

// AssociatedPersonResponse represents the response data for an associated person.
// This is returned after creating or retrieving associated persons (beneficial owners, controllers, signers).
type AssociatedPersonResponse struct {
	// ID is the unique identifier for this associated person.
	ID string `json:"id"`
	// Email is the person's contact email address.
	Email string `json:"email"`
	// FirstName is the person's legal first name.
	FirstName string `json:"first_name"`
	// MiddleName is the person's middle name (optional).
	MiddleName string `json:"middle_name,omitempty"`
	// LastName is the person's legal last name.
	LastName string `json:"last_name"`
	// BirthDate is the person's date of birth in ISO format (e.g., "1980-01-15").
	BirthDate string `json:"birth_date"`
	// PrimaryNationality is the person's primary nationality or citizenship (ISO 3166-1 alpha-2).
	PrimaryNationality string `json:"primary_nationality"`
	// ResidentialAddress is the person's current residential address.
	ResidentialAddress *Address `json:"residential_address"`
	// ApplicantType is the type based on role (e.g., "UltimateBeneficialOwner").
	ApplicantType string `json:"applicant_type"`
	// Title is the role titles (pipe-delimited IDs, e.g., "1|2").
	Title string `json:"title"`
	// HasOwnership indicates whether the person has ownership stake (≥25%).
	HasOwnership bool `json:"has_ownership"`
	// OwnershipPercentage is the percentage of ownership (0.01-100).
	OwnershipPercentage float64 `json:"ownership_percentage,omitempty"`
	// HasControl indicates whether the person has control over the business (CEO, CFO, etc.).
	HasControl bool `json:"has_control"`
	// IsSigner indicates whether the person is an authorized signer.
	IsSigner bool `json:"is_signer"`
	// IsDirector indicates whether the person serves as a director.
	IsDirector bool `json:"is_director"`
	// CreatedAt is the timestamp when the associated person was created (ISO 8601 format).
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the associated person was last updated (ISO 8601 format).
	UpdatedAt string `json:"updated_at"`
}

// UpdateAssociatedPersonRequest represents the request body for updating an associated person.
// This request supports partial updates - only the fields that are provided will be updated.
// All fields are optional (using pointers for primitive types and omitempty for complex types).
type UpdateAssociatedPersonRequest struct {
	// FirstName is the person's legal first name.
	FirstName *string `json:"first_name,omitempty"`
	// MiddleName is the person's middle name.
	MiddleName *string `json:"middle_name,omitempty"`
	// LastName is the person's legal last name.
	LastName *string `json:"last_name,omitempty"`
	// Email is the person's contact email address.
	Email *string `json:"email,omitempty"`
	// ResidentialAddress is the person's current residential address.
	ResidentialAddress *Address `json:"residential_address,omitempty"`
	// BirthDate is the person's date of birth in ISO format (e.g., "1980-01-15").
	BirthDate *string `json:"birth_date,omitempty"`
	// CountryOfBirth is the country where the person was born.
	CountryOfBirth *string `json:"country_of_birth,omitempty"`
	// Gender is the person's gender.
	Gender *Gender `json:"gender,omitempty"`
	// PrimaryNationality is the person's primary nationality or citizenship.
	PrimaryNationality *string `json:"primary_nationality,omitempty"`
	// HasOwnership indicates whether the person has ownership stake in the business.
	HasOwnership *bool `json:"has_ownership,omitempty"`
	// OwnershipPercentage is the percentage of ownership.
	OwnershipPercentage *int `json:"ownership_percentage,omitempty"`
	// HasControl indicates whether the person has control over the business operations.
	HasControl *bool `json:"has_control,omitempty"`
	// IsSigner indicates whether the person is authorized to sign on behalf of the business.
	IsSigner *bool `json:"is_signer,omitempty"`
	// IsDirector indicates whether the person serves as a director of the business.
	IsDirector *bool `json:"is_director,omitempty"`
	// IdentifyingInformation is a list of identification documents for this person.
	IdentifyingInformation []IdentifyingInformation `json:"identifying_information,omitempty"`
	// DualNationality is the person's second nationality if they hold dual citizenship.
	DualNationality *string `json:"dual_nationality,omitempty"`
	// CountryOfTax is the country where the person is subject to taxation.
	CountryOfTax *string `json:"country_of_tax,omitempty"`
	// TaxType is the type of tax identification.
	TaxType *TaxIDType `json:"tax_type,omitempty"`
	// TaxIDNumber is the person's tax identification number.
	TaxIDNumber *string `json:"tax_id_number,omitempty"`
	// POA is the Power of Attorney document in data-uri format.
	POA *string `json:"poa,omitempty"`
}

// ListAssociatedPersonsResponse represents the response data for listing associated persons.
// This contains the list of all associated persons for a specific customer.
type ListAssociatedPersonsResponse []AssociatedPersonResponse

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new customer service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateCustomer creates a new customer using the generic PostJSON function.
func (s *serviceImpl) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error) {
	resp, err := svc.PostJSON[*CreateCustomerRequest, CreateCustomerResponse](
		ctx,
		s.BaseService,
		ROUTE_PREFIX,
		req,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// ListCustomers retrieves a list of customers with optional filtering and pagination.
func (s *serviceImpl) ListCustomers(ctx context.Context, req *ListCustomersRequest) (*ListCustomersResponse, error) {
	params := make(map[string]string)

	if req != nil {
		if req.Page > 0 {
			params["page"] = fmt.Sprintf("%d", req.Page)
		}
		if req.PageSize > 0 {
			params["page_size"] = fmt.Sprintf("%d", req.PageSize)
		}
		if req.Status != "" {
			params["status"] = req.Status
		}
		if req.Email != "" {
			params["email"] = req.Email
		}
		if req.Name != "" {
			params["name"] = req.Name
		}
	}

	resp, err := svc.GetJSONWithParams[ListCustomersResponse](
		ctx,
		s.BaseService,
		ROUTE_PREFIX,
		params,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetCustomer retrieves a specific customer by ID.
func (s *serviceImpl) GetCustomer(ctx context.Context, customerID string) (*CustomerResponse, error) {
	path := fmt.Sprintf("%s/%s", ROUTE_PREFIX, customerID)
	resp, err := svc.GetJSON[CustomerResponse](ctx, s.BaseService, path)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// UpdateCustomer updates an existing customer with partial KYB information.
// Only the fields provided in the request will be updated; nil/omitted fields remain unchanged.
func (s *serviceImpl) UpdateCustomer(ctx context.Context, customerID string, req *UpdateCustomerRequest) (*UpdateCustomerResponse, error) {
	path := fmt.Sprintf("%s/%s", ROUTE_PREFIX, customerID)
	resp, err := svc.PatchJSON[*UpdateCustomerRequest, UpdateCustomerResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// CreateAssociatedPerson creates a new associated person for a customer.
func (s *serviceImpl) CreateAssociatedPerson(
	ctx context.Context,
	customerID string,
	req *CreateAssociatedPersonRequest,
) (*AssociatedPersonResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons", ROUTE_PREFIX, customerID)
	resp, err := svc.PostJSON[*CreateAssociatedPersonRequest, AssociatedPersonResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// ListAssociatedPersons retrieves all associated persons for a specific customer.
func (s *serviceImpl) ListAssociatedPersons(ctx context.Context, customerID string) (*ListAssociatedPersonsResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons", ROUTE_PREFIX, customerID)
	resp, err := svc.GetJSON[ListAssociatedPersonsResponse](ctx, s.BaseService, path)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// GetAssociatedPerson retrieves a specific associated person by ID.
func (s *serviceImpl) GetAssociatedPerson(
	ctx context.Context,
	customerID string,
	associatedPersonID string,
) (*AssociatedPersonResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons/%s", ROUTE_PREFIX, customerID, associatedPersonID)
	resp, err := svc.GetJSON[AssociatedPersonResponse](ctx, s.BaseService, path)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// UpdateAssociatedPerson updates an existing associated person with partial data.
// Only the fields provided in the request will be updated; nil/omitted fields remain unchanged.
func (s *serviceImpl) UpdateAssociatedPerson(
	ctx context.Context,
	customerID string,
	associatedPersonID string,
	req *UpdateAssociatedPersonRequest,
) (*AssociatedPersonResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons/%s", ROUTE_PREFIX, customerID, associatedPersonID)
	resp, err := svc.PatchJSON[*UpdateAssociatedPersonRequest, AssociatedPersonResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// DeleteAssociatedPerson soft-deletes a specific associated person.
func (s *serviceImpl) DeleteAssociatedPerson(
	ctx context.Context,
	customerID string,
	associatedPersonID string,
) error {
	path := fmt.Sprintf("%s/%s/associated_persons/%s", ROUTE_PREFIX, customerID, associatedPersonID)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}
