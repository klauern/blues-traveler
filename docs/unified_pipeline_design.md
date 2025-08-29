# Unified Pipeline (Removed)

The experimental unified multi-hook pipeline (single aggregated `cchooks.Runner` combining multiple hook contributors) has been removed. Each hook now executes independently through its own `Run()` method.

## Rationale

1. Actual usage only invoked a single hook per Claude Code event.
2. Aggregation constructs (contributor type aliases, ordering, filtering wrappers) increased complexity without clear user benefit.
3. Simpler operational model: `blues-traveler run <plugin-key>` instantiates the hook via [`GetPlugin`](plugin.go:21) and calls its [`Run()`](internal/hooks/base.go:20) (internal concrete implementation lines vary per file).

## Current Behavior

- Each hook (security, format, vet, debug, audit) builds a runner with just the handlers it implements (Pre / Post / Raw).
- Logging enablement still toggled globally via [`SetGlobalLoggingConfig`](internal/hooks/registry.go:174).
- Install / uninstall commands continue to write single-hook commands (e.g. `/path/to/blues-traveler run security`) into Claude Code settings.
- No implicit canonical ordering is applied between different hook types.

## Impact to Users

No migration steps required. Existing settings entries continue to function unchanged.

Example (still valid):

```
/absolute/path/blues-traveler run security
```

## Removed / Deprecated Artifacts

| Item | Status |
|------|--------|
| `internal/hooks/pipeline.go` aggregation logic | Replaced with minimal stub |
| Type aliases `PreContributor`, `PostContributor`, `RawConsumer` | Removed from effective API surface |
| Builder funcs `BuildUnifiedRunner`, `BuildRunnerForKeys` | Removed (stub file contains none) |
| Contributor extraction / filtering wrappers | Removed |
| Canonical ordering array `canonicalOrder` | Removed |
| Per-hook exported contributor helper functions (e.g. `SecurityPreContributor`) | Removed |

## Retained Artifacts

| Aspect | Notes |
|--------|-------|
| Individual hook `Run()` methods | Still create isolated runners |
| Base logging via [`LogHookEvent`](internal/hooks/base.go:172) | Unchanged |
| Global logging toggle | Unchanged |
| CLI subcommands (`run`, `install`, `uninstall`, etc.) | Simplified `run` path (no unified pipeline text) |

## Reason Not Fully Deleting `pipeline.go`

A minimal stub is kept (instead of deleting the file) to preserve import stability in case any external code or cached module references existed during transition. Future major version can delete the file entirely once confirmed unused.

## Future Considerations

If coordinated multi-hook sequencing is desired later:

- Prefer an explicit, user-configured pipeline specification (e.g. settings or CLI flag) instead of hard-coded canonical ordering.
- Provide opt-in composition rather than implicit aggregation.
- Consider metrics / timeout instrumentation at that time.

## Changelog Entry (Draft)

Removed: Unified multi-hook pipeline aggregation (`internal/hooks/pipeline.go` logic).
Changed: `blues-traveler run <plugin-key>` now directly invokes the hook's `Run()`; help text updated.
Removed: Contributor helper exports and aggregation type aliases.

(End of deprecation note.)
