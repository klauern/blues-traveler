# Cursor IDE Event Types & Triggers

**Complete reference for all 6 Cursor hook events**

---

## Complete Event Catalog

| Event | Trigger | Timing | Cancellable | Permission Modes | Context Available |
|-------|---------|--------|-------------|------------------|-------------------|
| beforeShellExecution | Shell command initiated | Before | ✅ Yes | allow/deny/ask | command, conversation_id, generation_id |
| beforeMCPExecution | MCP tool called | Before | ✅ Yes | allow/deny/ask | tool_name, arguments |
| beforeReadFile | File read requested | Before | ✅ Yes | allow/deny | file_path, content |
| afterFileEdit | File modified by agent | After | ❌ No | N/A | file_path, edits |
| beforeSubmitPrompt | User submits prompt | Before | ✅ Yes | continue only | prompt, attachments |
| stop | Agent completes task | After | ❌ No | followup only | status, conversation_id |

---

## Event Categories

### Pre-Execution Events (Gating)

These events fire **before** an action and can **block** execution.

- beforeShellExecution
- beforeMCPExecution
- beforeReadFile
- beforeSubmitPrompt

### Post-Execution Events (Informational)

These events fire **after** an action and are **informational only**.

- afterFileEdit
- stop

---

## Event Details

### 1. beforeShellExecution

**Purpose:** Control and validate shell commands before execution
**Type:** Pre-execution (gating)
**Can Cancel:** ✅ Yes

#### When It Fires

- Agent attempts to execute any shell command
- Fires before command reaches the shell
- Includes both user-initiated and agent-initiated commands

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

**Field Descriptions:**
- `command` (string): Full shell command to execute
- `conversation_id` (uuid): Unique ID for chat session
- `generation_id` (uuid): Unique ID for this prompt within conversation
- `hook_event_name` (string): Always "beforeShellExecution"
- `workspace_roots` (array): Project root paths

#### Output Schema

```json
{
  "permission": "allow" | "deny" | "ask",
  "userMessage": "Optional: User-facing message",
  "agentMessage": "Optional: Agent-facing technical details"
}
```

**Permission Modes:**

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
   - **Unique to Cursor** (not available in all IDEs)

#### Common Use Cases

**Security Blocking:**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if echo "$command" | grep -qE '(rm -rf /|sudo rm|mkfs|dd if=)'; then
  echo '{
    "permission": "deny",
    "userMessage": "Dangerous command blocked",
    "agentMessage": "Blocked pattern: filesystem destruction"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Tool Enforcement (gh over git):**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if [[ "$command" =~ ^git\ ]]; then
  echo '{
    "permission": "deny",
    "userMessage": "Use gh instead of git",
    "agentMessage": "Raw git commands blocked. Use GitHub CLI (gh) instead."
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Package Manager Control:**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if [[ "$command" =~ (^npm\ |\ npm\ ) ]]; then
  echo '{
    "permission": "deny",
    "agentMessage": "npm is not allowed, always use bun instead"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

### 2. beforeMCPExecution

**Purpose:** Control MCP (Model Context Protocol) tool execution
**Type:** Pre-execution (gating)
**Can Cancel:** ✅ Yes

#### When It Fires

- Agent attempts to execute MCP tool
- Fires before tool parameters sent to MCP server
- Applies to all MCP tool invocations

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

**Field Descriptions:**
- `tool_name` (string): MCP tool being invoked
- `arguments` (object): Tool-specific parameters
- Other fields same as beforeShellExecution

#### Output Schema

```json
{
  "permission": "allow" | "deny" | "ask",
  "userMessage": "Optional: User-facing message",
  "agentMessage": "Optional: Agent-facing technical details"
}
```

**Permission Modes:** Same as beforeShellExecution (allow/deny/ask)

#### Common Use Cases

**MCP Server Allowlist:**
```bash
#!/bin/bash
input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')

# Only allow approved MCP servers
if thv list | grep -q "$tool_name"; then
  echo '{"permission":"allow"}'
else
  echo '{
    "permission": "deny",
    "userMessage": "Unapproved MCP server",
    "agentMessage": "MCP server not managed by ToolHive. Contact admin."
  }'
fi
```

**Argument Validation:**
```bash
#!/bin/bash
input=$(cat)
arguments=$(echo "$input" | jq -r '.arguments')

if echo "$arguments" | grep -qE '(/etc/|/root/|~/.ssh/)'; then
  echo '{
    "permission": "deny",
    "userMessage": "Cannot access system directories via MCP"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

### 3. beforeReadFile

**Purpose:** Control file access and redact sensitive content
**Type:** Pre-execution (gating)
**Can Cancel:** ✅ Yes

#### When It Fires

- Agent attempts to read file contents
- Fires before file read from disk
- **Receives file path AND current file contents**
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

**Field Descriptions:**
- `file_path` (string): Absolute path to file
- `content` (string): Current file contents (before sending to agent)
- `attachments` (array): Additional file attachments (usually empty)

#### Output Schema

```json
{
  "permission": "allow" | "deny",
  "userMessage": "Optional: User-facing message",
  "agentMessage": "Optional: Agent-facing technical details"
}
```

**Note:** `ask` permission is **NOT supported** for `beforeReadFile`.

#### Common Use Cases

**Secret Detection:**
```bash
#!/bin/bash
input=$(cat)
content=$(echo "$input" | jq -r '.content')

if echo "$content" | grep -qE 'gh[ps]_[A-Za-z0-9]{36}|gh_api_[A-Za-z0-9]+'; then
  echo '{
    "permission": "deny",
    "userMessage": "Cannot read file containing API keys"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**File Extension Blocking:**
```bash
#!/bin/bash
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

**Purpose:** Post-process file edits (formatting, validation, audit)
**Type:** Post-execution (informational)
**Can Cancel:** ❌ No (edit already happened)

#### When It Fires

- Agent completes editing a file
- Fires after file written to disk
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

**Field Descriptions:**
- `file_path` (string): Absolute path to edited file
- `edits` (array): List of old_string → new_string replacements
- `old_string` (string): Content before edit
- `new_string` (string): Content after edit

#### Output Schema

**No JSON output expected.** Hook exit code indicates success/failure.

- Exit 0: Success
- Exit non-zero: Failure (logged but doesn't affect agent)

**Important:** Use stderr or log files for output. stdout would interfere with JSON parsing in other hooks.

#### Common Use Cases

**Auto-formatting (TypeScript):**
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
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

if [[ "$file_path" == *"package.json" ]]; then
  edits=$(echo "$input" | jq -r '.edits[0].new_string')
  echo "$edits" | check-malware >> /tmp/malware-scan.log
fi
```

---

### 5. beforeSubmitPrompt

**Purpose:** Gate or filter prompts before sending to model
**Type:** Pre-execution (flow control)
**Can Cancel:** ✅ Yes

#### When It Fires

- User submits a prompt in Cursor
- Fires before prompt sent to AI model
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

**Field Descriptions:**
- `prompt` (string): User's submitted prompt text
- `attachments` (array): Attached files or context

#### Output Schema

```json
{
  "continue": true | false
}
```

**Boolean only** - no permission modes, no messages.

- `true`: Allow prompt to be sent to model
- `false`: Block prompt submission

#### Common Use Cases

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
input=$(cat)
prompt=$(echo "$input" | jq -r '.prompt')

if [ ${#prompt} -lt 10 ]; then
  echo '{"continue":false}'
else
  echo '{"continue":true}'
fi
```

**Session Control (Work Hours):**
```bash
#!/bin/bash
current_hour=$(date +%H)

if [ $current_hour -ge 9 ] && [ $current_hour -lt 17 ]; then
  echo '{"continue":true}'
else
  echo '{"continue":false}'
fi
```

---

### 6. stop

**Purpose:** Cleanup, notifications, and summaries after agent completes
**Type:** Post-execution (completion)
**Can Cancel:** ❌ No

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

**Field Descriptions:**
- `status` (string): Completion status ("completed")

#### Output Schema

```json
{
  "followup_message": "Optional: Message to send to the agent"
}
```

#### Common Use Cases

**macOS Notifications:**
```bash
#!/bin/bash
osascript -e 'display notification "Cursor agent completed task" with title "Cursor"'
```

**Session Summary:**
```bash
#!/bin/bash
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

---

## Event Comparison Matrix

| Event | Type | Blocking | Permission Modes | Messages | Primary Use Case |
|-------|------|----------|------------------|----------|------------------|
| beforeShellExecution | Pre-action | Yes | allow/deny/ask | Both | Security, tool enforcement |
| beforeMCPExecution | Pre-action | Yes | allow/deny/ask | Both | MCP governance |
| beforeReadFile | Pre-action | Yes | allow/deny | Both | Secret protection |
| afterFileEdit | Post-action | No | N/A | N/A | Formatting, audit |
| beforeSubmitPrompt | Pre-action | Yes | continue only | None | Prompt gating |
| stop | Post-action | No | followup only | Agent | Cleanup, summary |

---

## Execution Semantics

### Synchronous Execution

**All hooks are synchronous:**
- Cursor waits for hook to complete
- Hook blocks agent action
- No async/parallel hook execution within event

### Multiple Hooks per Event

When multiple hooks defined for same event:

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

---

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
generation_id=$(echo "$input" | jq -r '.generation_id')
timestamp=$(date -Iseconds)
hook_event=$(echo "$input" | jq -r '.hook_event_name')

echo "[$timestamp][$generation_id][$hook_event] $input" >> /tmp/cursor-audit.log
```

---

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

---

**Sources:**
- [Cursor Official Documentation](https://cursor.com/docs/agent/hooks)
- [GitButler Deep Dive](https://blog.gitbutler.com/cursor-hooks-deep-dive)
- [TypeScript Types](https://github.com/johnlindquist/cursor-hooks)
- [Python Models](https://github.com/DevonFulcher/py-cursor-hooks)
- [Shell Examples](https://github.com/hamzafer/cursor-hooks)
