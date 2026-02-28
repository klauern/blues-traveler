# GitHub Copilot Hooks - Event Types & Triggers

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete - All 7 Events Documented

---

## Event Summary Table

| # | Event | Trigger | Timing | Cancellable | Output Format |
|---|-------|---------|--------|-------------|---------------|
| 1 | sessionStart | Session begins/resumes | Session Init | ❌ No | Ignored |
| 2 | sessionEnd | Session terminates | Session Close | ❌ No | Ignored |
| 3 | userPromptSubmitted | User submits prompt | Pre-Processing | ❌ No | Ignored |
| 4 | **preToolUse** | **Before tool execution** | **Pre-Action** | **✅ Yes** | **JSON Decision** |
| 5 | postToolUse | After tool success | Post-Action | ❌ No | Ignored |
| 6 | errorOccurred | Error during operation | On Error | ❌ No | Ignored |
| 7 | agentStop | Agent finishes responding | Post-Response | ❌ No | Ignored |

---

## 1. sessionStart

**Trigger:** When a new GitHub Copilot agent session begins or resumes
**Timing:** Once per session
**Frequency:** Low (session-level)
**Cancellable:** ❌ No

### When It Fires

- New session initialized
- Existing session resumed

### Input Schema

```json
{
  "timestamp": 1708560000000,
  "cwd": "/workspace/project",
  "source": "new",
  "initialPrompt": "User's first prompt (if new session)"
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560000000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |
| `source` | string | "new", "resume", or "startup" | `"new"` |
| `initialPrompt` | string | User's opening prompt (new sessions) | `"implement auth"` |

### Output

Hook output is **ignored**. Use for side effects only (logging, setup).

### Common Use Cases

- **Environment initialization** - Set up development environment
- **Project context loading** - Load git status, recent commits
- **Session logging** - Record session start for audit
- **Resource preparation** - Start services, validate dependencies
- **User notification** - Send session start alerts

### Example Implementation

```bash
#!/bin/bash
# .github/hooks/scripts/session-start.sh

INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
CWD=$(echo "$INPUT" | jq -r '.cwd')
SOURCE=$(echo "$INPUT" | jq -r '.source')

# Log session start
echo "{\"type\":\"session_start\",\"timestamp\":$TIMESTAMP,\"source\":\"$SOURCE\"}" >> audit.jsonl

# Initialize environment based on project type
if [ -f "$CWD/package.json" ]; then
  echo "Node.js project detected" >&2
  [ ! -d "$CWD/node_modules" ] && npm install
elif [ -f "$CWD/requirements.txt" ]; then
  echo "Python project detected" >&2
  [ ! -d "$CWD/venv" ] && python -m venv venv
fi

exit 0
```

**Configuration:**

```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{
      "type": "command",
      "bash": "./scripts/session-start.sh",
      "timeoutSec": 30
    }]
  }
}
```

---

## 2. sessionEnd

**Trigger:** When GitHub Copilot agent session completes or terminates
**Timing:** End of session
**Frequency:** Low (once per session)
**Cancellable:** ❌ No

### When It Fires

- Session completed successfully
- Session aborted
- Session timed out
- User exited

### Input Schema

```json
{
  "timestamp": 1708560100000,
  "cwd": "/workspace/project",
  "reason": "complete"
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560100000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |
| `reason` | string | "complete", "error", "abort", "timeout", "user_exit" | `"complete"` |

### Output

Hook output is **ignored**. Use for cleanup and final logging.

### Common Use Cases

- **Resource cleanup** - Stop services, clean temp files
- **Log archival** - Archive session logs
- **Session reporting** - Generate summary reports
- **Team notification** - Send completion alerts
- **Metrics collection** - Record session statistics

### Example Implementation

```bash
#!/bin/bash
# .github/hooks/scripts/session-end.sh

INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
REASON=$(echo "$INPUT" | jq -r '.reason')
CWD=$(echo "$INPUT" | jq -r '.cwd')

# Log session end
echo "{\"type\":\"session_end\",\"timestamp\":$TIMESTAMP,\"reason\":\"$REASON\"}" >> audit.jsonl

# Clean temporary files
find "$CWD" -name "*.tmp" -mmin +60 -delete

# Archive logs
if [ -f "copilot-session.log" ]; then
  gzip copilot-session.log
  mv copilot-session.log.gz "logs/session-$(date +%s).log.gz"
fi

# Send notification
if [ "$REASON" = "error" ]; then
  curl -X POST "$WEBHOOK_URL" -d "{\"text\":\"Copilot session ended with error\"}"
fi

exit 0
```

---

## 3. userPromptSubmitted

**Trigger:** When user submits input to the agent
**Timing:** After user input, before processing
**Frequency:** High (every user message)
**Cancellable:** ❌ No

### When It Fires

Every time the user submits a prompt to GitHub Copilot.

### Input Schema

```json
{
  "timestamp": 1708560005000,
  "cwd": "/workspace/project",
  "prompt": "User's submitted text"
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560005000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |
| `prompt` | string | Exact user input | `"implement user authentication"` |

### Output

Hook output is **ignored**. Cannot modify or block the prompt.

**Note:** Unlike Claude Code, GitHub Copilot's `userPromptSubmitted` cannot block or modify prompts.

### Common Use Cases

- **Usage tracking** - Log user requests for analytics
- **Compliance audit** - Record all user interactions
- **Pattern analysis** - Analyze prompt patterns
- **Usage statistics** - Track feature usage
- **Team analytics** - Understand how teams use Copilot

### Example Implementation

```bash
#!/bin/bash
# .github/hooks/scripts/log-prompt.sh

INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
PROMPT=$(echo "$INPUT" | jq -r '.prompt')

# Log to JSON Lines for easy processing
echo "{\"type\":\"prompt\",\"timestamp\":$TIMESTAMP,\"prompt\":\"$PROMPT\"}" >> prompts.jsonl

# Count prompts per day
DATE=$(date +%Y-%m-%d)
echo "$DATE" >> prompt-counts.txt

exit 0
```

---

## 4. preToolUse (Most Powerful)

**Trigger:** Before tool invocation, after agent creates tool parameters
**Timing:** Pre-action
**Frequency:** Very high (every tool use)
**Cancellable:** ✅ **Yes - Can block execution**

### When It Fires

Before any tool is executed by the agent:
- bash commands
- file edits
- file creation
- file viewing
- any agent tool

### Input Schema

```json
{
  "timestamp": 1708560010000,
  "cwd": "/workspace/project",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"rm -rf /\"}"
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560010000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |
| `toolName` | string | Tool identifier | `"bash"`, `"edit"`, `"view"`, `"create"` |
| `toolArgs` | string | JSON string of tool parameters | `"{\"command\":\"ls -la\"}"` |

### Output Schema

**Critical:** `preToolUse` can **block execution** by returning a deny decision.

```json
{
  "permissionDecision": "deny",
  "permissionDecisionReason": "Human-readable explanation shown to user"
}
```

**Decision Values:**
- `"allow"` - Permit tool execution (default if not specified)
- `"deny"` - **Block tool execution** and show reason to user
- `"ask"` - Request user approval (not currently implemented)

**Exit Codes:**
- `0` - Success, parse JSON output
- `2` - Blocking error, stop and report to model
- Other - Non-blocking warning, continue

### Tool Types and Arguments

**bash:**
```json
{
  "toolArgs": "{\"command\":\"git status\"}"
}
```

**edit:**
```json
{
  "toolArgs": "{\"path\":\"/path/to/file.js\",\"content\":\"...\"}"
}
```

**view:**
```json
{
  "toolArgs": "{\"path\":\"/path/to/file.js\"}"
}
```

**create:**
```json
{
  "toolArgs": "{\"path\":\"/path/to/new-file.js\",\"content\":\"...\"}"
}
```

### Common Use Cases

- **Security enforcement** - Block dangerous commands
- **Policy compliance** - Enforce coding standards
- **Secret scanning** - Prevent credential leaks
- **Approval workflows** - Require approval for sensitive ops
- **Audit logging** - Log all tool requests

### Example Implementation - Security

```bash
#!/bin/bash
# .github/hooks/scripts/security-check.sh

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')

# Only validate bash commands
if [ "$TOOL_NAME" != "bash" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi

COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command')

# Block dangerous patterns
if echo "$COMMAND" | grep -qE "(rm -rf|sudo|mkfs|dd if=|format)"; then
  echo '{
    "permissionDecision": "deny",
    "permissionDecisionReason": "Dangerous system command blocked. This operation requires manual review."
  }' | jq -c
  exit 0
fi

# Block download-and-execute
if echo "$COMMAND" | grep -qE "(curl.*\|.*bash|wget.*\|.*sh)"; then
  echo '{
    "permissionDecision": "deny",
    "permissionDecisionReason": "Download-and-execute pattern blocked for security."
  }' | jq -c
  exit 0
fi

# Allow safe commands
echo '{"permissionDecision":"allow"}' | jq -c
exit 0
```

---

## 5. postToolUse

**Trigger:** After tool execution completes successfully
**Timing:** Post-action
**Frequency:** Very high (every successful tool use)
**Cancellable:** ❌ No

### When It Fires

After any tool completes successfully.

### Input Schema

```json
{
  "timestamp": 1708560015000,
  "cwd": "/workspace/project",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"ls -la\"}",
  "toolResult": {
    "resultType": "success",
    "textResultForLlm": "Output shown to the model"
  }
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560015000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |
| `toolName` | string | Executed tool name | `"bash"` |
| `toolArgs` | string | Tool parameters (JSON string) | `"{\"command\":\"ls\"}"` |
| `toolResult` | object | Tool execution result | See below |

**toolResult Object:**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `resultType` | string | "success", "failure", or "denied" | `"success"` |
| `textResultForLlm` | string | Output sent to model | `"file1.js\nfile2.js"` |

### Output

Hook output is **ignored**. Cannot modify result.

**Note:** Unlike some systems, you cannot modify tool output. Use for logging only.

### Common Use Cases

- **Usage tracking** - Track which tools are used
- **Performance metrics** - Measure tool execution
- **Success rate tracking** - Monitor success/failure
- **Audit logging** - Comprehensive activity log
- **External notifications** - Alert on specific actions

### Example Implementation

```bash
#!/bin/bash
# .github/hooks/scripts/log-tool-result.sh

INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
RESULT_TYPE=$(echo "$INPUT" | jq -r '.toolResult.resultType')

# Log to JSON Lines
echo "{\"type\":\"tool_result\",\"timestamp\":$TIMESTAMP,\"tool\":\"$TOOL_NAME\",\"result\":\"$RESULT_TYPE\"}" >> tool-usage.jsonl

# Track metrics
if [ "$RESULT_TYPE" = "success" ]; then
  echo "$TOOL_NAME" >> successful-tools.txt
fi

exit 0
```

---

## 6. errorOccurred

**Trigger:** When agent encounters errors during operation
**Timing:** On error
**Frequency:** Low (only on errors)
**Cancellable:** ❌ No

### When It Fires

When GitHub Copilot agent encounters an error during execution.

### Input Schema

```json
{
  "timestamp": 1708560020000,
  "cwd": "/workspace/project",
  "error": {
    "message": "Error description",
    "name": "ErrorType",
    "stack": "Stack trace..."
  }
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560020000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |
| `error` | object | Error details | See below |

**error Object:**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `message` | string | Error message | `"Command failed"` |
| `name` | string | Error type | `"CommandError"` |
| `stack` | string | Stack trace | `"Error: ...\nat ..."` |

### Output

Hook output is **ignored**. Use for notifications and logging.

### Common Use Cases

- **Error logging** - Comprehensive error tracking
- **Team notifications** - Alert team of failures
- **Error analysis** - Track error patterns
- **External monitoring** - Send to monitoring systems
- **Incident response** - Trigger incident workflows

### Example Implementation

```bash
#!/bin/bash
# .github/hooks/scripts/error-handler.sh

INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
ERROR_MSG=$(echo "$INPUT" | jq -r '.error.message')
ERROR_NAME=$(echo "$INPUT" | jq -r '.error.name')

# Log error
echo "{\"type\":\"error\",\"timestamp\":$TIMESTAMP,\"error\":\"$ERROR_NAME: $ERROR_MSG\"}" >> errors.jsonl

# Send Slack notification
PAYLOAD=$(jq -n \
  --arg text "🚨 GitHub Copilot Error" \
  --arg error "$ERROR_NAME: $ERROR_MSG" \
  '{text: $text, blocks: [{type: "section", text: {type: "mrkdwn", text: $error}}]}')

curl -X POST "$SLACK_WEBHOOK_URL" \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD"

exit 0
```

---

## 7. agentStop

**Trigger:** When main agent or subagent finishes responding
**Timing:** Post-response
**Frequency:** Medium (per response)
**Cancellable:** ❌ No

### When It Fires

When the GitHub Copilot agent completes its response.

### Input Schema

```json
{
  "timestamp": 1708560025000,
  "cwd": "/workspace/project"
}
```

### Input Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | number | Unix milliseconds | `1708560025000` |
| `cwd` | string | Current working directory | `"/workspace/project"` |

### Output

Hook output is **ignored**. Use for cleanup and logging.

### Common Use Cases

- **Response logging** - Log agent completions
- **Cleanup** - Clean up response-specific resources
- **Metrics** - Track response times
- **Final validation** - Verify response quality
- **Post-processing** - Trigger follow-up actions

### Example Implementation

```bash
#!/bin/bash
# .github/hooks/scripts/agent-stop.sh

INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')

# Log agent stop
echo "{\"type\":\"agent_stop\",\"timestamp\":$TIMESTAMP}" >> agent-activity.jsonl

# Clean up temporary state
rm -f /tmp/copilot-temp-*

exit 0
```

---

## Hook Communication Protocol

### Input: JSON via stdin

All hooks receive JSON input through **stdin**:

```bash
#!/bin/bash
INPUT=$(cat)
FIELD=$(echo "$INPUT" | jq -r '.fieldName')
```

### Output: JSON via stdout (for preToolUse)

Only `preToolUse` processes stdout. All others ignore it.

**Success:**
```bash
echo '{"permissionDecision":"allow"}' | jq -c
exit 0
```

**Block:**
```bash
echo '{
  "permissionDecision": "deny",
  "permissionDecisionReason": "Explanation shown to user"
}' | jq -c
exit 0
```

### Exit Codes

| Code | Meaning | Effect |
|------|---------|--------|
| `0` | Success | Parse stdout (preToolUse), continue (others) |
| `2` | Blocking error | Stop execution, report to model |
| Other | Non-blocking warning | Log warning, continue |

---

## Event Ordering & Lifecycle

### Typical Session Flow

```
1. sessionStart
   ↓
2. userPromptSubmitted (user: "create a login page")
   ↓
3. preToolUse (create file)
   ↓
4. postToolUse (create succeeded)
   ↓
5. preToolUse (edit file)
   ↓
6. postToolUse (edit succeeded)
   ↓
7. agentStop (response complete)
   ↓
... more interactions ...
   ↓
8. sessionEnd (session terminated)
```

### Error Flow

```
1. sessionStart
   ↓
2. userPromptSubmitted
   ↓
3. preToolUse (bash)
   ↓
4. [Command fails]
   ↓
5. errorOccurred
   ↓
6. agentStop
```

---

## Performance Considerations

### Hook Execution Time

**Recommended Timeouts:**
- `preToolUse` (security): 3-5 seconds
- `postToolUse` (logging): 2-3 seconds
- `sessionStart` (setup): 30-60 seconds
- `sessionEnd` (cleanup): 30-60 seconds
- `errorOccurred`: 10-15 seconds

**Why:** Long-running hooks block agent progress and degrade user experience.

### Optimization Strategies

1. **Keep hooks fast** - Under 5 seconds for preToolUse
2. **Use caching** - Cache expensive validations
3. **Defer heavy work** - Log to queue, process later
4. **Optimize scripts** - Profile and optimize slow operations
5. **Batch processing** - Aggregate logs, process in batches

---

## Advanced Patterns

### Multiple Hooks Per Event

Execute multiple hooks sequentially:

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {"type": "command", "bash": "./security-check.sh"},
      {"type": "command", "bash": "./compliance-check.sh"},
      {"type": "command", "bash": "./audit-log.sh"}
    ]
  }
}
```

**Execution:** Sequential, in order. First deny wins.

### Context-Aware Hooks

Different behavior based on directory:

```bash
#!/bin/bash
INPUT=$(cat)
CWD=$(echo "$INPUT" | jq -r '.cwd')

if echo "$CWD" | grep -q "/production/"; then
  # Strict rules in production
  ENFORCE=true
else
  # Relaxed in dev
  ENFORCE=false
fi
```

### Conditional Execution

Skip expensive checks when not needed:

```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

# Only check bash commands
if [ "$TOOL_NAME" != "bash" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi

# ... expensive validation ...
```

---

## See Also

- [Architecture & Internals](./architecture.md) - How hooks execute
- [Configuration & Setup](./configuration.md) - Hook configuration
- [Common Patterns & Recipes](./examples.md) - Working examples
- [API Reference](./api-reference.md) - Complete schemas
- [Security & Safety](./security.md) - Security best practices

---

**Sources:**
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [About Hooks - GitHub Copilot](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-hooks)
- [Using Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

---

**Document Version:** 1.0
**Coverage:** 7/7 events (100%)
**Status:** ✅ Complete
