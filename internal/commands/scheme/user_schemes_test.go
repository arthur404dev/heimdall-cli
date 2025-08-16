package scheme

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommandWithSourceFlag(t *testing.T) {
	// Create temp directory with test schemes
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create user schemes
	createTestUserScheme(t, userDir, "user-theme1", "default")
	createTestUserScheme(t, userDir, "user-theme2", "variant1")

	// Set environment variable to use our test directory
	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	tests := []struct {
		name          string
		sourceFlag    string
		expectUser    bool
		expectBundled bool
		expectError   bool
	}{
		{
			name:          "list all sources by default",
			sourceFlag:    "",
			expectUser:    true,
			expectBundled: true,
			expectError:   false,
		},
		{
			name:          "list only user schemes",
			sourceFlag:    "user",
			expectUser:    true,
			expectBundled: false,
			expectError:   false,
		},
		{
			name:          "list only bundled schemes",
			sourceFlag:    "bundled",
			expectUser:    false,
			expectBundled: true,
			expectError:   false,
		},
		{
			name:          "list generated schemes (empty if none)",
			sourceFlag:    "generated",
			expectUser:    false,
			expectBundled: false,
			expectError:   false,
		},
		{
			name:          "invalid source filter",
			sourceFlag:    "invalid",
			expectUser:    false,
			expectBundled: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := NewListCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set flags
			if tt.sourceFlag != "" {
				cmd.Flags().Set("source", tt.sourceFlag)
			}

			// Execute
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()

			// Check for expected content
			if tt.expectUser {
				assert.Contains(t, output, "user-theme1", "Should contain user scheme")
				// Check for [user] indicator
				assert.Contains(t, output, "[user]", "Should have user source indicator")
			} else {
				assert.NotContains(t, output, "user-theme1", "Should not contain user scheme")
			}

			if tt.expectBundled {
				// At least one bundled scheme should be present
				hasBundled := strings.Contains(output, "catppuccin") ||
					strings.Contains(output, "gruvbox") ||
					strings.Contains(output, "onedark")
				assert.True(t, hasBundled, "Should contain bundled schemes")
			}
		})
	}
}

func TestListCommandSourceIndicators(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create test schemes
	createTestUserScheme(t, userDir, "my-custom-theme", "default")

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	// Test tree view
	cmd := NewListCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Use tree view (default)
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Check for colored indicators
	// The actual implementation uses color codes, so we check for the pattern
	assert.Contains(t, output, "my-custom-theme", "Should show user scheme")
	assert.Contains(t, output, "[user]", "Should have user indicator")

	// Test simple list view
	cmd = NewListCommand()
	buf = new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.Flags().Set("simple", "true")

	err = cmd.Execute()
	require.NoError(t, err)

	output = buf.String()
	assert.Contains(t, output, "my-custom-theme", "Should list user scheme in simple mode")
}

func TestGetCommandWithUserScheme(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create a user scheme
	schemeContent := `{
		"name": "My Custom Theme",
		"author": "Test User",
		"colors": {
			"base00": "#1a1a1a",
			"base01": "#2a2a2a",
			"base02": "#3a3a3a"
		}
	}`
	createTestUserSchemeWithContent(t, userDir, "custom-theme", "default", schemeContent)

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	// Test get command
	cmd := NewGetCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Set arguments
	cmd.SetArgs([]string{"custom-theme"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Check for scheme information
	assert.Contains(t, output, "My Custom Theme", "Should show scheme name")
	assert.Contains(t, output, "Test User", "Should show author")
	assert.Contains(t, output, "Source: user", "Should indicate user source")
	assert.Contains(t, output, "#1a1a1a", "Should show colors")
}

func TestInstallCommandWithUserFlag(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	tests := []struct {
		name        string
		schemeName  string
		userFlag    bool
		expectPath  string
		expectError bool
	}{
		{
			name:        "install bundled scheme to user directory",
			schemeName:  "catppuccin",
			userFlag:    true,
			expectPath:  filepath.Join(userDir, "catppuccin"),
			expectError: false,
		},
		{
			name:        "install without user flag (default behavior)",
			schemeName:  "gruvbox",
			userFlag:    false,
			expectPath:  "", // Would go to default location
			expectError: false,
		},
		{
			name:        "install non-existent scheme",
			schemeName:  "non-existent",
			userFlag:    true,
			expectPath:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInstallCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set arguments and flags
			cmd.SetArgs([]string{tt.schemeName})
			if tt.userFlag {
				cmd.Flags().Set("user", "true")
			}

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check if scheme was installed to user directory
			if tt.userFlag && tt.expectPath != "" {
				assert.DirExists(t, tt.expectPath, "Scheme should be installed to user directory")
			}

			output := buf.String()
			if tt.userFlag {
				assert.Contains(t, output, "user", "Should indicate installation to user directory")
			}
		})
	}
}

func TestSourceFilteringIntegration(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create various schemes
	createTestUserScheme(t, userDir, "user-only", "default")
	createTestUserScheme(t, userDir, "catppuccin", "custom") // Override bundled

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	// Test that user version takes priority
	cmd := NewGetCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"catppuccin", "custom"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Source: user", "User version should take priority")

	// Test listing with source filter shows correct schemes
	cmd = NewListCommand()
	buf = new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.Flags().Set("source", "user")
	cmd.Flags().Set("simple", "true")

	err = cmd.Execute()
	require.NoError(t, err)

	output = buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have both user schemes
	userSchemes := 0
	for _, line := range lines {
		if strings.Contains(line, "user-only") || strings.Contains(line, "catppuccin/custom") {
			userSchemes++
		}
	}
	assert.GreaterOrEqual(t, userSchemes, 2, "Should list user schemes")
}

// Helper functions

func createTestUserScheme(t *testing.T, baseDir, scheme, variant string) {
	t.Helper()
	content := `{
		"name": "Test Scheme",
		"author": "Test",
		"colors": {
			"base00": "#000000",
			"base01": "#111111"
		}
	}`
	createTestUserSchemeWithContent(t, baseDir, scheme, variant, content)
}

func createTestUserSchemeWithContent(t *testing.T, baseDir, scheme, variant, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, scheme, variant)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	file := filepath.Join(dir, "dark.json")
	err = os.WriteFile(file, []byte(content), 0644)
	require.NoError(t, err)
}

func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}
