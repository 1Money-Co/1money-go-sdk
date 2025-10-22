package credentials

import (
	"errors"
	"fmt"
	"strings"
)

// ChainProvider chains multiple providers and retrieves credentials from the first
// provider that returns valid credentials.
//
// This is similar to AWS SDK's credential chain:
// 1. Static credentials (command-line flags)
// 2. Environment variables
// 3. Config file (~/.onemoney/credentials)
type ChainProvider struct {
	providers []Provider
}

// NewChainProvider creates a new chain provider with the given providers.
// Providers are checked in order until one returns valid credentials.
func NewChainProvider(providers ...Provider) *ChainProvider {
	return &ChainProvider{
		providers: providers,
	}
}

// NewDefaultChainProvider creates a chain provider with the default provider chain:
// 1. Static provider (if credentials are provided)
// 2. Environment variable provider
// 3. File provider (with optional profile)
func NewDefaultChainProvider(accessKey, secretKey, baseURL, profile string) *ChainProvider {
	var providers []Provider

	// 1. Static credentials (highest priority)
	if accessKey != "" && secretKey != "" {
		providers = append(providers, NewStaticProvider(accessKey, secretKey, baseURL))
	}

	// 2. Environment variables
	providers = append(providers, NewEnvProvider())

	// 3. Config file (lowest priority)
	providers = append(providers, NewFileProvider("", profile))

	return &ChainProvider{
		providers: providers,
	}
}

// Retrieve attempts to retrieve credentials from each provider in the chain.
// Returns the first valid credentials found.
// If no provider can supply valid credentials, returns a detailed error listing all attempts.
func (p *ChainProvider) Retrieve() (*Credentials, error) {
	var providerErrors []string

	for _, provider := range p.providers {
		creds, err := provider.Retrieve()
		if err == nil && creds != nil && creds.IsValid() {
			// Successfully retrieved valid credentials
			return creds, nil
		}

		// Record the failure
		if err != nil {
			// Check if it's a ProviderError with details
			var provErr *ProviderError
			if errors.As(err, &provErr) {
				providerErrors = append(providerErrors, fmt.Sprintf("  - %s", err.Error()))
			} else {
				providerErrors = append(providerErrors, fmt.Sprintf("  - %s: %v", provider.Name(), err))
			}
		} else {
			// Credentials were returned but invalid
			providerErrors = append(providerErrors, fmt.Sprintf("  - %s: returned invalid credentials", provider.Name()))
		}
	}

	// Build detailed error message
	errorMsg := fmt.Sprintf("%w: attempted to load credentials from %d provider(s):\n%s",
		ErrNoCredentials,
		len(p.providers),
		strings.Join(providerErrors, "\n"))

	return nil, &ProviderError{
		Provider: p.Name(),
		Err:      ErrNoCredentials,
		Message:  errorMsg,
	}
}

// Name returns the provider name.
func (p *ChainProvider) Name() string {
	return "ChainProvider"
}
