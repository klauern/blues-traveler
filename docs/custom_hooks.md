# Custom Hooks Guide

Custom hooks let you define project- or user-specific automation for Claude Code events using simple YAML or JSON. This approach can replace most built-in hooks (security, formatting, testing) with scripts tailored to your workflow, while keeping the core model secure and predictable.

## Why Custom Hooks

- Flexible: Run any command or script
- Localized: Live next to your code in `.claude/hooks/`
- Powerful: Conditions (`only`, `skip`), globs, env vars, per-job timeouts
- Safe by default: Nothing runs unless installed via `config sync`

## Configuration Files

Preferred location (project scope): `./.claude/hooks/hooks.yml`

Also supported:
- Per-group files in `./.claude/hooks/*.yml`
- Global config in `~/.claude/hooks/hooks.yml`
- Embedded JSON via `blues-traveler-config.json` under `customHooks`

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
blues-traveler config validate
blues-traveler config sync            # install all groups/events
blues-traveler run config:my-project:format-go
```

Filter installs to a single group/event:

```bash
blues-traveler install custom my-project --event PostToolUse
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

Built-ins remain for convenience; custom hooks are recommended for most workflows.

## Tips

- Use `glob` to scope work and improve performance
- Add `timeout` to avoid long-running jobs
- Combine `only`/`skip` to be precise about when jobs run
- Version control your `.claude/hooks/` directory

