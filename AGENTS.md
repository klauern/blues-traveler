# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `klauer-hooks`, a CLI tool for managing and running Claude Code hooks. It has evolved from a template generator to a direct hook runner with a plugin architecture. The tool integrates with Claude Code's hook system to provide security, formatting, debugging, and audit capabilities.

## Architecture

### Core Components

- **Plugin Registry** (`plugin.go`): Dynamic plugin registration system with thread-safe access
- **Hook Implementations** (`hooks.go`): Concrete hook functions (security, format, debug, audit)
- **Settings Management** (`settings.go`): Handles both project-local (`./.claude/settings.json`) and global (`~/.claude/settings.json`) configuration
- **CLI Commands** (`commands.go`, `main.go`): Cobra-based CLI for listing, running, installing, and managing hooks

### Plugin System

The application uses a plugin architecture where:
- Plugins implement the `Plugin` interface with `Name()`, `Description()`, and `Run()` methods
- Built-in plugins register themselves via `init()` functions in `plugin.go`
- Plugins can be enabled/disabled via settings configuration
- Plugin state is checked hierarchically (project settings override global settings)

### Hook Types

- **security**: Blocks dangerous commands using pattern matching, regex detection, and macOS-specific protections
- **format**: Auto-formats code files after Edit/Write operations (supports Go, JS/TS, Python)
- **debug**: Logs all tool usage to `claude-hooks.log`
- **audit**: Comprehensive JSON audit logging to stdout

## Development Commands

### Build and Test
```bash
# Build the binary
task build

# Run all checks (format, lint, test)
task check

# Development workflow (format, lint, test, build)
task dev

# Run tests with coverage
task test-coverage
```

### Linting and Formatting
```bash
# Format Go code
task format

# Run linter (golangci-lint preferred, falls back to go vet)
task lint
```

### Running
```bash
# List available plugins
./hooks list

# Run a specific plugin
./hooks run <plugin-key>

# Install hook into Claude Code settings
./hooks install <plugin-key> [flags]

# List installed hooks from settings
./hooks list-installed
```

## File Structure

- `main.go`: CLI entry point with Cobra commands
- `plugin.go`: Plugin registry and built-in plugin registration
- `hooks.go`: Hook implementation functions using cchooks library
- `settings.go`: Settings file management with JSON marshaling
- `commands.go`: CLI command implementations
- `Taskfile.yml`: Task runner configuration for build/test/lint workflows

## Dependencies

- `github.com/brads3290/cchooks`: Claude Code hooks library for event handling
- `github.com/spf13/cobra`: CLI framework
- Go 1.25.0+ required

## Settings Configuration

Settings are managed in JSON format with two scopes:
- Project: `./.claude/settings.json` 
- Global: `~/.claude/settings.json`

The tool supports plugin-specific enable/disable configuration and preserves unknown JSON fields when reading/writing settings files.