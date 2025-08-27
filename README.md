# klauer-hooks

[![Go Version](https://img.shields.io/badge/go-1.25.0+-blue.svg)](https://golang.org/doc/go1.25)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/klauern/klauer-hooks)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A CLI tool for managing and running [Claude Code](https://claude.ai/code) hooks with built-in security, formatting, debugging, and audit capabilities.

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
go install github.com/klauern/klauer-hooks@latest

# Or build locally
git clone https://github.com/klauern/klauer-hooks.git
cd klauer-hooks
task build

# List available hooks
./hooks list

# Install security hook for Claude Code
./hooks install security --event PreToolUse

# Run a specific hook manually
./hooks run security --log

# View installed hooks
./hooks list-installed
```

## Installation

**Prerequisites:** Go 1.25.0 or later

```bash
# From source
go install github.com/klauern/klauer-hooks@latest

# Build from source
git clone https://github.com/klauern/klauer-hooks.git
cd klauer-hooks
task build

# Verify installation
hooks list
```

## Usage

### Core Commands

```bash
# List all available hooks
hooks list

# Run a specific hook
hooks run <hook-name> [--log] [--log-format jsonl|pretty]

# Install hook in Claude Code settings
hooks install <hook-name> [options]

# Remove hook from Claude Code settings  
hooks uninstall <hook-name|all> [--global]

# List installed hooks
hooks list-installed [--global]

# List available Claude Code events
hooks list-events

# Generate new hook from template
hooks generate <hook-name> [options]

# Configure log rotation
hooks config-log [options]
```

### Installation Options

```bash
# Install to project settings (./.claude/settings.json)
hooks install security

# Install to global settings (~/.claude/settings.json)
hooks install security --global

# Install for specific event with custom matcher
hooks install format --event PostToolUse --matcher "Edit,Write"

# Enable logging with custom format
hooks install debug --log --log-format pretty

# Set command timeout
hooks install security --timeout 30
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
hooks install security --event PreToolUse
```

### Format Hook (`format`)

Automatically formats code files after editing operations to maintain consistent code style.

**Supported Languages:**

- Go (`gofmt`, `goimports`)
- JavaScript/TypeScript (`prettier`, `eslint --fix`)
- Python (`black`, `autopep8`)

**Best for:** PostToolUse events after Edit/Write operations

```bash
hooks install format --event PostToolUse --matcher "Edit,Write"
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
hooks install debug --event PreToolUse --log --log-format pretty
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
hooks install audit --event PreToolUse
hooks install audit --event PostToolUse
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
hooks install vet --event PostToolUse --matcher "Edit,Write"
```

## Development

```bash
# Development workflow (format, lint, test, build)
task dev

# Run all checks
task check

# Generate new hook from template
hooks generate my-hook --description "Custom hook for X" --type both
```

## Configuration

### Settings Files

klauer-hooks supports hierarchical configuration:

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
            "command": "/path/to/hooks run security",
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
            "command": "/path/to/hooks run format --log"
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
hooks config-log --max-age 30 --max-size 10 --max-backups 5 --compress

# View current settings
hooks config-log --show

# Configure globally
hooks config-log --global --max-age 7 --max-size 5
```

## Examples

### Complete Setup

```bash
# Security + Format + Debug pipeline
hooks install security --event PreToolUse
hooks install format --event PostToolUse --matcher "Edit,Write"
hooks install debug --event PreToolUse --log --log-format pretty

# Configure log rotation
hooks config-log --max-age 7 --max-size 5 --compress
```

### Production Environment

```bash
# Audit all operations with global configuration
hooks install audit --event PreToolUse --global
hooks install audit --event PostToolUse --global
hooks install security --event PreToolUse --global
```