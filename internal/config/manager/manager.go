package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/arthur404dev/heimdall-cli/internal/config/providers"
	"github.com/arthur404dev/heimdall-cli/internal/config/schema"
	"github.com/arthur404dev/heimdall-cli/internal/config/types"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Manager coordinates all configuration providers
type Manager struct {
	providers   map[string]providers.Provider
	registry    *schema.Registry
	paths       *types.ConfigPaths
	mu          sync.RWMutex
	initialized bool
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]providers.Provider),
		registry:  schema.NewRegistry(),
		paths:     getDefaultPaths(),
	}
}

// getDefaultPaths returns default configuration paths
func getDefaultPaths() *types.ConfigPaths {
	// Check environment variables first
	baseDir := os.Getenv("HEIMDALL_CONFIG_DIR")
	if baseDir == "" {
		baseDir = paths.HeimdallConfigDir
	}

	schemaDir := os.Getenv("HEIMDALL_SCHEMA_DIR")
	if schemaDir == "" {
		schemaDir = filepath.Join(baseDir, "schemas")
	}

	backupDir := os.Getenv("HEIMDALL_BACKUP_DIR")
	if backupDir == "" {
		backupDir = filepath.Join(baseDir, "backups")
	}

	return &types.ConfigPaths{
		BaseDir:     baseDir,
		FilePattern: "%s.json",
		SchemaDir:   schemaDir,
		BackupDir:   backupDir,
		OutputPaths: make(map[string]string),
	}
}

// SetPaths updates the configuration paths
func (m *Manager) SetPaths(paths *types.ConfigPaths) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot change paths after initialization")
	}

	m.paths = paths
	return nil
}

// LoadPathsFromConfig loads paths from the main config
func (m *Manager) LoadPathsFromConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config struct {
		ConfigPaths *types.ConfigPaths `json:"config_paths,omitempty"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if config.ConfigPaths != nil {
		return m.SetPaths(config.ConfigPaths)
	}

	return nil
}

// Initialize sets up the manager with default providers
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	// Ensure directories exist
	if err := paths.EnsureDir(m.paths.BaseDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := paths.EnsureDir(m.paths.SchemaDir); err != nil {
		return fmt.Errorf("failed to create schema directory: %w", err)
	}
	if err := paths.EnsureDir(m.paths.BackupDir); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Register default providers
	// CLI provider for the main heimdall config
	cliProvider := providers.NewCLIProvider(filepath.Join(m.paths.BaseDir, "config.json"))
	if err := m.registerProviderLocked("cli", cliProvider); err != nil {
		return fmt.Errorf("failed to register CLI provider: %w", err)
	}

	// Shell provider for quickshell config
	shellConfigPath := filepath.Join(m.paths.BaseDir, fmt.Sprintf(m.paths.FilePattern, "shell"))
	shellProvider := providers.NewShellProvider(shellConfigPath, m.paths)
	if err := m.registerProviderLocked("shell", shellProvider); err != nil {
		return fmt.Errorf("failed to register shell provider: %w", err)
	}

	m.initialized = true
	return nil
}

// RegisterProvider adds a new configuration provider
func (m *Manager) RegisterProvider(domain string, provider providers.Provider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.registerProviderLocked(domain, provider)
}

// registerProviderLocked registers a provider (must be called with lock held)
func (m *Manager) registerProviderLocked(domain string, provider providers.Provider) error {
	// Initialize provider
	if err := provider.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Extract and register schema
	schema, err := provider.GetSchema()
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	if err := m.registry.Register(domain, schema); err != nil {
		return fmt.Errorf("failed to register schema: %w", err)
	}

	m.providers[domain] = provider
	return nil
}

// Get retrieves a configuration value
func (m *Manager) Get(domain, path string) (interface{}, error) {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown configuration domain: %s", domain)
	}

	return provider.Get(path)
}

// Set updates a configuration value
func (m *Manager) Set(domain, path string, value interface{}) error {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	schema := m.registry.GetSchema(domain)
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown configuration domain: %s", domain)
	}

	// Validate against schema if available
	if schema != nil {
		if err := schema.ValidateValue(path, value); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return provider.Set(path, value)
}

// GetAll retrieves all configuration values for a domain
func (m *Manager) GetAll(domain string) (map[string]interface{}, error) {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown configuration domain: %s", domain)
	}

	return provider.GetAll()
}

// SetAll updates all configuration values for a domain
func (m *Manager) SetAll(domain string, config map[string]interface{}) error {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	schema := m.registry.GetSchema(domain)
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown configuration domain: %s", domain)
	}

	// Validate against schema if available
	if schema != nil {
		if err := schema.Validate(config); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return provider.SetAll(config)
}

// Save persists configuration for a domain
func (m *Manager) Save(domain string) error {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown configuration domain: %s", domain)
	}

	return provider.Save()
}

// SaveAll persists all configurations
func (m *Manager) SaveAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []string
	for domain, provider := range m.providers {
		if err := provider.Save(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", domain, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to save configurations: %s", strings.Join(errs, "; "))
	}

	return nil
}

// Load reads configuration for a domain
func (m *Manager) Load(domain string) error {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown configuration domain: %s", domain)
	}

	return provider.Load()
}

// LoadAll reads all configurations
func (m *Manager) LoadAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []string
	for domain, provider := range m.providers {
		if err := provider.Load(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", domain, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to load configurations: %s", strings.Join(errs, "; "))
	}

	return nil
}

// ListDomains returns all registered configuration domains
func (m *Manager) ListDomains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]string, 0, len(m.providers))
	for domain := range m.providers {
		domains = append(domains, domain)
	}
	return domains
}

// GetSchema returns the schema for a domain
func (m *Manager) GetSchema(domain string) (*schema.Schema, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schema := m.registry.GetSchema(domain)
	if schema == nil {
		return nil, fmt.Errorf("no schema found for domain: %s", domain)
	}

	return schema, nil
}

// Validate checks if a configuration is valid
func (m *Manager) Validate(domain string) error {
	m.mu.RLock()
	provider, exists := m.providers[domain]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown configuration domain: %s", domain)
	}

	return provider.Validate()
}

// ApplyAll applies a function to all domains
func (m *Manager) ApplyAll(fn func(domain string, provider providers.Provider) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []string
	for domain, provider := range m.providers {
		if err := fn(domain, provider); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", domain, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("operation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetPaths returns the current configuration paths
func (m *Manager) GetPaths() *types.ConfigPaths {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.paths
}
