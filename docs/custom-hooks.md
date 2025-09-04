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

Examples with new variables/operators:
- PostToolUse formatting for Python files edited/written by Claude:
  ```yaml
  mygroup:
    PostToolUse:
      jobs:
        - name: format-py
          run: ruff format --fix ${TOOL_OUTPUT_FILE}
          only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
          glob: ["*.py"]
  ```
- Regex filter on changed files (any token matches):
  ```yaml
  mygroup:
    PostToolUse:
      jobs:
        - name: regex-sample
          run: echo "changed: ${FILES_CHANGED}"
          only: ${FILES_CHANGED} regex ".*\\.rb$"
  ```

Notes:
- Environment variables exposed: `EVENT_NAME`, `TOOL_NAME`, `PROJECT_ROOT`, plus custom `env` from job. For Edit/Write in PostToolUse, `FILES_CHANGED`, `TOOL_FILE`, and `TOOL_OUTPUT_FILE` are populated with the target file.
- Expression evaluator is minimal: supports `${VAR}` substitution, `==`, `!=`, `matches`, `&&`, `||`, unary `!`, and glob patterns on the right side of `matches`.
  - Added `regex`: `${FILES_CHANGED} regex ".*\\.rb$"` (matches any token when multiple files are present).

Next Steps:
- Expand event coverage and context extraction for additional event types.
- Improve `config show` to pretty-print YAML.
- Add tests for loader/merge/evaluator and install custom command.
- Document advanced patterns and security guidance.
