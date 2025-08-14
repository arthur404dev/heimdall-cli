# Heimdall CLI Makefile

# Variables
BINARY_NAME := heimdall
PACKAGE := github.com/arthur404dev/heimdall-cli
CMD_PATH := ./cmd/heimdall
BUILD_DIR := ./build
DIST_DIR := ./dist

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILT_BY := $(shell whoami)

# Go build flags
LDFLAGS := -ldflags "\
	-X 'github.com/arthur404dev/heimdall-cli/internal/commands.Version=$(VERSION)' \
	-X 'github.com/arthur404dev/heimdall-cli/internal/commands.Commit=$(COMMIT)' \
	-X 'github.com/arthur404dev/heimdall-cli/internal/commands.Date=$(DATE)' \
	-X 'github.com/arthur404dev/heimdall-cli/internal/commands.BuiltBy=$(BUILT_BY)' \
	-s -w"

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 linux/386 freebsd/amd64

.PHONY: all build clean test test-unit test-integration test-commands test-coverage test-coverage-html test-bench test-race test-short test-verbose test-watch test-clean fmt lint install uninstall run help

# Default target
all: clean fmt lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Quick build (no version info, for development)
quick:
	@echo "Quick building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} $(CMD_PATH); \
		echo "Built: $(DIST_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done

# Build optimized binary (smaller size)
build-release:
	@echo "Building optimized release binary..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -trimpath -o $(DIST_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Release build complete: $(DIST_DIR)/$(BINARY_NAME)"
	@echo "Binary size: $$(du -h $(DIST_DIR)/$(BINARY_NAME) | cut -f1)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "Clean complete"

# ============================================================================
# TEST TARGETS
# ============================================================================

# Run all tests (default test target)
test:
	@echo "Running all tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests completed. Coverage report: coverage.out"

# Run unit tests only (fast tests without integration)
test-unit:
	@echo "Running unit tests..."
	@$(GOTEST) -v -short -race ./...

# Run integration tests only (slower tests)
test-integration:
	@echo "Running integration tests..."
	@$(GOTEST) -v -run Integration ./...

# Run tests for specific commands
test-commands:
	@echo "Running command tests..."
	@$(GOTEST) -v -race ./internal/commands/...

# Run tests with coverage analysis
test-coverage:
	@echo "Running tests with coverage analysis..."
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -func=coverage.out
	@echo "Coverage profile saved to: coverage.out"

# Generate HTML coverage report
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view detailed coverage"

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	@$(GOTEST) -v -bench=. -benchmem ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	@$(GOTEST) -v -race ./...

# Run short tests only (skip long-running tests)
test-short:
	@echo "Running short tests..."
	@$(GOTEST) -v -short ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@$(GOTEST) -v ./...

# Run tests and watch for changes (requires entr)
test-watch:
	@echo "Running tests in watch mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -c $(GOTEST) -v ./...; \
	else \
		echo "entr not installed. Install with your package manager (e.g., apt install entr)"; \
		echo "Falling back to single test run..."; \
		$(MAKE) test; \
	fi

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -f *.test
	@rm -rf test-results/
	@echo "Test artifacts cleaned"

# Run specific test by name (usage: make test-run TEST=TestFunctionName)
test-run:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-run TEST=TestFunctionName"; \
		echo "Example: make test-run TEST=TestRootCommand"; \
		exit 1; \
	fi
	@echo "Running test: $(TEST)"
	@$(GOTEST) -v -run "$(TEST)" ./...

# Run tests for specific package (usage: make test-pkg PKG=./internal/commands/config)
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=./path/to/package"; \
		echo "Example: make test-pkg PKG=./internal/commands/config"; \
		exit 1; \
	fi
	@echo "Running tests for package: $(PKG)"
	@$(GOTEST) -v -race $(PKG)

# Run tests with timeout (usage: make test-timeout TIMEOUT=30s)
test-timeout:
	@TIMEOUT=$${TIMEOUT:-10s}; \
	echo "Running tests with timeout: $$TIMEOUT"; \
	$(GOTEST) -v -timeout $$TIMEOUT ./...

# Run tests and generate JUnit XML report (requires go-junit-report)
test-junit:
	@echo "Running tests and generating JUnit XML report..."
	@mkdir -p test-results
	@if command -v go-junit-report >/dev/null 2>&1; then \
		$(GOTEST) -v ./... 2>&1 | go-junit-report > test-results/junit.xml; \
		echo "JUnit XML report generated: test-results/junit.xml"; \
	else \
		echo "go-junit-report not installed. Install with: go install github.com/jstemmer/go-junit-report@latest"; \
		$(MAKE) test; \
	fi

# Run tests with memory profiling
test-memprofile:
	@echo "Running tests with memory profiling..."
	@$(GOTEST) -v -memprofile=mem.prof ./...
	@echo "Memory profile saved to: mem.prof"
	@echo "View with: go tool pprof mem.prof"

# Run tests with CPU profiling
test-cpuprofile:
	@echo "Running tests with CPU profiling..."
	@$(GOTEST) -v -cpuprofile=cpu.prof ./...
	@echo "CPU profile saved to: cpu.prof"
	@echo "View with: go tool pprof cpu.prof"

# Run all test variants (comprehensive test suite)
test-all: test-clean
	@echo "Running comprehensive test suite..."
	@echo "1. Unit tests..."
	@$(MAKE) test-unit
	@echo "2. Integration tests..."
	@$(MAKE) test-integration
	@echo "3. Race detection..."
	@$(MAKE) test-race
	@echo "4. Benchmarks..."
	@$(MAKE) test-bench
	@echo "5. Coverage analysis..."
	@$(MAKE) test-coverage-html
	@echo "Comprehensive test suite completed!"

# Legacy coverage target (for backward compatibility)
coverage: test-coverage-html

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@$(GOMOD) tidy
	@echo "Format complete"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install the binary to system
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Install the binary to user's local bin (~/.local/bin)
install-local: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@if [ -f ~/.local/bin/$(BINARY_NAME) ]; then \
		mv ~/.local/bin/$(BINARY_NAME) ~/.local/bin/$(BINARY_NAME).old 2>/dev/null || true; \
	fi
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@rm -f ~/.local/bin/$(BINARY_NAME).old
	@echo "Installation complete at ~/.local/bin/$(BINARY_NAME)"
	@echo "Make sure ~/.local/bin is in your PATH"

# Update the binary in ~/.local/bin (alias for install-local)
update: install-local

# Uninstall the binary from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Uninstall the binary from user's local bin
uninstall-local:
	@echo "Uninstalling $(BINARY_NAME) from ~/.local/bin..."
	@rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run with arguments
run-args: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Development run (with hot reload using air)
dev:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		$(MAKE) run; \
	fi

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "Dependencies updated"

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@$(GOMOD) verify
	@echo "Dependencies verified"

# Generate documentation
docs:
	@echo "Generating documentation..."
	@$(GOCMD) doc -all > docs/API.md
	@echo "Documentation generated: docs/API.md"

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@echo "Built by: $(BUILT_BY)"

# Help target
help:
	@echo "Heimdall CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  all          - Clean, format, lint, test, and build"
	@echo "  build        - Build the binary with version info"
	@echo "  quick        - Quick build for development (no version info)"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-release- Build optimized release binary"
	@echo "  clean        - Remove build artifacts"
	@echo ""
	@echo "Test Targets:"
	@echo "  test         - Run all tests with race detection and coverage"
	@echo "  test-unit    - Run unit tests only (fast)"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-commands- Run command tests specifically"
	@echo "  test-coverage- Run tests with coverage analysis"
	@echo "  test-coverage-html - Generate HTML coverage report"
	@echo "  test-bench   - Run benchmark tests"
	@echo "  test-race    - Run tests with race detection"
	@echo "  test-short   - Run short tests only"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-watch   - Run tests in watch mode (requires entr)"
	@echo "  test-clean   - Clean test artifacts"
	@echo "  test-run     - Run specific test (TEST=TestName)"
	@echo "  test-pkg     - Run tests for specific package (PKG=./path)"
	@echo "  test-timeout - Run tests with timeout (TIMEOUT=30s)"
	@echo "  test-junit   - Generate JUnit XML report (requires go-junit-report)"
	@echo "  test-memprofile - Run tests with memory profiling"
	@echo "  test-cpuprofile - Run tests with CPU profiling"
	@echo "  test-all     - Run comprehensive test suite"
	@echo "  coverage     - Alias for test-coverage-html"
	@echo ""
	@echo "Development Targets:"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  run          - Build and run the application"
	@echo "  run-args     - Build and run with arguments (ARGS=...)"
	@echo "  dev          - Run with hot reload (requires air)"
	@echo ""
	@echo "Installation Targets:"
	@echo "  install      - Install binary to /usr/local/bin (requires sudo)"
	@echo "  install-local- Install binary to ~/.local/bin"
	@echo "  update       - Update binary in ~/.local/bin (alias for install-local)"
	@echo "  uninstall    - Remove binary from /usr/local/bin (requires sudo)"
	@echo "  uninstall-local - Remove binary from ~/.local/bin"
	@echo ""
	@echo "Utility Targets:"
	@echo "  deps         - Update dependencies"
	@echo "  verify       - Verify dependencies"
	@echo "  docs         - Generate documentation"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make test-run TEST=TestRootCommand"
	@echo "  make test-pkg PKG=./internal/commands/config"
	@echo "  make test-timeout TIMEOUT=30s"
	@echo "  make run-args ARGS='--help'"

# Set default goal
.DEFAULT_GOAL := help