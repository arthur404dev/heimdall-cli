package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/viper"
)

// Config represents the main configuration structure
type Config struct {
	Version      string                  `mapstructure:"version" json:"version" yaml:"version" desc:"Configuration version for migration and compatibility checking" default:"0.2.0" example:"0.2.0"`
	MigratedFrom string                  `mapstructure:"migrated_from,omitempty" json:"migrated_from,omitempty" yaml:"migrated_from,omitempty" desc:"Previous version this config was migrated from" example:"0.1.0"`
	Theme        ThemeConfig             `mapstructure:"theme" json:"theme" yaml:"theme" desc:"Theme application settings for various applications"`
	Toggles      map[string]ToggleConfig `mapstructure:"toggles" json:"toggles" yaml:"toggles" desc:"Workspace-specific application toggle configurations"`
	Shell        ShellConfig             `mapstructure:"shell" json:"shell" yaml:"shell" desc:"Quickshell daemon configuration for the UI shell"`
	Scheme       SchemeConfig            `mapstructure:"scheme" json:"scheme" yaml:"scheme" desc:"Color scheme settings and Material You configuration"`
	Wallpaper    WallpaperConfig         `mapstructure:"wallpaper" json:"wallpaper" yaml:"wallpaper" desc:"Wallpaper management and filtering settings"`
	Screenshot   ScreenshotConfig        `mapstructure:"screenshot" json:"screenshot" yaml:"screenshot" desc:"Screenshot capture and processing settings"`
	Recording    RecordingConfig         `mapstructure:"recording" json:"recording" yaml:"recording" desc:"Screen recording configuration"`
	Clipboard    ClipboardConfig         `mapstructure:"clipboard" json:"clipboard" yaml:"clipboard" desc:"Clipboard history and management settings"`
	Emoji        EmojiConfig             `mapstructure:"emoji" json:"emoji" yaml:"emoji" desc:"Emoji picker configuration and data sources"`
	PIP          PIPConfig               `mapstructure:"pip" json:"pip" yaml:"pip" desc:"Picture-in-Picture window management settings"`
	Notification NotificationConfig      `mapstructure:"notification" json:"notification" yaml:"notification" desc:"System notification preferences and providers"`
	Paths        PathsConfig             `mapstructure:"paths" json:"paths" yaml:"paths" desc:"Custom paths for templates, schemes, and data directories"`
	Network      NetworkConfig           `mapstructure:"network" json:"network" yaml:"network" desc:"Network and IPC timeout configurations"`
	External     ExternalTools           `mapstructure:"external_tools" json:"external_tools" yaml:"external_tools" desc:"External tool paths and command overrides"`
}

// ThemeConfig represents theme configuration
type ThemeConfig struct {
	EnableTerm      bool             `mapstructure:"enableTerm" json:"enableTerm" yaml:"enableTerm" desc:"Apply themes to terminal emulators via escape sequences" default:"true" example:"true"`
	EnableHypr      bool             `mapstructure:"enableHypr" json:"enableHypr" yaml:"enableHypr" desc:"Apply themes to Hyprland window manager configuration" default:"true" example:"true"`
	EnableDiscord   bool             `mapstructure:"enableDiscord" json:"enableDiscord" yaml:"enableDiscord" desc:"Apply themes to Discord clients (Vesktop, Discord, Vencord, etc.)" default:"true" example:"true"`
	EnableSpicetify bool             `mapstructure:"enableSpicetify" json:"enableSpicetify" yaml:"enableSpicetify" desc:"Apply themes to Spotify via Spicetify" default:"true" example:"false"`
	EnableFuzzel    bool             `mapstructure:"enableFuzzel" json:"enableFuzzel" yaml:"enableFuzzel" desc:"Apply themes to Fuzzel launcher" default:"true" example:"true"`
	EnableBtop      bool             `mapstructure:"enableBtop" json:"enableBtop" yaml:"enableBtop" desc:"Apply themes to btop++ system monitor" default:"true" example:"true"`
	EnableGtk       bool             `mapstructure:"enableGtk" json:"enableGtk" yaml:"enableGtk" desc:"Apply themes to GTK 3 and GTK 4 applications" default:"true" example:"true"`
	EnableQt        bool             `mapstructure:"enableQt" json:"enableQt" yaml:"enableQt" desc:"Apply themes to Qt5 and Qt6 applications via qt5ct/qt6ct" default:"true" example:"false"`
	EnableKitty     bool             `mapstructure:"enableKitty" json:"enableKitty" yaml:"enableKitty" desc:"Apply themes to Kitty terminal emulator" default:"true" example:"true"`
	EnableAlacritty bool             `mapstructure:"enableAlacritty" json:"enableAlacritty" yaml:"enableAlacritty" desc:"Apply themes to Alacritty terminal emulator" default:"false" example:"true"`
	EnableWezterm   bool             `mapstructure:"enableWezterm" json:"enableWezterm" yaml:"enableWezterm" desc:"Apply themes to WezTerm terminal emulator" default:"false" example:"true"`
	EnableNvim      bool             `mapstructure:"enableNvim" json:"enableNvim" yaml:"enableNvim" desc:"Apply themes to Neovim editor (LazyVim integration)" default:"true" example:"true"`
	Paths           ThemePathsConfig `mapstructure:"paths" json:"paths" yaml:"paths" desc:"Custom paths for theme configuration files"`
}

// ThemePathsConfig represents custom paths for theme files
type ThemePathsConfig struct {
	Gtk3          string `mapstructure:"gtk3" json:"gtk3" yaml:"gtk3" desc:"Path to GTK 3 theme colors CSS file" example:"~/.config/gtk-3.0/colors.css"`
	Gtk4          string `mapstructure:"gtk4" json:"gtk4" yaml:"gtk4" desc:"Path to GTK 4 theme colors CSS file" example:"~/.config/gtk-4.0/colors.css"`
	Qt5           string `mapstructure:"qt5" json:"qt5" yaml:"qt5" desc:"Path to Qt5ct color scheme file" example:"~/.config/qt5ct/colors/heimdall.conf"`
	Qt6           string `mapstructure:"qt6" json:"qt6" yaml:"qt6" desc:"Path to Qt6ct color scheme file" example:"~/.config/qt6ct/colors/heimdall.conf"`
	Btop          string `mapstructure:"btop" json:"btop" yaml:"btop" desc:"Path to btop++ theme file" example:"~/.config/btop/themes/heimdall.theme"`
	Fuzzel        string `mapstructure:"fuzzel" json:"fuzzel" yaml:"fuzzel" desc:"Path to Fuzzel launcher colors configuration" example:"~/.config/fuzzel/colors.ini"`
	Spicetify     string `mapstructure:"spicetify" json:"spicetify" yaml:"spicetify" desc:"Path to Spicetify theme color.ini file" example:"~/.config/spicetify/Themes/heimdall/color.ini"`
	Kitty         string `mapstructure:"kitty" json:"kitty" yaml:"kitty" desc:"Path to Kitty terminal theme configuration" example:"~/.config/kitty/themes/heimdall.conf"`
	Alacritty     string `mapstructure:"alacritty" json:"alacritty" yaml:"alacritty" desc:"Path to Alacritty theme TOML file" example:"~/.config/alacritty/themes/heimdall.toml"`
	Wezterm       string `mapstructure:"wezterm" json:"wezterm" yaml:"wezterm" desc:"Path to WezTerm color scheme Lua file" example:"~/.config/wezterm/colors/heimdall.lua"`
	Nvim          string `mapstructure:"nvim" json:"nvim" yaml:"nvim" desc:"Path to Neovim LazyVim theme plugin file" example:"~/.config/nvim/lua/user/heimdall.lua"`
	Terminal      string `mapstructure:"terminal" json:"terminal" yaml:"terminal" desc:"Path to terminal escape sequences file" example:"~/.config/heimdall/sequences.txt"`
	Vesktop       string `mapstructure:"vesktop" json:"vesktop" yaml:"vesktop" desc:"Path to Vesktop theme CSS file" example:"~/.config/vesktop/themes/heimdall.css"`
	Discord       string `mapstructure:"discord" json:"discord" yaml:"discord" desc:"Path to Discord theme CSS file" example:"~/.config/discord/themes/heimdall.css"`
	DiscordCanary string `mapstructure:"discordCanary" json:"discordCanary" yaml:"discordCanary" desc:"Path to Discord Canary theme CSS file" example:"~/.config/discordcanary/themes/heimdall.css"`
	Vencord       string `mapstructure:"vencord" json:"vencord" yaml:"vencord" desc:"Path to Vencord theme CSS file" example:"~/.config/Vencord/themes/heimdall.css"`
	Equicord      string `mapstructure:"equicord" json:"equicord" yaml:"equicord" desc:"Path to Equicord theme CSS file" example:"~/.config/Equicord/themes/heimdall.css"`
	BetterDiscord string `mapstructure:"betterDiscord" json:"betterDiscord" yaml:"betterDiscord" desc:"Path to BetterDiscord theme CSS file" example:"~/.config/BetterDiscord/themes/heimdall.theme.css"`
}

// ToggleConfig represents workspace toggle configuration
type ToggleConfig struct {
	Apps map[string]AppConfig `mapstructure:"apps" json:"apps" yaml:"apps" desc:"Application-specific toggle configurations for this workspace"`
}

// AppConfig represents an application configuration for toggles
type AppConfig struct {
	Enable  bool             `mapstructure:"enable" json:"enable" yaml:"enable" desc:"Whether to enable this application toggle" default:"true" example:"true"`
	Match   []map[string]any `mapstructure:"match" json:"match" yaml:"match" desc:"Window matching rules (class, title, etc.)" example:"[{\"class\": \"firefox\"}]"`
	Command []string         `mapstructure:"command" json:"command" yaml:"command" desc:"Command to launch the application" example:"[\"firefox\", \"--new-window\"]"`
	Move    bool             `mapstructure:"move" json:"move" yaml:"move" desc:"Move existing windows to current workspace" default:"false" example:"true"`
}

// ShellConfig represents shell configuration
type ShellConfig struct {
	Command    string   `mapstructure:"command" json:"command" yaml:"command" desc:"Quickshell executable command" default:"qs" example:"qs"`
	Args       []string `mapstructure:"args" json:"args" yaml:"args" desc:"Arguments to pass to Quickshell" default:"[\"-c\", \"heimdall\", \"-n\"]" example:"[\"-c\", \"heimdall\", \"-n\"]"`
	LogRules   string   `mapstructure:"log_rules" json:"log_rules" yaml:"log_rules" desc:"Logging rules for Quickshell (Qt logging format)" example:"*.debug=false"`
	DaemonPort int      `mapstructure:"daemon_port" json:"daemon_port" yaml:"daemon_port" desc:"Port for Quickshell daemon IPC" default:"9999" example:"9999"`
	LogFile    string   `mapstructure:"log_file" json:"log_file" yaml:"log_file" desc:"Log file name for shell output" default:"shell.log" example:"shell.log"`
	PidFile    string   `mapstructure:"pid_file" json:"pid_file" yaml:"pid_file" desc:"PID file name for shell process" default:"shell.pid" example:"shell.pid"`
	IPCTimeout int      `mapstructure:"ipc_timeout" json:"ipc_timeout" yaml:"ipc_timeout" desc:"IPC timeout in seconds" default:"5" example:"10"`
}

// SchemeConfig represents scheme configuration
type SchemeConfig struct {
	Default       string   `mapstructure:"default" json:"default" yaml:"default" desc:"Default color scheme to use" default:"rosepine" example:"catppuccin-mocha"`
	AutoMode      bool     `mapstructure:"auto_mode" json:"auto_mode" yaml:"auto_mode" desc:"Automatically switch between light/dark variants based on time" default:"true" example:"true"`
	MaterialYou   bool     `mapstructure:"material_you" json:"material_you" yaml:"material_you" desc:"Generate Material You color schemes from wallpapers" default:"true" example:"false"`
	UserPaths     []string `mapstructure:"user_paths" json:"user_paths" yaml:"user_paths" desc:"Additional directories to search for user-defined schemes" example:"[\"~/.config/heimdall/schemes\", \"~/custom-schemes\"]"`
	GeneratedPath string   `mapstructure:"generated_path" json:"generated_path" yaml:"generated_path" desc:"Directory for storing generated Material You schemes" example:"~/.local/share/heimdall/schemes"`
}

// WallpaperConfig represents wallpaper configuration
type WallpaperConfig struct {
	Directory  string   `mapstructure:"directory" json:"directory" yaml:"directory" desc:"Directory containing wallpaper images" example:"~/Pictures/Wallpapers"`
	Filter     bool     `mapstructure:"filter" json:"filter" yaml:"filter" desc:"Filter wallpapers based on color similarity to current scheme" default:"true" example:"false"`
	Threshold  float64  `mapstructure:"threshold" json:"threshold" yaml:"threshold" desc:"Color similarity threshold for filtering (0.0-1.0, higher = stricter)" default:"0.8" example:"0.7"`
	SmartMode  bool     `mapstructure:"smart_mode" json:"smart_mode" yaml:"smart_mode" desc:"Use intelligent wallpaper selection based on scheme colors" default:"true" example:"true"`
	Extensions []string `mapstructure:"extensions" json:"extensions" yaml:"extensions" desc:"Supported image file extensions" default:"[\".jpg\", \".jpeg\", \".png\", \".webp\"]" example:"[\".jpg\", \".png\"]"`
}

// ScreenshotConfig represents screenshot configuration
type ScreenshotConfig struct {
	Directory           string `mapstructure:"directory" json:"directory" yaml:"directory" desc:"Directory to save screenshots" example:"~/Pictures/Screenshots"`
	FileFormat          string `mapstructure:"file_format" json:"file_format" yaml:"file_format" desc:"Image format (png, jpg, webp)" default:"png" example:"webp"`
	FileNamePattern     string `mapstructure:"file_name_pattern" json:"file_name_pattern" yaml:"file_name_pattern" desc:"Filename pattern with date format codes" default:"screenshot_%Y%m%d_%H%M%S" example:"screen_%Y-%m-%d_%H-%M-%S"`
	CopyToClipboard     bool   `mapstructure:"copy_to_clipboard" json:"copy_to_clipboard" yaml:"copy_to_clipboard" desc:"Copy screenshot to clipboard after capture" default:"true" example:"false"`
	OpenWithSwappy      bool   `mapstructure:"open_with_swappy" json:"open_with_swappy" yaml:"open_with_swappy" desc:"Open screenshot in Swappy editor after capture" default:"true" example:"false"`
	ShowNotification    bool   `mapstructure:"show_notification" json:"show_notification" yaml:"show_notification" desc:"Show notification after screenshot capture" default:"true" example:"true"`
	NotificationTimeout int    `mapstructure:"notification_timeout" json:"notification_timeout" yaml:"notification_timeout" desc:"Notification display duration in seconds" default:"3" example:"5"`
	FreezeFileName      string `mapstructure:"freeze_file_name" json:"freeze_file_name" yaml:"freeze_file_name" desc:"Temporary filename for freeze screenshots" default:"freeze.png" example:"temp_freeze.png"`
}

// RecordingConfig represents recording configuration
type RecordingConfig struct {
	Directory        string `mapstructure:"directory" json:"directory" yaml:"directory" desc:"Directory to save screen recordings" example:"~/Videos/Recordings"`
	FileFormat       string `mapstructure:"file_format" json:"file_format" yaml:"file_format" desc:"Video format (mp4, webm, mkv)" default:"mp4" example:"webm"`
	FileNamePattern  string `mapstructure:"file_name_pattern" json:"file_name_pattern" yaml:"file_name_pattern" desc:"Filename pattern with date format codes" default:"recording_%Y%m%d_%H%M%S" example:"rec_%Y-%m-%d_%H-%M-%S"`
	TempFileName     string `mapstructure:"temp_file_name" json:"temp_file_name" yaml:"temp_file_name" desc:"Temporary filename during recording" default:"recording.mp4" example:"temp_recording.mp4"`
	ShowNotification bool   `mapstructure:"show_notification" json:"show_notification" yaml:"show_notification" desc:"Show notification when recording starts/stops" default:"true" example:"false"`
	AudioSource      string `mapstructure:"audio_source" json:"audio_source" yaml:"audio_source" desc:"Audio source (auto, none, or specific device)" default:"auto" example:"none"`
}

// ClipboardConfig represents clipboard configuration
type ClipboardConfig struct {
	MaxEntries     int      `mapstructure:"max_entries" json:"max_entries" yaml:"max_entries" desc:"Maximum number of clipboard history entries" default:"100" example:"200"`
	FuzzelPrompt   string   `mapstructure:"fuzzel_prompt" json:"fuzzel_prompt" yaml:"fuzzel_prompt" desc:"Prompt text for clipboard picker" default:"Clipboard> " example:"📋 Select: "`
	FuzzelArgs     []string `mapstructure:"fuzzel_args" json:"fuzzel_args" yaml:"fuzzel_args" desc:"Additional arguments for Fuzzel launcher" default:"[\"--dmenu\", \"--width\", \"50\", \"--lines\", \"20\"]" example:"[\"--dmenu\", \"--width\", \"60\"]"`
	PreviewLength  int      `mapstructure:"preview_length" json:"preview_length" yaml:"preview_length" desc:"Maximum characters to show in preview" default:"50" example:"80"`
	DeleteOnSelect bool     `mapstructure:"delete_on_select" json:"delete_on_select" yaml:"delete_on_select" desc:"Remove entry from history after selection" default:"false" example:"true"`
}

// EmojiConfig represents emoji configuration
type EmojiConfig struct {
	DataDirectory   string   `mapstructure:"data_directory" json:"data_directory" yaml:"data_directory" desc:"Directory for emoji data files" example:"~/.local/share/heimdall/emoji"`
	Sources         []string `mapstructure:"sources" json:"sources" yaml:"sources" desc:"Emoji data source files to use" default:"[\"emoji.json\"]" example:"[\"emoji.json\", \"custom.json\"]"`
	FuzzelPrompt    string   `mapstructure:"fuzzel_prompt" json:"fuzzel_prompt" yaml:"fuzzel_prompt" desc:"Prompt text for emoji picker" default:"Emoji> " example:"😀 Pick: "`
	FuzzelArgs      []string `mapstructure:"fuzzel_args" json:"fuzzel_args" yaml:"fuzzel_args" desc:"Additional arguments for Fuzzel launcher" default:"[\"--dmenu\", \"--prompt\"]" example:"[\"--dmenu\", \"--width\", \"40\"]"`
	CopyToClipboard bool     `mapstructure:"copy_to_clipboard" json:"copy_to_clipboard" yaml:"copy_to_clipboard" desc:"Copy selected emoji to clipboard" default:"true" example:"false"`
	TypeDirectly    bool     `mapstructure:"type_directly" json:"type_directly" yaml:"type_directly" desc:"Type emoji directly into active window" default:"false" example:"true"`
	DownloadTimeout int      `mapstructure:"download_timeout" json:"download_timeout" yaml:"download_timeout" desc:"Timeout for downloading emoji data in seconds" default:"30" example:"60"`
}

// PIPConfig represents picture-in-picture configuration
type PIPConfig struct {
	Enabled        bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled" desc:"Enable picture-in-picture mode" default:"true" example:"false"`
	PidFile        string   `mapstructure:"pid_file" json:"pid_file" yaml:"pid_file" desc:"PID file name for PIP process tracking" default:"pip.pid" example:"pip.pid"`
	WindowSize     string   `mapstructure:"window_size" json:"window_size" yaml:"window_size" desc:"PIP window size as percentage of screen" default:"25%" example:"30%"`
	WindowPosition string   `mapstructure:"window_position" json:"window_position" yaml:"window_position" desc:"PIP window position on screen" default:"bottom-right" example:"top-left"`
	VideoApps      []string `mapstructure:"video_apps" json:"video_apps" yaml:"video_apps" desc:"Applications to detect for PIP mode" default:"[\"mpv\", \"vlc\", \"firefox\", \"chromium\", \"chrome\", \"brave\", \"youtube\", \"netflix\", \"twitch\", \"spotify\"]" example:"[\"firefox\", \"mpv\"]"`
	VideoKeywords  []string `mapstructure:"video_keywords" json:"video_keywords" yaml:"video_keywords" desc:"Window title keywords to detect video playback" default:"[\"youtube\", \"netflix\", \"twitch\", \"vimeo\", \"- playing\", \"▶\", \"►\", \"video\", \"stream\"]" example:"[\"youtube\", \"video\"]"`
	PinWindows     bool     `mapstructure:"pin_windows" json:"pin_windows" yaml:"pin_windows" desc:"Pin PIP windows to all workspaces" default:"true" example:"false"`
	AlwaysOnTop    bool     `mapstructure:"always_on_top" json:"always_on_top" yaml:"always_on_top" desc:"Keep PIP windows above other windows" default:"true" example:"true"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled" desc:"Enable system notifications" default:"true" example:"false"`
	Provider       string `mapstructure:"provider" json:"provider" yaml:"provider" desc:"Notification provider (notify-send, dunstify, auto)" default:"auto" example:"dunstify"`
	DefaultTimeout int    `mapstructure:"default_timeout" json:"default_timeout" yaml:"default_timeout" desc:"Default notification timeout in seconds" default:"5" example:"10"`
	AppName        string `mapstructure:"app_name" json:"app_name" yaml:"app_name" desc:"Application name shown in notifications" default:"heimdall" example:"Heimdall CLI"`
	DefaultUrgency string `mapstructure:"default_urgency" json:"default_urgency" yaml:"default_urgency" desc:"Default notification urgency (low, normal, critical)" default:"normal" example:"low"`
}

// PathsConfig represents custom paths configuration
type PathsConfig struct {
	Templates string `mapstructure:"templates" json:"templates" yaml:"templates" desc:"Custom directory for theme templates" example:"~/.config/heimdall/templates"`
	Schemes   string `mapstructure:"schemes" json:"schemes" yaml:"schemes" desc:"Custom directory for color schemes" example:"~/.config/heimdall/schemes"`
	StateDir  string `mapstructure:"state_dir" json:"state_dir" yaml:"state_dir" desc:"Directory for state files and runtime data" example:"~/.local/state/heimdall"`
	CacheDir  string `mapstructure:"cache_dir" json:"cache_dir" yaml:"cache_dir" desc:"Directory for cache files" example:"~/.cache/heimdall"`
	DataDir   string `mapstructure:"data_dir" json:"data_dir" yaml:"data_dir" desc:"Directory for application data" example:"~/.local/share/heimdall"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	IPCTimeout     int `mapstructure:"ipc_timeout" json:"ipc_timeout" yaml:"ipc_timeout" desc:"General IPC timeout in seconds" default:"5" example:"10"`
	HyprIPCTimeout int `mapstructure:"hypr_ipc_timeout" json:"hypr_ipc_timeout" yaml:"hypr_ipc_timeout" desc:"Hyprland IPC timeout in seconds" default:"5" example:"3"`
}

// ExternalTools represents external tool paths
type ExternalTools struct {
	Grim        string `mapstructure:"grim" json:"grim" yaml:"grim" desc:"Path to grim screenshot tool" default:"grim" example:"/usr/bin/grim"`
	Slurp       string `mapstructure:"slurp" json:"slurp" yaml:"slurp" desc:"Path to slurp selection tool" default:"slurp" example:"/usr/bin/slurp"`
	Swappy      string `mapstructure:"swappy" json:"swappy" yaml:"swappy" desc:"Path to swappy screenshot editor" default:"swappy" example:"/usr/bin/swappy"`
	WlClipboard string `mapstructure:"wl_clipboard" json:"wl_clipboard" yaml:"wl_clipboard" desc:"Path to wl-copy clipboard tool" default:"wl-copy" example:"/usr/bin/wl-copy"`
	WlScreenrec string `mapstructure:"wl_screenrec" json:"wl_screenrec" yaml:"wl_screenrec" desc:"Path to wl-screenrec recording tool" default:"wl-screenrec" example:"/usr/bin/wl-screenrec"`
	Cliphist    string `mapstructure:"cliphist" json:"cliphist" yaml:"cliphist" desc:"Path to cliphist clipboard manager" default:"cliphist" example:"/usr/bin/cliphist"`
	Fuzzel      string `mapstructure:"fuzzel" json:"fuzzel" yaml:"fuzzel" desc:"Path to fuzzel launcher" default:"fuzzel" example:"/usr/bin/fuzzel"`
	DartSass    string `mapstructure:"dart_sass" json:"dart_sass" yaml:"dart_sass" desc:"Path to Dart Sass compiler" default:"sass" example:"/usr/bin/sass"`
	Libnotify   string `mapstructure:"libnotify" json:"libnotify" yaml:"libnotify" desc:"Path to notify-send notification tool" default:"notify-send" example:"/usr/bin/notify-send"`
	Dunstify    string `mapstructure:"dunstify" json:"dunstify" yaml:"dunstify" desc:"Path to dunstify notification tool" default:"dunstify" example:"/usr/bin/dunstify"`
	Qs          string `mapstructure:"qs" json:"qs" yaml:"qs" desc:"Path to Quickshell executable" default:"qs" example:"/usr/bin/qs"`
	App2unit    string `mapstructure:"app2unit" json:"app2unit" yaml:"app2unit" desc:"Path to app2unit systemd integration tool" default:"app2unit" example:"/usr/bin/app2unit"`
	Xclip       string `mapstructure:"xclip" json:"xclip" yaml:"xclip" desc:"Path to xclip X11 clipboard tool" default:"xclip" example:"/usr/bin/xclip"`
	Pactl       string `mapstructure:"pactl" json:"pactl" yaml:"pactl" desc:"Path to PulseAudio control utility" default:"pactl" example:"/usr/bin/pactl"`
	Pidof       string `mapstructure:"pidof" json:"pidof" yaml:"pidof" desc:"Path to pidof process finder" default:"pidof" example:"/usr/bin/pidof"`
	Pkill       string `mapstructure:"pkill" json:"pkill" yaml:"pkill" desc:"Path to pkill process killer" default:"pkill" example:"/usr/bin/pkill"`
	Gdbus       string `mapstructure:"gdbus" json:"gdbus" yaml:"gdbus" desc:"Path to gdbus D-Bus tool" default:"gdbus" example:"/usr/bin/gdbus"`
}

// Global config instance
var cfg *Config
var userSetKeys = make(map[string]bool) // Track which keys are actually set by the user

// GetDefaults returns the default configuration values (exported)
func GetDefaults() *Config {
	cfg := getDefaults()
	return &cfg
}

// getDefaults returns the default configuration values
func getDefaults() Config {
	return Config{
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
			EnableKitty:     true,
			EnableAlacritty: false,
			EnableWezterm:   false,
			EnableNvim:      true,
			Paths: ThemePathsConfig{
				Gtk3:          filepath.Join(paths.ConfigDir, "gtk-3.0", "colors.css"),                        // GTK uses CSS, colors.css makes sense
				Gtk4:          filepath.Join(paths.ConfigDir, "gtk-4.0", "colors.css"),                        // GTK uses CSS, colors.css makes sense
				Qt5:           filepath.Join(paths.ConfigDir, "qt5ct", "colors", "heimdall.conf"),             // Qt5ct expects files in colors/ dir
				Qt6:           filepath.Join(paths.ConfigDir, "qt6ct", "colors", "heimdall.conf"),             // Qt6ct expects files in colors/ dir
				Btop:          filepath.Join(paths.ConfigDir, "btop", "themes", "heimdall.theme"),             // btop expects .theme files in themes/ dir
				Fuzzel:        filepath.Join(paths.ConfigDir, "fuzzel", "colors.ini"),                         // Fuzzel uses INI, colors.ini makes sense
				Spicetify:     filepath.Join(paths.ConfigDir, "spicetify", "Themes", "heimdall", "color.ini"), // Spicetify expects this structure
				Kitty:         filepath.Join(paths.ConfigDir, "kitty", "themes", "heimdall.conf"),             // Kitty can include from themes/ dir
				Alacritty:     filepath.Join(paths.ConfigDir, "alacritty", "themes", "heimdall.toml"),         // Alacritty can import from themes/ dir
				Wezterm:       filepath.Join(paths.ConfigDir, "wezterm", "colors", "heimdall.lua"),            // WezTerm color schemes go in colors/ dir
				Nvim:          filepath.Join(paths.ConfigDir, "nvim", "lua", "user", "heimdall.lua"),          // Neovim LazyVim plugin file
				Terminal:      filepath.Join(paths.ConfigDir, "heimdall", "sequences.txt"),                    // Our own sequences file
				Vesktop:       filepath.Join(paths.ConfigDir, "vesktop", "themes", "heimdall.css"),            // Discord clients use themes/ dir
				Discord:       filepath.Join(paths.ConfigDir, "discord", "themes", "heimdall.css"),
				DiscordCanary: filepath.Join(paths.ConfigDir, "discordcanary", "themes", "heimdall.css"),
				Vencord:       filepath.Join(paths.ConfigDir, "Vencord", "themes", "heimdall.css"),
				Equicord:      filepath.Join(paths.ConfigDir, "Equicord", "themes", "heimdall.css"),
				BetterDiscord: filepath.Join(paths.ConfigDir, "BetterDiscord", "themes", "heimdall.theme.css"),
			},
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
			Default:       "rosepine",
			AutoMode:      true,
			MaterialYou:   true,
			UserPaths:     []string{filepath.Join(paths.HeimdallConfigDir, "schemes")},
			GeneratedPath: filepath.Join(paths.DataDir, "schemes"),
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
}

// trackUserKeys recursively tracks all keys from user config
func trackUserKeys(settings map[string]interface{}, prefix string) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		userSetKeys[fullKey] = true

		// Recursively track nested maps
		if nested, ok := value.(map[string]interface{}); ok {
			trackUserKeys(nested, fullKey)
		}
	}
}

// IsUserSet checks if a config key was explicitly set by the user
func IsUserSet(key string) bool {
	return userSetKeys[key]
}

// Load loads the configuration from file
func Load() error {
	// Check if migration is needed first (from old formats)
	if CheckForMigration() {
		if err := MigrateConfig(); err != nil {
			// Log migration error but continue
			fmt.Fprintf(os.Stderr, "Warning: Failed to migrate old config: %v\n", err)
		}
	}

	// Clear userSetKeys map
	userSetKeys = make(map[string]bool)

	// Set defaults first - this ensures all fields have values
	setDefaults()

	// Set config type to JSON
	viper.SetConfigType("json")

	// Try to load from Heimdall config (JSON)
	heimdallConfig := filepath.Join(paths.HeimdallConfigDir, "config.json")

	configExists := paths.Exists(heimdallConfig)

	if configExists {
		viper.SetConfigFile(heimdallConfig)
		// Read existing config
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("failed to read config: %w", err)
			}
		}

		// Track which keys are actually in the user's config file
		// Read the file again to get only user-set keys
		userViper := viper.New()
		userViper.SetConfigType("json")
		userViper.SetConfigFile(heimdallConfig)
		if err := userViper.ReadInConfig(); err == nil {
			// Recursively track all keys from user config
			trackUserKeys(userViper.AllSettings(), "")
		}
	} else {
		// No config found, system will use defaults
		// We don't create a config file here - let the user explicitly save if they want
		viper.SetConfigFile(heimdallConfig)
	}

	// Check for environment variable override for scheme paths AFTER loading config
	if schemePaths := os.Getenv("HEIMDALL_SCHEME_PATHS"); schemePaths != "" {
		// Parse colon-separated paths
		userPaths := filepath.SplitList(schemePaths)
		viper.Set("scheme.user_paths", userPaths)
	}

	// Unmarshal into struct - this merges defaults with existing config
	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// NOTE: We no longer automatically save the config
	// This allows the system to work with no config file or partial configs
	// Users can explicitly save their config if they want to persist changes

	return nil
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		Load()
	}
	return cfg
}

// EffectiveConfig returns the complete merged configuration (user + defaults)
// This shows what the system is actually using at runtime
func EffectiveConfig() *Config {
	// Get() already returns the merged config since Load() applies defaults
	return Get()
}

// UserConfig returns only the user-specified configuration values
// This shows what's actually saved in the config file
func UserConfig() (map[string]interface{}, error) {
	configPath := filepath.Join(paths.HeimdallConfigDir, "config.json")

	// If no config file exists, return empty map
	if !paths.Exists(configPath) {
		return make(map[string]interface{}), nil
	}

	// Read the actual file content
	userViper := viper.New()
	userViper.SetConfigType("json")
	userViper.SetConfigFile(configPath)

	if err := userViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read user config: %w", err)
	}

	return userViper.AllSettings(), nil
}

// HasUserConfig checks if a user config file exists
func HasUserConfig() bool {
	configPath := filepath.Join(paths.HeimdallConfigDir, "config.json")
	return paths.Exists(configPath)
}

// Save saves the current configuration
// This now only saves user-modified values, not defaults
func Save() error {
	if cfg == nil {
		return fmt.Errorf("no configuration loaded")
	}

	configPath := filepath.Join(paths.HeimdallConfigDir, "config.json")

	// Ensure directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a new viper instance for saving only user values
	saveViper := viper.New()
	saveViper.SetConfigType("json")
	saveViper.SetConfigFile(configPath)

	// If config file exists, read it to preserve user values
	if paths.Exists(configPath) {
		if err := saveViper.ReadInConfig(); err != nil {
			// If we can't read existing config, start fresh
			saveViper = viper.New()
			saveViper.SetConfigType("json")
			saveViper.SetConfigFile(configPath)
		}
	}

	// Get the current viper settings (which has user overrides)
	// We'll only save values that differ from defaults
	defaults := getDefaults()

	// Helper function to check if value differs from default
	isDifferent := func(current, def interface{}) bool {
		return fmt.Sprintf("%v", current) != fmt.Sprintf("%v", def)
	}

	// Only set values that differ from defaults
	if isDifferent(cfg.Version, defaults.Version) {
		saveViper.Set("version", cfg.Version)
	}
	if isDifferent(cfg.MigratedFrom, defaults.MigratedFrom) && cfg.MigratedFrom != "" {
		saveViper.Set("migrated_from", cfg.MigratedFrom)
	}

	// For nested structs, we need to check each field
	// This preserves the minimal config approach
	if viper.IsSet("theme") {
		saveViper.Set("theme", viper.Get("theme"))
	}
	if viper.IsSet("shell") {
		saveViper.Set("shell", viper.Get("shell"))
	}
	if viper.IsSet("scheme") {
		saveViper.Set("scheme", viper.Get("scheme"))
	}
	if viper.IsSet("wallpaper") {
		saveViper.Set("wallpaper", viper.Get("wallpaper"))
	}
	if viper.IsSet("screenshot") {
		saveViper.Set("screenshot", viper.Get("screenshot"))
	}
	if viper.IsSet("recording") {
		saveViper.Set("recording", viper.Get("recording"))
	}
	if viper.IsSet("clipboard") {
		saveViper.Set("clipboard", viper.Get("clipboard"))
	}
	if viper.IsSet("emoji") {
		saveViper.Set("emoji", viper.Get("emoji"))
	}
	if viper.IsSet("pip") {
		saveViper.Set("pip", viper.Get("pip"))
	}
	if viper.IsSet("notification") {
		saveViper.Set("notification", viper.Get("notification"))
	}
	if viper.IsSet("paths") {
		saveViper.Set("paths", viper.Get("paths"))
	}
	if viper.IsSet("network") {
		saveViper.Set("network", viper.Get("network"))
	}
	if viper.IsSet("external") {
		saveViper.Set("external", viper.Get("external"))
	}
	if viper.IsSet("toggles") {
		saveViper.Set("toggles", viper.Get("toggles"))
	}

	// Write config
	if paths.Exists(configPath) {
		if err := saveViper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	} else {
		if err := saveViper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
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
	defaults := getDefaults()

	// Set top-level defaults
	viper.SetDefault("version", defaults.Version)
	viper.SetDefault("migrated_from", defaults.MigratedFrom)

	// Theme defaults - set each field individually for proper merging
	viper.SetDefault("theme.enableTerm", defaults.Theme.EnableTerm)
	viper.SetDefault("theme.enableHypr", defaults.Theme.EnableHypr)
	viper.SetDefault("theme.enableDiscord", defaults.Theme.EnableDiscord)
	viper.SetDefault("theme.enableSpicetify", defaults.Theme.EnableSpicetify)
	viper.SetDefault("theme.enableFuzzel", defaults.Theme.EnableFuzzel)
	viper.SetDefault("theme.enableBtop", defaults.Theme.EnableBtop)
	viper.SetDefault("theme.enableGtk", defaults.Theme.EnableGtk)
	viper.SetDefault("theme.enableQt", defaults.Theme.EnableQt)
	viper.SetDefault("theme.enableKitty", defaults.Theme.EnableKitty)
	viper.SetDefault("theme.enableAlacritty", defaults.Theme.EnableAlacritty)
	viper.SetDefault("theme.enableWezterm", defaults.Theme.EnableWezterm)
	viper.SetDefault("theme.enableNvim", defaults.Theme.EnableNvim)
	viper.SetDefault("theme.paths", defaults.Theme.Paths)

	// Shell defaults
	viper.SetDefault("shell.command", defaults.Shell.Command)
	viper.SetDefault("shell.args", defaults.Shell.Args)
	viper.SetDefault("shell.log_rules", defaults.Shell.LogRules)
	viper.SetDefault("shell.daemon_port", defaults.Shell.DaemonPort)
	viper.SetDefault("shell.log_file", defaults.Shell.LogFile)
	viper.SetDefault("shell.pid_file", defaults.Shell.PidFile)
	viper.SetDefault("shell.ipc_timeout", defaults.Shell.IPCTimeout)

	// Scheme defaults
	viper.SetDefault("scheme.default", defaults.Scheme.Default)
	viper.SetDefault("scheme.auto_mode", defaults.Scheme.AutoMode)
	viper.SetDefault("scheme.material_you", defaults.Scheme.MaterialYou)
	viper.SetDefault("scheme.user_paths", defaults.Scheme.UserPaths)
	viper.SetDefault("scheme.generated_path", defaults.Scheme.GeneratedPath)

	// Wallpaper defaults
	viper.SetDefault("wallpaper.directory", defaults.Wallpaper.Directory)
	viper.SetDefault("wallpaper.filter", defaults.Wallpaper.Filter)
	viper.SetDefault("wallpaper.threshold", defaults.Wallpaper.Threshold)
	viper.SetDefault("wallpaper.smart_mode", defaults.Wallpaper.SmartMode)
	viper.SetDefault("wallpaper.extensions", defaults.Wallpaper.Extensions)

	// Screenshot defaults
	viper.SetDefault("screenshot.directory", defaults.Screenshot.Directory)
	viper.SetDefault("screenshot.file_format", defaults.Screenshot.FileFormat)
	viper.SetDefault("screenshot.file_name_pattern", defaults.Screenshot.FileNamePattern)
	viper.SetDefault("screenshot.copy_to_clipboard", defaults.Screenshot.CopyToClipboard)
	viper.SetDefault("screenshot.open_with_swappy", defaults.Screenshot.OpenWithSwappy)
	viper.SetDefault("screenshot.show_notification", defaults.Screenshot.ShowNotification)
	viper.SetDefault("screenshot.notification_timeout", defaults.Screenshot.NotificationTimeout)
	viper.SetDefault("screenshot.freeze_file_name", defaults.Screenshot.FreezeFileName)

	// Recording defaults
	viper.SetDefault("recording.directory", defaults.Recording.Directory)
	viper.SetDefault("recording.file_format", defaults.Recording.FileFormat)
	viper.SetDefault("recording.file_name_pattern", defaults.Recording.FileNamePattern)
	viper.SetDefault("recording.temp_file_name", defaults.Recording.TempFileName)
	viper.SetDefault("recording.show_notification", defaults.Recording.ShowNotification)
	viper.SetDefault("recording.audio_source", defaults.Recording.AudioSource)

	// Clipboard defaults
	viper.SetDefault("clipboard.max_entries", defaults.Clipboard.MaxEntries)
	viper.SetDefault("clipboard.fuzzel_prompt", defaults.Clipboard.FuzzelPrompt)
	viper.SetDefault("clipboard.fuzzel_args", defaults.Clipboard.FuzzelArgs)
	viper.SetDefault("clipboard.preview_length", defaults.Clipboard.PreviewLength)
	viper.SetDefault("clipboard.delete_on_select", defaults.Clipboard.DeleteOnSelect)

	// Emoji defaults
	viper.SetDefault("emoji.data_directory", defaults.Emoji.DataDirectory)
	viper.SetDefault("emoji.sources", defaults.Emoji.Sources)
	viper.SetDefault("emoji.fuzzel_prompt", defaults.Emoji.FuzzelPrompt)
	viper.SetDefault("emoji.fuzzel_args", defaults.Emoji.FuzzelArgs)
	viper.SetDefault("emoji.copy_to_clipboard", defaults.Emoji.CopyToClipboard)
	viper.SetDefault("emoji.type_directly", defaults.Emoji.TypeDirectly)
	viper.SetDefault("emoji.download_timeout", defaults.Emoji.DownloadTimeout)

	// PIP defaults
	viper.SetDefault("pip.enabled", defaults.PIP.Enabled)
	viper.SetDefault("pip.pid_file", defaults.PIP.PidFile)
	viper.SetDefault("pip.window_size", defaults.PIP.WindowSize)
	viper.SetDefault("pip.window_position", defaults.PIP.WindowPosition)
	viper.SetDefault("pip.video_apps", defaults.PIP.VideoApps)
	viper.SetDefault("pip.video_keywords", defaults.PIP.VideoKeywords)
	viper.SetDefault("pip.pin_windows", defaults.PIP.PinWindows)
	viper.SetDefault("pip.always_on_top", defaults.PIP.AlwaysOnTop)

	// Notification defaults
	viper.SetDefault("notification.enabled", defaults.Notification.Enabled)
	viper.SetDefault("notification.provider", defaults.Notification.Provider)
	viper.SetDefault("notification.default_timeout", defaults.Notification.DefaultTimeout)
	viper.SetDefault("notification.app_name", defaults.Notification.AppName)
	viper.SetDefault("notification.default_urgency", defaults.Notification.DefaultUrgency)

	// Paths defaults
	viper.SetDefault("paths.templates", defaults.Paths.Templates)
	viper.SetDefault("paths.schemes", defaults.Paths.Schemes)
	viper.SetDefault("paths.state_dir", defaults.Paths.StateDir)
	viper.SetDefault("paths.cache_dir", defaults.Paths.CacheDir)
	viper.SetDefault("paths.data_dir", defaults.Paths.DataDir)

	// Network defaults
	viper.SetDefault("network.ipc_timeout", defaults.Network.IPCTimeout)
	viper.SetDefault("network.hypr_ipc_timeout", defaults.Network.HyprIPCTimeout)

	// External tools defaults
	viper.SetDefault("external.grim", defaults.External.Grim)
	viper.SetDefault("external.slurp", defaults.External.Slurp)
	viper.SetDefault("external.swappy", defaults.External.Swappy)
	viper.SetDefault("external.wl_clipboard", defaults.External.WlClipboard)
	viper.SetDefault("external.wl_screenrec", defaults.External.WlScreenrec)
	viper.SetDefault("external.cliphist", defaults.External.Cliphist)
	viper.SetDefault("external.fuzzel", defaults.External.Fuzzel)
	viper.SetDefault("external.dart_sass", defaults.External.DartSass)
	viper.SetDefault("external.libnotify", defaults.External.Libnotify)
	viper.SetDefault("external.dunstify", defaults.External.Dunstify)
	viper.SetDefault("external.qs", defaults.External.Qs)
	viper.SetDefault("external.app2unit", defaults.External.App2unit)
	viper.SetDefault("external.xclip", defaults.External.Xclip)
	viper.SetDefault("external.pactl", defaults.External.Pactl)
	viper.SetDefault("external.pidof", defaults.External.Pidof)
	viper.SetDefault("external.pkill", defaults.External.Pkill)
	viper.SetDefault("external.gdbus", defaults.External.Gdbus)
}

// Reload reloads the configuration from file
func Reload() error {
	return Load()
}

// Validate checks if the current configuration is valid
func Validate() error {
	c := Get()
	if c == nil {
		return fmt.Errorf("no configuration loaded")
	}

	var warnings []string
	var errors []string

	// Basic validation rules
	if c.Version == "" {
		errors = append(errors, "configuration version is required")
	}

	// Check for deprecated fields (using viper to check raw config)
	checkDeprecatedFields(&warnings)

	// Validate paths exist or can be created
	if c.Wallpaper.Directory != "" {
		expandedPath := c.Wallpaper.Directory
		if strings.HasPrefix(expandedPath, "~") {
			expandedPath = filepath.Join(os.Getenv("HOME"), expandedPath[1:])
		}
		if !paths.Exists(expandedPath) {
			warnings = append(warnings, fmt.Sprintf("Wallpaper directory does not exist: %s", c.Wallpaper.Directory))
		}
	}

	// Check for potentially problematic configurations
	if c.Theme.EnableGtk && c.Theme.EnableQt {
		// This is fine, but might cause conflicts
		warnings = append(warnings, "Both GTK and Qt theming are enabled. Ensure themes are compatible")
	}

	if len(c.Scheme.UserPaths) > 5 {
		warnings = append(warnings, fmt.Sprintf("You have %d user scheme paths. This may slow down scheme discovery", len(c.Scheme.UserPaths)))
	}

	// Validate numeric ranges
	if c.Clipboard.MaxEntries < 0 {
		errors = append(errors, "clipboard.max_entries must be non-negative")
	} else if c.Clipboard.MaxEntries > 10000 {
		warnings = append(warnings, "clipboard.max_entries is very high (>10000). This may impact performance")
	}

	if c.Clipboard.PreviewLength < 0 {
		errors = append(errors, "clipboard.preview_length must be non-negative")
	} else if c.Clipboard.PreviewLength > 500 {
		warnings = append(warnings, "clipboard.preview_length is very high (>500). Consider reducing for better UI")
	}

	if c.Screenshot.NotificationTimeout < 0 {
		errors = append(errors, "screenshot.notification_timeout must be non-negative")
	}

	if c.Wallpaper.Threshold < 0 || c.Wallpaper.Threshold > 100 {
		errors = append(errors, "wallpaper.threshold must be between 0 and 100")
	}

	// Validate file formats
	validImageFormats := []string{"png", "jpg", "jpeg", "webp"}
	if !contains(validImageFormats, c.Screenshot.FileFormat) {
		errors = append(errors, fmt.Sprintf("screenshot.file_format must be one of: %v", validImageFormats))
	}

	validVideoFormats := []string{"mp4", "webm", "mkv", "gif"}
	if !contains(validVideoFormats, c.Recording.FileFormat) {
		errors = append(errors, fmt.Sprintf("recording.file_format must be one of: %v", validVideoFormats))
	}

	// Validate notification urgency
	validUrgencies := []string{"low", "normal", "critical"}
	if !contains(validUrgencies, c.Notification.DefaultUrgency) {
		errors = append(errors, fmt.Sprintf("notification.default_urgency must be one of: %v", validUrgencies))
	}

	// Validate PIP window position
	validPositions := []string{"top-left", "top-right", "bottom-left", "bottom-right"}
	if !contains(validPositions, c.PIP.WindowPosition) {
		errors = append(errors, fmt.Sprintf("pip.window_position must be one of: %v", validPositions))
	}

	// Check for missing recommended tools
	checkExternalTools(c, &warnings)

	// Print warnings
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Configuration Warnings:\n")
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "   • %s\n", warning)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Return error if there are any errors
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  • %s", strings.Join(errors, "\n  • "))
	}

	return nil
}

// checkDeprecatedFields checks for deprecated configuration fields
func checkDeprecatedFields(warnings *[]string) {
	// Check for old field names that might still be in use
	deprecatedFields := map[string]string{
		"colorScheme":    "Use 'scheme.default' instead",
		"enableGTK":      "Use 'theme.enableGtk' instead",
		"enableQT":       "Use 'theme.enableQt' instead",
		"wallpaperDir":   "Use 'wallpaper.directory' instead",
		"wallpaperDirs":  "Use 'wallpaper.directories' instead",
		"enableHyprland": "Use 'theme.enableHypr' instead",
		"schemeDir":      "Use 'paths.schemes' instead",
	}

	for oldField, suggestion := range deprecatedFields {
		if viper.IsSet(oldField) {
			*warnings = append(*warnings, fmt.Sprintf("Deprecated field '%s' found. %s", oldField, suggestion))
		}
	}

	// Check for deprecated values
	if viper.GetString("notification.provider") == "notify-send" {
		*warnings = append(*warnings, "notification.provider 'notify-send' is deprecated. Use 'libnotify' instead")
	}
}

// checkExternalTools checks if recommended external tools are available
func checkExternalTools(c *Config, warnings *[]string) {
	// Check for critical tools based on enabled features
	if c.Screenshot.OpenWithSwappy {
		if _, err := exec.LookPath(c.External.Swappy); err != nil {
			*warnings = append(*warnings, fmt.Sprintf("Swappy is enabled but '%s' not found in PATH", c.External.Swappy))
		}
	}

	if c.Theme.EnableKitty {
		if _, err := exec.LookPath("kitty"); err != nil {
			*warnings = append(*warnings, "Kitty theming is enabled but 'kitty' not found in PATH")
		}
	}

	if c.Theme.EnableAlacritty {
		if _, err := exec.LookPath("alacritty"); err != nil {
			*warnings = append(*warnings, "Alacritty theming is enabled but 'alacritty' not found in PATH")
		}
	}

	if c.Notification.Enabled {
		provider := c.External.Libnotify
		if c.Notification.Provider == "dunstify" {
			provider = c.External.Dunstify
		}
		if _, err := exec.LookPath(provider); err != nil {
			*warnings = append(*warnings, fmt.Sprintf("Notifications enabled but '%s' not found in PATH", provider))
		}
	}
}

// EnsureConfigSaved ensures the configuration directory exists
// It no longer automatically saves defaults - users must explicitly save
func EnsureConfigSaved() error {
	// Ensure directory exists
	if err := paths.EnsureDir(paths.HeimdallConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// We no longer automatically save the config
	// The system works fine with defaults only
	return nil
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

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
