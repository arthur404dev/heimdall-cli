# Heimdall CLI Comprehensive Test Plan

## Overview

This document outlines a comprehensive testing strategy for the Heimdall CLI tool, covering all 12+ commands, utilities, and cross-cutting concerns. The plan is designed to ensure robust test coverage, maintainable test code, and reliable CI/CD integration.

**Test Framework**: Go's built-in testing framework with table-driven tests
**Coverage Target**: 85% line coverage, 75% branch coverage
**Test Categories**: Unit, Integration, End-to-End
**Mock Strategy**: Interface-based mocking for external dependencies

## Testing Strategy Overview

### Unit vs Integration Test Boundaries

#### Unit Tests (70% of test effort)
- **Scope**: Individual functions, methods, and components in isolation
- **Dependencies**: All external dependencies mocked (file system, network, external tools)
- **Focus**: Business logic, data transformations, error handling, edge cases
- **Execution**: Fast (<1s total), no external dependencies, parallel execution

#### Integration Tests (25% of test effort)
- **Scope**: Component interactions, external tool integration, file system operations
- **Dependencies**: Real file system (temp directories), mocked network calls
- **Focus**: Command execution flows, configuration loading, IPC communication
- **Execution**: Medium speed (<30s total), isolated test environments

#### End-to-End Tests (5% of test effort)
- **Scope**: Complete command workflows, user scenarios
- **Dependencies**: Controlled test environment with required tools
- **Focus**: Critical user paths, command combinations
- **Execution**: Slow, run in CI only, comprehensive scenarios

### Mock Strategy for External Dependencies

#### External Tools Mocking
```go
// Tool execution interface for mocking
type ToolExecutor interface {
    Execute(cmd string, args ...string) ([]byte, error)
    IsAvailable(tool string) bool
}

// Mock implementation for testing
type MockToolExecutor struct {
    commands map[string]MockCommand
    available map[string]bool
}
```

#### File System Operations
- Use `afero.Fs` interface for file system abstraction
- Memory-based file system for unit tests
- Temporary directories for integration tests

#### Network Operations
- HTTP client interface with mock implementations
- Predefined response fixtures for emoji data, etc.

#### IPC Communication
- Mock Hyprland IPC client
- Mock D-Bus interfaces
- Simulated process management

### Test Organization and Structure

```
tests/
├── unit/                           # Unit tests (fast, isolated)
│   ├── commands/                   # Command-specific tests
│   │   ├── clipboard_test.go
│   │   ├── config_test.go
│   │   ├── emoji_test.go
│   │   ├── idle_test.go
│   │   ├── pip_test.go
│   │   ├── record_test.go
│   │   ├── root_test.go
│   │   ├── scheme_test.go
│   │   ├── screenshot_test.go
│   │   ├── shell_test.go
│   │   ├── toggle_test.go
│   │   └── wallpaper_test.go
│   ├── config/                     # Configuration system tests
│   │   ├── manager_test.go
│   │   ├── providers_test.go
│   │   └── schema_test.go
│   ├── utils/                      # Utility tests
│   │   ├── color_test.go          # Already exists
│   │   ├── logger_test.go
│   │   ├── notify_test.go
│   │   ├── paths_test.go
│   │   └── wallpaper_test.go
│   └── testutils/                  # Test utilities and helpers
│       ├── mocks.go
│       ├── fixtures.go
│       └── helpers.go
├── integration/                    # Integration tests (medium speed)
│   ├── command_execution_test.go   # Command execution flows
│   ├── config_integration_test.go  # Configuration loading/saving
│   ├── external_tools_test.go      # External tool integration
│   ├── file_operations_test.go     # File system operations
│   └── ipc_integration_test.go     # IPC communication tests
├── e2e/                           # End-to-end tests (slow)
│   ├── user_workflows_test.go     # Complete user scenarios
│   └── command_combinations_test.go
└── fixtures/                      # Test data and fixtures
    ├── configs/                   # Sample configuration files
    ├── schemes/                   # Sample color schemes
    ├── images/                    # Sample wallpaper images
    └── responses/                 # Mock HTTP responses
```

### Coverage Targets and Quality Metrics

#### Coverage Targets
- **Overall Line Coverage**: 85%
- **Branch Coverage**: 75%
- **Function Coverage**: 90%
- **Critical Path Coverage**: 100%

#### Quality Metrics
- **Test Execution Time**: Unit tests <1s, Integration <30s
- **Test Reliability**: <1% flaky test rate
- **Test Maintainability**: Clear naming, minimal duplication
- **Documentation Coverage**: All public APIs documented with examples

## Test Infrastructure Requirements

### Test Utilities and Helpers

#### Core Test Utilities (`testutils/helpers.go`)
```go
// Test environment setup
func SetupTestEnvironment(t *testing.T) *TestEnvironment
func CleanupTestEnvironment(env *TestEnvironment)

// Temporary directory management
func CreateTempDir(t *testing.T, prefix string) string
func CreateTempFile(t *testing.T, dir, name, content string) string

// Configuration helpers
func CreateTestConfig(t *testing.T, config map[string]interface{}) string
func LoadTestScheme(t *testing.T, name string) *scheme.Scheme

// Assertion helpers
func AssertFileExists(t *testing.T, path string)
func AssertFileContains(t *testing.T, path, content string)
func AssertCommandSuccess(t *testing.T, cmd *cobra.Command, args []string)
func AssertCommandError(t *testing.T, cmd *cobra.Command, args []string, expectedError string)
```

#### Mock Implementations (`testutils/mocks.go`)
```go
// External tool executor mock
type MockToolExecutor struct {
    commands map[string]MockCommand
    available map[string]bool
}

// HTTP client mock
type MockHTTPClient struct {
    responses map[string]*http.Response
    errors    map[string]error
}

// File system mock (using afero)
type MockFileSystem struct {
    fs afero.Fs
}

// Hyprland IPC mock
type MockHyprlandClient struct {
    workspaces []Workspace
    windows    []Window
    responses  map[string]string
}

// Notification system mock
type MockNotifier struct {
    notifications []Notification
}
```

### Test Fixtures and Data Setup

#### Configuration Fixtures (`fixtures/configs/`)
- `minimal.json` - Minimal valid configuration
- `complete.json` - Complete configuration with all options
- `invalid.json` - Invalid configuration for error testing
- `legacy.json` - Legacy Caelestia configuration format

#### Scheme Fixtures (`fixtures/schemes/`)
- Sample color schemes for each supported format
- Invalid scheme files for error testing
- Material You generated schemes

#### Image Fixtures (`fixtures/images/`)
- Small test images in various formats (PNG, JPG, WebP)
- Invalid/corrupted image files
- Images with different color profiles

### CI/CD Integration Considerations

#### GitHub Actions Configuration
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y xvfb  # For X11 testing
    
    - name: Run unit tests
      run: go test -v -race -coverprofile=coverage.out ./tests/unit/...
    
    - name: Run integration tests
      run: go test -v -timeout=5m ./tests/integration/...
    
    - name: Run e2e tests
      run: go test -v -timeout=10m ./tests/e2e/...
      if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

#### Test Environment Requirements
- **OS**: Ubuntu 22.04 (primary), test on multiple distros
- **Go Version**: 1.21+ (test on latest 2 versions)
- **External Tools**: Mock in unit tests, install for integration tests
- **Display Server**: Xvfb for headless X11 testing

## Command-Specific Test Plans

### Root Command (`internal/commands/root.go`)

#### Test File: `tests/unit/commands/root_test.go`

#### Key Test Scenarios

##### Happy Path Tests
```go
func TestRootCommand_Execute(t *testing.T) {
    tests := []struct {
        name string
        args []string
        want string
    }{
        {"version flag", []string{"--version"}, "heimdall version"},
        {"help flag", []string{"--help"}, "Main control script"},
        {"no args", []string{}, "Main control script"},
    }
    // Implementation...
}
```

##### Configuration Loading Tests
```go
func TestRootCommand_ConfigLoading(t *testing.T) {
    tests := []struct {
        name       string
        configFile string
        envVars    map[string]string
        wantError  bool
    }{
        {"valid config file", "fixtures/configs/minimal.json", nil, false},
        {"missing config file", "nonexistent.json", nil, false}, // Should use defaults
        {"invalid config file", "fixtures/configs/invalid.json", nil, true},
        {"environment variables", "", map[string]string{"HEIMDALL_DEBUG": "true"}, false},
    }
    // Implementation...
}
```

##### Error Handling Tests
```go
func TestRootCommand_ErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        setup       func(*testing.T) string // Returns config path
        wantError   bool
        errorString string
    }{
        {"unreadable config file", setupUnreadableConfig, true, "permission denied"},
        {"corrupted config file", setupCorruptedConfig, true, "invalid JSON"},
        {"missing home directory", setupMissingHome, false, ""}, // Should fallback
    }
    // Implementation...
}
```

#### Mock Requirements
- File system operations (config file reading)
- Environment variable access
- Home directory detection

#### Expected Coverage Areas
- Command initialization and setup
- Configuration file discovery and loading
- Environment variable processing
- Version information display
- Help system functionality
- Error handling and recovery

---

### Clipboard Command (`internal/commands/clipboard/clipboard.go`)

#### Test File: `tests/unit/commands/clipboard_test.go`

#### Key Test Scenarios

##### Happy Path Tests
```go
func TestClipboardCommand_Run(t *testing.T) {
    tests := []struct {
        name           string
        deleteFlag     bool
        cliphistOutput string
        fuzzelSelection string
        wantCopied     string
    }{
        {"select item", false, "item1\nitem2\nitem3", "item2", "item2"},
        {"delete item", true, "item1\nitem2\nitem3", "item2", ""},
        {"empty history", false, "", "", ""},
    }
    // Implementation...
}
```

##### Edge Cases Tests
```go
func TestClipboardCommand_EdgeCases(t *testing.T) {
    tests := []struct {
        name           string
        cliphistOutput string
        fuzzelBehavior string
        wantError      bool
    }{
        {"user cancellation", "item1\nitem2", "ESC", false},
        {"very large items", strings.Repeat("x", 10000), "selection", false},
        {"special characters", "item with\nnewlines\tand\ttabs", "selection", false},
        {"binary data", "\x00\x01\x02\x03", "selection", false},
    }
    // Implementation...
}
```

##### Error Cases Tests
```go
func TestClipboardCommand_ErrorCases(t *testing.T) {
    tests := []struct {
        name         string
        toolsAvailable map[string]bool
        wantError    bool
        errorString  string
    }{
        {"cliphist missing", map[string]bool{"cliphist": false}, true, "cliphist not found"},
        {"fuzzel missing", map[string]bool{"fuzzel": false}, true, "fuzzel not found"},
        {"wl-copy missing", map[string]bool{"wl-copy": false}, true, "wl-copy not found"},
        {"all tools missing", map[string]bool{}, true, "required tools not available"},
    }
    // Implementation...
}
```

#### Mock Requirements
- External tool executor (cliphist, fuzzel, wl-copy)
- Tool availability checker
- Command output simulation

#### Expected Coverage Areas
- Command flag parsing (--delete)
- External tool execution and output parsing
- User interaction simulation
- Clipboard manipulation
- Error handling for missing tools
- User cancellation handling

---

### Config Command (`internal/commands/config/config.go`)

#### Test File: `tests/unit/commands/config_test.go`

#### Key Test Scenarios

##### Configuration Operations Tests
```go
func TestConfigCommand_Operations(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        domain    string
        key       string
        value     interface{}
        wantError bool
    }{
        {"list domains", "list", "", "", nil, false},
        {"get simple value", "get", "shell", "theme", nil, false},
        {"set simple value", "set", "shell", "theme", "dark", false},
        {"get nested value", "get", "idle", "providers.x11.enabled", nil, false},
        {"set nested value", "set", "idle", "providers.x11.enabled", true, false},
        {"invalid domain", "get", "nonexistent", "key", nil, true},
        {"invalid key path", "get", "shell", "invalid.path", nil, true},
    }
    // Implementation...
}
```

##### Schema Validation Tests
```go
func TestConfigCommand_SchemaValidation(t *testing.T) {
    tests := []struct {
        name      string
        domain    string
        config    map[string]interface{}
        wantError bool
        errorType string
    }{
        {"valid config", "shell", validShellConfig, false, ""},
        {"missing required field", "shell", missingRequiredConfig, true, "required"},
        {"invalid type", "shell", invalidTypeConfig, true, "type"},
        {"invalid enum value", "shell", invalidEnumConfig, true, "enum"},
    }
    // Implementation...
}
```

##### Bulk Operations Tests
```go
func TestConfigCommand_BulkOperations(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        configs   map[string]map[string]interface{}
        wantError bool
    }{
        {"save all configs", "save", allValidConfigs, false},
        {"load all configs", "load", nil, false},
        {"validate all configs", "validate", allValidConfigs, false},
        {"mixed valid/invalid", "validate", mixedConfigs, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- Configuration manager interface
- File system operations (config file I/O)
- JSON schema validation
- Configuration providers

#### Expected Coverage Areas
- All configuration operations (list, get, set, validate, save, load)
- Domain-based configuration management
- JSON path notation parsing
- Schema validation and error reporting
- Bulk operations
- File system error handling

---

### Emoji Command (`internal/commands/emoji/emoji.go`)

#### Test File: `tests/unit/commands/emoji_test.go`

#### Key Test Scenarios

##### Data Fetching Tests
```go
func TestEmojiCommand_DataFetching(t *testing.T) {
    tests := []struct {
        name         string
        httpResponse *http.Response
        httpError    error
        wantError    bool
    }{
        {"successful fetch", mockEmojiResponse(), nil, false},
        {"network timeout", nil, &net.OpError{Op: "dial", Err: errors.New("timeout")}, true},
        {"404 not found", &http.Response{StatusCode: 404}, nil, true},
        {"invalid JSON", mockInvalidJSONResponse(), nil, true},
    }
    // Implementation...
}
```

##### Emoji Picker Tests
```go
func TestEmojiCommand_Picker(t *testing.T) {
    tests := []struct {
        name           string
        emojiData      []Emoji
        fuzzelSelection string
        wantCopied     string
        wantError      bool
    }{
        {"select emoji", sampleEmojiData, "😀 grinning face", "😀", false},
        {"search and select", sampleEmojiData, "❤️ red heart", "❤️", false},
        {"user cancellation", sampleEmojiData, "", "", false},
        {"empty data", []Emoji{}, "", "", true},
    }
    // Implementation...
}
```

##### Search Functionality Tests
```go
func TestEmojiCommand_Search(t *testing.T) {
    tests := []struct {
        name      string
        query     string
        emojiData []Emoji
        wantCount int
    }{
        {"exact name match", "heart", heartEmojiData, 5},
        {"partial match", "face", faceEmojiData, 20},
        {"tag search", "love", loveTaggedEmojis, 8},
        {"no matches", "xyz123", allEmojiData, 0},
        {"case insensitive", "HEART", heartEmojiData, 5},
    }
    // Implementation...
}
```

#### Mock Requirements
- HTTP client for emoji data fetching
- File system for local emoji database
- External tool executor (fuzzel, wl-copy)
- Network error simulation

#### Expected Coverage Areas
- Emoji data fetching and caching
- Local database management
- Interactive picker functionality
- Search and filtering
- Clipboard integration
- Network error handling
- Data corruption handling

---

### Idle Command (`internal/commands/idle/idle.go`)

#### Test File: `tests/unit/commands/idle_test.go`

#### Key Test Scenarios

##### Provider Management Tests
```go
func TestIdleCommand_ProviderManagement(t *testing.T) {
    tests := []struct {
        name              string
        availableProviders []string
        environment       map[string]string
        wantProvider      string
        wantError         bool
    }{
        {"X11 environment", []string{"x11", "fallback"}, map[string]string{"DISPLAY": ":0"}, "x11", false},
        {"Wayland environment", []string{"dbus", "fallback"}, map[string]string{"WAYLAND_DISPLAY": "wayland-0"}, "dbus", false},
        {"systemd available", []string{"systemd", "fallback"}, map[string]string{}, "systemd", false},
        {"fallback only", []string{"fallback"}, map[string]string{}, "fallback", false},
        {"no providers", []string{}, map[string]string{}, "", true},
    }
    // Implementation...
}
```

##### Session Management Tests
```go
func TestIdleCommand_SessionManagement(t *testing.T) {
    tests := []struct {
        name        string
        operation   string
        sessionID   string
        duration    string
        wantError   bool
    }{
        {"start session", "start", "", "30m", false},
        {"start with timer", "start", "", "2h", false},
        {"stop session", "stop", "session-123", "", false},
        {"status check", "status", "", "", false},
        {"invalid duration", "start", "", "invalid", true},
        {"stop nonexistent", "stop", "nonexistent", "", true},
    }
    // Implementation...
}
```

##### Timer Parsing Tests
```go
func TestIdleCommand_TimerParsing(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantDuration time.Duration
        wantError bool
    }{
        {"minutes only", "30m", 30 * time.Minute, false},
        {"hours only", "2h", 2 * time.Hour, false},
        {"hours and minutes", "1h30m", 90 * time.Minute, false},
        {"seconds", "45s", 45 * time.Second, false},
        {"complex", "2h15m30s", 2*time.Hour + 15*time.Minute + 30*time.Second, false},
        {"invalid format", "2x", 0, true},
        {"empty string", "", 0, true},
        {"negative", "-30m", 0, true},
    }
    // Implementation...
}
```

#### Subcomponent Tests

##### Detector Tests (`tests/unit/commands/idle/detector_test.go`)
```go
func TestDetector_EnvironmentDetection(t *testing.T) {
    tests := []struct {
        name        string
        environment map[string]string
        wantDisplay string
        wantDesktop string
    }{
        {"X11 with GNOME", map[string]string{"DISPLAY": ":0", "XDG_CURRENT_DESKTOP": "GNOME"}, "x11", "gnome"},
        {"Wayland with KDE", map[string]string{"WAYLAND_DISPLAY": "wayland-0", "XDG_CURRENT_DESKTOP": "KDE"}, "wayland", "kde"},
        {"Hyprland", map[string]string{"HYPRLAND_INSTANCE_SIGNATURE": "abc123"}, "wayland", "hyprland"},
        {"unknown", map[string]string{}, "unknown", "unknown"},
    }
    // Implementation...
}
```

##### Provider Tests (`tests/unit/commands/idle/providers_test.go`)
```go
func TestProviders_Availability(t *testing.T) {
    tests := []struct {
        name         string
        provider     string
        environment  map[string]string
        mockCommands map[string]bool
        wantAvailable bool
    }{
        {"X11 provider available", "x11", map[string]string{"DISPLAY": ":0"}, map[string]bool{"xset": true}, true},
        {"D-Bus provider available", "dbus", map[string]string{}, map[string]bool{"dbus-send": true}, true},
        {"systemd available", "systemd", map[string]string{}, map[string]bool{"systemctl": true}, true},
        {"fallback always available", "fallback", map[string]string{}, map[string]bool{}, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- Environment variable access
- External tool availability checking
- D-Bus interface mocking
- Process management
- Timer and session state

#### Expected Coverage Areas
- Environment detection (X11/Wayland, desktop environment)
- Provider selection and initialization
- Session lifecycle management
- Timer parsing and validation
- Daemon mode operation
- Provider-specific inhibition logic
- Error handling and fallbacks

---

### PIP Command (`internal/commands/pip/pip.go`)

#### Test File: `tests/unit/commands/pip_test.go`

#### Key Test Scenarios

##### Daemon Management Tests
```go
func TestPIPCommand_DaemonManagement(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        pidFile   string
        wantError bool
    }{
        {"start daemon", "start", "", false},
        {"stop daemon", "stop", "valid-pid-file", false},
        {"status check", "status", "valid-pid-file", false},
        {"start when running", "start", "valid-pid-file", true},
        {"stop when not running", "stop", "", true},
    }
    // Implementation...
}
```

##### Video Window Detection Tests
```go
func TestPIPCommand_VideoWindowDetection(t *testing.T) {
    tests := []struct {
        name        string
        windows     []Window
        wantVideo   []Window
    }{
        {"YouTube in browser", browserWindows, []Window{youtubeWindow}},
        {"VLC player", mediaWindows, []Window{vlcWindow}},
        {"multiple videos", multiVideoWindows, []Window{youtubeWindow, vlcWindow}},
        {"no video windows", textWindows, []Window{}},
        {"mixed windows", mixedWindows, []Window{youtubeWindow}},
    }
    // Implementation...
}
```

##### PIP Mode Activation Tests
```go
func TestPIPCommand_PIPActivation(t *testing.T) {
    tests := []struct {
        name      string
        window    Window
        wantError bool
    }{
        {"valid video window", validVideoWindow, false},
        {"already in PIP", pipWindow, false}, // Should be idempotent
        {"invalid window", invalidWindow, true},
        {"window disappeared", nonexistentWindow, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- Hyprland IPC client
- Process management (PID files, daemon control)
- Window detection and manipulation
- Notification system

#### Expected Coverage Areas
- Daemon lifecycle management
- Video window detection algorithms
- PIP mode activation/deactivation
- Window monitoring and event handling
- Process management and PID file handling
- Error handling for Hyprland unavailability

---

### Record Command (`internal/commands/record/record.go`)

#### Test File: `tests/unit/commands/record_test.go`

#### Key Test Scenarios

##### Recording Management Tests
```go
func TestRecordCommand_RecordingManagement(t *testing.T) {
    tests := []struct {
        name        string
        operation   string
        regionFlag  bool
        soundFlag   bool
        wantError   bool
    }{
        {"start full screen", "start", false, false, false},
        {"start with region", "start", true, false, false},
        {"start with sound", "start", false, true, false},
        {"start with both", "start", true, true, false},
        {"stop recording", "stop", false, false, false},
        {"stop when not recording", "stop", false, false, true},
    }
    // Implementation...
}
```

##### Region Selection Tests
```go
func TestRecordCommand_RegionSelection(t *testing.T) {
    tests := []struct {
        name         string
        slurpOutput  string
        wantRegion   string
        wantError    bool
    }{
        {"valid region", "100,100 200x150", "100,100 200x150", false},
        {"user cancellation", "", "", false},
        {"invalid format", "invalid", "", true},
        {"slurp error", "", "", true},
    }
    // Implementation...
}
```

##### Audio Source Detection Tests
```go
func TestRecordCommand_AudioSources(t *testing.T) {
    tests := []struct {
        name         string
        pactlOutput  string
        wantSources  []string
        wantError    bool
    }{
        {"multiple sources", multiSourceOutput, []string{"source1", "source2"}, false},
        {"single source", singleSourceOutput, []string{"default"}, false},
        {"no sources", "", []string{}, false},
        {"pactl error", "", nil, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- External tool executor (wl-screenrec, slurp, pactl)
- File system operations (output file management)
- Process management (recording process control)
- Audio system interface

#### Expected Coverage Areas
- Recording start/stop functionality
- Region selection with slurp integration
- Audio source detection and configuration
- File naming and path management
- Process control and signal handling
- Error handling for missing tools

---

### Scheme Command (`internal/commands/scheme/scheme.go`)

#### Test File: `tests/unit/commands/scheme_test.go`

#### Key Test Scenarios

##### Scheme Operations Tests
```go
func TestSchemeCommand_Operations(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        args      []string
        wantError bool
    }{
        {"list schemes", "list", []string{}, false},
        {"get current", "get", []string{}, false},
        {"set scheme", "set", []string{"catppuccin", "mocha", "dark"}, false},
        {"install bundled", "install", []string{"gruvbox"}, false},
        {"show bundled", "bundled", []string{}, false},
        {"invalid scheme", "set", []string{"nonexistent"}, true},
        {"incomplete args", "set", []string{"catppuccin"}, true},
    }
    // Implementation...
}
```

##### Theme Application Tests
```go
func TestSchemeCommand_ThemeApplication(t *testing.T) {
    tests := []struct {
        name      string
        scheme    *Scheme
        wantError bool
    }{
        {"valid scheme", validTestScheme, false},
        {"scheme with missing colors", incompleteScheme, false}, // Should use defaults
        {"invalid color format", invalidColorScheme, true},
        {"empty scheme", emptyScheme, true},
    }
    // Implementation...
}
```

##### Random Selection Tests
```go
func TestSchemeCommand_RandomSelection(t *testing.T) {
    tests := []struct {
        name           string
        availableSchemes []string
        iterations     int
        wantVariety    bool // Should get different schemes across iterations
    }{
        {"multiple schemes", []string{"catppuccin", "gruvbox", "nord"}, 10, true},
        {"single scheme", []string{"catppuccin"}, 5, false},
        {"no schemes", []string{}, 1, false},
    }
    // Implementation...
}
```

#### Mock Requirements
- Scheme manager interface
- Theme applier interface
- File system operations (scheme loading/saving)
- Embedded asset access

#### Expected Coverage Areas
- All scheme operations (list, get, set, install, bundled)
- Scheme validation and loading
- Theme application workflow
- Random scheme selection
- Bundled scheme installation
- Error handling for invalid schemes

---

### Screenshot Command (`internal/commands/screenshot/screenshot.go`)

#### Test File: `tests/unit/commands/screenshot_test.go`

#### Key Test Scenarios

##### Screenshot Capture Tests
```go
func TestScreenshotCommand_Capture(t *testing.T) {
    tests := []struct {
        name       string
        regionFlag bool
        freezeFlag bool
        wantError  bool
    }{
        {"full screen", false, false, false},
        {"region selection", true, false, false},
        {"freeze screen", false, true, false},
        {"region with freeze", true, true, false},
    }
    // Implementation...
}
```

##### File Management Tests
```go
func TestScreenshotCommand_FileManagement(t *testing.T) {
    tests := []struct {
        name         string
        outputDir    string
        filename     string
        wantPath     string
        wantError    bool
    }{
        {"default directory", "", "", "~/Pictures/Screenshots/", false},
        {"custom directory", "/tmp/screenshots", "", "/tmp/screenshots/", false},
        {"custom filename", "", "test.png", "test.png", false},
        {"invalid directory", "/root/restricted", "", "", true},
    }
    // Implementation...
}
```

##### Integration Tests
```go
func TestScreenshotCommand_Integration(t *testing.T) {
    tests := []struct {
        name           string
        regionFlag     bool
        slurpOutput    string
        grimSuccess    bool
        swappyAvailable bool
        wantError      bool
    }{
        {"successful full capture", false, "", true, false, false},
        {"successful region capture", true, "100,100 200x150", true, false, false},
        {"region with editing", true, "100,100 200x150", true, true, false},
        {"user cancels region", true, "", false, false, false},
        {"grim fails", false, "", false, false, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- External tool executor (grim, slurp, swappy, wl-copy)
- File system operations (directory creation, file writing)
- Clipboard integration

#### Expected Coverage Areas
- Screenshot capture (full screen and region)
- Region selection with slurp
- File naming and directory management
- Optional editing integration (swappy)
- Clipboard integration
- Error handling for tool failures

---

### Shell Command (`internal/commands/shell/shell.go`)

#### Test File: `tests/unit/commands/shell_test.go`

#### Key Test Scenarios

##### Daemon Management Tests
```go
func TestShellCommand_DaemonManagement(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        wantError bool
    }{
        {"start daemon", "start", false},
        {"stop daemon", "stop", false},
        {"kill daemon", "kill", false},
        {"status check", "status", false},
        {"start when running", "start", true},
    }
    // Implementation...
}
```

##### IPC Communication Tests (`tests/unit/commands/shell/ipc_test.go`)
```go
func TestShellIPC_Communication(t *testing.T) {
    tests := []struct {
        name        string
        message     string
        serverResp  string
        wantError   bool
    }{
        {"simple message", "hello", "world", false},
        {"json message", `{"command":"test"}`, `{"result":"ok"}`, false},
        {"server error", "error", "", true},
        {"connection timeout", "timeout", "", true},
    }
    // Implementation...
}
```

##### Process Management Tests
```go
func TestShellCommand_ProcessManagement(t *testing.T) {
    tests := []struct {
        name         string
        mode         string
        pidFileExists bool
        processRunning bool
        wantError    bool
    }{
        {"start attached", "attached", false, false, false},
        {"start daemon", "daemon", false, false, false},
        {"stop running daemon", "stop", true, true, false},
        {"stop non-running", "stop", false, false, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- TCP socket communication
- Process management (PID files, signal handling)
- File system operations (log files, PID files)
- Network interface mocking

#### Expected Coverage Areas
- Daemon lifecycle management
- IPC message handling and protocol
- Process control and signal handling
- Log management and streaming
- Error handling for network issues
- PID file management

---

### Toggle Command (`internal/commands/toggle/toggle.go`)

#### Test File: `tests/unit/commands/toggle_test.go`

#### Key Test Scenarios

##### Workspace Management Tests
```go
func TestToggleCommand_WorkspaceManagement(t *testing.T) {
    tests := []struct {
        name           string
        workspace      string
        currentState   string
        wantAction     string
        wantError      bool
    }{
        {"show hidden workspace", "communication", "hidden", "show", false},
        {"hide visible workspace", "music", "visible", "hide", false},
        {"toggle empty workspace", "sysmon", "empty", "show", false},
        {"invalid workspace", "invalid", "", "", true},
    }
    // Implementation...
}
```

##### Application Detection Tests
```go
func TestToggleCommand_ApplicationDetection(t *testing.T) {
    tests := []struct {
        name         string
        workspace    string
        windows      []Window
        wantApp      string
        wantSpawn    bool
    }{
        {"discord in communication", "communication", discordWindows, "discord", false},
        {"no app in communication", "communication", []Window{}, "", true},
        {"spotify in music", "music", spotifyWindows, "spotify", false},
        {"btop in sysmon", "sysmon", btopWindows, "btop", false},
    }
    // Implementation...
}
```

##### Window Movement Tests
```go
func TestToggleCommand_WindowMovement(t *testing.T) {
    tests := []struct {
        name        string
        window      Window
        targetWS    string
        wantError   bool
    }{
        {"move to special workspace", validWindow, "special:communication", false},
        {"move to regular workspace", validWindow, "1", false},
        {"move nonexistent window", nonexistentWindow, "special:music", true},
        {"invalid workspace", validWindow, "invalid", true},
    }
    // Implementation...
}
```

#### Mock Requirements
- Hyprland IPC client
- Window detection and manipulation
- Process spawning for applications
- Configuration system for app mappings

#### Expected Coverage Areas
- Special workspace detection and management
- Application-specific workspace logic
- Window matching and movement
- Smart workspace toggling behavior
- Application spawning when needed
- Error handling for Hyprland unavailability

---

### Wallpaper Command (`internal/commands/wallpaper/wallpaper.go`)

#### Test File: `tests/unit/commands/wallpaper_test.go`

#### Key Test Scenarios

##### Wallpaper Setting Tests
```go
func TestWallpaperCommand_Setting(t *testing.T) {
    tests := []struct {
        name      string
        path      string
        wantError bool
    }{
        {"valid image file", "fixtures/images/test.png", false},
        {"valid jpg file", "fixtures/images/test.jpg", false},
        {"nonexistent file", "nonexistent.png", true},
        {"invalid image", "fixtures/images/corrupted.png", true},
        {"non-image file", "fixtures/configs/minimal.json", true},
    }
    // Implementation...
}
```

##### Random Selection Tests
```go
func TestWallpaperCommand_RandomSelection(t *testing.T) {
    tests := []struct {
        name         string
        directory    string
        minSize      string
        iterations   int
        wantVariety  bool
    }{
        {"multiple images", "fixtures/images/", "", 10, true},
        {"size filtering", "fixtures/images/", "1920x1080", 5, true},
        {"single image", "fixtures/images/single/", "", 5, false},
        {"empty directory", "fixtures/images/empty/", "", 1, false},
    }
    // Implementation...
}
```

##### Color Extraction Tests
```go
func TestWallpaperCommand_ColorExtraction(t *testing.T) {
    tests := []struct {
        name      string
        imagePath string
        wantColors int
        wantError bool
    }{
        {"colorful image", "fixtures/images/colorful.png", 16, false},
        {"grayscale image", "fixtures/images/gray.png", 16, false},
        {"high contrast", "fixtures/images/contrast.png", 16, false},
        {"corrupted image", "fixtures/images/corrupted.png", 0, true},
    }
    // Implementation...
}
```

##### Material You Generation Tests
```go
func TestWallpaperCommand_MaterialYou(t *testing.T) {
    tests := []struct {
        name      string
        imagePath string
        wantScheme bool
        wantError bool
    }{
        {"generate from image", "fixtures/images/test.png", true, false},
        {"vibrant colors", "fixtures/images/vibrant.png", true, false},
        {"muted colors", "fixtures/images/muted.png", true, false},
        {"invalid image", "fixtures/images/corrupted.png", false, true},
    }
    // Implementation...
}
```

#### Mock Requirements
- External tool executor (hyprctl, swww)
- Image processing libraries
- File system operations (directory scanning)
- Material You color generator
- Wallpaper analyzer

#### Expected Coverage Areas
- Wallpaper setting and validation
- Random wallpaper selection with filtering
- Directory scanning and image detection
- Color scheme extraction from images
- Material You scheme generation
- Integration with wallpaper tools
- Error handling for invalid images

---

## Cross-Cutting Concerns

### Configuration System Testing

#### Test File: `tests/unit/config/manager_test.go`

##### Configuration Loading Tests
```go
func TestConfigManager_Loading(t *testing.T) {
    tests := []struct {
        name         string
        configFiles  []string
        envVars      map[string]string
        wantConfig   map[string]interface{}
        wantError    bool
    }{
        {"single config file", []string{"minimal.json"}, nil, minimalConfig, false},
        {"multiple config files", []string{"base.json", "override.json"}, nil, mergedConfig, false},
        {"environment override", []string{"base.json"}, envOverrides, envMergedConfig, false},
        {"missing config file", []string{"nonexistent.json"}, nil, defaultConfig, false},
        {"invalid config file", []string{"invalid.json"}, nil, nil, true},
    }
    // Implementation...
}
```

##### Configuration Providers Tests
```go
func TestConfigProviders_Integration(t *testing.T) {
    tests := []struct {
        name         string
        provider     string
        config       map[string]interface{}
        wantError    bool
    }{
        {"CLI provider", "cli", cliConfig, false},
        {"shell provider", "shell", shellConfig, false},
        {"invalid provider", "invalid", nil, true},
    }
    // Implementation...
}
```

### Error Handling Patterns

#### Test File: `tests/unit/utils/errors_test.go`

##### Error Wrapping Tests
```go
func TestErrorHandling_Wrapping(t *testing.T) {
    tests := []struct {
        name        string
        baseError   error
        context     string
        wantMessage string
    }{
        {"wrap system error", os.ErrNotExist, "config file", "config file: file does not exist"},
        {"wrap network error", &net.OpError{}, "emoji fetch", "emoji fetch: network operation failed"},
        {"wrap custom error", customError, "validation", "validation: custom error occurred"},
    }
    // Implementation...
}
```

### File System Operations

#### Test File: `tests/unit/utils/filesystem_test.go`

##### Atomic Operations Tests
```go
func TestFileSystem_AtomicOperations(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        content   string
        wantError bool
    }{
        {"atomic write", "write", "test content", false},
        {"atomic update", "update", "updated content", false},
        {"concurrent writes", "concurrent", "content", false},
        {"disk full", "write", largeContent, true},
    }
    // Implementation...
}
```

### Process Management

#### Test File: `tests/unit/utils/process_test.go`

##### Process Control Tests
```go
func TestProcessManagement_Control(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        pid       int
        wantError bool
    }{
        {"start process", "start", 0, false},
        {"stop process", "stop", validPID, false},
        {"kill process", "kill", validPID, false},
        {"check status", "status", validPID, false},
        {"invalid PID", "stop", -1, true},
    }
    // Implementation...
}
```

### IPC Communication

#### Test File: `tests/unit/utils/ipc_test.go`

##### Communication Protocol Tests
```go
func TestIPC_Protocol(t *testing.T) {
    tests := []struct {
        name        string
        message     interface{}
        wantReply   interface{}
        wantError   bool
    }{
        {"simple string", "hello", "world", false},
        {"json object", jsonMessage, jsonReply, false},
        {"binary data", binaryData, binaryReply, false},
        {"invalid message", invalidMessage, nil, true},
    }
    // Implementation...
}
```

## Implementation Phases

### Phase 1: Foundation and Core Utilities (Week 1-2)
**Priority**: Critical
**Dependencies**: None

#### Deliverables
- Test infrastructure setup (`testutils/`)
- Mock implementations for external dependencies
- Core utility tests (color, logger, paths, notify)
- Configuration system tests
- Root command tests

#### Success Criteria
- All utility functions have >90% coverage
- Mock framework is functional and documented
- CI/CD pipeline is running unit tests
- Configuration loading/saving works reliably

### Phase 2: Command Infrastructure (Week 3-4)
**Priority**: High
**Dependencies**: Phase 1

#### Deliverables
- Command framework tests (cobra integration)
- External tool integration tests
- File system operation tests
- Error handling pattern tests

#### Success Criteria
- Command execution framework is tested
- External tool mocking is comprehensive
- File operations are atomic and tested
- Error handling is consistent across commands

### Phase 3: Core Commands (Week 5-7)
**Priority**: High
**Dependencies**: Phase 2

#### Deliverables
- Config command tests (highest complexity)
- Scheme command tests (core functionality)
- Wallpaper command tests (Material You integration)
- Shell command tests (IPC communication)

#### Success Criteria
- Configuration management is fully tested
- Scheme operations work reliably
- Material You generation is tested
- IPC communication is robust

### Phase 4: System Integration Commands (Week 8-9)
**Priority**: Medium
**Dependencies**: Phase 3

#### Deliverables
- Idle command tests (provider system)
- Toggle command tests (Hyprland integration)
- Screenshot command tests (external tool chain)
- Record command tests (process management)

#### Success Criteria
- Provider system is thoroughly tested
- Hyprland integration is mocked and tested
- External tool chains are validated
- Process management is reliable

### Phase 5: Utility Commands (Week 10)
**Priority**: Medium
**Dependencies**: Phase 4

#### Deliverables
- Clipboard command tests
- Emoji command tests
- PIP command tests

#### Success Criteria
- All utility commands are tested
- User interaction flows are validated
- Network operations are mocked and tested

### Phase 6: Integration and E2E Tests (Week 11-12)
**Priority**: Low
**Dependencies**: Phase 5

#### Deliverables
- Integration test suite
- End-to-end workflow tests
- Performance benchmarks
- Documentation and examples

#### Success Criteria
- Integration tests cover command interactions
- E2E tests validate user workflows
- Performance is within acceptable bounds
- Documentation is complete and accurate

## Quality Assurance

### Test Quality Checklist
- [ ] All tests have descriptive names following pattern: `Test[Component]_[Scenario]_[ExpectedOutcome]`
- [ ] Table-driven tests are used for multiple similar scenarios
- [ ] Error cases are tested with specific error message validation
- [ ] Edge cases are identified and tested (empty inputs, boundary values, etc.)
- [ ] Mock objects implement complete interfaces, not just tested methods
- [ ] Tests are independent and can run in any order
- [ ] Test data is isolated and doesn't affect other tests
- [ ] Benchmarks are included for performance-critical code
- [ ] Tests include both positive and negative scenarios
- [ ] Documentation includes examples of running tests

### Coverage Quality Standards
- **Critical Paths**: 100% coverage (configuration loading, scheme application, etc.)
- **Business Logic**: 95% coverage (command operations, data processing)
- **Error Handling**: 90% coverage (all error paths tested)
- **Utility Functions**: 85% coverage (helper functions, formatters)
- **Integration Points**: 80% coverage (external tool interfaces)

### Continuous Improvement
- Weekly coverage reports with trend analysis
- Monthly test performance reviews
- Quarterly test strategy reviews
- Annual testing framework updates

This comprehensive test plan provides a structured approach to testing the entire Heimdall CLI tool, ensuring reliability, maintainability, and comprehensive coverage of all functionality.