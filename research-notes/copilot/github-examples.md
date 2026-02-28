# GitHub Copilot - Real-World Examples & Patterns

**Research Date:** 2026-02-21

## Official Examples Repository

### github/awesome-copilot

**Repository:** [github/awesome-copilot](https://github.com/github/awesome-copilot)

**Description:** Community-contributed instructions, prompts, and configurations to help you make the most of GitHub Copilot.

**Contents:**
- Hook examples and templates
- Best practices
- Configuration patterns
- Community contributions

### Hooks Directory Structure

**Location:** `hooks/` directory in awesome-copilot

**Organization:**
```
hooks/
├── hook-name/
│   ├── README.md         # Hook documentation
│   └── hooks.json        # Hook configuration
├── another-hook/
│   ├── README.md
│   └── hooks.json
...
```

**Documentation:** [awesome-copilot/docs/README.hooks.md](https://github.com/github/awesome-copilot/blob/main/docs/README.hooks.md)

## Hook Configuration Examples

### Example 1: Session Logging

**Purpose:** Record session initiation timestamps for auditing

**Hook Type:** `sessionStart`

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{
      "type": "command",
      "bash": "echo \"Session started: $(date)\" >> session-log.txt",
      "powershell": "Add-Content -Path session-log.txt -Value \"Session started: $(Get-Date)\"",
      "cwd": ".",
      "timeoutSec": 10
    }]
  }
}
```

**Use Case:** Audit trail for compliance and usage tracking

---

### Example 2: User Prompt Logging

**Purpose:** Log user requests for auditing and usage analysis

**Hook Type:** `userPromptSubmitted`

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "userPromptSubmitted": [{
      "type": "command",
      "bash": "./scripts/log-prompt.sh",
      "env": {
        "LOG_LEVEL": "info"
      },
      "timeoutSec": 5
    }]
  }
}
```

**Script (`log-prompt.sh`):**
```bash
#!/bin/bash
INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
PROMPT=$(echo "$INPUT" | jq -r '.prompt')
CWD=$(echo "$INPUT" | jq -r '.cwd')

echo "{\"timestamp\":$TIMESTAMP,\"prompt\":\"$PROMPT\",\"cwd\":\"$CWD\"}" >> prompts.jsonl
```

**Use Case:** Build analytics on prompt patterns and user behavior

---

### Example 3: Security - Blocking Dangerous Commands

**Purpose:** Block dangerous system commands before execution

**Hook Type:** `preToolUse`

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/security-check.sh",
      "timeoutSec": 5
    }]
  }
}
```

**Script (`security-check.sh`):**
```bash
#!/bin/bash
set -e

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')

# Only validate bash commands
if [ "$TOOL_NAME" != "bash" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi

# Extract command from toolArgs
COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command')

# Block dangerous patterns
if echo "$COMMAND" | grep -qE "(rm -rf|sudo|mkfs|dd if=|format|curl.*\|.*bash|wget.*\|.*sh)"; then
  echo '{
    "permissionDecision": "deny",
    "permissionDecisionReason": "Dangerous system command detected. This operation requires manual approval."
  }' | jq -c
  exit 0
fi

# Allow safe commands
echo '{"permissionDecision":"allow"}' | jq -c
```

**Blocked Patterns:**
- `rm -rf` - Recursive deletion
- `sudo` - Elevated privileges
- `mkfs` - Filesystem formatting
- `dd if=` - Direct disk writes
- `format` - Drive formatting
- `curl ... | bash` - Download and execute
- `wget ... | sh` - Download and execute

**Use Case:** Prevent accidental destructive operations

---

### Example 4: Tool Usage Tracking

**Purpose:** Track which tools are used and how often

**Hook Type:** `postToolUse`

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "postToolUse": [{
      "type": "command",
      "bash": "./scripts/track-usage.sh",
      "timeoutSec": 5
    }]
  }
}
```

**Script (`track-usage.sh`):**
```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
RESULT_TYPE=$(echo "$INPUT" | jq -r '.toolResult.resultType')

# Append to CSV for analytics
echo "$TIMESTAMP,$TOOL_NAME,$RESULT_TYPE" >> tool-usage.csv

# Allow continuation (postToolUse doesn't affect execution)
exit 0
```

**Output Format (CSV):**
```csv
timestamp,tool_name,result_type
1708560000000,bash,success
1708560005000,edit,success
1708560010000,view,success
1708560015000,bash,failure
```

**Use Case:** Analytics on tool usage patterns, success rates, and performance

---

### Example 5: Multi-Hook Security + Logging

**Purpose:** Combine security enforcement with comprehensive logging

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "bash": "./scripts/security-check.sh",
        "timeoutSec": 5
      },
      {
        "type": "command",
        "bash": "./scripts/log-tool-request.sh",
        "timeoutSec": 3
      }
    ],
    "postToolUse": [{
      "type": "command",
      "bash": "./scripts/log-tool-result.sh",
      "timeoutSec": 3
    }]
  }
}
```

**Execution Flow:**
1. Security check runs first
2. If allowed, logging hook runs
3. Tool executes (if not denied)
4. Result logging hook runs

**Use Case:** Comprehensive security and audit trail

---

### Example 6: External Integration - Slack Notifications

**Purpose:** Send notifications to Slack on errors

**Hook Type:** `errorOccurred`

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "errorOccurred": [{
      "type": "command",
      "bash": "./scripts/notify-slack.sh",
      "env": {
        "SLACK_WEBHOOK_URL": "${SLACK_WEBHOOK_URL}"
      },
      "timeoutSec": 10
    }]
  }
}
```

**Script (`notify-slack.sh`):**
```bash
#!/bin/bash
INPUT=$(cat)
ERROR_MESSAGE=$(echo "$INPUT" | jq -r '.error.message')
ERROR_NAME=$(echo "$INPUT" | jq -r '.error.name')
CWD=$(echo "$INPUT" | jq -r '.cwd')

PAYLOAD=$(jq -n \
  --arg text "🚨 GitHub Copilot Error" \
  --arg error "$ERROR_NAME: $ERROR_MESSAGE" \
  --arg cwd "$CWD" \
  '{
    "text": $text,
    "blocks": [
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*Error:* `\($error)`\n*Location:* `\($cwd)`"
        }
      }
    ]
  }')

curl -X POST "$SLACK_WEBHOOK_URL" \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD"
```

**Use Case:** Real-time error monitoring and team notifications

---

### Example 7: Session Cleanup

**Purpose:** Clean up temporary resources when session ends

**Hook Type:** `sessionEnd`

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "sessionEnd": [{
      "type": "command",
      "bash": "./scripts/cleanup.sh",
      "timeoutSec": 30
    }]
  }
}
```

**Script (`cleanup.sh`):**
```bash
#!/bin/bash
INPUT=$(cat)
REASON=$(echo "$INPUT" | jq -r '.reason')
CWD=$(echo "$INPUT" | jq -r '.cwd')

# Generate session report
echo "Session ended: $REASON at $(date)" >> session-report.txt

# Clean up temp files
find "$CWD" -name "*.tmp" -mmin +60 -delete

# Archive logs
if [ -f "session-log.txt" ]; then
  mv session-log.txt "logs/session-$(date +%s).txt"
fi
```

**Use Case:** Resource management and log archival

---

## Testing Patterns

### Local Hook Testing

**Pattern:** Test hooks locally before deployment

```bash
# Create test input JSON
cat > test-input.json <<EOF
{
  "timestamp": 1708560000000,
  "cwd": "/workspace",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"rm -rf /\"}"
}
EOF

# Test the hook
cat test-input.json | ./scripts/security-check.sh

# Expected output for blocked command:
# {"permissionDecision":"deny","permissionDecisionReason":"Dangerous system command detected..."}
```

**Best Practice:** Always test with various inputs before committing to repository

---

### Gradual Rollout Pattern

**Phase 1: Logging Only** (Weeks 1-2)
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/log-only.sh"
    }]
  }
}
```

Script returns `{"permissionDecision":"allow"}` for all commands but logs everything.

**Phase 2: Review and Analysis** (Weeks 3-4)
- Analyze logs to understand usage patterns
- Identify false positives
- Refine security rules

**Phase 3: Enforcement** (Week 5+)
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/security-check.sh"
    }]
  }
}
```

Now denies dangerous commands based on learned patterns.

---

## Repository Configuration Patterns

### Monorepo Pattern

**Challenge:** Different teams/projects in same repository need different hooks

**Solution:** Use `cwd` to scope hook behavior

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/context-aware-security.sh",
      "timeoutSec": 5
    }]
  }
}
```

**Script:**
```bash
#!/bin/bash
INPUT=$(cat)
CWD=$(echo "$INPUT" | jq -r '.cwd')

# Apply different rules based on directory
if echo "$CWD" | grep -q "/backend/"; then
  # Stricter rules for backend
  # ... security checks ...
elif echo "$CWD" | grep -q "/frontend/"; then
  # Different rules for frontend
  # ... different checks ...
fi
```

---

### CI/CD Integration Pattern

**Goal:** Different behavior in CI vs local development

```bash
#!/bin/bash
INPUT=$(cat)

# Detect CI environment
if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ]; then
  # Stricter enforcement in CI
  ENFORCE=true
else
  # Warning only in local dev
  ENFORCE=false
fi

# ... security checks ...

if [ "$DANGEROUS" = true ]; then
  if [ "$ENFORCE" = true ]; then
    echo '{"permissionDecision":"deny","permissionDecisionReason":"CI: Dangerous command blocked"}' | jq -c
  else
    echo '{"permissionDecision":"allow"}' | jq -c
    echo "WARNING: Dangerous command detected but allowed in dev" >&2
  fi
fi
```

---

## Performance Optimization Patterns

### Timeout Strategy

**Pattern:** Set appropriate timeouts based on hook complexity

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "bash": "./scripts/fast-check.sh",
        "timeoutSec": 3
      }
    ],
    "sessionEnd": [
      {
        "type": "command",
        "bash": "./scripts/slow-cleanup.sh",
        "timeoutSec": 60
      }
    ]
  }
}
```

**Guideline:**
- Fast checks (security): 3-5 seconds
- Logging: 2-3 seconds
- Cleanup: 30-60 seconds
- External API calls: 10-15 seconds

---

### Caching Pattern

**Pattern:** Cache expensive validations

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND_HASH=$(echo "$INPUT" | jq -r '.toolArgs.command' | sha256sum | cut -d' ' -f1)
CACHE_FILE="/tmp/copilot-hook-cache-$COMMAND_HASH"

# Check cache
if [ -f "$CACHE_FILE" ]; then
  CACHED_AGE=$(($(date +%s) - $(stat -f %m "$CACHE_FILE")))
  if [ $CACHED_AGE -lt 300 ]; then  # 5 minute cache
    cat "$CACHE_FILE"
    exit 0
  fi
fi

# Perform expensive check
RESULT=$(expensive_validation "$INPUT")

# Cache result
echo "$RESULT" | tee "$CACHE_FILE" | jq -c
```

---

## Documentation from Examples

All examples follow the **GitHub Copilot hooks specification**:

**Official Reference:** [Hooks configuration](https://docs.github.com/en/copilot/reference/hooks-configuration)

**Key Patterns Demonstrated:**
1. ✅ JSON input via stdin
2. ✅ JSON output via stdout
3. ✅ Exit code 0 for success
4. ✅ Exit code 2 for blocking errors
5. ✅ Proper timeout handling
6. ✅ Environment variable usage
7. ✅ Platform-specific commands
8. ✅ Multiple hooks per event

---

## Community Examples

**Resource:** [GitHub Copilot awesome-copilot](https://github.com/github/awesome-copilot)

**Contributions Welcome:** The awesome-copilot repository accepts community contributions for:
- New hook patterns
- Best practices
- Integration examples
- Use case documentation

---

## Summary

### Most Common Hook Patterns

1. **Security Enforcement** - `preToolUse` with deny logic
2. **Audit Logging** - `userPromptSubmitted` + `postToolUse`
3. **Session Management** - `sessionStart` + `sessionEnd`
4. **External Integration** - `errorOccurred` with webhooks
5. **Usage Analytics** - `postToolUse` with metrics

### Best Practices from Examples

- ✅ Test locally with sample inputs
- ✅ Start with logging, add enforcement gradually
- ✅ Use appropriate timeouts
- ✅ Validate and sanitize all inputs
- ✅ Use structured logging (JSON Lines)
- ✅ Provide clear denial reasons
- ✅ Cache expensive operations
- ✅ Handle both bash and PowerShell
- ✅ Document your hooks thoroughly

**Key Takeaway:** GitHub Copilot provides extensive real-world examples demonstrating practical automation patterns for security, compliance, and workflow customization.
