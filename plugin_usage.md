# Plugin System Usage

This document explains how to work with the hook plugin architecture for adding linters, formatters, checkers, and other behavioral extensions.

## Overview

Plugins encapsulate hook behavior (security checks, formatting, auditing, etc.).
They register themselves at init() time via a lightweight registry defined in [`plugin.go`](plugin.go).

Key goals:

- Zero boilerplate to add a new plugin
- Stable programmatic listing & execution (`hooks list`, `hooks run <key>`)
- Per‑plugin enable/disable controls via settings (`plugins` section)
- Backwards compatibility with existing install/uninstall flow

## Core Types

Registry + interface (see [`plugin.go`](plugin.go)):

```go
type Plugin interface {
    Name() string
    Description() string
    Run() error
}

func RegisterPlugin(key string, p Plugin) error
func MustRegisterPlugin(key string, p Plugin)
func GetPlugin(key string) (Plugin, bool)
func PluginKeys() []string
```

A helper adapter allows function-style implementations:

```go
p := NewFuncPlugin("My Plugin Name", "Does something", func() error {
    // logic
    return nil
})
MustRegisterPlugin("my-key", p)
```

## Built-In Plugins

Registered in [`plugin.go`](plugin.go):

| Key       | Purpose |
|-----------|---------|
| security  | Blocks dangerous shell commands |
| format    | Formats code after edits/writes |
| debug     | Logs tool usage to a file |
| audit     | Emits structured JSON events |

Underlying logic lives in [`hooks.go`](hooks.go).

## Settings Integration

Settings structure (see [`settings.go`](settings.go)) now includes:

```jsonc
{
  "plugins": {
    "security": { "enabled": true },
    "format":   { "enabled": false }
  }
}
```

Rules:

- Omitted plugin OR `{}` means enabled by default
- `"enabled": false` explicitly disables it
- Future plugin-specific config fields can be added inside each object

Helper method:

```go
func (s *Settings) IsPluginEnabled(key string) bool
```

## Adding a New Plugin

1. Create a new file (e.g. `plugin_<name>.go`)
2. Implement logic (directly or via `NewFuncPlugin`)
3. Register in `init()`:

```go
func init() {
    MustRegisterPlugin("mylinter", NewFuncPlugin(
        "MyLinter",
        "Enforces XYZ style rules",
        runMyLinterHook,
    ))
}
```

4. Provide the implementation:

```go
func runMyLinterHook() error {
    // Instantiate cchooks.Runner or any needed machinery
    // ...
    return nil
}
```

5. (Optional) Add config support by reading from `Settings.Plugins["mylinter"]`.

## Execution Flow

- Listing: `hooks list`
- Direct run (usually by Claude Code): `hooks run <plugin-key>`
- Installation into Claude settings uses existing CLI:

  ```
  hooks install security -e PreToolUse -m "*"
  ```

This writes a command like:

```
<absolute-exe-path> run security
```

into the settings hook entries.

## Disabling a Plugin Without Removing Hook Entries

If you want to keep the hook wiring but temporarily disable behavior:

```json
{
  "plugins": {
    "security": { "enabled": false }
  }
}
```

Your wrapper (caller) should consult `settings.IsPluginEnabled("security")` before invoking (future enhancement—currently direct execution assumes active).

## Extending: Plugin-Specific Configuration

Add fields to `PluginConfig`:

```go
type PluginConfig struct {
    Enabled *bool  `json:"enabled,omitempty"`
    MaxIssues *int `json:"maxIssues,omitempty"`
}
```

Then read inside the plugin:

```go
cfg := settings.Plugins["mylinter"]
if cfg.MaxIssues != nil { /* enforce */ }
```

## Testing Plugins

See forthcoming test file (e.g. `plugin_test.go`) for patterns:

- Verify registration presence
- Ensure duplicate registration errors
- Validate `IsPluginEnabled` semantics

## Migration Notes

Removed:

- Old `HookType` struct & `GetHookTypes()` in favor of registry

Retained:

- CLI verbs (`list`, `run`, `install`, `uninstall`, `list-installed`)
- Hook execution semantics (still shelling out via settings-installed commands)

## Recommended Future Enhancements

- Lazy plugin activation based on event type
- Structured error categorization for plugin failures
- Metrics collector plugin (timings, error counts)
- External plugin discovery via `GOHOOK_PATH` directory scanning + Go plugin `.so` builds (optional advanced step)

## Quick Start Template

```go
// plugin_mylinter.go
package main

func runMyLinterHook() error {
    // TODO: implement logic
    return nil
}

func init() {
    MustRegisterPlugin("mylinter", NewFuncPlugin(
        "My Linter",
        "Checks custom project rules",
        runMyLinterHook,
    ))
}
```

Done—`hooks list` now shows it.

## Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| Plugin missing from list | init() not executed (wrong package name) | Ensure `package main` and file included in build |
| Duplicate key panic | `MustRegisterPlugin` on existing key | Use unique key or switch to `RegisterPlugin` and handle error |
| Settings not honoring disable | Caller not checking `IsPluginEnabled` | Add guard before invoking plugin logic |

## Summary

The new architecture decouples:

- Registration (static init)
- Discovery (registry enumeration)
- Execution (uniform `Run()` method)
- Configuration (`settings.Plugins`)

This minimizes friction for adding new capability classes (linters, formatters, audit layers) while keeping backward compatibility with existing hook installation semantics.
