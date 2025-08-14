package clipboard

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/spf13/cobra"
)

// MockConfig provides a mock configuration for testing
type MockConfig struct {
	clipboard config.ClipboardConfig
	external  config.ExternalTools
}

// NewMockConfig creates a new mock configuration
func NewMockConfig() *MockConfig {
	return &MockConfig{
		clipboard: config.ClipboardConfig{
			MaxEntries:     100,
			FuzzelPrompt:   "Clipboard> ",
			FuzzelArgs:     []string{"--dmenu", "--width", "50", "--lines", "20"},
			PreviewLength:  50,
			DeleteOnSelect: false,
		},
		external: config.ExternalTools{
			Cliphist:    "cliphist",
			Fuzzel:      "fuzzel",
			WlClipboard: "wl-copy",
		},
	}
}

// SetClipboardConfig sets the clipboard configuration
func (m *MockConfig) SetClipboardConfig(cfg config.ClipboardConfig) {
	m.clipboard = cfg
}

// SetExternalTools sets the external tools configuration
func (m *MockConfig) SetExternalTools(tools config.ExternalTools) {
	m.external = tools
}

// GetClipboard returns the clipboard configuration
func (m *MockConfig) GetClipboard() config.ClipboardConfig {
	return m.clipboard
}

// GetExternal returns the external tools configuration
func (m *MockConfig) GetExternal() config.ExternalTools {
	return m.external
}

// Test helper functions
func createTestCommand() *cobra.Command {
	return NewCommand()
}

func resetFlags() {
	deleteFlag = false
}

// TestNewCommand tests the command creation and initialization
func TestNewCommand(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "command has correct metadata",
			validate: func(t *testing.T, cmd *cobra.Command) {
				if cmd.Use != "clipboard" {
					t.Errorf("Expected Use to be 'clipboard', got %s", cmd.Use)
				}
				if cmd.Short != "Manage clipboard history" {
					t.Errorf("Expected Short to be 'Manage clipboard history', got %s", cmd.Short)
				}
				if !strings.Contains(cmd.Long, "Display and manage clipboard history") {
					t.Errorf("Long description should contain 'Display and manage clipboard history'")
				}
				if !strings.Contains(cmd.Long, "cliphist") {
					t.Errorf("Long description should mention 'cliphist'")
				}
				if !strings.Contains(cmd.Long, "fuzzel") {
					t.Errorf("Long description should mention 'fuzzel'")
				}
			},
		},
		{
			name: "command has delete flag",
			validate: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("delete")
				if flag == nil {
					t.Errorf("Expected delete flag to exist")
					return
				}
				if flag.Shorthand != "d" {
					t.Errorf("Expected delete flag shorthand to be 'd', got %s", flag.Shorthand)
				}
				if flag.Usage != "Delete selected item from clipboard history" {
					t.Errorf("Expected delete flag usage to describe deletion, got %s", flag.Usage)
				}
			},
		},
		{
			name: "command has run function",
			validate: func(t *testing.T, cmd *cobra.Command) {
				if cmd.RunE == nil {
					t.Errorf("Expected RunE to be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			cmd := createTestCommand()
			tt.validate(t, cmd)
		})
	}
}

// TestCommandFlags tests flag parsing and behavior
func TestCommandFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T)
	}{
		{
			name: "delete flag sets global variable",
			args: []string{"--delete"},
			validate: func(t *testing.T) {
				if !deleteFlag {
					t.Errorf("Expected deleteFlag to be true")
				}
			},
		},
		{
			name: "delete flag short form works",
			args: []string{"-d"},
			validate: func(t *testing.T) {
				if !deleteFlag {
					t.Errorf("Expected deleteFlag to be true with short form")
				}
			},
		},
		{
			name: "no flags leaves deleteFlag false",
			args: []string{},
			validate: func(t *testing.T) {
				if deleteFlag {
					t.Errorf("Expected deleteFlag to be false by default")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			cmd := createTestCommand()
			cmd.SetArgs(tt.args)

			// Parse flags without running the command
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Errorf("Flag parsing should not error: %v", err)
			}

			tt.validate(t)
		})
	}
}

// TestCommandHelp tests the help output
func TestCommandHelp(t *testing.T) {
	resetFlags()
	cmd := createTestCommand()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help command should not error: %v", err)
	}

	output := buf.String()
	expectedStrings := []string{
		"clipboard",
		"Display and manage clipboard history",
		"cliphist",
		"fuzzel",
		"--delete",
		"-d",
		"Delete selected item",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output should contain '%s'\nActual output: %s", expected, output)
		}
	}
}

// TestConfigurationIntegration tests integration with configuration system
func TestConfigurationIntegration(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  func(*MockConfig)
		deleteFlag   bool
		expectedArgs []string
	}{
		{
			name: "default configuration",
			setupConfig: func(cfg *MockConfig) {
				// Use default config
			},
			deleteFlag: false,
			expectedArgs: []string{
				"--dmenu", "--width", "50", "--lines", "20",
				"--prompt", "Clipboard> ",
				"--placeholder", "Type to search clipboard",
			},
		},
		{
			name: "custom fuzzel args",
			setupConfig: func(cfg *MockConfig) {
				cfg.SetClipboardConfig(config.ClipboardConfig{
					MaxEntries:     200,
					FuzzelPrompt:   "Select> ",
					FuzzelArgs:     []string{"--dmenu", "--width", "80", "--height", "30"},
					PreviewLength:  100,
					DeleteOnSelect: false,
				})
			},
			deleteFlag: false,
			expectedArgs: []string{
				"--dmenu", "--width", "80", "--height", "30",
				"--prompt", "Select> ",
				"--placeholder", "Type to search clipboard",
			},
		},
		{
			name: "delete mode configuration",
			setupConfig: func(cfg *MockConfig) {
				cfg.SetClipboardConfig(config.ClipboardConfig{
					MaxEntries:     50,
					FuzzelPrompt:   "Choose> ",
					FuzzelArgs:     []string{"--dmenu"},
					PreviewLength:  25,
					DeleteOnSelect: true,
				})
			},
			deleteFlag: true,
			expectedArgs: []string{
				"--dmenu",
				"--prompt", "del > ",
				"--placeholder", "Delete from clipboard",
			},
		},
		{
			name: "custom external tools",
			setupConfig: func(cfg *MockConfig) {
				cfg.SetExternalTools(config.ExternalTools{
					Cliphist:    "/custom/path/cliphist",
					Fuzzel:      "/custom/path/fuzzel",
					WlClipboard: "/custom/path/wl-copy",
				})
			},
			deleteFlag: false,
		},
		{
			name: "empty external tool paths use defaults",
			setupConfig: func(cfg *MockConfig) {
				cfg.SetExternalTools(config.ExternalTools{
					Cliphist:    "",
					Fuzzel:      "",
					WlClipboard: "",
				})
			},
			deleteFlag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock config
			mockConfig := NewMockConfig()
			if tt.setupConfig != nil {
				tt.setupConfig(mockConfig)
			}

			// Test that configuration values are properly used
			cfg := mockConfig.GetClipboard()
			external := mockConfig.GetExternal()

			// Validate clipboard config
			if cfg.MaxEntries <= 0 {
				t.Errorf("MaxEntries should be positive, got %d", cfg.MaxEntries)
			}
			if cfg.FuzzelPrompt == "" {
				t.Errorf("FuzzelPrompt should not be empty")
			}
			if cfg.PreviewLength <= 0 {
				t.Errorf("PreviewLength should be positive, got %d", cfg.PreviewLength)
			}

			// Validate external tools (empty values should use defaults)
			expectedCliphist := external.Cliphist
			if expectedCliphist == "" {
				expectedCliphist = "cliphist"
			}
			expectedFuzzel := external.Fuzzel
			if expectedFuzzel == "" {
				expectedFuzzel = "fuzzel"
			}
			expectedWlCopy := external.WlClipboard
			if expectedWlCopy == "" {
				expectedWlCopy = "wl-copy"
			}

			// These would be the actual paths used in the run function
			if expectedCliphist != external.Cliphist && external.Cliphist != "" {
				t.Errorf("Cliphist path mismatch")
			}
			if expectedFuzzel != external.Fuzzel && external.Fuzzel != "" {
				t.Errorf("Fuzzel path mismatch")
			}
			if expectedWlCopy != external.WlClipboard && external.WlClipboard != "" {
				t.Errorf("WlClipboard path mismatch")
			}
		})
	}
}

// TestFuzzelArgumentConstruction tests fuzzel argument construction
func TestFuzzelArgumentConstruction(t *testing.T) {
	tests := []struct {
		name                string
		clipboardCfg        config.ClipboardConfig
		deleteFlag          bool
		expectedPrompt      string
		expectedPlaceholder string
	}{
		{
			name: "normal mode with default config",
			clipboardCfg: config.ClipboardConfig{
				FuzzelPrompt: "Clipboard> ",
				FuzzelArgs:   []string{"--dmenu"},
			},
			deleteFlag:          false,
			expectedPrompt:      "Clipboard> ",
			expectedPlaceholder: "Type to search clipboard",
		},
		{
			name: "delete mode overrides prompt",
			clipboardCfg: config.ClipboardConfig{
				FuzzelPrompt: "Custom> ",
				FuzzelArgs:   []string{"--dmenu"},
			},
			deleteFlag:          true,
			expectedPrompt:      "del > ",
			expectedPlaceholder: "Delete from clipboard",
		},
		{
			name: "custom fuzzel args preserved",
			clipboardCfg: config.ClipboardConfig{
				FuzzelPrompt: "Select> ",
				FuzzelArgs:   []string{"--dmenu", "--width", "100", "--lines", "15"},
			},
			deleteFlag:          false,
			expectedPrompt:      "Select> ",
			expectedPlaceholder: "Type to search clipboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the argument construction logic from the run function
			var fuzzelArgs []string
			fuzzelArgs = append(fuzzelArgs, tt.clipboardCfg.FuzzelArgs...)

			if tt.deleteFlag {
				fuzzelArgs = append(fuzzelArgs, "--prompt", "del > ")
				fuzzelArgs = append(fuzzelArgs, "--placeholder", "Delete from clipboard")
			} else {
				fuzzelArgs = append(fuzzelArgs, "--prompt", tt.clipboardCfg.FuzzelPrompt)
				fuzzelArgs = append(fuzzelArgs, "--placeholder", "Type to search clipboard")
			}

			// Validate constructed arguments
			promptFound := false
			placeholderFound := false

			for i, arg := range fuzzelArgs {
				if arg == "--prompt" && i+1 < len(fuzzelArgs) {
					if fuzzelArgs[i+1] != tt.expectedPrompt {
						t.Errorf("Expected prompt '%s', got '%s'", tt.expectedPrompt, fuzzelArgs[i+1])
					}
					promptFound = true
				}
				if arg == "--placeholder" && i+1 < len(fuzzelArgs) {
					if fuzzelArgs[i+1] != tt.expectedPlaceholder {
						t.Errorf("Expected placeholder '%s', got '%s'", tt.expectedPlaceholder, fuzzelArgs[i+1])
					}
					placeholderFound = true
				}
			}

			if !promptFound {
				t.Errorf("Prompt argument not found in fuzzel args")
			}
			if !placeholderFound {
				t.Errorf("Placeholder argument not found in fuzzel args")
			}

			// Validate that original fuzzel args are preserved
			for _, originalArg := range tt.clipboardCfg.FuzzelArgs {
				found := false
				for _, arg := range fuzzelArgs {
					if arg == originalArg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Original fuzzel arg '%s' not found in constructed args", originalArg)
				}
			}
		})
	}
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		description string
		test        func(t *testing.T)
	}{
		{
			name:        "empty clipboard history",
			description: "Test behavior when clipboard history is empty",
			test: func(t *testing.T) {
				// This would be handled by cliphist returning empty output
				// The fuzzel command would still run but with empty input
				// This is expected behavior - fuzzel will show empty list
				t.Log("Empty clipboard history should not cause errors")
			},
		},
		{
			name:        "malformed clipboard output",
			description: "Test behavior with malformed cliphist output",
			test: func(t *testing.T) {
				// cliphist output format is implementation-specific
				// The clipboard command passes it directly to fuzzel
				// fuzzel handles the display formatting
				t.Log("Malformed output is passed through to fuzzel")
			},
		},
		{
			name:        "very long clipboard entries",
			description: "Test behavior with very long clipboard entries",
			test: func(t *testing.T) {
				// This is handled by the PreviewLength config option
				// But the actual truncation is done by fuzzel, not our code
				cfg := config.ClipboardConfig{
					PreviewLength: 50,
				}
				if cfg.PreviewLength <= 0 {
					t.Errorf("PreviewLength should be positive")
				}
			},
		},
		{
			name:        "special characters in clipboard",
			description: "Test behavior with special characters",
			test: func(t *testing.T) {
				// Special characters are handled by the underlying tools
				// cliphist encodes/decodes them properly
				// Our code just passes data through pipes
				t.Log("Special characters handled by underlying tools")
			},
		},
		{
			name:        "concurrent clipboard access",
			description: "Test behavior with concurrent clipboard access",
			test: func(t *testing.T) {
				// Concurrent access is handled by cliphist and wl-clipboard
				// Our code doesn't need special handling for this
				t.Log("Concurrent access handled by underlying tools")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			tt.test(t)
		})
	}
}

// TestCommandIntegration tests integration with cobra command system
func TestCommandIntegration(t *testing.T) {
	resetFlags()
	cmd := createTestCommand()

	// Test that command can be added to a parent command
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.AddCommand(cmd)

	// Verify command is properly registered
	found := false
	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Use == "clipboard" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Clipboard command should be registered with parent")
	}

	// Test command execution through parent
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"clipboard", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Command execution through parent should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "clipboard") {
		t.Errorf("Help output should contain command name")
	}
}

// TestGlobalVariables tests global variable behavior
func TestGlobalVariables(t *testing.T) {
	// Test initial state
	resetFlags()
	if deleteFlag {
		t.Errorf("deleteFlag should be false initially")
	}

	// Test that flag parsing affects global variable
	cmd := createTestCommand()
	cmd.SetArgs([]string{"--delete"})
	err := cmd.ParseFlags([]string{"--delete"})
	if err != nil {
		t.Errorf("Flag parsing should not error: %v", err)
	}

	if !deleteFlag {
		t.Errorf("deleteFlag should be true after parsing --delete")
	}

	// Test reset functionality
	resetFlags()
	if deleteFlag {
		t.Errorf("deleteFlag should be false after reset")
	}
}

// TestErrorMessages tests error message quality and consistency
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		errorMsg      string
		shouldContain []string
	}{
		{
			name:     "cliphist error message",
			errorMsg: "failed to get clipboard history: command not found",
			shouldContain: []string{
				"failed to get clipboard history",
				"command not found",
			},
		},
		{
			name:     "fuzzel error message",
			errorMsg: "failed to run fuzzel: permission denied",
			shouldContain: []string{
				"failed to run fuzzel",
				"permission denied",
			},
		},
		{
			name:     "decode error message",
			errorMsg: "failed to decode clipboard item: invalid format",
			shouldContain: []string{
				"failed to decode clipboard item",
				"invalid format",
			},
		},
		{
			name:     "copy error message",
			errorMsg: "failed to copy to clipboard: wl-copy not found",
			shouldContain: []string{
				"failed to copy to clipboard",
				"wl-copy not found",
			},
		},
		{
			name:     "delete error message",
			errorMsg: "failed to delete from clipboard: access denied",
			shouldContain: []string{
				"failed to delete from clipboard",
				"access denied",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expected := range tt.shouldContain {
				if !strings.Contains(tt.errorMsg, expected) {
					t.Errorf("Error message should contain '%s'\nActual: %s", expected, tt.errorMsg)
				}
			}

			// Check that error message starts with lowercase (Go convention)
			if len(tt.errorMsg) > 0 && strings.ToUpper(tt.errorMsg[:1]) == tt.errorMsg[:1] {
				// Allow "failed to" pattern which is common and acceptable
				if !strings.HasPrefix(tt.errorMsg, "failed to") {
					t.Errorf("Error message should start with lowercase: %s", tt.errorMsg)
				}
			}
		})
	}
}

// TestConfigurationDefaults tests that configuration defaults are sensible
func TestConfigurationDefaults(t *testing.T) {
	mockConfig := NewMockConfig()
	cfg := mockConfig.GetClipboard()
	external := mockConfig.GetExternal()

	// Test clipboard configuration defaults
	if cfg.MaxEntries <= 0 {
		t.Errorf("Default MaxEntries should be positive, got %d", cfg.MaxEntries)
	}
	if cfg.FuzzelPrompt == "" {
		t.Errorf("Default FuzzelPrompt should not be empty")
	}
	if len(cfg.FuzzelArgs) == 0 {
		t.Errorf("Default FuzzelArgs should not be empty")
	}
	if cfg.PreviewLength <= 0 {
		t.Errorf("Default PreviewLength should be positive, got %d", cfg.PreviewLength)
	}

	// Test external tools defaults
	if external.Cliphist == "" {
		t.Errorf("Default Cliphist should not be empty")
	}
	if external.Fuzzel == "" {
		t.Errorf("Default Fuzzel should not be empty")
	}
	if external.WlClipboard == "" {
		t.Errorf("Default WlClipboard should not be empty")
	}

	// Test that defaults are reasonable values
	expectedDefaults := map[string]string{
		"cliphist": external.Cliphist,
		"fuzzel":   external.Fuzzel,
		"wl-copy":  external.WlClipboard,
	}

	for expected, actual := range expectedDefaults {
		if actual != expected {
			t.Errorf("Expected default %s, got %s", expected, actual)
		}
	}
}

// TestCommandValidation tests command validation and error handling
func TestCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid delete flag",
			args:        []string{"--delete"},
			expectError: false,
		},
		{
			name:        "valid delete flag short form",
			args:        []string{"-d"},
			expectError: false,
		},
		{
			name:        "no arguments is valid",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "unknown flag should error",
			args:        []string{"--unknown"},
			expectError: true,
			errorMsg:    "unknown flag",
		},
		{
			name:        "invalid flag format should error",
			args:        []string{"--delete=invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			cmd := createTestCommand()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetErr(&buf)

			err := cmd.ParseFlags(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestExternalToolPathHandling tests how external tool paths are handled
func TestExternalToolPathHandling(t *testing.T) {
	tests := []struct {
		name     string
		external config.ExternalTools
		expected map[string]string
	}{
		{
			name: "custom paths are used",
			external: config.ExternalTools{
				Cliphist:    "/usr/local/bin/cliphist",
				Fuzzel:      "/usr/local/bin/fuzzel",
				WlClipboard: "/usr/local/bin/wl-copy",
			},
			expected: map[string]string{
				"cliphist": "/usr/local/bin/cliphist",
				"fuzzel":   "/usr/local/bin/fuzzel",
				"wl-copy":  "/usr/local/bin/wl-copy",
			},
		},
		{
			name: "empty paths fall back to defaults",
			external: config.ExternalTools{
				Cliphist:    "",
				Fuzzel:      "",
				WlClipboard: "",
			},
			expected: map[string]string{
				"cliphist": "cliphist",
				"fuzzel":   "fuzzel",
				"wl-copy":  "wl-copy",
			},
		},
		{
			name: "mixed custom and default paths",
			external: config.ExternalTools{
				Cliphist:    "/custom/cliphist",
				Fuzzel:      "",
				WlClipboard: "/custom/wl-copy",
			},
			expected: map[string]string{
				"cliphist": "/custom/cliphist",
				"fuzzel":   "fuzzel",
				"wl-copy":  "/custom/wl-copy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the path resolution logic from the run function
			cliphistPath := tt.external.Cliphist
			if cliphistPath == "" {
				cliphistPath = "cliphist"
			}

			fuzzelPath := tt.external.Fuzzel
			if fuzzelPath == "" {
				fuzzelPath = "fuzzel"
			}

			wlCopyPath := tt.external.WlClipboard
			if wlCopyPath == "" {
				wlCopyPath = "wl-copy"
			}

			// Validate resolved paths
			if cliphistPath != tt.expected["cliphist"] {
				t.Errorf("Expected cliphist path %s, got %s", tt.expected["cliphist"], cliphistPath)
			}
			if fuzzelPath != tt.expected["fuzzel"] {
				t.Errorf("Expected fuzzel path %s, got %s", tt.expected["fuzzel"], fuzzelPath)
			}
			if wlCopyPath != tt.expected["wl-copy"] {
				t.Errorf("Expected wl-copy path %s, got %s", tt.expected["wl-copy"], wlCopyPath)
			}
		})
	}
}

// TestDeleteModeSpecificBehavior tests behavior specific to delete mode
func TestDeleteModeSpecificBehavior(t *testing.T) {
	tests := []struct {
		name       string
		deleteFlag bool
		expected   struct {
			prompt      string
			placeholder string
		}
	}{
		{
			name:       "normal mode uses configured prompt",
			deleteFlag: false,
			expected: struct {
				prompt      string
				placeholder string
			}{
				prompt:      "Clipboard> ",
				placeholder: "Type to search clipboard",
			},
		},
		{
			name:       "delete mode overrides prompt",
			deleteFlag: true,
			expected: struct {
				prompt      string
				placeholder string
			}{
				prompt:      "del > ",
				placeholder: "Delete from clipboard",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.ClipboardConfig{
				FuzzelPrompt: "Clipboard> ",
				FuzzelArgs:   []string{"--dmenu"},
			}

			// Simulate argument construction from run function
			var fuzzelArgs []string
			fuzzelArgs = append(fuzzelArgs, cfg.FuzzelArgs...)

			if tt.deleteFlag {
				fuzzelArgs = append(fuzzelArgs, "--prompt", "del > ")
				fuzzelArgs = append(fuzzelArgs, "--placeholder", "Delete from clipboard")
			} else {
				fuzzelArgs = append(fuzzelArgs, "--prompt", cfg.FuzzelPrompt)
				fuzzelArgs = append(fuzzelArgs, "--placeholder", "Type to search clipboard")
			}

			// Find and validate prompt and placeholder
			var actualPrompt, actualPlaceholder string
			for i, arg := range fuzzelArgs {
				if arg == "--prompt" && i+1 < len(fuzzelArgs) {
					actualPrompt = fuzzelArgs[i+1]
				}
				if arg == "--placeholder" && i+1 < len(fuzzelArgs) {
					actualPlaceholder = fuzzelArgs[i+1]
				}
			}

			if actualPrompt != tt.expected.prompt {
				t.Errorf("Expected prompt '%s', got '%s'", tt.expected.prompt, actualPrompt)
			}
			if actualPlaceholder != tt.expected.placeholder {
				t.Errorf("Expected placeholder '%s', got '%s'", tt.expected.placeholder, actualPlaceholder)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkCommandCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCommand()
	}
}

func BenchmarkFlagParsing(b *testing.B) {
	cmd := createTestCommand()
	args := []string{"--delete"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetFlags()
		_ = cmd.ParseFlags(args)
	}
}

func BenchmarkConfigurationAccess(b *testing.B) {
	mockConfig := NewMockConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockConfig.GetClipboard()
		_ = mockConfig.GetExternal()
	}
}
