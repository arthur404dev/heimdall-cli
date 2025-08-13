package providers

import "github.com/arthur404dev/heimdall-cli/internal/config/schema"

// Provider defines the interface for configuration providers
type Provider interface {
	// Initialize sets up the provider
	Initialize() error

	// GetSchema returns the JSON schema for this provider
	GetSchema() (*schema.Schema, error)

	// Load reads the configuration from storage
	Load() error

	// Save writes the configuration to storage
	Save() error

	// Get retrieves a value by path (e.g., "appearance.theme")
	Get(path string) (interface{}, error)

	// Set updates a value by path
	Set(path string, value interface{}) error

	// GetAll returns the entire configuration
	GetAll() (map[string]interface{}, error)

	// SetAll replaces the entire configuration
	SetAll(config map[string]interface{}) error

	// Validate checks the entire configuration
	Validate() error

	// GetConfigPath returns the file path for this configuration
	GetConfigPath() string

	// GetDomain returns the domain name for this provider
	GetDomain() string
}
