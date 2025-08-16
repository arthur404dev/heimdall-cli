# Heimdall CLI Configuration Reference

This document provides a comprehensive reference for all configuration options available in heimdall-cli.

## Table of Contents

- [Clipboard Configuration](#clipboard-configuration)
- [Emoji Configuration](#emoji-configuration)
- [External_tools Configuration](#external_tools-configuration)
- [Migrated_from Configuration](#migrated_from-configuration)
- [Network Configuration](#network-configuration)
- [Notification Configuration](#notification-configuration)
- [Paths Configuration](#paths-configuration)
- [Pip Configuration](#pip-configuration)
- [Recording Configuration](#recording-configuration)
- [Scheme Configuration](#scheme-configuration)
- [Screenshot Configuration](#screenshot-configuration)
- [Shell Configuration](#shell-configuration)
- [Theme Configuration](#theme-configuration)
- [Toggles Configuration](#toggles-configuration)
- [Version Configuration](#version-configuration)
- [Wallpaper Configuration](#wallpaper-configuration)

## Quick Start

Heimdall CLI uses sensible defaults for all configuration options. You only need to create a configuration file if you want to customize the behavior.

### Minimal Configuration

Create a file at `~/.config/heimdall/config.json` with only the settings you want to change:

```json
{
  "scheme": {
    "default": "catppuccin-mocha"
  }
}
```

All other settings will use their default values.

## Clipboard Configuration

### `clipboard.delete_on_select`

Remove entry from history after selection

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `false` |

**Example:**

```json
{
  "clipboard": {
    "delete_on_select": true
  }
}
```

### `clipboard.fuzzel_args`

Additional arguments for Fuzzel launcher

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"["--dmenu", "--width", "50", "--lines", "20"]"` |

**Example:**

```json
{
  "clipboard": {
    "fuzzel_args": ["--dmenu", "--width", "60"]
  }
}
```

### `clipboard.fuzzel_prompt`

Prompt text for clipboard picker

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"Clipboard> "` |

**Example:**

```json
{
  "clipboard": {
    "fuzzel_prompt": "📋 Select: "
  }
}
```

### `clipboard.max_entries`

Maximum number of clipboard history entries

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `100` |

**Example:**

```json
{
  "clipboard": {
    "max_entries": 200
  }
}
```

### `clipboard.preview_length`

Maximum characters to show in preview

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `50` |

**Example:**

```json
{
  "clipboard": {
    "preview_length": 80
  }
}
```

## Emoji Configuration

### `emoji.copy_to_clipboard`

Copy selected emoji to clipboard

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "emoji": {
    "copy_to_clipboard": false
  }
}
```

### `emoji.data_directory`

Directory for emoji data files

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "emoji": {
    "data_directory": "~/.local/share/heimdall/emoji"
  }
}
```

### `emoji.download_timeout`

Timeout for downloading emoji data in seconds

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `30` |

**Example:**

```json
{
  "emoji": {
    "download_timeout": 60
  }
}
```

### `emoji.fuzzel_args`

Additional arguments for Fuzzel launcher

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"["--dmenu", "--prompt"]"` |

**Example:**

```json
{
  "emoji": {
    "fuzzel_args": ["--dmenu", "--width", "40"]
  }
}
```

### `emoji.fuzzel_prompt`

Prompt text for emoji picker

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"Emoji> "` |

**Example:**

```json
{
  "emoji": {
    "fuzzel_prompt": "😀 Pick: "
  }
}
```

### `emoji.sources`

Emoji data source files to use

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"["emoji.json"]"` |

**Example:**

```json
{
  "emoji": {
    "sources": ["emoji.json", "custom.json"]
  }
}
```

### `emoji.type_directly`

Type emoji directly into active window

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `false` |

**Example:**

```json
{
  "emoji": {
    "type_directly": true
  }
}
```

## External_tools Configuration

### `external_tools.app2unit`

Path to app2unit systemd integration tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"app2unit"` |

**Example:**

```json
{
  "external_tools": {
    "app2unit": "/usr/bin/app2unit"
  }
}
```

### `external_tools.cliphist`

Path to cliphist clipboard manager

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"cliphist"` |

**Example:**

```json
{
  "external_tools": {
    "cliphist": "/usr/bin/cliphist"
  }
}
```

### `external_tools.dart_sass`

Path to Dart Sass compiler

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"sass"` |

**Example:**

```json
{
  "external_tools": {
    "dart_sass": "/usr/bin/sass"
  }
}
```

### `external_tools.dunstify`

Path to dunstify notification tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"dunstify"` |

**Example:**

```json
{
  "external_tools": {
    "dunstify": "/usr/bin/dunstify"
  }
}
```

### `external_tools.fuzzel`

Path to fuzzel launcher

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"fuzzel"` |

**Example:**

```json
{
  "external_tools": {
    "fuzzel": "/usr/bin/fuzzel"
  }
}
```

### `external_tools.gdbus`

Path to gdbus D-Bus tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"gdbus"` |

**Example:**

```json
{
  "external_tools": {
    "gdbus": "/usr/bin/gdbus"
  }
}
```

### `external_tools.grim`

Path to grim screenshot tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"grim"` |

**Example:**

```json
{
  "external_tools": {
    "grim": "/usr/bin/grim"
  }
}
```

### `external_tools.libnotify`

Path to notify-send notification tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"notify-send"` |

**Example:**

```json
{
  "external_tools": {
    "libnotify": "/usr/bin/notify-send"
  }
}
```

### `external_tools.pactl`

Path to PulseAudio control utility

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"pactl"` |

**Example:**

```json
{
  "external_tools": {
    "pactl": "/usr/bin/pactl"
  }
}
```

### `external_tools.pidof`

Path to pidof process finder

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"pidof"` |

**Example:**

```json
{
  "external_tools": {
    "pidof": "/usr/bin/pidof"
  }
}
```

### `external_tools.pkill`

Path to pkill process killer

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"pkill"` |

**Example:**

```json
{
  "external_tools": {
    "pkill": "/usr/bin/pkill"
  }
}
```

### `external_tools.qs`

Path to Quickshell executable

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"qs"` |

**Example:**

```json
{
  "external_tools": {
    "qs": "/usr/bin/qs"
  }
}
```

### `external_tools.slurp`

Path to slurp selection tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"slurp"` |

**Example:**

```json
{
  "external_tools": {
    "slurp": "/usr/bin/slurp"
  }
}
```

### `external_tools.swappy`

Path to swappy screenshot editor

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"swappy"` |

**Example:**

```json
{
  "external_tools": {
    "swappy": "/usr/bin/swappy"
  }
}
```

### `external_tools.wl_clipboard`

Path to wl-copy clipboard tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"wl-copy"` |

**Example:**

```json
{
  "external_tools": {
    "wl_clipboard": "/usr/bin/wl-copy"
  }
}
```

### `external_tools.wl_screenrec`

Path to wl-screenrec recording tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"wl-screenrec"` |

**Example:**

```json
{
  "external_tools": {
    "wl_screenrec": "/usr/bin/wl-screenrec"
  }
}
```

### `external_tools.xclip`

Path to xclip X11 clipboard tool

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"xclip"` |

**Example:**

```json
{
  "external_tools": {
    "xclip": "/usr/bin/xclip"
  }
}
```

## Migrated_from Configuration

## Network Configuration

### `network.hypr_ipc_timeout`

Hyprland IPC timeout in seconds

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `5` |

**Example:**

```json
{
  "network": {
    "hypr_ipc_timeout": 3
  }
}
```

### `network.ipc_timeout`

General IPC timeout in seconds

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `5` |

**Example:**

```json
{
  "network": {
    "ipc_timeout": 10
  }
}
```

## Notification Configuration

### `notification.app_name`

Application name shown in notifications

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"heimdall"` |

**Example:**

```json
{
  "notification": {
    "app_name": "Heimdall CLI"
  }
}
```

### `notification.default_timeout`

Default notification timeout in seconds

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `5` |

**Example:**

```json
{
  "notification": {
    "default_timeout": 10
  }
}
```

### `notification.default_urgency`

Default notification urgency (low, normal, critical)

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"normal"` |

**Example:**

```json
{
  "notification": {
    "default_urgency": "low"
  }
}
```

### `notification.enabled`

Enable system notifications

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "notification": {
    "enabled": false
  }
}
```

### `notification.provider`

Notification provider (notify-send, dunstify, auto)

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"auto"` |

**Example:**

```json
{
  "notification": {
    "provider": "dunstify"
  }
}
```

## Paths Configuration

Custom paths for configuration files

### `paths.cache_dir`

Directory for cache files

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "paths": {
    "cache_dir": "~/.cache/heimdall"
  }
}
```

### `paths.data_dir`

Directory for application data

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "paths": {
    "data_dir": "~/.local/share/heimdall"
  }
}
```

### `paths.schemes`

Custom directory for color schemes

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "paths": {
    "schemes": "~/.config/heimdall/schemes"
  }
}
```

### `paths.state_dir`

Directory for state files and runtime data

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "paths": {
    "state_dir": "~/.local/state/heimdall"
  }
}
```

### `paths.templates`

Custom directory for theme templates

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "paths": {
    "templates": "~/.config/heimdall/templates"
  }
}
```

## Pip Configuration

### `pip.always_on_top`

Keep PIP windows above other windows

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "pip": {
    "always_on_top": true
  }
}
```

### `pip.enabled`

Enable picture-in-picture mode

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "pip": {
    "enabled": false
  }
}
```

### `pip.pid_file`

PID file name for PIP process tracking

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"pip.pid"` |

**Example:**

```json
{
  "pip": {
    "pid_file": "pip.pid"
  }
}
```

### `pip.pin_windows`

Pin PIP windows to all workspaces

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "pip": {
    "pin_windows": false
  }
}
```

### `pip.video_apps`

Applications to detect for PIP mode

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"["mpv", "vlc", "firefox", "chromium", "chrome", "brave", "youtube", "netflix", "twitch", "spotify"]"` |

**Example:**

```json
{
  "pip": {
    "video_apps": ["firefox", "mpv"]
  }
}
```

### `pip.video_keywords`

Window title keywords to detect video playback

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"["youtube", "netflix", "twitch", "vimeo", "- playing", "▶", "►", "video", "stream"]"` |

**Example:**

```json
{
  "pip": {
    "video_keywords": ["youtube", "video"]
  }
}
```

### `pip.window_position`

PIP window position on screen

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"bottom-right"` |

**Example:**

```json
{
  "pip": {
    "window_position": "top-left"
  }
}
```

### `pip.window_size`

PIP window size as percentage of screen

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"25%"` |

**Example:**

```json
{
  "pip": {
    "window_size": "30%"
  }
}
```

## Recording Configuration

### `recording.audio_source`

Audio source (auto, none, or specific device)

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"auto"` |

**Example:**

```json
{
  "recording": {
    "audio_source": "none"
  }
}
```

### `recording.directory`

Directory to save screen recordings

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "recording": {
    "directory": "~/Videos/Recordings"
  }
}
```

### `recording.file_format`

Video format (mp4, webm, mkv)

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"mp4"` |

**Example:**

```json
{
  "recording": {
    "file_format": "webm"
  }
}
```

### `recording.file_name_pattern`

Filename pattern with date format codes

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"recording_%Y%m%d_%H%M%S"` |

**Example:**

```json
{
  "recording": {
    "file_name_pattern": "rec_%Y-%m-%d_%H-%M-%S"
  }
}
```

### `recording.show_notification`

Show notification when recording starts/stops

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "recording": {
    "show_notification": false
  }
}
```

### `recording.temp_file_name`

Temporary filename during recording

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"recording.mp4"` |

**Example:**

```json
{
  "recording": {
    "temp_file_name": "temp_recording.mp4"
  }
}
```

## Scheme Configuration

Color scheme management and generation

### `scheme.auto_mode`

Automatically switch between light/dark variants based on time

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "scheme": {
    "auto_mode": true
  }
}
```

### `scheme.default`

Default color scheme to use

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"rosepine"` |

**Example:**

```json
{
  "scheme": {
    "default": "catppuccin-mocha"
  }
}
```

### `scheme.generated_path`

Directory for storing generated Material You schemes

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "scheme": {
    "generated_path": "~/.local/share/heimdall/schemes"
  }
}
```

### `scheme.material_you`

Generate Material You color schemes from wallpapers

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "scheme": {
    "material_you": false
  }
}
```

### `scheme.user_paths`

Additional directories to search for user-defined schemes

| Property | Value |
|----------|-------|
| **Type** | `[]string` |

**Example:**

```json
{
  "scheme": {
    "user_paths": ["~/.config/heimdall/schemes", "~/custom-schemes"]
  }
}
```

## Screenshot Configuration

### `screenshot.copy_to_clipboard`

Copy screenshot to clipboard after capture

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "screenshot": {
    "copy_to_clipboard": false
  }
}
```

### `screenshot.directory`

Directory to save screenshots

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "screenshot": {
    "directory": "~/Pictures/Screenshots"
  }
}
```

### `screenshot.file_format`

Image format (png, jpg, webp)

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"png"` |

**Example:**

```json
{
  "screenshot": {
    "file_format": "webp"
  }
}
```

### `screenshot.file_name_pattern`

Filename pattern with date format codes

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"screenshot_%Y%m%d_%H%M%S"` |

**Example:**

```json
{
  "screenshot": {
    "file_name_pattern": "screen_%Y-%m-%d_%H-%M-%S"
  }
}
```

### `screenshot.freeze_file_name`

Temporary filename for freeze screenshots

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"freeze.png"` |

**Example:**

```json
{
  "screenshot": {
    "freeze_file_name": "temp_freeze.png"
  }
}
```

### `screenshot.notification_timeout`

Notification display duration in seconds

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `3` |

**Example:**

```json
{
  "screenshot": {
    "notification_timeout": 5
  }
}
```

### `screenshot.open_with_swappy`

Open screenshot in Swappy editor after capture

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "screenshot": {
    "open_with_swappy": false
  }
}
```

### `screenshot.show_notification`

Show notification after screenshot capture

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "screenshot": {
    "show_notification": true
  }
}
```

## Shell Configuration

### `shell.args`

Arguments to pass to Quickshell

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"["-c", "heimdall", "-n"]"` |

**Example:**

```json
{
  "shell": {
    "args": ["-c", "heimdall", "-n"]
  }
}
```

### `shell.command`

Quickshell executable command

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"qs"` |

**Example:**

```json
{
  "shell": {
    "command": "qs"
  }
}
```

### `shell.daemon_port`

Port for Quickshell daemon IPC

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `9999` |

**Example:**

```json
{
  "shell": {
    "daemon_port": 9999
  }
}
```

### `shell.ipc_timeout`

IPC timeout in seconds

| Property | Value |
|----------|-------|
| **Type** | `int` |
| **Default** | `5` |

**Example:**

```json
{
  "shell": {
    "ipc_timeout": 10
  }
}
```

### `shell.log_file`

Log file name for shell output

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"shell.log"` |

**Example:**

```json
{
  "shell": {
    "log_file": "shell.log"
  }
}
```

### `shell.log_rules`

Logging rules for Quickshell (Qt logging format)

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "shell": {
    "log_rules": "*.debug=false"
  }
}
```

### `shell.pid_file`

PID file name for shell process

| Property | Value |
|----------|-------|
| **Type** | `string` |
| **Default** | `"shell.pid"` |

**Example:**

```json
{
  "shell": {
    "pid_file": "shell.pid"
  }
}
```

## Theme Configuration

Theme application settings for various applications

### `theme.enableAlacritty`

Apply themes to Alacritty terminal emulator

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `false` |

**Example:**

```json
{
  "theme": {
    "enableAlacritty": true
  }
}
```

### `theme.enableBtop`

Apply themes to btop++ system monitor

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableBtop": true
  }
}
```

### `theme.enableDiscord`

Apply themes to Discord clients (Vesktop, Discord, Vencord, etc.)

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableDiscord": true
  }
}
```

### `theme.enableFuzzel`

Apply themes to Fuzzel launcher

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableFuzzel": true
  }
}
```

### `theme.enableGtk`

Apply themes to GTK 3 and GTK 4 applications

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableGtk": true
  }
}
```

### `theme.enableHypr`

Apply themes to Hyprland window manager configuration

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableHypr": true
  }
}
```

### `theme.enableKitty`

Apply themes to Kitty terminal emulator

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableKitty": true
  }
}
```

### `theme.enableNvim`

Apply themes to Neovim editor (LazyVim integration)

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableNvim": true
  }
}
```

### `theme.enableQt`

Apply themes to Qt5 and Qt6 applications via qt5ct/qt6ct

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableQt": false
  }
}
```

### `theme.enableSpicetify`

Apply themes to Spotify via Spicetify

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableSpicetify": false
  }
}
```

### `theme.enableTerm`

Apply themes to terminal emulators via escape sequences

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "theme": {
    "enableTerm": true
  }
}
```

### `theme.enableWezterm`

Apply themes to WezTerm terminal emulator

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `false` |

**Example:**

```json
{
  "theme": {
    "enableWezterm": true
  }
}
```

#### `theme.paths.alacritty`

Path to Alacritty theme TOML file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "alacritty": "~/.config/alacritty/themes/heimdall.toml"
    }
  }
}
```

#### `theme.paths.betterDiscord`

Path to BetterDiscord theme CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "betterDiscord": "~/.config/BetterDiscord/themes/heimdall.theme.css"
    }
  }
}
```

#### `theme.paths.btop`

Path to btop++ theme file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "btop": "~/.config/btop/themes/heimdall.theme"
    }
  }
}
```

#### `theme.paths.discord`

Path to Discord theme CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "discord": "~/.config/discord/themes/heimdall.css"
    }
  }
}
```

#### `theme.paths.discordCanary`

Path to Discord Canary theme CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "discordCanary": "~/.config/discordcanary/themes/heimdall.css"
    }
  }
}
```

#### `theme.paths.equicord`

Path to Equicord theme CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "equicord": "~/.config/Equicord/themes/heimdall.css"
    }
  }
}
```

#### `theme.paths.fuzzel`

Path to Fuzzel launcher colors configuration

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "fuzzel": "~/.config/fuzzel/colors.ini"
    }
  }
}
```

#### `theme.paths.gtk3`

Path to GTK 3 theme colors CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "gtk3": "~/.config/gtk-3.0/colors.css"
    }
  }
}
```

#### `theme.paths.gtk4`

Path to GTK 4 theme colors CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "gtk4": "~/.config/gtk-4.0/colors.css"
    }
  }
}
```

#### `theme.paths.kitty`

Path to Kitty terminal theme configuration

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "kitty": "~/.config/kitty/themes/heimdall.conf"
    }
  }
}
```

#### `theme.paths.nvim`

Path to Neovim LazyVim theme plugin file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "nvim": "~/.config/nvim/lua/user/heimdall.lua"
    }
  }
}
```

#### `theme.paths.qt5`

Path to Qt5ct color scheme file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "qt5": "~/.config/qt5ct/colors/heimdall.conf"
    }
  }
}
```

#### `theme.paths.qt6`

Path to Qt6ct color scheme file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "qt6": "~/.config/qt6ct/colors/heimdall.conf"
    }
  }
}
```

#### `theme.paths.spicetify`

Path to Spicetify theme color.ini file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "spicetify": "~/.config/spicetify/Themes/heimdall/color.ini"
    }
  }
}
```

#### `theme.paths.terminal`

Path to terminal escape sequences file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "terminal": "~/.config/heimdall/sequences.txt"
    }
  }
}
```

#### `theme.paths.vencord`

Path to Vencord theme CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "vencord": "~/.config/Vencord/themes/heimdall.css"
    }
  }
}
```

#### `theme.paths.vesktop`

Path to Vesktop theme CSS file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "vesktop": "~/.config/vesktop/themes/heimdall.css"
    }
  }
}
```

#### `theme.paths.wezterm`

Path to WezTerm color scheme Lua file

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "theme": {
    "paths": {
      "wezterm": "~/.config/wezterm/colors/heimdall.lua"
    }
  }
}
```

## Toggles Configuration

## Version Configuration

Configuration version management

## Wallpaper Configuration

Wallpaper management and Material You integration

### `wallpaper.directory`

Directory containing wallpaper images

| Property | Value |
|----------|-------|
| **Type** | `string` |

**Example:**

```json
{
  "wallpaper": {
    "directory": "~/Pictures/Wallpapers"
  }
}
```

### `wallpaper.extensions`

Supported image file extensions

| Property | Value |
|----------|-------|
| **Type** | `[]string` |
| **Default** | `"[".jpg", ".jpeg", ".png", ".webp"]"` |

**Example:**

```json
{
  "wallpaper": {
    "extensions": [".jpg", ".png"]
  }
}
```

### `wallpaper.filter`

Filter wallpapers based on color similarity to current scheme

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "wallpaper": {
    "filter": false
  }
}
```

### `wallpaper.smart_mode`

Use intelligent wallpaper selection based on scheme colors

| Property | Value |
|----------|-------|
| **Type** | `bool` |
| **Default** | `true` |

**Example:**

```json
{
  "wallpaper": {
    "smart_mode": true
  }
}
```

### `wallpaper.threshold`

Color similarity threshold for filtering (0.0-1.0, higher = stricter)

| Property | Value |
|----------|-------|
| **Type** | `float` |
| **Default** | `0.8` |

**Example:**

```json
{
  "wallpaper": {
    "threshold": 0.7
  }
}
```

## Default Values

To see all default values, run:

```bash
heimdall config defaults --show
```

## Validation

Heimdall CLI validates your configuration on load. To check if your configuration is valid:

```bash
heimdall config validate
```

## Environment Variables

You can override configuration values using environment variables:

- `HEIMDALL_CONFIG_PATH`: Override the configuration file location
- `HEIMDALL_SCHEME`: Override the default color scheme
- `HEIMDALL_DEBUG`: Enable debug logging

## Examples

See the [examples directory](examples/) for various configuration examples:

- [Minimal theme configuration](examples/minimal-theme-only.json)
- [Material You wallpaper theming](examples/minimal-material-you.json)
- [Quickshell integration](examples/minimal-quickshell.json)
- [Full configuration with all options](examples/config-full-example.json)

