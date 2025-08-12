# Heimdall CLI

<div align="center">

[![Go Version](https://img.shields.io/badge/go-%3E%3D%201.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/heimdall-cli/heimdall/build.yml)](https://github.com/heimdall-cli/heimdall/actions)

**A powerful CLI tool for managing dotfiles, color schemes, wallpapers, and system theming**

</div>

## Overview

Heimdall is a comprehensive command-line interface tool designed for managing your Linux desktop environment with a focus on Hyprland window manager integration. It provides seamless control over theming, wallpapers, color schemes, and application management through a single, efficient Go binary.

### Key Features

- 🎨 **Dynamic Theming** - Material You color generation from wallpapers
- 🖼️ **Smart Wallpaper Management** - Intelligent filtering based on color analysis
- 🎭 **Color Scheme Control** - Multiple scheme flavors and modes support
- 🪟 **Workspace Toggles** - Automated workspace management for applications
- 📸 **Screenshot & Recording** - Built-in screen capture utilities
- 📋 **Clipboard Management** - Integration with clipboard history
- 🐚 **Shell Integration** - Interactive shell with IPC daemon
- 😊 **Emoji Picker** - Quick emoji selection tool
- 🖼️ **PiP Mode** - Picture-in-Picture support for applications
- ⚙️ **Fully Configurable** - Extensive JSON configuration for all features

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/heimdall-cli/heimdall.git
cd heimdall

# Build and install
make build
sudo make install
```

### Binary Release

Download the latest release for your platform:

```bash
# Linux AMD64
wget https://github.com/heimdall-cli/heimdall/releases/latest/download/heimdall-linux-amd64
chmod +x heimdall-linux-amd64
sudo mv heimdall-linux-amd64 /usr/local/bin/heimdall

# Linux ARM64
wget https://github.com/heimdall-cli/heimdall/releases/latest/download/heimdall-linux-arm64
chmod +x heimdall-linux-arm64
sudo mv heimdall-linux-arm64 /usr/local/bin/heimdall
```

## Quick Start

### Initialize Configuration

```bash
# Create default configuration
heimdall init

# Or migrate from Caelestia dotfiles
heimdall migrate --from-caelestia
```

### Configuration

Heimdall uses a JSON configuration file located at `~/.config/heimdall/config.json`. All aspects of Heimdall's behavior can be customized through this file.

📚 **[View the complete Configuration Guide](docs/CONFIGURATION.md)** for detailed information about all available options.

Quick configuration example:
```bash
# Copy example configuration
cp config-example.json ~/.config/heimdall/config.json

# Edit with your preferred editor
nano ~/.config/heimdall/config.json
```

### Basic Usage

```bash
# Set a wallpaper with Material You theming
heimdall wallpaper ~/Pictures/wallpaper.jpg --generate-scheme

# Set a random wallpaper with color filtering
heimdall wallpaper --random --filter --threshold 0.3

# Apply a color scheme
heimdall scheme set catppuccin-mocha

# Take a screenshot
heimdall screenshot --region slurp

# Toggle workspace applications
heimdall toggle communication
```

## Commands

### `wallpaper` - Wallpaper Management

Manage wallpapers with intelligent filtering and Material You integration.

```bash
# Set specific wallpaper
heimdall wallpaper /path/to/image.jpg

# Random wallpaper with filtering
heimdall wallpaper --random --filter --threshold 0.3

# Generate Material You scheme from wallpaper
heimdall wallpaper /path/to/image.jpg --generate-scheme

# Get wallpaper information
heimdall wallpaper --info /path/to/image.jpg
```

**Options:**
- `--random, -r` - Select random wallpaper from configured directory
- `--filter, -f` - Enable colourfulness filtering
- `--threshold, -t` - Set colourfulness threshold (0.0-1.0)
- `--generate-scheme, -g` - Generate Material You color scheme
- `--info, -i` - Display wallpaper analysis information

### `scheme` - Color Scheme Management

Control system-wide color schemes with multiple flavors and modes.

```bash
# List available schemes
heimdall scheme list

# Get current scheme
heimdall scheme get

# Set a scheme
heimdall scheme set catppuccin-mocha

# Set scheme with specific mode
heimdall scheme set --mode dark catppuccin
```

**Subcommands:**
- `list` - List available schemes, flavours, or modes
- `get` - Get current scheme or specific property
- `set` - Set the active scheme

### `screenshot` - Screen Capture

Take screenshots with various capture modes.

```bash
# Full screen screenshot
heimdall screenshot

# Region selection with freeze
heimdall screenshot --region slurp --freeze

# Custom region
heimdall screenshot --region "100,100 500x300"
```

**Options:**
- `--region, -r` - Capture specific region
- `--freeze, -f` - Freeze screen during selection

### `record` - Screen Recording

Record screen content with audio support.

```bash
# Start recording
heimdall record start

# Record specific region
heimdall record start --region slurp

# Stop recording
heimdall record stop

# Record with audio
heimdall record start --audio
```

**Options:**
- `--region, -r` - Record specific region
- `--audio, -a` - Include audio in recording
- `--freeze, -f` - Freeze screen during region selection

### `toggle` - Workspace Application Management

Manage application workspaces with automated launching and positioning.

```bash
# Toggle communication apps (Discord, WhatsApp, etc.)
heimdall toggle communication

# Toggle music players
heimdall toggle music

# Toggle system monitors
heimdall toggle sysmon

# Toggle todo applications
heimdall toggle todo
```

### `shell` - Interactive Shell

Launch an interactive shell with IPC daemon support.

```bash
# Start interactive shell
heimdall shell

# With custom log rules
heimdall shell --log-rules "debug"
```

### `clipboard` - Clipboard Management

Manage clipboard history and content.

```bash
# Show clipboard history picker
heimdall clipboard

# Clear clipboard history
heimdall clipboard --clear

# Copy specific item
heimdall clipboard --copy "text to copy"
```

### `emoji` - Emoji Picker

Quick emoji selection and insertion.

```bash
# Open emoji picker
heimdall emoji

# Search for specific emoji
heimdall emoji --search "smile"
```

### `pip` - Picture-in-Picture Mode

Enable PiP mode for supported applications.

```bash
# Enable PiP for current window
heimdall pip

# Enable PiP for specific application
heimdall pip --app firefox
```

## Configuration

Heimdall uses a comprehensive JSON configuration system that allows you to customize every aspect of its behavior.

📚 **[Read the Complete Configuration Guide](docs/CONFIGURATION.md)**

### Quick Overview

The configuration file is located at `~/.config/heimdall/config.json` and includes settings for:

- **Theme Management** - Control which applications receive theme updates
- **Screenshot & Recording** - Custom directories, file formats, and notifications
- **Clipboard & Emoji** - Fuzzel integration and picker settings
- **PiP Mode** - Picture-in-Picture window management
- **Notifications** - Desktop notification preferences
- **External Tools** - Custom paths for all external dependencies
- **And much more!**

### Example Configuration Snippet

```json
{
  "version": "1.0.0",
  "screenshot": {
    "directory": "~/Pictures/Screenshots",
    "file_format": "png",
    "file_name_pattern": "screenshot_%Y%m%d_%H%M%S",
    "copy_to_clipboard": true,
    "open_with_swappy": true
  },
  "notification": {
    "enabled": true,
    "provider": "auto",
    "default_timeout": 5
  }

}
```

See [`config-example.json`](config-example.json) for a complete example configuration.

## Development

### Building from Source

```bash
# Standard build
make build

# Build for all platforms
make build-all

# Optimized release build
make build-release

# Development with hot reload
make dev
```

### Testing

```bash
# Run tests
make test

# Generate coverage report
make coverage

# Run linter
make lint
```

### Project Structure

```
heimdall-cli/
├── cmd/heimdall/          # Main application entry point
├── internal/
│   ├── commands/          # CLI command implementations
│   │   ├── clipboard/     # Clipboard management
│   │   ├── emoji/         # Emoji picker
│   │   ├── pip/           # Picture-in-Picture
│   │   ├── record/        # Screen recording
│   │   ├── scheme/        # Color scheme management
│   │   ├── screenshot/    # Screenshot capture
│   │   ├── shell/         # Interactive shell
│   │   ├── toggle/        # Workspace toggles
│   │   └── wallpaper/     # Wallpaper management
│   ├── config/            # Configuration management
│   ├── scheme/            # Scheme manager
│   ├── theme/             # Theme engine and applier
│   └── utils/             # Utility packages
│       ├── color/         # Color manipulation
│       ├── hypr/          # Hyprland IPC
│       ├── logger/        # Logging utilities
│       ├── material/      # Material You generation
│       ├── notify/        # Desktop notifications
│       ├── paths/         # Path management
│       └── wallpaper/     # Wallpaper analysis
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── go.sum                 # Dependency checksums
```

## Dependencies

### Required System Tools

- **Hyprland** - Window manager (for full functionality)
- **grim** - Screenshot utility
- **slurp** - Region selection
- **swappy** - Screenshot annotation
- **wl-clipboard** - Wayland clipboard utilities
- **wl-screenrec** - Screen recording
- **cliphist** - Clipboard history
- **fuzzel** - Application launcher
- **swww** - Wallpaper daemon
- **hyprpicker** - Color picker

### Go Dependencies

- `spf13/cobra` - CLI framework
- `spf13/viper` - Configuration management
- `fsnotify/fsnotify` - File system notifications

## Compatibility

- **OS**: Linux (with Wayland compositor)
- **Window Manager**: Optimized for Hyprland
- **Go Version**: 1.24.5 or higher
- **Architecture**: AMD64, ARM64, 386

## Migration from Caelestia

If you're migrating from the original Caelestia dotfiles:

```bash
# Automatic migration
heimdall migrate --from-caelestia

# Manual migration
cp ~/.config/caelestia/config.toml ~/.config/heimdall/config.yaml
heimdall config convert
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original Caelestia dotfiles for inspiration
- Material You design system by Google
- Catppuccin theme for color schemes
- Hyprland community for the excellent window manager

## Support

For issues, questions, or suggestions:
- Open an issue on [GitHub](https://github.com/heimdall-cli/heimdall/issues)
- Check the [Wiki](https://github.com/heimdall-cli/heimdall/wiki) for detailed documentation

---

<div align="center">
Made with ❤️ for the Linux desktop theming community
</div>