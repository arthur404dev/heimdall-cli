# Configuration Documentation

This document describes all configuration options for the `config-documented.json` file.

## Theme Settings

Controls which applications receive theme updates.

- `theme.enableTerm`: Apply themes to terminal emulators
- `theme.enableHypr`: Apply themes to Hyprland
- `theme.enableDiscord`: Apply themes to Discord clients
- `theme.enableSpicetify`: Apply themes to Spotify via Spicetify
- `theme.enableFuzzel`: Apply themes to Fuzzel launcher
- `theme.enableBtop`: Apply themes to btop system monitor
- `theme.enableGtk`: Apply themes to GTK applications
- `theme.enableQt`: Apply themes to Qt applications
- `theme.enableKitty`: Apply themes to Kitty terminal
- `theme.enableAlacritty`: Apply themes to Alacritty terminal
- `theme.enableWezterm`: Apply themes to WezTerm terminal

## Shell Integration

Settings for Quickshell integration.

- `shell.command`: Command to execute shell (default: "qs")
- `shell.args`: Arguments for shell command
- `shell.daemon_port`: Port for daemon communication
- `shell.log_file`: Path to log file
- `shell.pid_file`: Path to PID file
- `shell.ipc_timeout`: IPC timeout in seconds

## Scheme Settings

Color scheme management configuration.

- `scheme.default`: Default color scheme to use
- `scheme.auto_mode`: Automatically switch between light/dark variants
- `scheme.material_you`: Generate Material You schemes from wallpapers
- `scheme.user_paths`: Paths to search for user-defined schemes
- `scheme.generated_path`: Path to store generated schemes

## Wallpaper Settings

Wallpaper management configuration.

- `wallpaper.directory`: Directory containing wallpapers
- `wallpaper.filter`: Enable smart filtering based on aspect ratio
- `wallpaper.threshold`: Similarity threshold for filtering (0.0-1.0)
- `wallpaper.smart_mode`: Enable intelligent wallpaper selection
- `wallpaper.extensions`: Supported image file extensions

## Screenshot Settings

Screenshot capture configuration.

- `screenshot.directory`: Directory to save screenshots
- `screenshot.file_format`: Image format (png, jpg, etc.)
- `screenshot.file_name_pattern`: Pattern for filename generation
- `screenshot.copy_to_clipboard`: Copy screenshot to clipboard
- `screenshot.open_after_capture`: Open screenshot after capture
- `screenshot.capture_mouse`: Include mouse cursor in screenshot
- `screenshot.capture_decorations`: Include window decorations
- `screenshot.delay`: Delay before capture in seconds
- `screenshot.quality`: Image quality (1-100)

