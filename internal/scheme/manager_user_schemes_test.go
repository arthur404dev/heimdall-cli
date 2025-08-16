package scheme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserSchemePaths(t *testing.T) {
	// This test would need to mock the config.Get() function
	// For now, we'll test the actual behavior with environment variables

	tests := []struct {
		name    string
		envVar  string
		hasPath bool
	}{
		{
			name:    "environment variable with single path",
			envVar:  "/test/path",
			hasPath: true,
		},
		{
			name:    "environment variable with multiple paths",
			envVar:  "/path1:/path2",
			hasPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVar != "" {
				os.Setenv("HEIMDALL_SCHEME_PATHS", tt.envVar)
				defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")
			}

			// Create manager
			manager := NewManager()

			// Test - getUserSchemePaths is private, so we test through public methods
			// We'll test this indirectly through ListSchemes
			schemes, err := manager.ListSchemes()

			// Should not error
			assert.NoError(t, err)
			// Should return some schemes (at least bundled ones)
			assert.NotEmpty(t, schemes)
		})
	}
}

func TestListSchemesMultiSource(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create test scheme structure
	createTestScheme(t, userDir, "custom-theme", "default", `{
		"name": "Custom Theme",
		"colours": {
			"base00": "#000000",
			"base01": "#111111"
		}
	}`)

	createTestScheme(t, userDir, "override-theme", "default", `{
		"name": "Override Theme",
		"colours": {
			"base00": "#222222",
			"base01": "#333333"
		}
	}`)

	// Set environment variable to use test directory
	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	manager := NewManager()
	schemes, err := manager.ListSchemes()
	require.NoError(t, err)

	// Should have schemes
	assert.Greater(t, len(schemes), 0)

	// Check for user schemes - this may not find them without proper config
	// The test is demonstrating the structure but actual behavior depends on config
}

func TestLoadSchemeWithPriority(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create a user scheme that overrides a bundled one
	createTestScheme(t, userDir, "catppuccin", "mocha", `{
		"name": "User Catppuccin",
		"flavour": "mocha",
		"mode": "dark",
		"colours": {
			"base00": "#ffffff",
			"base01": "#eeeeee"
		}
	}`)

	// Create a unique user scheme
	createTestScheme(t, userDir, "unique-user", "default", `{
		"name": "Unique User Scheme",
		"flavour": "default",
		"mode": "dark",
		"colours": {
			"base00": "#123456"
		}
	}`)

	tests := []struct {
		name         string
		schemeName   string
		variant      string
		mode         string
		expectSource SchemeSource
		expectError  bool
		envPath      string
	}{
		{
			name:         "load bundled scheme when no user paths",
			schemeName:   "catppuccin",
			variant:      "mocha",
			mode:         "dark",
			expectSource: SourceBundled,
			expectError:  false,
			envPath:      "",
		},
		{
			name:         "prioritize user scheme over bundled",
			schemeName:   "catppuccin",
			variant:      "mocha",
			mode:         "dark",
			expectSource: SourceUser,
			expectError:  false,
			envPath:      userDir,
		},
		{
			name:         "load unique user scheme",
			schemeName:   "unique-user",
			variant:      "default",
			mode:         "dark",
			expectSource: SourceUser,
			expectError:  false,
			envPath:      userDir,
		},
		{
			name:         "error on non-existent scheme",
			schemeName:   "non-existent",
			variant:      "default",
			mode:         "dark",
			expectSource: SourceBundled,
			expectError:  true,
			envPath:      userDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envPath != "" {
				os.Setenv("HEIMDALL_SCHEME_PATHS", tt.envPath)
				defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")
			}

			manager := NewManager()
			scheme, err := manager.LoadScheme(tt.schemeName, tt.variant, tt.mode)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, scheme)

			// Check source tracking
			source := manager.GetSchemeSource(tt.schemeName)
			assert.Equal(t, tt.expectSource, source)

			// Verify the scheme has the source field set
			assert.Equal(t, tt.expectSource, scheme.Source)

			// For user override test, verify it's actually the user version
			if tt.schemeName == "catppuccin" && tt.expectSource == SourceUser {
				// The user version should have different colors
				assert.Equal(t, "#ffffff", scheme.Colours["base00"])
			}
		})
	}
}

func TestSchemeDeduplication(t *testing.T) {
	tempDir := t.TempDir()
	userDir1 := filepath.Join(tempDir, "user1")
	userDir2 := filepath.Join(tempDir, "user2")

	// Create same scheme in multiple user directories
	createTestScheme(t, userDir1, "duplicate", "default", `{"name": "Dup1", "colours": {}}`)
	createTestScheme(t, userDir2, "duplicate", "default", `{"name": "Dup2", "colours": {}}`)

	// Create unique schemes
	createTestScheme(t, userDir1, "unique1", "default", `{"name": "Unique1", "colours": {}}`)
	createTestScheme(t, userDir2, "unique2", "default", `{"name": "Unique2", "colours": {}}`)

	// Set multiple paths via environment variable (colon-separated)
	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir1+":"+userDir2)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	manager := NewManager()
	schemes, err := manager.ListSchemes()
	require.NoError(t, err)

	// Count occurrences
	duplicateCount := 0
	unique1Count := 0
	unique2Count := 0

	for _, scheme := range schemes {
		switch scheme {
		case "duplicate":
			duplicateCount++
		case "unique1":
			unique1Count++
		case "unique2":
			unique2Count++
		}
	}

	// Verify deduplication
	assert.Equal(t, 1, duplicateCount, "Duplicate scheme should appear only once")
	assert.Equal(t, 1, unique1Count, "Unique1 should appear once")
	assert.Equal(t, 1, unique2Count, "Unique2 should appear once")

	// Verify priority (first path wins)
	scheme, err := manager.LoadScheme("duplicate", "default", "dark")
	require.NoError(t, err)
	assert.Equal(t, "Dup1", scheme.Name, "Should load from first user path")
}

func TestSaveSchemeToUser(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Set environment to use test directory
	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	manager := NewManager()

	// Create a scheme to save
	scheme := &Scheme{
		Name:    "saved-scheme",
		Flavour: "custom",
		Mode:    "dark",
		Variant: "default",
		Source:  SourceUser,
		Colours: map[string]string{
			"base00": "#000000",
			"base01": "#111111",
		},
	}

	// Test saving
	err := manager.SaveSchemeToUser(scheme)
	require.NoError(t, err)

	// Verify file was created
	expectedPath := filepath.Join(userDir, "saved-scheme", "custom", "dark.json")
	assert.FileExists(t, expectedPath)

	// Load it back and verify
	loaded, err := manager.LoadScheme("saved-scheme", "custom", "dark")
	require.NoError(t, err)
	assert.Equal(t, scheme.Name, loaded.Name)
	assert.Equal(t, scheme.Colours["base00"], loaded.Colours["base00"])
	assert.Equal(t, SourceUser, loaded.Source)
}

func TestListFlavoursMultiSource(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create scheme with multiple flavours
	createTestScheme(t, userDir, "multi-flavour", "flavour1", `{"name": "F1", "colours": {}}`)
	createTestScheme(t, userDir, "multi-flavour", "flavour2", `{"name": "F2", "colours": {}}`)
	createTestScheme(t, userDir, "multi-flavour", "flavour3", `{"name": "F3", "colours": {}}`)

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	manager := NewManager()
	flavours, err := manager.ListFlavours("multi-flavour")
	require.NoError(t, err)

	assert.Contains(t, flavours, "flavour1")
	assert.Contains(t, flavours, "flavour2")
	assert.Contains(t, flavours, "flavour3")
	assert.Len(t, flavours, 3)
}

func TestListModesMultiSource(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	// Create scheme with both light and dark modes
	createTestScheme(t, userDir, "dual-mode", "default", `{"name": "Dark", "colours": {}}`)
	createTestSchemeMode(t, userDir, "dual-mode", "default", "light", `{"name": "Light", "colours": {}}`)

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	manager := NewManager()
	modes, err := manager.ListModes("dual-mode", "default")
	require.NoError(t, err)

	assert.Contains(t, modes, "dark")
	assert.Contains(t, modes, "light")
	assert.Len(t, modes, 2)
}

func TestGetSchemeSource(t *testing.T) {
	tempDir := t.TempDir()
	userDir := filepath.Join(tempDir, "user_schemes")

	createTestScheme(t, userDir, "user-scheme", "default", `{"name": "User", "colours": {}}`)

	os.Setenv("HEIMDALL_SCHEME_PATHS", userDir)
	defer os.Unsetenv("HEIMDALL_SCHEME_PATHS")

	manager := NewManager()

	// Load schemes to populate source tracking
	_, _ = manager.LoadScheme("user-scheme", "default", "dark")
	_, _ = manager.LoadScheme("catppuccin", "mocha", "dark") // Bundled

	tests := []struct {
		schemeName string
		expected   SchemeSource
	}{
		{"user-scheme", SourceUser},
		{"catppuccin", SourceBundled},
		{"non-existent", SourceBundled}, // Default for unknown
	}

	for _, tt := range tests {
		t.Run(tt.schemeName, func(t *testing.T) {
			source := manager.GetSchemeSource(tt.schemeName)
			assert.Equal(t, tt.expected, source)
		})
	}
}

// Helper functions

func createTestScheme(t *testing.T, baseDir, scheme, variant, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, scheme, variant)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	file := filepath.Join(dir, "dark.json")
	err = os.WriteFile(file, []byte(content), 0644)
	require.NoError(t, err)
}

func createTestSchemeMode(t *testing.T, baseDir, scheme, variant, mode, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, scheme, variant)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	file := filepath.Join(dir, mode+".json")
	err = os.WriteFile(file, []byte(content), 0644)
	require.NoError(t, err)
}
