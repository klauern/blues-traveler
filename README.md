# Blues Traveler

[![Go Version](https://img.shields.io/badge/go-1.25.0+-blue.svg)](https://golang.org/doc/go1.25)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/klauern/blues-traveler)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> *"The hook brings you back"* - A [Claude Code hooks](https://docs.anthropic.com/en/docs/claude-code/hooks) management tool

A CLI tool for managing and running [Claude Code](https://claude.ai/code) hooks with built-in security, formatting, debugging, and audit capabilities. Just like the classic Blues Traveler song, our hooks will bring you back to clean, secure, and well-formatted code every time.

## Features

- **🛡️ Security Hook**: Blocks dangerous commands using pattern matching and regex detection
- **🎨 Format Hook**: Automatically formats code files (Go, JavaScript/TypeScript, Python) after editing
- **🐛 Debug Hook**: Comprehensive tool usage logging for troubleshooting
- **📋 Audit Hook**: JSON audit logging for compliance and monitoring
- **✅ Vet Hook**: Code quality and best practices enforcement
- **🔌 Plugin Architecture**: Extensible system with thread-safe plugin registry
- **⚙️ Settings Management**: Supports both project-local and global configuration
- **📊 Log Rotation**: Configurable log rotation with cleanup and compression

## Quick Start

```bash
# Install from source
go install github.com/klauern/blues-traveler@latest

# Or build locally
git clone https://github.com/klauern/blues-traveler.git
cd blues-traveler
task build

# List available hooks
./blues-traveler list

# Install security hook for Claude Code
./blues-traveler install security --event PreToolUse

# Run a specific hook manually
./blues-traveler run security --log

# View installed hooks
./blues-traveler list-installed
```

## Installation

**Prerequisites:** Go 1.25.0 or later

```bash
# From source
go install github.com/klauern/blues-traveler@latest

# Build from source
git clone https://github.com/klauern/blues-traveler.git
cd blues-traveler
task build

# Verify installation
blues-traveler list
```

## Usage

### Core Commands

```bash
# List all available hooks
blues-traveler list

# Run a specific hook
blues-traveler run <hook-name> [--log] [--log-format jsonl|pretty]

# Install hook in Claude Code settings
blues-traveler install <hook-name> [options]

# Remove hook from Claude Code settings
blues-traveler uninstall <hook-name|all> [--global]

# List installed hooks
blues-traveler list-installed [--global]

# List available Claude Code events
blues-traveler list-events

# Generate new hook from template
blues-traveler generate <hook-name> [options]

# Configure log rotation
blues-traveler config-log [options]
```

### Installation Options

```bash
# Install to project settings (./.claude/settings.json)
blues-traveler install security

# Install to global settings (~/.claude/settings.json)
blues-traveler install security --global

# Install for specific event with custom matcher
blues-traveler install format --event PostToolUse --matcher "Edit,Write"

# Enable logging with custom format
blues-traveler install debug --log --log-format pretty

# Set command timeout
blues-traveler install security --timeout 30
```

## Hook Types

### Security Hook (`security`)

Blocks dangerous commands and provides security controls to prevent accidental system damage.

**Features:**

- Pattern matching for dangerous commands (`rm -rf`, `sudo`, etc.)
- Regex-based detection of risky operations
- macOS-specific protections
- Configurable blocking rules

**Best for:** PreToolUse events to intercept commands before execution

```bash
blues-traveler install security --event PreToolUse
```

### Format Hook (`format`)

Automatically formats code files after editing operations to maintain consistent code style.

**Supported Languages:**

- Go (`gofmt`, `goimports`)
- JavaScript/TypeScript (`prettier`, `eslint --fix`)
- Python (`black`, `autopep8`)

**Best for:** PostToolUse events after Edit/Write operations

```bash
blues-traveler install format --event PostToolUse --matcher "Edit,Write"
```

### Debug Hook (`debug`)

Comprehensive logging of all tool usage for debugging and troubleshooting Claude Code operations.

**Features:**

- Detailed event logging to `.claude/hooks/debug.log`
- Raw event capture and parsing
- Configurable log formats (JSON Lines or pretty-printed)
- Log rotation support

**Best for:** All events when debugging issues

```bash
blues-traveler install debug --event PreToolUse --log --log-format pretty
```

### Audit Hook (`audit`)

JSON audit logging for compliance, monitoring, and analysis of Claude Code usage.

**Features:**

- Structured JSON output to stdout
- Pre and post tool use event capture
- Timestamped entries with detailed metadata
- Suitable for log aggregation systems

**Best for:** All events in production environments

```bash
blues-traveler install audit --event PreToolUse
blues-traveler install audit --event PostToolUse
```

### Vet Hook (`vet`)

Code quality and best practices enforcement to catch common issues and maintain code standards.

**Features:**

- Static analysis integration
- Best practices checking
- Custom rule configuration
- Integration with language-specific linters

**Best for:** PostToolUse events after code modifications

```bash
blues-traveler install vet --event PostToolUse --matcher "Edit,Write"
```

## Development

```bash
# Development workflow (format, lint, test, build)
task dev

# Run all checks
task check

# Generate new hook from template
blues-traveler generate my-hook --description "Custom hook for X" --type both
```

## Configuration

### Settings Files

Blues Traveler supports hierarchical configuration:

- **Project**: `./.claude/settings.json` (takes precedence)
- **Global**: `~/.claude/settings.json` (fallback)

### Settings Structure

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "command": "/path/to/blues-traveler run security",
            "timeout": 30
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit,Write",
        "hooks": [
          {
            "command": "/path/to/blues-traveler run format --log"
          }
        ]
      }
    ]
  }
}
```

### Log Rotation Configuration

```bash
# Configure log rotation settings
blues-traveler config-log --max-age 30 --max-size 10 --max-backups 5 --compress

# View current settings
blues-traveler config-log --show

# Configure globally
blues-traveler config-log --global --max-age 7 --max-size 5
```

## Examples

### Complete Setup

```bash
# Security + Format + Debug pipeline
blues-traveler install security --event PreToolUse
blues-traveler install format --event PostToolUse --matcher "Edit,Write"
blues-traveler install debug --event PreToolUse --log --log-format pretty

# Configure log rotation
blues-traveler config-log --max-age 7 --max-size 5 --compress
```

### Production Environment

```bash
# Audit all operations with global configuration
blues-traveler install audit --event PreToolUse --global
blues-traveler install audit --event PostToolUse --global
blues-traveler install security --event PreToolUse --global
```
