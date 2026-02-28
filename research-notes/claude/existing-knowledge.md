# Existing Knowledge - Blues Traveler's Claude Code Documentation

**Research Date:** 2026-02-21
**Sources:** Internal blues-traveler documentation
**Status:** ✅ Complete

---

## Overview

Blues Traveler is a CLI tool for managing Claude Code hooks. It provides both **built-in hooks** and **custom hook** management capabilities.

**Key Insight:** Blues Traveler bridges the gap between simple built-in hooks and fully custom hook configurations, offering a "lefthook-style" configuration system for Claude Code.

---

## What Blues Traveler Documents

### 1. Built-in Hook Registry

Blues Traveler provides 7 pre-built hooks:

| Hook | Purpose | Event | Description |
|------|---------|-------|-------------|
| **security** | Block dangerous commands | PreToolUse | Blocks `rm -rf`, `sudo`, shell injection |
| **format** | Auto-format code | PostToolUse | Go, JS/TS, Python formatting |
| **debug** | Log tool usage | Any event | JSON/pretty logging for troubleshooting |
| **audit** | Compliance logging | Any event | JSON audit logs for production |
| **vet** | Code quality | PostToolUse | Best practices enforcement |
| **fetch-blocker** | Block auth URLs | PreToolUse | Blocks WebFetch to authenticated URLs |
| **find-blocker** | Suggest fd | PreToolUse | Suggests `fd` instead of `find` |

### 2. Custom Hooks System

**Architecture:** "Lefthook-style" YAML/JSON configuration

**Configuration Locations:**
- **Project:** `./.claude/hooks/hooks.yml` or `./.claude/hooks.yml`
- **Global:** `~/.claude/hooks/hooks.yml` or `~/.claude/hooks.yml`
- **Embedded:** `~/.config/blues-traveler/projects/<name>.json` or `~/.config/blues-traveler/global.json`

**Priority:** Project > Global, Embedded > Separate files

### 3. Custom Hook Schema

```yaml
mygroup:
  PreToolUse:
    jobs:
      - name: security-check
        run: |
          if echo "$TOOL_ARGS" | grep -E "(rm -rf|sudo)"; then
            echo "Dangerous command"; exit 1; fi
        only: ${TOOL_NAME} == "Bash"
  PostToolUse:
    jobs:
      - name: format-go
        run: gofmt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
```

**Job Fields:**
- `name` - Job identifier
- `run` - Shell command to execute
- `only` - Conditional expression (optional)
- `skip` - Inverse conditional (optional)
- `glob` - File pattern matching (optional)
- `env` - Custom environment variables (optional)
- `timeout` - Job timeout in seconds (optional)
- `workdir` - Working directory (optional)

---

## Environment Variables (Blues Traveler Custom Hooks)

Blues Traveler documents these variables for custom hooks:

| Variable | Available In | Description |
|----------|--------------|-------------|
| `EVENT_NAME` | All events | Claude Code event name |
| `TOOL_NAME` | All events | Tool being used |
| `PROJECT_ROOT` | All events | Current working directory |
| `FILES_CHANGED` | PostToolUse only | Space-separated list of changed files |
| `TOOL_FILE` | PostToolUse only | First file from FILES_CHANGED |
| `TOOL_OUTPUT_FILE` | PostToolUse only | Same as TOOL_FILE (Edit/Write) |
| `USER_PROMPT` | UserPromptSubmit only | User's prompt text |

**Important:** These are **Blues Traveler custom hook** variables, not official Claude Code variables.

---

## Expression Evaluator

Blues Traveler implements a minimal expression evaluator for `only`/`skip` conditions:

**Supported Operators:**
- `${VAR}` - Variable substitution
- `==` - Equality
- `!=` - Inequality
- `&&` - Logical AND
- `||` - Logical OR
- `!` - Unary NOT
- `matches` - Glob pattern matching (right side is glob)
- `regex` - Regex matching (right side is regex pattern)

**Examples:**

```yaml
# Equality check
only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"

# Glob matching
only: ${FILES_CHANGED} matches "*.py"

# Regex matching (any token when multiple files)
only: ${FILES_CHANGED} regex ".*\\.rb$"
```

**Note:** When `FILES_CHANGED` has multiple tokens, any match passes the condition.

---

## CLI Commands

### Hook Operations

```bash
# List all available hooks
blues-traveler hooks list

# List installed hooks
blues-traveler hooks list --installed [--global]

# List available events
blues-traveler hooks list --events

# Run hook manually
blues-traveler hooks run <hook-name> [--log] [--log-format jsonl|pretty]

# Install hook
blues-traveler hooks install <hook-name> [--global] [--event <event>] \
  [--matcher <pattern>] [--timeout <seconds>] [--log] [--log-format <format>]

# Uninstall hook
blues-traveler hooks uninstall <hook-name|all> [--global] [--yes]
```

### Custom Hooks

```bash
# Initialize config
blues-traveler hooks custom init [--group NAME] [--name FILE] [--global] [--overwrite]

# Validate config
blues-traveler hooks custom validate

# List groups
blues-traveler hooks custom list

# Show config
blues-traveler hooks custom show [--format yaml|json] [--global]

# Sync to settings
blues-traveler hooks custom sync [group] [--global] [--dry-run] \
  [--event E] [--matcher <pattern>] [--timeout <seconds>]

# Install group
blues-traveler hooks custom install <group> [--global] [--event E] \
  [--matcher GLOB] [--timeout S] [--list] [--init] [--prune]

# Manage blocked URLs (fetch-blocker)
blues-traveler hooks custom blocked list [--global]
blues-traveler hooks custom blocked add <prefix> [--suggestion TEXT] [--global]
blues-traveler hooks custom blocked remove <prefix> [--global]
blues-traveler hooks custom blocked clear [--global]
```

### Configuration Management

```bash
# Migrate to XDG structure
blues-traveler config migrate [--dry-run] [--verbose] [--all]

# List tracked configs
blues-traveler config list [--verbose] [--paths-only]

# Edit config
blues-traveler config edit [--global] [--project <path>] [--editor <editor>]

# Clean orphaned configs
blues-traveler config clean [--dry-run]

# Show status
blues-traveler config status [--project <path>]

# Configure log rotation
blues-traveler config log [--global] [--max-age <days>] [--max-size <MB>] \
  [--max-backups <count>] [--compress] [--show]
```

---

## Blues Traveler Configuration Structure

### Embedded Config Format

```json
{
  "logRotation": {
    "maxAge": 30,
    "maxSize": 10,
    "maxBackups": 5,
    "compress": true
  },
  "customHooks": {
    "group-name": {
      "PreToolUse": {
        "jobs": [
          {
            "name": "job-name",
            "run": "command",
            "only": "${TOOL_NAME} == \"Bash\"",
            "glob": ["*.py"]
          }
        ]
      }
    }
  },
  "blockedUrls": [
    {
      "prefix": "https://github.com/*/*/private/*",
      "suggestion": "Use 'gh api' for private repos"
    }
  ]
}
```

---

## Implementation Details (From blues-traveler source)

### Phase Progress

**Phase 1 - Configuration & Registry (Complete)**
- Config file support in `.claude/` and `~/.claude/`
- YAML/JSON parsing with `gopkg.in/yaml.v3`
- Merge logic: project overrides global
- `-local` variant support
- New `config` command group

**Phase 2 - CLI & Documentation (Complete)**
- Updated help text
- Walkthroughs and examples
- Developer experience improvements

**Phase 3 - JSON Transformation & Runtime Execution (Complete)**
- Full JSON transformation pipeline
- Cursor-style and Claude Code JSON normalization
- Merged with environment context
- Validation before dispatch
- Chained jobs with transformed JSON payloads

### Runtime Execution

**ConfigHook Executor:**
- Runs jobs with environment hydration
- Evaluates `skip`/`only` conditions
- Respects timeouts and glob patterns
- Workdir support

**Environment Builder:**
- Populates `EVENT_NAME`, `TOOL_NAME`, `PROJECT_ROOT`
- For Edit/Write PostToolUse: `FILES_CHANGED`, `TOOL_FILE`, `TOOL_OUTPUT_FILE`
- `FILES_CHANGED` is space-separated (`strings.Join(..., " ")`)
- Custom `env` from job config

---

## Cursor Compatibility

Blues Traveler documents **Cursor IDE compatibility** extensively:

### Event Name Mapping

| Claude Code Event | Cursor Aliases |
|-------------------|---------------|
| PreToolUse | beforeToolUse, beforeShellExecution, beforeFileEdit, beforeFileWrite |
| PostToolUse | afterToolUse, afterShellExecution, afterFileEdit, afterFileWrite |
| UserPromptSubmit | onPromptSubmit, beforePrompt, onUserInput |
| Notification | onNotification, onPermissionRequest |
| Stop | onStop, onAgentStop, afterResponse |
| SubagentStop | onSubagentStop, afterSubagent, onTaskComplete |
| PreCompact | beforeCompact, onCompact |
| SessionStart | onSessionStart, onStart, onSessionBegin |
| SessionEnd | onSessionEnd, onEnd, onSessionClose |

**Usage:** Blues Traveler auto-resolves Cursor aliases to canonical Claude Code names.

```bash
# Both work identically
blues-traveler hooks install security --event PreToolUse
blues-traveler hooks install security --event beforeShellExecution
```

---

## Common Patterns (From blues-traveler docs)

### Essential Security Setup

```bash
blues-traveler hooks install security --event PreToolUse
blues-traveler hooks install fetch-blocker --event PreToolUse
blues-traveler hooks install find-blocker --event PreToolUse
```

### Code Quality Pipeline

```bash
blues-traveler hooks install format --event PostToolUse --matcher "Edit,Write"
blues-traveler hooks install vet --event PostToolUse --matcher "Edit,Write"
blues-traveler hooks install debug --event PreToolUse --log --log-format pretty
```

### Production Monitoring

```bash
# Global audit logging
blues-traveler hooks install audit --event PreToolUse --global
blues-traveler hooks install audit --event PostToolUse --global
blues-traveler hooks install security --event PreToolUse --global
```

### Developer Workflow

```bash
blues-traveler hooks install security --event PreToolUse
blues-traveler hooks install format --event PostToolUse --matcher "Edit,Write"
blues-traveler hooks install debug --event PreToolUse --log
blues-traveler hooks install find-blocker --event PreToolUse
```

---

## Custom Hooks Sync Feature

**Smart Cleanup:** Automatically removes hooks from settings when removed from config

**Benefits:**
- Group management (sync specific or all)
- Safe preview (`--dry-run`)
- Event filtering
- Stale detection and cleanup

```bash
# Sync all
blues-traveler hooks custom sync

# Sync specific group
blues-traveler hooks custom sync my-python-group

# Preview
blues-traveler hooks custom sync --dry-run

# Global
blues-traveler hooks custom sync --global

# Event-specific
blues-traveler hooks custom sync --event PostToolUse
```

---

## Configuration Examples

### Project-Specific Hook (YAML)

```yaml
# ./.claude/hooks/hooks.yml
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
```

### Global Hook (Embedded JSON)

```json
{
  "customHooks": {
    "python-global": {
      "PreToolUse": {
        "jobs": [
          {
            "name": "flake8-check",
            "run": "flake8 .",
            "only": "${TOOL_NAME} == \"Bash\"",
            "glob": ["*.py"]
          }
        ]
      },
      "PostToolUse": {
        "jobs": [
          {
            "name": "black-format",
            "run": "black ${TOOL_OUTPUT_FILE}",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["*.py"]
          },
          {
            "name": "isort-imports",
            "run": "isort ${TOOL_OUTPUT_FILE}",
            "only": "${TOOL_NAME} == \"Edit\" || ${TOOL_NAME} == \"Write\"",
            "glob": ["*.py"]
          }
        ]
      }
    }
  }
}
```

---

## Settings Installation Pattern

Blues Traveler transforms custom hooks into Claude Code settings.json format:

**Custom Hook Config:**
```yaml
mygroup:
  PostToolUse:
    jobs:
      - name: format-py
        run: ruff format --fix ${TOOL_OUTPUT_FILE}
        glob: ["*.py"]
```

**Generated Settings Entry:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "blues-traveler run config:mygroup:format-py"
          }
        ]
      }
    ]
  }
}
```

**Key Insight:** Blues Traveler acts as a runtime for custom hooks, translating group:job references into hook execution with proper environment hydration.

---

## Architecture Insights

### Static Hook Registry

Blues Traveler uses a **static registration** pattern:

- Hooks registered at startup via `init()`
- No dynamic plugin loading (security)
- Independent execution (isolation)
- Enable/disable via settings

### Hook Interface

```go
type Hook interface {
    Name() string
    Description() string
    Run() error
    IsEnabled() bool
}

type BaseHook struct {
    name        string
    displayName string
    description string
    context     *HookContext
}
```

### Adding New Built-in Hooks

1. Create `internal/hooks/myhook.go`
2. Implement Hook interface using `core.BaseHook`
3. Register in `internal/hooks/init.go`
4. Add tests in `internal/hooks/myhook_test.go`
5. Document in README

---

## XDG Configuration Migration

Blues Traveler migrated to XDG Base Directory specification:

**New Locations:**
- Global config: `~/.config/blues-traveler/global.json`
- Project configs: `~/.config/blues-traveler/projects/<name>.json`
- Logs: `~/.local/state/blues-traveler/logs/`

**Migration Command:**
```bash
blues-traveler config migrate
```

---

## Limitations & Known Issues

From blues-traveler documentation:

1. **Environment variables** - Limited to PostToolUse for file-related vars (`FILES_CHANGED`, etc.)
2. **Expression evaluator** - Minimal implementation (no complex expressions)
3. **Event coverage** - Focus on PreToolUse and PostToolUse
4. **Multiple files** - `FILES_CHANGED` handling with space-separated tokens

---

## Future Enhancements (Documented)

From `docs/custom-hooks.md`:

**Optional Improvements:**
- Broaden event coverage
- Improve `config show` to pretty-print YAML
- Deeper examples for advanced patterns
- Security guidance

---

## Sources

Primary internal sources:
- `README.md`
- `docs/custom-hooks.md`
- `docs/cursor-compatibility.md`
- `.claude/settings.json`

Research date: 2026-02-21
