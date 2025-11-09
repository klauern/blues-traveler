# AGENTS.md

This file provides guidance to AI assistants (like Claude Code) when working with code in this repository.

## Project Overview

This is `blues-traveler`, a CLI tool for managing and running Claude Code hooks. It provides a **static hook registry** with built-in security, formatting, debugging, and audit capabilities.

## Current Architecture

### Core Design Principles

- **Static Registration**: All hooks are registered at startup via `init()` functions
- **Independent Execution**: Each hook runs in isolation for security and reliability
- **No Dynamic Loading**: Prevents security risks and ensures predictable behavior
- **Simple Lifecycle**: Create → Execute → Cleanup

### Key Components

- **CLI Layer** (`internal/cmd/`): urfave/cli v3 command implementations
- **Registry** (`internal/core/registry.go`): Static hook registration and management
- **Hooks** (`internal/hooks/`): Concrete hook implementations
- **Settings** (`internal/config/`): Configuration management
- **Custom Hooks** (`internal/config/hooks_config.go`, `internal/cmd/hooks_config.go`): YAML/JSON-driven hooks synced into Claude Code
- **Core** (`internal/core/`): Event handling and execution

### Hook System

The application uses a **static hook registry** where:

- Built-in hooks register themselves via `init()` functions in `internal/hooks/init.go`
- Each hook implements the `core.Hook` interface with `Key()`, `Name()`, `Description()`, `Run()`, and `IsEnabled()` methods
- Hooks can be enabled/disabled via settings configuration
- Plugin state is checked hierarchically (project settings override global settings)

### Built-in Hooks

| Key | Purpose | Best Event |
|-----|---------|------------|
| `security` | Blocks dangerous commands using pattern matching and regex detection | `PreToolUse` |
| `format` | Auto-formats code files after Edit/Write operations (Go, JS/TS, Python) | `PostToolUse` |
| `debug` | Logs all tool usage to `blues-traveler.log` | Any event |
| `audit` | Comprehensive JSON audit logging to stdout | Any event |
| `vet` | Code quality and best practices enforcement | `PostToolUse` |
| `fetch-blocker` | Blocks fetch requests for security | `PreToolUse` |
| `find-blocker` | Blocks find commands for security | `PreToolUse` |

Note: Custom hooks can implement similar behavior using your own scripts. Prefer custom hooks for project-specific security, formatting, testing, and workflows; use built-ins for quick starts.

## Development Commands

**IMPORTANT**: Use Taskfile tasks for tests, building, formatting, etc.

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
# List available hooks
./blues-traveler hooks list

# List installed hooks from settings
./blues-traveler hooks list --installed

# List available events
./blues-traveler hooks list --events

# Run a specific hook
./blues-traveler hooks run <hook-key>

# Install hook into Claude Code settings
./blues-traveler hooks install <hook-key> [flags]

# Install custom hook group
./blues-traveler hooks custom install <group-name> [flags]

# Sync custom hooks to settings
./blues-traveler hooks custom sync [group] [flags]
```

## File Structure

- `main.go`: CLI entry point with urfave/cli v3 commands
- `internal/hooks/init.go`: Hook registration and built-in hook initialization
- `internal/hooks/*.go`: Individual hook implementations
- `internal/core/registry.go`: Hook registry and management
- `internal/config/settings.go`: Settings file management with JSON marshaling
- `internal/cmd/hooks.go`: Consolidated hooks command with all hook operations
- `internal/cmd/config_xdg.go`: Configuration management commands
- `internal/cmd/generate.go`: Code generation commands
- `internal/cmd/version.go`: Version information command
- `Taskfile.yml`: Task runner configuration for build/test/lint workflows

## Dependencies

- `github.com/brads3290/cchooks`: Claude Code hooks library for event handling
- `github.com/urfave/cli/v3`: CLI framework
- Go 1.25.4+ required

## Settings Configuration

Settings are managed in JSON format with two scopes:

- **Project**: `./.claude/settings.json` (takes precedence)
- **Global**: `~/.claude/settings.json` (fallback)

The tool supports hook-specific enable/disable configuration and preserves unknown JSON fields when reading/writing settings files.

## Adding New Hooks

To add a new hook:

1. **Create Implementation**: Add new file in `internal/hooks/` implementing `core.Hook` interface
2. **Register Hook**: Add to `builtinHooks` map in `internal/hooks/init.go`
3. **Add Tests**: Create test file following existing patterns
4. **Update Docs**: Add to README.md and relevant documentation

### Example Hook Structure

```go
type MyHook struct{
    *core.BaseHook
}

func NewMyHook(ctx *core.HookContext) core.Hook {
    base := core.NewBaseHook("myhook", "MyHook", "Does something useful", ctx)
    return &MyHook{BaseHook: base}
}

func (h *MyHook) Run() error {
    // Hook logic here
    return nil
}
```

## Important Notes for AI Assistants

### What NOT to Do

- Don't suggest dynamic plugin loading or runtime registration
- Don't reference the old pipeline system (it's been removed)
- Don't suggest modifying the registry at runtime
- Don't reference Cobra (the project now uses urfave/cli v3)

### What TO Do

- Suggest adding hooks to the static registry in `init.go`
- Recommend implementing the `core.Hook` interface (use `core.BaseHook` for common functionality)
- Suggest using the existing settings system for configuration
- Point to existing hook implementations as examples
- Reference urfave/cli v3 for CLI-related functionality

### Common Patterns

- Hooks are stateless and created fresh for each execution
- Use `h.LogHookEvent()` (from `core.BaseHook`) for logging within hooks
- Check `h.IsEnabled()` for configuration-based behavior
- Implement event handlers in the `Run()` method
- Use `core.BaseHook` to get common functionality like `Key()`, `Name()`, `Description()`, and `IsEnabled()`

## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**

```bash
bd ready --json
```

**Create new issues:**

```bash
bd create "Issue title" -t bug|feature|task -p 0-4 --json
bd create "Issue title" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**

```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**

```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs with git:

- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### MCP Server (Recommended)

If using Claude or MCP-compatible clients, install the beads MCP server:

```bash
pip install beads-mcp
```

Add to MCP config (e.g., `~/.config/claude/config.json`):

```json
{
  "beads": {
    "command": "beads-mcp",
    "args": []
  }
}
```

Then use `mcp__beads__*` functions instead of CLI commands.

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and QUICKSTART.md.

## Migration Notes

The current architecture is simpler and more secure than previous versions:

- **Removed**: Complex pipeline aggregation logic
- **Simplified**: Direct hook execution model
- **Improved**: Better security and reliability
- **Maintained**: All existing CLI commands and functionality

No migration steps are required for users.
