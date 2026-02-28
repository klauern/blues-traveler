# OpenAI Codex - Scripting & Execution

**Last Updated:** 2026-02-21

---

## Overview

Codex's scripting model differs fundamentally from local tools because execution happens in **cloud sandboxes**, not on the user's machine. This enables background mode (hours-long tasks), parallel execution, and resource isolation.

---

## Background Mode

### Purpose

Execute long-running tasks without timeout constraints.

**Traditional APIs:** 30s-10min timeouts
**Codex Background Mode:** Unlimited (hours/days)

### Usage

```python
from openai import OpenAI

client = OpenAI()

# Start background task
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,  # ← Enable background mode
    messages=[{
        "role": "user",
        "content": "Refactor entire codebase to TypeScript"
    }]
)

# Response ID for tracking
response_id = response.id

# Task runs in cloud sandbox
# No timeout, no need to keep connection alive
```

### Polling for Status

```python
# Poll for completion
import time

while True:
    status = client.responses.retrieve(response_id)

    if status.status == "completed":
        print(f"Task complete! Output: {status.output}")
        break
    elif status.status == "failed":
        print(f"Task failed: {status.error}")
        break
    elif status.status == "in_progress":
        print(f"Still running... ({status.progress}%)")
        time.sleep(30)  # Check every 30 seconds
```

### Webhook Alternative (Better)

```python
# Instead of polling, use webhooks (configured in dashboard)
# Webhook fires when task completes

@app.post("/webhooks/codex")
async def handle_completion(request: Request):
    event = await request.json()

    if event["type"] == "background_completion":
        response_id = event["data"]["response_id"]
        result = client.responses.retrieve(response_id)

        # Process result
        await create_pull_request(result)

    return {"status": "ok"}
```

---

## Streaming + Background

### Best of Both Worlds

```python
# Get immediate progress updates AND long-running execution
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,    # ← Task persists if connection drops
    stream=True,        # ← Get real-time events
    messages=[{...}]
)

# Stream events in real-time
for event in response.events:
    print(f"{event.type}: {event.data}")

    if event.type == "file_change":
        print(f"  Modified: {event.data['file']}")
    elif event.type == "tool_call":
        print(f"  Used tool: {event.data['tool']}")

# If connection drops, task continues
# Reconnect and resume streaming, or wait for webhook
```

---

## Hook Execution Model

### Synchronous Execution

**Event hooks block Codex until completion:**

```toml
[hooks.file.before_write]
command = "/usr/local/bin/validate-file"
timeout = 30  # Default: 30 seconds

# Execution flow:
# 1. Codex wants to write file
# 2. Before-write hook starts (blocks)
# 3. Hook validates (30s max)
# 4. Hook exits 0 → allow, or exit 1 → block
# 5. Codex proceeds or aborts
```

### Exit Codes

| Exit Code | Meaning | Effect |
|-----------|---------|--------|
| `0` | Success | Allow operation |
| `1` | Failure | Block operation |
| `42` | Retry | Retry operation (webhooks only) |
| Other | Error | Treat as failure |

### Hook Timeout Handling

```toml
[hooks.file.before_write]
command = "/usr/local/bin/slow-validator"
timeout = 60  # Override default 30s

# If hook exceeds timeout:
# - Process killed (SIGKILL)
# - Treated as exit code 1 (failure)
# - Operation blocked
```

---

## Supported Languages

### Native Support (via shebang)

Any language with shebang support:

```bash
#!/bin/bash
# Bash hook

#!/usr/bin/env python3
# Python hook

#!/usr/bin/env node
# Node.js hook

#!/usr/bin/env ruby
# Ruby hook
```

### Hook Script Examples

**Python:**

```python
#!/usr/bin/env python3
import sys
import json

# Read from arguments
tool_name = sys.argv[1]
tool_args = json.loads(sys.argv[2])

# Read from stdin (alternative)
# tool_args = json.load(sys.stdin)

# Validate
if tool_name == "Bash" and "rm -rf" in tool_args.get("command", ""):
    print("ERROR: Dangerous command blocked", file=sys.stderr)
    sys.exit(1)  # Block

sys.exit(0)  # Allow
```

**Node.js:**

```javascript
#!/usr/bin/env node
const fs = require('fs');

const filePath = process.argv[2];
const content = fs.readFileSync(filePath, 'utf8');

// Check for secrets
if (content.match(/(api_key|password|secret).*=.*["']/)) {
  console.error('ERROR: Potential secret detected');
  process.exit(1);  // Block write
}

// Auto-format
const prettier = require('prettier');
const formatted = prettier.format(content, { parser: 'typescript' });
fs.writeFileSync(filePath, formatted);

process.exit(0);  // Allow
```

**Go:**

```go
#!/usr/bin/env go run
package main

import (
    "fmt"
    "os"
    "regexp"
)

func main() {
    filePath := os.Args[1]
    content, _ := os.ReadFile(filePath)

    // Detect secrets
    secretPattern := regexp.MustCompile(`(api_key|password|secret).*=.*["']`)
    if secretPattern.Match(content) {
        fmt.Fprintln(os.Stderr, "ERROR: Secret detected")
        os.Exit(1)  // Block
    }

    os.Exit(0)  // Allow
}
```

---

## Error Handling

### stdout vs stderr

```bash
#!/bin/bash

# stdout: General output (visible to user)
echo "Validating file: $1"

# stderr: Errors and warnings
if ! validate "$1"; then
    echo "ERROR: Validation failed" >&2
    exit 1
fi

# Both captured by Codex
# stderr shown as errors in UI
```

### Error Propagation

```
Hook fails (exit 1)
    ↓
Operation blocked
    ↓
Error shown to user
    ↓
Agent retries or asks for help
```

---

## Timeouts & Resource Limits

### Hook Timeouts

| Hook Type | Default Timeout | Configurable | Max |
|-----------|----------------|--------------|-----|
| Tool Hooks | 30s | ✅ | 300s (5min) |
| File Hooks | 30s | ✅ | 300s |
| Event Hooks | 30s | ✅ | 300s |
| Notifications | 10s | ✅ | 60s |

```toml
[hooks.file.before_write]
command = "/usr/local/bin/slow-validator"
timeout = 120  # 2 minutes
```

### Cloud Sandbox Limits

| Resource | Limit | Notes |
|----------|-------|-------|
| CPU | 4 cores | Per sandbox |
| Memory | 8 GB | Per sandbox |
| Disk | 50 GB | Ephemeral |
| Network | Metered | Outbound only |
| Time | Unlimited | Background mode |

---

## Async & Parallel Execution

### Parallel Tool Execution

**Codex can execute multiple tools simultaneously:**

```
Turn starts
    ↓
├── Bash (install deps) → 30s
├── Read (auth.ts) → 1s
├── Read (middleware.ts) → 1s
└── Read (config.ts) → 1s
    ↓ (all complete)
Write (refactored code)
```

**Total time: ~30s (not 33s sequential)**

### Multi-Agent Parallel Execution

```python
# Run multiple Codex agents in parallel
import asyncio

async def run_agents():
    tasks = [
        client.responses.create_async(
            model="gpt-5.2-codex",
            background=True,
            messages=[{"role": "user", "content": "Add logging"}]
        ),
        client.responses.create_async(
            model="gpt-5.2-codex",
            background=True,
            messages=[{"role": "user", "content": "Write tests"}]
        ),
        client.responses.create_async(
            model="gpt-5.2-codex",
            background=True,
            messages=[{"role": "user", "content": "Update docs"}]
        )
    ]

    results = await asyncio.gather(*tasks)
    return results

# All 3 agents run in parallel in separate sandboxes
results = asyncio.run(run_agents())
```

---

## CLI Non-Interactive Mode

### codex exec

**Purpose:** Run Codex non-interactively (CI/CD, scripts)

```bash
# Basic usage
codex exec "Add error handling to all API endpoints"

# Output to file
codex exec "Refactor auth" --output changes.patch

# Apply changes automatically
codex exec "Fix linting errors" --apply

# Use specific model
codex exec "Add tests" --model gpt-5.3-codex

# Background mode
codex exec "Large refactoring" --background
```

### GitHub Actions Integration

```yaml
# .github/workflows/codex-autofix.yml
- uses: openai/codex-action@v1
  with:
    command: codex exec "Fix CI failures" --apply
  env:
    OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

---

## Performance Optimization

### Prompt Caching

**75% discount on cached tokens:**

```python
# First request: Full price
response = client.responses.create(
    model="gpt-5.2-codex",
    messages=[
        {
            "role": "system",
            "content": huge_codebase  # Expensive
        },
        {
            "role": "user",
            "content": "Add error handling"
        }
    ]
)

# Second request: 75% cheaper on cached content
response2 = client.responses.create(
    model="gpt-5.2-codex",
    messages=[
        {
            "role": "system",
            "content": huge_codebase  # Cached!
        },
        {
            "role": "user",
            "content": "Add logging"  # Only this is full price
        }
    ]
)
```

### Batch Processing

```python
# Process multiple files in one request (cheaper)
files = ["auth.ts", "middleware.ts", "config.ts", "routes.ts"]

response = client.responses.create(
    model="gpt-5.2-codex",
    messages=[{
        "role": "user",
        "content": f"Add error handling to: {', '.join(files)}"
    }]
)

# vs. separate requests (more expensive)
for file in files:
    response = client.responses.create(...)  # 4x API calls
```

---

## Best Practices

### 1. Use Background Mode for Long Tasks

```python
# ✅ Good: Large tasks in background
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,
    messages=[{
        "role": "user",
        "content": "Migrate entire codebase to TypeScript"
    }]
)

# ❌ Bad: Synchronous for hours-long task
# Will timeout or require connection to stay alive
```

### 2. Stream for Progress Updates

```python
# ✅ Good: Stream + background for visibility
response = client.responses.create(
    background=True,
    stream=True,
    messages=[{...}]
)

for event in response.events:
    print(f"Progress: {event.type}")
```

### 3. Set Appropriate Hook Timeouts

```toml
# ✅ Good: Realistic timeout
[hooks.file.before_write]
command = "/usr/local/bin/lint-and-format"
timeout = 60  # Linting can take time

# ❌ Bad: Too short, will kill long-running linters
timeout = 5
```

### 4. Handle Errors Gracefully

```bash
#!/bin/bash
set -e  # Exit on error

# Trap errors
trap 'echo "ERROR: Hook failed at line $LINENO" >&2; exit 1' ERR

# Your logic here
validate_file "$1" || exit 1
```

---

## Sources

- [Responses API Guide](https://platform.openai.com/docs/guides/)
- [Background Mode Documentation](https://platform.openai.com/docs/guides/background)
- [Streaming Responses](https://platform.openai.com/docs/guides/streaming-responses)
- Research: [Automation Patterns](../../../research-notes/codex/automation-patterns.md)

---

**Last Updated:** 2026-02-21
