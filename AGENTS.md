# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
# List available hooks
./blues-traveler list

# Run a specific hook
./blues-traveler run <hook-key>

# Install hook into Claude Code settings
./blues-traveler install <hook-key> [flags]

# List installed hooks from settings
./blues-traveler list-installed
```

## File Structure

- `main.go`: CLI entry point with urfave/cli v3 commands
- `internal/hooks/init.go`: Hook registration and built-in hook initialization
- `internal/hooks/*.go`: Individual hook implementations
- `internal/core/registry.go`: Hook registry and management
- `internal/config/settings.go`: Settings file management with JSON marshaling
- `internal/cmd/*.go`: CLI command implementations
- `Taskfile.yml`: Task runner configuration for build/test/lint workflows

## Dependencies

- `github.com/brads3290/cchooks`: Claude Code hooks library for event handling
- `github.com/urfave/cli/v3`: CLI framework
- Go 1.25.0+ required

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

## Migration Notes

The current architecture is simpler and more secure than previous versions:

- **Removed**: Complex pipeline aggregation logic
- **Simplified**: Direct hook execution model
- **Improved**: Better security and reliability
- **Maintained**: All existing CLI commands and functionality

No migration steps are required for users.
