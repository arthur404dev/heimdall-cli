package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/arthur404dev/heimdall-cli/internal/config/schema"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// BaseProvider provides common functionality for configuration providers
type BaseProvider struct {
	configPath string
	domain     string
	schema     *schema.Schema
	data       map[string]interface{}
	mu         sync.RWMutex
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(domain, configPath string) *BaseProvider {
	return &BaseProvider{
		domain:     domain,
		configPath: configPath,
		data:       make(map[string]interface{}),
	}
}

// Initialize sets up the provider
func (b *BaseProvider) Initialize() error {
	// Ensure parent directory exists
	if err := paths.EnsureParentDir(b.configPath); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config if it exists
	if paths.Exists(b.configPath) {
		return b.Load()
	}

	// Initialize with defaults from schema if available
	if b.schema != nil {
		b.initializeDefaults()
	}

	return nil
}

// initializeDefaults sets default values from schema
func (b *BaseProvider) initializeDefaults() {
	if b.schema == nil || b.schema.Properties == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for key, prop := range b.schema.Properties {
		if prop.Default != nil {
			b.data[key] = prop.Default
		}
	}
}

// SetSchema sets the schema for this provider
func (b *BaseProvider) SetSchema(s *schema.Schema) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.schema = s
}

// GetSchema returns the JSON schema for this provider
func (b *BaseProvider) GetSchema() (*schema.Schema, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.schema == nil {
		return nil, fmt.Errorf("no schema defined for provider %s", b.domain)
	}
	return b.schema, nil
}

// Load reads the configuration from storage
func (b *BaseProvider) Load() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := os.ReadFile(b.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	b.data = config
	return nil
}

// Save writes the configuration to storage
func (b *BaseProvider) Save() error {
	b.mu.RLock()
	data, err := json.MarshalIndent(b.data, "", "  ")
	b.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent directory exists
	if err := paths.EnsureParentDir(b.configPath); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write atomically
	tempFile := b.configPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := os.Rename(tempFile, b.configPath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// Get retrieves a value by path (e.g., "appearance.theme")
func (b *BaseProvider) Get(path string) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	parts := strings.Split(path, ".")
	current := b.data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, return the value
			value, exists := current[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found", path)
			}
			return value, nil
		}

		// Navigate deeper
		next, exists := current[part]
		if !exists {
			return nil, fmt.Errorf("key '%s' not found", strings.Join(parts[:i+1], "."))
		}

		// Check if it's an object
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("'%s' is not an object", strings.Join(parts[:i+1], "."))
		}

		current = nextMap
	}

	return nil, fmt.Errorf("key '%s' not found", path)
}

// Set updates a value by path
func (b *BaseProvider) Set(path string, value interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	parts := strings.Split(path, ".")
	current := b.data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, set the value
			current[part] = value
			return nil
		}

		// Navigate deeper, creating objects as needed
		next, exists := current[part]
		if !exists {
			// Create new object
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		} else {
			// Check if it's an object
			nextMap, ok := next.(map[string]interface{})
			if !ok {
				return fmt.Errorf("'%s' is not an object", strings.Join(parts[:i+1], "."))
			}
			current = nextMap
		}
	}

	return nil
}

// GetAll returns the entire configuration
func (b *BaseProvider) GetAll() (map[string]interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string]interface{})
	for k, v := range b.data {
		result[k] = deepCopy(v)
	}

	return result, nil
}

// SetAll replaces the entire configuration
func (b *BaseProvider) SetAll(config map[string]interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create a copy to prevent external modifications
	newData := make(map[string]interface{})
	for k, v := range config {
		newData[k] = deepCopy(v)
	}

	b.data = newData
	return nil
}

// Validate checks the entire configuration
func (b *BaseProvider) Validate() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.schema == nil {
		// No schema, consider valid
		return nil
	}

	return b.schema.Validate(b.data)
}

// GetConfigPath returns the file path for this configuration
func (b *BaseProvider) GetConfigPath() string {
	return b.configPath
}

// GetDomain returns the domain name for this provider
func (b *BaseProvider) GetDomain() string {
	return b.domain
}

// deepCopy creates a deep copy of a value
func deepCopy(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = deepCopy(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = deepCopy(v)
		}
		return result
	default:
		return v
	}
}
