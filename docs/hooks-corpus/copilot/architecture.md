# GitHub Copilot Hooks - Architecture & Internals

**Version:** 1.0
**Last Updated:** 2026-02-21

---

## Execution Architecture

### Overview

GitHub Copilot's hook system executes **shell commands** at strategic points in the agent's workflow. The architecture is designed for:

- **Security** - Hooks can block operations before execution
- **Simplicity** - JSON input/output, exit codes for control
- **Reliability** - Timeouts, error handling, fallbacks
- **Performance** - Fast execution, minimal overhead

### Process Model

```
┌─────────────────────────────────────────────────────────┐
│                  GitHub Copilot Agent                    │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 1. User submits prompt                             │ │
│  └────────────────┬──────────────────────────────────┘ │
│                   │                                     │
│                   ▼                                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 2. Check for userPromptSubmitted hook              │ │
│  │    - Execute hook if configured                    │ │
│  │    - Log output (cannot block)                     │ │
│  └────────────────┬──────────────────────────────────┘ │
│                   │                                     │
│                   ▼                                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 3. Agent generates response with tool calls        │ │
│  └────────────────┬──────────────────────────────────┘ │
│                   │                                     │
│                   ▼                                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 4. For each tool call:                             │ │
│  │    a. Trigger preToolUse hook                      │ │
│  │    b. Execute hook, read decision                  │ │
│  │    c. If "deny": skip tool, inform agent           │ │
│  │    d. If "allow": proceed                          │ │
│  └────────────────┬──────────────────────────────────┘ │
│                   │                                     │
│                   ▼                                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 5. Execute tool (if allowed)                       │ │
│  └────────────────┬──────────────────────────────────┘ │
│                   │                                     │
│                   ▼                                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 6. Trigger postToolUse hook                        │ │
│  │    - Execute hook                                  │ │
│  │    - Log output (cannot modify result)            │ │
│  └────────────────┬──────────────────────────────────┘ │
│                   │                                     │
│                   ▼                                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │ 7. Agent completes response                        │ │
│  │    - Trigger agentStop hook                        │ │
│  └──────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────┘
```

### Hook Execution Flow

**Step-by-step:**

1. **Event triggers** - Hook event occurs (sessionStart, preToolUse, etc.)
2. **Hook lookup** - Check configuration for registered hooks
3. **Environment preparation** - Set up environment variables, working directory
4. **JSON input preparation** - Construct JSON payload with event data
5. **Process spawn** - Execute shell command (bash/PowerShell)
6. **Input delivery** - Pipe JSON to hook's stdin
7. **Hook execution** - Hook processes input, performs logic
8. **Output collection** - Read stdout, stderr
9. **Exit code check** - Determine success/failure/blocking
10. **Decision processing** - Parse JSON output (preToolUse only)
11. **Action** - Continue, block, or error based on result
12. **Cleanup** - Terminate hook process, log execution

---

## Hook Lifecycle

### Lifecycle Stages

**1. Configuration Loading**
- Read `.github/hooks/*.json` from repository default branch
- Merge with workspace/user hooks (if applicable)
- Validate JSON schema
- Register hooks for each event type

**2. Hook Registration**
- Associate hooks with events
- Prepare hook metadata (timeout, environment, cwd)
- Cache hook scripts for performance

**3. Event Detection**
- Monitor agent workflow for hook events
- Trigger registered hooks when events occur
- Queue multiple hooks for sequential execution

**4. Hook Execution**
- Spawn shell process (bash or PowerShell)
- Set working directory and environment
- Pipe JSON input to stdin
- Capture stdout and stderr
- Monitor for timeout

**5. Result Processing**
- Parse exit code
- Parse JSON output (preToolUse only)
- Determine action (allow, deny, error)
- Log execution details

**6. Cleanup**
- Terminate hook process
- Close file descriptors
- Free resources
- Log completion

---

## Performance Considerations

### Execution Model

**Synchronous Execution:**
- Hooks execute **synchronously** - agent waits for completion
- Multiple hooks execute **sequentially**, not in parallel
- Total time = sum of all hook execution times

**Performance Impact:**
```
Agent response time = Agent processing + Σ(Hook execution times)
```

**Example:**
- Agent processing: 2 seconds
- preToolUse hook 1: 1 second
- preToolUse hook 2: 1 second
- postToolUse hook: 0.5 seconds
- **Total: 4.5 seconds**

### Performance Best Practices

**1. Keep Hooks Fast**
```bash
# ❌ Bad - Slow external call
curl -X POST $API_URL -d "$DATA" --max-time 30

# ✅ Good - Fast validation, defer slow work
echo "$DATA" >> queue.txt  # Process later
validate_fast "$DATA"      # Local check only
```

**2. Use Timeouts Appropriately**
```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./quick-check.sh",
      "timeoutSec": 5  // ✅ Short timeout for blocking hook
    }],
    "sessionEnd": [{
      "bash": "./cleanup.sh",
      "timeoutSec": 60  // ✅ Longer for cleanup
    }]
  }
}
```

**3. Cache Expensive Operations**
```bash
#!/bin/bash
# Cache validation results
CACHE_KEY=$(echo "$INPUT" | sha256sum | cut -d' ' -f1)
CACHE_FILE="/tmp/hook-cache-$CACHE_KEY"

if [ -f "$CACHE_FILE" ] && [ $(($(date +%s) - $(stat -f %m "$CACHE_FILE"))) -lt 300 ]; then
  cat "$CACHE_FILE"
  exit 0
fi

# Expensive validation
RESULT=$(expensive_check "$INPUT")
echo "$RESULT" | tee "$CACHE_FILE"
```

**4. Minimize Hook Count**
```json
// ❌ Bad - Many hooks add latency
{
  "preToolUse": [
    {"bash": "./check1.sh"},
    {"bash": "./check2.sh"},
    {"bash": "./check3.sh"},
    {"bash": "./check4.sh"}
  ]
}

// ✅ Good - Single combined hook
{
  "preToolUse": [
    {"bash": "./combined-check.sh"}  // Runs all checks internally
  ]
}
```

**5. Defer Heavy Work**
```bash
#!/bin/bash
# ✅ Log to queue, process asynchronously
echo "$INPUT" | jq -c >> /var/log/copilot-events.jsonl

# Quick decision
echo '{"permissionDecision":"allow"}' | jq -c
exit 0

# Separate process handles heavy lifting:
# tail -f /var/log/copilot-events.jsonl | process-events.sh &
```

---

## Security Model

### Execution Permissions

**Hooks run with:**
- Same user permissions as GitHub Copilot process
- Access to repository files
- Access to system commands
- Environment variables from user session

**No sandboxing:**
- ⚠️ Hooks are **not sandboxed**
- ⚠️ Can execute arbitrary code
- ⚠️ Trust only hooks from trusted sources
- ⚠️ Review hooks carefully before use

### Security Features

**1. Repository Scoping**
- Hooks must be on repository default branch
- Cannot execute hooks from unmerged code
- Prevents malicious PR hook execution

**2. Timeout Protection**
- All hooks have timeout limits
- Prevents runaway processes
- Default: 30 seconds (configurable)

**3. Exit Code Control**
- Exit code 2 = blocking error (stops execution)
- Provides error handling mechanism
- Prevents cascading failures

**4. Built-in Secret Scanning**
- GitHub Copilot has built-in secret detection
- Warns about detected credentials
- Complements hook-based secret scanning

### Security Recommendations

**1. Input Validation**
```bash
#!/bin/bash
INPUT=$(cat)

# ✅ Validate JSON structure
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON input" >&2
  exit 2
fi

# ✅ Sanitize extracted values
COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command // ""')
if [ -z "$COMMAND" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi
```

**2. Avoid Logging Secrets**
```bash
#!/bin/bash
# ❌ Bad - Logs may contain secrets
echo "Command: $COMMAND" >> hook.log

# ✅ Good - Sanitize before logging
SAFE_CMD=$(echo "$COMMAND" | sed -E 's/(password|token|key)=[^ ]+/\1=****/g')
echo "Command: $SAFE_CMD" >> hook.log
```

**3. Escape Shell Commands**
```bash
#!/bin/bash
# ❌ Bad - Shell injection risk
eval "$COMMAND"

# ✅ Good - Proper quoting
echo "$COMMAND" | grep -qE "$PATTERN"
```

**4. Limit Privileges**
```bash
#!/bin/bash
# ✅ Run validation without elevated privileges
# Don't use sudo in hooks unless absolutely necessary

# ❌ Bad
sudo validate-security

# ✅ Good
validate-security  # Run as current user
```

---

## Resource Limits

### Timeout Limits

**Default:** 30 seconds
**Configurable:** 1-90 seconds (practical limit)

```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./hook.sh",
      "timeoutSec": 10
    }]
  }
}
```

**Timeout Behavior:**
- Process is **terminated** after timeout
- Treated as failure (non-blocking warning)
- Error logged to console
- Agent continues execution

### Memory Limits

**No explicit limits** - Hooks inherit process limits from parent

**Best practices:**
- Keep memory usage minimal
- Stream large data instead of loading into memory
- Clean up resources after use

### CPU Limits

**No explicit limits** - Subject to system scheduler

**Best practices:**
- Avoid CPU-intensive operations
- Use efficient algorithms
- Profile and optimize slow hooks

---

## Error Handling

### Hook Failures

**Exit Code Meanings:**

| Exit Code | Meaning | Agent Behavior |
|-----------|---------|----------------|
| 0 | Success | Continue, parse output (preToolUse) |
| 1 | Failure | Log warning, continue (non-blocking) |
| 2 | Blocking error | Stop, report to model, user intervention |
| 3-255 | Failure | Log warning, continue (non-blocking) |

### Error Recovery

**Timeout:**
```
Hook timeout → Log warning → Continue with default behavior
```

**Parse Error (preToolUse):**
```
Invalid JSON output → Log warning → Treat as "allow"
```

**Process Spawn Failure:**
```
Cannot execute hook → Log error → Continue with default behavior
```

### Fallback Behavior

**preToolUse:** Default to **allow** if hook fails
**All others:** Continue normal execution, log error

**Rationale:** Failing open prevents hooks from blocking legitimate work

---

## Platform Differences

### bash vs. PowerShell

**Configuration:**
```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./unix-hook.sh",
      "powershell": "./windows-hook.ps1"
    }]
  }
}
```

**Platform Detection:**
- Automatic based on OS
- Unix/Linux/macOS → bash
- Windows → PowerShell

**Cross-Platform Tips:**

1. **Use jq on both platforms:**
```bash
# bash
FIELD=$(echo "$INPUT" | jq -r '.field')
```
```powershell
# PowerShell
$field = ($input | jq -r '.field')
```

2. **Handle paths correctly:**
```bash
# bash
FILE_PATH="$CWD/file.txt"
```
```powershell
# PowerShell
$FilePath = Join-Path $cwd "file.txt"
```

3. **Consistent JSON output:**
```bash
# bash
echo '{"permissionDecision":"allow"}' | jq -c
```
```powershell
# PowerShell
@{permissionDecision="allow"} | ConvertTo-Json -Compress
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Hook events
- [Scripting & Execution](./scripting.md) - Writing hooks
- [Security & Safety](./security.md) - Security best practices
- [Troubleshooting](./troubleshooting.md) - Debug issues

---

**Sources:**
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [About Hooks - GitHub Copilot](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-hooks)
- [Using Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

---

**Document Version:** 1.0
**Status:** ✅ Complete
