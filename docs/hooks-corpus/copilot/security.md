# GitHub Copilot Hooks - Security & Safety

**Version:** 1.0
**Last Updated:** 2026-02-21

---

## Security Enforcement with preToolUse

The `preToolUse` hook is the **primary security enforcement mechanism** in GitHub Copilot. It can **block operations before execution**.

### Basic Security Pattern

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

# Block dangerous operations
if echo "$COMMAND" | grep -qE "(rm -rf|sudo|mkfs|dd if=|format)"; then
  echo '{
    "permissionDecision": "deny",
    "permissionDecisionReason": "Dangerous system command blocked by security policy"
  }' | jq -c
  exit 0
fi

# Allow safe commands
echo '{"permissionDecision":"allow"}' | jq -c
```

---

## Blocking Dangerous Operations

### System Commands to Block

**Destructive commands:**
```bash
# Block patterns
rm -rf          # Recursive deletion
mkfs           # Filesystem formatting
dd if=         # Direct disk writes
format         # Drive formatting
fdisk          # Partition manipulation
```

**Privilege escalation:**
```bash
# Block patterns
sudo           # Run as root
su             # Switch user
pkexec         # PolicyKit execute
```

**Download and execute:**
```bash
# Block patterns
curl.*\|.*bash     # curl URL | bash
wget.*\|.*sh       # wget URL | sh
.\|.*sh           # Any pipe to shell
```

**Network exposure:**
```bash
# Block patterns
nc.*-l             # Netcat listen mode
python.*-m.*http   # Python HTTP server
npm.*start         # May start servers
```

### Complete Security Hook

```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

# Only check bash commands
if [ "$TOOL_NAME" != "bash" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi

COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command')

# Dangerous patterns
PATTERNS=(
  "rm -rf"
  "sudo"
  "mkfs"
  "dd if="
  "format"
  "fdisk"
  "curl.*\|.*bash"
  "wget.*\|.*sh"
  "nc.*-l"
)

for PATTERN in "${PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qE "$PATTERN"; then
    echo "{
      \"permissionDecision\": \"deny\",
      \"permissionDecisionReason\": \"Blocked dangerous pattern: $PATTERN\"
    }" | jq -c
    exit 0
  fi
done

# Allow if no dangerous patterns
echo '{"permissionDecision":"allow"}' | jq -c
```

---

## Secret Scanning

### Detect Common Secret Patterns

```bash
#!/bin/bash
# .github/hooks/scripts/secret-scanner.sh

INPUT=$(cat)
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')

# Secret patterns (regex)
PATTERNS=(
  'AKIA[0-9A-Z]{16}'                    # AWS Access Key
  'ghp_[0-9a-zA-Z]{36}'                 # GitHub PAT
  'sk-[0-9a-zA-Z]{48}'                  # OpenAI API Key
  '[0-9]+-[0-9A-Za-z_]{32}'            # Slack Token
  'AIza[0-9A-Za-z\-_]{35}'             # Google API Key
  'ya29\.[0-9A-Za-z\-_]+'              # Google OAuth
  '[Bb]earer [A-Za-z0-9\-\._~\+\/]+=*' # Bearer token
  'password\s*=\s*["\047][^"\047]+["\047]' # Password in quotes
)

for PATTERN in "${PATTERNS[@]}"; do
  if echo "$TOOL_ARGS" | grep -qE "$PATTERN"; then
    echo '{
      "permissionDecision": "deny",
      "permissionDecisionReason": "Potential secret detected. Remove credentials before proceeding."
    }' | jq -c
    exit 0
  fi
done

echo '{"permissionDecision":"allow"}' | jq -c
```

**Note:** GitHub Copilot has **built-in secret scanning**. This hook provides additional coverage.

---

## Compliance & Audit Logging

### Comprehensive Audit Trail

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{"bash": "./audit/log-session-start.sh"}],
    "userPromptSubmitted": [{"bash": "./audit/log-prompt.sh"}],
    "preToolUse": [{"bash": "./audit/log-tool-request.sh"}],
    "postToolUse": [{"bash": "./audit/log-tool-result.sh"}],
    "errorOccurred": [{"bash": "./audit/log-error.sh"}],
    "sessionEnd": [{"bash": "./audit/log-session-end.sh"}]
  }
}
```

**Audit Log Format (JSON Lines):**
```jsonl
{"type":"session_start","timestamp":1708560000000,"source":"new","user":"alice"}
{"type":"prompt","timestamp":1708560005000,"prompt":"implement auth","user":"alice"}
{"type":"tool_request","timestamp":1708560010000,"tool":"edit","path":"auth.js","user":"alice"}
{"type":"tool_result","timestamp":1708560015000,"tool":"edit","result":"success","user":"alice"}
{"type":"session_end","timestamp":1708560100000,"reason":"complete","user":"alice"}
```

**Query audit logs:**
```bash
# Find all blocked operations
jq 'select(.decision == "deny")' audit.jsonl

# Find all tool usage by user
jq 'select(.user == "alice" and .type == "tool_request")' audit.jsonl

# Generate compliance report
jq -r '[.timestamp, .type, .user, .decision // "allow"] | @csv' audit.jsonl > report.csv
```

---

## Permission Model

### Hook Permissions

**Hooks run with:**
- ✅ Same permissions as GitHub Copilot process
- ✅ Access to repository files
- ✅ Access to system commands
- ✅ User environment variables

**No sandboxing:**
- ⚠️ Hooks are **NOT sandboxed**
- ⚠️ Can execute arbitrary code
- ⚠️ Trust only hooks from trusted sources

### Security Recommendations

**1. Repository default branch requirement:**
- Hooks must be on default branch (main/master)
- Prevents malicious PR hook execution
- Code review hooks like application code

**2. Review hooks carefully:**
```bash
# Review hook changes in PRs
git diff main...feature -- .github/hooks/

# Audit existing hooks
cat .github/hooks/scripts/*.sh | shellcheck
```

**3. Principle of least privilege:**
```bash
#!/bin/bash
# ❌ Bad - requires sudo
sudo validate-security

# ✅ Good - runs as current user
validate-security
```

**4. Input validation:**
```bash
#!/bin/bash
# Always validate JSON input
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid input" >&2
  exit 2
fi
```

---

## Secret Management

### Best Practices

**1. Never log sensitive data:**
```bash
#!/bin/bash
# ❌ Bad
echo "Command: $COMMAND" >> hook.log

# ✅ Good
SAFE=$(echo "$COMMAND" | sed -E 's/(password|token|key)=[^ ]+/\1=****/gi')
echo "Command: $SAFE" >> hook.log
```

**2. Use environment variables for secrets:**
```json
{
  "hooks": {
    "errorOccurred": [{
      "bash": "./notify.sh",
      "env": {
        "SLACK_WEBHOOK_URL": "${SLACK_WEBHOOK_URL}"
      }
    }]
  }
}
```

Load from environment (not hardcoded):
```bash
# ✅ Good - from environment
curl -X POST "$SLACK_WEBHOOK_URL" -d "$DATA"

# ❌ Bad - hardcoded
curl -X POST "https://hooks.slack.com/..." -d "$DATA"
```

**3. Secure external calls:**
```bash
#!/bin/bash
# Validate webhook URL
if [[ ! "$WEBHOOK_URL" =~ ^https:// ]]; then
  echo "Only HTTPS webhooks allowed" >&2
  exit 2
fi

# Use secure curl options
curl -X POST "$WEBHOOK_URL" \
  --max-time 10 \
  --fail-with-body \
  --silent \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD"
```

---

## Enterprise Governance

### Multi-Layer Security

**Layer 1: Secret scanning**
```bash
./scripts/scan-secrets.sh  # Check for credentials
```

**Layer 2: Command validation**
```bash
./scripts/validate-command.sh  # Check for dangerous patterns
```

**Layer 3: Policy enforcement**
```bash
./scripts/enforce-policy.sh  # Check against company policy
```

**Layer 4: Audit logging**
```bash
./scripts/audit-log.sh  # Log all requests
```

**Configuration:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {"bash": "./scripts/scan-secrets.sh", "timeoutSec": 3},
      {"bash": "./scripts/validate-command.sh", "timeoutSec": 3},
      {"bash": "./scripts/enforce-policy.sh", "timeoutSec": 5},
      {"bash": "./scripts/audit-log.sh", "timeoutSec": 2}
    ]
  }
}
```

### Gradual Rollout Strategy

**Phase 1: Observation (Weeks 1-2)**
```bash
#!/bin/bash
# Log only, allow everything
echo "$INPUT" >> security-observations.jsonl
echo '{"permissionDecision":"allow"}' | jq -c
```

**Phase 2: Warning (Weeks 3-4)**
```bash
#!/bin/bash
# Warn but allow
if is_dangerous; then
  echo "WARNING: Dangerous operation detected" >&2
  echo "{\"type\":\"warning\",\"command\":\"$COMMAND\"}" >> warnings.jsonl
fi
echo '{"permissionDecision":"allow"}' | jq -c
```

**Phase 3: Enforcement (Week 5+)**
```bash
#!/bin/bash
# Block dangerous operations
if is_dangerous; then
  echo '{"permissionDecision":"deny","permissionDecisionReason":"Policy violation"}' | jq -c
  exit 0
fi
echo '{"permissionDecision":"allow"}' | jq -c
```

---

## Context-Aware Security

### Environment-Specific Rules

```bash
#!/bin/bash
INPUT=$(cat)
CWD=$(echo "$INPUT" | jq -r '.cwd')

# Strict rules in production
if echo "$CWD" | grep -q "/production/"; then
  ALLOWED_TOOLS="view"
  ENFORCE_STRICT=true
# Medium rules in staging
elif echo "$CWD" | grep -q "/staging/"; then
  ALLOWED_TOOLS="view|edit"
  ENFORCE_STRICT=true
# Relaxed in development
else
  ALLOWED_TOOLS=".*"
  ENFORCE_STRICT=false
fi

# Apply rules based on environment
if [ "$ENFORCE_STRICT" = true ]; then
  TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
  if ! echo "$TOOL_NAME" | grep -qE "^($ALLOWED_TOOLS)$"; then
    echo '{
      "permissionDecision": "deny",
      "permissionDecisionReason": "Not allowed in this environment"
    }' | jq -c
    exit 0
  fi
fi

echo '{"permissionDecision":"allow"}' | jq -c
```

---

## Security Checklist

### Before Deploying Hooks

- [ ] Review all hook scripts for security issues
- [ ] Validate no hardcoded secrets
- [ ] Test with malicious inputs
- [ ] Ensure proper error handling
- [ ] Verify timeout settings
- [ ] Check file permissions (executable)
- [ ] Review on default branch requirement
- [ ] Document expected behavior
- [ ] Set up audit logging
- [ ] Test gradual rollout plan

### Ongoing Security

- [ ] Regularly review audit logs
- [ ] Update secret patterns
- [ ] Refine dangerous command patterns
- [ ] Monitor hook performance
- [ ] Review hook changes in PRs
- [ ] Update documentation
- [ ] Train team on security policies

---

## See Also

- [Event Types & Triggers](./events-reference.md) - preToolUse hook
- [Examples](./examples.md) - Security patterns
- [Configuration](./configuration.md) - Deployment
- [Troubleshooting](./troubleshooting.md) - Debug security issues

---

**Sources:**
- [About Hooks - GitHub Copilot](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-hooks)
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [Copilot CLI Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

---

**Document Version:** 1.0
**Status:** ✅ Complete
