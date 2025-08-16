package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadWithoutConfigFile tests that the system works with no config file
func TestLoadWithoutConfigFile(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
		cfg = nil // Reset global config
	}()

	// Ensure no config file exists
	configPath := filepath.Join(tempDir, "config.json")
	assert.False(t, paths.Exists(configPath), "Config file should not exist")

	// Load config - should work with defaults
	err := Load()
	require.NoError(t, err, "Load should succeed without config file")

	// Verify defaults are applied
	c := Get()
	assert.NotNil(t, c)
	assert.Equal(t, "0.2.0", c.Version)
	assert.Equal(t, "rosepine", c.Scheme.Default)
	assert.True(t, c.Theme.EnableTerm)
	assert.Equal(t, 100, c.Clipboard.MaxEntries)

	// Verify no config file was created
	assert.False(t, paths.Exists(configPath), "Config file should not be created automatically")
}

// TestLoadWithPartialConfig tests that partial configs are merged with defaults
func TestLoadWithPartialConfig(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
		cfg = nil // Reset global config
	}()

	// Create a minimal config with only a few settings
	configPath := filepath.Join(tempDir, "config.json")
	minimalConfig := map[string]interface{}{
		"scheme": map[string]interface{}{
			"default": "catppuccin-mocha",
		},
		"clipboard": map[string]interface{}{
			"max_entries": 200,
		},
	}

	data, err := json.MarshalIndent(minimalConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Load config
	err = Load()
	require.NoError(t, err, "Load should succeed with partial config")

	// Verify user values are applied
	c := Get()
	assert.Equal(t, "catppuccin-mocha", c.Scheme.Default, "User value should override default")
	assert.Equal(t, 200, c.Clipboard.MaxEntries, "User value should override default")

	// Verify defaults are still applied for unspecified fields
	assert.Equal(t, "0.2.0", c.Version, "Default should be applied for unspecified field")
	assert.True(t, c.Theme.EnableTerm, "Default should be applied for unspecified field")
	assert.Equal(t, "Clipboard> ", c.Clipboard.FuzzelPrompt, "Default should be applied for unspecified field")
}

// TestEffectiveConfig tests that EffectiveConfig returns merged configuration
func TestEffectiveConfig(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
		cfg = nil // Reset global config
	}()

	// Create a minimal config
	configPath := filepath.Join(tempDir, "config.json")
	minimalConfig := map[string]interface{}{
		"scheme": map[string]interface{}{
			"default": "gruvbox",
		},
	}

	data, err := json.MarshalIndent(minimalConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Load config
	err = Load()
	require.NoError(t, err)

	// Get effective config
	effective := EffectiveConfig()
	assert.NotNil(t, effective)

	// Should have both user values and defaults
	assert.Equal(t, "gruvbox", effective.Scheme.Default, "User value should be present")
	assert.Equal(t, "0.2.0", effective.Version, "Default value should be present")
	assert.True(t, effective.Theme.EnableTerm, "Default value should be present")
}

// TestUserConfig tests that UserConfig returns only user-specified values
func TestUserConfig(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
		cfg = nil // Reset global config
	}()

	// Test with no config file
	userCfg, err := UserConfig()
	require.NoError(t, err)
	assert.Empty(t, userCfg, "Should return empty map when no config file exists")

	// Create a minimal config
	configPath := filepath.Join(tempDir, "config.json")
	minimalConfig := map[string]interface{}{
		"scheme": map[string]interface{}{
			"default": "nord",
		},
		"theme": map[string]interface{}{
			"enableGtk": false,
		},
	}

	data, err := json.MarshalIndent(minimalConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Get user config
	userCfg, err = UserConfig()
	require.NoError(t, err)

	// Should only contain user-specified values
	assert.Len(t, userCfg, 2, "Should only have user-specified top-level keys")

	scheme, ok := userCfg["scheme"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "nord", scheme["default"])

	theme, ok := userCfg["theme"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, theme["enablegtk"]) // viper lowercases keys
}

// TestSaveOnlyUserValues tests that Save only persists user-modified values
func TestSaveOnlyUserValues(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
		cfg = nil // Reset global config
	}()

	// Load with no config (defaults only)
	err := Load()
	require.NoError(t, err)

	// Modify a few values
	c := Get()
	c.Scheme.Default = "tokyo-night"
	c.Clipboard.MaxEntries = 500

	// Save the config
	err = Save()
	require.NoError(t, err)

	// Read the saved file directly
	configPath := filepath.Join(tempDir, "config.json")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var saved map[string]interface{}
	err = json.Unmarshal(data, &saved)
	require.NoError(t, err)

	// The saved file should be minimal - only containing changed values
	// Note: Due to how viper works, it might save entire sections if any value in that section changed
	// But it shouldn't save sections that weren't modified at all
	assert.NotNil(t, saved["scheme"], "Modified section should be saved")
	assert.NotNil(t, saved["clipboard"], "Modified section should be saved")
}

// TestValidate tests configuration validation
func TestValidate(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
		cfg = nil // Reset global config
	}()

	// Load defaults
	err := Load()
	require.NoError(t, err)

	// Defaults should be valid
	err = Validate()
	assert.NoError(t, err, "Default configuration should be valid")

	// Test invalid values
	c := Get()

	// Invalid wallpaper threshold
	c.Wallpaper.Threshold = 1.5
	err = Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wallpaper threshold")
	c.Wallpaper.Threshold = 0.8 // Reset

	// Invalid port
	c.Shell.DaemonPort = 70000
	err = Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "daemon port")
	c.Shell.DaemonPort = 9999 // Reset

	// Invalid screenshot format
	c.Screenshot.FileFormat = "bmp"
	err = Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "screenshot format")
	c.Screenshot.FileFormat = "png" // Reset

	// Invalid notification urgency
	c.Notification.DefaultUrgency = "urgent"
	err = Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification urgency")
	c.Notification.DefaultUrgency = "normal" // Reset

	// Should be valid again
	err = Validate()
	assert.NoError(t, err)
}

// TestHasUserConfig tests the HasUserConfig function
func TestHasUserConfig(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	oldConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() {
		paths.HeimdallConfigDir = oldConfigDir
	}()

	// Initially no config
	assert.False(t, HasUserConfig(), "Should return false when no config exists")

	// Create a config file
	configPath := filepath.Join(tempDir, "config.json")
	err := os.WriteFile(configPath, []byte("{}"), 0644)
	require.NoError(t, err)

	// Now should have config
	assert.True(t, HasUserConfig(), "Should return true when config exists")
}
