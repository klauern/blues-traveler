# Custom Hooks (Lefthook-style) for blues-traveler

Status: Phase 1 implemented (config loader, registry integration, install custom, basic CLI). Remaining: richer env extraction, expression features, docs polish.

Implemented:
- Config file support in `.claude/` and `~/.claude/` with priority.
- YAML/JSON parsing (`gopkg.in/yaml.v3`).
- Merge logic: project overrides global; `-local` variants included.
- New `config` command group: `init`, `validate`, `groups`, `show`.
- `install custom <group>` to add commands pointing to `config:<group>:<job>`.
- Runtime registration of config hooks on startup.
- `ConfigHook` executes `run` with env, timeout, workdir, skip/only via simple evaluator.

Usage:
- Create `.claude/hooks.yml` with groups and events.
- List groups: `blues-traveler config groups`.
- Validate: `blues-traveler config validate`.
- Install group: `blues-traveler install custom <group> [--event E] [--matcher GLOB] [--timeout S]`.

Notes:
- Environment variables exposed: `EVENT_NAME`, `TOOL_NAME`, `PROJECT_ROOT`, plus custom `env` from job. `FILES_CHANGED` is currently populated best-effort only for PostToolUse; further enrichment may come later.
- Expression evaluator is minimal: supports `${VAR}` substitution, `==`, `!=`, `matches`, `&&`, `||`, unary `!`, and glob patterns on the right side of `matches`.

Next Steps:
- Expand event coverage and context extraction for additional event types.
- Improve `config show` to pretty-print YAML.
- Add tests for loader/merge/evaluator and install custom command.
- Document advanced patterns and security guidance.

