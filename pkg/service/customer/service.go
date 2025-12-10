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
//	    BusinessIndustry:           "541519", // NAICS code
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

const ROUTE_PREFIX = "/v1/customers"

// Service defines the customer service interface for managing customer accounts.
type Service interface {
	// CreateTOSLink creates a session token for signing the Terms of Service agreement.
	// This is the first step in the customer onboarding flow.
	// The session expires in 1 hour.
	// Pass nil for req if no redirect URI is needed.
	CreateTOSLink(ctx context.Context, req *CreateTOSLinkRequest) (*TOSLinkResponse, error)
	// SignTOSAgreement signs the Terms of Service agreement using the session token.
	// This is the second step in the customer onboarding flow.
	// Returns a signed_agreement_id to be used in customer creation.
	SignTOSAgreement(ctx context.Context, sessionToken string) (*SignAgreementResponse, error)
	// CreateCustomer creates a new business customer account with KYB information.
	CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error)
	// ListCustomers retrieves a list of customer accounts with pagination support.
	ListCustomers(ctx context.Context, req *ListCustomersRequest) (*ListCustomersResponse, error)
	// GetCustomer retrieves a specific customer by ID.
	GetCustomer(ctx context.Context, id svc.CustomerID) (*CustomerResponse, error)
	// UpdateCustomer updates an existing business customer account with partial KYB information.
	UpdateCustomer(ctx context.Context, id svc.CustomerID, req *UpdateCustomerRequest) (*UpdateCustomerResponse, error)
	// CreateAssociatedPerson creates a new associated person (beneficial owner, controller, signer) for a customer.
	CreateAssociatedPerson(
		ctx context.Context, id svc.CustomerID, req *CreateAssociatedPersonRequest,
	) (*AssociatedPersonResponse, error)
	// ListAssociatedPersons retrieves all associated persons for a specific customer.
	ListAssociatedPersons(ctx context.Context, id svc.CustomerID) (*ListAssociatedPersonsResponse, error)
	// GetAssociatedPerson retrieves a specific associated person by ID.
	GetAssociatedPerson(
		ctx context.Context, id svc.CustomerID, associatedPersonID string,
	) (*AssociatedPersonResponse, error)
	// UpdateAssociatedPerson updates an existing associated person with partial data.
	UpdateAssociatedPerson(
		ctx context.Context, id svc.CustomerID, associatedPersonID string, req *UpdateAssociatedPersonRequest,
	) (*AssociatedPersonResponse, error)
	// DeleteAssociatedPerson soft-deletes a specific associated person.
	DeleteAssociatedPerson(ctx context.Context, id svc.CustomerID, associatedPersonID string) error
}

// Common types for customer and associated person operations.
type (
	// Address represents a physical or registered address for business or personal use.
	// This structure is used for both registered business addresses and residential addresses.
	Address struct {
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
	IdentifyingInformation struct {
		// Type is the type of identification document (e.g., "drivers_license", "passport").
		Type IDType `json:"type"`
		// IssuingCountry is the country that issued the identification document.
		IssuingCountry string `json:"issuing_country"`
		// ImageFront is the front image of the ID document in data-uri format.
		// Supported formats: jpeg, jpg, png, heic, tif.
		ImageFront string `json:"image_front"`
		// ImageBack is the back image of the ID document in data-uri format.
		// Supported formats: jpeg, jpg, png, heic, tif.
		ImageBack string `json:"image_back"`
		// NationalIdentityNumber is the national identity number from the ID document.
		// This field is required for KYC verification. Maximum length: 128 characters.
		NationalIdentityNumber string `json:"national_identity_number"`
	}

	// Document represents a business document attachment for KYB verification.
	// Documents may include certificates of incorporation, operating agreements, or other legal documents.
	Document struct {
		// DocType is the type of document (e.g., "certificate_of_incorporation").
		DocType DocumentType `json:"doc_type"`
		// File is the document file in data-uri format.
		// Format: "data:image/[type];base64,[base64_data]" where type is jpeg, jpg, png, heic, or tif.
		File string `json:"file"`
		// Description is an optional description of the document.
		Description string `json:"description,omitempty"`
	}
)

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
	// Gender is the person's gender (required). Valid values: "male", "female".
	Gender Gender `json:"gender"`
	// ResidentialAddress is the person's current residential address.
	ResidentialAddress *Address `json:"residential_address"`
	// BirthDate is the person's date of birth in ISO format (e.g., "1980-01-15").
	BirthDate string `json:"birth_date"`
	// CountryOfBirth is the country where the person was born (required).
	CountryOfBirth string `json:"country_of_birth"`
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
	// TaxID is the person's tax identification number.
	TaxID string `json:"tax_id"`
	// POA is the Proof of Address document in data-uri format (required).
	// Format: "data:[mime_type];base64,[base64_data]"
	// Supported formats: jpeg, jpg, png, pdf, csv, xls, xlsx.
	POA string `json:"poa"`
	// POAType is the type of proof of address document (required).
	// Examples: "utility_bill", "bank_statement", "government_letter". Maximum length: 64 characters.
	POAType string `json:"poa_type"`
}

// CreateCustomer request and response types.
type (
	// CreateCustomerRequest represents the request body for creating a business customer.
	// This request includes comprehensive KYB (Know Your Business) information required for
	// regulatory compliance, including business details, beneficial ownership, and risk assessment.
	CreateCustomerRequest struct {
		// BusinessLegalName is the official registered legal name of the business entity.
		BusinessLegalName string `json:"business_legal_name"`
		// BusinessDescription provides a detailed description of the business operations.
		BusinessDescription string `json:"business_description"`
		// BusinessRegistrationNumber is the official business registration or incorporation number.
		BusinessRegistrationNumber string `json:"business_registration_number"`
		// Email is the primary contact email address for the business.
		Email string `json:"email"`
		// BusinessType specifies the legal structure (e.g., "cooperative", "corporation", "llc").
		BusinessType BusinessType `json:"business_type"`
		// BusinessIndustry is a NAICS code representing the business industry (e.g., "541519").
		BusinessIndustry string `json:"business_industry"`
		// RegisteredAddress is the official registered address of the business.
		RegisteredAddress *Address `json:"registered_address"`
		// DateOfIncorporation is the date when the business was officially incorporated (ISO format).
		DateOfIncorporation string `json:"date_of_incorporation"`
		// PhysicalAddress is the actual operating address if different from registered address.
		PhysicalAddress *Address `json:"physical_address,omitempty"`
		// SignedAgreementID is the identifier of the signed service agreement.
		SignedAgreementID string `json:"signed_agreement_id"`
		// IsDAO indicates whether this is a Decentralized Autonomous Organization.
		IsDAO bool `json:"is_dao"`
		// AssociatedPersons is a list of all persons associated with the business.
		AssociatedPersons []AssociatedPerson `json:"associated_persons"`
		// AccountPurpose describes the primary purpose of the account.
		AccountPurpose AccountPurpose `json:"account_purpose"`
		// SourceOfFunds is a list of sources for the funds being used.
		SourceOfFunds []SourceOfFunds `json:"source_of_funds"`
		// SourceOfWealth is a list of sources for the business's wealth.
		SourceOfWealth []SourceOfWealth `json:"source_of_wealth"`
		// Documents is a list of supporting documents for KYB verification (optional).
		Documents []Document `json:"documents,omitempty"`
		// PrimaryWebsite is the business's primary website URL (optional).
		PrimaryWebsite string `json:"primary_website,omitempty"`
		// PubliclyTraded indicates whether the business is publicly traded on a stock exchange.
		PubliclyTraded bool `json:"publicly_traded"`
		// EstimatedAnnualRevenueUSD is the estimated annual revenue range (e.g., "0_99999").
		EstimatedAnnualRevenueUSD MoneyRange `json:"estimated_annual_revenue_usd"`
		// ExpectedMonthlyFiatDeposits is the expected monthly fiat deposit range.
		ExpectedMonthlyFiatDeposits MoneyRange `json:"expected_monthly_fiat_deposits"`
		// ExpectedMonthlyFiatWithdrawals is the expected monthly fiat withdrawal range.
		ExpectedMonthlyFiatWithdrawals MoneyRange `json:"expected_monthly_fiat_withdrawals"`
		// AccountPurposeOther provides additional details if AccountPurpose is "other".
		AccountPurposeOther string `json:"account_purpose_other,omitempty"`
		// HighRiskActivities is a list of high-risk business activities.
		HighRiskActivities []HighRiskActivity `json:"high_risk_activities,omitempty"`
		// HighRiskActivitiesExplanation provides additional context for high-risk activities.
		HighRiskActivitiesExplanation string `json:"high_risk_activities_explanation,omitempty"`
		// TaxID is the business tax identification number.
		TaxID string `json:"tax_id"`
		// TaxType is the type of tax ID (e.g., "EIN", "TIN").
		TaxType TaxIDType `json:"tax_type"`
		// TaxCountry is the country where the business is subject to taxation (ISO 3166-1 alpha-3).
		TaxCountry string `json:"tax_country"`
	}

	// CustomerResponse represents the standard customer response data.
	// This structure is used for customer creation, retrieval, and update operations.
	CustomerResponse struct {
		// CustomerID is the unique identifier of the customer.
		CustomerID string `json:"customer_id"`
		// Email is the primary contact email for the customer.
		Email string `json:"email"`
		// BusinessLegalName is the legal business name.
		BusinessLegalName string `json:"business_legal_name"`
		// BusinessDescription provides a detailed description of the business operations.
		BusinessDescription string `json:"business_description,omitempty"`
		// BusinessType is the type of business entity.
		BusinessType BusinessType `json:"business_type"`
		// BusinessIndustry is a NAICS code representing the business industry.
		BusinessIndustry string `json:"business_industry,omitempty"`
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
		// Status is the current KYB verification status.
		Status KybStatus `json:"status"`
		// SubmittedAt is the timestamp when the customer application was submitted (ISO 8601 format).
		SubmittedAt string `json:"submitted_at,omitempty"`
		// CreatedAt is the timestamp when the customer account was created (ISO 8601 format).
		CreatedAt string `json:"created_at"`
		// UpdatedAt is the timestamp when the customer account was last updated (ISO 8601 format).
		UpdatedAt string `json:"updated_at"`
	}

	// CreateCustomerResponse is an alias for CustomerResponse.
	CreateCustomerResponse = CustomerResponse
)

// UpdateCustomer request and response types.
type (
	// UpdateCustomerRequest represents the request body for updating an existing business customer.
	// This request supports partial updates - only the fields that are provided will be updated.
	UpdateCustomerRequest struct {
		// BusinessLegalName is the official registered legal name of the business entity.
		BusinessLegalName *string `json:"business_legal_name,omitempty"`
		// BusinessDescription provides a detailed description of the business operations.
		BusinessDescription *string `json:"business_description,omitempty"`
		// BusinessRegistrationNumber is the official business registration or incorporation number.
		BusinessRegistrationNumber *string `json:"business_registration_number,omitempty"`
		// Email is the primary contact email address for the business.
		Email *string `json:"email,omitempty"`
		// BusinessType specifies the legal structure (e.g., "cooperative", "corporation", "llc").
		BusinessType *BusinessType `json:"business_type,omitempty"`
		// BusinessIndustry is a NAICS code representing the business industry.
		BusinessIndustry *string `json:"business_industry,omitempty"`
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
		// AssociatedPersons is a list of all persons associated with the business.
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
		// TaxID is the business tax identification number.
		TaxID *string `json:"tax_id,omitempty"`
		// TaxType is the type of tax ID (e.g., "EIN", "TIN").
		TaxType *TaxIDType `json:"tax_type,omitempty"`
		// TaxCountry is the country where the business is subject to taxation (ISO 3166-1 alpha-3).
		TaxCountry *string `json:"tax_country,omitempty"`
	}

	// UpdateCustomerResponse is an alias for CustomerResponse.
	UpdateCustomerResponse = CustomerResponse
)

// ListCustomers request and response types.
type (
	// ListCustomersRequest represents the request parameters for listing customers.
	ListCustomersRequest struct {
		// PageSize is the number of records per page (1-100, default 10).
		PageSize int `json:"page_size,omitempty"`
		// PageNum is the page number, 0-indexed (default 0).
		PageNum int `json:"page_num,omitempty"`
		// KybStatus filters customers by their KYB verification status.
		KybStatus string `json:"kyb_status,omitempty"`
	}

	// CustomerSummary represents a summary of a customer account in list responses.
	CustomerSummary struct {
		// CustomerID is the unique identifier of the customer.
		CustomerID string `json:"customer_id"`
		// Email is the primary contact email.
		Email string `json:"email"`
		// BusinessLegalName is the legal business name.
		BusinessLegalName string `json:"business_legal_name"`
		// BusinessType is the type of business entity.
		BusinessType BusinessType `json:"business_type"`
		// Status is the current KYB verification status.
		Status KybStatus `json:"status"`
		// CreatedAt is the timestamp when the customer was created (ISO 8601 format).
		CreatedAt string `json:"created_at"`
		// UpdatedAt is the timestamp when the customer was last updated (ISO 8601 format).
		UpdatedAt string `json:"updated_at"`
	}

	// ListCustomersResponse represents the response data for listing customers.
	ListCustomersResponse struct {
		// Customers is the list of customer summaries returned by the API.
		Customers []CustomerSummary `json:"customers"`
		// Total is the number of customers in the current response.
		Total int `json:"total"`
	}
)

// AssociatedPerson request and response types.
type (
	// CreateAssociatedPersonRequest represents the request body for creating an associated person.
	CreateAssociatedPersonRequest struct {
		AssociatedPerson
	}

	// AssociatedPersonResponse represents the response data for an associated person.
	AssociatedPersonResponse struct {
		// AssociatedPersonID is the unique identifier for this associated person.
		AssociatedPersonID string `json:"associated_person_id"`
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
	UpdateAssociatedPersonRequest struct {
		// FirstName is the person's legal first name.
		FirstName *string `json:"first_name,omitempty"`
		// MiddleName is the person's middle name.
		MiddleName *string `json:"middle_name,omitempty"`
		// LastName is the person's legal last name.
		LastName *string `json:"last_name,omitempty"`
		// Email is the person's contact email address.
		Email *string `json:"email,omitempty"`
		// Gender is the person's gender. Valid values: "male", "female".
		Gender *Gender `json:"gender,omitempty"`
		// ResidentialAddress is the person's current residential address.
		ResidentialAddress *Address `json:"residential_address,omitempty"`
		// BirthDate is the person's date of birth in ISO format (e.g., "1980-01-15").
		BirthDate *string `json:"birth_date,omitempty"`
		// CountryOfBirth is the country where the person was born.
		CountryOfBirth *string `json:"country_of_birth,omitempty"`
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
		// TaxID is the person's tax identification number.
		TaxID *string `json:"tax_id,omitempty"`
		// POA is the Proof of Address document in data-uri format.
		POA *string `json:"poa,omitempty"`
		// POAType is the type of proof of address document.
		POAType *string `json:"poa_type,omitempty"`
	}

	// ListAssociatedPersonsResponse represents the response data for listing associated persons.
	ListAssociatedPersonsResponse []AssociatedPersonResponse
)

// TOS (Terms of Service) request and response types.
type (
	// CreateTOSLinkRequest represents the request body for creating a TOS signing link.
	CreateTOSLinkRequest struct {
		// RedirectUri is the URL where the user will be redirected after signing the TOS.
		// The URL will be appended to the TOS link as a query parameter.
		RedirectUri string `json:"redirectUri,omitempty"`
	}

	// TOSLinkResponse represents the response data for creating a TOS signing link.
	TOSLinkResponse struct {
		// Url is the hosted TOS signing page URL.
		// If RedirectUri was provided, it will be included as a query parameter.
		Url string `json:"url"`
		// SessionToken is the unique token for the TOS signing session.
		SessionToken string `json:"sessionToken"`
		// ExpiresIn is the number of seconds until the session token expires.
		ExpiresIn int `json:"expiresIn"`
	}

	// SignAgreementResponse represents the response data for signing a TOS agreement.
	SignAgreementResponse struct {
		// SignedAgreementID is the unique identifier for the signed agreement.
		SignedAgreementID string `json:"signedAgreementId"`
	}
)

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new customer service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// CreateTOSLink creates a session token for signing the Terms of Service agreement.
// This is the first step in the customer onboarding flow. The session expires in 1 hour.
func (s *serviceImpl) CreateTOSLink(ctx context.Context, req *CreateTOSLinkRequest) (*TOSLinkResponse, error) {
	path := fmt.Sprintf("%s/tos_links", ROUTE_PREFIX)
	if req == nil {
		req = &CreateTOSLinkRequest{}
	}
	return svc.PostJSON[*CreateTOSLinkRequest, TOSLinkResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
}

// SignTOSAgreement signs the Terms of Service agreement using the session token.
// This is the second step in the customer onboarding flow.
// Returns a signed_agreement_id to be used in customer creation.
func (s *serviceImpl) SignTOSAgreement(ctx context.Context, sessionToken string) (*SignAgreementResponse, error) {
	path := fmt.Sprintf("%s/tos_links/%s/sign", ROUTE_PREFIX, sessionToken)
	return svc.PostJSON[any, SignAgreementResponse](
		ctx,
		s.BaseService,
		path,
		nil,
	)
}

// CreateCustomer creates a new customer using the generic PostJSON function.
func (s *serviceImpl) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error) {
	return svc.PostJSON[*CreateCustomerRequest, CreateCustomerResponse](
		ctx,
		s.BaseService,
		ROUTE_PREFIX,
		req,
	)
}

// ListCustomers retrieves a list of customers with optional filtering and pagination.
func (s *serviceImpl) ListCustomers(ctx context.Context, req *ListCustomersRequest) (*ListCustomersResponse, error) {
	params := make(map[string]string)

	if req != nil {
		if req.PageSize > 0 {
			params["page_size"] = fmt.Sprintf("%d", req.PageSize)
		}
		if req.PageNum > 0 {
			params["page_num"] = fmt.Sprintf("%d", req.PageNum)
		}
		if req.KybStatus != "" {
			params["kyb_status"] = req.KybStatus
		}
	}

	return svc.GetJSONWithParams[ListCustomersResponse](
		ctx,
		s.BaseService,
		ROUTE_PREFIX,
		params,
	)
}

// GetCustomer retrieves a specific customer by ID.
func (s *serviceImpl) GetCustomer(ctx context.Context, id svc.CustomerID) (*CustomerResponse, error) {
	path := fmt.Sprintf("%s/%s", ROUTE_PREFIX, id)
	return svc.GetJSON[CustomerResponse](ctx, s.BaseService, path)
}

// UpdateCustomer updates an existing customer with partial KYB information.
// Only the fields provided in the request will be updated; nil/omitted fields remain unchanged.
func (s *serviceImpl) UpdateCustomer(
	ctx context.Context, id svc.CustomerID, req *UpdateCustomerRequest,
) (*UpdateCustomerResponse, error) {
	path := fmt.Sprintf("%s/%s", ROUTE_PREFIX, id)
	return svc.PutJSON[*UpdateCustomerRequest, UpdateCustomerResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
}

// CreateAssociatedPerson creates a new associated person for a customer.
func (s *serviceImpl) CreateAssociatedPerson(
	ctx context.Context,
	id svc.CustomerID,
	req *CreateAssociatedPersonRequest,
) (*AssociatedPersonResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons", ROUTE_PREFIX, id)
	return svc.PostJSON[*CreateAssociatedPersonRequest, AssociatedPersonResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
}

// ListAssociatedPersons retrieves all associated persons for a specific customer.
func (s *serviceImpl) ListAssociatedPersons(ctx context.Context, id svc.CustomerID) (*ListAssociatedPersonsResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons", ROUTE_PREFIX, id)
	return svc.GetJSON[ListAssociatedPersonsResponse](ctx, s.BaseService, path)
}

// GetAssociatedPerson retrieves a specific associated person by ID.
func (s *serviceImpl) GetAssociatedPerson(
	ctx context.Context,
	id svc.CustomerID,
	associatedPersonID string,
) (*AssociatedPersonResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons/%s", ROUTE_PREFIX, id, associatedPersonID)
	return svc.GetJSON[AssociatedPersonResponse](ctx, s.BaseService, path)
}

// UpdateAssociatedPerson updates an existing associated person with partial data.
// Only the fields provided in the request will be updated; nil/omitted fields remain unchanged.
func (s *serviceImpl) UpdateAssociatedPerson(
	ctx context.Context,
	id svc.CustomerID,
	associatedPersonID string,
	req *UpdateAssociatedPersonRequest,
) (*AssociatedPersonResponse, error) {
	path := fmt.Sprintf("%s/%s/associated_persons/%s", ROUTE_PREFIX, id, associatedPersonID)
	return svc.PutJSON[*UpdateAssociatedPersonRequest, AssociatedPersonResponse](
		ctx,
		s.BaseService,
		path,
		req,
	)
}

// DeleteAssociatedPerson soft-deletes a specific associated person.
func (s *serviceImpl) DeleteAssociatedPerson(
	ctx context.Context,
	id svc.CustomerID,
	associatedPersonID string,
) error {
	path := fmt.Sprintf("%s/%s/associated_persons/%s", ROUTE_PREFIX, id, associatedPersonID)
	_, err := svc.DeleteJSON[any](ctx, s.BaseService, path)
	return err
}
