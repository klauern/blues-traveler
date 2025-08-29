# Blues Traveler

[![Go Version](https://img.shields.io/badge/go-1.25.0+-blue.svg)](https://golang.org/doc/go1.25)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/klauern/blues-traveler)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> *"The Hook brings you back"* - A [Claude Code hooks](https://docs.anthropic.com/en/docs/claude-code/hooks) management tool

A CLI tool for managing and running [Claude Code](https://claude.ai/code) hooks with built-in security, formatting, debugging, and audit capabilities. Just like the classic Blues Traveler song, our hooks will bring you back to clean, secure, and well-formatted code every time.

## ✨ Features

Blues Traveler provides **pre-built hooks** that integrate seamlessly with Claude Code:

| Hook | Description | Best For |
|------|-------------|----------|
| **🛡️ Security** | Blocks dangerous commands (`rm -rf`, `sudo`, etc.) | `PreToolUse` events |
| **🎨 Format** | Auto-formats code after editing (Go, JS/TS, Python) | `PostToolUse` with Edit/Write |
| **🐛 Debug** | Logs all tool usage for troubleshooting | Any event type |
| **📋 Audit** | JSON audit logging for compliance and monitoring | Production environments |
| **✅ Vet** | Code quality and best practices enforcement | `PostToolUse` with code changes |
| **🚫 Fetch Blocker** | Blocks web fetches requiring authentication | `PreToolUse` events |
| **🔍 Find Blocker** | Suggests `fd` instead of `find` for better performance | `PreToolUse` events |

## 🚀 Quick Start

```bash
# Install from source
go install github.com/klauern/blues-traveler@latest

# Or build locally
git clone https://github.com/klauern/blues-traveler.git
cd blues-traveler
task build

# List available hooks
blues-traveler list

# Install your first hook (security recommended)
blues-traveler install security --event PreToolUse

# Verify installation
blues-traveler list-installed
```

## 📖 Core Commands

### Basic Operations

```bash
# List all available hooks
blues-traveler list

# Run a specific hook manually
blues-traveler run <hook-name> [--log] [--log-format jsonl|pretty]

# Install hook in Claude Code settings
blues-traveler install <hook-name> [options]

# Remove hook from Claude Code settings
blues-traveler uninstall <hook-name|all> [--global]

# List installed hooks
blues-traveler list-installed [--global]

# List available Claude Code events
blues-traveler list-events
```

### Installation Options

```bash
# Install to project settings (default)
blues-traveler install <hook-name>

# Install globally for all projects
blues-traveler install <hook-name> --global

# Install with specific event and matcher
blues-traveler install format --event PostToolUse --matcher "Edit,Write"

# Enable logging with custom format
blues-traveler install debug --log --log-format pretty

# Set custom timeout (seconds)
blues-traveler install security --timeout 30
```

## 🎯 Common Usage Patterns

### Essential Security Setup

Protect against dangerous commands and risky operations:

```bash
# Block dangerous commands
blues-traveler install security --event PreToolUse

# Block unauthorized web fetches
blues-traveler install fetch-blocker --event PreToolUse

# Suggest better alternatives to find
blues-traveler install find-blocker --event PreToolUse
```

### Code Quality Pipeline

Maintain consistent code quality and formatting:

```bash
# Auto-format code after edits
blues-traveler install format --event PostToolUse --matcher "Edit,Write"

# Enforce code quality standards
blues-traveler install vet --event PostToolUse --matcher "Edit,Write"

# Debug and monitor operations
blues-traveler install debug --event PreToolUse --log --log-format pretty
```

### Production Monitoring

Comprehensive audit logging for production environments:

```bash
# Global audit logging for all operations
blues-traveler install audit --event PreToolUse --global
blues-traveler install audit --event PostToolUse --global

# Global security enforcement
blues-traveler install security --event PreToolUse --global
```

### Developer Workflow

Optimal setup for development:

```bash
# Security + formatting + debugging
blues-traveler install security --event PreToolUse
blues-traveler install format --event PostToolUse --matcher "Edit,Write"
blues-traveler install debug --event PreToolUse --log
blues-traveler install find-blocker --event PreToolUse  # Use fd instead
```

## ⚙️ Configuration

### Settings Hierarchy

Blues Traveler uses a hierarchical configuration system:

1. **Project Settings**: `./.claude/settings.json` (takes precedence)
2. **Global Settings**: `~/.claude/settings.json` (fallback)

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
  },
  "plugins": {
    "security": { "enabled": true },
    "format": { "enabled": true },
    "debug": { "enabled": false }
  }
}
```

### Disabling Hooks

Hooks can be disabled without removing them from settings:

```json
{
  "plugins": {
    "security": { "enabled": false }
  }
}
```

## 🛠️ Development

### Building from Source

```bash
# Install dependencies
task deps

# Development workflow (format, lint, test, build)
task dev

# Run all checks
task check

# Build binary
task build

# Run tests with coverage
task test-coverage
```

### Adding Custom Hooks

To create a new hook:

1. **Create implementation** in `internal/hooks/myhook.go`
2. **Implement the Hook interface** using `core.BaseHook`
3. **Register** in `internal/hooks/init.go`
4. **Add tests** in `internal/hooks/myhook_test.go`
5. **Document** in README and docs

Example hook structure:

```go
type MyHook struct {
    *core.BaseHook
}

func NewMyHook(ctx *core.HookContext) core.Hook {
    base := core.NewBaseHook("myhook", "MyHook", "Description", ctx)
    return &MyHook{BaseHook: base}
}

func (h *MyHook) Run() error {
    if !h.IsEnabled() {
        return nil
    }
    // Hook logic here
    return nil
}
```

## 🏗️ Architecture

Blues Traveler uses a **static hook registry** architecture:

- ✅ **Static Registration**: Hooks registered at startup via `init()`
- ✅ **Independent Execution**: Each hook runs in isolation
- ✅ **Security First**: No dynamic plugin loading
- ✅ **Configurable**: Enable/disable via settings
- ✅ **Extensible**: Easy to add new hooks

## 📚 Documentation

For detailed documentation, see:

- **[Quick Start Guide](docs/quick_start.md)** - Get up and running in minutes
- **[Developer Guide](docs/developer_guide.md)** - Create custom hooks
- **[Architecture Design](docs/unified_pipeline_design.md)** - Technical deep dive
- **[Documentation Index](docs/index.md)** - All documentation

## 🔧 Troubleshooting

| Issue | Solution |
|-------|----------|
| Hook not found | Run `blues-traveler list` to see available hooks |
| Hook not working | Check if enabled: `blues-traveler list-installed` |
| Settings not applied | Verify path: project `./.claude/settings.json` or global `~/.claude/settings.json` |
| Format not working | Ensure formatters installed: `gofmt`, `prettier`, `black` |
| Logs not appearing | Use `--log` flag and check `.claude/hooks/` directory |
| Permission denied | Ensure binary has execute permissions: `chmod +x blues-traveler` |

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`task test`)
5. Submit a pull request

See [Developer Guide](docs/developer_guide.md) for detailed contribution guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Claude Code](https://claude.ai/code) for the hooks system
- [cchooks](https://github.com/brads3290/cchooks) library for event handling
- [Blues Traveler](https://en.wikipedia.org/wiki/Blues_Traveler) for the inspiration

---

*"It doesn't matter what I say, as long as I sing with inflection"* - Hook by Blues Traveler
