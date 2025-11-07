# Custom Hooks Guide

Custom hooks let you define project- or user-specific automation for Claude Code events using simple YAML or JSON. This approach can replace most built-in hooks (security, formatting, testing, performance monitoring) with scripts tailored to your workflow, while keeping the core model secure and predictable.

## Why Custom Hooks

- Flexible: Run any command or script
- Localized: Live in your user configuration at `~/.config/blues-traveler/`
- Powerful: Conditions (`only`, `skip`), globs, env vars, per-job timeouts
- Safe by default: Nothing runs unless installed via `config sync`

## Configuration Files

Preferred location (project scope): `~/.config/blues-traveler/projects/<project-name>.yml`

Also supported:
- Per-project files in `~/.config/blues-traveler/projects/*.yml`
- Global config in `~/.config/blues-traveler/global.yml`
- Embedded JSON via XDG configuration files under `customHooks`

For backwards compatibility, legacy locations (`.claude/hooks/`) are still supported but deprecated.

## YAML Example

```yaml
my-project:
  PreToolUse:
    jobs:
      - name: security-check
        run: |
          if echo "$TOOL_ARGS" | grep -E "(rm -rf|sudo|curl.*\\|.*sh)"; then
            echo "Dangerous command detected"; exit 1; fi
        only: ${TOOL_NAME} == "Bash"
  PostToolUse:
    jobs:
      - name: format-go
        run: gofmt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
      - name: unit-tests
        run: go test ./...
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        skip: ${FILES_CHANGED} regex ".*_test\\.go$"
        timeout: 60
```

## Install and Test

```bash
blues-traveler hooks custom validate
blues-traveler hooks custom sync            # install all groups/events
blues-traveler hooks run config:my-project:format-go
```

Filter installs to a single group/event:

```bash
blues-traveler hooks custom install my-project --event PostToolUse
```

## Variables Available

- `TOOL_NAME`: Tool (Bash, Edit, Write, etc.)
- `TOOL_OUTPUT_FILE`: File path for Edit/Write
- `FILES_CHANGED`: Comma-separated list of changed files
- `USER_PROMPT`: User’s prompt text
- `EVENT_NAME`: Current event name
- `TOOL_ARGS`: Raw tool arguments where applicable

## Replacing Built-ins

- Security: Implement your policies in a `PreToolUse` script that exits non-zero to block
- Formatting: Run formatters in `PostToolUse` conditioned on `Edit/Write` and file globs
- Testing/Vet: Trigger tests and linters after edits
- Audit/Debug: Emit logs from your scripts or rely on built-in logging flags
- Performance: Monitor tool execution timing with `PreToolUse` and `PostToolUse` handlers

Built-ins remain for convenience; custom hooks are recommended for most workflows.

## Tips

- Use `glob` to scope work and improve performance
- Add `timeout` to avoid long-running jobs
- Combine `only`/`skip` to be precise about when jobs run
- Version control your `.claude/hooks/` directory
