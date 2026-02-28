# Claude Code Hooks - Environment & Context

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete

---

## Overview

Claude Code hooks receive context through two primary mechanisms:
1. **JSON input via stdin** (standard for all hooks)
2. **Environment variables** (limited, system-provided only)

**Important:** Most context comes through JSON stdin, not environment variables.

---

## System-Provided Environment Variables

### Available in All Hook Types

| Variable | Type | Description | Example | Available When |
|----------|------|-------------|---------|----------------|
| `$CLAUDE_PROJECT_DIR` | string | Project root directory | `/Users/name/project` | All events |
| `$CLAUDE_PLUGIN_ROOT` | string | Plugin's root directory | `/path/to/plugin` | Plugin hooks only |
| `$CLAUDE_CODE_REMOTE` | string | Set to "true" in remote environments | `"true"` | Web/remote sessions |

### SessionStart Only

| Variable | Type | Description | Example | Available When |
|----------|------|-------------|---------|----------------|
| `$CLAUDE_ENV_FILE` | string | File path for persisting env vars | `/tmp/claude-env-abc123` | SessionStart event only |

---

## Common Input Fields (JSON via stdin)

All hooks receive these fields via JSON on stdin:

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/directory",
  "permission_mode": "default|plan|acceptEdits|dontAsk|bypassPermissions",
  "hook_event_name": "EventName"
}
```

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `session_id` | string | Current session identifier | `"abc123"` |
| `transcript_path` | string | Path to conversation JSON | `"/Users/.../.claude/projects/.../transcript.jsonl"` |
| `cwd` | string | Current working directory | `"/Users/name/project"` |
| `permission_mode` | string | Current permission mode | `"default"`, `"plan"`, `"acceptEdits"`, `"dontAsk"`, `"bypassPermissions"` |
| `hook_event_name` | string | Name of event that fired | `"PreToolUse"`, `"PostToolUse"`, etc. |

---

## Event-Specific JSON Fields

### PreToolUse, PermissionRequest

```json
{
  "tool_name": "Bash|Edit|Write|Read|Glob|Grep|Task|WebFetch|WebSearch|mcp__*",
  "tool_input": { /* tool-specific fields */ },
  "tool_use_id": "toolu_01ABC123..."  // PreToolUse only
}
```

### PostToolUse, PostToolUseFailure

```json
{
  "tool_name": "Write",
  "tool_input": { /* same as PreToolUse */ },
  "tool_response": { /* tool-specific response */ },  // PostToolUse only
  "tool_use_id": "toolu_01ABC123...",
  "error": "Error description",  // PostToolUseFailure only
  "is_interrupt": false  // PostToolUseFailure only, optional
}
```

### UserPromptSubmit

```json
{
  "prompt": "User's submitted text"
}
```

### SessionStart

```json
{
  "source": "startup|resume|clear|compact",
  "model": "claude-sonnet-4-6",
  "agent_type": "agent_name"  // Optional
}
```

### Stop, SubagentStop

```json
{
  "stop_hook_active": true,
  "last_assistant_message": "Final response text",
  // SubagentStop adds:
  "agent_id": "def456",
  "agent_type": "Explore",
  "agent_transcript_path": "/path/to/subagent/transcript.jsonl"
}
```

### SubagentStart

```json
{
  "agent_id": "agent-abc123",
  "agent_type": "Explore|Bash|Plan|custom"
}
```

### Notification

```json
{
  "message": "Notification text",
  "title": "Notification title",  // Optional
  "notification_type": "permission_prompt|idle_prompt|auth_success|elicitation_dialog"
}
```

### TeammateIdle

```json
{
  "teammate_name": "researcher",
  "team_name": "my-project"
}
```

### TaskCompleted

```json
{
  "task_id": "task-001",
  "task_subject": "Task title",
  "task_description": "Task details",  // Optional
  "teammate_name": "implementer",  // Optional
  "team_name": "my-project"  // Optional
}
```

### ConfigChange

```json
{
  "source": "user_settings|project_settings|local_settings|policy_settings|skills",
  "file_path": "/path/to/changed/file"
}
```

### WorktreeCreate

```json
{
  "name": "feature-auth"
}
```

### WorktreeRemove

```json
{
  "worktree_path": "/absolute/path/to/worktree"
}
```

### PreCompact

```json
{
  "trigger": "manual|auto",
  "custom_instructions": "User's instructions or empty"
}
```

### SessionEnd

```json
{
  "reason": "clear|logout|prompt_input_exit|bypass_permissions_disabled|other"
}
```

---

## Blues Traveler Custom Hook Variables

**Note:** These are **specific to Blues Traveler's custom hook system**, not official Claude Code variables.

| Variable | Available In | Description | Example |
|----------|--------------|-------------|---------|
| `EVENT_NAME` | All events | Claude Code event name | `"PreToolUse"` |
| `TOOL_NAME` | All events | Tool being used | `"Edit"`, `"Bash"` |
| `PROJECT_ROOT` | All events | Current working directory | `"/path/to/project"` |
| `FILES_CHANGED` | PostToolUse only | Space-separated changed files | `"src/main.go src/utils.go"` |
| `TOOL_FILE` | PostToolUse only | First file from FILES_CHANGED | `"src/main.go"` |
| `TOOL_OUTPUT_FILE` | PostToolUse only | Same as TOOL_FILE (Edit/Write) | `"src/main.go"` |
| `USER_PROMPT` | UserPromptSubmit only | User's prompt text | `"Add error handling"` |

**Usage Context:** Blues Traveler custom hooks (YAML/JSON config), not standard shell command hooks.

---

## Using Environment Variables in Hooks

### Referencing Project Directory

**Use case:** Run scripts relative to project root

```bash
#!/bin/bash
# Reference script in project's .claude/hooks/ directory
SCRIPT="$CLAUDE_PROJECT_DIR/.claude/hooks/validate.sh"
"$SCRIPT"
```

**In settings.json:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/format.sh"
          }
        ]
      }
    ]
  }
}
```

**Important:** Quote the variable to handle paths with spaces.

---

### Plugin Scripts

**Use case:** Run scripts bundled with plugin

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/format.sh"
          }
        ]
      }
    ]
  }
}
```

---

### Remote Environment Detection

**Use case:** Different behavior in web vs. local

```bash
#!/bin/bash
if [[ "$CLAUDE_CODE_REMOTE" == "true" ]]; then
  echo "Running in web/remote environment" >&2
  # Remote-specific logic
else
  echo "Running in local CLI" >&2
  # Local-specific logic
fi
```

---

### Persisting Environment Variables (SessionStart)

**Use case:** Set environment variables for all subsequent Bash commands

```bash
#!/bin/bash
# SessionStart hook

if [ -n "$CLAUDE_ENV_FILE" ]; then
  # Method 1: Individual exports
  echo 'export NODE_ENV=production' >> "$CLAUDE_ENV_FILE"
  echo 'export DEBUG=true' >> "$CLAUDE_ENV_FILE"
  echo 'export PATH="$PATH:./node_modules/.bin"' >> "$CLAUDE_ENV_FILE"
fi

exit 0
```

**Method 2: Capture environment changes**

```bash
#!/bin/bash
# SessionStart hook

ENV_BEFORE=$(export -p | sort)

# Run setup commands that modify environment
source ~/.nvm/nvm.sh
nvm use 20

if [ -n "$CLAUDE_ENV_FILE" ]; then
  ENV_AFTER=$(export -p | sort)
  # Write only the differences
  comm -13 <(echo "$ENV_BEFORE") <(echo "$ENV_AFTER") >> "$CLAUDE_ENV_FILE"
fi

exit 0
```

**What gets persisted:**
- Exported variables only
- Available in all subsequent Bash tool calls
- Not available in hooks (hooks run before Bash commands)

---

## Accessing JSON Fields in Shell Scripts

### Basic Extraction

```bash
#!/bin/bash
INPUT=$(cat)  # Read JSON from stdin

# Extract fields with jq
EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
SESSION=$(echo "$INPUT" | jq -r '.session_id')
```

### Safe Extraction (with defaults)

```bash
#!/bin/bash
INPUT=$(cat)

# Use // empty for missing fields
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Check if field exists
if [[ -z "$FILE" ]]; then
  echo "No file path in input" >&2
  exit 0
fi
```

### Extracting Nested Fields

```bash
#!/bin/bash
INPUT=$(cat)

# Tool-specific input
BASH_COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')
WRITE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
WRITE_CONTENT=$(echo "$INPUT" | jq -r '.tool_input.content // empty')

# Tool response (PostToolUse)
SUCCESS=$(echo "$INPUT" | jq -r '.tool_response.success // false')
RESULT_PATH=$(echo "$INPUT" | jq -r '.tool_response.filePath // empty')
```

### Extracting Arrays

```bash
#!/bin/bash
INPUT=$(cat)

# Permission suggestions (PermissionRequest)
SUGGESTIONS=$(echo "$INPUT" | jq -r '.permission_suggestions[]?.type // empty')

# Array to bash array
readarray -t TYPES < <(echo "$INPUT" | jq -r '.permission_suggestions[]?.type // empty')
for TYPE in "${TYPES[@]}"; do
  echo "Suggestion type: $TYPE" >&2
done
```

---

## Variable Substitution in Output

### In Shell Scripts

```bash
#!/bin/bash
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Output JSON with variable substitution
jq -n --arg file "$FILE" '{
  hookSpecificOutput: {
    hookEventName: "PostToolUse",
    additionalContext: ("Formatted file: " + $file)
  }
}'
```

### In Blues Traveler Custom Hooks

**YAML config:**
```yaml
mygroup:
  PostToolUse:
    jobs:
      - name: format
        run: prettier --write ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
```

**Variable substitution:**
- `${VAR}` - Replaced by Blues Traveler before execution
- Only works in Blues Traveler custom hooks
- Not available in standard shell command hooks

---

## Common Patterns

### Pattern 1: Extract Tool Info

```bash
#!/bin/bash
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')

case "$TOOL" in
  Bash)
    COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')
    echo "Bash command: $COMMAND" >&2
    ;;
  Edit|Write)
    FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')
    echo "File operation: $FILE" >&2
    ;;
esac
```

---

### Pattern 2: Build Context String

```bash
#!/bin/bash
INPUT=$(cat)

# Build context from multiple fields
EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
SESSION=$(echo "$INPUT" | jq -r '.session_id')
CWD=$(echo "$INPUT" | jq -r '.cwd')
MODE=$(echo "$INPUT" | jq -r '.permission_mode')

CONTEXT="Session: $SESSION | Event: $EVENT | Dir: $CWD | Mode: $MODE"
echo "$CONTEXT" >&2
```

---

### Pattern 3: Conditional Logic

```bash
#!/bin/bash
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
MODE=$(echo "$INPUT" | jq -r '.permission_mode')

# Different behavior in bypass mode
if [[ "$MODE" == "bypassPermissions" ]]; then
  echo "Warning: Bypass permissions active" >&2
fi

# Tool-specific checks
if [[ "$TOOL" == "Bash" ]]; then
  COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')
  # Validate command...
fi
```

---

## Known Issues

### Environment Variable Substitution Bug

**GitHub Issues:** anthropics/claude-code#5489, #9567

**Problem:**
- Some users report `CLAUDE_TOOL_NAME` and `CLAUDE_TOOL_PARAMS` not substituting
- Variables appear as literal strings instead of values

**Workaround:**
- Use JSON stdin instead of environment variables
- Extract values with `jq` as shown above

**Status:** Documented issue, use JSON stdin for reliability

---

### Shell Profile Interference

**Problem:**
- Shell profile prints to stdout
- Interferes with JSON parsing
- Hooks fail silently

**Solution:**
```bash
#!/bin/bash
# Redirect debug output to stderr
echo "Debug: Processing hook" >&2

# Keep stdout clean for JSON
INPUT=$(cat)
# ... process ...
```

---

## Variable Availability Matrix

| Variable/Field | SessionStart | UserPromptSubmit | PreToolUse | PermissionRequest | PostToolUse | PostToolUseFailure | Stop | SubagentStop | Other Events |
|----------------|--------------|------------------|------------|-------------------|-------------|-------------------|------|--------------|--------------|
| `$CLAUDE_PROJECT_DIR` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `$CLAUDE_PLUGIN_ROOT` | ✅* | ✅* | ✅* | ✅* | ✅* | ✅* | ✅* | ✅* | ✅* |
| `$CLAUDE_CODE_REMOTE` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `$CLAUDE_ENV_FILE` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `session_id` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `transcript_path` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `cwd` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `permission_mode` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `hook_event_name` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `tool_name` | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `tool_input` | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `tool_use_id` | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `tool_response` | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `error` | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `prompt` | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `last_assistant_message` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |
| `stop_hook_active` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |

\* Plugin hooks only

---

## Best Practices

### 1. Always Quote Variables

```bash
# Bad
rm $FILE

# Good
rm "$FILE"
```

### 2. Check if Fields Exist

```bash
# Bad
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')
rm "$FILE"  # Could be null or "null"!

# Good
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
if [[ -n "$FILE" && -f "$FILE" ]]; then
  rm "$FILE"
fi
```

### 3. Use Strict Error Handling

```bash
#!/bin/bash
set -euo pipefail  # Exit on error, undefined var, pipe failure

INPUT=$(cat)
# Your logic...
```

### 4. Redirect Debug Output to stderr

```bash
#!/bin/bash
# Keep stdout clean for JSON
echo "Debug info" >&2

# Output JSON to stdout
jq -n '{...}'
```

### 5. Validate JSON Input

```bash
#!/bin/bash
INPUT=$(cat)

# Validate JSON
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON input" >&2
  exit 1
fi

# Continue processing...
```

---

## Sources

- **Official Hooks Reference:** https://code.claude.com/docs/en/hooks
- **Blues Traveler Implementation:** [`../../custom-hooks.md`](../../custom-hooks.md)
- **GitHub Issues:** anthropics/claude-code#5489, #9567

Research date: 2026-02-21
