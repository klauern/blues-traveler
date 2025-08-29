# Plugin System Usage

This document explains how to work with the hook plugin (hook) architecture for adding linters, formatters, checkers, and other behavioral extensions.

## Overview

Hooks encapsulate behavior (security checking, formatting, auditing, debugging).
They are defined as concrete implementations under `internal/hooks` and referenced by key (e.g. `security`, `format`).
Dynamic runtime registration has been deprecated. The previous helper `NewFuncPlugin` and `RegisterPlugin` pattern were removed in favor of a static, explicit set of hook implementations.

Key goals:

- Simple discovery & execution (`blues-traveler list`, `blues-traveler run <key>`)
- Per‑hook enable/disable controls via settings (`plugins` section)
- Deterministic set of built‑ins with clear maintenance surface
- Backwards compatibility layer (shim in [`plugin.go`](plugin.go)) for existing consumers calling `GetPlugin` / `PluginKeys`

## Core Types (Shim Layer)

For backward compatibility the legacy types are still exposed (see [`plugin.go`](plugin.go)):

```go
type Plugin interface {
    Name() string
    Description() string
    Run() error
}

func GetPlugin(key string) (Plugin, bool)
func PluginKeys() []string
```

`RegisterPlugin`, `MustRegisterPlugin` now return (or panic with) an error informing that dynamic registration is deprecated.

## Built-In Hooks

| Key       | Purpose |
|-----------|---------|
| security  | Blocks dangerous shell commands |
| format    | Formats code after edits/writes |
| debug     | Logs tool usage to a file |
| audit     | Emits structured JSON events |

Implementations live under `internal/hooks/*.go`.

## Settings Integration

Settings structure (see [`settings.go`](settings.go)):

```jsonc
{
  "plugins": {
    "security": { "enabled": true },
    "format":   { "enabled": false }
  }
}
```

Rules:

- Omitted plugin OR empty object means enabled by default
- `"enabled": false` explicitly disables it
- Future plugin‑specific config fields may be added per key

Helper:

```go
func (s *Settings) IsPluginEnabled(key string) bool
```

## Adding a New Built-In Hook

Because dynamic registration is removed, adding a new hook requires:

1. Create implementation file under `internal/hooks/` implementing the `hooks.Hook` interface.
2. Add its factory/constructor to the registry list (refer to existing files such as `security.go`).
3. Update documentation if needed.
4. Add tests asserting its presence in `PluginKeys()` (via shim) and enabled behavior.

Minimal pattern:

```go
// internal/hooks/myhook.go
package hooks

type myHook struct{}

func (h *myHook) Name() string        { return "MyHook" }
func (h *myHook) Description() string { return "Does something useful" }
func (h *myHook) Run() error {
    // logic
    return nil
}

// In registry (implementation dependent), ensure key "myhook" creates *myHook.
```

## Execution Flow

- List: `blues-traveler list`
- Run:  `blues-traveler run <key>`
- Install into Claude settings:

  ```
  blues-traveler install security -e PreToolUse -m "*"
  ```

Stored command resembles:

```
<absolute-exe-path> run security
```

## Disabling a Hook Without Removing Settings Entries

```json
{
  "plugins": {
    "security": { "enabled": false }
  }
}
```

The CLI performs the `IsPluginEnabled` check prior to execution.

## Extending: Hook-Specific Configuration

Extend `PluginConfig`:

```go
type PluginConfig struct {
    Enabled  *bool `json:"enabled,omitempty"`
    MaxIssues *int `json:"maxIssues,omitempty"`
}
```

Use inside the hook by fetching settings from the global context (see existing hooks for patterns).

## Testing

Typical assertions (see `plugin_test.go`):

- Built‑in keys present and sorted
- `IsPluginEnabled` default / explicit enable / disable logic

Dynamic registration tests were removed with deprecation of runtime registration.

## Migration Notes

Removed:

- Dynamic runtime registration (`NewFuncPlugin`, functional adapters)
- Pipeline aggregation constructs

Retained:

- CLI verbs (`list`, `run`, `install`, `uninstall`, `list-installed`)
- Uniform `Run()` execution model
- Settings enable/disable semantics

## Future Enhancements (Potential)

- Lazy activation based on event type
- Structured error categorization
- Metrics (timings, error counts)
- Optional external discovery (would require a new, explicit extension mechanism)

## Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| Key missing from list | Hook not added to registry | Add constructor entry in registry |
| Disabled hook still runs | Settings not reloaded / wrong scope | Verify project vs global settings and enable flag |
| Formatting not applied | File types unsupported | Extend formatter logic / add hook |

## Logging Formats

Structured hook event logs written to per-hook log files support two formats controlled by the `--log-format` flag (default: `jsonl`) when combined with `--log`:

- `jsonl` (default): Each event is a single compact JSON object on one line (JSON Lines). Easier to stream and parse (e.g. `jq -c`, `grep`).
- `pretty`: Multi-line, indented JSON for human inspection.

Example:

```
blues-traveler run audit --log                      # jsonl (default)
blues-traveler run audit --log --log-format pretty  # pretty printed
```

The flag is also persisted when installing a hook if you specify `--log` and a non-default format:

```
blues-traveler install audit --event PreToolUse --log --log-format pretty
```

(If you omit `--log-format` or use `jsonl`, nothing extra is stored.)

Scope / limitations:

- Affects only the structured hook event file logging (shared logger in [`internal/hooks/logging.go`](internal/hooks/logging.go:1)).
- Audit hook stdout output is already compact single-line JSON (unchanged).
- Debug hook continues emitting human-readable lines; `--log-format` does not alter its legacy text output (future unification possible).

## Summary

The revised architecture favors explicit, static hook definitions for predictability and reduced surface area. A thin shim maintains backward compatibility for existing external tooling expecting the prior registry API while discouraging dynamic mutation. The logging format flag introduces observability flexibility without breaking previous behavior (default now optimized for machine parsing).
