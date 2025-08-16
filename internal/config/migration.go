package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/viper"
)

// migrateFromYAML migrates configuration from YAML format to JSON format
func migrateFromYAML(yamlConfigPath string) error {
	// Read existing YAML config
	viper.SetConfigFile(yamlConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read YAML config: %w", err)
	}

	// Create backup of YAML config
	backupPath := yamlConfigPath + ".backup"
	if err := paths.CopyFile(yamlConfigPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup YAML config: %w", err)
	}

	// Ensure Heimdall config directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save as JSON
	jsonConfigPath := filepath.Join(paths.HeimdallConfigDir, "config.json")
	viper.SetConfigType("json")
	if err := viper.WriteConfigAs(jsonConfigPath); err != nil {
		return fmt.Errorf("failed to save JSON config: %w", err)
	}

	// Remove old YAML config
	if err := os.Remove(yamlConfigPath); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: Failed to remove old YAML config: %v\n", err)
	}

	fmt.Printf("Successfully migrated configuration from YAML to JSON\n")
	fmt.Printf("Backup saved at: %s\n", backupPath)
	fmt.Printf("New config saved at: %s\n", jsonConfigPath)

	return nil
}

// CheckForMigration checks if migration is needed
func CheckForMigration() bool {
	// Check if Heimdall JSON config exists
	heimdallConfigJSON := filepath.Join(paths.HeimdallConfigDir, "config.json")
	if paths.Exists(heimdallConfigJSON) {
		return false // Already configured with JSON
	}

	// Check if Heimdall YAML config exists (needs migration to JSON)
	heimdallConfigYAML := filepath.Join(paths.HeimdallConfigDir, "config.yaml")
	if paths.Exists(heimdallConfigYAML) {
		return true // Need to migrate from YAML to JSON
	}

	// Check for old config.yml variant
	heimdallConfigYML := filepath.Join(paths.HeimdallConfigDir, "config.yml")
	if paths.Exists(heimdallConfigYML) {
		return true // Need to migrate from YAML to JSON
	}

	// Check for caelestia config (old name)
	caelestiaConfigDir := filepath.Join(os.Getenv("HOME"), ".config", "caelestia")
	caelestiaConfigJSON := filepath.Join(caelestiaConfigDir, "config.json")
	caelestiaConfigYAML := filepath.Join(caelestiaConfigDir, "config.yaml")

	if paths.Exists(caelestiaConfigJSON) || paths.Exists(caelestiaConfigYAML) {
		return true // Need to migrate from caelestia
	}

	return false
}

// GetMigrationPath returns the path of the config file that needs migration
func GetMigrationPath() string {
	// Check for Heimdall YAML configs
	heimdallConfigYAML := filepath.Join(paths.HeimdallConfigDir, "config.yaml")
	if paths.Exists(heimdallConfigYAML) {
		return heimdallConfigYAML
	}

	heimdallConfigYML := filepath.Join(paths.HeimdallConfigDir, "config.yml")
	if paths.Exists(heimdallConfigYML) {
		return heimdallConfigYML
	}

	// Check for caelestia configs
	caelestiaConfigDir := filepath.Join(os.Getenv("HOME"), ".config", "caelestia")
	caelestiaConfigJSON := filepath.Join(caelestiaConfigDir, "config.json")
	if paths.Exists(caelestiaConfigJSON) {
		return caelestiaConfigJSON
	}

	caelestiaConfigYAML := filepath.Join(caelestiaConfigDir, "config.yaml")
	if paths.Exists(caelestiaConfigYAML) {
		return caelestiaConfigYAML
	}

	return ""
}

// MigrateConfig performs automatic migration from old config formats
func MigrateConfig() error {
	migrationPath := GetMigrationPath()
	if migrationPath == "" {
		return nil // No migration needed
	}

	fmt.Printf("🔄 Found old configuration at: %s\n", migrationPath)
	fmt.Println("   Migrating to new format...")

	// Determine file type
	ext := filepath.Ext(migrationPath)
	if ext == ".yaml" || ext == ".yml" {
		return migrateFromYAML(migrationPath)
	}

	// For JSON files (e.g., from caelestia), just copy and potentially update structure
	return migrateFromOldJSON(migrationPath)
}

// migrateFromOldJSON migrates from old JSON format (e.g., caelestia)
func migrateFromOldJSON(oldConfigPath string) error {
	// Read existing JSON config
	viper.SetConfigFile(oldConfigPath)
	viper.SetConfigType("json")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read old JSON config: %w", err)
	}

	// Create backup
	backupPath := oldConfigPath + ".backup"
	if err := paths.CopyFile(oldConfigPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old config: %w", err)
	}

	// Ensure Heimdall config directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check for any field renames or structure changes
	migrateFieldNames()

	// Save as new JSON config
	jsonConfigPath := filepath.Join(paths.HeimdallConfigDir, "config.json")
	if err := viper.WriteConfigAs(jsonConfigPath); err != nil {
		return fmt.Errorf("failed to save new JSON config: %w", err)
	}

	fmt.Printf("✓ Successfully migrated configuration\n")
	fmt.Printf("  Backup saved at: %s\n", backupPath)
	fmt.Printf("  New config saved at: %s\n", jsonConfigPath)

	// If migrating from caelestia, suggest removing old directory
	if strings.Contains(oldConfigPath, "caelestia") {
		fmt.Printf("\n💡 You can remove the old caelestia config directory:\n")
		fmt.Printf("   rm -rf %s\n", filepath.Dir(oldConfigPath))
	}

	return nil
}

// migrateFieldNames handles any field renames or structure changes
func migrateFieldNames() {
	// Example migrations (add as needed):

	// Rename old field names to new ones
	if viper.IsSet("colorScheme") {
		viper.Set("scheme.default", viper.Get("colorScheme"))
		// Don't delete old field, let it be ignored
	}

	// Handle structure changes
	if viper.IsSet("enableGTK") {
		viper.Set("theme.enableGtk", viper.Get("enableGTK"))
	}

	if viper.IsSet("enableQT") {
		viper.Set("theme.enableQt", viper.Get("enableQT"))
	}

	// Migrate wallpaper settings
	if viper.IsSet("wallpaperDir") {
		viper.Set("wallpaper.directory", viper.Get("wallpaperDir"))
	}

	if viper.IsSet("wallpaperDirs") {
		viper.Set("wallpaper.directories", viper.Get("wallpaperDirs"))
	}
}
