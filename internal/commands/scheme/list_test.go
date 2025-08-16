package scheme

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "creates list command with correct properties",
			test: testListCommandCreation,
		},
		{
			name: "lists scheme names with -n flag",
			test: testListCommandNamesFlag,
		},
		{
			name: "lists flavours with -f flag",
			test: testListCommandFlavoursFlag,
		},
		{
			name: "lists modes with -m flag",
			test: testListCommandModesFlag,
		},
		{
			name: "lists Material You variants with -v flag",
			test: testListCommandVariantsFlag,
		},
		{
			name: "outputs Caelestia format by default",
			test: testListCommandCaelestiaFormat,
		},
		{
			name: "lists flavours for specific scheme",
			test: testListCommandSpecificScheme,
		},
		{
			name: "lists modes for specific scheme and flavour",
			test: testListCommandSpecificSchemeFlavour,
		},
		{
			name: "handles manager errors",
			test: testListCommandManagerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testListCommandCreation(t *testing.T) {
	// Act
	cmd := listCommand()

	// Assert
	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}
	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got '%s'", cmd.Use)
	}
	if cmd.Short != "List available schemes, flavours, or modes" {
		t.Errorf("Expected Short to be 'List available schemes, flavours, or modes', got '%s'", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "List available color schemes") {
		t.Error("Expected Long description to contain 'List available color schemes'")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Check flags
	flags := cmd.Flags()
	if flags.Lookup("scheme") == nil {
		t.Error("Expected --scheme/-s flag to be present")
	}
	if flags.Lookup("flavour") == nil {
		t.Error("Expected --flavour flag to be present")
	}
	if flags.Lookup("names") == nil {
		t.Error("Expected --names/-n flag to be present")
	}
	if flags.Lookup("flavours") == nil {
		t.Error("Expected --flavours/-f flag to be present")
	}
	if flags.Lookup("modes") == nil {
		t.Error("Expected --modes/-m flag to be present")
	}
	if flags.Lookup("variants") == nil {
		t.Error("Expected --variants/-v flag to be present")
	}
}

func testListCommandNamesFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddScheme("catppuccin", []string{"mocha", "latte"})
	mockManager.AddScheme("gruvbox", []string{"dark", "light"})

	cmd := listCommand()
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

	outputStr := output.String()
	if !strings.Contains(outputStr, "catppuccin") {
		t.Error("Expected output to contain 'catppuccin'")
	}
	if !strings.Contains(outputStr, "gruvbox") {
		t.Error("Expected output to contain 'gruvbox'")
	}
}

func testListCommandFlavoursFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	currentScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
	}
	mockManager.SetCurrentScheme(currentScheme)
	mockManager.AddScheme("catppuccin", []string{"mocha", "latte", "frappe", "macchiato"})

	cmd := listCommand()
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

	outputStr := output.String()
	flavours := []string{"mocha", "latte", "frappe", "macchiato"}
	for _, flavour := range flavours {
		if !strings.Contains(outputStr, flavour) {
			t.Errorf("Expected output to contain flavour '%s'", flavour)
		}
	}
}

func testListCommandModesFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	currentScheme := &scheme.Scheme{
		Name:    "gruvbox",
		Flavour: "hard",
		Mode:    "dark",
	}
	mockManager.SetCurrentScheme(currentScheme)
	mockManager.AddFlavour("gruvbox", "hard", []string{"dark", "light"})

	cmd := listCommand()
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

	outputStr := output.String()
	if !strings.Contains(outputStr, "dark") {
		t.Error("Expected output to contain 'dark'")
	}
	if !strings.Contains(outputStr, "light") {
		t.Error("Expected output to contain 'light'")
	}
}

func testListCommandVariantsFlag(t *testing.T) {
	// Arrange
	cmd := listCommand()
	cmd.SetArgs([]string{"-v"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()
	expectedVariants := []string{
		"tonalspot", "neutral", "vibrant", "expressive",
		"rainbow", "fruitsalad", "content", "monochrome",
	}
	for _, variant := range expectedVariants {
		if !strings.Contains(outputStr, variant) {
			t.Errorf("Expected output to contain variant '%s'", variant)
		}
	}
}

func testListCommandCaelestiaFormat(t *testing.T) {
	// Arrange
	cmd := listCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the listJSONFormat function
	originalListJSONFormat := listJSONFormatFunc
	defer func() { listJSONFormatFunc = originalListJSONFormat }()
	listJSONFormatFunc = func() error {
		// Return a simple JSON structure for testing
		testOutput := map[string]map[string]map[string]string{
			"catppuccin": {
				"mocha": {
					"base": "1e1e2e",
					"text": "cdd6f4",
				},
			},
		}
		jsonData, _ := json.Marshal(testOutput)
		cmd.OutOrStdout().Write(jsonData)
		return nil
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

	// Verify JSON structure
	if _, exists := result["catppuccin"]; !exists {
		t.Error("Expected JSON to contain 'catppuccin' scheme")
	}
}

func testListCommandSpecificScheme(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddScheme("rosepine", []string{"main", "moon", "dawn"})

	cmd := listCommand()
	cmd.SetArgs([]string{"-s", "rosepine"})
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
	if !strings.Contains(outputStr, "Available flavours for rosepine:") {
		t.Error("Expected output to contain flavours header")
	}
	flavours := []string{"main", "moon", "dawn"}
	for _, flavour := range flavours {
		if !strings.Contains(outputStr, flavour) {
			t.Errorf("Expected output to contain flavour '%s'", flavour)
		}
	}
}

func testListCommandSpecificSchemeFlavour(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddFlavour("catppuccin", "mocha", []string{"dark"})

	cmd := listCommand()
	cmd.SetArgs([]string{"-s", "catppuccin", "--flavour", "mocha"})
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
	if !strings.Contains(outputStr, "Available modes for catppuccin/mocha:") {
		t.Error("Expected output to contain modes header")
	}
	if !strings.Contains(outputStr, "dark") {
		t.Error("Expected output to contain 'dark' mode")
	}
}

func testListCommandManagerError(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.SetError("ListSchemes", &mockError{msg: "failed to list schemes"})

	cmd := listCommand()
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
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list schemes") {
		t.Errorf("Expected error to contain manager error message, got: %v", err)
	}
}

// Variable to allow mocking the listJSONFormat function
var listJSONFormatFunc = listJSONFormat

func TestListCommandIntegration(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupManager   func(*MockSchemeManager)
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "list names only",
			args: []string{"--names"},
			setupManager: func(m *MockSchemeManager) {
				m.AddScheme("scheme1", []string{"flavour1"})
				m.AddScheme("scheme2", []string{"flavour2"})
			},
			expectedOutput: []string{"scheme1", "scheme2"},
		},
		{
			name: "list flavours for current scheme",
			args: []string{"--flavours"},
			setupManager: func(m *MockSchemeManager) {
				m.SetCurrentScheme(&scheme.Scheme{
					Name:    "test-scheme",
					Flavour: "test-flavour",
					Mode:    "dark",
				})
				m.AddScheme("test-scheme", []string{"flavour1", "flavour2", "flavour3"})
			},
			expectedOutput: []string{"flavour1", "flavour2", "flavour3"},
		},
		{
			name: "list modes for current scheme and flavour",
			args: []string{"--modes"},
			setupManager: func(m *MockSchemeManager) {
				m.SetCurrentScheme(&scheme.Scheme{
					Name:    "test-scheme",
					Flavour: "test-flavour",
					Mode:    "dark",
				})
				m.AddFlavour("test-scheme", "test-flavour", []string{"dark", "light"})
			},
			expectedOutput: []string{"dark", "light"},
		},
		{
			name: "list flavours for specific scheme",
			args: []string{"--scheme", "specific-scheme"},
			setupManager: func(m *MockSchemeManager) {
				m.AddScheme("specific-scheme", []string{"specific-flavour1", "specific-flavour2"})
			},
			expectedOutput: []string{"Available flavours for specific-scheme:", "specific-flavour1", "specific-flavour2"},
		},
		{
			name: "list modes for specific scheme and flavour",
			args: []string{"--scheme", "specific-scheme", "--flavour", "specific-flavour"},
			setupManager: func(m *MockSchemeManager) {
				m.AddFlavour("specific-scheme", "specific-flavour", []string{"mode1", "mode2"})
			},
			expectedOutput: []string{"Available modes for specific-scheme/specific-flavour:", "mode1", "mode2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockManager := NewMockSchemeManager()
			if tt.setupManager != nil {
				tt.setupManager(mockManager)
			}

			cmd := listCommand()
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
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't. Output:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestListCommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		shortFlag    string
		expectedType string
	}{
		{
			name:         "scheme flag",
			flagName:     "scheme",
			shortFlag:    "s",
			expectedType: "string",
		},
		{
			name:         "flavour flag",
			flagName:     "flavour",
			shortFlag:    "",
			expectedType: "string",
		},
		{
			name:         "names flag",
			flagName:     "names",
			shortFlag:    "n",
			expectedType: "bool",
		},
		{
			name:         "flavours flag",
			flagName:     "flavours",
			shortFlag:    "f",
			expectedType: "bool",
		},
		{
			name:         "modes flag",
			flagName:     "modes",
			shortFlag:    "m",
			expectedType: "bool",
		},
		{
			name:         "variants flag",
			flagName:     "variants",
			shortFlag:    "v",
			expectedType: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cmd := listCommand()
			flags := cmd.Flags()

			// Act
			flag := flags.Lookup(tt.flagName)

			// Assert
			if flag == nil {
				t.Errorf("Expected flag '%s' to be present", tt.flagName)
				return
			}

			if tt.shortFlag != "" && flag.Shorthand != tt.shortFlag {
				t.Errorf("Expected flag '%s' to have shorthand '%s', got '%s'", tt.flagName, tt.shortFlag, flag.Shorthand)
			}

			if flag.Value.Type() != tt.expectedType {
				t.Errorf("Expected flag '%s' to be of type '%s', got '%s'", tt.flagName, tt.expectedType, flag.Value.Type())
			}
		})
	}
}

func TestListMaterialYouVariants(t *testing.T) {
	// Act
	cmd := listCommand()
	cmd.SetArgs([]string{"--variants"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()
	expectedVariants := []string{
		"tonalspot",
		"neutral",
		"vibrant",
		"expressive",
		"rainbow",
		"fruitsalad",
		"content",
		"monochrome",
	}

	for _, variant := range expectedVariants {
		if !strings.Contains(outputStr, variant) {
			t.Errorf("Expected output to contain Material You variant '%s'", variant)
		}
	}
}
