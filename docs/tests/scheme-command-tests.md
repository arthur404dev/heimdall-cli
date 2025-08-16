# Scheme Command Test Suite

This document describes the comprehensive unit test suite created for the scheme command functionality in `internal/commands/scheme/`.

## Overview

The scheme command is one of the most complex commands in heimdall-cli, featuring multiple subcommands, Material You integration, file system operations, network operations, and theme application logic. The test suite provides comprehensive coverage of all functionality.

## Test Files Created

### 1. Main Command Tests (`scheme_test.go`)
Tests the main scheme command structure and subcommand registration:

- **Command Creation**: Verifies the scheme command is created with correct properties
- **Subcommand Registration**: Ensures all 5 subcommands (list, get, set, install, bundled) are properly registered
- **Command Metadata**: Validates command descriptions and help text
- **Integration**: Tests command execution without arguments and with invalid subcommands

### 2. Mock Infrastructure (`mocks_test.go`)
Provides comprehensive mock implementations for testing:

- **MockSchemeManager**: Full mock implementation of the scheme manager interface
  - Supports all CRUD operations on schemes
  - Tracks method calls for verification
  - Configurable error injection
  - Thread-safe implementation

- **MockBundledSchemeProvider**: Mock for bundled scheme operations
  - Simulates embedded scheme loading
  - Configurable scheme data
  - Error simulation capabilities

- **MockThemeApplier**: Mock for theme application
  - Tracks applied themes
  - Configurable error responses
  - Verification of theme application calls

- **MockNotifier**: Mock notification system
  - Tracks sent notifications
  - Configurable availability and errors
  - Verification of notification content

### 3. Bundled Command Tests (`bundled_test.go`)
Tests the `scheme bundled` subcommand:

- **Command Creation**: Verifies command structure and properties
- **Basic Execution**: Tests command runs without errors
- **Output Verification**: Ensures command produces expected output
- **Integration**: Tests real bundled scheme loading

**Key Test Scenarios:**
- Command creation and metadata validation
- Successful execution with bundled schemes
- Handling of empty scheme lists
- Error handling for scheme loading failures

### 4. Get Command Tests (`get_test.go`)
Tests the `scheme get` subcommand with comprehensive flag and argument testing:

- **Command Creation**: Validates command structure and all flags
- **Flag Operations**: Tests all individual flags (-n, -f, -m, -v, --json, --no-color)
- **Property Retrieval**: Tests getting specific properties and colors
- **JSON Output**: Validates JSON formatting and content
- **Error Handling**: Tests unknown properties and manager errors

**Key Test Scenarios:**
- Default scheme information display
- Individual flag operations (name, flavour, mode, variant)
- JSON output formatting
- Specific property and color retrieval
- No-color output mode
- Error handling for invalid properties

### 5. Install Command Tests (`install_test.go`)
Tests the `scheme install` subcommand:

- **Command Creation**: Verifies command structure and flags
- **Installation Operations**: Tests individual and bulk scheme installation
- **Argument Handling**: Tests scheme names with spaces
- **Error Handling**: Tests various failure scenarios

**Key Test Scenarios:**
- Installing all bundled schemes with `--all` flag
- Installing specific schemes by name
- Handling scheme names with spaces
- Listing available schemes when no arguments provided
- Error handling for missing schemes and installation failures

### 6. List Command Tests (`list_test.go`)
Tests the `scheme list` subcommand with complex flag combinations:

- **Command Creation**: Validates command structure and all flags
- **Flag Operations**: Tests all listing modes (-n, -f, -m, -v)
- **Caelestia Format**: Tests JSON output compatibility
- **Specific Queries**: Tests scheme-specific and flavour-specific listings

**Key Test Scenarios:**
- Listing scheme names only
- Listing flavours for current or specific schemes
- Listing modes for scheme/flavour combinations
- Material You variant listing
- Caelestia-compatible JSON output
- Error handling for manager failures

### 7. Set Command Tests (`set_test.go`)
Tests the most complex `scheme set` subcommand:

- **Command Creation**: Validates command structure and all flags
- **Positional Arguments**: Tests scheme setting with arguments
- **Flag-based Setting**: Tests setting schemes using flags
- **Random Selection**: Tests random scheme selection
- **Theme Application**: Tests theme application and --no-apply flag
- **Notifications**: Tests notification integration
- **Error Handling**: Comprehensive error scenario testing

**Key Test Scenarios:**
- Setting schemes with positional arguments
- Setting schemes using individual flags
- Random scheme selection with `-r` flag
- Theme application control with `--no-apply`
- Desktop notifications with `--notify`
- Default flavour selection when not specified
- Error handling for invalid modes, missing schemes, and theme failures

## Test Architecture

### Dependency Injection Pattern
The tests use a dependency injection pattern with function variables that can be mocked:

```go
var newManagerFunc = func() SchemeManagerInterface {
    return scheme.NewManager()
}
```

This allows tests to inject mock implementations without modifying the original code.

### Interface-Based Mocking
All major dependencies are abstracted behind interfaces:

- `SchemeManagerInterface`: Abstracts scheme management operations
- Mock implementations provide full interface compatibility
- Thread-safe mock implementations with proper synchronization

### Comprehensive Error Testing
Each test file includes extensive error testing:

- Manager operation failures
- Network and file system errors
- Invalid input validation
- Theme application failures
- Notification system errors

## Test Coverage

The test suite provides comprehensive coverage of:

### Functional Areas
- ✅ Command structure and metadata
- ✅ All subcommand functionality
- ✅ Flag parsing and validation
- ✅ Argument processing
- ✅ Scheme management operations
- ✅ Theme application
- ✅ Notification integration
- ✅ Error handling and validation

### Edge Cases
- ✅ Empty scheme lists
- ✅ Invalid arguments and flags
- ✅ Network failures
- ✅ File system errors
- ✅ Missing dependencies
- ✅ Malformed data
- ✅ Concurrent operations

### Integration Scenarios
- ✅ Real command execution
- ✅ Manager integration
- ✅ Theme system integration
- ✅ Notification system integration

## Running the Tests

### Run All Scheme Tests
```bash
go test ./internal/commands/scheme/... -v
```

### Run Specific Test Files
```bash
go test ./internal/commands/scheme/ -run TestCommand -v
go test ./internal/commands/scheme/ -run TestBundledCommand -v
go test ./internal/commands/scheme/ -run TestGetCommand -v
go test ./internal/commands/scheme/ -run TestInstallCommand -v
go test ./internal/commands/scheme/ -run TestListCommand -v
go test ./internal/commands/scheme/ -run TestSetCommand -v
```

### Run with Coverage
```bash
go test ./internal/commands/scheme/... -cover -v
```

## Test Quality Standards

The test suite follows these quality standards:

### Meaningful Tests
- Every test has a clear, specific purpose
- Test names describe the scenario and expected outcome
- No placeholder or trivial tests

### Comprehensive Coverage
- All public methods and functions tested
- Both success and failure paths covered
- Edge cases and error conditions included

### Maintainable Code
- DRY principles with helper functions and fixtures
- Clear test structure with Arrange-Act-Assert pattern
- Proper cleanup and resource management

### Fast Execution
- Tests run quickly (< 1 second for unit tests)
- Minimal external dependencies
- Efficient mock implementations

### Deterministic Results
- No random failures or flaky tests
- Proper synchronization for concurrent operations
- Consistent test data and expectations

## Future Enhancements

Potential areas for test suite enhancement:

1. **Performance Testing**: Add benchmarks for scheme operations
2. **Concurrency Testing**: More extensive concurrent operation testing
3. **Integration Testing**: End-to-end testing with real file systems
4. **Property-Based Testing**: Generate random test data for robustness
5. **Mutation Testing**: Verify test quality with mutation testing tools

## Conclusion

This comprehensive test suite provides excellent coverage of the scheme command functionality, ensuring reliability and maintainability of this critical component. The tests follow Go best practices and provide a solid foundation for continued development and refactoring.