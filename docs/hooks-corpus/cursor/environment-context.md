# Cursor IDE Environment & Context

**Understanding the data available to hooks**

---

## Context Data Structures

### Common Fields (All Events)

Every hook receives these fields via JSON stdin:

```json
{
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "beforeShellExecution",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| conversation_id | UUID | Unique ID for chat session | `550e8400-e29b-...` |
| generation_id | UUID | Unique ID for this prompt | `6ba7b810-9dad-...` |
| hook_event_name | String | Event that triggered hook | `beforeShellExecution` |
| workspace_roots | Array[String] | Project root paths | `["/path/to/project"]` |

---

## Event-Specific Fields

### beforeShellExecution

```json
{
  "command": "npm install lodash@4.17.21",
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "beforeShellExecution",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

**Unique Fields:**
- `command` (string): Full shell command to execute

**Access in Bash:**
```bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')
conversation_id=$(echo "$input" | jq -r '.conversation_id')
```

**Access in TypeScript:**
```typescript
import type { BeforeShellExecutionPayload } from "cursor-hooks";
const input: BeforeShellExecutionPayload = await Bun.stdin.json();
console.log(input.command);
```

---

### beforeMCPExecution

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

**Unique Fields:**
- `tool_name` (string): MCP tool being invoked
- `arguments` (object): Tool-specific parameters (structure varies)

**Access in Bash:**
```bash
input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')
path_arg=$(echo "$input" | jq -r '.arguments.path')
```

---

### beforeReadFile

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

**Unique Fields:**
- `file_path` (string): Absolute path to file
- `content` (string): Current file contents
- `attachments` (array): Additional attachments (usually empty)

**Access in Bash:**
```bash
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')
content=$(echo "$input" | jq -r '.content')
```

---

### afterFileEdit

```json
{
  "file_path": "/Users/username/projects/myapp/src/index.ts",
  "edits": [
    {
      "old_string": "const x = 1;",
      "new_string": "const x = 2;"
    }
  ],
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "afterFileEdit",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

**Unique Fields:**
- `file_path` (string): Absolute path to edited file
- `edits` (array): List of string replacements
  - `old_string` (string): Content before edit
  - `new_string` (string): Content after edit

**Access in TypeScript:**
```typescript
import type { AfterFileEditPayload } from "cursor-hooks";
const input: AfterFileEditPayload = await Bun.stdin.json();

for (const edit of input.edits) {
  console.log("Changed:", edit.old_string, "->", edit.new_string);
}
```

---

### beforeSubmitPrompt

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

**Unique Fields:**
- `prompt` (string): User's submitted prompt text
- `attachments` (array): Attached files/context

---

### stop

```json
{
  "status": "completed",
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "generation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "hook_event_name": "stop",
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

**Unique Fields:**
- `status` (string): Completion status (e.g., "completed")

---

## Conversation vs Generation IDs

### Conversation ID

**Purpose:** Track entire chat session

**Lifetime:**
- Created when user starts new chat
- Persists across multiple prompts
- Reset when starting new chat

**Use Cases:**
- Session-level state
- Conversation-wide logging
- Multi-prompt workflows

**Example:**
```bash
conversation_id=$(echo "$input" | jq -r '.conversation_id')
echo "$data" >> "/tmp/conv-${conversation_id}.log"
```

### Generation ID

**Purpose:** Track individual prompt within conversation

**Lifetime:**
- New ID for each user prompt
- Unique within conversation
- Used for prompt-level state

**Use Cases:**
- Temporary state sharing between hooks
- Per-prompt tracking
- Cleanup after prompt completes

**Example:**
```bash
generation_id=$(echo "$input" | jq -r '.generation_id')
echo "finding" > "/tmp/gen-${generation_id}.txt"

# Later hook or stop event reads it
findings=$(cat "/tmp/gen-${generation_id}.txt")
rm "/tmp/gen-${generation_id}.txt"
```

---

## Workspace Roots

### Single Workspace

```json
{
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

**Most common scenario:** One project root

### Multi-Root Workspace

```json
{
  "workspace_roots": [
    "/Users/username/projects/frontend",
    "/Users/username/projects/backend",
    "/Users/username/projects/shared-lib"
  ]
}
```

**Use Case:** VS Code multi-root workspaces

### Working with Workspace Roots

**Validate file is in workspace:**
```bash
file_path=$(echo "$input" | jq -r '.file_path')
workspace_roots=$(echo "$input" | jq -r '.workspace_roots[]')

in_workspace=false
for root in $workspace_roots; do
  if [[ "$file_path" == "$root"* ]]; then
    in_workspace=true
    break
  fi
done

if [ "$in_workspace" = false ]; then
  echo '{"permission":"deny","userMessage":"File outside workspace"}'
  exit 0
fi
```

---

## Environment Variables

### Available Variables

**Undocumented:** Cursor does not officially document environment variables passed to hooks.

**Likely Available (Standard Shell):**
- `HOME` - User home directory
- `USER` - Username
- `PWD` - Current working directory
- `PATH` - System PATH
- `SHELL` - User's shell

**Cursor-Specific (Unconfirmed):**
- `CURSOR_*` - Potentially Cursor-specific variables
- Testing needed (see gaps.md)

### Testing for Env Vars

```bash
#!/bin/bash
# Log all environment variables
env | grep -i cursor > /tmp/cursor-env.txt
env >> /tmp/all-env.txt
```

### Custom Environment Variables

**Not Supported:** Cannot pass custom env vars via hooks.json

**Workaround:** Use wrapper scripts

```bash
#!/bin/bash
# wrapper.sh
export MY_CUSTOM_VAR="value"
export API_KEY=$(security find-generic-password -s my-service -w)
./actual-hook.sh
```

---

## State Management

### Per-Generation State (Temporary)

**Use Case:** Share data between hooks in same prompt

```bash
# Hook 1 (beforeShellExecution): Write state
generation_id=$(echo "$input" | jq -r '.generation_id')
echo "suspicious" > "/tmp/state-${generation_id}.txt"

# Hook 2 (stop): Read and cleanup
generation_id=$(echo "$input" | jq -r '.generation_id')
if [ -f "/tmp/state-${generation_id}.txt" ]; then
  findings=$(cat "/tmp/state-${generation_id}.txt")
  echo "{\"followup_message\":\"Found: $findings\"}"
  rm "/tmp/state-${generation_id}.txt"
fi
```

### Per-Conversation State (Session)

**Use Case:** Track data across multiple prompts

```bash
# Append to conversation log
conversation_id=$(echo "$input" | jq -r '.conversation_id')
echo "$(date -Iseconds): $event_data" >> "/tmp/conv-${conversation_id}.log"
```

### Persistent State (Global)

**Use Case:** Track data across all conversations

```bash
# Global audit log
echo "$(date -Iseconds): $event_data" >> /tmp/cursor-global-audit.log
```

---

## JSON Parsing Examples

### Bash (jq)

```bash
#!/bin/bash
input=$(cat)

# Extract simple field
command=$(echo "$input" | jq -r '.command')

# Extract nested field
path=$(echo "$input" | jq -r '.arguments.path')

# Extract array elements
workspace_roots=$(echo "$input" | jq -r '.workspace_roots[]')

# Check if field exists
if echo "$input" | jq -e '.command' > /dev/null; then
  echo "Command field exists"
fi

# Conditional extraction
if [ "$(echo "$input" | jq -r '.hook_event_name')" = "beforeShellExecution" ]; then
  command=$(echo "$input" | jq -r '.command')
fi
```

### TypeScript (Bun)

```typescript
import type { BeforeShellExecutionPayload } from "cursor-hooks";

const input: BeforeShellExecutionPayload = await Bun.stdin.json();

// Access fields with type safety
console.log(input.command);
console.log(input.conversation_id);
console.log(input.workspace_roots[0]);

// Type guards
if ("command" in input) {
  console.log("This is a beforeShellExecution event");
}
```

### Python

```python
import sys
import json

# Read JSON from stdin
input_data = json.load(sys.stdin)

# Extract fields
command = input_data.get("command")
conversation_id = input_data.get("conversation_id")
workspace_roots = input_data.get("workspace_roots", [])

# Check field existence
if "command" in input_data:
    print("Command field exists")
```

---

## Output JSON Examples

### Permission-Based Response

```bash
#!/bin/bash
cat << 'JSON'
{
  "permission": "allow",
  "userMessage": "Command approved",
  "agentMessage": "No security issues detected"
}
JSON
```

### Continue Response

```bash
#!/bin/bash
echo '{"continue":true}'
```

### Followup Message

```bash
#!/bin/bash
echo '{"followup_message":"Session completed successfully"}'
```

---

**Sources:**
- [Cursor Official Documentation](https://cursor.com/docs/agent/hooks)
- [TypeScript Types](https://github.com/johnlindquist/cursor-hooks)
- [Python Models](https://github.com/DevonFulcher/py-cursor-hooks)
- [GitButler Deep Dive](https://blog.gitbutler.com/cursor-hooks-deep-dive)
