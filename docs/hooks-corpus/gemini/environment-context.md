# Gemini - Environment & Context

**Last Updated:** 2026-02-21
**Version:** Gemini CLI v0.26.0+

---

## Overview

Gemini hooks receive rich context data via JSON on stdin, allowing hooks to make informed decisions. This document catalogs all available context fields for each hook event.

**Sources:**
- [Gemini CLI Hooks Reference](https://geminicli.com/docs/hooks/reference/)
- [Writing Hooks Documentation](https://geminicli.com/docs/hooks/writing-hooks/)

---

## Environment Variables

### System-Provided Variables

Gemini CLI injects these environment variables into hook execution contexts:

| Variable | Type | Availability | Description | Example |
|----------|------|--------------|-------------|---------|
| `GEMINI_PROJECT_DIR` | string | All hooks | Project root directory (absolute path) | `/Users/user/projects/myapp` |
| `GEMINI_USER_DIR` | string | All hooks | User's home directory | `/Users/user` |
| `GEMINI_SESSION_ID` | string | All hooks | Unique session identifier | `abc123def456...` |
| `GEMINI_EVENT_TYPE` | string | All hooks | Current hook event type | `BeforeTool` |
| `GEMINI_CLI_VERSION` | string | All hooks | CLI version | `0.26.0` |

**Usage in Hooks:**
```bash
#!/bin/bash
# Access environment variables directly

PROJECT_DIR="$GEMINI_PROJECT_DIR"
SESSION_ID="$GEMINI_SESSION_ID"

echo "Running in project: $PROJECT_DIR"
echo "Session: $SESSION_ID"
```

### Custom Environment Variables

Set custom variables in hook configuration:

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["./hooks/notify.sh"],
            "env": {
              "SLACK_WEBHOOK": "https://hooks.slack.com/...",
              "NOTIFICATION_CHANNEL": "#dev-alerts",
              "LOG_LEVEL": "info"
            }
          }
        ]
      }
    ]
  }
}
```

**Access in Hook:**
```bash
#!/bin/bash
WEBHOOK="$SLACK_WEBHOOK"
CHANNEL="$NOTIFICATION_CHANNEL"

# Use custom variables
curl -X POST "$WEBHOOK" -d "{\"channel\": \"$CHANNEL\", \"text\": \"Alert!\"}"
```

---

## Context Data Structures

### Common Fields (All Events)

All hook events receive these base fields:

```json
{
  "event": "EventType",
  "timestamp": "2026-02-21T12:00:00Z",
  "sessionId": "unique-session-identifier",
  "requestId": "unique-request-identifier"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `event` | string | Hook event type (e.g., `"BeforeTool"`) |
| `timestamp` | string | ISO 8601 timestamp |
| `sessionId` | string | Unique session identifier |
| `requestId` | string | Unique request identifier (for agent turns) |

---

## Event-Specific Context

### SessionStart

**Full Context:**
```json
{
  "event": "SessionStart",
  "timestamp": "2026-02-21T12:00:00Z",
  "sessionId": "abc123...",
  "projectPath": "/absolute/path/to/project",
  "user": {
    "id": "user-identifier",
    "email": "user@example.com",
    "name": "User Name"
  },
  "cliVersion": "0.26.0",
  "isResume": false
}
```

**Field Descriptions:**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `projectPath` | string | Absolute path to project root | `/Users/user/myproject` |
| `user.id` | string | User identifier | `user-123` |
| `user.email` | string | User email | `user@example.com` |
| `user.name` | string | User display name | `John Doe` |
| `cliVersion` | string | Gemini CLI version | `0.26.0` |
| `isResume` | boolean | True if resuming session | `false` |

**Access in Hook:**
```python
#!/usr/bin/env python3
import json
import sys

context = json.load(sys.stdin)

project = context['projectPath']
user_email = context['user']['email']

print(f"Session started for {user_email} in {project}")
```

---

### SessionEnd

**Full Context:**
```json
{
  "event": "SessionEnd",
  "timestamp": "2026-02-21T13:00:00Z",
  "sessionId": "abc123...",
  "stats": {
    "duration": 3600,
    "toolsExecuted": 42,
    "tokensUsed": 15000,
    "agentTurns": 8,
    "errors": 0
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `stats.duration` | number | Session duration in seconds |
| `stats.toolsExecuted` | number | Total tools executed |
| `stats.tokensUsed` | number | Total tokens consumed |
| `stats.agentTurns` | number | Number of agent turns |
| `stats.errors` | number | Number of errors |

---

### BeforeAgent

**Full Context:**
```json
{
  "event": "BeforeAgent",
  "timestamp": "2026-02-21T12:05:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "query": "user's input prompt text",
  "conversationHistory": [
    {
      "role": "user",
      "content": "previous user message"
    },
    {
      "role": "assistant",
      "content": "previous assistant response"
    }
  ],
  "context": "existing context string from previous hooks",
  "availableTools": [
    {
      "name": "write_file",
      "description": "Write content to a file",
      "parameters": {...}
    }
  ],
  "modelParameters": {
    "temperature": 0.7,
    "maxTokens": 2048
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `query` | string | User's input prompt |
| `conversationHistory` | array | Previous messages |
| `context` | string | Accumulated context |
| `availableTools` | array | Tools available to agent |
| `modelParameters` | object | LLM parameters |

**Modifiable Fields:**
- `query` - Modify user prompt
- `context` - Add/append context
- `availableTools` - (Only in BeforeToolSelection)
- `modelParameters` - (Only in BeforeModel)

---

### BeforeToolSelection

**Full Context:**
```json
{
  "event": "BeforeToolSelection",
  "timestamp": "2026-02-21T12:06:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "query": "user's request",
  "availableTools": [
    {
      "name": "write_file",
      "description": "Write content to a file",
      "expensive": false,
      "parameters": {
        "type": "object",
        "properties": {
          "path": {"type": "string"},
          "content": {"type": "string"}
        }
      }
    },
    {
      "name": "search_codebase",
      "description": "Search entire codebase",
      "expensive": true,
      "parameters": {...}
    }
  ]
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `query` | string | User's request |
| `availableTools` | array | Full list of available tools |
| `availableTools[].name` | string | Tool name |
| `availableTools[].description` | string | Tool description |
| `availableTools[].expensive` | boolean | Cost indicator |
| `availableTools[].parameters` | object | JSON schema for parameters |

**Modifiable Fields:**
- `availableTools` - Filter tools array

**Example - Filter Expensive Tools:**
```javascript
#!/usr/bin/env node
const input = JSON.parse(require('fs').readFileSync(0, 'utf-8'));

const isSimple = input.query.length < 50;

if (isSimple) {
  const filtered = input.availableTools.filter(t => !t.expensive);
  console.log(JSON.stringify({
    decision: 'continue',
    availableTools: filtered
  }));
} else {
  console.log(JSON.stringify({decision: 'allow'}));
}
```

---

### BeforeTool

**Full Context:**
```json
{
  "event": "BeforeTool",
  "timestamp": "2026-02-21T12:07:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "tool": {
    "name": "write_file",
    "params": {
      "path": "/path/to/file.js",
      "content": "console.log('hello');",
      "mode": "create"
    }
  },
  "metadata": {
    "estimatedCost": 0.001,
    "requiresPermission": true
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `tool.name` | string | Tool being executed |
| `tool.params` | object | Tool parameters (varies by tool) |
| `metadata.estimatedCost` | number | Estimated cost in USD |
| `metadata.requiresPermission` | boolean | Whether user permission needed |

**Modifiable Fields:**
- `tool.params` - Modify tool parameters

**Common Tool Parameters:**

**write_file:**
```json
{
  "path": "/path/to/file",
  "content": "file content",
  "mode": "create" | "overwrite" | "append"
}
```

**execute_command:**
```json
{
  "command": "npm test",
  "cwd": "/working/directory",
  "timeout": 30000
}
```

**read_file:**
```json
{
  "path": "/path/to/file"
}
```

---

### AfterTool

**Full Context:**
```json
{
  "event": "AfterTool",
  "timestamp": "2026-02-21T12:08:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "tool": {
    "name": "write_file",
    "params": {...}
  },
  "result": {
    "success": true,
    "output": "File written successfully",
    "exitCode": 0,
    "stdout": "...",
    "stderr": "...",
    "duration": 0.15
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `result.success` | boolean | Tool execution succeeded |
| `result.output` | string | Human-readable result |
| `result.exitCode` | number | Exit code (for commands) |
| `result.stdout` | string | Standard output |
| `result.stderr` | string | Standard error |
| `result.duration` | number | Execution time in seconds |

---

### BeforeModel

**Full Context:**
```json
{
  "event": "BeforeModel",
  "timestamp": "2026-02-21T12:09:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "modelInput": {
    "prompt": "full prompt including context and tools",
    "tools": ["write_file", "read_file"],
    "parameters": {
      "temperature": 0.7,
      "maxTokens": 2048,
      "topP": 0.9,
      "topK": 40
    },
    "model": "gemini-2.0-flash-exp"
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `modelInput.prompt` | string | Complete prompt sent to model |
| `modelInput.tools` | array | Available tool names |
| `modelInput.parameters.temperature` | number | Sampling temperature (0-1) |
| `modelInput.parameters.maxTokens` | number | Max tokens to generate |
| `modelInput.parameters.topP` | number | Nucleus sampling parameter |
| `modelInput.parameters.topK` | number | Top-k sampling parameter |
| `modelInput.model` | string | Model identifier |

**Modifiable Fields:**
- `modelInput.prompt`
- `modelInput.parameters.*`

---

### AfterModel

**Full Context:**
```json
{
  "event": "AfterModel",
  "timestamp": "2026-02-21T12:10:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "modelOutput": {
    "text": "model's response text",
    "functionCalls": [
      {
        "name": "write_file",
        "params": {
          "path": "/file.js",
          "content": "..."
        }
      }
    ],
    "tokensUsed": 150
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `modelOutput.text` | string | Generated text response |
| `modelOutput.functionCalls` | array | Tools model wants to call |
| `modelOutput.tokensUsed` | number | Tokens consumed |

**Modifiable Fields:**
- `modelOutput.text` - Filter/redact response
- `modelOutput.functionCalls` - Modify tool calls

---

### AfterChunk

**Full Context:**
```json
{
  "event": "AfterChunk",
  "timestamp": "2026-02-21T12:11:00Z",
  "sessionId": "abc123...",
  "requestId": "req456...",
  "chunk": {
    "text": "partial response text",
    "index": 5,
    "totalChunks": 10,
    "isLast": false
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `chunk.text` | string | Partial response text |
| `chunk.index` | number | Chunk sequence number |
| `chunk.totalChunks` | number | Estimated total chunks |
| `chunk.isLast` | boolean | Whether this is last chunk |

**Note:** AfterChunk does NOT support modifying output or flow control

---

### Notification

**Full Context:**
```json
{
  "event": "Notification",
  "timestamp": "2026-02-21T12:30:00Z",
  "sessionId": "abc123...",
  "notificationType": "idle" | "confirmation" | "attention",
  "message": "Description of notification",
  "context": {
    "waitingFor": "user input",
    "idleDuration": 300
  }
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `notificationType` | string | Type of notification |
| `message` | string | Notification message |
| `context.waitingFor` | string | What CLI is waiting for |
| `context.idleDuration` | number | Idle time in seconds |

---

### BeforeCompress

**Full Context:**
```json
{
  "event": "BeforeCompress",
  "timestamp": "2026-02-21T12:45:00Z",
  "sessionId": "abc123...",
  "contextSize": 128000,
  "targetSize": 64000,
  "compressionRatio": 0.5,
  "itemsToCompress": ["message-1", "message-2", "..."]
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `contextSize` | number | Current context size (tokens) |
| `targetSize` | number | Target size after compression |
| `compressionRatio` | number | Compression ratio (0-1) |
| `itemsToCompress` | array | IDs of items to be compressed |

**Note:** Flow-control fields ignored (observability only)

---

### AfterAlert

**Full Context:**
```json
{
  "event": "AfterAlert",
  "timestamp": "2026-02-21T12:50:00Z",
  "sessionId": "abc123...",
  "alertType": "permission" | "warning" | "error",
  "alertMessage": "Alert text shown to user",
  "userResponse": "allow" | "deny" | null
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `alertType` | string | Type of alert |
| `alertMessage` | string | Alert message shown |
| `userResponse` | string\|null | User's response (if applicable) |

**Note:** Flow-control fields ignored (observability only)

---

## Context Injection Patterns

### Adding Context via systemMessage

**BeforeAgent Hook:**
```bash
#!/bin/bash
INPUT=$(cat)

# Gather context
GIT_STATUS=$(git status --short)
CONTEXT="Current git status:\n$GIT_STATUS"

# Inject via systemMessage
jq -n \
  --arg ctx "$CONTEXT" \
  '{decision: "continue", systemMessage: $ctx}'
```

### Modifying Existing Context

**BeforeAgent Hook:**
```python
#!/usr/bin/env python3
import json
import sys

input_data = json.load(sys.stdin)

# Append to existing context
existing = input_data.get('context', '')
new_context = existing + "\n\nAdditional context here"

output = {
    'decision': 'continue',
    'context': new_context
}

print(json.dumps(output))
```

---

## Best Practices

### Accessing Context

**1. Always Validate Fields:**
```python
import json
import sys

context = json.load(sys.stdin)

# Safe access with defaults
tool_name = context.get('tool', {}).get('name', 'unknown')
file_path = context.get('tool', {}).get('params', {}).get('path', '')
```

**2. Handle Missing Fields:**
```bash
#!/bin/bash
INPUT=$(cat)

FILE=$(echo "$INPUT" | jq -r '.tool.params.path // empty')

if [ -z "$FILE" ]; then
  echo "No file path provided"
  echo '{"decision": "allow"}'
  exit 0
fi
```

**3. Use jq for JSON Processing:**
```bash
#!/bin/bash
INPUT=$(cat)

# Extract fields safely
TOOL=$(echo "$INPUT" | jq -r '.tool.name // "unknown"')
PARAMS=$(echo "$INPUT" | jq '.tool.params')

# Process...
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Complete event catalog
- [Scripting & Execution](./scripting.md) - Writing hooks
- [API Reference](./api-reference.md) - Complete JSON schemas
- [Examples](./examples.md) - Real-world examples

---

**Document Version:** 1.0
**Last Updated:** 2026-02-21
**Sources:** Official documentation, research analysis
