# Architecture Design

This document explains the current architecture of Blues Traveler and how hooks are executed.

## Current Architecture

Blues Traveler uses a **static hook registry** where each hook runs independently:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Commands  │───▶│  Hook Registry   │───▶│  Hook Impls     │
│   (Cobra)      │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Settings Mgmt  │    │  Event Handling  │    │  Logging &      │
│                 │    │                  │    │  Configuration  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Key Design Decisions

1. **Static Registration**: All hooks are registered at startup via `init()` functions
2. **Independent Execution**: Each hook runs in isolation for security and reliability
3. **No Dynamic Loading**: Prevents security risks and ensures predictable behavior
4. **Simple Lifecycle**: Create → Execute → Cleanup

## Hook Execution Flow

### 1. Command Execution

```bash
blues-traveler run security
```

### 2. Hook Lookup

The command looks up the hook in the registry:

```go
hook, exists := registry.Create("security")
if !exists {
    return fmt.Errorf("hook 'security' not found")
}
```

### 3. Hook Execution

The hook's `Run()` method is called:

```go
if err := hook.Run(); err != nil {
    return fmt.Errorf("hook execution failed: %v", err)
}
```

### 4. Cleanup

The hook instance is discarded after execution.

## Benefits of Current Design

### Security

- No dynamic code loading
- Predictable execution environment
- Controlled hook lifecycle

### Reliability

- No runtime registration errors
- Consistent behavior across runs
- Easy to test and debug

### Simplicity

- Clear execution model
- Straightforward debugging
- Minimal complexity

## Event Handling

Each hook can implement handlers for different Claude Code events:

- **PreToolUse**: Before tool execution
- **PostToolUse**: After tool execution
- **UserPromptSubmit**: When user submits prompt
- **SessionStart/End**: Session lifecycle

### Example Hook Structure

```go
type SecurityHook struct {
    // Hook implementation
}

func (h *SecurityHook) Run() error {
    // Create runner for this hook
    runner := cchooks.NewRunner()

    // Add event handlers
    runner.OnPreToolUse(h.handlePreToolUse)
    runner.OnPostToolUse(h.handlePostToolUse)

    // Execute
    return runner.Run()
}
```

## Configuration

Hooks are configured via settings files:

```json
{
  "plugins": {
    "security": { "enabled": true },
    "format": { "enabled": false }
  }
}
```

### Settings Hierarchy

1. **Project**: `./.claude/settings.json` (takes precedence)
2. **Global**: `~/.claude/settings.json` (fallback)

## Logging

Hooks can use the built-in logging system:

```go
// Enable logging
core.SetGlobalLoggingConfig(true, ".claude/hooks", "jsonl")

// Log events
core.LogHookEvent("security", "Command blocked", data)
```

### Log Formats

- **JSON Lines**: Machine-readable, one JSON object per line
- **Pretty**: Human-readable, formatted output

## Future Considerations

If coordinated multi-hook sequencing is needed later:

- Prefer explicit, user-configured pipeline specification
- Provide opt-in composition rather than implicit aggregation
- Consider metrics and timeout instrumentation
- Maintain the current security and reliability guarantees

## Migration from Previous Versions

The current architecture is simpler and more secure than previous versions:

- **Removed**: Complex pipeline aggregation logic
- **Simplified**: Direct hook execution model
- **Improved**: Better security and reliability
- **Maintained**: All existing CLI commands and functionality

No migration steps are required for users.
