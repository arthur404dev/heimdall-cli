package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/heimdall-cli/heimdall/internal/utils/paths"
	"github.com/spf13/viper"
)

// Config represents the main configuration structure
type Config struct {
	Version      string                  `mapstructure:"version" json:"version" yaml:"version"`
	MigratedFrom string                  `mapstructure:"migrated_from,omitempty" json:"migrated_from,omitempty" yaml:"migrated_from,omitempty"`
	Theme        ThemeConfig             `mapstructure:"theme" json:"theme" yaml:"theme"`
	Toggles      map[string]ToggleConfig `mapstructure:"toggles" json:"toggles" yaml:"toggles"`
	Shell        ShellConfig             `mapstructure:"shell" json:"shell" yaml:"shell"`
	Scheme       SchemeConfig            `mapstructure:"scheme" json:"scheme" yaml:"scheme"`
	Wallpaper    WallpaperConfig         `mapstructure:"wallpaper" json:"wallpaper" yaml:"wallpaper"`
	External     ExternalTools           `mapstructure:"external_tools" json:"external_tools" yaml:"external_tools"`
}

// ThemeConfig represents theme configuration
type ThemeConfig struct {
	EnableTerm      bool `mapstructure:"enableTerm" json:"enableTerm" yaml:"enableTerm"`
	EnableHypr      bool `mapstructure:"enableHypr" json:"enableHypr" yaml:"enableHypr"`
	EnableDiscord   bool `mapstructure:"enableDiscord" json:"enableDiscord" yaml:"enableDiscord"`
	EnableSpicetify bool `mapstructure:"enableSpicetify" json:"enableSpicetify" yaml:"enableSpicetify"`
	EnableFuzzel    bool `mapstructure:"enableFuzzel" json:"enableFuzzel" yaml:"enableFuzzel"`
	EnableBtop      bool `mapstructure:"enableBtop" json:"enableBtop" yaml:"enableBtop"`
	EnableGtk       bool `mapstructure:"enableGtk" json:"enableGtk" yaml:"enableGtk"`
	EnableQt        bool `mapstructure:"enableQt" json:"enableQt" yaml:"enableQt"`
}

// ToggleConfig represents workspace toggle configuration
type ToggleConfig struct {
	Apps map[string]AppConfig `mapstructure:"apps" json:"apps" yaml:"apps"`
}

// AppConfig represents an application configuration for toggles
type AppConfig struct {
	Enable  bool             `mapstructure:"enable" json:"enable" yaml:"enable"`
	Match   []map[string]any `mapstructure:"match" json:"match" yaml:"match"`
	Command []string         `mapstructure:"command" json:"command" yaml:"command"`
	Move    bool             `mapstructure:"move" json:"move" yaml:"move"`
}

// ShellConfig represents shell configuration
type ShellConfig struct {
	Command    string   `mapstructure:"command" json:"command" yaml:"command"`
	Args       []string `mapstructure:"args" json:"args" yaml:"args"`
	LogRules   string   `mapstructure:"log_rules" json:"log_rules" yaml:"log_rules"`
	DaemonPort int      `mapstructure:"daemon_port" json:"daemon_port" yaml:"daemon_port"`
}

// SchemeConfig represents scheme configuration
type SchemeConfig struct {
	Default     string `mapstructure:"default" json:"default" yaml:"default"`
	AutoMode    bool   `mapstructure:"auto_mode" json:"auto_mode" yaml:"auto_mode"`
	MaterialYou bool   `mapstructure:"material_you" json:"material_you" yaml:"material_you"`
}

// WallpaperConfig represents wallpaper configuration
type WallpaperConfig struct {
	Directory string  `mapstructure:"directory" json:"directory" yaml:"directory"`
	Filter    bool    `mapstructure:"filter" json:"filter" yaml:"filter"`
	Threshold float64 `mapstructure:"threshold" json:"threshold" yaml:"threshold"`
	SmartMode bool    `mapstructure:"smart_mode" json:"smart_mode" yaml:"smart_mode"`
}

// ExternalTools represents external tool paths
type ExternalTools struct {
	Grim        string `mapstructure:"grim" json:"grim" yaml:"grim"`
	Slurp       string `mapstructure:"slurp" json:"slurp" yaml:"slurp"`
	Swappy      string `mapstructure:"swappy" json:"swappy" yaml:"swappy"`
	WlClipboard string `mapstructure:"wl_clipboard" json:"wl_clipboard" yaml:"wl_clipboard"`
	WlScreenrec string `mapstructure:"wl_screenrec" json:"wl_screenrec" yaml:"wl_screenrec"`
	Cliphist    string `mapstructure:"cliphist" json:"cliphist" yaml:"cliphist"`
	Fuzzel      string `mapstructure:"fuzzel" json:"fuzzel" yaml:"fuzzel"`
	DartSass    string `mapstructure:"dart_sass" json:"dart_sass" yaml:"dart_sass"`
	Libnotify   string `mapstructure:"libnotify" json:"libnotify" yaml:"libnotify"`
	Qs          string `mapstructure:"qs" json:"qs" yaml:"qs"`
	App2unit    string `mapstructure:"app2unit" json:"app2unit" yaml:"app2unit"`
}

var (
	// Global config instance
	cfg *Config

	// Default configuration values
	defaults = Config{
		Version: "1.0.0",
		Theme: ThemeConfig{
			EnableTerm:      true,
			EnableHypr:      true,
			EnableDiscord:   true,
			EnableSpicetify: true,
			EnableFuzzel:    true,
			EnableBtop:      true,
			EnableGtk:       true,
			EnableQt:        true,
		},
		Shell: ShellConfig{
			Command:    "qs",
			Args:       []string{"-c", "heimdall", "-n"},
			DaemonPort: 9999,
		},
		Scheme: SchemeConfig{
			Default:     "rosepine",
			AutoMode:    true,
			MaterialYou: true,
		},
		Wallpaper: WallpaperConfig{
			Directory: paths.WallpapersDir,
			Filter:    true,
			Threshold: 0.8,
			SmartMode: true,
		},
		External: ExternalTools{
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
		},
	}
)

// Load loads the configuration from file
func Load() error {
	// Set defaults
	setDefaults()

	// Try to load from Heimdall config
	heimdallConfig := filepath.Join(paths.HeimdallConfigDir, "config.yaml")
	if paths.Exists(heimdallConfig) {
		viper.SetConfigFile(heimdallConfig)
	} else {
		// Check for legacy Caelestia config
		caelestiaConfig := filepath.Join(paths.CaelestiaConfigDir, "cli.json")
		if paths.Exists(caelestiaConfig) {
			// Attempt migration
			if err := migrateFromCaelestia(caelestiaConfig); err != nil {
				return fmt.Errorf("failed to migrate Caelestia config: %w", err)
			}
			viper.SetConfigFile(heimdallConfig)
		} else {
			// No config found, create default
			viper.SetConfigFile(heimdallConfig)
			if err := SaveDefaults(); err != nil {
				return fmt.Errorf("failed to save default config: %w", err)
			}
		}
	}

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Unmarshal into struct
	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		Load()
	}
	return cfg
}

// Save saves the current configuration
func Save() error {
	if cfg == nil {
		return fmt.Errorf("no configuration loaded")
	}

	configPath := filepath.Join(paths.HeimdallConfigDir, "config.yaml")

	// Ensure directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SaveDefaults saves the default configuration
func SaveDefaults() error {
	setDefaults()

	configPath := filepath.Join(paths.HeimdallConfigDir, "config.yaml")

	// Ensure directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	if err := viper.SafeWriteConfigAs(configPath); err != nil {
		if os.IsExist(err) {
			return nil // Config already exists
		}
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// setDefaults sets default values in Viper
func setDefaults() {
	viper.SetDefault("version", defaults.Version)
	viper.SetDefault("theme", defaults.Theme)
	viper.SetDefault("shell", defaults.Shell)
	viper.SetDefault("scheme", defaults.Scheme)
	viper.SetDefault("wallpaper", defaults.Wallpaper)
	viper.SetDefault("external", defaults.External)
}

// Reload reloads the configuration from file
func Reload() error {
	return Load()
}

// GetTheme returns the theme configuration
func GetTheme() ThemeConfig {
	c := Get()
	return c.Theme
}

// GetShell returns the shell configuration
func GetShell() ShellConfig {
	c := Get()
	return c.Shell
}

// GetScheme returns the scheme configuration
func GetScheme() SchemeConfig {
	c := Get()
	return c.Scheme
}

// GetWallpaper returns the wallpaper configuration
func GetWallpaper() WallpaperConfig {
	c := Get()
	return c.Wallpaper
}

// GetExternal returns the external tools configuration
func GetExternal() ExternalTools {
	c := Get()
	return c.External
}
