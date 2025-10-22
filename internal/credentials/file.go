package credentials

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

const (
	// DefaultProfile is the default profile name.
	DefaultProfile = "default"

	// DefaultConfigDir is the default config directory.
	DefaultConfigDir = ".onemoney"

	// DefaultCredentialsFile is the default credentials file name.
	DefaultCredentialsFile = "credentials"
)

// FileProvider retrieves credentials from a config file (similar to ~/.aws/credentials).
type FileProvider struct {
	filePath string
	profile  string
}

// NewFileProvider creates a new file-based credential provider.
// If filePath is empty, it uses ~/.onemoney/credentials.
// If profile is empty, it uses "default".
func NewFileProvider(filePath, profile string) *FileProvider {
	if filePath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			filePath = filepath.Join(homeDir, DefaultConfigDir, DefaultCredentialsFile)
		}
	}

	if profile == "" {
		profile = DefaultProfile
	}

	return &FileProvider{
		filePath: filePath,
		profile:  profile,
	}
}

// Retrieve retrieves credentials from the config file.
// Returns ProviderError with detailed information about why credentials could not be loaded.
func (p *FileProvider) Retrieve() (*Credentials, error) {
	// Check if file exists
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      ErrNoCredentials,
			Message:  fmt.Sprintf("credentials file not found: %s", p.filePath),
		}
	}

	// Load INI file
	cfg, err := ini.Load(p.filePath)
	if err != nil {
		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      err,
			Message:  fmt.Sprintf("failed to parse credentials file: %s", p.filePath),
		}
	}

	// Get profile section
	section, err := cfg.GetSection(p.profile)
	if err != nil {
		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      ErrNoCredentials,
			Message:  fmt.Sprintf("profile '%s' not found in %s", p.profile, p.filePath),
		}
	}

	// Read credentials
	accessKey := section.Key("access_key").String()
	secretKey := section.Key("secret_key").String()
	baseURL := section.Key("base_url").String()

	// Check which required keys are missing
	var missing []string
	if accessKey == "" {
		missing = append(missing, "access_key")
	}
	if secretKey == "" {
		missing = append(missing, "secret_key")
	}

	if len(missing) > 0 {
		return nil, &ProviderError{
			Provider: p.Name(),
			Err:      ErrNoCredentials,
			Message:  fmt.Sprintf("missing required keys in profile '%s': %s", p.profile, joinStringsFile(missing, ", ")),
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
			Message:  fmt.Sprintf("credentials in profile '%s' are invalid", p.profile),
		}
	}

	return creds, nil
}

// joinStringsFile joins strings with a separator (helper function).
func joinStringsFile(strs []string, sep string) string {
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
func (p *FileProvider) Name() string {
	return "FileProvider"
}
