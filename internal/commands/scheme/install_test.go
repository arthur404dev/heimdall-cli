package scheme

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
)

func TestInstallCommand(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "creates install command with correct properties",
			test: testInstallCommandCreation,
		},
		{
			name: "installs all bundled schemes with --all flag",
			test: testInstallCommandAllFlag,
		},
		{
			name: "installs specific scheme by name",
			test: testInstallCommandSpecificScheme,
		},
		{
			name: "lists available schemes when no args provided",
			test: testInstallCommandNoArgs,
		},
		{
			name: "handles scheme installation error",
			test: testInstallCommandInstallError,
		},
		{
			name: "handles install all error",
			test: testInstallCommandInstallAllError,
		},
		{
			name: "handles list bundled schemes error",
			test: testInstallCommandListError,
		},
		{
			name: "installs scheme with spaces in name",
			test: testInstallCommandSchemeWithSpaces,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testInstallCommandCreation(t *testing.T) {
	// Act
	cmd := installCommand()

	// Assert
	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}
	if cmd.Use != "install [scheme-name]" {
		t.Errorf("Expected Use to be 'install [scheme-name]', got '%s'", cmd.Use)
	}
	if cmd.Short != "Install bundled color schemes" {
		t.Errorf("Expected Short to be 'Install bundled color schemes', got '%s'", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "Install bundled color schemes") {
		t.Error("Expected Long description to contain 'Install bundled color schemes'")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Check flags
	flags := cmd.Flags()
	if flags.Lookup("all") == nil {
		t.Error("Expected --all flag to be present")
	}
}

func testInstallCommandAllFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()

	cmd := installCommand()
	cmd.SetArgs([]string{"--all"})
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

	if !mockManager.GetInstallAllCalled() {
		t.Error("Expected InstallAllBundledSchemes to be called")
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Installing all bundled schemes") {
		t.Error("Expected output to contain 'Installing all bundled schemes'")
	}
	if !strings.Contains(outputStr, "Successfully installed all bundled schemes") {
		t.Error("Expected output to contain success message")
	}
}

func testInstallCommandSpecificScheme(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddBundledScheme(scheme.BundledScheme{
		Name:    "Catppuccin Mocha",
		Family:  "catppuccin",
		Flavour: "mocha",
	})

	cmd := installCommand()
	cmd.SetArgs([]string{"Catppuccin", "Mocha"})
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

	installedSchemes := mockManager.GetInstallBundledCalled()
	if len(installedSchemes) != 1 {
		t.Errorf("Expected 1 scheme to be installed, got %d", len(installedSchemes))
	}
	if installedSchemes[0] != "Catppuccin Mocha" {
		t.Errorf("Expected 'Catppuccin Mocha' to be installed, got '%s'", installedSchemes[0])
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Installing scheme: Catppuccin Mocha") {
		t.Error("Expected output to contain installation message")
	}
	if !strings.Contains(outputStr, "Successfully installed Catppuccin Mocha") {
		t.Error("Expected output to contain success message")
	}
}

func testInstallCommandNoArgs(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()

	cmd := installCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Mock the ListBundledSchemeNames function
	originalListBundledSchemeNames := listBundledSchemeNamesFunc
	defer func() { listBundledSchemeNamesFunc = originalListBundledSchemeNames }()
	listBundledSchemeNamesFunc = func() ([]string, error) {
		return []string{"Catppuccin Mocha", "Gruvbox Dark", "Rosé Pine Main"}, nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Available bundled schemes:") {
		t.Error("Expected output to contain 'Available bundled schemes:'")
	}
	if !strings.Contains(outputStr, "Catppuccin Mocha") {
		t.Error("Expected output to contain 'Catppuccin Mocha'")
	}
	if !strings.Contains(outputStr, "Gruvbox Dark") {
		t.Error("Expected output to contain 'Gruvbox Dark'")
	}
	if !strings.Contains(outputStr, "To install a scheme, run:") {
		t.Error("Expected output to contain usage instructions")
	}
}

func testInstallCommandInstallError(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.SetError("InstallBundledScheme", &mockError{msg: "scheme not found"})

	cmd := installCommand()
	cmd.SetArgs([]string{"NonExistent", "Scheme"})
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
	if !strings.Contains(err.Error(), "failed to install scheme") {
		t.Errorf("Expected error to contain 'failed to install scheme', got: %v", err)
	}
	if !strings.Contains(err.Error(), "scheme not found") {
		t.Errorf("Expected error to contain underlying error message, got: %v", err)
	}
}

func testInstallCommandInstallAllError(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.SetError("InstallAllBundledSchemes", &mockError{msg: "installation failed"})

	cmd := installCommand()
	cmd.SetArgs([]string{"--all"})
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
	if !strings.Contains(err.Error(), "failed to install bundled schemes") {
		t.Errorf("Expected error to contain 'failed to install bundled schemes', got: %v", err)
	}
	if !strings.Contains(err.Error(), "installation failed") {
		t.Errorf("Expected error to contain underlying error message, got: %v", err)
	}
}

func testInstallCommandListError(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()

	cmd := installCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock the scheme.NewManager function
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Mock the ListBundledSchemeNames function to return an error
	originalListBundledSchemeNames := listBundledSchemeNamesFunc
	defer func() { listBundledSchemeNamesFunc = originalListBundledSchemeNames }()
	listBundledSchemeNamesFunc = func() ([]string, error) {
		return nil, &mockError{msg: "failed to list schemes"}
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list bundled schemes") {
		t.Errorf("Expected error to contain 'failed to list bundled schemes', got: %v", err)
	}
}

func testInstallCommandSchemeWithSpaces(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddBundledScheme(scheme.BundledScheme{
		Name:    "Rosé Pine Dawn",
		Family:  "rosepine",
		Flavour: "dawn",
	})

	cmd := installCommand()
	cmd.SetArgs([]string{"Rosé", "Pine", "Dawn"})
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

	installedSchemes := mockManager.GetInstallBundledCalled()
	if len(installedSchemes) != 1 {
		t.Errorf("Expected 1 scheme to be installed, got %d", len(installedSchemes))
	}
	if installedSchemes[0] != "Rosé Pine Dawn" {
		t.Errorf("Expected 'Rosé Pine Dawn' to be installed, got '%s'", installedSchemes[0])
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Installing scheme: Rosé Pine Dawn") {
		t.Error("Expected output to contain installation message with full name")
	}
}

// Variable to allow mocking the ListBundledSchemeNames function
var listBundledSchemeNamesFunc = scheme.ListBundledSchemeNames

func TestInstallCommandIntegration(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		availableSchemes []string
		expectError      bool
		expectedOutput   []string
	}{
		{
			name:             "install single scheme",
			args:             []string{"Test", "Scheme"},
			availableSchemes: []string{"Test Scheme"},
			expectedOutput:   []string{"Installing scheme: Test Scheme", "Successfully installed Test Scheme"},
		},
		{
			name:             "install all schemes",
			args:             []string{"--all"},
			availableSchemes: []string{"Scheme1", "Scheme2"},
			expectedOutput:   []string{"Installing all bundled schemes", "Successfully installed all bundled schemes"},
		},
		{
			name:             "list available schemes",
			args:             []string{},
			availableSchemes: []string{"Available Scheme 1", "Available Scheme 2"},
			expectedOutput:   []string{"Available bundled schemes:", "Available Scheme 1", "Available Scheme 2"},
		},
		{
			name:             "no available schemes",
			args:             []string{},
			availableSchemes: []string{},
			expectedOutput:   []string{"No bundled schemes available"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockManager := NewMockSchemeManager()
			for _, schemeName := range tt.availableSchemes {
				mockManager.AddBundledScheme(scheme.BundledScheme{
					Name: schemeName,
				})
			}

			cmd := installCommand()
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

			// Mock the ListBundledSchemeNames function
			originalListBundledSchemeNames := listBundledSchemeNamesFunc
			defer func() { listBundledSchemeNamesFunc = originalListBundledSchemeNames }()
			listBundledSchemeNamesFunc = func() ([]string, error) {
				if tt.expectError {
					return nil, &mockError{msg: "test error"}
				}
				return tt.availableSchemes, nil
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

func TestInstallCommandFlags(t *testing.T) {
	// Act
	cmd := installCommand()

	// Assert
	flags := cmd.Flags()

	// Check --all flag
	allFlag := flags.Lookup("all")
	if allFlag == nil {
		t.Error("Expected --all flag to be present")
	}
	if allFlag.Usage != "Install all bundled schemes" {
		t.Errorf("Expected --all flag usage to be 'Install all bundled schemes', got '%s'", allFlag.Usage)
	}
}

func TestInstallCommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "single argument",
			args:        []string{"scheme"},
			expectError: false,
		},
		{
			name:        "multiple arguments",
			args:        []string{"scheme", "name", "with", "spaces"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cmd := installCommand()

			// Act
			var err error
			if cmd.Args != nil {
				err = cmd.Args(cmd, tt.args)
			}

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
