package scheme

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommand(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "creates scheme command with correct properties",
			test: testSchemeCommandCreation,
		},
		{
			name: "registers all required subcommands",
			test: testSchemeSubcommandRegistration,
		},
		{
			name: "has correct command metadata",
			test: testSchemeCommandMetadata,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testSchemeCommandCreation(t *testing.T) {
	// Act
	cmd := Command()

	// Assert
	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}
	if cmd.Use != "scheme" {
		t.Errorf("Expected Use to be 'scheme', got '%s'", cmd.Use)
	}
	if cmd.Short != "Manage color schemes" {
		t.Errorf("Expected Short to be 'Manage color schemes', got '%s'", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "Manage color schemes for theming") {
		t.Error("Expected Long description to contain 'Manage color schemes for theming'")
	}
	if !strings.Contains(cmd.Long, "Available subcommands:") {
		t.Error("Expected Long description to contain 'Available subcommands:'")
	}
}

func testSchemeSubcommandRegistration(t *testing.T) {
	// Act
	cmd := Command()

	// Assert
	subcommands := cmd.Commands()
	if len(subcommands) != 5 {
		t.Errorf("Expected 5 subcommands, got %d", len(subcommands))
	}

	// Check that all expected subcommands are present
	expectedSubcommands := []string{"list", "get", "set", "install", "bundled"}
	actualSubcommands := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		// Extract just the command name from the Use field (e.g., "get [property]" -> "get")
		parts := strings.Fields(subcmd.Use)
		if len(parts) > 0 {
			actualSubcommands[i] = parts[0]
		} else {
			actualSubcommands[i] = subcmd.Use
		}
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
			t.Errorf("Missing subcommand: %s", expected)
		}
	}
}

func testSchemeCommandMetadata(t *testing.T) {
	// Act
	cmd := Command()

	// Assert
	if cmd.Use != "scheme" {
		t.Errorf("Expected Use to be 'scheme', got '%s'", cmd.Use)
	}
	if cmd.Short != "Manage color schemes" {
		t.Errorf("Expected Short to be 'Manage color schemes', got '%s'", cmd.Short)
	}

	// Check that long description contains all expected subcommands
	longDesc := cmd.Long
	expectedSubcommands := []string{"list", "get", "set", "install", "bundled"}
	for _, subcmd := range expectedSubcommands {
		if !strings.Contains(longDesc, subcmd) {
			t.Errorf("Long description should mention subcommand: %s", subcmd)
		}
	}
}

func TestSchemeCommandIntegration(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "scheme command without args shows help",
			test: testSchemeCommandWithoutArgs,
		},
		{
			name: "scheme command with invalid subcommand shows error",
			test: testSchemeCommandWithInvalidSubcommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testSchemeCommandWithoutArgs(t *testing.T) {
	// Arrange
	cmd := Command()

	// Act
	err := cmd.Execute()

	// Assert
	// The command should not error when run without args (it should show help)
	if err != nil {
		t.Errorf("Expected no error when running command without args, got: %v", err)
	}
}

func testSchemeCommandWithInvalidSubcommand(t *testing.T) {
	// Arrange
	cmd := Command()
	cmd.SetArgs([]string{"invalid-subcommand"})

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error when running command with invalid subcommand")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected error to contain 'unknown command', got: %v", err)
	}
}

func TestSubcommandCreation(t *testing.T) {
	tests := []struct {
		name        string
		createFunc  func() *cobra.Command
		expectedUse string
	}{
		{
			name:        "list command creation",
			createFunc:  listCommand,
			expectedUse: "list",
		},
		{
			name:        "get command creation",
			createFunc:  getCommand,
			expectedUse: "get [property]",
		},
		{
			name:        "set command creation",
			createFunc:  setCommand,
			expectedUse: "set [scheme] [flavour] [mode]",
		},
		{
			name:        "install command creation",
			createFunc:  installCommand,
			expectedUse: "install [scheme-name]",
		},
		{
			name:        "bundled command creation",
			createFunc:  bundledCommand,
			expectedUse: "bundled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			cmd := tt.createFunc()

			// Assert
			if cmd == nil {
				t.Fatal("Expected command to be created, got nil")
			}
			if cmd.Use != tt.expectedUse {
				t.Errorf("Expected Use to be '%s', got '%s'", tt.expectedUse, cmd.Use)
			}
			if cmd.Short == "" {
				t.Error("Expected Short description to be non-empty")
			}
			if cmd.Long == "" {
				t.Error("Expected Long description to be non-empty")
			}
			if cmd.RunE == nil {
				t.Error("Command should have a RunE function")
			}
		})
	}
}
