# Custom Hooks (Lefthook-style) for blues-traveler

[![Phase 2 Complete – 2025-09-30](https://img.shields.io/badge/Phase%202-Complete%20%E2%80%94%202025--09--30-1f7a1f.svg)](#phase-2-cli--documentation)
[![Phase 3 Complete – 2025-09-30](https://img.shields.io/badge/Phase%203-Complete%20%E2%80%94%202025--09--30-1f7a1f.svg)](#phase-3-json-transformation--runtime-execution)

> **Status:** Phase 3 complete — JSON input is transformed into runtime-ready jobs and executed through the config hook pipeline.
>
> **Last Updated:** 2025-09-30

The custom hooks initiative now covers the end-to-end workflow: configuration loading, CLI tooling, documentation, JSON transformation, and runtime execution. The sections below recap what shipped in each phase and how to take advantage of the finished pipeline.

## Phase Progress

### Phase 1 – Configuration & Registry (Complete)
Phase 1 established the foundation for custom hooks: configuration files load from project and global scopes, YAML/JSON parsing feeds the registry merge logic, and the CLI exposes `init`, `validate`, `groups`, and `show` entry points. The config hook executor runs jobs with environment hydration, skip/only evaluation, and timeouts.

### Phase 2 – CLI & Documentation (Complete)
Phase 2 focused on rounding out the developer experience and education materials. Updated help text, walkthroughs, and examples ship with the CLI so teams can confidently adopt custom hooks without reading source code.

### Phase 3 – JSON Transformation & Runtime Execution (Complete)
Phase 3 delivered the full JSON transformation pipeline and hook execution flow. Cursor-style and Claude Code JSON definitions are normalized into a shared internal model, merged with environment context, and executed through the config hook runner. Validation ensures each job’s matcher, timeout, and command payloads are respected before dispatch. This phase unlocks seamless execution of custom hook jobs installed via the CLI, including chained jobs that rely on transformed JSON payloads.

Implemented:
- Config file support in `.claude/` and `~/.claude/` with priority.
- YAML/JSON parsing (`gopkg.in/yaml.v3`).
- Merge logic: project overrides global; `-local` variants included.
- New `config` command group: `init`, `validate`, `groups`, `show`.
- `install custom <group>` to add commands pointing to `config:<group>:<job>`.
- Runtime registration of config hooks on startup.
- `ConfigHook` executes `run` with env, timeout, workdir, skip/only via simple evaluator.
- Phase 2 documentation updates for CLI commands and usage examples.
- Phase 3 runtime execution tests confirming JSON-to-job transformation and hook dispatch.

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
- Environment variables exposed: `EVENT_NAME`, `TOOL_NAME`, `PROJECT_ROOT`, plus custom
  `env` from job. For Edit/Write in PostToolUse, `FILES_CHANGED`, `TOOL_FILE`, and
  `TOOL_OUTPUT_FILE` are populated with the target file. `FILES_CHANGED` is
  space-separated (matching the `strings.Join(..., " ")` behavior in the environment
  builder).
- Expression evaluator is minimal: supports `${VAR}` substitution, `==`, `!=`, `matches`, `&&`, `||`, unary `!`, and glob patterns on the right side of `matches`.
  - Added `regex`: `${FILES_CHANGED} regex ".*\\.rb$"` (matches any token when multiple files are present).

Future Enhancements (Optional):
- Broaden event coverage and context extraction for additional event types.
- Improve `config show` to pretty-print YAML.
- Add deeper examples for advanced patterns and security guidance.
