# OpenAI Codex - Environment & Context

**Last Updated:** 2026-02-21

---

## Overview

Codex provides rich context to agents and hooks through environment variables, API payloads, and structured data fields. Unlike local tools, Codex runs in cloud sandboxes with controlled environment injection.

---

## Environment Variables

### System-Provided Variables (Cloud Sandbox)

| Variable | Type | Available In | Description | Example |
|----------|------|--------------|-------------|---------|
| `CODEX_RESPONSE_ID` | string | All contexts | Unique response ID | `resp_abc123` |
| `CODEX_THREAD_ID` | string | Conversations | Thread/conversation ID | `thread_xyz789` |
| `CODEX_TURN_ID` | string | Each turn | Current turn ID | `turn_456` |
| `CODEX_MODEL` | string | All contexts | Model being used | `gpt-5.2-codex` |
| `CODEX_WORKSPACE` | path | Cloud sandbox | Workspace directory | `/sandbox/workspace` |
| `CODEX_PROJECT_DIR` | path | CLI mode | Project root | `/Users/name/project` |
| `HOME` | path | Cloud sandbox | Home directory | `/sandbox/home` |
| `PWD` | path | Cloud sandbox | Current directory | `/sandbox/workspace` |

### User-Injected Variables

```python
# Inject environment variables via API
response = client.responses.create(
    model="gpt-5.2-codex",
    environment={
        "DATABASE_URL": os.environ["DB_URL"],
        "API_KEY": os.environ["API_KEY"],
        "CUSTOM_VAR": "value"
    },
    messages=[{...}]
)
```

**Security Note:** Never hardcode secrets in code. Use environment variable references.

---

## Hook Context Data

### Tool Hook Context

**Before Tool Execution:**

```bash
#!/bin/bash
# Hook receives as arguments:
TOOL_NAME=$1       # e.g., "Bash", "Write", "Edit"
TOOL_ARGS_JSON=$2  # JSON string of tool arguments

# Parse tool arguments
echo "$TOOL_ARGS_JSON" | jq .
# Example for Bash tool:
# {
#   "command": "ls -la",
#   "workingDirectory": "/sandbox/workspace"
# }
```

**After Tool Execution:**

```bash
#!/bin/bash
TOOL_NAME=$1
TOOL_OUTPUT=$2      # Tool's output (stdout/stderr)
TOOL_EXIT_CODE=$3   # 0 for success, non-zero for failure
TOOL_DURATION=$4    # Execution time in milliseconds
```

### File Hook Context

**Before Write:**

```bash
#!/bin/bash
FILE_PATH=$1        # Absolute path to file
PROPOSED_CONTENT=$2  # Content to be written (stdin also available)

# Can read proposed content
echo "$PROPOSED_CONTENT" | grep "secret"  # Detect secrets

# Or from stdin
CONTENT=$(cat)
```

**After Write:**

```bash
#!/bin/bash
FILE_PATH=$1
BYTES_WRITTEN=$2
MODIFICATION_TIME=$3  # ISO 8601 timestamp
```

### Event Hook Context

```bash
#!/bin/bash
EVENT_TYPE=$1  # e.g., "turn_complete", "stop", "notification"
EVENT_DATA=$2  # JSON payload (stdin also available)

case $EVENT_TYPE in
  "turn_complete")
    TOOLS_USED=$(echo "$EVENT_DATA" | jq -r '.tools_used | join(", ")')
    FILES_MODIFIED=$(echo "$EVENT_DATA" | jq -r '.files_modified | join(", ")')
    echo "Turn complete: used $TOOLS_USED, modified $FILES_MODIFIED"
    ;;
  "notification")
    MESSAGE=$(echo "$EVENT_DATA" | jq -r '.message')
    SEVERITY=$(echo "$EVENT_DATA" | jq -r '.severity')
    echo "[$SEVERITY] $MESSAGE"
    ;;
esac
```

---

## API Request Context

### Responses API Context

**Request Payload:**

```json
{
  "model": "gpt-5.2-codex",
  "messages": [
    {
      "role": "system",
      "content": "You are a code refactoring expert."
    },
    {
      "role": "user",
      "content": "Refactor authentication module"
    }
  ],
  "environment": {
    "PROJECT_NAME": "MyApp",
    "FRAMEWORK": "Express",
    "DATABASE": "PostgreSQL"
  },
  "metadata": {
    "user_id": "user_123",
    "session_id": "sess_456"
  }
}
```

**Response Context:**

```json
{
  "id": "resp_abc123",
  "model": "gpt-5.2-codex",
  "created": 1708531200,
  "output": "...",
  "usage": {
    "prompt_tokens": 1500,
    "completion_tokens": 800,
    "total_tokens": 2300,
    "cached_tokens": 1200
  },
  "metadata": {
    "tools_used": ["Bash", "Write", "Edit"],
    "files_modified": ["auth.ts", "middleware.ts"],
    "duration_ms": 45000
  }
}
```

---

## App Server Protocol Context

### Event Payload Structure

**Generic Event:**

```json
{
  "jsonrpc": "2.0",
  "method": "event",
  "params": {
    "type": "file_change",
    "timestamp": "2026-02-21T10:30:00Z",
    "thread_id": "thread_xyz789",
    "turn_id": "turn_456",
    "data": {
      "file": "/sandbox/workspace/auth.ts",
      "change_type": "modified",
      "diff": "..."
    }
  }
}
```

### Tool Call Event:

```json
{
  "type": "tool_call",
  "timestamp": "2026-02-21T10:30:00Z",
  "data": {
    "tool": "Bash",
    "arguments": {
      "command": "npm test",
      "workingDirectory": "/sandbox/workspace"
    },
    "result": {
      "stdout": "All tests passed",
      "stderr": "",
      "exitCode": 0
    },
    "duration_ms": 3500
  }
}
```

### Approval Request Context:

```json
{
  "type": "approval_request",
  "id": "approval_789",
  "data": {
    "operation": "bash_command",
    "command": "rm -rf node_modules",
    "reason": "Reinstalling dependencies",
    "risk_level": "medium"
  }
}
```

---

## Webhook Payload Context

### Background Completion Webhook:

```json
{
  "type": "background_completion",
  "timestamp": "2026-02-21T10:30:00Z",
  "data": {
    "response_id": "resp_abc123",
    "status": "completed",
    "model": "gpt-5.2-codex",
    "created_at": "2026-02-21T08:00:00Z",
    "completed_at": "2026-02-21T10:30:00Z",
    "duration_seconds": 9000,
    "metadata": {
      "tools_used": ["Bash", "Write", "Edit", "Read"],
      "files_modified": 47,
      "total_tokens": 125000,
      "cached_tokens": 80000
    }
  }
}
```

---

## Custom Context Injection

### Via AGENTS.md

```markdown
# Project Context

## Tech Stack
- Framework: Next.js 14
- Database: PostgreSQL with Prisma
- Auth: NextAuth.js
- Deployment: Vercel

## Code Conventions
- Use TypeScript strict mode
- Functional components with hooks
- Tailwind for styling
- Jest + React Testing Library

## Important Paths
- API routes: `/app/api/**/*.ts`
- Components: `/components/**/*.tsx`
- Database schema: `/prisma/schema.prisma`
```

**Codex automatically reads and uses this context**

### Via API Metadata

```python
response = client.responses.create(
    model="gpt-5.2-codex",
    messages=[{...}],
    metadata={
        "project": "MyApp",
        "component": "authentication",
        "jira_ticket": "PROJ-123",
        "reviewer": "alice@example.com"
    }
)
```

**Metadata flows through webhooks and events**

---

## Context Availability Matrix

| Context Type | Webhooks | Event Hooks | App Server Events | API Response |
|--------------|----------|-------------|-------------------|--------------|
| Response ID | ✅ | ✅ | ✅ | ✅ |
| Thread ID | ✅ | ✅ | ✅ | ✅ |
| Model Used | ✅ | ✅ | ✅ | ✅ |
| Tools Used | ✅ | ✅ | ✅ | ✅ |
| Files Modified | ✅ | ✅ | ✅ | ✅ |
| Token Usage | ✅ | ❌ | ❌ | ✅ |
| Duration | ✅ | ❌ | ❌ | ✅ |
| User Metadata | ✅ | ✅ | ✅ | ✅ |

---

## Accessing Context in Hooks

### Bash Hook Example

```bash
#!/bin/bash
# /usr/local/bin/log-tool-usage

# Arguments
TOOL_NAME=$1
TOOL_ARGS=$2

# Environment variables
RESPONSE_ID=$CODEX_RESPONSE_ID
THREAD_ID=$CODEX_THREAD_ID

# Log to file
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ)|$RESPONSE_ID|$TOOL_NAME|$TOOL_ARGS" \
  >> ~/.codex/tool-usage.log

# Send to analytics
curl -X POST https://analytics.example.com/events \
  -H "Content-Type: application/json" \
  -d "{
    \"event\": \"tool_used\",
    \"response_id\": \"$RESPONSE_ID\",
    \"thread_id\": \"$THREAD_ID\",
    \"tool\": \"$TOOL_NAME\",
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
  }"
```

### Python Hook Example

```python
#!/usr/bin/env python3
# /usr/local/bin/notify-slack

import os
import sys
import json
import requests

# Read from arguments and stdin
event_type = sys.argv[1]
event_data = json.loads(sys.stdin.read())

# Access environment variables
response_id = os.environ.get('CODEX_RESPONSE_ID')
thread_id = os.environ.get('CODEX_THREAD_ID')

# Format message
message = {
    "text": f"🤖 Codex {event_type}",
    "blocks": [{
        "type": "section",
        "fields": [
            {"type": "mrkdwn", "text": f"*Response:* {response_id}"},
            {"type": "mrkdwn", "text": f"*Thread:* {thread_id}"},
            {"type": "mrkdwn", "text": f"*Event:* {event_type}"}
        ]
    }]
}

# Send to Slack
webhook_url = os.environ.get('SLACK_WEBHOOK')
requests.post(webhook_url, json=message)
```

---

## Best Practices

### 1. Use Environment Variables for Secrets

```python
# ✅ Good: Reference environment variables
environment={
    "DATABASE_URL": os.environ["DB_URL"],
    "API_KEY": os.environ["API_KEY"]
}

# ❌ Bad: Hardcode secrets
environment={
    "DATABASE_URL": "postgres://user:pass@host/db",
    "API_KEY": "sk-abc123"
}
```

### 2. Validate Context in Hooks

```bash
#!/bin/bash
# Validate required environment variables
if [ -z "$CODEX_RESPONSE_ID" ]; then
    echo "ERROR: CODEX_RESPONSE_ID not set" >&2
    exit 1
fi

if [ -z "$CODEX_THREAD_ID" ]; then
    echo "ERROR: CODEX_THREAD_ID not set" >&2
    exit 1
fi
```

### 3. Include Context in Logs

```bash
# Log with full context
echo "[$(date -u)] [$CODEX_RESPONSE_ID] [$EVENT_TYPE] $MESSAGE" \
  >> ~/.codex/events.log
```

### 4. Pass Context to Downstream Systems

```python
# Forward context to analytics
analytics.track(
    user_id=os.environ['CODEX_USER_ID'],
    event="codex_completion",
    properties={
        "response_id": os.environ['CODEX_RESPONSE_ID'],
        "thread_id": os.environ['CODEX_THREAD_ID'],
        "model": os.environ['CODEX_MODEL'],
        "tools_used": event_data['tools_used'],
        "duration_ms": event_data['duration_ms']
    }
)
```

---

## Sources

- [Responses API Documentation](https://platform.openai.com/docs/guides/)
- [App Server Protocol](https://developers.openai.com/codex/app-server/)
- [Hooks Reference](https://developers.openai.com/codex/config-reference/#hooks)
- Research: [API Capabilities](../../../research-notes/codex/api-capabilities.md)

---

**Last Updated:** 2026-02-21
