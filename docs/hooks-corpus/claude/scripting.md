# Claude Code Hooks - Scripting & Execution

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete

---

## Table of Contents

1. [Supported Languages](#supported-languages)
2. [Exit Code System](#exit-code-system)
3. [JSON Output Format](#json-output-format)
4. [Timeout Handling](#timeout-handling)
5. [Error Handling](#error-handling)
6. [Async Execution](#async-execution)
7. [Language-Specific Examples](#language-specific-examples)

---

## Supported Languages

### Any Language with Shebang

Claude Code hooks can be written in **any language** that supports the shebang (`#!`) convention.

**Requirements:**
- Executable permission (`chmod +x`)
- Valid shebang line
- Reads JSON from stdin
- Writes to stdout/stderr
- Returns appropriate exit code

**Supported:**
- Bash/Shell scripts (`#!/bin/bash`, `#!/bin/sh`, `#!/usr/bin/env zsh`)
- Python (`#!/usr/bin/env python3`)
- Node.js (`#!/usr/bin/env node`)
- Ruby (`#!/usr/bin/env ruby`)
- Perl (`#!/usr/bin/env perl`)
- Go (compiled binaries)
- Rust (compiled binaries)
- Any executable

**Source:** [Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

---

### Inline Commands

Simple hooks can use inline shell commands:

```json
{
  "type": "command",
  "command": "echo 'Hook triggered!' >&2"
}
```

**Use for:**
- Simple one-liners
- Quick notifications
- Basic logging

**Avoid for:**
- Complex logic
- Multi-step processes
- JSON output generation

---

## Exit Code System

### Exit Code Meanings

| Exit Code | Meaning | Effect | stderr Handling |
|-----------|---------|--------|-----------------|
| `0` | Success | Parse stdout for JSON or add text to context | Ignored (unless verbose mode) |
| `2` | Blocking error | Prevent action (if event is cancellable) | Fed to Claude or user as error |
| Other | Non-blocking error | Log error, continue execution | Logged in debug mode only |

**Source:** [Official Documentation](https://code.claude.com/docs/en/hooks)

---

### Exit Code 0 (Success)

**Behavior:**
- Stdout parsed for JSON decision
- If not valid JSON, stdout added to context as plain text
- stderr ignored (unless verbose mode)
- Execution continues normally

**Example:**
```bash
#!/bin/bash
INPUT=$(cat)
echo "Processing hook..." >&2
echo "Additional context for Claude"
exit 0
```

---

### Exit Code 2 (Blocking Error)

**Behavior:**
- **If event is cancellable:** Action blocked, stderr becomes error message
- **If event is non-cancellable:** stderr shown to Claude, execution continues
- Claude or user sees stderr content

**Example:**
```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

if echo "$COMMAND" | grep -q "rm -rf"; then
  echo "Dangerous command blocked: $COMMAND" >&2
  exit 2
fi

exit 0
```

**Source:** [Hooks Guide - Exit Codes](https://code.claude.com/docs/en/hooks-guide#exit-codes)

---

### Other Exit Codes (Non-Blocking Error)

**Behavior:**
- Logged in debug mode only
- Execution continues as if hook succeeded
- stderr not shown to user or Claude
- Use for non-critical failures

**Example:**
```bash
#!/bin/bash
# Optional notification hook
curl -X POST "$SLACK_WEBHOOK" -d "$DATA" || exit 1
exit 0
```

---

### Event-Specific Exit Code Effects

| Event | Exit 0 | Exit 2 | Other |
|-------|--------|--------|-------|
| PreToolUse | Continue, parse JSON | Block tool call | Continue, log error |
| PermissionRequest | Continue, parse JSON | Deny permission | Continue, log error |
| UserPromptSubmit | Continue, parse JSON | Block prompt, erase from context | Continue, log error |
| PostToolUse | Continue, parse JSON | Show stderr to Claude | Continue, log error |
| Stop | Allow stop | Prevent stop, continue working | Allow stop |
| SubagentStop | Allow stop | Prevent stop | Allow stop |
| TeammateIdle | Allow idle | Prevent idle, feed stderr to teammate | Allow idle |
| TaskCompleted | Allow completion | Prevent completion, feed stderr to model | Allow completion |
| ConfigChange | Allow change | Block change (except policy) | Allow change |
| WorktreeCreate | Use printed path | Fail creation | Fail creation |
| SessionStart | Add context | Show stderr to user | Log only |
| SessionEnd | Continue | Continue | Continue |
| PostToolUseFailure | Add context | Show stderr to Claude | Log only |
| Notification | Add context | Show stderr to user | Log only |
| SubagentStart | Add context | Show stderr to user | Log only |
| PreCompact | Continue | Continue | Continue |
| WorktreeRemove | Continue | Log only | Log only |

---

## JSON Output Format

### Universal Fields

Available to all hooks when exiting with code 0:

```json
{
  "continue": true,          // default: true
  "stopReason": "Message",   // shown to user if continue=false
  "suppressOutput": false,   // default: false, hides hook output
  "systemMessage": "Warning" // system-level warning message
}
```

**Source:** [Official Hooks Reference](https://code.claude.com/docs/en/hooks)

---

### Decision Control Patterns

#### Pattern 1: Top-Level Decision

**Events:** UserPromptSubmit, PostToolUse, PostToolUseFailure, Stop, SubagentStop, ConfigChange

```json
{
  "decision": "block",
  "reason": "Explanation for Claude or user"
}
```

**Example:**
```bash
#!/bin/bash
INPUT=$(cat)
# Check something...
if [[ condition ]]; then
  jq -n '{
    decision: "block",
    reason: "Tests are failing. Please fix before stopping."
  }'
  exit 0
fi
exit 0
```

---

#### Pattern 2: PreToolUse Decision

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|deny|ask",
    "permissionDecisionReason": "Explanation",
    "updatedInput": {
      "field": "new_value"
    },
    "additionalContext": "Context for Claude"
  }
}
```

**Example:**
```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

if echo "$COMMAND" | grep -qE "rm -rf"; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Dangerous rm -rf command blocked"
    }
  }'
  exit 0
fi

# Allow
exit 0
```

---

#### Pattern 3: PermissionRequest Decision

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PermissionRequest",
    "decision": {
      "behavior": "allow|deny",
      "updatedInput": {
        "field": "modified_value"
      },
      "updatedPermissions": {},
      "message": "Denial reason for Claude",
      "interrupt": false
    }
  }
}
```

**Example:**
```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

# Auto-approve safe npm commands
if echo "$COMMAND" | grep -q "^npm \\(test\\|run\\|install\\)"; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PermissionRequest",
      decision: {
        behavior: "allow"
      }
    }
  }'
  exit 0
fi

# Ask user for everything else
exit 0
```

---

#### Pattern 4: Context Injection

**Events:** SessionStart, UserPromptSubmit, PostToolUse, SubagentStart

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Text added to Claude's context"
  }
}
```

**Or simply output plain text:**
```bash
#!/bin/bash
echo "## Current Project State"
echo ""
echo "### Git Status"
git status --short
exit 0
```

Plain text automatically added to context.

---

### Prompt/Agent Hook Responses

**For hooks with `type: "prompt"` or `type: "agent"`:**

```json
{
  "ok": true,
  "reason": "Required when ok is false"
}
```

**Example:**
```python
#!/usr/bin/env python3
import json
import sys

# Read event
event = json.load(sys.stdin)

# Evaluate
is_safe = evaluate_command(event['tool_input']['command'])

# Respond
response = {
    "ok": is_safe,
    "reason": "Command contains dangerous pattern" if not is_safe else None
}

print(json.dumps(response))
sys.exit(0)
```

**Source:** [Hooks Guide - Prompt Hooks](https://code.claude.com/docs/en/hooks-guide#prompt-hooks)

---

## Timeout Handling

### Default Timeouts

| Hook Type | Default Timeout | Configurable |
|-----------|----------------|--------------|
| Command | 600 seconds (10 minutes) | Yes |
| Prompt | 30 seconds | Yes |
| Agent | 60 seconds | Yes |

**Source:** [Hooks Configuration](https://code.claude.com/docs/en/hooks)

---

### Custom Timeouts

```json
{
  "type": "command",
  "command": "/path/to/slow-script.sh",
  "timeout": 300  // 5 minutes
}
```

**Best Practices:**
- Set conservative timeouts for critical hooks
- Use aggressive timeouts for non-critical hooks
- Consider async for slow operations

---

### Timeout Behavior

**When timeout occurs:**
- Process is killed (SIGTERM)
- Treated as non-blocking error (exit code != 0, != 2)
- stderr logged in debug mode only
- Execution continues

**No retry:** Timed-out hooks are not retried.

---

## Error Handling

### stdout vs stderr

**stdout:**
- Used for JSON decisions
- Used for context injection (plain text)
- Parsed and processed by Claude Code

**stderr:**
- Used for error messages
- Used for debug output
- Behavior depends on exit code:
  - Exit 0: Ignored (unless verbose)
  - Exit 2: Fed to Claude or user
  - Other: Logged in debug only

**Best Practice:**
```bash
#!/bin/bash
# Debug/progress to stderr
echo "Processing hook..." >&2

# JSON/context to stdout
echo "Additional context for Claude"
```

---

### Error Message Guidelines

**Good error messages (stderr with exit 2):**
```bash
echo "Cannot edit files on protected branch 'main'" >&2
echo "Create a feature branch: git checkout -b feature/your-name" >&2
exit 2
```

**Bad error messages:**
```bash
echo "Error" >&2  # Too vague
exit 2
```

**Tips:**
- Be specific about what failed
- Provide actionable suggestions
- Include relevant context (file name, branch, command)

---

### Graceful Degradation

```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)

# Try to format, but don't fail if formatter missing
if command -v prettier &>/dev/null; then
  FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
  if [[ -n "$FILE" && -f "$FILE" ]]; then
    prettier --write "$FILE" 2>&1 >&2 || echo "Formatting failed" >&2
  fi
else
  echo "Prettier not installed, skipping format" >&2
fi

exit 0  # Always succeed
```

---

## Async Execution

### Async Command Hooks

**Enable with `"async": true`:**

```json
{
  "type": "command",
  "command": "/path/to/slow-notification.sh",
  "async": true,
  "timeout": 300
}
```

**Behavior:**
- Hook spawned in background
- Claude Code continues immediately
- Output delivered on next conversation turn
- Cannot block actions
- Decision fields ignored

**Use cases:**
- Notifications (Slack, email)
- Logging
- Analytics
- Long-running non-critical tasks

**Limitations:**
- No guaranteed delivery
- No deduplication (may run multiple times for same event)
- Cannot block or modify tool execution

**Source:** [Hooks Guide - Async](https://code.claude.com/docs/en/hooks-guide#async)

---

### Async Best Practices

**1. Make idempotent:**
```bash
#!/bin/bash
# Async notification hook
LOG_FILE="$HOME/.claude/notifications.log"
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") Event: $EVENT" >> "$LOG_FILE"
curl -X POST "$SLACK_WEBHOOK" -d "$DATA" || true
exit 0
```

**2. Handle failures gracefully:**
```bash
#!/bin/bash
# Don't exit 2 in async hooks (ignored anyway)
curl -X POST "$ENDPOINT" -d "$DATA" || {
  echo "Failed to send notification" >> /tmp/hook-errors.log
  exit 1
}
exit 0
```

**3. Use timeouts:**
```json
{
  "async": true,
  "timeout": 60  // Kill after 1 minute
}
```

---

## Language-Specific Examples

### Bash

```bash
#!/bin/bash
set -euo pipefail

# Read JSON input
INPUT=$(cat)

# Extract fields
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Logic
if [[ "$FILE" == *.py ]]; then
  black "$FILE"
  isort "$FILE"
  echo "Formatted Python file: $FILE" >&2
fi

# Output JSON decision (optional)
jq -n '{
  hookSpecificOutput: {
    hookEventName: "PostToolUse",
    additionalContext: "Auto-formatted Python files"
  }
}'

exit 0
```

---

### Python

```python
#!/usr/bin/env python3
import json
import sys
import subprocess

# Read JSON from stdin
event = json.load(sys.stdin)

# Extract fields
tool_name = event.get('tool_name', '')
file_path = event.get('tool_input', {}).get('file_path', '')

# Logic
if file_path.endswith('.py'):
    try:
        subprocess.run(['black', file_path], check=True, capture_output=True)
        subprocess.run(['isort', file_path], check=True, capture_output=True)
        print(f"Formatted {file_path}", file=sys.stderr)
    except subprocess.CalledProcessError as e:
        print(f"Formatting failed: {e}", file=sys.stderr)
        sys.exit(1)

# Output JSON
response = {
    "hookSpecificOutput": {
        "hookEventName": "PostToolUse",
        "additionalContext": "Auto-formatted Python file"
    }
}
print(json.dumps(response))
sys.exit(0)
```

---

### Node.js

```javascript
#!/usr/bin/env node
const fs = require('fs');
const { execSync } = require('child_process');

// Read JSON from stdin
const input = fs.readFileSync(0, 'utf-8');
const event = JSON.parse(input);

// Extract fields
const toolName = event.tool_name || '';
const filePath = event.tool_input?.file_path || '';

// Logic
if (filePath.endsWith('.js') || filePath.endsWith('.ts')) {
  try {
    execSync(`prettier --write "${filePath}"`, { stdio: 'pipe' });
    console.error(`Formatted ${filePath}`);
  } catch (err) {
    console.error(`Formatting failed: ${err.message}`);
    process.exit(1);
  }
}

// Output JSON
const response = {
  hookSpecificOutput: {
    hookEventName: 'PostToolUse',
    additionalContext: 'Auto-formatted JavaScript/TypeScript file'
  }
};
console.log(JSON.stringify(response));
process.exit(0);
```

---

### Ruby

```ruby
#!/usr/bin/env ruby
require 'json'

# Read JSON from stdin
input = JSON.parse($stdin.read)

# Extract fields
tool_name = input['tool_name'] || ''
file_path = input.dig('tool_input', 'file_path') || ''

# Logic
if file_path.end_with?('.rb')
  begin
    `rubocop -a "#{file_path}"`
    $stderr.puts "Formatted #{file_path}"
  rescue => e
    $stderr.puts "Formatting failed: #{e.message}"
    exit 1
  end
end

# Output JSON
response = {
  hookSpecificOutput: {
    hookEventName: 'PostToolUse',
    additionalContext: 'Auto-formatted Ruby file'
  }
}
puts JSON.generate(response)
exit 0
```

---

### Go (Compiled Binary)

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strings"
)

type Event struct {
    ToolName  string                 `json:"tool_name"`
    ToolInput map[string]interface{} `json:"tool_input"`
}

type Response struct {
    HookSpecificOutput struct {
        HookEventName     string `json:"hookEventName"`
        AdditionalContext string `json:"additionalContext"`
    } `json:"hookSpecificOutput"`
}

func main() {
    // Read JSON from stdin
    inputBytes, _ := io.ReadAll(os.Stdin)
    var event Event
    json.Unmarshal(inputBytes, &event)

    // Extract fields
    filePath, _ := event.ToolInput["file_path"].(string)

    // Logic
    if strings.HasSuffix(filePath, ".go") {
        cmd := exec.Command("gofmt", "-w", filePath)
        if err := cmd.Run(); err != nil {
            fmt.Fprintf(os.Stderr, "Formatting failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Fprintf(os.Stderr, "Formatted %s\n", filePath)
    }

    // Output JSON
    var response Response
    response.HookSpecificOutput.HookEventName = "PostToolUse"
    response.HookSpecificOutput.AdditionalContext = "Auto-formatted Go file"

    output, _ := json.Marshal(response)
    fmt.Println(string(output))
    os.Exit(0)
}
```

Compile: `go build -o format-go hook.go`

---

## Best Practices

### 1. Use Strict Error Handling

```bash
#!/bin/bash
set -euo pipefail  # Exit on error, undefined var, pipe failure
```

### 2. Validate Input

```bash
#!/bin/bash
INPUT=$(cat)

# Validate JSON
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON input" >&2
  exit 1
fi

# Extract with defaults
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
if [[ -z "$FILE" ]]; then
  exit 0  # No file, nothing to do
fi
```

### 3. Quote Variables

```bash
# Bad
rm $FILE

# Good
rm "$FILE"
```

### 4. Separate Debug from Output

```bash
#!/bin/bash
# Debug to stderr
echo "Processing file..." >&2

# Output to stdout
echo "Additional context"
```

### 5. Make Idempotent

```bash
#!/bin/bash
# Can run multiple times safely
if [[ ! -f "$FILE.bak" ]]; then
  cp "$FILE" "$FILE.bak"
fi
```

### 6. Use Timeouts Wisely

```json
{
  "type": "command",
  "command": "/path/to/quick-check.sh",
  "timeout": 5  // Fail fast
}
```

---

## See Also

- [Events Reference](./events-reference.md) - Complete event catalog
- [Configuration](./configuration.md) - Settings format
- [Examples](./examples.md) - Ready-to-use patterns
- [Security](./security.md) - Security best practices

---

**Sources:**
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)
- [GitHub Examples](https://github.com/disler/claude-code-hooks-mastery)

**Document Version:** 1.0
**Research Date:** 2026-02-21
