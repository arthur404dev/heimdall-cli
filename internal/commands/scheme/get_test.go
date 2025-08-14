package scheme

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
)

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "creates get command with correct properties",
			test: testGetCommandCreation,
		},
		{
			name: "displays current scheme info by default",
			test: testGetCommandDefault,
		},
		{
			name: "returns scheme name with -n flag",
			test: testGetCommandNameFlag,
		},
		{
			name: "returns flavour with -f flag",
			test: testGetCommandFlavourFlag,
		},
		{
			name: "returns mode with -m flag",
			test: testGetCommandModeFlag,
		},
		{
			name: "returns variant with -v flag",
			test: testGetCommandVariantFlag,
		},
		{
			name: "returns JSON output with --json flag",
			test: testGetCommandJSONFlag,
		},
		{
			name: "returns specific property by argument",
			test: testGetCommandPropertyArg,
		},
		{
			name: "returns specific color value",
			test: testGetCommandColorValue,
		},
		{
			name: "returns all colors as JSON",
			test: testGetCommandColorsJSON,
		},
		{
			name: "handles unknown property error",
			test: testGetCommandUnknownProperty,
		},
		{
			name: "handles manager error",
			test: testGetCommandManagerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testGetCommandCreation(t *testing.T) {
	// Act
	cmd := getCommand()

	// Assert
	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}
	if cmd.Use != "get [property]" {
		t.Errorf("Expected Use to be 'get [property]', got '%s'", cmd.Use)
	}
	if cmd.Short != "Get current scheme or specific property" {
		t.Errorf("Expected Short to be 'Get current scheme or specific property', got '%s'", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "Get the current color scheme") {
		t.Error("Expected Long description to contain 'Get the current color scheme'")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Check flags
	flags := cmd.Flags()
	if flags.Lookup("json") == nil {
		t.Error("Expected --json flag to be present")
	}
	if flags.Lookup("name") == nil {
		t.Error("Expected --name/-n flag to be present")
	}
	if flags.Lookup("flavour") == nil {
		t.Error("Expected --flavour/-f flag to be present")
	}
	if flags.Lookup("mode") == nil {
		t.Error("Expected --mode/-m flag to be present")
	}
	if flags.Lookup("variant") == nil {
		t.Error("Expected --variant/-v flag to be present")
	}
	if flags.Lookup("no-color") == nil {
		t.Error("Expected --no-color flag to be present")
	}
}

func testGetCommandDefault(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Variant: "tonalspot",
		Colours: map[string]string{
			"base": "1e1e2e",
			"text": "cdd6f4",
		},
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "catppuccin") {
		t.Error("Expected output to contain scheme name")
	}
	if !strings.Contains(outputStr, "mocha") {
		t.Error("Expected output to contain flavour")
	}
	if !strings.Contains(outputStr, "dark") {
		t.Error("Expected output to contain mode")
	}
}

func testGetCommandNameFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "gruvbox",
		Flavour: "dark",
		Mode:    "dark",
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"-n"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())
	if outputStr != "gruvbox" {
		t.Errorf("Expected output to be 'gruvbox', got '%s'", outputStr)
	}
}

func testGetCommandFlavourFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "latte",
		Mode:    "light",
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"-f"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())
	if outputStr != "latte" {
		t.Errorf("Expected output to be 'latte', got '%s'", outputStr)
	}
}

func testGetCommandModeFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "rosepine",
		Flavour: "dawn",
		Mode:    "light",
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"-m"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())
	if outputStr != "light" {
		t.Errorf("Expected output to be 'light', got '%s'", outputStr)
	}
}

func testGetCommandVariantFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "dynamic",
		Flavour: "default",
		Mode:    "dark",
		Variant: "vibrant",
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"-v"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())
	if outputStr != "vibrant" {
		t.Errorf("Expected output to be 'vibrant', got '%s'", outputStr)
	}
}

func testGetCommandJSONFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Variant: "tonalspot",
		Colours: map[string]string{
			"base": "1e1e2e",
			"text": "cdd6f4",
		},
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"--json"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(outputStr), &result); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	// Verify JSON contains expected fields
	if result["name"] != "catppuccin" {
		t.Errorf("Expected JSON to contain name 'catppuccin', got %v", result["name"])
	}
	if result["flavour"] != "mocha" {
		t.Errorf("Expected JSON to contain flavour 'mocha', got %v", result["flavour"])
	}
}

func testGetCommandPropertyArg(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "gruvbox",
		Flavour: "hard",
		Mode:    "dark",
		Variant: "tonalspot",
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"name"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())
	if outputStr != "gruvbox" {
		t.Errorf("Expected output to be 'gruvbox', got '%s'", outputStr)
	}
}

func testGetCommandColorValue(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Colours: map[string]string{
			"base":    "1e1e2e",
			"text":    "cdd6f4",
			"primary": "89b4fa",
		},
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"base"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "1e1e2e") {
		t.Error("Expected output to contain base color value")
	}
}

func testGetCommandColorsJSON(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Colours: map[string]string{
			"base": "1e1e2e",
			"text": "cdd6f4",
		},
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"colors", "--json"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()

	// Verify it's valid JSON
	var colors map[string]string
	if err := json.Unmarshal([]byte(outputStr), &colors); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	// Verify colors are present
	if colors["base"] != "1e1e2e" {
		t.Errorf("Expected base color '1e1e2e', got '%s'", colors["base"])
	}
	if colors["text"] != "cdd6f4" {
		t.Errorf("Expected text color 'cdd6f4', got '%s'", colors["text"])
	}
}

func testGetCommandUnknownProperty(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Colours: map[string]string{
			"base": "1e1e2e",
		},
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"unknown-property"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error for unknown property")
	}
	if !strings.Contains(err.Error(), "unknown property or color") {
		t.Errorf("Expected error to contain 'unknown property or color', got: %v", err)
	}
}

func testGetCommandManagerError(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.SetError("GetCurrent", &mockError{msg: "failed to get current scheme"})

	cmd := getCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error from manager")
	}
	if !strings.Contains(err.Error(), "failed to get current scheme") {
		t.Errorf("Expected error to contain manager error message, got: %v", err)
	}
}

// SchemeManagerInterface defines the interface for scheme managers
type SchemeManagerInterface interface {
	GetCurrent() (*scheme.Scheme, error)
	SetScheme(*scheme.Scheme) error
	ListSchemes() ([]string, error)
	ListFlavours(string) ([]string, error)
	ListModes(string, string) ([]string, error)
	LoadScheme(string, string, string) (*scheme.Scheme, error)
	LoadSchemeWithFallback(string, string, string) (*scheme.Scheme, error)
	SaveScheme(*scheme.Scheme) error
	InstallBundledScheme(string) error
	InstallAllBundledSchemes() error
}

// Variable to allow mocking the scheme.NewManager function
var newManagerFunc = func() SchemeManagerInterface {
	return scheme.NewManager()
}

func TestGetCommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedFlag string
	}{
		{
			name:         "short name flag",
			args:         []string{"-n"},
			expectedFlag: "name",
		},
		{
			name:         "long name flag",
			args:         []string{"--name"},
			expectedFlag: "name",
		},
		{
			name:         "short flavour flag",
			args:         []string{"-f"},
			expectedFlag: "flavour",
		},
		{
			name:         "long flavour flag",
			args:         []string{"--flavour"},
			expectedFlag: "flavour",
		},
		{
			name:         "short mode flag",
			args:         []string{"-m"},
			expectedFlag: "mode",
		},
		{
			name:         "long mode flag",
			args:         []string{"--mode"},
			expectedFlag: "mode",
		},
		{
			name:         "short variant flag",
			args:         []string{"-v"},
			expectedFlag: "variant",
		},
		{
			name:         "long variant flag",
			args:         []string{"--variant"},
			expectedFlag: "variant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockManager := NewMockSchemeManager()
			testScheme := &scheme.Scheme{
				Name:    "test-scheme",
				Flavour: "test-flavour",
				Mode:    "dark",
				Variant: "test-variant",
			}
			mockManager.SetCurrentScheme(testScheme)

			cmd := getCommand()
			cmd.SetArgs(tt.args)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Mock the scheme.NewManager function
			originalNewManager := newManagerFunc
			defer func() { newManagerFunc = originalNewManager }()
			newManagerFunc = func() SchemeManagerInterface {
				return mockManager
			}

			// Act
			err := cmd.Execute()

			// Assert
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			outputStr := strings.TrimSpace(output.String())

			// Verify the correct value is returned based on the flag
			switch tt.expectedFlag {
			case "name":
				if outputStr != "test-scheme" {
					t.Errorf("Expected 'test-scheme', got '%s'", outputStr)
				}
			case "flavour":
				if outputStr != "test-flavour" {
					t.Errorf("Expected 'test-flavour', got '%s'", outputStr)
				}
			case "mode":
				if outputStr != "dark" {
					t.Errorf("Expected 'dark', got '%s'", outputStr)
				}
			case "variant":
				if outputStr != "test-variant" {
					t.Errorf("Expected 'test-variant', got '%s'", outputStr)
				}
			}
		})
	}
}

func TestGetCommandNoColorFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Colours: map[string]string{
			"base": "1e1e2e",
			"text": "cdd6f4",
		},
	}
	mockManager.SetCurrentScheme(testScheme)

	cmd := getCommand()
	cmd.SetArgs([]string{"base", "--no-color"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())
	if outputStr != "1e1e2e" {
		t.Errorf("Expected plain color value '1e1e2e', got '%s'", outputStr)
	}
}
