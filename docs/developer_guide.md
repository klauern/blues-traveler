# Developer Guide

This guide explains how to extend Blues Traveler with new hooks, understand the architecture, and contribute to the project.

## Architecture Overview

Blues Traveler uses a **static hook registry** architecture for security and reliability:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Commands  │───▶│  Hook Registry   │───▶│  Hook Impls     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Settings Mgmt  │    │  Event Handling  │    │  Logging &      │
│                 │    │                  │    │  Configuration  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Key Components

- **CLI Layer** (`internal/cmd/`): urfave/cli v3 command implementations
- **Registry** (`internal/core/registry.go`): Static hook registration and management
- **Hooks** (`internal/hooks/`): Concrete hook implementations
- **Settings** (`internal/config/`): Configuration management
- **Core** (`internal/core/`): Event handling and execution

## Adding a New Hook

### 1. Create Hook Implementation

Create a new file in `internal/hooks/` (e.g., `internal/hooks/myhook.go`):

```go
package hooks

import (
    "github.com/klauern/blues-traveler/internal/core"
)

type MyHook struct {
    *core.BaseHook
}

func NewMyHook(ctx *core.HookContext) core.Hook {
    base := core.NewBaseHook("myhook", "MyHook", "Does something useful", ctx)
    return &MyHook{BaseHook: base}
}

func (h *MyHook) Run() error {
    // Your hook logic here
    return nil
}
```

### 2. Register the Hook

Add your hook to `internal/hooks/init.go`:

```go
func init() {
    builtinHooks := map[string]core.HookFactory{
        "security":      NewSecurityHook,
        "format":        NewFormatHook,
        "debug":         NewDebugHook,
        "audit":         NewAuditHook,
        "vet":           NewVetHook,
        "fetch-blocker": NewFetchBlockerHook,
        "find-blocker":  NewFindBlockerHook,
        "myhook":        NewMyHook, // Add your hook here
    }
    core.RegisterBuiltinHooks(builtinHooks)
}
```

### 3. Add Tests

Create tests in `internal/hooks/myhook_test.go`:

```go
package hooks

import (
    "testing"
    "github.com/klauern/blues-traveler/internal/core"
)

func TestMyHook(t *testing.T) {
    ctx := &core.HookContext{}
    hook := NewMyHook(ctx)

    if hook.Name() != "MyHook" {
        t.Errorf("Expected name 'MyHook', got '%s'", hook.Name())
    }

    if hook.Key() != "myhook" {
        t.Errorf("Expected key 'myhook', got '%s'", hook.Key())
    }

    if err := hook.Run(); err != nil {
        t.Errorf("Hook.Run() failed: %v", err)
    }
}
```

### 4. Update Documentation

- Add your hook to the README.md features list
- Document any configuration options
- Provide usage examples

## Hook Interface

All hooks must implement the `core.Hook` interface:

```go
type Hook interface {
    Key() string
    Name() string
    Description() string
    Run() error
    IsEnabled() bool
}
```

### Using BaseHook for Common Functionality

Instead of implementing all methods manually, you can embed `core.BaseHook` to get common functionality:

```go
type MyHook struct {
    *core.BaseHook
}

func NewMyHook(ctx *core.HookContext) core.Hook {
    base := core.NewBaseHook("myhook", "MyHook", "Does something useful", ctx)
    return &MyHook{BaseHook: base}
}

func (h *MyHook) Run() error {
    // Your hook logic here
    return nil
}
```

This gives you:
- `Key()`: Returns the hook identifier
- `Name()`: Returns the human-readable name
- `Description()`: Returns the hook description
- `IsEnabled()`: Checks if the hook is enabled via settings
- `Context()`: Access to the hook context

### Hook Lifecycle

1. **Registration**: Hook is registered at startup via `init()`
2. **Creation**: Hook instance created when needed via factory function
3. **Execution**: Hook's `Run()` method called by CLI or event system
4. **Cleanup**: Hook instance discarded after execution

## Event Handling

Hooks can handle different Claude Code events:

- **PreToolUse**: Before tool execution (e.g., security checks)
- **PostToolUse**: After tool execution (e.g., formatting, validation)
- **UserPromptSubmit**: When user submits a prompt
- **SessionStart/End**: Session lifecycle events

### Event Context

Hooks receive context through the `HookContext`:

```go
type HookContext struct {
    FileSystem      FileSystem
    CommandExecutor CommandExecutor
    RunnerFactory   RunnerFactory
    SettingsChecker func(string) bool
}
```

## Configuration

### Settings Structure

Hooks can be configured via the `plugins` section in settings:

```json
{
  "plugins": {
    "myhook": {
      "enabled": true,
      "customOption": "value"
    }
  }
}
```

### Accessing Settings

Use the `IsEnabled()` method to check if your hook is enabled:

```go
func (h *MyHook) Run() error {
    if !h.IsEnabled() {
        return nil // Hook disabled
    }

    // Your hook logic here
    return nil
}
```

## Logging

### Built-in Logging

Hooks can use the built-in logging via `BaseHook.LogHookEvent`:

```go
import "github.com/klauern/blues-traveler/internal/core"

func (h *MyHook) Run() error {
    // Emits only when enabled via core.SetGlobalLoggingConfig
    h.LogHookEvent("myhook_start", "", map[string]interface{}{"msg": "Starting"}, nil)

    // Your logic here

    h.LogHookEvent("myhook_done", "", map[string]interface{}{"status": "ok"}, nil)
    return nil
}
```

### Log Formats

Logs support two formats:

- **JSON Lines**: Machine-readable, one JSON object per line
- **Pretty**: Human-readable, formatted output

## Testing

### Running Tests

```bash
# Run all tests
task test

# Run specific package tests
go test ./internal/hooks/

# Run with coverage
task test-coverage
```

### Test Patterns

- Test hook registration and discovery
- Test hook execution and error handling
- Test configuration and settings integration
- Test logging and output formats

## Development Workflow

### 1. Find Work

```bash
# See ready-to-work issues
bd ready

# Pick an issue and claim it
bd update blues-traveler-X --status in_progress
```

### 2. Setup Environment

```bash
# Install dependencies
task deps

# Setup development tools
task setup-dev
```

### 3. Make Changes

```bash
# Edit code
# Add tests
# Update documentation
```

### 4. Verify Changes

```bash
# Run all checks
task check

# Build and test
task dev
```

### 5. Complete Work

```bash
# Close the issue
bd close blues-traveler-X "Implemented and tested"

# Commit and push
git add .
git commit -m "feat: add new hook for X (closes blues-traveler-X)"
git push
```

## Best Practices

### Hook Design

- **Single Responsibility**: Each hook should do one thing well
- **Error Handling**: Return meaningful errors, don't panic
- **Configuration**: Use settings for customization, not hardcoded values
- **Logging**: Log important events for debugging and monitoring
- **Testing**: Write comprehensive tests for all functionality

### Code Style

- Follow Go conventions and idioms
- Use descriptive names for functions and variables
- Add comments for complex logic
- Keep functions small and focused
- Handle errors explicitly

### Performance

- Avoid expensive operations in hooks
- Use timeouts for external calls
- Cache expensive computations when possible
- Profile hooks if performance becomes an issue

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Hook not found | Not registered in init.go | Add to builtinHooks map |
| Hook disabled | Settings configuration | Check plugins section |
| Tests failing | Missing dependencies | Run `task deps` |
| Build errors | Go version mismatch | Use Go 1.25+ |

### Debugging

- Use `--log` flag when running hooks
- Check log files in `.claude/hooks/`
- Enable verbose logging with `--log-format pretty`
- Use `blues-traveler hooks list --installed` to verify configuration

## Contributing

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and documentation
5. Run `task check` to verify
6. Submit a pull request

### Code Review

- Ensure all tests pass
- Follow Go coding standards
- Add appropriate documentation
- Consider backward compatibility
- Test with different configurations

## Issue Tracking

This project uses **beads** for issue tracking:

```bash
# Find work to do
bd ready

# See all open issues
bd list --status open

# Get issue details
bd show <issue-id>

# Update issue status
bd update <issue-id> --status in_progress

# Close completed work
bd close <issue-id> "Description of what was done"
```

For detailed workflow documentation, see [Beads Workflow](development/beads-workflow.md).

**Important**: Do NOT create backlog.md or similar files. Use beads for all issue tracking.

## Resources

- [Go Documentation](https://golang.org/doc/)
- [urfave/cli v3](https://github.com/urfave/cli)
- [Claude Code Hooks](https://docs.anthropic.com/en/docs/claude-code/hooks)
- [Beads Workflow](development/beads-workflow.md) - Issue tracking
- [Project Issues (beads)](.beads/) - Local issue database
