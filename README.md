# Blues Traveler

[![Go Version](https://img.shields.io/badge/go-1.25.0+-blue.svg)](https://golang.org/doc/go1.25)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/klauern/blues-traveler)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> _"The Hook brings you back"_ - A [Claude Code hooks](https://docs.anthropic.com/en/docs/claude-code/hooks) management tool

CLI tool for managing and running [Claude Code](https://claude.ai/code) and [Cursor](https://cursor.sh) hooks with built-in security, formatting, debugging, and audit capabilities. Powered by `urfave/cli v3` with a static hook registry. Our hooks bring you back to clean, secure, and well-formatted code every time.

**NEW**: Now supports Cursor IDE! Use `--platform cursor` to install hooks for Cursor. See [docs/cursor-support.md](docs/cursor-support.md) for details.

## ✨ Features

Blues Traveler provides **pre-built hooks** that integrate seamlessly with Claude Code:

| Hook                 | Description                                            | Best For                        |
| -------------------- | ------------------------------------------------------ | ------------------------------- |
| **🛡️ Security**      | Blocks dangerous commands (`rm -rf`, `sudo`, etc.)     | `PreToolUse` events             |
| **🎨 Format**        | Auto-formats code after editing (Go, JS/TS, Python)    | `PostToolUse` with Edit/Write   |
| **🐛 Debug**         | Logs all tool usage for troubleshooting                | Any event type                  |
| **📋 Audit**         | JSON audit logging for compliance and monitoring       | Production environments         |
| **✅ Vet**           | Code quality and best practices enforcement            | `PostToolUse` with code changes |
| **🚫 Fetch Blocker** | Blocks web fetches requiring authentication            | `PreToolUse` events             |
| **🔍 Find Blocker**  | Suggests `fd` instead of `find` for better performance | `PreToolUse` events             |

Note: Custom hooks can implement all of the above (and more) using your own scripts. Built-ins are provided for quick setup; custom hooks are recommended for most workflows.

## 🚀 Quick Start

### Custom Hooks (Recommended)

Define your own security and formatting with custom hooks. You can create **project-specific** or **global** custom hook groups.

#### Project-Specific Custom Hooks

Create hooks that apply only to the current project:

```bash
# Initialize a project custom hook group
blues-traveler hooks custom init --group my-project

# This creates: ./.claude/hooks/hooks.yml
```

```yaml
# ./.claude/hooks/hooks.yml
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

#### Global Custom Hooks

Create hooks that apply to all your projects:

```bash
# Initialize a global custom hook group
blues-traveler hooks custom init --group ruby-global --global

# This creates: ~/.claude/hooks/hooks.yml
```

```yaml
# ~/.claude/hooks/hooks.yml
ruby-global:
  PreToolUse:
    jobs:
      - name: rubocop-check
        run: bundle exec rubocop .
        only: ${TOOL_NAME} == "Bash"
        glob: ["*.rb"]
  PostToolUse:
    jobs:
      - name: ruby-test
        run: bundle exec rake test
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*_test.rb", "*_spec.rb"]
      - name: rubocop-fix
        run: bundle exec rubocop --auto-correct ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.rb"]

python-global:
  PreToolUse:
    jobs:
      - name: flake8-check
        run: flake8 .
        only: ${TOOL_NAME} == "Bash"
        glob: ["*.py"]
  PostToolUse:
    jobs:
      - name: python-test
        run: python -m pytest ${FILES_CHANGED}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["test_*.py", "*_test.py"]
      - name: black-format
        run: black ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
      - name: isort-imports
        run: isort ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
```

#### Installing Custom Hooks

After creating your custom hook groups, install them into Claude Code settings:

```bash
# Install project-specific hooks
blues-traveler hooks custom install my-project

# Install global hooks for specific languages
blues-traveler hooks custom install ruby-global --global
blues-traveler hooks custom install python-global --global

# Or sync all custom hooks at once
blues-traveler hooks custom sync
blues-traveler hooks custom sync --global  # for global hooks
```

#### Testing Custom Hooks

You can test individual jobs locally:

```bash
# Test project hooks
blues-traveler hooks run config:my-project:format-go

# Test global hooks
blues-traveler hooks run config:ruby-global:rubocop-check
blues-traveler hooks run config:python-global:flake8-check
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

Blues Traveler configuration can be stored in two ways:

#### 1. Embedded in Main Config (Recommended)

Custom hooks can be embedded directly in the main Blues Traveler config:

- **Project**: `~/.config/blues-traveler/projects/<project-name>.json`
- **Global**: `~/.config/blues-traveler/global.json`

Key sections:

- `logRotation`: Log rotation settings used by `--log` mode.
- `customHooks`: Custom hook groups (by name) with events and jobs.
- `blockedUrls`: URL prefixes used by the `fetch-blocker` hook.

#### 2. Separate Hook Config Files (Legacy)

Custom hooks can also be defined in separate YAML files:

- **Project**: `./.claude/hooks/hooks.yml` (or `./.claude/hooks.yml`)
- **Global**: `~/.claude/hooks/hooks.yml` (or `~/.claude/hooks.yml`)

**Priority Order**: Project configs override global configs, and embedded configs override separate files.

Custom hooks support environment variables and simple expressions to control when jobs run:

#### Available Environment Variables

| Variable           | Available In          | Description                                 | Example Value                   |
| ------------------ | --------------------- | ------------------------------------------- | ------------------------------- |
| `EVENT_NAME`       | All events            | The Claude Code event name                  | `"PreToolUse"`, `"PostToolUse"` |
| `TOOL_NAME`        | All events            | The tool being used                         | `"Edit"`, `"Write"`, `"Bash"`   |
| `PROJECT_ROOT`     | All events            | Current working directory                   | `"/path/to/project"`            |
| `FILES_CHANGED`    | PostToolUse only      | Space-separated list of changed files       | `"src/main.go src/utils.go"`    |
| `TOOL_FILE`        | PostToolUse only      | First file from FILES_CHANGED (convenience) | `"src/main.go"`                 |
| `TOOL_OUTPUT_FILE` | PostToolUse only      | Same as TOOL_FILE (for Edit/Write)          | `"src/main.go"`                 |
| `USER_PROMPT`      | UserPromptSubmit only | The user's prompt text                      | `"Add error handling"`          |

**Important Notes:**

- `FILES_CHANGED`, `TOOL_FILE`, and `TOOL_OUTPUT_FILE` are **only available in PostToolUse events** when files are actually changed (Edit/Write tools)
- PreToolUse events only have access to `EVENT_NAME`, `TOOL_NAME`, and `PROJECT_ROOT`
- Use `glob` patterns to filter which files trigger the job, and `only`/`skip` conditions to control execution

#### Expression Syntax

Expressions in `only`/`skip` conditions support:

- **Boolean operators**: `&&`, `||`, unary `!`
- **Comparisons**: `==`, `!=`
- **Glob matching**: `matches` (right side is a glob pattern)
- **Regex matching**: `regex` (right side is a Go regex pattern)

When `FILES_CHANGED` contains multiple tokens, any match passes the condition.

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

#### Creating Global Custom Hooks (Embedded Config)

You can create global custom hooks by editing the embedded config directly:

```bash
# Edit global config
blues-traveler config edit --global
```

Add your custom hooks to the `customHooks` section:

```json
{
  "logRotation": {
    "maxAge": 30,
    "maxSize": 10,
    "maxBackups": 5,
    "compress": true
  },
  "customHooks": {
    "ruby-global": {
      "PreToolUse": {
        "jobs": [
          {
            "name": "rubocop-check",
            "run": "bundle exec rubocop .",
            "only": "${TOOL_NAME} == \"Bash\"",
            "glob": ["*.rb"]
          }
        ]
      },
      "PostToolUse": {
        "jobs": [
          {
            "name": "ruby-test",
            "run": "bundle exec rake test",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["*_test.rb", "*_spec.rb"]
          },
          {
            "name": "rubocop-fix",
            "run": "bundle exec rubocop --auto-correct ${TOOL_OUTPUT_FILE}",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["*.rb"]
          }
        ]
      }
    },
    "python-global": {
      "PreToolUse": {
        "jobs": [
          {
            "name": "flake8-check",
            "run": "flake8 .",
            "only": "${TOOL_NAME} == \"Bash\"",
            "glob": ["*.py"]
          }
        ]
      },
      "PostToolUse": {
        "jobs": [
          {
            "name": "python-test",
            "run": "python -m pytest ${FILES_CHANGED}",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["test_*.py", "*_test.py"]
          },
          {
            "name": "black-format",
            "run": "black ${TOOL_OUTPUT_FILE}",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["*.py"]
          },
          {
            "name": "isort-imports",
            "run": "isort ${TOOL_OUTPUT_FILE}",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["*.py"]
          }
        ]
      }
    }
  },
  "blockedUrls": [
    {
      "prefix": "https://github.com/*/*/private/*",
      "suggestion": "Use 'gh api' for private repos"
    },
    { "prefix": "https://api.company.com/private/*" }
  ]
}
```

Then install the global hooks:

```bash
# Install global custom hooks
blues-traveler hooks custom install ruby-global --global
blues-traveler hooks custom install python-global --global
```

#### Project-Specific Example

For project-specific hooks, create a similar structure in your project config:

```json
{
  "customHooks": {
    "ruby": {
      "PreToolUse": {
        "jobs": [
          {
            "name": "rubocop-check",
            "run": "bundle exec rubocop ${FILES_CHANGED}",
            "glob": ["*.rb"]
          }
        ]
      },
      "PostToolUse": {
        "jobs": [
          {
            "name": "ruby-test",
            "run": "bundle exec rspec ${FILES_CHANGED}",
            "glob": ["*_spec.rb"]
          }
        ]
      }
    }
  }
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

| Issue                | Solution                                                                           |
| -------------------- | ---------------------------------------------------------------------------------- |
| Hook not found       | Run `blues-traveler list` to see available hooks                                   |
| Hook not working     | Check if enabled: `blues-traveler list-installed`                                  |
| Settings not applied | Verify path: project `./.claude/settings.json` or global `~/.claude/settings.json` |
| Format not working   | Ensure formatters installed: `gofmt`, `prettier`, `black`                          |
| Logs not appearing   | Use `--log` flag and check `~/.config/blues-traveler/` directory                   |
| Permission denied    | Ensure binary has execute permissions: `chmod +x blues-traveler`                   |
| Config sync issues   | Use `--dry-run` to preview changes, check config with `config validate`            |
| Stale hook entries   | Run `config sync` - it automatically cleans up removed groups                      |

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

_"It doesn't matter what I say, as long as I sing with inflection"_ - Hook by Blues Traveler
