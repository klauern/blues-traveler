# Blues Traveler

[![Go Version](https://img.shields.io/badge/go-1.25.0+-blue.svg)](https://golang.org/doc/go1.25)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/klauern/blues-traveler)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> *"The Hook brings you back"* - A [Claude Code hooks](https://docs.anthropic.com/en/docs/claude-code/hooks) management tool

CLI tool for managing and running [Claude Code](https://claude.ai/code) hooks with built-in security, formatting, debugging, and audit capabilities. Powered by `urfave/cli v3` with a static hook registry. Our hooks bring you back to clean, secure, and well-formatted code every time.

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

Note: Custom hooks can implement all of the above (and more) using your own scripts. Built-ins are provided for quick setup; custom hooks are recommended for most workflows.

## 🚀 Quick Start

### Custom Hooks (Recommended)

Define your own security and formatting with custom hooks:

```yaml
# ~/.config/blues-traveler/projects/my-project.yml
my-project:
  PreToolUse:
    jobs:
      - name: security-check
        run: |
          if echo "$TOOL_ARGS" | grep -E "(rm -rf|sudo|curl.*\\|.*sh)"; then
            echo "Dangerous command detected"; exit 1; fi
        only: ${TOOL_NAME} == "Bash"
  PostToolUse:
    jobs:
      - name: format-go
        run: gofmt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
```

Then sync into Claude Code settings:

```bash
blues-traveler hooks custom validate
blues-traveler hooks custom sync
```

You can test a single job locally:

```bash
blues-traveler hooks run config:my-project:format-go
```

### Installation

```bash
# Homebrew (macOS/Linux)
brew tap klauern/blues-traveler
brew install blues-traveler

# Go install
go install github.com/klauern/blues-traveler@latest

# Build from source
git clone https://github.com/klauern/blues-traveler.git
cd blues-traveler
task build

# List available hooks
blues-traveler hooks list

# Install your first hook (security recommended)
blues-traveler hooks install security --event PreToolUse

# Verify installation
blues-traveler hooks list --installed
```

## 📖 Core Commands

### Hook Operations

```bash
# List all available hooks
blues-traveler hooks list

# List installed hooks
blues-traveler hooks list --installed [--global]

# List available Claude Code events
blues-traveler hooks list --events

# Run a specific hook manually
blues-traveler hooks run <hook-name> [--log] [--log-format jsonl|pretty]

# Install hook in Claude Code settings
blues-traveler hooks install <hook-name> [--global] [--event <event>] [--matcher <pattern>] [--timeout <seconds>] [--log] [--log-format <format>]

# Remove hook from Claude Code settings
blues-traveler hooks uninstall <hook-name|all> [--global] [--yes]
```

### Custom Hooks Management

```bash
# Initialize custom hooks configuration
blues-traveler hooks custom init [--group NAME] [--name FILE] [--global] [--overwrite]

# Validate custom hooks configuration
blues-traveler hooks custom validate

# List available custom hook groups
blues-traveler hooks custom list

# Show custom hooks configuration
blues-traveler hooks custom show [--format yaml|json] [--global]

# Sync custom hooks to Claude Code settings
blues-traveler hooks custom sync [group] [--global] [--dry-run] [--event E] [--matcher <pattern>] [--timeout <seconds>]

# Install custom hook group
blues-traveler hooks custom install <group> [--global] [--event E] [--matcher GLOB] [--timeout S] [--list] [--init] [--prune]

# Manage blocked URLs (fetch-blocker)
blues-traveler hooks custom blocked list [--global]
blues-traveler hooks custom blocked add <prefix> [--suggestion TEXT] [--global]
blues-traveler hooks custom blocked remove <prefix> [--global]
blues-traveler hooks custom blocked clear [--global]
```

### Configuration Management

```bash
# Migrate existing configurations to XDG structure
blues-traveler config migrate [--dry-run] [--verbose] [--all]

# List tracked project configurations
blues-traveler config list [--verbose] [--paths-only]

# Edit configuration files
blues-traveler config edit [--global] [--project <path>] [--editor <editor>]

# Clean up orphaned configurations
blues-traveler config clean [--dry-run]

# Show configuration status
blues-traveler config status [--project <path>]

# Configure log rotation settings
blues-traveler config log [--global] [--max-age <days>] [--max-size <MB>] [--max-backups <count>] [--compress] [--show]

# Enable logging with custom format
blues-traveler hooks install debug --log --log-format pretty

# Set custom timeout (seconds)
blues-traveler hooks install security --timeout 30
```

## 🎯 Common Usage Patterns

### Essential Security Setup

Protect against dangerous commands and risky operations:

```bash
# Block dangerous commands
blues-traveler hooks install security --event PreToolUse

# Block unauthorized web fetches
blues-traveler hooks install fetch-blocker --event PreToolUse

# Suggest better alternatives to find
blues-traveler hooks install find-blocker --event PreToolUse
```

### Code Quality Pipeline

Maintain consistent code quality and formatting:

```bash
# Auto-format code after edits
blues-traveler hooks install format --event PostToolUse --matcher "Edit,Write"

# Enforce code quality standards
blues-traveler hooks install vet --event PostToolUse --matcher "Edit,Write"

# Debug and monitor operations
blues-traveler hooks install debug --event PreToolUse --log --log-format pretty
```

### Production Monitoring

Comprehensive audit logging for production environments:

```bash
# Global audit logging for all operations
blues-traveler hooks install audit --event PreToolUse --global
blues-traveler hooks install audit --event PostToolUse --global

# Global security enforcement
blues-traveler hooks install security --event PreToolUse --global
```

### Developer Workflow

Optimal setup for development:

```bash
# Security + formatting + debugging
blues-traveler hooks install security --event PreToolUse
blues-traveler hooks install format --event PostToolUse --matcher "Edit,Write"
blues-traveler hooks install debug --event PreToolUse --log
blues-traveler hooks install find-blocker --event PreToolUse  # Use fd instead
```

### Custom Hooks Sync

Sync custom hooks from your configuration into Claude Code settings:

```bash
# Sync all custom hooks from config to settings
blues-traveler hooks custom sync

# Sync only a specific group
blues-traveler hooks custom sync my-python-group

# Preview changes without applying them
blues-traveler hooks custom sync --dry-run

# Sync to global settings instead of project
blues-traveler hooks custom sync --global

# Sync only hooks for a specific event
blues-traveler hooks custom sync --event PostToolUse
```

**Key Benefits:**

- **Smart Cleanup**: Automatically removes hooks from settings when they're removed from config
- **Group Management**: Sync specific groups or all at once
- **Safe Preview**: Use `--dry-run` to see what changes will be made
- **Event Filtering**: Sync only hooks for specific Claude Code events
- **Stale Detection**: Identifies and cleans up outdated hook entries

The sync command ensures your Claude Code settings stay perfectly aligned with your configuration files, automatically handling additions, updates, and removals.

## ⚙️ Configuration

### Settings Hierarchy

Blues Traveler uses a hierarchical configuration system:

1. **Project Settings**: `./.claude/settings.json` (takes precedence)
2. **Global Settings**: `~/.claude/settings.json` (fallback)

### Blues Traveler Config (embedded)

Project and global Blues Traveler configuration is stored at:

- Project: `~/.config/blues-traveler/projects/<project-name>.json`
- Global: `~/.config/blues-traveler/global.json`

Key sections:

- `logRotation`: Log rotation settings used by `--log` mode.
- `customHooks`: Custom hook groups (by name) with events and jobs.
- `blockedUrls`: URL prefixes used by the `fetch-blocker` hook.

Custom hooks support environment variables and simple expressions to control when jobs run:

- Env vars:
  - `EVENT_NAME`, `TOOL_NAME`, `PROJECT_ROOT`
  - `FILES_CHANGED` (space-separated list)
  - `TOOL_FILE`, `TOOL_OUTPUT_FILE` (for `PostToolUse` on `Edit`/`Write`)
  - `USER_PROMPT` (for `UserPromptSubmit`)
- Expressions in `only`/`skip`:
  - Boolean ops: `&&`, `||`, unary `!`
  - Comparisons: `==`, `!=`
  - Glob: `matches` (right side is a glob). When `FILES_CHANGED` contains multiple tokens, any match passes.
  - Regex: `regex` (right side is a Go regex). Matches any token when multiple files are present.

Examples:

```yaml
mygroup:
  PostToolUse:
    jobs:
      - name: format-py
        run: ruff format --fix ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]

      - name: controller-tests
        run: ./scripts/run-tests.sh
        only: ${FILES_CHANGED} regex ".*controller.*\\.rb$"
```

Example:

```json
{
  "logRotation": { "maxAge": 30, "maxSize": 10, "maxBackups": 5, "compress": true },
  "customHooks": {
    "ruby": {
      "PreToolUse": { "jobs": [{ "name": "rubocop-check", "run": "bundle exec rubocop ${FILES_CHANGED}", "glob": ["*.rb"] }] },
      "PostToolUse": { "jobs": [{ "name": "ruby-test", "run": "bundle exec rspec ${FILES_CHANGED}", "glob": ["*_spec.rb"] }] }
    }
  },
  "blockedUrls": [
    { "prefix": "https://github.com/*/*/private/*", "suggestion": "Use 'gh api' for private repos" },
    { "prefix": "https://api.company.com/private/*" }
  ]
}
```

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
| Logs not appearing | Use `--log` flag and check `~/.config/blues-traveler/` directory |
| Permission denied | Ensure binary has execute permissions: `chmod +x blues-traveler` |
| Config sync issues | Use `--dry-run` to preview changes, check config with `config validate` |
| Stale hook entries | Run `config sync` - it automatically cleans up removed groups |

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
