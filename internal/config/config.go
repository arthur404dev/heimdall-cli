package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
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
	Screenshot   ScreenshotConfig        `mapstructure:"screenshot" json:"screenshot" yaml:"screenshot"`
	Recording    RecordingConfig         `mapstructure:"recording" json:"recording" yaml:"recording"`
	Clipboard    ClipboardConfig         `mapstructure:"clipboard" json:"clipboard" yaml:"clipboard"`
	Emoji        EmojiConfig             `mapstructure:"emoji" json:"emoji" yaml:"emoji"`
	PIP          PIPConfig               `mapstructure:"pip" json:"pip" yaml:"pip"`
	Notification NotificationConfig      `mapstructure:"notification" json:"notification" yaml:"notification"`
	Paths        PathsConfig             `mapstructure:"paths" json:"paths" yaml:"paths"`
	Network      NetworkConfig           `mapstructure:"network" json:"network" yaml:"network"`
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
	LogFile    string   `mapstructure:"log_file" json:"log_file" yaml:"log_file"`
	PidFile    string   `mapstructure:"pid_file" json:"pid_file" yaml:"pid_file"`
	IPCTimeout int      `mapstructure:"ipc_timeout" json:"ipc_timeout" yaml:"ipc_timeout"` // seconds
}

// SchemeConfig represents scheme configuration
type SchemeConfig struct {
	Default     string `mapstructure:"default" json:"default" yaml:"default"`
	AutoMode    bool   `mapstructure:"auto_mode" json:"auto_mode" yaml:"auto_mode"`
	MaterialYou bool   `mapstructure:"material_you" json:"material_you" yaml:"material_you"`
}

// WallpaperConfig represents wallpaper configuration
type WallpaperConfig struct {
	Directory  string   `mapstructure:"directory" json:"directory" yaml:"directory"`
	Filter     bool     `mapstructure:"filter" json:"filter" yaml:"filter"`
	Threshold  float64  `mapstructure:"threshold" json:"threshold" yaml:"threshold"`
	SmartMode  bool     `mapstructure:"smart_mode" json:"smart_mode" yaml:"smart_mode"`
	Extensions []string `mapstructure:"extensions" json:"extensions" yaml:"extensions"`
}

// ScreenshotConfig represents screenshot configuration
type ScreenshotConfig struct {
	Directory           string `mapstructure:"directory" json:"directory" yaml:"directory"`
	FileFormat          string `mapstructure:"file_format" json:"file_format" yaml:"file_format"` // png, jpg, webp
	FileNamePattern     string `mapstructure:"file_name_pattern" json:"file_name_pattern" yaml:"file_name_pattern"`
	CopyToClipboard     bool   `mapstructure:"copy_to_clipboard" json:"copy_to_clipboard" yaml:"copy_to_clipboard"`
	OpenWithSwappy      bool   `mapstructure:"open_with_swappy" json:"open_with_swappy" yaml:"open_with_swappy"`
	ShowNotification    bool   `mapstructure:"show_notification" json:"show_notification" yaml:"show_notification"`
	NotificationTimeout int    `mapstructure:"notification_timeout" json:"notification_timeout" yaml:"notification_timeout"` // seconds
	FreezeFileName      string `mapstructure:"freeze_file_name" json:"freeze_file_name" yaml:"freeze_file_name"`
}

// RecordingConfig represents recording configuration
type RecordingConfig struct {
	Directory        string `mapstructure:"directory" json:"directory" yaml:"directory"`
	FileFormat       string `mapstructure:"file_format" json:"file_format" yaml:"file_format"` // mp4, webm, mkv
	FileNamePattern  string `mapstructure:"file_name_pattern" json:"file_name_pattern" yaml:"file_name_pattern"`
	TempFileName     string `mapstructure:"temp_file_name" json:"temp_file_name" yaml:"temp_file_name"`
	ShowNotification bool   `mapstructure:"show_notification" json:"show_notification" yaml:"show_notification"`
	AudioSource      string `mapstructure:"audio_source" json:"audio_source" yaml:"audio_source"` // auto, none, specific device
}

// ClipboardConfig represents clipboard configuration
type ClipboardConfig struct {
	MaxEntries     int      `mapstructure:"max_entries" json:"max_entries" yaml:"max_entries"`
	FuzzelPrompt   string   `mapstructure:"fuzzel_prompt" json:"fuzzel_prompt" yaml:"fuzzel_prompt"`
	FuzzelArgs     []string `mapstructure:"fuzzel_args" json:"fuzzel_args" yaml:"fuzzel_args"`
	PreviewLength  int      `mapstructure:"preview_length" json:"preview_length" yaml:"preview_length"`
	DeleteOnSelect bool     `mapstructure:"delete_on_select" json:"delete_on_select" yaml:"delete_on_select"`
}

// EmojiConfig represents emoji configuration
type EmojiConfig struct {
	DataDirectory   string   `mapstructure:"data_directory" json:"data_directory" yaml:"data_directory"`
	Sources         []string `mapstructure:"sources" json:"sources" yaml:"sources"` // emoji sources to use
	FuzzelPrompt    string   `mapstructure:"fuzzel_prompt" json:"fuzzel_prompt" yaml:"fuzzel_prompt"`
	FuzzelArgs      []string `mapstructure:"fuzzel_args" json:"fuzzel_args" yaml:"fuzzel_args"`
	CopyToClipboard bool     `mapstructure:"copy_to_clipboard" json:"copy_to_clipboard" yaml:"copy_to_clipboard"`
	TypeDirectly    bool     `mapstructure:"type_directly" json:"type_directly" yaml:"type_directly"`
	DownloadTimeout int      `mapstructure:"download_timeout" json:"download_timeout" yaml:"download_timeout"` // seconds
}

// PIPConfig represents picture-in-picture configuration
type PIPConfig struct {
	Enabled        bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	PidFile        string   `mapstructure:"pid_file" json:"pid_file" yaml:"pid_file"`
	WindowSize     string   `mapstructure:"window_size" json:"window_size" yaml:"window_size"`             // e.g., "25%"
	WindowPosition string   `mapstructure:"window_position" json:"window_position" yaml:"window_position"` // e.g., "bottom-right"
	VideoApps      []string `mapstructure:"video_apps" json:"video_apps" yaml:"video_apps"`
	VideoKeywords  []string `mapstructure:"video_keywords" json:"video_keywords" yaml:"video_keywords"`
	PinWindows     bool     `mapstructure:"pin_windows" json:"pin_windows" yaml:"pin_windows"`
	AlwaysOnTop    bool     `mapstructure:"always_on_top" json:"always_on_top" yaml:"always_on_top"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Provider       string `mapstructure:"provider" json:"provider" yaml:"provider"`                      // notify-send, dunstify, auto
	DefaultTimeout int    `mapstructure:"default_timeout" json:"default_timeout" yaml:"default_timeout"` // seconds
	AppName        string `mapstructure:"app_name" json:"app_name" yaml:"app_name"`
	DefaultUrgency string `mapstructure:"default_urgency" json:"default_urgency" yaml:"default_urgency"` // low, normal, critical
}

// PathsConfig represents custom paths configuration
type PathsConfig struct {
	Templates string `mapstructure:"templates" json:"templates" yaml:"templates"`
	Schemes   string `mapstructure:"schemes" json:"schemes" yaml:"schemes"`
	StateDir  string `mapstructure:"state_dir" json:"state_dir" yaml:"state_dir"`
	CacheDir  string `mapstructure:"cache_dir" json:"cache_dir" yaml:"cache_dir"`
	DataDir   string `mapstructure:"data_dir" json:"data_dir" yaml:"data_dir"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	IPCTimeout     int `mapstructure:"ipc_timeout" json:"ipc_timeout" yaml:"ipc_timeout"`                // seconds
	HyprIPCTimeout int `mapstructure:"hypr_ipc_timeout" json:"hypr_ipc_timeout" yaml:"hypr_ipc_timeout"` // seconds
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
	Dunstify    string `mapstructure:"dunstify" json:"dunstify" yaml:"dunstify"`
	Qs          string `mapstructure:"qs" json:"qs" yaml:"qs"`
	App2unit    string `mapstructure:"app2unit" json:"app2unit" yaml:"app2unit"`
	Xclip       string `mapstructure:"xclip" json:"xclip" yaml:"xclip"`
	Pactl       string `mapstructure:"pactl" json:"pactl" yaml:"pactl"`
	Pidof       string `mapstructure:"pidof" json:"pidof" yaml:"pidof"`
	Pkill       string `mapstructure:"pkill" json:"pkill" yaml:"pkill"`
	Gdbus       string `mapstructure:"gdbus" json:"gdbus" yaml:"gdbus"`
}

var (
	// Global config instance
	cfg *Config

	// Default configuration values
	defaults = Config{
		Version: "0.2.0",
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
			LogFile:    "shell.log",
			PidFile:    "shell.pid",
			IPCTimeout: 5,
		},
		Scheme: SchemeConfig{
			Default:     "rosepine",
			AutoMode:    true,
			MaterialYou: true,
		},
		Wallpaper: WallpaperConfig{
			Directory:  paths.WallpapersDir,
			Filter:     true,
			Threshold:  0.8,
			SmartMode:  true,
			Extensions: []string{".jpg", ".jpeg", ".png", ".webp"},
		},
		Screenshot: ScreenshotConfig{
			Directory:           paths.ScreenshotsDir,
			FileFormat:          "png",
			FileNamePattern:     "screenshot_%Y%m%d_%H%M%S",
			CopyToClipboard:     true,
			OpenWithSwappy:      true,
			ShowNotification:    true,
			NotificationTimeout: 3,
			FreezeFileName:      "freeze.png",
		},
		Recording: RecordingConfig{
			Directory:        paths.RecordingsDir,
			FileFormat:       "mp4",
			FileNamePattern:  "recording_%Y%m%d_%H%M%S",
			TempFileName:     "recording.mp4",
			ShowNotification: true,
			AudioSource:      "auto",
		},
		Clipboard: ClipboardConfig{
			MaxEntries:     100,
			FuzzelPrompt:   "Clipboard> ",
			FuzzelArgs:     []string{"--dmenu", "--width", "50", "--lines", "20"},
			PreviewLength:  50,
			DeleteOnSelect: false,
		},
		Emoji: EmojiConfig{
			DataDirectory:   filepath.Join(paths.DataDir, "emoji"),
			Sources:         []string{"emoji.json"},
			FuzzelPrompt:    "Emoji> ",
			FuzzelArgs:      []string{"--dmenu", "--prompt"},
			CopyToClipboard: true,
			TypeDirectly:    false,
			DownloadTimeout: 30,
		},
		PIP: PIPConfig{
			Enabled:        true,
			PidFile:        "pip.pid",
			WindowSize:     "25%",
			WindowPosition: "bottom-right",
			VideoApps: []string{
				"mpv", "vlc", "firefox", "chromium", "chrome",
				"brave", "youtube", "netflix", "twitch", "spotify",
			},
			VideoKeywords: []string{
				"youtube", "netflix", "twitch", "vimeo",
				"- playing", "▶", "►", "video", "stream",
			},
			PinWindows:  true,
			AlwaysOnTop: true,
		},
		Notification: NotificationConfig{
			Enabled:        true,
			Provider:       "auto",
			DefaultTimeout: 5,
			AppName:        "heimdall",
			DefaultUrgency: "normal",
		},
		Paths: PathsConfig{
			Templates: "",
			Schemes:   "",
			StateDir:  "",
			CacheDir:  "",
			DataDir:   "",
		},
		Network: NetworkConfig{
			IPCTimeout:     5,
			HyprIPCTimeout: 5,
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
			Dunstify:    "dunstify",
			Qs:          "qs",
			App2unit:    "app2unit",
			Xclip:       "xclip",
			Pactl:       "pactl",
			Pidof:       "pidof",
			Pkill:       "pkill",
			Gdbus:       "gdbus",
		},
	}
)

// Load loads the configuration from file
func Load() error {
	// Set defaults
	setDefaults()

	// Set config type to JSON
	viper.SetConfigType("json")

	// Try to load from Heimdall config (JSON)
	heimdallConfig := filepath.Join(paths.HeimdallConfigDir, "config.json")

	// Check if we need to migrate from old YAML config
	oldYamlConfig := filepath.Join(paths.HeimdallConfigDir, "config.yaml")
	if !paths.Exists(heimdallConfig) && paths.Exists(oldYamlConfig) {
		// Migrate from YAML to JSON
		if err := migrateFromYAML(oldYamlConfig); err != nil {
			return fmt.Errorf("failed to migrate from YAML: %w", err)
		}
	}

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

	configPath := filepath.Join(paths.HeimdallConfigDir, "config.json")

	// Ensure directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set config type to JSON
	viper.SetConfigType("json")

	// Write config
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SaveDefaults saves the default configuration
func SaveDefaults() error {
	setDefaults()

	configPath := filepath.Join(paths.HeimdallConfigDir, "config.json")

	// Ensure directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set config type to JSON
	viper.SetConfigType("json")

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
	viper.SetDefault("screenshot", defaults.Screenshot)
	viper.SetDefault("recording", defaults.Recording)
	viper.SetDefault("clipboard", defaults.Clipboard)
	viper.SetDefault("emoji", defaults.Emoji)
	viper.SetDefault("pip", defaults.PIP)
	viper.SetDefault("notification", defaults.Notification)
	viper.SetDefault("paths", defaults.Paths)
	viper.SetDefault("network", defaults.Network)
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

// GetScreenshot returns the screenshot configuration
func GetScreenshot() ScreenshotConfig {
	c := Get()
	return c.Screenshot
}

// GetRecording returns the recording configuration
func GetRecording() RecordingConfig {
	c := Get()
	return c.Recording
}

// GetClipboard returns the clipboard configuration
func GetClipboard() ClipboardConfig {
	c := Get()
	return c.Clipboard
}

// GetEmoji returns the emoji configuration
func GetEmoji() EmojiConfig {
	c := Get()
	return c.Emoji
}

// GetPIP returns the PIP configuration
func GetPIP() PIPConfig {
	c := Get()
	return c.PIP
}

// GetNotification returns the notification configuration
func GetNotification() NotificationConfig {
	c := Get()
	return c.Notification
}

// GetPaths returns the paths configuration
func GetPaths() PathsConfig {
	c := Get()
	return c.Paths
}

// GetNetwork returns the network configuration
func GetNetwork() NetworkConfig {
	c := Get()
	return c.Network
}

// GetExternal returns the external tools configuration
func GetExternal() ExternalTools {
	c := Get()
	return c.External
}

// GetNotificationTimeout returns the notification timeout as a Duration
func (c NotificationConfig) GetTimeout() time.Duration {
	return time.Duration(c.DefaultTimeout) * time.Second
}

// GetScreenshotNotificationTimeout returns the screenshot notification timeout as a Duration
func (c ScreenshotConfig) GetNotificationTimeout() time.Duration {
	return time.Duration(c.NotificationTimeout) * time.Second
}

// GetEmojiDownloadTimeout returns the emoji download timeout as a Duration
func (c EmojiConfig) GetDownloadTimeout() time.Duration {
	return time.Duration(c.DownloadTimeout) * time.Second
}

// GetShellIPCTimeout returns the shell IPC timeout as a Duration
func (c ShellConfig) GetIPCTimeout() time.Duration {
	return time.Duration(c.IPCTimeout) * time.Second
}

// GetNetworkIPCTimeout returns the network IPC timeout as a Duration
func (c NetworkConfig) GetIPCTimeout() time.Duration {
	return time.Duration(c.IPCTimeout) * time.Second
}

// GetHyprIPCTimeout returns the Hypr IPC timeout as a Duration
func (c NetworkConfig) GetHyprIPCTimeout() time.Duration {
	return time.Duration(c.HyprIPCTimeout) * time.Second
}
