# Heimdall CLI Configuration Quick Reference

A quick reference guide for common configuration options.

## Common Configurations

### Set Default Color Scheme

```json
{
  "scheme": {
    "default": "catppuccin-mocha"
  }
}
```

### Enable Material You Theming

```json
{
  "scheme": {
    "materialYou": true
  },
  "wallpaper": {
    "generateMaterialYou": true
  }
}
```

### Disable Specific Applications

```json
{
  "theme": {
    "enableGtk": false,
    "enableDiscord": false
  }
}
```

### Configure Idle Detection

```json
{
  "idle": {
    "enabled": true,
    "timeout": 300,
    "scheme": "rosepine",
    "wallpaper": "/path/to/idle-wallpaper.jpg"
  }
}
```

## All Configuration Options

| Path | Type | Default | Description |
|------|------|---------|-------------|
| `clipboard.delete_on_select` | bool | false | Remove entry from history after selection |
| `clipboard.fuzzel_args` | []string | ["--dmenu", "--wi... | Additional arguments for Fuzzel launcher |
| `clipboard.fuzzel_prompt` | string | Clipboard>  | Prompt text for clipboard picker |
| `clipboard.max_entries` | int | 100 | Maximum number of clipboard history entries |
| `clipboard.preview_length` | int | 50 | Maximum characters to show in preview |
| `emoji.copy_to_clipboard` | bool | true | Copy selected emoji to clipboard |
| `emoji.data_directory` | string | - | Directory for emoji data files |
| `emoji.download_timeout` | int | 30 | Timeout for downloading emoji data in seconds |
| `emoji.fuzzel_args` | []string | ["--dmenu", "--pr... | Additional arguments for Fuzzel launcher |
| `emoji.fuzzel_prompt` | string | Emoji>  | Prompt text for emoji picker |
| `emoji.sources` | []string | ["emoji.json"] | Emoji data source files to use |
| `emoji.type_directly` | bool | false | Type emoji directly into active window |
| `external_tools.app2unit` | string | app2unit | Path to app2unit systemd integration tool |
| `external_tools.cliphist` | string | cliphist | Path to cliphist clipboard manager |
| `external_tools.dart_sass` | string | sass | Path to Dart Sass compiler |
| `external_tools.dunstify` | string | dunstify | Path to dunstify notification tool |
| `external_tools.fuzzel` | string | fuzzel | Path to fuzzel launcher |
| `external_tools.gdbus` | string | gdbus | Path to gdbus D-Bus tool |
| `external_tools.grim` | string | grim | Path to grim screenshot tool |
| `external_tools.libnotify` | string | notify-send | Path to notify-send notification tool |
| `external_tools.pactl` | string | pactl | Path to PulseAudio control utility |
| `external_tools.pidof` | string | pidof | Path to pidof process finder |
| `external_tools.pkill` | string | pkill | Path to pkill process killer |
| `external_tools.qs` | string | qs | Path to Quickshell executable |
| `external_tools.slurp` | string | slurp | Path to slurp selection tool |
| `external_tools.swappy` | string | swappy | Path to swappy screenshot editor |
| `external_tools.wl_clipboard` | string | wl-copy | Path to wl-copy clipboard tool |
| `external_tools.wl_screenrec` | string | wl-screenrec | Path to wl-screenrec recording tool |
| `external_tools.xclip` | string | xclip | Path to xclip X11 clipboard tool |
| `migrated_from` | string | - | Previous version this config was migrated from |
| `network.hypr_ipc_timeout` | int | 5 | Hyprland IPC timeout in seconds |
| `network.ipc_timeout` | int | 5 | General IPC timeout in seconds |
| `notification.app_name` | string | heimdall | Application name shown in notifications |
| `notification.default_timeout` | int | 5 | Default notification timeout in seconds |
| `notification.default_urgency` | string | normal | Default notification urgency (low, normal, critical) |
| `notification.enabled` | bool | true | Enable system notifications |
| `notification.provider` | string | auto | Notification provider (notify-send, dunstify, auto) |
| `paths.cache_dir` | string | - | Directory for cache files |
| `paths.data_dir` | string | - | Directory for application data |
| `paths.schemes` | string | - | Custom directory for color schemes |
| `paths.state_dir` | string | - | Directory for state files and runtime data |
| `paths.templates` | string | - | Custom directory for theme templates |
| `pip.always_on_top` | bool | true | Keep PIP windows above other windows |
| `pip.enabled` | bool | true | Enable picture-in-picture mode |
| `pip.pid_file` | string | pip.pid | PID file name for PIP process tracking |
| `pip.pin_windows` | bool | true | Pin PIP windows to all workspaces |
| `pip.video_apps` | []string | ["mpv", "vlc", "f... | Applications to detect for PIP mode |
| `pip.video_keywords` | []string | ["youtube", "netf... | Window title keywords to detect video playback |
| `pip.window_position` | string | bottom-right | PIP window position on screen |
| `pip.window_size` | string | 25% | PIP window size as percentage of screen |
| `recording.audio_source` | string | auto | Audio source (auto, none, or specific device) |
| `recording.directory` | string | - | Directory to save screen recordings |
| `recording.file_format` | string | mp4 | Video format (mp4, webm, mkv) |
| `recording.file_name_pattern` | string | recording_%Y%m%d_... | Filename pattern with date format codes |
| `recording.show_notification` | bool | true | Show notification when recording starts/stops |
| `recording.temp_file_name` | string | recording.mp4 | Temporary filename during recording |
| `scheme.auto_mode` | bool | true | Automatically switch between light/dark variants based on... |
| `scheme.default` | string | rosepine | Default color scheme to use |
| `scheme.generated_path` | string | - | Directory for storing generated Material You schemes |
| `scheme.material_you` | bool | true | Generate Material You color schemes from wallpapers |
| `scheme.user_paths` | []string | - | Additional directories to search for user-defined schemes |
| `screenshot.copy_to_clipboard` | bool | true | Copy screenshot to clipboard after capture |
| `screenshot.directory` | string | - | Directory to save screenshots |
| `screenshot.file_format` | string | png | Image format (png, jpg, webp) |
| `screenshot.file_name_pattern` | string | screenshot_%Y%m%d... | Filename pattern with date format codes |
| `screenshot.freeze_file_name` | string | freeze.png | Temporary filename for freeze screenshots |
| `screenshot.notification_timeout` | int | 3 | Notification display duration in seconds |
| `screenshot.open_with_swappy` | bool | true | Open screenshot in Swappy editor after capture |
| `screenshot.show_notification` | bool | true | Show notification after screenshot capture |
| `shell.args` | []string | ["-c", "heimdall"... | Arguments to pass to Quickshell |
| `shell.command` | string | qs | Quickshell executable command |
| `shell.daemon_port` | int | 9999 | Port for Quickshell daemon IPC |
| `shell.ipc_timeout` | int | 5 | IPC timeout in seconds |
| `shell.log_file` | string | shell.log | Log file name for shell output |
| `shell.log_rules` | string | - | Logging rules for Quickshell (Qt logging format) |
| `shell.pid_file` | string | shell.pid | PID file name for shell process |
| `theme.enableAlacritty` | bool | false | Apply themes to Alacritty terminal emulator |
| `theme.enableBtop` | bool | true | Apply themes to btop++ system monitor |
| `theme.enableDiscord` | bool | true | Apply themes to Discord clients (Vesktop, Discord, Vencor... |
| `theme.enableFuzzel` | bool | true | Apply themes to Fuzzel launcher |
| `theme.enableGtk` | bool | true | Apply themes to GTK 3 and GTK 4 applications |
| `theme.enableHypr` | bool | true | Apply themes to Hyprland window manager configuration |
| `theme.enableKitty` | bool | true | Apply themes to Kitty terminal emulator |
| `theme.enableNvim` | bool | true | Apply themes to Neovim editor (LazyVim integration) |
| `theme.enableQt` | bool | true | Apply themes to Qt5 and Qt6 applications via qt5ct/qt6ct |
| `theme.enableSpicetify` | bool | true | Apply themes to Spotify via Spicetify |
| `theme.enableTerm` | bool | true | Apply themes to terminal emulators via escape sequences |
| `theme.enableWezterm` | bool | false | Apply themes to WezTerm terminal emulator |
| `theme.paths.alacritty` | string | - | Path to Alacritty theme TOML file |
| `theme.paths.betterDiscord` | string | - | Path to BetterDiscord theme CSS file |
| `theme.paths.btop` | string | - | Path to btop++ theme file |
| `theme.paths.discord` | string | - | Path to Discord theme CSS file |
| `theme.paths.discordCanary` | string | - | Path to Discord Canary theme CSS file |
| `theme.paths.equicord` | string | - | Path to Equicord theme CSS file |
| `theme.paths.fuzzel` | string | - | Path to Fuzzel launcher colors configuration |
| `theme.paths.gtk3` | string | - | Path to GTK 3 theme colors CSS file |
| `theme.paths.gtk4` | string | - | Path to GTK 4 theme colors CSS file |
| `theme.paths.kitty` | string | - | Path to Kitty terminal theme configuration |
| `theme.paths.nvim` | string | - | Path to Neovim LazyVim theme plugin file |
| `theme.paths.qt5` | string | - | Path to Qt5ct color scheme file |
| `theme.paths.qt6` | string | - | Path to Qt6ct color scheme file |
| `theme.paths.spicetify` | string | - | Path to Spicetify theme color.ini file |
| `theme.paths.terminal` | string | - | Path to terminal escape sequences file |
| `theme.paths.vencord` | string | - | Path to Vencord theme CSS file |
| `theme.paths.vesktop` | string | - | Path to Vesktop theme CSS file |
| `theme.paths.wezterm` | string | - | Path to WezTerm color scheme Lua file |
| `toggles` | map[string]object | - | Workspace-specific application toggle configurations |
| `version` | string | 0.2.0 | Configuration version for migration and compatibility che... |
| `wallpaper.directory` | string | - | Directory containing wallpaper images |
| `wallpaper.extensions` | []string | [".jpg", ".jpeg",... | Supported image file extensions |
| `wallpaper.filter` | bool | true | Filter wallpapers based on color similarity to current sc... |
| `wallpaper.smart_mode` | bool | true | Use intelligent wallpaper selection based on scheme colors |
| `wallpaper.threshold` | float | 0.8 | Color similarity threshold for filtering (0.0-1.0, higher... |

## Useful Commands

```bash
# Show all configuration options
heimdall config list

# Search for specific options
heimdall config search theme

# Show current effective configuration
heimdall config effective

# Show only modified values
heimdall config list --modified

# Describe a specific option
heimdall config describe scheme.default

# Validate configuration
heimdall config validate
```

