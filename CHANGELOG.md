# Changelog

All notable changes to Heimdall CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-08-12

### Added
- **Comprehensive JSON Configuration System**
  - All hardcoded values are now configurable
  - Extensive configuration options for all features
  - Support for custom paths, timeouts, and behaviors
  - Configuration documentation in `docs/CONFIGURATION.md`
  - Quick reference guide in `docs/CONFIG_QUICK_REFERENCE.md`
  - Example configuration file `config-example.json`

- **Material You Color Scheme Support**
  - Full Material You color token implementation (100+ colors)
  - Compatibility with Caelestia's color format
  - Dynamic color generation from wallpapers
  - Embedded color schemes in binary

- **Enhanced Features**
  - Screenshot configuration (directory, format, naming patterns)
  - Recording configuration (audio sources, formats, notifications)
  - Clipboard manager settings (max entries, preview length)
  - Emoji picker configuration (sources, download timeout)
  - PIP mode settings (window size, position, video detection)
  - Notification system configuration (provider, timeouts, urgency)
  - Network and IPC timeout settings
  - External tool path customization

### Changed
- **Configuration Format Migration**
  - Migrated from YAML to JSON configuration format
  - Automatic migration from YAML with backup creation
  - Improved configuration structure and organization
  
- **Color Scheme System**
  - Schemes now use Material You format (.txt files)
  - Removed legacy JSON scheme files
  - Updated scheme list output to match Caelestia format
  - Cleaned up unsupported themes

### Fixed
- Embedded assets now work correctly from any directory
- Scheme loading no longer requires external files
- Configuration file paths are now properly resolved

### Removed
- Legacy YAML configuration support (auto-migrated)
- Old JSON color scheme files (replaced with Material You format)
- Unsupported color schemes (dracula, nord, tokyonight)

## [0.1.0] - 2024-08-09

### Initial Release
- **Core Features**
  - Color scheme management with multiple flavors
  - Wallpaper management with smart filtering
  - Screenshot and screen recording utilities
  - Clipboard history management
  - Emoji picker
  - Picture-in-Picture (PIP) mode
  - Shell integration with IPC daemon
  - Workspace toggle management

- **Theme Integration**
  - Support for terminal emulators
  - Hyprland window manager theming
  - Discord (via BetterDiscord) theming
  - Spotify (via Spicetify) theming
  - Fuzzel launcher theming
  - btop system monitor theming
  - GTK and Qt application theming

- **Color Schemes Included**
  - Catppuccin (4 flavors)
  - Gruvbox (6 flavors)
  - Rose Pine (3 flavors)
  - OneDark
  - OldWorld
  - ShadoTheme

- **External Tool Integration**
  - grim (screenshots)
  - slurp (region selection)
  - swappy (screenshot editor)
  - wl-screenrec (screen recording)
  - cliphist (clipboard history)
  - fuzzel (application launcher)
  - And more...

[0.2.0]: https://github.com/heimdall-cli/heimdall/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/heimdall-cli/heimdall/releases/tag/v0.1.0