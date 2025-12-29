# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-cli-template is a Go CLI application configured for cross-platform distribution using GoReleaser.

**Module**: `com.github.olifink.go-cli-template`
**Go Version**: 1.25.5

## Build and Development Commands

### Local Development
```bash
# Build the application
go build -o go-cli-template main.go

# Run the application
go run main.go

# Format code
go fmt ./...

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./path/to/package

# Tidy dependencies
go mod tidy

# Run code generation
go generate ./...
```

### Release Building
```bash
# Build release artifacts for all platforms (Linux, Windows, macOS)
goreleaser build --snapshot --clean

# Create a full release (requires tags and proper git state)
goreleaser release --snapshot --clean

# Test the release configuration without building
goreleaser check
```

## Architecture

The project is currently minimal with:
- `main.go`: Entry point containing the main function
- `src/`: Directory for future source code organization (currently empty)
- `.goreleaser.yaml`: GoReleaser configuration for cross-platform builds with CGO disabled

Future code should be organized in the `src/` directory with proper package structure as the project grows.
