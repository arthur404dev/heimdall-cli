# Heimdall CLI Makefile

# Variables
BINARY_NAME := heimdall
PACKAGE := github.com/heimdall-cli/heimdall
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
	-X 'github.com/heimdall-cli/heimdall/internal/commands.Version=$(VERSION)' \
	-X 'github.com/heimdall-cli/heimdall/internal/commands.Commit=$(COMMIT)' \
	-X 'github.com/heimdall-cli/heimdall/internal/commands.Date=$(DATE)' \
	-X 'github.com/heimdall-cli/heimdall/internal/commands.BuiltBy=$(BUILT_BY)' \
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

.PHONY: all build clean test coverage fmt lint install uninstall run help

# Default target
all: clean fmt lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
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

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
coverage: test
	@echo "Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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

# Uninstall the binary from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
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
	@echo "Targets:"
	@echo "  all          - Clean, format, lint, test, and build"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-release- Build optimized release binary"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage report"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  install      - Install binary to /usr/local/bin"
	@echo "  uninstall    - Remove binary from /usr/local/bin"
	@echo "  run          - Build and run the application"
	@echo "  run-args     - Build and run with arguments (ARGS=...)"
	@echo "  dev          - Run with hot reload (requires air)"
	@echo "  deps         - Update dependencies"
	@echo "  verify       - Verify dependencies"
	@echo "  docs         - Generate documentation"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help message"

# Set default goal
.DEFAULT_GOAL := help