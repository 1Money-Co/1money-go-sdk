package credentials

// StaticProvider provides credentials from static values (e.g., command-line flags).
type StaticProvider struct {
	creds Credentials
}

// NewStaticProvider creates a new static credential provider.
func NewStaticProvider(accessKey, secretKey, baseURL string) *StaticProvider {
	return &StaticProvider{
		creds: Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
			BaseURL:   baseURL,
		},
	}
}

// Retrieve returns the static credentials.
// Returns ErrNoCredentials if credentials are not provided or invalid.
func (p *StaticProvider) Retrieve() (*Credentials, error) {
	if !p.creds.IsValid() {
		var missing []string
		if p.creds.AccessKey == "" {
			missing = append(missing, "access_key")
		}
		if p.creds.SecretKey == "" {
			missing = append(missing, "secret_key")
		}

		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      ErrNoCredentials,
			Message:  "missing required credentials: " + joinStringsHelper(missing, ", "),
		}
	}
	return &p.creds, nil
}

// joinStringsHelper joins strings with a separator (helper function).
func joinStringsHelper(strs []string, sep string) string {
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
func (p *StaticProvider) Name() string {
	return "StaticProvider"
}
