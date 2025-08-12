package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/heimdall-cli/heimdall/internal/utils/paths"
	"github.com/spf13/viper"
)

// CaelestiaConfig represents the legacy Caelestia configuration structure
type CaelestiaConfig struct {
	Theme   CaelestiaTheme             `json:"theme"`
	Toggles map[string]CaelestiaToggle `json:"toggles"`
}

// CaelestiaTheme represents the legacy theme configuration
type CaelestiaTheme struct {
	EnableTerm      bool `json:"enableTerm"`
	EnableHypr      bool `json:"enableHypr"`
	EnableDiscord   bool `json:"enableDiscord"`
	EnableSpicetify bool `json:"enableSpicetify"`
	EnableFuzzel    bool `json:"enableFuzzel"`
	EnableBtop      bool `json:"enableBtop"`
	EnableGtk       bool `json:"enableGtk"`
	EnableQt        bool `json:"enableQt"`
}

// CaelestiaToggle represents the legacy toggle configuration
type CaelestiaToggle struct {
	Apps map[string]CaelestiaApp `json:"apps"`
}

// CaelestiaApp represents the legacy app configuration
type CaelestiaApp struct {
	Enable  bool             `json:"enable"`
	Match   []map[string]any `json:"match"`
	Command []string         `json:"command"`
	Move    bool             `json:"move"`
}

// migrateFromCaelestia migrates configuration from Caelestia format to Heimdall format
func migrateFromCaelestia(caelestiaConfigPath string) error {
	// Read Caelestia config
	file, err := os.Open(caelestiaConfigPath)
	if err != nil {
		return fmt.Errorf("failed to open Caelestia config: %w", err)
	}
	defer file.Close()

	var caelestiaConfig CaelestiaConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&caelestiaConfig); err != nil {
		return fmt.Errorf("failed to decode Caelestia config: %w", err)
	}

	// Create backup of Caelestia config
	backupPath := caelestiaConfigPath + ".backup"
	if err := paths.CopyFile(caelestiaConfigPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup Caelestia config: %w", err)
	}

	// Set defaults first
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("theme", ThemeConfig{
		EnableTerm:      true,
		EnableHypr:      true,
		EnableDiscord:   true,
		EnableSpicetify: true,
		EnableFuzzel:    true,
		EnableBtop:      true,
		EnableGtk:       true,
		EnableQt:        true,
	})
	viper.SetDefault("shell", ShellConfig{
		Command:    "qs",
		Args:       []string{"-c", "heimdall", "-n"},
		DaemonPort: 9999,
	})
	viper.SetDefault("scheme", SchemeConfig{
		Default:     "rosepine",
		AutoMode:    true,
		MaterialYou: true,
	})
	viper.SetDefault("wallpaper", WallpaperConfig{
		Directory: paths.WallpapersDir,
		Filter:    true,
		Threshold: 0.8,
		SmartMode: true,
	})
	viper.SetDefault("external", ExternalTools{
		Grim:        "grim",
		Slurp:       "slurp",
		Swappy:      "swappy",
		WlClipboard: "wl-copy",
		WlScreenrec: "wl-screenrec",
		Cliphist:    "cliphist",
		Fuzzel:      "fuzzel",
		DartSass:    "sass",
		Libnotify:   "notify-send",
		Qs:          "qs",
		App2unit:    "app2unit",
	})

	// Migrate theme configuration
	viper.Set("theme.enableTerm", caelestiaConfig.Theme.EnableTerm)
	viper.Set("theme.enableHypr", caelestiaConfig.Theme.EnableHypr)
	viper.Set("theme.enableDiscord", caelestiaConfig.Theme.EnableDiscord)
	viper.Set("theme.enableSpicetify", caelestiaConfig.Theme.EnableSpicetify)
	viper.Set("theme.enableFuzzel", caelestiaConfig.Theme.EnableFuzzel)
	viper.Set("theme.enableBtop", caelestiaConfig.Theme.EnableBtop)
	viper.Set("theme.enableGtk", caelestiaConfig.Theme.EnableGtk)
	viper.Set("theme.enableQt", caelestiaConfig.Theme.EnableQt)

	// Migrate toggles configuration
	toggles := make(map[string]ToggleConfig)
	for name, toggle := range caelestiaConfig.Toggles {
		apps := make(map[string]AppConfig)
		for appName, app := range toggle.Apps {
			apps[appName] = AppConfig{
				Enable:  app.Enable,
				Match:   app.Match,
				Command: app.Command,
				Move:    app.Move,
			}
		}
		toggles[name] = ToggleConfig{Apps: apps}
	}
	viper.Set("toggles", toggles)

	// Set migration info
	viper.Set("migrated_from", "caelestia")
	viper.Set("version", "1.0.0")

	// Ensure Heimdall config directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create Heimdall config directory: %w", err)
	}

	// Save migrated config as JSON
	heimdallConfigPath := filepath.Join(paths.HeimdallConfigDir, "config.json")
	viper.SetConfigType("json") // Set to JSON
	if err := viper.SafeWriteConfigAs(heimdallConfigPath); err != nil {
		return fmt.Errorf("failed to save migrated config: %w", err)
	}

	// Migrate state files if they exist
	if err := migrateStateFiles(); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: Failed to migrate state files: %v\n", err)
	}

	fmt.Printf("Successfully migrated configuration from Caelestia to Heimdall\n")
	fmt.Printf("Backup saved at: %s\n", backupPath)
	fmt.Printf("New config saved at: %s\n", heimdallConfigPath)

	return nil
}

// migrateStateFiles migrates state files from Caelestia to Heimdall
func migrateStateFiles() error {
	// Migrate scheme state
	caelestiaScheme := filepath.Join(paths.CaelestiaStateDir, "scheme.json")
	if paths.Exists(caelestiaScheme) {
		heimdallScheme := filepath.Join(paths.HeimdallStateDir, "scheme.json")
		if err := paths.EnsureParentDir(heimdallScheme); err != nil {
			return fmt.Errorf("failed to create scheme directory: %w", err)
		}
		if err := paths.CopyFile(caelestiaScheme, heimdallScheme); err != nil {
			return fmt.Errorf("failed to copy scheme state: %w", err)
		}
	}

	// Migrate wallpaper state
	caelestiaWallpaper := filepath.Join(paths.CaelestiaStateDir, "wallpaper")
	if paths.IsDir(caelestiaWallpaper) {
		heimdallWallpaper := filepath.Join(paths.HeimdallStateDir, "wallpaper")
		if err := paths.EnsureDir(heimdallWallpaper); err != nil {
			return fmt.Errorf("failed to create wallpaper directory: %w", err)
		}

		// Copy wallpaper path file
		caelestiaPath := filepath.Join(caelestiaWallpaper, "path.txt")
		if paths.Exists(caelestiaPath) {
			heimdallPath := filepath.Join(heimdallWallpaper, "path.txt")
			if err := paths.CopyFile(caelestiaPath, heimdallPath); err != nil {
				return fmt.Errorf("failed to copy wallpaper path: %w", err)
			}
		}

		// Copy wallpaper thumbnail
		caelestiaThumbnail := filepath.Join(caelestiaWallpaper, "thumbnail.jpg")
		if paths.Exists(caelestiaThumbnail) {
			heimdallThumbnail := filepath.Join(heimdallWallpaper, "thumbnail.jpg")
			if err := paths.CopyFile(caelestiaThumbnail, heimdallThumbnail); err != nil {
				return fmt.Errorf("failed to copy wallpaper thumbnail: %w", err)
			}
		}
	}

	// Migrate theme directory
	caelestiaTheme := filepath.Join(paths.CaelestiaStateDir, "theme")
	if paths.IsDir(caelestiaTheme) {
		heimdallTheme := filepath.Join(paths.HeimdallStateDir, "theme")
		if err := paths.EnsureDir(heimdallTheme); err != nil {
			return fmt.Errorf("failed to create theme directory: %w", err)
		}
		// Note: We might want to recursively copy theme files here
	}

	return nil
}

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

// CheckForMigration checks if migration from Caelestia is needed
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

	// Check if Caelestia config exists
	caelestiaConfig := filepath.Join(paths.CaelestiaConfigDir, "cli.json")
	return paths.Exists(caelestiaConfig)
}
