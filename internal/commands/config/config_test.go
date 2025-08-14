package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/config/manager"
	"github.com/arthur404dev/heimdall-cli/internal/config/schema"
	"github.com/arthur404dev/heimdall-cli/internal/config/types"
	"github.com/spf13/cobra"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	domain        string
	configPath    string
	config        map[string]interface{}
	schema        *schema.Schema
	initError     error
	loadError     error
	saveError     error
	getError      error
	setError      error
	validateError error
	schemaError   error
	initialized   bool
	loaded        bool
	saved         bool
}

// NewMockProvider creates a new mock provider
func NewMockProvider(domain string, configPath string) *MockProvider {
	return &MockProvider{
		domain:     domain,
		configPath: configPath,
		config:     make(map[string]interface{}),
		schema:     createMockSchema(domain),
	}
}

// createMockSchema creates a basic schema for testing
func createMockSchema(domain string) *schema.Schema {
	schemaJSON := fmt.Sprintf(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title": "%s Configuration",
		"description": "Configuration schema for %s domain",
		"type": "object",
		"properties": {
			"theme": {
				"type": "string",
				"description": "Theme name",
				"enum": ["dark", "light"]
			},
			"appearance": {
				"type": "object",
				"properties": {
					"colorScheme": {
						"type": "string",
						"description": "Color scheme name"
					},
					"fontSize": {
						"type": "number",
						"minimum": 8,
						"maximum": 72
					}
				}
			},
			"enabled": {
				"type": "boolean",
				"description": "Whether the feature is enabled"
			}
		},
		"required": ["theme"]
	}`, domain, domain)

	s, _ := schema.NewSchema([]byte(schemaJSON))
	return s
}

// Provider interface implementations
func (mp *MockProvider) Initialize() error {
	if mp.initError != nil {
		return mp.initError
	}
	mp.initialized = true
	return nil
}

func (mp *MockProvider) GetSchema() (*schema.Schema, error) {
	if mp.schemaError != nil {
		return nil, mp.schemaError
	}
	return mp.schema, nil
}

func (mp *MockProvider) Load() error {
	if mp.loadError != nil {
		return mp.loadError
	}
	mp.loaded = true
	return nil
}

func (mp *MockProvider) Save() error {
	if mp.saveError != nil {
		return mp.saveError
	}
	mp.saved = true
	return nil
}

func (mp *MockProvider) Get(path string) (interface{}, error) {
	if mp.getError != nil {
		return nil, mp.getError
	}

	parts := strings.Split(path, ".")
	current := mp.config

	for i, part := range parts {
		if i == len(parts)-1 {
			if value, exists := current[part]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("path not found: %s", path)
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("path not found: %s", path)
		}
	}

	return nil, fmt.Errorf("path not found: %s", path)
}

func (mp *MockProvider) Set(path string, value interface{}) error {
	if mp.setError != nil {
		return mp.setError
	}

	parts := strings.Split(path, ".")
	current := mp.config

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return nil
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			// Create intermediate objects
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	return nil
}

func (mp *MockProvider) GetAll() (map[string]interface{}, error) {
	if mp.getError != nil {
		return nil, mp.getError
	}
	return mp.config, nil
}

func (mp *MockProvider) SetAll(config map[string]interface{}) error {
	if mp.setError != nil {
		return mp.setError
	}
	mp.config = config
	return nil
}

func (mp *MockProvider) Validate() error {
	if mp.validateError != nil {
		return mp.validateError
	}
	return mp.schema.Validate(mp.config)
}

func (mp *MockProvider) GetConfigPath() string {
	return mp.configPath
}

func (mp *MockProvider) GetDomain() string {
	return mp.domain
}

// Test helper functions
func setupTestManager(t *testing.T) (*manager.Manager, string) {
	tempDir := t.TempDir()

	mgr := manager.NewManager()
	paths := &types.ConfigPaths{
		BaseDir:     tempDir,
		FilePattern: "%s.json",
		SchemaDir:   filepath.Join(tempDir, "schemas"),
		BackupDir:   filepath.Join(tempDir, "backups"),
		OutputPaths: make(map[string]string),
	}

	if err := mgr.SetPaths(paths); err != nil {
		t.Fatalf("Failed to set paths: %v", err)
	}

	return mgr, tempDir
}

func setupTestCommand(t *testing.T) (*cobra.Command, *manager.Manager, string) {
	testMgr, tempDir := setupTestManager(t)

	// Override the global manager for testing
	originalMgr := mgr
	mgr = testMgr

	t.Cleanup(func() {
		mgr = originalMgr
	})

	cmd := Command()
	return cmd, testMgr, tempDir
}

func captureOutput(t *testing.T, fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = old
	return <-done
}

// Test Command Creation and Structure
func TestCommand(t *testing.T) {
	cmd := Command()

	if cmd.Use != "config [domain] [operation] [args...]" {
		t.Errorf("Expected Use to be 'config [domain] [operation] [args...]', got %s", cmd.Use)
	}

	if cmd.Short != "Manage configuration files" {
		t.Errorf("Expected Short to be 'Manage configuration files', got %s", cmd.Short)
	}

	// Check that all expected subcommands are present
	expectedSubcommands := []string{"list", "get", "set", "validate", "save", "load", "schema", "all"}
	actualSubcommands := make([]string, 0, len(cmd.Commands()))

	for _, subcmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subcmd.Name())
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, actual := range actualSubcommands {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

func TestCommandPersistentPreRunE(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful initialization",
			setupEnv: func() {
				// Clean environment
				os.Unsetenv("HEIMDALL_CONFIG")
			},
			expectError: false,
		},
		{
			name: "with HEIMDALL_CONFIG environment variable",
			setupEnv: func() {
				tempDir := t.TempDir()
				tempFile := filepath.Join(tempDir, "config.json")
				configData := fmt.Sprintf(`{"config_paths": {"base_dir": "%s", "schema_dir": "%s/schemas", "backup_dir": "%s/backups", "file_pattern": "%%s.json"}}`,
					tempDir, tempDir, tempDir)
				os.WriteFile(tempFile, []byte(configData), 0644)
				os.Setenv("HEIMDALL_CONFIG", tempFile)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, _ := setupTestCommand(t)
			tt.setupEnv()
			defer os.Unsetenv("HEIMDALL_CONFIG")

			err := cmd.PersistentPreRunE(cmd, []string{})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test List Command
func TestListCommand(t *testing.T) {
	_, testMgr, _ := setupTestCommand(t)

	// Register mock providers
	cliProvider := NewMockProvider("cli", "/tmp/cli.json")
	shellProvider := NewMockProvider("shell", "/tmp/shell.json")

	testMgr.RegisterProvider("cli", cliProvider)
	testMgr.RegisterProvider("shell", shellProvider)

	listCmd := listCommand()

	output := captureOutput(t, func() {
		err := listCmd.RunE(listCmd, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Available configuration domains:") {
		t.Errorf("Expected output to contain 'Available configuration domains:', got %s", output)
	}

	if !strings.Contains(output, "cli") || !strings.Contains(output, "shell") {
		t.Errorf("Expected output to contain both 'cli' and 'shell' domains, got %s", output)
	}
}

// Test Get Command
func TestGetCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		setupData    func(*MockProvider)
		expectError  bool
		errorMsg     string
		expectOutput string
	}{
		{
			name: "get string value",
			args: []string{"cli", "theme"},
			setupData: func(mp *MockProvider) {
				mp.config["theme"] = "dark"
			},
			expectError:  false,
			expectOutput: "dark",
		},
		{
			name: "get nested value",
			args: []string{"cli", "appearance.colorScheme"},
			setupData: func(mp *MockProvider) {
				mp.config["appearance"] = map[string]interface{}{
					"colorScheme": "gruvbox",
				}
			},
			expectError:  false,
			expectOutput: "gruvbox",
		},
		{
			name: "get boolean value",
			args: []string{"cli", "enabled"},
			setupData: func(mp *MockProvider) {
				mp.config["enabled"] = true
			},
			expectError:  false,
			expectOutput: "true",
		},
		{
			name: "get complex object",
			args: []string{"cli", "appearance"},
			setupData: func(mp *MockProvider) {
				mp.config["appearance"] = map[string]interface{}{
					"colorScheme": "gruvbox",
					"fontSize":    12,
				}
			},
			expectError:  false,
			expectOutput: "colorScheme",
		},
		{
			name:        "invalid number of arguments",
			args:        []string{"cli"},
			expectError: true,
			errorMsg:    "accepts 2 arg(s), received 1",
		},
		{
			name: "unknown domain",
			args: []string{"unknown", "theme"},
			setupData: func(mp *MockProvider) {
				mp.getError = fmt.Errorf("unknown configuration domain: unknown")
			},
			expectError: true,
			errorMsg:    "unknown configuration domain: unknown",
		},
		{
			name: "path not found",
			args: []string{"cli", "nonexistent"},
			setupData: func(mp *MockProvider) {
				mp.getError = fmt.Errorf("path not found: nonexistent")
			},
			expectError: true,
			errorMsg:    "path not found: nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testMgr, _ := setupTestCommand(t)

			// Register mock provider
			cliProvider := NewMockProvider("cli", "/tmp/cli.json")
			if tt.setupData != nil {
				tt.setupData(cliProvider)
			}
			testMgr.RegisterProvider("cli", cliProvider)

			getCmd := getCommand()

			if tt.expectError {
				// For cobra argument validation errors, we need to use Execute
				getCmd.SetArgs(tt.args)
				err := getCmd.Execute()
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				output := captureOutput(t, func() {
					err := getCmd.RunE(getCmd, tt.args)
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				})

				if tt.expectOutput != "" && !strings.Contains(output, tt.expectOutput) {
					t.Errorf("Expected output to contain '%s', got %s", tt.expectOutput, output)
				}
			}
		})
	}
}

// Test Set Command
func TestSetCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		expectValue interface{}
		expectPath  string
	}{
		{
			name:        "set string value",
			args:        []string{"cli", "theme", "light"},
			expectError: false,
			expectValue: "light",
			expectPath:  "theme",
		},
		{
			name:        "set JSON boolean",
			args:        []string{"cli", "enabled", "true"},
			expectError: false,
			expectValue: true,
			expectPath:  "enabled",
		},
		{
			name:        "set JSON number",
			args:        []string{"cli", "appearance.fontSize", "14"},
			expectError: false,
			expectValue: float64(14),
			expectPath:  "appearance.fontSize",
		},
		{
			name:        "set JSON object",
			args:        []string{"cli", "appearance", `{"colorScheme": "gruvbox", "fontSize": 12}`},
			expectError: false,
			expectValue: map[string]interface{}{"colorScheme": "gruvbox", "fontSize": float64(12)},
			expectPath:  "appearance",
		},
		{
			name:        "invalid number of arguments",
			args:        []string{"cli", "theme"},
			expectError: true,
			errorMsg:    "accepts 3 arg(s), received 2",
		},
		{
			name:        "unknown domain",
			args:        []string{"unknown", "theme", "dark"},
			expectError: true,
			errorMsg:    "unknown configuration domain: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testMgr, _ := setupTestCommand(t)

			// Register mock provider
			cliProvider := NewMockProvider("cli", "/tmp/cli.json")
			cliProvider.config["theme"] = "dark" // Initial value
			testMgr.RegisterProvider("cli", cliProvider)

			setCmd := setCommand()

			if tt.expectError {
				// For cobra argument validation errors, we need to use Execute
				setCmd.SetArgs(tt.args)
				err := setCmd.Execute()
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				output := captureOutput(t, func() {
					err := setCmd.RunE(setCmd, tt.args)
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				})

				// Verify the value was set
				value, err := cliProvider.Get(tt.expectPath)
				if err != nil {
					t.Errorf("Failed to get value after set: %v", err)
				}

				if !equalValues(value, tt.expectValue) {
					t.Errorf("Expected value %v, got %v", tt.expectValue, value)
				}

				// Verify save was called
				if !cliProvider.saved {
					t.Errorf("Expected provider to be saved after set")
				}

				// Verify success message
				if !strings.Contains(output, "✓ Set") {
					t.Errorf("Expected success message in output, got %s", output)
				}
			}
		})
	}
}

// Test Validate Command
func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupData   func(*MockProvider)
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			args: []string{"cli"},
			setupData: func(mp *MockProvider) {
				mp.config["theme"] = "dark"
			},
			expectError: false,
		},
		{
			name: "invalid configuration",
			args: []string{"cli"},
			setupData: func(mp *MockProvider) {
				mp.validateError = fmt.Errorf("required field 'theme' is missing")
			},
			expectError: true,
			errorMsg:    "validation failed: required field 'theme' is missing",
		},
		{
			name:        "invalid number of arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "unknown domain",
			args:        []string{"unknown"},
			expectError: true,
			errorMsg:    "unknown configuration domain: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testMgr, _ := setupTestCommand(t)

			// Register mock provider
			cliProvider := NewMockProvider("cli", "/tmp/cli.json")
			if tt.setupData != nil {
				tt.setupData(cliProvider)
			}
			testMgr.RegisterProvider("cli", cliProvider)

			validateCmd := validateCommand()

			if tt.expectError {
				// For cobra argument validation errors, we need to use Execute
				validateCmd.SetArgs(tt.args)
				err := validateCmd.Execute()
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				output := captureOutput(t, func() {
					err := validateCmd.RunE(validateCmd, tt.args)
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				})

				if !strings.Contains(output, "✓ Configuration 'cli' is valid") {
					t.Errorf("Expected success message in output, got %s", output)
				}
			}
		})
	}
}

// Test Save Command
func TestSaveCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupData   func(*MockProvider)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful save",
			args:        []string{"cli"},
			expectError: false,
		},
		{
			name: "save error",
			args: []string{"cli"},
			setupData: func(mp *MockProvider) {
				mp.saveError = fmt.Errorf("permission denied")
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
		{
			name:        "invalid number of arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "unknown domain",
			args:        []string{"unknown"},
			expectError: true,
			errorMsg:    "unknown configuration domain: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testMgr, _ := setupTestCommand(t)

			// Register mock provider
			cliProvider := NewMockProvider("cli", "/tmp/cli.json")
			if tt.setupData != nil {
				tt.setupData(cliProvider)
			}
			testMgr.RegisterProvider("cli", cliProvider)

			saveCmd := saveCommand()

			if tt.expectError {
				// For cobra argument validation errors, we need to use Execute
				saveCmd.SetArgs(tt.args)
				err := saveCmd.Execute()
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				output := captureOutput(t, func() {
					err := saveCmd.RunE(saveCmd, tt.args)
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				})

				if !strings.Contains(output, "✓ Saved configuration 'cli'") {
					t.Errorf("Expected success message in output, got %s", output)
				}

				if !cliProvider.saved {
					t.Errorf("Expected provider to be saved")
				}
			}
		})
	}
}

// Test Load Command
func TestLoadCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupData   func(*MockProvider)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful load",
			args:        []string{"cli"},
			expectError: false,
		},
		{
			name: "load error",
			args: []string{"cli"},
			setupData: func(mp *MockProvider) {
				mp.loadError = fmt.Errorf("file not found")
			},
			expectError: true,
			errorMsg:    "file not found",
		},
		{
			name:        "invalid number of arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "unknown domain",
			args:        []string{"unknown"},
			expectError: true,
			errorMsg:    "unknown configuration domain: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testMgr, _ := setupTestCommand(t)

			// Register mock provider
			cliProvider := NewMockProvider("cli", "/tmp/cli.json")
			if tt.setupData != nil {
				tt.setupData(cliProvider)
			}
			testMgr.RegisterProvider("cli", cliProvider)

			loadCmd := loadCommand()

			if tt.expectError {
				// For cobra argument validation errors, we need to use Execute
				loadCmd.SetArgs(tt.args)
				err := loadCmd.Execute()
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				output := captureOutput(t, func() {
					err := loadCmd.RunE(loadCmd, tt.args)
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				})

				if !strings.Contains(output, "✓ Loaded configuration 'cli'") {
					t.Errorf("Expected success message in output, got %s", output)
				}

				if !cliProvider.loaded {
					t.Errorf("Expected provider to be loaded")
				}
			}
		})
	}
}

// Test Schema Command
func TestSchemaCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupData   func(*MockProvider)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful schema display",
			args:        []string{"cli"},
			expectError: false,
		},
		{
			name: "schema error",
			args: []string{"cli"},
			setupData: func(mp *MockProvider) {
				mp.schemaError = fmt.Errorf("schema not found")
			},
			expectError: true,
			errorMsg:    "no schema found for domain: cli",
		},
		{
			name:        "invalid number of arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "unknown domain",
			args:        []string{"unknown"},
			expectError: true,
			errorMsg:    "no schema found for domain: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, testMgr, _ := setupTestCommand(t)

			// Register mock provider
			cliProvider := NewMockProvider("cli", "/tmp/cli.json")
			if tt.setupData != nil {
				tt.setupData(cliProvider)
			}
			testMgr.RegisterProvider("cli", cliProvider)

			schemaCmd := schemaCommand()

			if tt.expectError {
				// For cobra argument validation errors, we need to use Execute
				schemaCmd.SetArgs(tt.args)
				err := schemaCmd.Execute()
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got %v", tt.errorMsg, err)
				}
			} else {
				output := captureOutput(t, func() {
					err := schemaCmd.RunE(schemaCmd, tt.args)
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				})

				if !strings.Contains(output, "cli Configuration") {
					t.Errorf("Expected schema output to contain 'cli Configuration', got %s", output)
				}
			}
		})
	}
}

// Test All Command
func TestAllCommand(t *testing.T) {
	_, testMgr, _ := setupTestCommand(t)

	// Register multiple mock providers
	cliProvider := NewMockProvider("cli", "/tmp/cli.json")
	shellProvider := NewMockProvider("shell", "/tmp/shell.json")

	cliProvider.config["theme"] = "dark"
	shellProvider.config["theme"] = "light"

	testMgr.RegisterProvider("cli", cliProvider)
	testMgr.RegisterProvider("shell", shellProvider)

	allCmd := allCommand()

	// Test validate all
	validateAllCmd := findSubcommand(allCmd, "validate")
	if validateAllCmd == nil {
		t.Fatal("validate subcommand not found in all command")
	}

	output := captureOutput(t, func() {
		err := validateAllCmd.RunE(validateAllCmd, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "✓ cli: valid") || !strings.Contains(output, "✓ shell: valid") {
		t.Errorf("Expected validation success for both domains, got %s", output)
	}

	// Test save all
	saveAllCmd := findSubcommand(allCmd, "save")
	if saveAllCmd == nil {
		t.Fatal("save subcommand not found in all command")
	}

	output = captureOutput(t, func() {
		err := saveAllCmd.RunE(saveAllCmd, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "✓ Saved all configurations") {
		t.Errorf("Expected save all success message, got %s", output)
	}

	// Test load all
	loadAllCmd := findSubcommand(allCmd, "load")
	if loadAllCmd == nil {
		t.Fatal("load subcommand not found in all command")
	}

	output = captureOutput(t, func() {
		err := loadAllCmd.RunE(loadAllCmd, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "✓ Loaded all configurations") {
		t.Errorf("Expected load all success message, got %s", output)
	}
}

// Test All Get Command - Basic functionality
func TestAllGetCommand(t *testing.T) {
	_, testMgr, _ := setupTestCommand(t)

	// Register multiple mock providers with data
	cliProvider := NewMockProvider("cli", "/tmp/cli.json")
	shellProvider := NewMockProvider("shell", "/tmp/shell.json")

	cliProvider.config["theme"] = "dark"
	shellProvider.config["theme"] = "light"

	testMgr.RegisterProvider("cli", cliProvider)
	testMgr.RegisterProvider("shell", shellProvider)

	allCmd := allCommand()
	getAllCmd := findSubcommand(allCmd, "get")
	if getAllCmd == nil {
		t.Fatal("get subcommand not found in all command")
	}

	// Test successful get from multiple domains
	output := captureOutput(t, func() {
		err := getAllCmd.RunE(getAllCmd, []string{"theme"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "cli: dark") || !strings.Contains(output, "shell: light") {
		t.Errorf("Expected output to contain both domain values, got %s", output)
	}

	// Test path not found
	err := getAllCmd.RunE(getAllCmd, []string{"nonexistent"})
	if err == nil {
		t.Errorf("Expected error for nonexistent path")
	} else if !strings.Contains(err.Error(), "path 'nonexistent' not found in any configuration") {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

// Test All Set Command - Basic functionality
func TestAllSetCommand(t *testing.T) {
	_, testMgr, _ := setupTestCommand(t)

	// Register multiple mock providers with data
	cliProvider := NewMockProvider("cli", "/tmp/cli.json")
	shellProvider := NewMockProvider("shell", "/tmp/shell.json")

	cliProvider.config["theme"] = "dark"
	shellProvider.config["theme"] = "dark"

	testMgr.RegisterProvider("cli", cliProvider)
	testMgr.RegisterProvider("shell", shellProvider)

	allCmd := allCommand()
	setAllCmd := findSubcommand(allCmd, "set")
	if setAllCmd == nil {
		t.Fatal("set subcommand not found in all command")
	}

	// Test successful set to multiple domains
	output := captureOutput(t, func() {
		err := setAllCmd.RunE(setAllCmd, []string{"theme", "light"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "✓ Set cli.theme to light") || !strings.Contains(output, "✓ Set shell.theme to light") {
		t.Errorf("Expected success messages for both domains, got %s", output)
	}

	// Test path not found
	err := setAllCmd.RunE(setAllCmd, []string{"nonexistent", "value"})
	if err == nil {
		t.Errorf("Expected error for nonexistent path")
	} else if !strings.Contains(err.Error(), "path 'nonexistent' not found in any configuration") {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

// Test All Validate Command with Errors
func TestAllValidateCommandWithErrors(t *testing.T) {
	_, testMgr, _ := setupTestCommand(t)

	// Register mock providers with validation errors
	cliProvider := NewMockProvider("cli", "/tmp/cli.json")
	shellProvider := NewMockProvider("shell", "/tmp/shell.json")

	cliProvider.validateError = fmt.Errorf("required field 'theme' is missing")
	shellProvider.config["theme"] = "light" // This one is valid

	testMgr.RegisterProvider("cli", cliProvider)
	testMgr.RegisterProvider("shell", shellProvider)

	allCmd := allCommand()
	validateAllCmd := findSubcommand(allCmd, "validate")
	if validateAllCmd == nil {
		t.Fatal("validate subcommand not found in all command")
	}

	output := captureOutput(t, func() {
		err := validateAllCmd.RunE(validateAllCmd, []string{})
		if err == nil {
			t.Errorf("Expected error but got none")
		} else if !strings.Contains(err.Error(), "validation failed for 1 domain(s)") {
			t.Errorf("Expected error about validation failure, got %v", err)
		}
	})

	if !strings.Contains(output, "✓ shell: valid") {
		t.Errorf("Expected shell to be valid, got %s", output)
	}

	if !strings.Contains(output, "✗ cli: required field 'theme' is missing") {
		t.Errorf("Expected cli validation error, got %s", output)
	}
}

// Test Helper Functions
func TestFormatPath(t *testing.T) {
	tests := []struct {
		domain   string
		path     string
		expected string
	}{
		{"cli", "theme", "cli.theme"},
		{"shell", "", "shell"},
		{"", "theme", ".theme"},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := formatPath(tt.domain, tt.path)
		if result != tt.expected {
			t.Errorf("formatPath(%s, %s) = %s, expected %s", tt.domain, tt.path, result, tt.expected)
		}
	}
}

func TestParseDomainPath(t *testing.T) {
	tests := []struct {
		combined       string
		expectedDomain string
		expectedPath   string
	}{
		{"cli.theme", "cli", "theme"},
		{"shell.appearance.colorScheme", "shell", "appearance.colorScheme"},
		{"cli", "cli", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		domain, path := parseDomainPath(tt.combined)
		if domain != tt.expectedDomain || path != tt.expectedPath {
			t.Errorf("parseDomainPath(%s) = (%s, %s), expected (%s, %s)",
				tt.combined, domain, path, tt.expectedDomain, tt.expectedPath)
		}
	}
}

// Utility functions for tests
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == name {
			return subcmd
		}
	}
	return nil
}

func equalValues(a, b interface{}) bool {
	// Handle JSON unmarshaling differences (float64 vs int)
	switch va := a.(type) {
	case float64:
		if vb, ok := b.(float64); ok {
			return va == vb
		}
		if vb, ok := b.(int); ok {
			return va == float64(vb)
		}
	case int:
		if vb, ok := b.(int); ok {
			return va == vb
		}
		if vb, ok := b.(float64); ok {
			return float64(va) == vb
		}
	case map[string]interface{}:
		if vb, ok := b.(map[string]interface{}); ok {
			if len(va) != len(vb) {
				return false
			}
			for k, v := range va {
				if !equalValues(v, vb[k]) {
					return false
				}
			}
			return true
		}
	}

	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
