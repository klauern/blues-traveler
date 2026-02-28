# OpenAI Codex - Event Types & Automation Mechanisms

**Last Updated:** 2026-02-21

---

## Overview

Codex provides **5 distinct automation mechanisms** that operate at different levels of the platform. Unlike traditional hook systems that only provide lifecycle callbacks, Codex offers API-level webhooks, app-level automations, and protocol-level event streaming.

**Key Insight:** Codex is fundamentally different from local tools (like Claude Code) because it's a **cloud-first platform** with automation built into the API layer, not just configuration.

---

## Complete Automation Catalog

| Mechanism | Level | Trigger Type | Cancellable | Async | Config Location |
|-----------|-------|--------------|-------------|-------|-----------------|
| **1. Webhooks** | OpenAI API | Completion events | ❌ | ✅ | OpenAI Dashboard |
| **2. Event Hooks** | Codex CLI/App | Lifecycle (tool/file/event) | ✅ | ❌ | `~/.codex/config.toml` |
| **3. Notifications** | Codex CLI/App | Agent events | ❌ | ✅ | `~/.codex/config.toml` |
| **4. Automations** | Codex App | Schedule/triggers | ❌ | ✅ | App UI |
| **5. App Server Events** | Protocol | Bidirectional JSON-RPC | Varies | ✅ | App Server protocol |

---

## 1. Webhooks (OpenAI API Level)

### Overview

**Standard:** [Standard Webhooks Specification](https://www.standardwebhooks.com/)
**Transport:** HTTP POST to your endpoint
**Authentication:** Signature verification (HMAC-SHA256)
**Configuration:** OpenAI Platform Dashboard

### Supported Webhook Events

| Event Type | Trigger | Payload Contains | Use Case |
|------------|---------|------------------|----------|
| `batch.completed` | Batch processing finishes | Batch ID, status, output file ID | Process large-scale results |
| `background_completion` | Background response finishes | Response ID, status, output | Trigger PR creation |
| `fine_tuning.completed` | Fine-tuning job finishes | Fine-tuned model ID | Deploy new model |
| `thread.archived` | Thread archived (App Server v2) | Thread ID, archive reason | Cleanup workflows |
| `thread.unarchived` | Thread restored | Thread ID | Resume work |

### Webhook Configuration

**1. Configure endpoint in OpenAI Dashboard:**

```
Dashboard → Settings → Webhooks
  ↓
URL: https://your-domain.com/webhooks/codex
Events: [background_completion, batch.completed]
Secret: wh_secret_xxx (generated automatically)
```

**2. Implement webhook handler:**

```python
import hmac
import hashlib
from fastapi import Request, HTTPException

WEBHOOK_SECRET = os.environ["OPENAI_WEBHOOK_SECRET"]

@app.post("/webhooks/codex")
async def handle_webhook(request: Request):
    # Verify signature (Standard Webhooks spec)
    signature = request.headers.get("webhook-signature")
    timestamp = request.headers.get("webhook-timestamp")
    body = await request.body()

    expected_sig = hmac.new(
        WEBHOOK_SECRET.encode(),
        f"{timestamp}.{body.decode()}".encode(),
        hashlib.sha256
    ).hexdigest()

    if not hmac.compare_digest(signature, f"v1,{expected_sig}"):
        raise HTTPException(status_code=401, detail="Invalid signature")

    # Process event
    event = await request.json()

    if event["type"] == "background_completion":
        response_id = event["data"]["response_id"]
        result = client.responses.retrieve(response_id)

        # Trigger downstream workflow
        await create_pull_request(result)
        await notify_team_slack(result)

    return {"status": "ok"}
```

### Webhook Payload Schema

**background_completion event:**

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
    "duration_seconds": 9000
  }
}
```

### Use Cases

- **Event-Driven CI/CD**: Trigger PR creation when background refactoring completes
- **Batch Processing Pipelines**: Chain multiple Codex tasks together
- **Notification Workflows**: Alert team when long-running tasks finish
- **Integration Triggers**: Start downstream systems (deploy, test, review)

---

## 2. Event Hooks (Codex-Specific Lifecycle)

### Overview

**Purpose:** Execute custom scripts at specific lifecycle points during Codex operations
**Configuration:** `~/.codex/config.toml` or project `codex.json`
**Execution:** Synchronous (blocks Codex until hook completes)
**Cancellation:** Hooks can block/cancel operations via exit codes

### Hook Categories

#### Tool Hooks

Trigger when Codex uses tools (Bash, Read, Write, Edit, etc.)

```toml
[hooks.tool.before]
command = "/usr/local/bin/pre-tool-validation"
# Receives: tool name, tool arguments
# Can: Block tool execution (exit 1)

[hooks.tool.after]
command = "/usr/local/bin/post-tool-logging"
# Receives: tool name, tool output, success/failure
# Can: Log for analytics, trigger notifications
```

#### File Hooks

Trigger when Codex modifies files:

```toml
[hooks.file.before_write]
command = "/usr/local/bin/validate-file"
args = ["--strict"]
# Receives: file path, proposed content
# Can: Block write if validation fails

[hooks.file.after_write]
command = "/usr/local/bin/format-and-lint"
# Receives: file path
# Can: Auto-format, run linters, trigger tests
```

#### Event Hooks

Trigger on agent lifecycle events:

```toml
[hooks.event.prompt_gating]
command = "/usr/local/bin/filter-prompt"
# Receives: user prompt
# Can: Modify or block prompts

[hooks.event.stop]
command = "/usr/local/bin/on-completion"
# Receives: final agent state
# Can: Verify work, send notifications

[hooks.event.notification]
command = "/usr/local/bin/send-notification"
events = ["turn_complete", "approval_required", "error"]
# Receives: notification type, details
# Can: Forward to Slack, email, etc.
```

### Hook Execution Model

```
User Request
    ↓
Pre-Tool Hook (before each tool use)
    ↓ (if exit 0)
Tool Execution
    ↓
Post-Tool Hook (after each tool)
    ↓ (if file modified)
Before-Write Hook
    ↓ (if exit 0)
File Write
    ↓
After-Write Hook
    ↓
Event Hook (on turn complete, stop, etc.)
```

### Hook Script Example

```bash
#!/bin/bash
# /usr/local/bin/validate-file

FILE_PATH=$1
CONTENT=$2

# Prevent secrets from being committed
if grep -E '(api_key|password|secret|token).*=.*["\']' "$FILE_PATH"; then
    echo "ERROR: Potential secret detected in $FILE_PATH" >&2
    exit 1  # Block write
fi

# Run linter
if [[ "$FILE_PATH" == *.py ]]; then
    if ! ruff check "$FILE_PATH"; then
        echo "ERROR: Linting failed for $FILE_PATH" >&2
        exit 1  # Block write
    fi
fi

exit 0  # Allow write
```

### Hook Configuration Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | ✅ | Path to executable |
| `args` | array[string] | ❌ | Additional arguments |
| `events` | array[string] | ❌ | Filter to specific event types |
| `pattern` | string (regex) | ❌ | Match tool/file patterns |
| `timeout` | int (seconds) | ❌ | Hook execution timeout (default: 30s) |

---

## 3. Notification System

### Overview

**Purpose:** Execute external programs for side-channel alerts
**Configuration:** `~/.codex/config.toml`
**Execution:** Asynchronous (non-blocking)

### Configuration

```toml
[notify]
command = "/usr/local/bin/notify-webhook"
args = ["--webhook", "https://hooks.slack.com/services/xxx"]
events = ["turn_complete", "approval_required", "error"]
```

### Notification Events

| Event | Trigger | Payload |
|-------|---------|---------|
| `turn_complete` | Agent finishes a turn | Turn ID, tools used, files modified |
| `approval_required` | Agent needs user approval | Approval type, operation details |
| `error` | Agent encounters error | Error message, context |
| `background_start` | Background task starts | Task ID, description |
| `background_complete` | Background task finishes | Task ID, duration, result |

### Example Notification Handler

```python
#!/usr/bin/env python3
# /usr/local/bin/notify-webhook

import sys
import json
import requests

SLACK_WEBHOOK = sys.argv[2]  # from args
event_data = json.loads(sys.stdin.read())

message = {
    "text": f"🤖 Codex {event_data['type']}",
    "blocks": [{
        "type": "section",
        "text": {"type": "mrkdwn", "text": event_data['details']}
    }]
}

requests.post(SLACK_WEBHOOK, json=message)
```

---

## 4. Automations (Codex App)

### Overview

**Purpose:** Scheduled and triggered workflows
**Configuration:** Codex App UI
**Execution:** Background, unprompted

### Automation Components

```yaml
name: "Daily Issue Triage"
schedule: "0 9 * * MON-FRI"  # Cron syntax
instruction: |
  Review all open GitHub issues.
  Categorize by type (bug, feature, question).
  Flag high-priority items.
  Create summary report.
skills:
  - github-triage
  - issue-classifier
reporting:
  mode: inbox  # or "auto_archive"
  notify: true
```

### Schedule Syntax

**Cron format:** `minute hour day month weekday`

```
0 9 * * *       # Daily at 9 AM
0 */4 * * *     # Every 4 hours
0 0 * * SUN     # Weekly on Sunday midnight
0 9 * * MON-FRI # Weekdays at 9 AM
```

### Automation Triggers

**Current (Feb 2026):**
- Schedule-based (cron)

**Coming Soon (roadmap):**
- GitHub push events
- PR opened/updated
- CI failure
- Issue created
- Custom webhooks

### OpenAI's Internal Automations

```yaml
# Issue Triage (9 AM daily)
- Categorize all open issues
- Flag urgent bugs
- Detect duplicates
- Assign labels

# CI Failure Summary (10 AM daily)
- Aggregate last 24h failures
- Group by failure type
- Suggest fixes

# Release Brief (8 AM daily)
- Compile yesterday's merged PRs
- Generate change summary
- Format for standup

# Bug Detection (midnight daily)
- Scan recent commits
- Identify potential bugs
- Create tracking issues

# Documentation Drift (Sunday weekly)
- Check for outdated docs
- Flag inconsistencies
- Suggest updates
```

**Impact Metrics:**
- 60% reduction in manual triage time
- Faster CI failure resolution
- Improved code quality
- Better team awareness

---

## 5. App Server Events (Protocol Level)

### Overview

**Protocol:** JSON-RPC over stdio (JSONL format)
**Direction:** Bidirectional (client ↔ server)
**Purpose:** Real-time event streaming for custom integrations

### Event Types

#### Client → Server Requests

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "conversation/create",
  "params": {
    "message": "Add error handling to all endpoints"
  }
}
```

#### Server → Client Events

```json
{
  "jsonrpc": "2.0",
  "method": "event",
  "params": {
    "type": "agent_message",
    "content": "I'll add try-catch blocks...",
    "timestamp": "2026-02-21T10:30:00Z"
  }
}
```

### Semantic Event Types

| Event Type | Purpose | Contains |
|------------|---------|----------|
| `user_message` | User sent message | Message text |
| `agent_message` | Agent responded | Response text, reasoning |
| `command_run` | Shell command executed | Command, output, exit code |
| `file_change` | File modified | File path, diff |
| `tool_call` | Tool used | Tool name, args, result |
| `approval_request` | Agent needs approval | Operation details |
| `notification` | General notification | Type, message |
| `thread_archived` | Thread closed | Thread ID, reason |

### Protocol Integration Example

```python
import json
import subprocess

# Launch App Server
server = subprocess.Popen(
    ["codex-app-server"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    text=True,
    bufsize=1
)

# Send request
request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "conversation/create",
    "params": {"message": "Refactor authentication"}
}
server.stdin.write(json.dumps(request) + "\n")
server.stdin.flush()

# Receive events
for line in server.stdout:
    event = json.loads(line)

    if "method" in event and event["method"] == "event":
        event_type = event["params"]["type"]

        if event_type == "file_change":
            print(f"Modified: {event['params']['file']}")
        elif event_type == "approval_request":
            # Handle approval
            approval = input(event['params']['message'] + " (y/n): ")
            response = {
                "jsonrpc": "2.0",
                "id": event['params']['id'],
                "result": {"approved": approval == 'y'}
            }
            server.stdin.write(json.dumps(response) + "\n")
            server.stdin.flush()
```

---

## Event Ordering & Lifecycle

### Typical Event Sequence

```
1. user_message (user sends prompt)
     ↓
2. Pre-Tool Hook (validation)
     ↓
3. tool_call (Bash) → command_run event
     ↓
4. Post-Tool Hook (logging)
     ↓
5. Before-Write Hook (validation)
     ↓
6. tool_call (Write) → file_change event
     ↓
7. After-Write Hook (formatting)
     ↓
8. agent_message (response)
     ↓
9. Event Hook (turn_complete)
     ↓
10. notification (Slack alert)
```

### Concurrent Events

- Multiple tool_calls can occur in parallel
- File hooks are sequential (one file at a time)
- App Server events stream in real-time
- Webhooks fire after full completion
- Notifications are async (non-blocking)

---

## Comparison: Hooks vs Webhooks vs Automations

| Feature | Webhooks | Event Hooks | Automations |
|---------|----------|-------------|-------------|
| **Level** | API | CLI/App | App |
| **Trigger** | Completion | Lifecycle | Schedule |
| **Async** | ✅ | ❌ | ✅ |
| **Cancellable** | ❌ | ✅ | ❌ |
| **Config** | Dashboard | config.toml | App UI |
| **Use Case** | Pipelines | Validation | Recurring |

---

## Sources

- [OpenAI Webhooks Guide](https://developers.openai.com/api/docs/guides/webhooks/)
- [Codex Config Reference](https://developers.openai.com/codex/config-reference/)
- [App Server Architecture](https://openai.com/index/unlocking-the-codex-harness/)
- [Automations Guide](https://developers.openai.com/codex/app/automations/)
- Research: [API Capabilities](../../../research-notes/codex/api-capabilities.md)

---

**Last Updated:** 2026-02-21
**Coverage:** 5/5 automation mechanisms (100%)
