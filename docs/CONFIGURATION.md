# Heimdall Configuration Guide

**🎉 NEW: Heimdall now works with ZERO configuration!** Just run any command and Heimdall will use smart defaults. Create a config file only when you want to customize behavior.

This guide covers Heimdall's powerful configuration system, which now features:
- ✨ Zero-config operation with smart defaults
- 🔍 Interactive configuration discovery and exploration
- 📝 Minimal configs - only specify what you want to change
- 🔄 Automatic migration from old formats
- 🎯 Shell completions for all config commands
- 📚 Self-documenting with generated examples

## Table of Contents

- [Quick Start](#quick-start)
- [Zero Configuration](#zero-configuration)
- [Configuration Discovery](#configuration-discovery)
- [Migration Guide](#migration-guide)
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
- [Related Documentation](#related-documentation)

## Quick Start

### Running Without Configuration
```bash
# Just run heimdall - no config needed!
heimdall scheme set catppuccin-mocha
heimdall wallpaper set ~/Pictures/wallpaper.jpg
```

### Creating a Minimal Config
If you want to customize settings, create a minimal config with only the values you want to change:

```json
{
  "scheme": {
    "default": "catppuccin-mocha"
  },
  "theme": {
    "enableDiscord": false
  }
}
```

### Using Example Configs
```bash
# View available example configurations
ls ~/.config/heimdall/docs/examples/

# Use a minimal config for specific needs
cp ~/.config/heimdall/docs/examples/minimal-theme-only.json ~/.config/heimdall/config.json
```

## Zero Configuration

Heimdall now operates perfectly without any configuration file! The system uses smart defaults that work for most users:

### Default Behavior
- **Themes**: Applied to all supported applications
- **Color Scheme**: Uses "rosepine" as default
- **Wallpapers**: Looks in `~/Pictures/Wallpapers`
- **Screenshots**: Saved to `~/Pictures/Screenshots`
- **Material You**: Enabled for wallpaper-based theming
- **Shell**: Uses quickshell if available

### When You Need a Config
You only need to create a config file when you want to:
- Change the default color scheme
- Disable theming for specific applications
- Use custom directories for wallpapers/screenshots
- Configure advanced features like PIP or idle management
- Set custom paths for external tools

## Configuration Discovery

Heimdall provides powerful interactive tools to explore and understand all configuration options:

### Browse All Options
```bash
# List all configuration options with descriptions and current values
heimdall config list

# Show options in a specific category
heimdall config list --category theme
heimdall config list --category scheme

# Filter by type
heimdall config list --type bool    # Show all boolean options
heimdall config list --type string  # Show all string options

# Show only modified values
heimdall config list --modified

# Copy a config path to clipboard
heimdall config list --copy
```

### Search for Options
```bash
# Search by name or description
heimdall config search wallpaper
heimdall config search "discord"
heimdall config search gtk

# Show all options (unfiltered)
heimdall config search --all
```

### Get Detailed Information
```bash
# Describe any configuration option
heimdall config describe theme.enableGtk
heimdall config describe scheme.materialYou
heimdall config describe wallpaper.directory

# Shows:
# - Full description
# - Type information
# - Default value
# - Current value
# - Example usage
```

### View Configuration State
```bash
# Show current effective configuration (user + defaults merged)
heimdall config effective

# Show differences from defaults
heimdall config effective --diff

# Show in JSON format
heimdall config effective --format json

# Show only default values
heimdall config defaults --show
heimdall config defaults --show --format json
```

### Understanding the Visual Indicators

When using `config list`, values are color-coded:
- 🔵 **Gray (●)** - Using default value
- 🟣 **Magenta (●)** - Modified from default
- 🟠 **Orange (●)** - User-set but matches default
- ✅ **Green (✓)** - Enabled boolean
- ❌ **Red (✗)** - Disabled boolean

## Migration Guide

### For Existing Users

If you have an existing heimdall configuration:

1. **Your config continues to work** - No changes required!
2. **Automatic migration** - Old YAML/YML configs are converted automatically
3. **Simplify your config** - Remove any values that match defaults:
   ```bash
   # See what you've customized
   heimdall config list --modified
   
   # View your effective config
   heimdall config effective --diff
   ```

4. **Clean up your config** - Keep only customizations:
   ```json
   // Before (full config)
   {
     "version": "0.2.0",
     "theme": {
       "enableTerm": true,    // This is default
       "enableGtk": true,     // This is default
       "enableDiscord": false // Keep this (customized)
     }
   }
   
   // After (minimal config)
   {
     "theme": {
       "enableDiscord": false
     }
   }
   ```

### From Old Formats

Heimdall automatically migrates:
- **YAML/YML configs** → JSON format
- **Old field names** → New field names
- **Caelestia configs** → Heimdall format

The original config is backed up before migration.

### Shell Completions

Enable shell completions for better experience:

```bash
# Bash
heimdall completion bash > /etc/bash_completion.d/heimdall

# Zsh
heimdall completion zsh > "${fpath[1]}/_heimdall"

# Fish
heimdall completion fish > ~/.config/fish/completions/heimdall.fish

# PowerShell
heimdall completion powershell > heimdall.ps1
```

With completions enabled:
- Tab-complete config paths: `heimdall config get theme.<TAB>`
- Tab-complete categories: `heimdall config list --category <TAB>`
- Tab-complete commands and options

## Configuration File Location

The configuration file is stored at:
```
~/.config/heimdall/config.json
```

If no configuration exists, Heimdall will use built-in defaults. A configuration file is only created when you explicitly save settings.

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
  "default": "rosepine",                          // Default color scheme
  "auto_mode": true,                              // Automatically switch between light/dark
  "material_you": true,                           // Enable Material You color generation
  "user_paths": ["~/.config/heimdall/schemes"],   // Paths to search for user-defined schemes
  "generated_path": "~/.local/share/heimdall/schemes" // Path where generated schemes are stored
}
```

- `user_paths`: Array of directories where user-defined color schemes are stored. Schemes in these directories take precedence over bundled schemes.
- `generated_path`: Directory where dynamically generated schemes (from wallpapers, Material You, etc.) are saved.

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

## Benefits of the New System

### For New Users
- **Zero Learning Curve**: Start using heimdall immediately without reading documentation
- **Discoverable**: Interactive tools help you explore options as needed
- **Minimal Setup**: Only configure what you want to change
- **Smart Defaults**: Sensible defaults that work for most users
- **IDE Support**: JSON Schema provides autocompletion and validation

### For Power Users
- **Efficient Workflows**: Shell completions speed up configuration
- **Advanced Discovery**: Search and filter to find exactly what you need
- **Visual Feedback**: Color-coded display shows customizations at a glance
- **Validation**: Helpful warnings guide you to best practices
- **Programmatic Access**: Config metadata available for scripting

### For Developers
- **Self-Documenting**: Struct tags keep documentation with code
- **Type Safety**: Strong typing with validation
- **Automatic Generation**: Examples and docs generated from code
- **Test Coverage**: Comprehensive tests ensure reliability
- **Clean Architecture**: Well-organized with clear separation of concerns

## Example Configurations

### Minimal Examples

Heimdall provides focused minimal configurations for common use cases:

```bash
# View all minimal examples
ls ~/.config/heimdall/docs/examples/minimal-*.json

# Theme management only
cat ~/.config/heimdall/docs/examples/minimal-theme-only.json

# Wallpaper management only
cat ~/.config/heimdall/docs/examples/minimal-wallpaper-only.json

# Material You theming
cat ~/.config/heimdall/docs/examples/minimal-material-you.json
```

### Complete Examples

For comprehensive configuration references:

```bash
# Full configuration with all options and defaults
cat ~/.config/heimdall/docs/examples/config-full-example.json

# Configuration with inline documentation (JSONC format)
cat ~/.config/heimdall/docs/examples/config-with-comments.jsonc

# Configuration with accompanying documentation
cat ~/.config/heimdall/docs/examples/config-documented.md
```

### Using Example Configurations

```bash
# Start with a minimal config
cp ~/.config/heimdall/docs/examples/minimal-scheme-only.json ~/.config/heimdall/config.json

# Or start with the documented example
cp ~/.config/heimdall/docs/examples/config-with-comments.jsonc ~/.config/heimdall/config.json
```

## Automatic Migration

Heimdall automatically handles migration from:

### Old Formats
- **YAML/YML files** → JSON format
- **Caelestia configs** → Heimdall format
- **Old field names** → Current field names

### Migration Process
1. Detects old configuration format
2. Creates backup (e.g., `config.yaml.backup`)
3. Converts to new format
4. Validates converted configuration
5. Saves new JSON configuration

### Manual Migration
If automatic migration fails:
```bash
# Backup existing config
cp ~/.config/heimdall/config.yaml ~/.config/heimdall/config.yaml.backup

# Remove old config
rm ~/.config/heimdall/config.yaml

# Start fresh (heimdall will use defaults)
heimdall scheme list
```

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

1. **Check if config file exists** (it's optional now!):
   ```bash
   ls -la ~/.config/heimdall/config.json
   # No file? That's fine - heimdall uses defaults
   ```

2. **Validate JSON syntax** (if you have a config):
   ```bash
   jq . ~/.config/heimdall/config.json
   ```

3. **Check what configuration is being used**:
   ```bash
   # Show effective configuration
   heimdall config effective
   
   # Show only your customizations
   heimdall config effective --diff
   ```

4. **Validate your configuration**:
   ```bash
   # Heimdall automatically validates and shows warnings
   heimdall config validate
   ```

### Understanding Configuration State

```bash
# Check if you have a config file
heimdall config status

# See what values you've customized
heimdall config list --modified

# Compare your config with defaults
heimdall config effective --diff

# Search for specific options
heimdall config search "discord"
```

### Reverting to Defaults

```bash
# Option 1: Remove config file (use all defaults)
rm ~/.config/heimdall/config.json

# Option 2: Reset specific values to defaults
heimdall config set theme.enableDiscord --default

# Option 3: View defaults without changing anything
heimdall config defaults --show
```

### Common Issues

#### "Config file not found" - This is normal!
Heimdall works without a config file. This message is informational, not an error.

#### Changes not taking effect
```bash
# Ensure your config is valid
heimdall config validate

# Check the effective configuration
heimdall config effective | grep -i "your_setting"

# Restart any affected services
heimdall shell restart
```

#### Old YAML config still present
```bash
# Heimdall automatically migrates, but you can force it:
heimdall config migrate

# Or manually convert:
mv ~/.config/heimdall/config.yaml ~/.config/heimdall/config.yaml.old
heimdall config save  # Creates new JSON config
```

## Related Documentation

- **[CONFIG_REFERENCE.md](CONFIG_REFERENCE.md)** - Complete reference for all configuration options
- **[CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md)** - Quick reference guide for common configurations
- **[CONFIG_MINIMAL.md](CONFIG_MINIMAL.md)** - Guide to minimal configuration approach
- **[examples/](examples/)** - Directory containing various configuration examples:
  - `config-full-example.json` - Complete configuration with all options
  - `config-with-comments.jsonc` - Documented configuration with inline comments
  - `minimal-*.json` - Minimal configurations for specific use cases
  - `config-schema.json` - JSON Schema for validation and IDE support

## JSON Schema Support

Heimdall provides a JSON Schema for configuration validation and IDE support. You can use this schema in your editor for autocompletion and validation:

```json
{
  "$schema": "https://github.com/heimdall-cli/heimdall-cli/blob/main/docs/examples/config-schema.json",
  "scheme": {
    "default": "catppuccin-mocha"
  }
}
```

Many editors like VS Code will automatically provide IntelliSense when the `$schema` field is present.