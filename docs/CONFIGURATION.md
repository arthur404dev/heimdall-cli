# Heimdall Configuration Guide

Heimdall uses a JSON configuration file located at `~/.config/heimdall/config.json`. This file controls all aspects of Heimdall's behavior, from theme management to external tool paths.

## Table of Contents

- [Configuration File Location](#configuration-file-location)
- [Configuration Structure](#configuration-structure)
- [Configuration Sections](#configuration-sections)
  - [Theme Configuration](#theme-configuration)
  - [Shell Configuration](#shell-configuration)
  - [Scheme Configuration](#scheme-configuration)
  - [Wallpaper Configuration](#wallpaper-configuration)
  - [Screenshot Configuration](#screenshot-configuration)
  - [Recording Configuration](#recording-configuration)
  - [Clipboard Configuration](#clipboard-configuration)
  - [Emoji Configuration](#emoji-configuration)
  - [PIP Configuration](#pip-configuration)
  - [Notification Configuration](#notification-configuration)
  - [Paths Configuration](#paths-configuration)
  - [Network Configuration](#network-configuration)
  - [External Tools](#external-tools)
- [Example Configuration](#example-configuration)
- [Migration from YAML](#migration-from-yaml)

## Configuration File Location

The configuration file is stored at:
```
~/.config/heimdall/config.json
```

If no configuration exists, Heimdall will create a default configuration on first run.

## Configuration Structure

The configuration is a JSON object with the following top-level sections:

```json
{
  "version": "1.0.0",
  "theme": { ... },
  "shell": { ... },
  "scheme": { ... },
  "wallpaper": { ... },
  "screenshot": { ... },
  "recording": { ... },
  "clipboard": { ... },
  "emoji": { ... },
  "pip": { ... },
  "notification": { ... },
  "paths": { ... },
  "network": { ... },
  "external_tools": { ... }
}
```

## Configuration Sections

### Theme Configuration

Controls which applications receive theme updates when schemes change.

```json
"theme": {
  "enableTerm": true,      // Apply themes to terminal emulators
  "enableHypr": true,      // Apply themes to Hyprland
  "enableDiscord": true,   // Apply themes to Discord (requires BetterDiscord)
  "enableSpicetify": true, // Apply themes to Spotify (requires Spicetify)
  "enableFuzzel": true,    // Apply themes to Fuzzel launcher
  "enableBtop": true,      // Apply themes to btop system monitor
  "enableGtk": true,       // Apply themes to GTK applications
  "enableQt": true         // Apply themes to Qt applications
}
```

### Shell Configuration

Configuration for the Heimdall shell daemon.

```json
"shell": {
  "command": "qs",                    // Shell command to execute
  "args": ["-c", "heimdall", "-n"],   // Arguments for the shell command
  "daemon_port": 9999,                // Port for daemon communication
  "log_file": "shell.log",            // Log file name (relative to state dir)
  "pid_file": "shell.pid",            // PID file name (relative to state dir)
  "ipc_timeout": 5                    // IPC timeout in seconds
}
```

### Scheme Configuration

Color scheme settings.

```json
"scheme": {
  "default": "rosepine",    // Default color scheme
  "auto_mode": true,        // Automatically switch between light/dark
  "material_you": true      // Enable Material You color generation
}
```

### Wallpaper Configuration

Wallpaper management settings.

```json
"wallpaper": {
  "directory": "~/Pictures/Wallpapers",           // Wallpaper directory
  "filter": true,                                 // Enable smart filtering
  "threshold": 0.8,                                // Filter threshold (0.0-1.0)
  "smart_mode": true,                             // Enable smart mode
  "extensions": [".jpg", ".jpeg", ".png", ".webp"] // Supported file extensions
}
```

### Screenshot Configuration

Screenshot capture settings.

```json
"screenshot": {
  "directory": "~/Pictures/Screenshots",      // Screenshot save directory
  "file_format": "png",                      // Output format: png, jpg, webp
  "file_name_pattern": "screenshot_%Y%m%d_%H%M%S", // Filename pattern
  "copy_to_clipboard": true,                 // Copy to clipboard after capture
  "open_with_swappy": true,                  // Open in Swappy editor
  "show_notification": true,                 // Show notification after capture
  "notification_timeout": 3,                 // Notification timeout in seconds
  "freeze_file_name": "freeze.png"           // Temporary file for freeze mode
}
```

**Filename Pattern Placeholders:**
- `%Y%m%d` - Date (YYYYMMDD)
- `%H%M%S` - Time (HHMMSS)

### Recording Configuration

Screen recording settings.

```json
"recording": {
  "directory": "~/Videos/Recordings",         // Recording save directory
  "file_format": "mp4",                      // Output format: mp4, webm, mkv
  "file_name_pattern": "recording_%Y%m%d_%H%M%S", // Filename pattern
  "temp_file_name": "recording.mp4",         // Temporary recording file
  "show_notification": true,                 // Show notifications
  "audio_source": "auto"                     // Audio source: auto, none, or device name
}
```

### Clipboard Configuration

Clipboard manager settings.

```json
"clipboard": {
  "max_entries": 100,                        // Maximum clipboard history entries
  "fuzzel_prompt": "Clipboard> ",            // Fuzzel prompt text
  "fuzzel_args": ["--dmenu", "--width", "50", "--lines", "20"], // Fuzzel arguments
  "preview_length": 50,                      // Text preview character limit
  "delete_on_select": false                  // Delete entry after selection
}
```

### Emoji Configuration

Emoji picker settings.

```json
"emoji": {
  "data_directory": "~/.local/share/emoji",  // Emoji data directory
  "sources": ["emoji.json"],                 // Emoji data sources
  "fuzzel_prompt": "Emoji> ",                // Fuzzel prompt text
  "fuzzel_args": ["--dmenu", "--prompt"],    // Fuzzel arguments
  "copy_to_clipboard": true,                 // Copy emoji to clipboard
  "type_directly": false,                    // Type emoji directly
  "download_timeout": 30                     // Download timeout in seconds
}
```

### PIP Configuration

Picture-in-Picture daemon settings.

```json
"pip": {
  "enabled": true,                           // Enable PIP daemon
  "pid_file": "pip.pid",                     // PID file name
  "window_size": "25%",                      // PIP window size
  "window_position": "bottom-right",         // Window position
  "video_apps": [                            // Applications to enable PIP for
    "mpv", "vlc", "firefox", "chromium", 
    "chrome", "brave", "youtube", "netflix", 
    "twitch", "spotify"
  ],
  "video_keywords": [                        // Keywords to detect video content
    "youtube", "netflix", "twitch", "vimeo",
    "- playing", "▶", "►", "video", "stream"
  ],
  "pin_windows": true,                       // Pin windows to all workspaces
  "always_on_top": true                      // Keep PIP windows on top
}
```

### Notification Configuration

Desktop notification settings.

```json
"notification": {
  "enabled": true,                           // Enable notifications
  "provider": "auto",                        // Provider: auto, notify-send, dunstify
  "default_timeout": 5,                      // Default timeout in seconds
  "app_name": "heimdall",                    // Application name for notifications
  "default_urgency": "normal"                // Default urgency: low, normal, critical
}
```

### Paths Configuration

Custom directory paths. Leave empty to use defaults.

```json
"paths": {
  "templates": "",                           // Custom templates directory
  "schemes": "",                             // Custom schemes directory
  "state_dir": "",                           // Custom state directory
  "cache_dir": "",                           // Custom cache directory
  "data_dir": ""                             // Custom data directory
}
```

### Network Configuration

Network and IPC settings.

```json
"network": {
  "ipc_timeout": 5,                          // General IPC timeout in seconds
  "hypr_ipc_timeout": 5                      // Hyprland IPC timeout in seconds
}
```

### External Tools

Paths to external tools. Leave as default to use system PATH.

```json
"external_tools": {
  "grim": "grim",                            // Screenshot tool
  "slurp": "slurp",                          // Region selection tool
  "swappy": "swappy",                        // Screenshot editor
  "wl_clipboard": "wl-copy",                 // Wayland clipboard tool
  "wl_screenrec": "wl-screenrec",            // Screen recording tool
  "cliphist": "cliphist",                    // Clipboard history tool
  "fuzzel": "fuzzel",                        // Application launcher
  "dart_sass": "sass",                       // Sass compiler
  "libnotify": "notify-send",                // Notification tool
  "dunstify": "dunstify",                    // Dunst notification tool
  "qs": "qs",                                // Quick shell
  "app2unit": "app2unit",                    // App to systemd unit tool
  "xclip": "xclip",                          // X11 clipboard tool
  "pactl": "pactl",                          // PulseAudio control
  "pidof": "pidof",                          // Process ID finder
  "pkill": "pkill",                          // Process killer
  "gdbus": "gdbus"                           // D-Bus tool
}
```

## Example Configuration

A complete example configuration is provided in [`config-example.json`](../config-example.json).

To use the example configuration:
```bash
cp config-example.json ~/.config/heimdall/config.json
# Edit the file to match your preferences
```

## Migration from YAML

If you have an existing YAML configuration from an older version, Heimdall will automatically migrate it to JSON format on first run. The old configuration will be backed up to `config.yaml.backup`.

## Environment Variables

Heimdall respects the following environment variables:

- `XDG_CONFIG_HOME` - Configuration directory (default: `~/.config`)
- `XDG_DATA_HOME` - Data directory (default: `~/.local/share`)
- `XDG_STATE_HOME` - State directory (default: `~/.local/state`)
- `XDG_CACHE_HOME` - Cache directory (default: `~/.cache`)
- `XDG_PICTURES_DIR` - Pictures directory (default: `~/Pictures`)
- `XDG_VIDEOS_DIR` - Videos directory (default: `~/Videos`)

## Tips

1. **Timestamp Patterns**: Use `%Y%m%d` for date and `%H%M%S` for time in filename patterns
2. **Relative Paths**: Paths starting with `~` are expanded to your home directory
3. **Timeouts**: All timeout values are specified in seconds
4. **External Tools**: Specify full paths for external tools if they're not in your PATH
5. **Notifications**: Set `provider` to `dunstify` for more features like notification IDs and replacement

## Troubleshooting

### Configuration Not Loading

If your configuration changes aren't being applied:

1. Check for JSON syntax errors:
   ```bash
   jq . ~/.config/heimdall/config.json
   ```

2. Verify the file location:
   ```bash
   ls -la ~/.config/heimdall/config.json
   ```

3. Check Heimdall logs for errors:
   ```bash
   heimdall --debug
   ```

### Reverting to Defaults

To reset to default configuration:
```bash
rm ~/.config/heimdall/config.json
heimdall scheme list  # Any command will recreate defaults
```

### Manual Migration from YAML

If automatic migration fails:
```bash
# Backup existing config
cp ~/.config/heimdall/config.yaml ~/.config/heimdall/config.yaml.backup

# Remove old config
rm ~/.config/heimdall/config.yaml

# Run any Heimdall command to generate new JSON config
heimdall scheme list
```