# Custom Hooks Guide

Custom hooks let you define project- or user-specific automation for Claude Code and Cursor events using simple YAML or JSON. This approach can replace most built-in hooks (security, formatting, testing) with scripts tailored to your workflow, while keeping the core model secure and predictable.

## Why Custom Hooks

- **Flexible**: Run any command or script
- **Localized**: Live in your user configuration at `~/.config/blues-traveler/`
- **Powerful**: Conditions (`only`, `skip`), globs, env vars, per-job timeouts
- **Safe by default**: Nothing runs unless installed via `hooks custom sync`

## Current Implementation Status

**Phase 1** implemented (config loader, registry integration, install custom, basic CLI).

**Remaining**: Richer env extraction, expression features, docs polish.

### Implemented Features

- Config file support in `.claude/` and `~/.claude/` with priority
- YAML/JSON parsing (`gopkg.in/yaml.v3`)
- Merge logic: project overrides global; `-local` variants included
- New `config` command group: `init`, `validate`, `groups`, `show`
- `hooks custom install <group>` to add commands pointing to `config:<group>:<job>`
- Runtime registration of config hooks on startup
- `ConfigHook` executes `run` with env, timeout, workdir, skip/only via simple evaluator

## Configuration Files

**Preferred location** (project scope): `~/.config/blues-traveler/projects/<project-name>.yml`

**Also supported**:

- Per-project files in `~/.config/blues-traveler/projects/*.yml`
- Global config in `~/.config/blues-traveler/global.yml`
- Embedded JSON via XDG configuration files under `customHooks`

**Legacy** (still supported but deprecated):

- `.claude/hooks.yml` (project scope)
- `~/.claude/hooks.yml` (user scope)

## YAML Configuration Format

### Basic Structure

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

### Advanced Examples

#### PostToolUse Formatting for Python Files

```yaml
mygroup:
  PostToolUse:
    jobs:
      - name: format-py
        run: ruff format --fix ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
```

#### Regex Filter on Changed Files

```yaml
mygroup:
  PostToolUse:
    jobs:
      - name: regex-sample
        run: echo "changed: ${FILES_CHANGED}"
        only: ${FILES_CHANGED} regex ".*\\.rb$"
```

## Usage Commands

### Validate Configuration

```bash
blues-traveler hooks custom validate
```

### List Available Groups

```bash
blues-traveler config groups
```

### Show Configuration

```bash
blues-traveler config show
```

### Install and Sync

```bash
# Install all groups/events
blues-traveler hooks custom sync

# Install specific group
blues-traveler hooks custom install my-project --event PostToolUse

# Test specific hook
blues-traveler hooks run config:my-project:format-go
```

## Environment Variables

Custom hooks have access to these environment variables:

- `EVENT_NAME`: Current event name (PreToolUse, PostToolUse, etc.)
- `TOOL_NAME`: Tool being used (Bash, Edit, Write, etc.)
- `TOOL_OUTPUT_FILE`: File path for Edit/Write operations
- `FILES_CHANGED`: Comma-separated list of changed files
- `USER_PROMPT`: User's prompt text
- `TOOL_ARGS`: Raw tool arguments where applicable
- `PROJECT_ROOT`: Project root directory
- Plus any custom `env` variables defined in the job

### Special Variables for Edit/Write

For `Edit` and `Write` operations in `PostToolUse` events:

- `FILES_CHANGED`: The target file being edited/written
- `TOOL_FILE`: Same as FILES_CHANGED
- `TOOL_OUTPUT_FILE`: File path for the operation

## Expression Evaluator

The expression evaluator is minimal but supports:

- **Variable substitution**: `${VAR}`
- **Comparison operators**: `==`, `!=`
- **Pattern matching**: `matches` (glob patterns)
- **Regex matching**: `regex` (regex patterns, matches any token when multiple files)
- **Logical operators**: `&&`, `||`
- **Negation**: Unary `!`

### Examples

```yaml
# Simple equality
only: ${TOOL_NAME} == "Edit"

# Multiple conditions with OR
only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"

# Regex matching
only: ${FILES_CHANGED} regex ".*\\.rb$"

# Glob matching
glob: ["*.go", "*.py"]

# Negation with skip
skip: ${FILES_CHANGED} regex ".*_test\\.go$"
```

## Job Configuration Options

Each job supports these options:

- `name`: Job identifier (required)
- `run`: Command to execute (required)
- `only`: Condition expression - job runs only if true
- `skip`: Condition expression - job skipped if true
- `glob`: Array of file patterns to match
- `timeout`: Timeout in seconds (default: no timeout)
- `env`: Additional environment variables (map)
- `workdir`: Working directory (default: project root)

## Replacing Built-in Hooks

Custom hooks can replace most built-in functionality:

### Security Checks

```yaml
security:
  PreToolUse:
    jobs:
      - name: block-dangerous
        run: |
          if echo "$TOOL_ARGS" | grep -E "(rm -rf|sudo|curl.*\\|.*sh)"; then
            echo "Blocked dangerous command"
            exit 1
          fi
        only: ${TOOL_NAME} == "Bash"
```

### Code Formatting

```yaml
formatting:
  PostToolUse:
    jobs:
      - name: format-go
        run: gofumpt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]

      - name: format-python
        run: ruff format --fix ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.py"]
```

### Testing and Validation

```yaml
testing:
  PostToolUse:
    jobs:
      - name: run-tests
        run: go test ./...
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
        skip: ${FILES_CHANGED} regex ".*_test\\.go$"
        timeout: 60

      - name: lint
        run: golangci-lint run ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
```

### Audit and Logging

```yaml
audit:
  PreToolUse:
    jobs:
      - name: log-command
        run: echo "[$(date)] ${TOOL_NAME}: ${TOOL_ARGS}" >> .audit.log
```

## Tips and Best Practices

1. **Use glob patterns** to scope work and improve performance
2. **Add timeouts** to avoid long-running jobs blocking the workflow
3. **Combine only/skip** to be precise about when jobs run
4. **Version control** your configuration files
5. **Test hooks manually** before syncing: `blues-traveler hooks run config:group:job`
6. **Start simple** and add complexity as needed
7. **Document your hooks** with descriptive job names

## Advantages Over Built-ins

- **Project-specific**: Tailor to your exact workflow
- **No code changes**: Pure configuration
- **Quick iteration**: Edit YAML, run `sync`, test
- **Portable**: Configuration files work across machines
- **Composable**: Combine multiple jobs for complex workflows

Built-in hooks remain available for convenience and quick setup, but custom hooks are recommended for most production workflows.

## Next Steps

Planned improvements:

- Expand event coverage and context extraction for additional event types
- Improve `config show` to pretty-print YAML
- Add tests for loader/merge/evaluator and install custom command
- Document advanced patterns and security guidance
- Cursor platform support for custom hooks
- Cross-platform hook sync

## Troubleshooting

### Hook Not Running

1. Validate configuration: `blues-traveler hooks custom validate`
2. Check that hook is synced: `blues-traveler hooks list --installed`
3. Verify conditions: Check `only` and `skip` expressions
4. Test manually: `blues-traveler hooks run config:group:job`

### Expression Errors

- Ensure variable names are correct and available for the event type
- Check syntax of operators (use `==` not `=`)
- Quote regex patterns properly in YAML
- Test expressions incrementally

### Performance Issues

- Add `glob` patterns to reduce unnecessary executions
- Use `timeout` to prevent hanging
- Optimize your scripts for speed
- Consider splitting large jobs into smaller ones

---

**See Also**:

- [Quick Start Guide](quick_start.md) - General setup
- [Developer Guide](developer_guide.md) - Hook internals
- [Cursor Support](cursor-support.md) - Cursor-specific usage
