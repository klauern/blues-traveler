# GitHub Copilot Hooks - Environment & Context

**Version:** 1.0
**Last Updated:** 2026-02-21

---

## Input Context Data

All hooks receive JSON via stdin with event-specific fields.

### Common Fields (All Events)

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560000000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |

### Event-Specific Fields

**sessionStart:**
```json
{
  "timestamp": 1708560000000,
  "cwd": "/workspace/project",
  "source": "new",
  "initialPrompt": "User's first prompt"
}
```

**userPromptSubmitted:**
```json
{
  "timestamp": 1708560005000,
  "cwd": "/workspace/project",
  "prompt": "implement authentication"
}
```

**preToolUse / postToolUse:**
```json
{
  "timestamp": 1708560010000,
  "cwd": "/workspace/project",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"ls -la\"}"
}
```

**postToolUse additional:**
```json
{
  "toolResult": {
    "resultType": "success",
    "textResultForLlm": "Output text"
  }
}
```

**errorOccurred:**
```json
{
  "timestamp": 1708560020000,
  "cwd": "/workspace/project",
  "error": {
    "message": "Error description",
    "name": "ErrorType",
    "stack": "Stack trace"
  }
}
```

---

## Environment Variables

### System-Provided Variables

Hooks run with the same environment as the Copilot process:

| Variable | Description | Example |
|----------|-------------|---------|
| `HOME` | User home directory | `/Users/username` |
| `USER` | Current user | `username` |
| `PATH` | System PATH | `/usr/bin:/usr/local/bin` |
| `PWD` | Current directory (= `cwd`) | `/workspace/project` |

### Custom Environment Variables

Set via hook configuration:

```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./hook.sh",
      "env": {
        "LOG_LEVEL": "debug",
        "API_KEY": "${API_KEY}",
        "CUSTOM_VAR": "value"
      }
    }]
  }
}
```

Access in hook:
```bash
#!/bin/bash
echo "Log level: $LOG_LEVEL" >&2
echo "Custom: $CUSTOM_VAR" >&2
```

---

## Accessing Context Data

### Extract JSON Fields

```bash
#!/bin/bash
INPUT=$(cat)

# Extract fields
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
CWD=$(echo "$INPUT" | jq -r '.cwd')
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName // empty')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs // empty')

# Parse nested JSON in toolArgs
if [ -n "$TOOL_ARGS" ]; then
  COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command // empty')
fi
```

### Handle Missing Fields

```bash
#!/bin/bash
INPUT=$(cat)

# Provide default if missing
PROMPT=$(echo "$INPUT" | jq -r '.prompt // ""')
if [ -z "$PROMPT" ]; then
  echo "No prompt in this event" >&2
  exit 0
fi
```

### Validate Input

```bash
#!/bin/bash
INPUT=$(cat)

# Validate JSON structure
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON input" >&2
  exit 2
fi

# Validate required field
if ! echo "$INPUT" | jq -e '.toolName' > /dev/null 2>&1; then
  echo "Missing toolName field" >&2
  exit 1
fi
```

---

## Tool-Specific Context

### bash Tool

```json
{
  "toolName": "bash",
  "toolArgs": "{\"command\":\"git status\"}"
}
```

Extract command:
```bash
COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command')
```

### edit Tool

```json
{
  "toolName": "edit",
  "toolArgs": "{\"path\":\"/path/file.js\",\"content\":\"...\"}"
}
```

Extract path:
```bash
FILE_PATH=$(echo "$TOOL_ARGS" | jq -r '.path')
CONTENT=$(echo "$TOOL_ARGS" | jq -r '.content')
```

### view Tool

```json
{
  "toolName": "view",
  "toolArgs": "{\"path\":\"/path/file.js\"}"
}
```

### create Tool

```json
{
  "toolName": "create",
  "toolArgs": "{\"path\":\"/path/new.js\",\"content\":\"...\"}"
}
```

---

## Complete Example

```bash
#!/bin/bash
# Extract all common context

INPUT=$(cat)

# Common fields
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
CWD=$(echo "$INPUT" | jq -r '.cwd')
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName // empty')

# Log context
{
  echo "Timestamp: $TIMESTAMP"
  echo "CWD: $CWD"
  echo "Tool: $TOOL_NAME"

  # Tool-specific
  if [ "$TOOL_NAME" = "bash" ]; then
    COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command')
    echo "Command: $COMMAND"
  elif [ "$TOOL_NAME" = "edit" ]; then
    PATH=$(echo "$INPUT" | jq -r '.toolArgs.path')
    echo "File: $PATH"
  fi
} >&2

# Make decision
echo '{"permissionDecision":"allow"}' | jq -c
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Input schemas by event
- [Scripting & Execution](./scripting.md) - Using context in scripts
- [API Reference](./api-reference.md) - Complete field reference

---

**Sources:**
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)

**Document Version:** 1.0
**Status:** ✅ Complete
