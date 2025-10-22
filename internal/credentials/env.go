package credentials

import (
	"os"
)

const (
	// Environment variable names (similar to AWS).
	EnvAccessKey = "ONEMONEY_ACCESS_KEY"
	EnvSecretKey = "ONEMONEY_SECRET_KEY"
	EnvBaseURL   = "ONEMONEY_BASE_URL"
)

// EnvProvider retrieves credentials from environment variables.
type EnvProvider struct{}

// NewEnvProvider creates a new environment variable credential provider.
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// Retrieve retrieves credentials from environment variables.
// Returns ErrNoCredentials if required environment variables are not set.
// Required: ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY
// Optional: ONEMONEY_BASE_URL
func (p *EnvProvider) Retrieve() (*Credentials, error) {
	accessKey := os.Getenv(EnvAccessKey)
	secretKey := os.Getenv(EnvSecretKey)
	baseURL := os.Getenv(EnvBaseURL)

	// Check which required variables are missing
	var missing []string
	if accessKey == "" {
		missing = append(missing, EnvAccessKey)
	}
	if secretKey == "" {
		missing = append(missing, EnvSecretKey)
	}

	if len(missing) > 0 {
		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      ErrNoCredentials,
			Message:  "missing required environment variables: " + joinStrings(missing, ", "),
		}
	}

	creds := &Credentials{
		AccessKey: accessKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}

	// Validate the retrieved credentials
	if !creds.IsValid() {
		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      ErrInvalidCredentials,
			Message:  "environment variables are set but credentials are invalid",
		}
	}

	return creds, nil
}

// joinStrings joins strings with a separator (helper to avoid importing strings package).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// Name returns the provider name.
func (p *EnvProvider) Name() string {
	return "EnvProvider"
}
