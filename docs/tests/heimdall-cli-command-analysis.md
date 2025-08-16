# Heimdall CLI Command Structure Analysis

## Overview

This document provides a comprehensive analysis of all commands in the Heimdall CLI tool, identifying their functionality, testing requirements, dependencies, and critical paths. This analysis will inform our testing strategy and help identify areas requiring comprehensive test coverage.

## Command Structure

### Root Command (root.go, test.go)

#### Purpose
- Main entry point for the Heimdall CLI
- Manages global configuration and initialization
- Provides version information and help system

#### Key Functions
- `Execute()` - Main command execution entry point
- `initConfig()` - Configuration initialization with Viper
- `addCommands()` - Registers all subcommands
- Version command handling

#### Dependencies
- Viper for configuration management
- Cobra for CLI framework
- Logger utility
- All subcommand packages

#### Input/Output Patterns
- Global flags: `--config`, `--verbose`, `--debug`
- Version output with build information
- Configuration file discovery and loading

#### Error Handling Scenarios
- Missing configuration files (graceful fallback)
- Invalid configuration format
- Home directory access failures
- Environment variable parsing errors

#### Configuration Requirements
- Config file paths: `$HOME/.config/heimdall/config.json`
- Backward compatibility with Caelestia paths
- Environment variable prefix: `HEIMDALL_`

#### Test Scenarios
**Happy Path:**
- Execute with valid config file
- Execute with environment variables
- Version command execution
- Help command display

**Edge Cases:**
- Missing home directory
- Corrupted config file
- Invalid environment variables
- No config file present (should use defaults)

**Error Cases:**
- Unreadable config file
- Invalid JSON in config
- Permission denied on config directory

---

### Test Command (test.go)

#### Purpose
- Hidden development command for testing Phase 2 utilities
- Tests color utilities, Hyprland IPC, and notifications

#### Key Functions
- Color utility testing (hex parsing, color manipulation)
- Hyprland IPC connection testing
- Notification system testing

#### Dependencies
- Color utilities
- Hyprland IPC client
- Notification system

#### Test Scenarios
**Happy Path:**
- All utilities working correctly
- Hyprland running and accessible
- Notification system available

**Edge Cases:**
- Hyprland not running
- Notification system unavailable
- Invalid color values

**Error Cases:**
- IPC connection failures
- Color parsing errors
- Notification send failures

---

### Clipboard Command

#### Purpose
- Manage clipboard history using cliphist and fuzzel
- Support for viewing and deleting clipboard items

#### Key Functions
- `run()` - Main execution logic
- Clipboard history retrieval via cliphist
- Interactive selection via fuzzel
- Delete functionality

#### Dependencies
- External tools: cliphist, fuzzel, wl-copy
- Configuration system

#### Input/Output Patterns
- Flag: `--delete` / `-d` for deletion mode
- Interactive fuzzel interface
- Clipboard content manipulation

#### Error Handling Scenarios
- Missing external tools
- User cancellation (ESC/Ctrl+C)
- Empty clipboard history
- Tool execution failures

#### Test Scenarios
**Happy Path:**
- Select and copy clipboard item
- Delete clipboard item
- Cancel selection gracefully

**Edge Cases:**
- Empty clipboard history
- Very large clipboard items
- Special characters in clipboard
- Binary data in clipboard

**Error Cases:**
- cliphist not installed
- fuzzel not available
- wl-copy failure
- Permission denied

---

### Config Command

#### Purpose
- Unified configuration management system
- Support for multiple configuration domains
- Schema validation and JSON operations

#### Key Functions
- `listCommand()` - List configuration domains
- `getCommand()` - Get configuration values
- `setCommand()` - Set configuration values
- `validateCommand()` - Validate configurations
- `saveCommand()` / `loadCommand()` - Persistence operations
- `schemaCommand()` - Display schemas
- `allCommand()` - Bulk operations

#### Dependencies
- Configuration manager
- Configuration providers
- JSON schema system

#### Input/Output Patterns
- Domain-based operations: `heimdall config [domain] [operation]`
- JSON path notation for nested values
- JSON output support

#### Error Handling Scenarios
- Invalid domain names
- Invalid JSON paths
- Schema validation failures
- File permission issues

#### Test Scenarios
**Happy Path:**
- List all domains
- Get/set simple values
- Get/set nested objects
- Validate configurations
- Bulk operations

**Edge Cases:**
- Non-existent domains
- Invalid JSON paths
- Type mismatches
- Empty configurations

**Error Cases:**
- Schema validation failures
- File system errors
- Invalid JSON input
- Permission denied

---

### Emoji Command

#### Purpose
- Emoji picker and search functionality
- Fetch and update emoji databases
- Integration with fuzzel for selection

#### Key Functions
- `updateEmojiData()` - Fetch emoji data from remote sources
- `runEmojiPicker()` - Interactive emoji selection
- `searchEmoji()` - Search by name/tags
- `loadEmojiData()` - Load local emoji database

#### Dependencies
- HTTP client for data fetching
- File system for data storage
- External tools: fuzzel, wl-copy
- Configuration system

#### Input/Output Patterns
- Flags: `--fetch` / `-f`, `--picker` / `-p`
- Search arguments
- Clipboard integration

#### Error Handling Scenarios
- Network failures during fetch
- Corrupted emoji data
- Missing external tools
- User cancellation

#### Test Scenarios
**Happy Path:**
- Fetch emoji data successfully
- Interactive picker selection
- Search by name/tags
- Copy to clipboard

**Edge Cases:**
- No internet connection
- Empty search results
- Large emoji databases
- Special Unicode characters

**Error Cases:**
- Network timeouts
- Invalid emoji data format
- Tool execution failures
- Clipboard copy failures

---

### Idle Command

#### Purpose
- System idle prevention (caffeinate/caffeine equivalent)
- Cross-platform support (X11, Wayland, systemd)
- Session management with timers

#### Key Functions
- `run()` - Main command logic
- `handleStart()` - Start idle prevention
- `handleStop()` - Stop idle prevention
- `handleStatus()` - Show current status
- `runDaemon()` - Daemon mode execution

#### Dependencies
- Idle providers (D-Bus, X11, systemd, fallback)
- Environment detection
- Session management
- Timer system
- Notification system

#### Input/Output Patterns
- Timer duration parsing (30m, 2h, 1h30m)
- Daemon mode support
- Status reporting
- Session listing

#### Error Handling Scenarios
- No available providers
- Provider initialization failures
- Timer parsing errors
- Daemon process management

#### Test Scenarios
**Happy Path:**
- Start/stop idle prevention
- Timer-based sessions
- Daemon mode operation
- Status checking

**Edge Cases:**
- Multiple concurrent sessions
- Very short/long timers
- System sleep during session
- Provider switching

**Error Cases:**
- All providers unavailable
- Invalid timer formats
- Daemon startup failures
- Session cleanup failures

#### Subcomponents Analysis

##### Detector (detector.go)
- **Purpose**: Environment detection (X11/Wayland, desktop environment)
- **Key Functions**: `Detect()`, `detectDisplayServer()`, `detectDesktopEnvironment()`
- **Test Focus**: Environment variable parsing, fallback logic

##### Manager (manager.go, session.go, timer.go)
- **Purpose**: Session lifecycle management
- **Key Functions**: Session creation/removal, state persistence, timer management
- **Test Focus**: Concurrent session handling, state recovery, timer accuracy

##### Providers (provider.go, dbus.go, x11.go, systemd.go, fallback.go)
- **Purpose**: Platform-specific idle prevention implementations
- **Key Functions**: Provider registration, availability checking, inhibition management
- **Test Focus**: Provider selection logic, error handling, cleanup

---

### PIP Command

#### Purpose
- Picture-in-picture daemon for automatic window management
- Monitor active windows and enable PIP for video applications

#### Key Functions
- `startDaemon()` - Start PIP daemon
- `runDaemon()` - Main daemon loop
- `isVideoWindow()` - Video window detection
- `enablePIP()` - Enable PIP mode for windows

#### Dependencies
- Hyprland IPC client
- Process management
- Configuration system
- Notification system

#### Input/Output Patterns
- Daemon mode operation
- PID file management
- Log file output
- Status reporting

#### Error Handling Scenarios
- Hyprland not available
- Window detection failures
- PIP mode activation errors
- Daemon process issues

#### Test Scenarios
**Happy Path:**
- Start/stop daemon
- Video window detection
- PIP mode activation
- Status monitoring

**Edge Cases:**
- Multiple video windows
- Window class changes
- Hyprland restart
- Configuration changes

**Error Cases:**
- Hyprland IPC failures
- Window manipulation errors
- Daemon startup failures
- Process management issues

---

### Record Command

#### Purpose
- Screen recording using wl-screenrec
- Region selection and audio recording support

#### Key Functions
- `startRecording()` - Start screen recording
- `stopRecording()` - Stop and save recording
- Region selection with slurp
- Audio source management

#### Dependencies
- External tools: wl-screenrec, slurp, pactl
- File system operations
- Configuration system
- Notification system

#### Input/Output Patterns
- Flags: `--region` / `-r`, `--sound` / `-s`
- File path management
- Process control

#### Error Handling Scenarios
- Missing external tools
- Audio source detection failures
- File system errors
- Process management issues

#### Test Scenarios
**Happy Path:**
- Start/stop recording
- Region selection
- Audio recording
- File management

**Edge Cases:**
- No audio sources
- Disk space issues
- Long recordings
- System interruptions

**Error Cases:**
- Tool execution failures
- File permission errors
- Audio system failures
- Process termination issues

---

### Scheme Command

#### Purpose
- Color scheme management with multiple subcommands
- Support for bundled and custom schemes
- Material You integration

#### Key Functions
- `listCommand()` - List available schemes
- `getCommand()` - Get current scheme information
- `setCommand()` - Set active scheme
- `installCommand()` - Install bundled schemes
- `bundledCommand()` - Show bundled schemes

#### Dependencies
- Scheme manager
- Theme applier
- Configuration system
- Embedded assets
- Material You generator

#### Input/Output Patterns
- Hierarchical scheme structure (name/flavour/mode)
- JSON output support
- Color preview functionality
- Random scheme selection

#### Error Handling Scenarios
- Invalid scheme combinations
- Missing scheme files
- Theme application failures
- Asset loading errors

#### Test Scenarios
**Happy Path:**
- List/get/set schemes
- Install bundled schemes
- Apply themes
- Random selection

**Edge Cases:**
- Large scheme collections
- Custom scheme formats
- Theme application partial failures
- Color format variations

**Error Cases:**
- Scheme file corruption
- Theme application failures
- Asset loading errors
- Invalid color formats

---

### Screenshot Command

#### Purpose
- Screenshot capture with region selection
- Integration with image editing tools
- Clipboard and notification support

#### Key Functions
- `runScreenshot()` - Main screenshot logic
- Region selection with slurp
- Freeze functionality
- Clipboard integration

#### Dependencies
- External tools: grim, slurp, swappy, wl-copy
- File system operations
- Configuration system
- Notification system

#### Input/Output Patterns
- Flags: `--region` / `-r`, `--freeze` / `-f`
- File naming patterns
- Directory management

#### Error Handling Scenarios
- Missing external tools
- File system errors
- User cancellation
- Clipboard failures

#### Test Scenarios
**Happy Path:**
- Full screen capture
- Region selection
- File saving
- Clipboard copy

**Edge Cases:**
- Very large screenshots
- Special characters in filenames
- Disk space issues
- Multiple monitors

**Error Cases:**
- Tool execution failures
- File permission errors
- Clipboard copy failures
- Directory creation errors

---

### Shell Command

#### Purpose
- Shell daemon management and IPC communication
- Process lifecycle management
- Log streaming and monitoring

#### Key Functions
- `startAttached()` / `startDaemon()` - Start shell processes
- `sendMessage()` - IPC communication
- `StopDaemon()` / `KillDaemon()` - Process termination
- Log management and streaming

#### Dependencies
- Process management
- IPC system (TCP sockets)
- Configuration system
- File system operations

#### Input/Output Patterns
- Daemon mode vs attached mode
- IPC message passing
- Log file management
- PID file handling

#### Error Handling Scenarios
- Process startup failures
- IPC connection issues
- Log file access problems
- Signal handling

#### Test Scenarios
**Happy Path:**
- Start/stop daemon
- Send IPC messages
- Log streaming
- Process monitoring

**Edge Cases:**
- Process crashes
- IPC connection drops
- Log rotation
- Signal handling

**Error Cases:**
- Process startup failures
- IPC communication errors
- File system errors
- Permission issues

#### Subcomponent Analysis

##### IPC (ipc.go)
- **Purpose**: TCP-based inter-process communication
- **Key Functions**: Client/server implementation, message handling
- **Test Focus**: Connection management, message serialization, error handling

---

### Toggle Command

#### Purpose
- Hyprland special workspace management
- Application-specific workspace toggling
- Smart workspace detection

#### Key Functions
- `run()` - Main toggle logic
- `handleSpecialWorkspace()` - Smart workspace handling
- `handleClientConfig()` - Application-specific logic
- Window matching and movement

#### Dependencies
- Hyprland IPC client
- Configuration system
- Process management

#### Input/Output Patterns
- Workspace names (communication, music, sysmon, todo)
- Window matching rules
- Application spawning

#### Error Handling Scenarios
- Hyprland IPC failures
- Window matching errors
- Application spawn failures
- Configuration issues

#### Test Scenarios
**Happy Path:**
- Toggle workspaces
- Application detection
- Window movement
- Smart workspace handling

**Edge Cases:**
- Multiple matching windows
- Application startup delays
- Workspace state changes
- Configuration updates

**Error Cases:**
- Hyprland IPC failures
- Application spawn errors
- Window manipulation failures
- Configuration parsing errors

---

### Wallpaper Command

#### Purpose
- Wallpaper management with Material You integration
- Random wallpaper selection with filtering
- Color scheme generation from images

#### Key Functions
- `setWallpaper()` - Set specific wallpaper
- `setRandomWallpaperFromDir()` - Random selection
- `printColorScheme()` - Color extraction
- `generateMaterialYouScheme()` - Dynamic scheme generation

#### Dependencies
- Image processing libraries
- Material You generator
- Wallpaper analyzer
- Hyprland IPC client
- External tools: hyprctl, swww

#### Input/Output Patterns
- File path handling
- Directory scanning
- JSON color output
- Size filtering parameters

#### Error Handling Scenarios
- Invalid image files
- Missing external tools
- Color extraction failures
- File system errors

#### Test Scenarios
**Happy Path:**
- Set wallpaper by path
- Random wallpaper selection
- Color scheme extraction
- Material You generation

**Edge Cases:**
- Very large images
- Unusual image formats
- Empty directories
- Color extraction edge cases

**Error Cases:**
- Corrupted image files
- Tool execution failures
- File system errors
- Color generation failures

---

## Critical Testing Areas

### High Priority (Must Test)
1. **Configuration System** - Core functionality for all commands
2. **Error Handling** - Graceful degradation when external tools missing
3. **File System Operations** - Path handling, permissions, atomic operations
4. **External Tool Integration** - Command execution, output parsing, error handling
5. **IPC Communication** - Hyprland integration, shell daemon communication

### Medium Priority (Should Test)
1. **Color Processing** - Scheme generation, color format conversion
2. **Process Management** - Daemon lifecycle, signal handling
3. **Image Processing** - Wallpaper analysis, format support
4. **Network Operations** - Emoji data fetching, timeout handling

### Low Priority (Could Test)
1. **UI Integration** - Fuzzel interactions, notification display
2. **Performance** - Large file handling, memory usage
3. **Platform Compatibility** - Different desktop environments

## Common Testing Patterns

### External Tool Dependency Testing
- Mock external tool availability
- Test graceful fallback when tools missing
- Validate command construction and execution
- Test output parsing and error handling

### Configuration Testing
- Test default value handling
- Validate configuration merging
- Test schema validation
- Test file system error handling

### File System Testing
- Test path expansion and validation
- Test atomic file operations
- Test permission handling
- Test directory creation

### Error Propagation Testing
- Test error message clarity
- Test error recovery mechanisms
- Test partial failure scenarios
- Test cleanup on errors

## Recommended Test Structure

```
tests/
├── unit/
│   ├── commands/
│   │   ├── clipboard_test.go
│   │   ├── config_test.go
│   │   ├── emoji_test.go
│   │   ├── idle_test.go
│   │   ├── pip_test.go
│   │   ├── record_test.go
│   │   ├── scheme_test.go
│   │   ├── screenshot_test.go
│   │   ├── shell_test.go
│   │   ├── toggle_test.go
│   │   └── wallpaper_test.go
│   ├── utils/
│   └── config/
├── integration/
│   ├── external_tools_test.go
│   ├── hyprland_integration_test.go
│   └── file_system_test.go
└── e2e/
    ├── command_execution_test.go
    └── workflow_test.go
```

This analysis provides a comprehensive foundation for developing a robust testing strategy that covers all critical paths, edge cases, and error scenarios in the Heimdall CLI tool.