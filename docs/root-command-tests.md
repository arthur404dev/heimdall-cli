# Root Command Tests Documentation

This document describes the comprehensive unit test suite for the Heimdall CLI root command functionality.

## Overview

The test suite covers the core command functionality in `internal/commands/root.go` and `internal/commands/test.go`, providing comprehensive validation of:

- Command initialization and setup
- Configuration loading and validation
- Version flag handling
- Help text generation
- Error handling for invalid configurations
- Subcommand registration verification
- Hidden command functionality
- Development utilities testing

## Test Files

### 1. `root_test.go`
Main test file for root command functionality with the following test categories:

#### Core Command Tests (`TestRootCommand`)
- **Help flag functionality**: Validates `--help` flag shows proper usage information
- **Version flag functionality**: Validates `--version` flag displays version information
- **Version command**: Tests the dedicated `version` subcommand
- **Default behavior**: Tests behavior when no arguments are provided
- **Error handling**: Tests invalid command scenarios

#### Flag Tests (`TestRootCommandFlags`)
- **Verbose flag**: Validates `--verbose` flag sets viper configuration
- **Debug flag**: Validates `--debug` flag sets viper configuration  
- **Config flag**: Tests `--config` flag with custom configuration file paths

#### Configuration Tests (`TestInitConfig`)
- **Home directory config**: Tests loading config from `~/.config/heimdall/config.json`
- **Backward compatibility**: Tests loading legacy config from `~/.config/caelestia/config.json`
- **Environment variables**: Tests environment variable override functionality

#### Registration Tests (`TestSubcommandRegistration`)
- Validates all expected subcommands are properly registered
- Checks for presence of: config, shell, toggle, scheme, screenshot, record, clipboard, emoji, wallpaper, pip, idle, version, test

#### Metadata Tests (`TestCommandMetadata`, `TestVersionInformation`)
- Validates command metadata (Use, Short, Long descriptions)
- Tests version information consistency
- Validates help text includes all important flags

#### Error Handling Tests (`TestErrorHandling`)
- Tests invalid config file paths
- Tests unknown flag scenarios
- Validates graceful error handling

#### Performance Tests
- **BenchmarkRootCommandCreation**: Measures command creation performance
- **BenchmarkVersionCommand**: Measures version command execution performance

### 2. `test_test.go`
Test file for the hidden test command functionality:

#### Test Command Tests (`TestTestCommand`)
- **Basic execution**: Validates test command runs successfully
- **Hyprland integration**: Tests behavior with/without Hyprland running
- **Hidden functionality**: Validates command is hidden from help

#### Utility Tests
- **Color utilities**: Tests color generation and manipulation
- **Hyprland IPC**: Tests Hyprland communication functionality
- **Notifications**: Tests notification system integration

#### Output Tests
- **Section validation**: Ensures all expected output sections are present
- **Format validation**: Validates output format and structure
- **Error handling**: Tests graceful handling of utility failures

### 3. `testutils_test.go`
Comprehensive test utilities and helpers:

#### TestUtilities Class
- **Environment management**: Set/unset environment variables with cleanup
- **File system utilities**: Create temporary files and directories
- **Output capture**: Capture stdout/stderr from functions
- **Cleanup management**: Automatic cleanup of test resources

#### MockCommand Builder
- **Fluent API**: Builder pattern for creating mock cobra commands
- **Flag support**: Add string/bool flags to mock commands
- **Subcommand support**: Add nested subcommands
- **Flexible configuration**: Support for various command configurations

#### CommandTester Class
- **Execution testing**: Execute commands with arguments
- **Timeout support**: Execute commands with timeout limits
- **Output validation**: Assert on stdout/stderr content
- **Buffer management**: Reset and manage output buffers

#### ConfigTester Class
- **Configuration management**: Set/restore viper configuration
- **Temporary configs**: Create and load temporary config files
- **State isolation**: Ensure test isolation with proper cleanup

#### TestAssertions Class
- **Standard assertions**: NoError, Error, Equal, NotEqual
- **String assertions**: Contains, NotContains
- **Boolean assertions**: True, False
- **Null assertions**: Nil, NotNil

### 4. `mocks_test.go`
Mock implementations for external dependencies:

#### MockLogger
- **Message tracking**: Track all logged messages by level
- **Level filtering**: Filter messages by log level
- **Thread-safe**: Safe for concurrent use
- **Validation helpers**: Check for specific messages

#### MockHyprlandClient
- **State simulation**: Simulate Hyprland running/not running states
- **Command mocking**: Mock IPC command responses
- **Workspace/Window simulation**: Mock workspace and window data
- **Error simulation**: Simulate various error conditions

#### MockNotifier
- **Notification tracking**: Track all sent notifications
- **Availability simulation**: Simulate notification system availability
- **Error simulation**: Simulate notification sending failures
- **Validation helpers**: Check for specific notifications

#### MockFileSystem
- **File operations**: Mock file read/write operations
- **Directory simulation**: Simulate directory structures
- **Error simulation**: Simulate file system errors
- **State management**: Track files and directories

## Test Coverage

The test suite provides comprehensive coverage of:

### ✅ Covered Functionality
- Command initialization and setup
- Flag parsing and binding
- Configuration loading (multiple sources)
- Version information display
- Help text generation
- Subcommand registration
- Error handling and validation
- Environment variable processing
- Backward compatibility
- Performance characteristics

### ⚠️ Limitations
- **Concurrent execution**: Disabled due to viper global state race conditions
- **Real external dependencies**: Uses mocks instead of real Hyprland/notification systems
- **File system operations**: Limited to basic scenarios
- **Complex configuration scenarios**: Some edge cases may not be covered

## Running Tests

### Run All Tests
```bash
go test ./internal/commands -v
```

### Run Specific Test Categories
```bash
# Root command tests only
go test ./internal/commands -v -run TestRootCommand

# Test command tests only
go test ./internal/commands -v -run TestTestCommand

# Configuration tests only
go test ./internal/commands -v -run TestInitConfig
```

### Run Benchmarks
```bash
# All benchmarks
go test ./internal/commands -bench=.

# Specific benchmark
go test ./internal/commands -bench=BenchmarkRootCommandCreation
```

### Generate Coverage Report
```bash
go test ./internal/commands -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Best Practices Demonstrated

### 1. Test Organization
- **Table-driven tests**: Used for testing multiple scenarios
- **Descriptive names**: Test names clearly describe what is being tested
- **Logical grouping**: Related tests are grouped together

### 2. Test Isolation
- **Independent tests**: Each test is independent and can run alone
- **Cleanup management**: Proper cleanup of resources and state
- **Mock usage**: External dependencies are mocked for reliability

### 3. Comprehensive Coverage
- **Happy path testing**: Normal operation scenarios
- **Error path testing**: Error conditions and edge cases
- **Performance testing**: Benchmark tests for performance regression detection

### 4. Maintainable Code
- **Helper functions**: Reusable test utilities
- **Mock implementations**: Comprehensive mocks for external dependencies
- **Clear assertions**: Descriptive error messages for test failures

## Future Improvements

### 1. Enhanced Coverage
- Add integration tests with real external dependencies
- Add more edge case scenarios
- Improve configuration loading test coverage

### 2. Performance Optimization
- Optimize test execution time
- Add more granular performance benchmarks
- Profile memory usage in tests

### 3. Test Infrastructure
- Add test data generators
- Implement property-based testing
- Add mutation testing for test quality validation

### 4. Documentation
- Add more inline documentation
- Create test case documentation
- Add troubleshooting guides

## Conclusion

This comprehensive test suite provides robust validation of the Heimdall CLI root command functionality. It follows Go testing best practices, provides excellent coverage of critical paths, and includes helpful utilities for maintaining and extending the test suite.

The tests serve as both validation and documentation of the expected behavior, making it easier for developers to understand and modify the codebase with confidence.