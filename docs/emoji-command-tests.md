# Emoji Command Test Documentation

## Overview

This document describes the comprehensive test suite for the emoji command functionality in `internal/commands/emoji/emoji.go`. The tests cover all major functionality including command initialization, emoji database management, external tool integration, and error handling.

## Test Structure

### Test Files
- `internal/commands/emoji/emoji_test.go` - Main test file with comprehensive coverage

### Test Categories

#### 1. Command Structure Tests (`TestCommand`)
- **Purpose**: Validates command initialization and structure
- **Coverage**:
  - Command creation and basic properties
  - Flag registration (`--fetch`, `--picker`)
  - Command metadata (Use, Short, Long descriptions)
  - RunE function assignment

#### 2. Database Management Tests (`TestUpdateEmojiData`)
- **Purpose**: Tests emoji database fetching and updating
- **Coverage**:
  - Successful data fetching from remote sources
  - HTTP error handling (server errors, timeouts)
  - File system operations (directory creation, file writing)
  - Multiple data source handling (emoji.json, nerd-fonts.json)
- **Mock Strategy**: Uses `httptest.Server` to simulate remote API responses

#### 3. Data Loading Tests (`TestLoadEmojiData`)
- **Purpose**: Tests local emoji data loading and parsing
- **Coverage**:
  - Successful JSON parsing and emoji structure validation
  - Missing file error handling
  - Malformed JSON error handling
  - Empty data file handling
- **Test Data**: Uses `createSampleEmojiData()` helper for consistent test data

#### 4. Search Functionality Tests (`TestSearchEmoji`)
- **Purpose**: Tests emoji search by aliases and tags
- **Coverage**:
  - Search by alias matching
  - Search by tag matching
  - Case-insensitive search
  - Multiple result handling
  - No results found scenarios
  - Empty query handling
- **Output Validation**: Captures stdout to verify search results

#### 5. Interactive Picker Tests (`TestRunEmojiPicker`)
- **Purpose**: Tests the interactive emoji picker functionality
- **Coverage**:
  - Successful picker execution
  - Missing data auto-fetching
  - User cancellation handling
  - External tool (fuzzel) failure handling
- **Limitations**: Tests are environment-aware and handle fuzzel display issues gracefully

#### 6. Clipboard Integration Tests (`TestCopyToClipboard`)
- **Purpose**: Tests clipboard operations
- **Coverage**:
  - Default wl-copy command usage
  - Custom wl-copy path configuration
  - Configuration loading error handling
- **Environment Handling**: Gracefully handles missing clipboard tools in test environment

#### 7. Command Execution Tests (`TestCommandExecution`)
- **Purpose**: Integration tests for complete command execution
- **Coverage**:
  - Default picker mode execution
  - Explicit picker flag usage
  - Search mode execution
  - Fetch mode execution
- **Integration**: Tests full command flow with real configuration

#### 8. Data Structure Tests (`TestEmojiStructure`)
- **Purpose**: Validates emoji data structure and JSON serialization
- **Coverage**:
  - JSON marshaling/unmarshaling
  - Field validation
  - Data integrity checks

#### 9. Performance Tests (Benchmarks)
- **BenchmarkLoadEmojiData**: Tests emoji data loading performance with 1000 emoji dataset
- **BenchmarkSearchEmoji**: Tests search performance with sample dataset

#### 10. Integration Tests (`TestEmojiCommandIntegration`)
- **Purpose**: End-to-end integration testing
- **Coverage**:
  - Complete search workflow
  - Command structure validation
- **Execution**: Skipped in short mode, runs full integration in normal mode

## Test Utilities and Mocks

### Custom Test Utilities
- **TestUtilities**: Provides common test setup and cleanup
- **setupTestConfig**: Configures test environment with temporary directories
- **createSampleEmojiData**: Generates consistent test emoji data

### Mock Implementations
- **MockHTTPClient**: Simulates HTTP responses for data fetching tests
- **MockExecCommand**: Provides mock command execution (planned for future use)

### Test Data Management
- **Temporary Directories**: Each test uses isolated temporary directories
- **Configuration Mocking**: Uses viper to mock configuration settings
- **Environment Cleanup**: Automatic cleanup of test artifacts

## Error Handling Strategy

### Expected Errors
- Missing emoji data files
- Malformed JSON data
- Network failures during data fetching
- External tool unavailability

### Environment-Aware Testing
- **Fuzzel Failures**: Tests handle fuzzel display connection issues gracefully
- **Clipboard Tool Availability**: Tests adapt to missing wl-copy/xclip tools
- **Network Conditions**: HTTP tests simulate various network scenarios

## Test Coverage Areas

### ✅ Fully Covered
- Command structure and initialization
- Emoji data loading and parsing
- Search functionality (aliases, tags, case-insensitive)
- Database updating with HTTP mocking
- Error handling for missing/malformed data
- JSON serialization/deserialization
- Configuration integration
- Performance benchmarking

### ⚠️ Partially Covered
- External tool integration (fuzzel, wl-copy) - limited by test environment
- Interactive picker flow - depends on display availability
- Real network operations - mocked for reliability

### 🔄 Future Enhancements
- Dependency injection for better external tool mocking
- More comprehensive integration tests with containerized environments
- Performance tests with larger datasets
- Concurrent access testing

## Running Tests

### Basic Test Execution
```bash
# Run all emoji tests
go test ./internal/commands/emoji/... -v

# Run tests in short mode (skips integration tests)
go test ./internal/commands/emoji/... -v -short

# Run specific test
go test ./internal/commands/emoji/... -run TestSearchEmoji -v
```

### Benchmark Tests
```bash
# Run all benchmarks
go test ./internal/commands/emoji/... -bench=. -run=^$

# Run specific benchmark
go test ./internal/commands/emoji/... -bench=BenchmarkLoadEmojiData -run=^$
```

### Coverage Analysis
```bash
# Generate coverage report
go test ./internal/commands/emoji/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Quality Metrics

### Test Characteristics
- **Deterministic**: All tests produce consistent results
- **Isolated**: Each test uses independent temporary directories
- **Fast**: Unit tests complete in milliseconds (excluding network operations)
- **Comprehensive**: Covers happy paths, error cases, and edge conditions
- **Maintainable**: Clear test names and well-structured test data

### Best Practices Followed
- **AAA Pattern**: Arrange, Act, Assert structure
- **Table-Driven Tests**: Used for multiple input scenarios
- **Mock External Dependencies**: HTTP servers, file systems
- **Environment Cleanup**: Automatic resource cleanup
- **Clear Error Messages**: Descriptive test failure messages

## Dependencies

### Test Dependencies
- Standard Go testing package
- `net/http/httptest` for HTTP mocking
- `github.com/spf13/viper` for configuration testing
- Temporary file system operations

### External Tools (Optional)
- `fuzzel` - Interactive selector (tests handle absence gracefully)
- `wl-copy` - Clipboard operations (tests handle absence gracefully)

## Maintenance Notes

### Adding New Tests
1. Follow existing naming conventions (`TestFunctionName`)
2. Use table-driven tests for multiple scenarios
3. Include both success and failure cases
4. Add appropriate cleanup in test utilities
5. Update this documentation

### Modifying Existing Tests
1. Ensure backward compatibility
2. Update test data if emoji structure changes
3. Maintain environment-aware error handling
4. Update documentation for significant changes

### Performance Considerations
- Benchmark tests use realistic data sizes
- Network operations are mocked for speed
- File operations use temporary directories
- Tests clean up resources promptly