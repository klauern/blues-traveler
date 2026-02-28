# GitHub Copilot Hooks - Common Patterns & Recipes

**Version:** 1.0
**Last Updated:** 2026-02-21

---

## Quick Reference

| Pattern | Hook Used | Complexity | Use Case |
|---------|-----------|------------|----------|
| [Security Enforcement](#security-enforcement) | preToolUse | Medium | Block dangerous commands |
| [Secret Scanning](#secret-scanning) | preToolUse | Medium | Prevent credential leaks |
| [Audit Logging](#audit-logging) | Multiple | Low | Compliance tracking |
| [Session Management](#session-management) | sessionStart/End | Low | Setup/cleanup |
| [External Notifications](#external-notifications) | errorOccurred | Medium | Slack/email alerts |
| [Usage Metrics](#usage-metrics) | postToolUse | Low | Analytics |
| [Approval Workflow](#approval-workflow) | preToolUse | High | Manual approvals |

---

## Security Enforcement

### Block Dangerous Commands

**Hook:** `preToolUse`
**Purpose:** Prevent destructive operations before execution

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/security-check.sh",
      "powershell": "./scripts/security-check.ps1",
      "timeoutSec": 5
    }]
  }
}
```

**Script:** `.github/hooks/scripts/security-check.sh`
```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

# Only validate bash commands
if [ "$TOOL_NAME" != "bash" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi

COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command')

# Dangerous patterns
if echo "$COMMAND" | grep -qE "(rm -rf|sudo|mkfs|dd if=|curl.*\|.*bash|wget.*\|.*sh)"; then
  echo '{
    "permissionDecision": "deny",
    "permissionDecisionReason": "Dangerous system command blocked by security policy. Review and execute manually if needed."
  }' | jq -c
  exit 0
fi

# Allow safe commands
echo '{"permissionDecision":"allow"}' | jq -c
```

**Result:** Commands like `rm -rf /`, `sudo rm`, `curl url | bash` are blocked.

---

## Secret Scanning

### Prevent Credential Leaks

**Hook:** `preToolUse`
**Purpose:** Detect and block operations containing secrets

**Script:** `.github/hooks/scripts/secret-scanner.sh`
```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')

# Secret patterns
PATTERNS=(
  'AKIA[0-9A-Z]{16}'                    # AWS Access Key
  'ghp_[0-9a-zA-Z]{36}'                 # GitHub PAT
  'sk-[0-9a-zA-Z]{48}'                  # OpenAI API Key
  '[0-9]+-[0-9A-Za-z_]{32}'            # Slack Token
  'AIza[0-9A-Za-z\-_]{35}'             # Google API Key
  '[Bb]earer [A-Za-z0-9\-\._~\+\/]+=*' # Bearer token
  'password\s*=\s*["\047][^"\047]+'    # Password assignment
)

for PATTERN in "${PATTERNS[@]}"; do
  if echo "$TOOL_ARGS" | grep -qE "$PATTERN"; then
    echo '{
      "permissionDecision": "deny",
      "permissionDecisionReason": "Potential secret detected. Remove credentials and use environment variables instead."
    }' | jq -c
    exit 0
  fi
done

echo '{"permissionDecision":"allow"}' | jq -c
```

---

## Audit Logging

### Comprehensive Compliance Tracking

**Hooks:** Multiple events
**Purpose:** Generate complete audit trail for compliance

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{
      "bash": "./audit/log-session-start.sh",
      "timeoutSec": 3
    }],
    "userPromptSubmitted": [{
      "bash": "./audit/log-prompt.sh",
      "timeoutSec": 2
    }],
    "preToolUse": [{
      "bash": "./audit/log-tool-request.sh",
      "timeoutSec": 2
    }],
    "postToolUse": [{
      "bash": "./audit/log-tool-result.sh",
      "timeoutSec": 2
    }],
    "sessionEnd": [{
      "bash": "./audit/log-session-end.sh",
      "timeoutSec": 3
    }]
  }
}
```

**Script:** `.github/hooks/audit/log-tool-request.sh`
```bash
#!/bin/bash
INPUT=$(cat)

TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')

# Log to JSON Lines format
echo "{
  \"type\": \"tool_request\",
  \"timestamp\": $TIMESTAMP,
  \"tool\": \"$TOOL_NAME\",
  \"args\": $TOOL_ARGS,
  \"user\": \"$USER\",
  \"cwd\": \"$(pwd)\"
}" | jq -c >> /var/log/copilot-audit.jsonl

exit 0
```

**Query audit logs:**
```bash
# Find all denied operations
jq 'select(.decision == "deny")' /var/log/copilot-audit.jsonl

# Generate CSV report
jq -r '[.timestamp, .type, .user, .tool] | @csv' /var/log/copilot-audit.jsonl > report.csv

# Count operations by tool
jq -r '.tool' /var/log/copilot-audit.jsonl | sort | uniq -c
```

---

## Session Management

### Environment Initialization

**Hook:** `sessionStart`
**Purpose:** Automatically set up development environment

**Script:** `.github/hooks/scripts/session-start.sh`
```bash
#!/bin/bash
INPUT=$(cat)

CWD=$(echo "$INPUT" | jq -r '.cwd')
SOURCE=$(echo "$INPUT" | jq -r '.source')

echo "Initializing session in $CWD..." >&2

# Detect and setup project type
if [ -f "$CWD/package.json" ]; then
  echo "Node.js project detected" >&2
  if [ ! -d "$CWD/node_modules" ]; then
    echo "Installing dependencies..." >&2
    npm install > /dev/null 2>&1 &
  fi
elif [ -f "$CWD/requirements.txt" ]; then
  echo "Python project detected" >&2
  if [ ! -d "$CWD/venv" ]; then
    echo "Creating virtual environment..." >&2
    python -m venv venv
  fi
elif [ -f "$CWD/go.mod" ]; then
  echo "Go project detected" >&2
  go mod download > /dev/null 2>&1 &
fi

# Log session start
echo "{\"type\":\"session_start\",\"timestamp\":$(date +%s)000,\"source\":\"$SOURCE\"}" \
  >> /var/log/copilot-sessions.jsonl

exit 0
```

### Session Cleanup

**Hook:** `sessionEnd`
**Purpose:** Clean up resources and archive logs

**Script:** `.github/hooks/scripts/session-end.sh`
```bash
#!/bin/bash
INPUT=$(cat)

TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
REASON=$(echo "$INPUT" | jq -r '.reason')
CWD=$(echo "$INPUT" | jq -r '.cwd')

echo "Session ended: $REASON" >&2

# Clean temporary files
find "$CWD" -name "*.tmp" -mmin +60 -delete 2>/dev/null
find "$CWD" -name ".DS_Store" -delete 2>/dev/null

# Archive session logs
if [ -f "/var/log/copilot-session.log" ]; then
  gzip /var/log/copilot-session.log
  mv /var/log/copilot-session.log.gz \
    "/var/log/archive/session-$(date +%s).log.gz"
fi

# Log session end
echo "{\"type\":\"session_end\",\"timestamp\":$TIMESTAMP,\"reason\":\"$REASON\"}" \
  >> /var/log/copilot-sessions.jsonl

exit 0
```

---

## External Notifications

### Slack Alerts on Errors

**Hook:** `errorOccurred`
**Purpose:** Send real-time error notifications to Slack

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "errorOccurred": [{
      "bash": "./scripts/notify-slack.sh",
      "env": {
        "SLACK_WEBHOOK_URL": "${SLACK_WEBHOOK_URL}"
      },
      "timeoutSec": 10
    }]
  }
}
```

**Script:** `.github/hooks/scripts/notify-slack.sh`
```bash
#!/bin/bash
INPUT=$(cat)

ERROR_MSG=$(echo "$INPUT" | jq -r '.error.message')
ERROR_NAME=$(echo "$INPUT" | jq -r '.error.name')
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')

# Create Slack payload
PAYLOAD=$(jq -n \
  --arg text "🚨 GitHub Copilot Error" \
  --arg error "$ERROR_NAME" \
  --arg msg "$ERROR_MSG" \
  --arg time "$(date -d @$((TIMESTAMP/1000)) -Iseconds)" \
  '{
    "text": $text,
    "blocks": [
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*Error Type:* `\($error)`\n*Message:* \($msg)\n*Time:* \($time)"
        }
      }
    ]
  }')

# Send to Slack
curl -X POST "$SLACK_WEBHOOK_URL" \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD" \
  --max-time 5 \
  --silent \
  --fail-with-body

exit 0
```

### Email Notifications

**Script:** `.github/hooks/scripts/notify-email.sh`
```bash
#!/bin/bash
INPUT=$(cat)

ERROR_MSG=$(echo "$INPUT" | jq -r '.error.message')
TO_EMAIL="devops@company.com"

# Send email (requires sendmail or similar)
{
  echo "Subject: Copilot Error Alert"
  echo "To: $TO_EMAIL"
  echo ""
  echo "Error occurred in GitHub Copilot:"
  echo ""
  echo "$ERROR_MSG"
} | sendmail -t

exit 0
```

---

## Usage Metrics

### Track Tool Usage

**Hook:** `postToolUse`
**Purpose:** Collect analytics on tool usage patterns

**Script:** `.github/hooks/scripts/track-usage.sh`
```bash
#!/bin/bash
INPUT=$(cat)

TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
RESULT=$(echo "$INPUT" | jq -r '.toolResult.resultType')

# Append to metrics file (JSON Lines)
echo "{
  \"timestamp\": $TIMESTAMP,
  \"tool\": \"$TOOL_NAME\",
  \"result\": \"$RESULT\",
  \"user\": \"$USER\"
}" | jq -c >> /var/log/copilot-metrics.jsonl

# Update daily summary
DATE=$(date +%Y-%m-%d)
echo "$TOOL_NAME,$RESULT" >> "/var/log/metrics-$DATE.csv"

exit 0
```

**Generate reports:**
```bash
# Daily tool usage
jq -r 'select(.timestamp >= '$(date -d "today" +%s)'000) | .tool' \
  /var/log/copilot-metrics.jsonl | sort | uniq -c

# Success rate by tool
jq -r '[.tool, .result] | @csv' /var/log/copilot-metrics.jsonl | \
  awk -F',' '{tools[$1]++; if($2=="success") success[$1]++}
              END {for(t in tools) print t, success[t]/tools[t]*100"%"}'

# Export for analytics
jq -r '[.timestamp, .tool, .result, .user] | @csv' \
  /var/log/copilot-metrics.jsonl > analytics.csv
```

---

## Approval Workflow

### Manual Approval for Sensitive Operations

**Hook:** `preToolUse`
**Purpose:** Require manual approval for high-risk operations

**Note:** GitHub Copilot doesn't natively support "ask" decision, so this implements a polling-based workaround.

**Script:** `.github/hooks/scripts/approval-workflow.sh`
```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')

# Define operations requiring approval
needs_approval() {
  local command=$1
  echo "$command" | grep -qE "(npm publish|git push.*production|kubectl.*production)"
}

if [ "$TOOL_NAME" = "bash" ]; then
  COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command')

  if needs_approval "$COMMAND"; then
    # Create approval request
    APPROVAL_ID=$(uuidgen)
    REQUEST_FILE="/tmp/copilot-approval-$APPROVAL_ID"

    echo "{
      \"id\": \"$APPROVAL_ID\",
      \"command\": \"$COMMAND\",
      \"user\": \"$USER\",
      \"timestamp\": $(date +%s)
    }" > "$REQUEST_FILE"

    # Notify approvers
    echo "Approval required for: $COMMAND" >&2
    echo "Waiting for approval (25 seconds)..." >&2

    # Poll for approval (within hook timeout)
    for i in {1..25}; do
      if [ -f "$REQUEST_FILE.approved" ]; then
        echo "Approved!" >&2
        rm -f "$REQUEST_FILE" "$REQUEST_FILE.approved"
        echo '{"permissionDecision":"allow"}' | jq -c
        exit 0
      fi
      sleep 1
    done

    # Timeout - deny
    rm -f "$REQUEST_FILE"
    echo '{
      "permissionDecision": "deny",
      "permissionDecisionReason": "Approval timeout. Operation requires manual approval."
    }' | jq -c
    exit 0
  fi
fi

echo '{"permissionDecision":"allow"}' | jq -c
```

**Approval helper:**
```bash
#!/bin/bash
# approve.sh - Run this to approve pending requests

REQUEST_FILE="$1"
if [ ! -f "$REQUEST_FILE" ]; then
  echo "No pending request found"
  exit 1
fi

cat "$REQUEST_FILE"
read -p "Approve this operation? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  touch "$REQUEST_FILE.approved"
  echo "Approved"
fi
```

---

## Multi-Hook Patterns

### Layered Security

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "bash": "./security/scan-secrets.sh",
        "timeoutSec": 3
      },
      {
        "bash": "./security/validate-command.sh",
        "timeoutSec": 3
      },
      {
        "bash": "./audit/log-request.sh",
        "timeoutSec": 2
      }
    ]
  }
}
```

**Execution:** Hooks run sequentially. First deny decision wins.

---

## Context-Aware Patterns

### Environment-Specific Rules

**Script:** `.github/hooks/scripts/context-aware-security.sh`
```bash
#!/bin/bash
INPUT=$(cat)

CWD=$(echo "$INPUT" | jq -r '.cwd')
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

# Determine environment
if echo "$CWD" | grep -q "/production/"; then
  ENV="production"
  ALLOWED_TOOLS="view"
elif echo "$CWD" | grep -q "/staging/"; then
  ENV="staging"
  ALLOWED_TOOLS="view|edit"
else
  ENV="development"
  ALLOWED_TOOLS=".*"
fi

echo "Environment: $ENV" >&2

# Apply environment-specific rules
if [ "$ENV" != "development" ]; then
  if ! echo "$TOOL_NAME" | grep -qE "^($ALLOWED_TOOLS)$"; then
    echo "{
      \"permissionDecision\": \"deny\",
      \"permissionDecisionReason\": \"Tool '$TOOL_NAME' not allowed in $ENV environment\"
    }" | jq -c
    exit 0
  fi
fi

echo '{"permissionDecision":"allow"}' | jq -c
```

---

## Testing Patterns

### Local Hook Testing

**Test script:** `test-hooks.sh`
```bash
#!/bin/bash
set -e

echo "Testing GitHub Copilot hooks..."

# Test 1: Dangerous command should be denied
echo -n "Test 1 (dangerous command): "
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"rm -rf /\"}"}' | \
  .github/hooks/scripts/security-check.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "deny"' > /dev/null; then
  echo "✅ PASS"
else
  echo "❌ FAIL"
  exit 1
fi

# Test 2: Safe command should be allowed
echo -n "Test 2 (safe command): "
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"ls -la\"}"}' | \
  .github/hooks/scripts/security-check.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "allow"' > /dev/null; then
  echo "✅ PASS"
else
  echo "❌ FAIL"
  exit 1
fi

# Test 3: Secret should be detected
echo -n "Test 3 (secret detection): "
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"export KEY=AKIAIOSFODNN7EXAMPLE\"}"}' | \
  .github/hooks/scripts/secret-scanner.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "deny"' > /dev/null; then
  echo "✅ PASS"
else
  echo "❌ FAIL"
  exit 1
fi

echo ""
echo "All tests passed! ✅"
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Available hooks
- [Security & Safety](./security.md) - Security best practices
- [Configuration](./configuration.md) - Setup instructions
- [Scripting](./scripting.md) - Writing hooks

---

**Sources:**
- [awesome-copilot](https://github.com/github/awesome-copilot) - Community examples
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [Using Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

---

**Document Version:** 1.0
**Status:** ✅ Complete
