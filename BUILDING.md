# Building Heimdall CLI

## Quick Start

Always use the Makefile to build Heimdall to ensure binaries are placed in the correct location:

```bash
# Standard build with version information
make build

# Quick build for development (faster, no version info)
make quick

# Install to ~/.local/bin
make install-local
# or
make update
```

## Build Output

All build artifacts are placed in the `./build/` directory. The project is configured to prevent binaries from being created at the project root.

## Available Make Targets

### Building
- `make build` - Build with full version information
- `make quick` - Quick development build (no version info)
- `make build-all` - Build for all supported platforms
- `make build-release` - Create optimized release binary

### Installation
- `make install` - Install to /usr/local/bin (requires sudo)
- `make install-local` - Install to ~/.local/bin (no sudo needed)
- `make update` - Alias for install-local, updates your local installation

### Development
- `make run` - Build and run immediately
- `make run-args ARGS="..."` - Build and run with arguments
- `make dev` - Run with hot reload (requires air)

### Maintenance
- `make clean` - Remove all build artifacts
- `make fmt` - Format code
- `make lint` - Run linter
- `make test` - Run tests
- `make coverage` - Generate test coverage report

## Manual Building

If you need to build manually for any reason, always specify the output directory:

```bash
# Correct way - outputs to build/
go build -o build/heimdall cmd/heimdall/main.go

# Or use the provided wrapper
./gobuild

# WRONG - creates binary at project root
# go build cmd/heimdall/main.go  # Don't do this!
```

## Platform-Specific Builds

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o build/heimdall-linux-amd64 cmd/heimdall/main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o build/heimdall-linux-arm64 cmd/heimdall/main.go

# Or use make
make build-all
```

## Troubleshooting

### Binary at Project Root

If you accidentally create a binary at the project root:
1. It will be ignored by git (listed in .gitignore)
2. Remove it with: `rm heimdall`
3. Use `make build` or `make quick` instead

### Build Fails

1. Ensure Go 1.21+ is installed: `go version`
2. Update dependencies: `make deps`
3. Clean and rebuild: `make clean && make build`

## CI/CD

The project uses GitHub Actions for automated builds. Every commit triggers:
- Code formatting check
- Linting
- Tests
- Multi-platform builds

Release binaries are automatically built and published when a new tag is pushed.