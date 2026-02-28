# Cursor IDE Event System Deep Dive

**Research Date:** 2026-02-21

## Event Architecture

Cursor implements a **lifecycle-based event system** for its AI agent, with 6 distinct events that fire at specific points during agent execution.

## Complete Event Catalog

### 1. beforeShellExecution

**Type:** Permission-based (gating)
**Execution:** Synchronous, blocking
**Direction:** Pre-action

#### When It Fires
- Agent attempts to execute any shell command
- Fires before the command reaches the shell
- Fires for both user-initiated and agent-initiated commands

#### Input Payload
```json
{
  "command": "npm install lodash@4.17.21",
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "beforeShellExecution",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

#### Output Schema
```json
{
  "permission": "allow" | "deny" | "ask",
  "userMessage": "Optional: User-facing message",
  "agentMessage": "Optional: Agent-facing technical details"
}
```

#### Permission Modes

1. **`allow`** - Permit command execution
   - Command proceeds to shell
   - No user interaction required
   - Can include informational messages

2. **`deny`** - Block command execution
   - Command is not executed
   - Agent receives rejection
   - Should include explanation in messages

3. **`ask`** - Request user confirmation
   - Cursor displays approval prompt to user
   - User can approve or deny manually
   - Only available in Cursor (not all IDEs support this)

#### Real-World Use Cases

**Security Blocking:**
```bash
#!/bin/bash
# Block dangerous patterns
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if echo "$command" | grep -qE '(rm -rf /|sudo rm|mkfs|dd if=)'; then
  echo '{"permission":"deny","userMessage":"Dangerous command blocked"}'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Tool Enforcement:**
```bash
#!/bin/bash
# Block git, enforce gh
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if [[ "$command" =~ ^git\ ]]; then
  echo '{
    "permission":"deny",
    "userMessage":"Use gh instead of git",
    "agentMessage":"Raw git commands blocked. Use GitHub CLI (gh) instead."
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Package Manager Control:**
```bash
#!/bin/bash
# Block npm, enforce bun
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if [[ "$command" =~ (^npm\ |\ npm\ ) ]]; then
  echo '{
    "permission":"deny",
    "agentMessage":"npm is not allowed, always use bun instead"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

### 2. beforeMCPExecution

**Type:** Permission-based (gating)
**Execution:** Synchronous, blocking
**Direction:** Pre-action

#### When It Fires
- Agent attempts to execute MCP (Model Context Protocol) tool
- Fires before tool parameters are sent to MCP server
- Fires for all MCP tool invocations

#### Input Payload
```json
{
  "tool_name": "filesystem_read",
  "arguments": {
    "path": "/etc/secrets",
    "recursive": true
  },
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "beforeMCPExecution",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

#### Output Schema
```json
{
  "permission": "allow" | "deny" | "ask",
  "userMessage": "Optional: User-facing message",
  "agentMessage": "Optional: Agent-facing technical details"
}
```

#### Real-World Use Cases

**MCP Server Allowlist (ToolHive Integration):**
```bash
#!/bin/bash
# Only allow ToolHive-managed MCP servers
input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')

# Query ToolHive for approved servers
if thv list | grep -q "$tool_name"; then
  echo '{"permission":"allow"}'
else
  echo '{
    "permission":"deny",
    "userMessage":"Unapproved MCP server",
    "agentMessage":"MCP server not managed by ToolHive. Contact admin to add."
  }'
fi
```

**Argument Validation:**
```bash
#!/bin/bash
# Block MCP calls with dangerous arguments
input=$(cat)
arguments=$(echo "$input" | jq -r '.arguments')

if echo "$arguments" | grep -qE '(/etc/|/root/|~/.ssh/)'; then
  echo '{
    "permission":"deny",
    "userMessage":"Cannot access system directories via MCP"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

### 3. beforeReadFile

**Type:** Permission-based (gating)
**Execution:** Synchronous, blocking
**Direction:** Pre-action

#### When It Fires
- Agent attempts to read file contents
- Fires before file is read from disk
- Receives file path AND current file contents
- Can be used for secret redaction

#### Input Payload
```json
{
  "file_path": "/Users/username/projects/myapp/.env",
  "content": "API_KEY=ghp_1234567890abcdef\nDATABASE_URL=postgres://...",
  "attachments": [],
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "beforeReadFile",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

#### Output Schema
```json
{
  "permission": "allow" | "deny",
  "userMessage": "Optional: User-facing message",
  "agentMessage": "Optional: Agent-facing technical details"
}
```

**Note:** `ask` permission is NOT supported for `beforeReadFile`.

#### Real-World Use Cases

**Secret Detection and Blocking:**
```bash
#!/bin/bash
# Block reading files with GitHub tokens
input=$(cat)
content=$(echo "$input" | jq -r '.content')

if echo "$content" | grep -qE 'gh[ps]_[A-Za-z0-9]{36}|gh_api_[A-Za-z0-9]+'; then
  echo '{
    "permission":"deny",
    "userMessage":"Cannot read file containing API keys"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**File Extension Blocking:**
```bash
#!/bin/bash
# Block reading of certain file types
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

case "$file_path" in
  *.pem|*.key|*.p12|*.pfx)
    echo '{"permission":"deny","userMessage":"Cannot read certificate/key files"}'
    exit 0
    ;;
esac

echo '{"permission":"allow"}'
```

**Directory-based Protection:**
```bash
#!/bin/bash
# Protect specific directories
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

if [[ "$file_path" =~ ^/etc/ ]] || [[ "$file_path" =~ /.ssh/ ]]; then
  echo '{"permission":"deny","userMessage":"System files protected"}'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

### 4. afterFileEdit

**Type:** Informational (post-action)
**Execution:** Synchronous, non-blocking for agent
**Direction:** Post-action

#### When It Fires
- Agent completes editing a file
- Fires after file has been written to disk
- Provides old and new content strings
- Cannot prevent the edit (already done)

#### Input Payload
```json
{
  "file_path": "/Users/username/projects/myapp/src/index.ts",
  "edits": [
    {
      "old_string": "const x = 1;",
      "new_string": "const x = 2;"
    },
    {
      "old_string": "function foo() {}",
      "new_string": "function foo() {\n  console.log('bar');\n}"
    }
  ],
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "afterFileEdit",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

#### Output Schema

**No JSON output expected.** Hook exit code indicates success/failure.

- Exit 0: Success
- Exit non-zero: Failure (logged but doesn't affect agent)

Use stderr or log files for output (stdout would interfere with JSON parsing in other hooks).

#### Real-World Use Cases

**Auto-formatting:**
```typescript
import type { AfterFileEditPayload } from "cursor-hooks";

const input: AfterFileEditPayload = await Bun.stdin.json();

if (input.file_path.endsWith(".ts") || input.file_path.endsWith(".tsx")) {
  await Bun.$`bunx prettier --write ${input.file_path}`;
}

if (input.file_path.endsWith(".py")) {
  await Bun.$`ruff format ${input.file_path}`;
}
```

**Audit Logging:**
```bash
#!/bin/bash
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')
timestamp=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$timestamp] Edited: $file_path" >> /tmp/agent-audit.log
echo "$input" >> /tmp/agent-audit.log
```

**Dependency Scanning:**
```bash
#!/bin/bash
# Scan package.json for malware
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

if [[ "$file_path" == *"package.json" ]]; then
  # Extract dependencies and check against malware DB
  edits=$(echo "$input" | jq -r '.edits[0].new_string')
  # Call malware scanning API
  # Log results
  echo "$edits" | check-malware >> /tmp/malware-scan.log
fi
```

**Documentation Updates:**
```bash
#!/bin/bash
# Auto-update README when files change
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

if [[ "$file_path" == *".go" ]]; then
  # Regenerate Go docs
  go doc -all > docs/api.md
fi
```

---

### 5. beforeSubmitPrompt

**Type:** Flow control (gating)
**Execution:** Synchronous, blocking
**Direction:** Pre-action

#### When It Fires
- User submits a prompt in Cursor
- Fires before prompt is sent to AI model
- Can gate prompt submission entirely
- **Note:** Context injection NOT supported

#### Input Payload
```json
{
  "prompt": "Add authentication to the login endpoint",
  "attachments": [],
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "beforeSubmitPrompt",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

#### Output Schema
```json
{
  "continue": true | false
}
```

**Boolean only** - no permission modes, no messages.

- `true`: Allow prompt to be sent to model
- `false`: Block prompt submission

#### Real-World Use Cases

**Keyword Gating:**
```typescript
import type { BeforeSubmitPromptPayload, BeforeSubmitPromptResponse } from "cursor-hooks";

const input: BeforeSubmitPromptPayload = await Bun.stdin.json();

const output: BeforeSubmitPromptResponse = {
  continue: input.prompt.includes("allow") || input.prompt.includes("proceed")
};

console.log(JSON.stringify(output));
```

**Prompt Validation:**
```bash
#!/bin/bash
# Require prompts to be > 10 characters
input=$(cat)
prompt=$(echo "$input" | jq -r '.prompt')

if [ ${#prompt} -lt 10 ]; then
  echo '{"continue":false}'
else
  echo '{"continue":true}'
fi
```

**Session Control:**
```bash
#!/bin/bash
# Allow only during work hours
current_hour=$(date +%H)

if [ $current_hour -ge 9 ] && [ $current_hour -lt 17 ]; then
  echo '{"continue":true}'
else
  echo '{"continue":false}'
fi
```

---

### 6. stop

**Type:** Informational (post-action)
**Execution:** Synchronous, non-blocking
**Direction:** Post-action

#### When It Fires
- AI agent finishes execution/task
- Fires when agent completes current generation
- Can be used for cleanup, notifications, summaries

#### Input Payload
```json
{
  "status": "completed",
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "stop",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

#### Output Schema
```json
{
  "followup_message": "Optional: Message to send to the agent"
}
```

#### Real-World Use Cases

**macOS Notifications:**
```bash
#!/bin/bash
# Display notification when agent completes
osascript -e 'display notification "Cursor agent completed task" with title "Cursor"'
```

**Session Summary:**
```bash
#!/bin/bash
# Read malware detections from earlier hooks and report
input=$(cat)
generation_id=$(echo "$input" | jq -r '.generation_id')

if [ -f "./malware_detected_packages_${generation_id}.txt" ]; then
  summary=$(cat "./malware_detected_packages_${generation_id}.txt")
  echo "{\"followup_message\":\"Malware detected: $summary\"}"
  rm "./malware_detected_packages_${generation_id}.txt"
else
  echo "{\"followup_message\":\"Session completed successfully\"}"
fi
```

**Cleanup:**
```bash
#!/bin/bash
# Remove temporary files created during session
input=$(cat)
generation_id=$(echo "$input" | jq -r '.generation_id')

rm -f /tmp/cursor-session-${generation_id}-*.tmp
```

---

## Event Lifecycle Flow

```
User submits prompt
        │
        ▼
[beforeSubmitPrompt] ──(deny)──> STOP
        │(allow)
        ▼
Agent processes prompt
        │
        ├──> Read file?
        │    [beforeReadFile] ──(deny)──> Skip file
        │         │(allow)
        │         ▼
        │    Read file contents
        │
        ├──> Execute shell?
        │    [beforeShellExecution] ──(deny/ask-denied)──> Skip command
        │         │(allow/ask-approved)
        │         ▼
        │    Execute shell command
        │
        ├──> Call MCP tool?
        │    [beforeMCPExecution] ──(deny/ask-denied)──> Skip tool
        │         │(allow/ask-approved)
        │         ▼
        │    Execute MCP tool
        │
        ├──> Edit file?
        │    Agent edits file
        │         │
        │         ▼
        │    [afterFileEdit] (informational)
        │
        ▼
Agent completes
        │
        ▼
[stop] (cleanup/summary)
```

## Event Comparison Matrix

| Event | Type | Blocking | Permission Modes | Messages | Use Case |
|-------|------|----------|------------------|----------|----------|
| beforeShellExecution | Pre-action | Yes | allow/deny/ask | Both | Security, tool enforcement |
| beforeMCPExecution | Pre-action | Yes | allow/deny/ask | Both | MCP governance |
| beforeReadFile | Pre-action | Yes | allow/deny | Both | Secret protection |
| afterFileEdit | Post-action | No | N/A | N/A | Formatting, audit |
| beforeSubmitPrompt | Pre-action | Yes | continue only | None | Prompt gating |
| stop | Post-action | No | followup only | Agent | Cleanup, summary |

## Execution Semantics

### Synchronous vs Asynchronous

**All hooks are synchronous:**
- Cursor waits for hook to complete
- Hook blocks agent action
- No async/parallel hook execution

### Multiple Hooks per Event

When multiple hooks are defined for same event:
```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hook1.sh" },
      { "command": "./hook2.sh" },
      { "command": "./hook3.sh" }
    ]
  }
}
```

**Execution order:**
- Hooks execute in array order (top to bottom)
- Each hook runs to completion before next starts
- First `deny` blocks action (subsequent hooks may not run)
- All `allow` results needed for final allow

### Error Propagation

Hook failure handling:
- Hook crash → Fail-safe (typically deny/block)
- Invalid JSON → Block with error message
- Timeout (if implemented) → Likely blocks
- Exit non-zero → Depends on JSON output

## Environment Context

### Conversation vs Generation

**Conversation ID:**
- Unique per chat session
- Persists across multiple prompts
- Resets when starting new chat

**Generation ID:**
- Unique per prompt within conversation
- New ID for each user message
- Used to track hook execution within session

### Workspace Roots

**Single workspace:**
```json
"workspace_roots": ["/Users/username/project"]
```

**Multi-root workspace (VS Code):**
```json
"workspace_roots": [
  "/Users/username/project-frontend",
  "/Users/username/project-backend"
]
```

## Advanced Patterns

### State Sharing Between Hooks

Hooks can share state via filesystem:

**beforeShellExecution writes:**
```bash
generation_id=$(echo "$input" | jq -r '.generation_id')
echo "suspicious-package" > "/tmp/scan-${generation_id}.txt"
```

**stop reads:**
```bash
generation_id=$(echo "$input" | jq -r '.generation_id')
if [ -f "/tmp/scan-${generation_id}.txt" ]; then
  findings=$(cat "/tmp/scan-${generation_id}.txt")
  echo "{\"followup_message\":\"Issues found: $findings\"}"
  rm "/tmp/scan-${generation_id}.txt"
fi
```

### Audit Trail

Complete session audit using generation_id:
```bash
# All hooks append to same log
generation_id=$(echo "$input" | jq -r '.generation_id')
timestamp=$(date -Iseconds)
hook_event=$(echo "$input" | jq -r '.hook_event_name')

echo "[$timestamp][$generation_id][$hook_event] $input" \
  >> "/tmp/cursor-audit.log"
```

## Limitations

### What You CANNOT Do

1. **Modify prompt content** (beforeSubmitPrompt)
   - Can only allow/deny, not edit
   - No context injection

2. **Prevent file edits** (afterFileEdit)
   - Edit already happened
   - Can only react, not prevent

3. **Async execution**
   - All hooks are synchronous
   - No parallel execution

4. **Hook chaining control**
   - Cannot control execution order beyond array order
   - Cannot skip subsequent hooks

5. **Custom environment variables**
   - Cannot pass env vars via hooks.json
   - Must rely on shell environment

6. **Timeout configuration**
   - No per-hook timeout in config
   - Global timeout (if any) not documented

## References

- GitButler Deep Dive: https://blog.gitbutler.com/cursor-hooks-deep-dive
- TypeScript Types: https://github.com/johnlindquist/cursor-hooks
- Python Models: https://github.com/DevonFulcher/py-cursor-hooks
- Shell Examples: https://github.com/hamzafer/cursor-hooks
