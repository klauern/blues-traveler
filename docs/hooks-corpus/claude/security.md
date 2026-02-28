# Claude Code Hooks - Security & Safety

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete

---

## Table of Contents

1. [Permission Model](#permission-model)
2. [Blocking Dangerous Operations](#blocking-dangerous-operations)
3. [Secret Management](#secret-management)
4. [Audit Logging](#audit-logging)
5. [Enterprise Security Patterns](#enterprise-security-patterns)
6. [Common Security Hooks](#common-security-hooks)

---

## Permission Model

### No Sandboxing

**Official Warning:**
> Hooks execute shell commands with your full user permissions. They can modify, delete, or access any files your user account can access.

**Implications:**
- Hooks run with **full user permissions**
- Can read, write, delete any file user can access
- Can execute any command
- Can make network requests
- No resource isolation
- No capability restrictions

**Trust Model:** Hooks are trusted code. Never run hooks from untrusted sources.

**Source:** [Official Security Warning](https://code.claude.com/docs/en/hooks#security)

---

### Policy Hooks

**Managed Policy Settings:**
- Organization admins can enforce hooks via policy
- Users cannot disable or modify policy hooks
- Policy hooks always execute
- `"disableAllHooks": true` does not affect policy hooks

**Enterprise Use:**
- Company-wide security enforcement
- Compliance requirements
- Audit logging
- Standard workflows

---

## Blocking Dangerous Operations

### Dangerous Bash Commands

**Block patterns:**

```bash
#!/bin/bash
# .claude/hooks/security-check.sh

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Dangerous patterns to block
DANGEROUS_PATTERNS=(
  "rm -rf"
  "sudo rm"
  "mkfs"
  "dd if="
  ":(){ :|:& };:"  # Fork bomb
  "curl.*\|.*sh"   # Pipe to shell
  "wget.*\|.*sh"
  "> /dev/sd"      # Disk writes
  "chmod 777"
  "chown root"
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qE "$pattern"; then
    echo "BLOCKED: Dangerous command pattern detected: $pattern" >&2
    echo "Command: $COMMAND" >&2
    exit 2
  fi
done

exit 0
```

**Configuration:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/security-check.sh"
          }
        ]
      }
    ]
  }
}
```

**Source:** [GitHub - Trail of Bits Config](https://github.com/trailofbits/claude-code-config)

---

### File Path Validation

**Prevent path traversal:**

```bash
#!/bin/bash
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Block path traversal
if [[ "$FILE" == *".."* ]]; then
  echo "Path traversal detected: $FILE" >&2
  exit 2
fi

# Ensure file is within project directory
if [[ ! "$FILE" =~ ^"$CLAUDE_PROJECT_DIR" ]]; then
  echo "File outside project directory: $FILE" >&2
  exit 2
fi

# Block sensitive files
if [[ "$FILE" =~ \.(env|key|pem|p12)$ ]]; then
  echo "Cannot modify sensitive file: $FILE" >&2
  exit 2
fi

exit 0
```

---

### Branch Protection

**Block commits to protected branches:**

```bash
#!/bin/bash
# .claude/hooks/branch-protection.sh

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")

PROTECTED_BRANCHES=("main" "master" "production" "release")

for protected in "${PROTECTED_BRANCHES[@]}"; do
  if [[ "$BRANCH" == "$protected" ]]; then
    if [[ "$TOOL" == "Edit" || "$TOOL" == "Write" ]]; then
      echo "Cannot modify files on protected branch: $BRANCH" >&2
      echo "Create a feature branch: git checkout -b feature/your-name" >&2
      exit 2
    fi
  fi
done

exit 0
```

**Source:** [GitHub - ChrisWiles/claude-code-showcase](https://github.com/ChrisWiles/claude-code-showcase)

---

### Sensitive Content Filtering

**Block secrets in prompts:**

```bash
#!/bin/bash
# .claude/hooks/secret-filter.sh

INPUT=$(cat)
PROMPT=$(echo "$INPUT" | jq -r '.prompt // empty')

# Patterns that indicate secrets
SECRET_PATTERNS=(
  "password"
  "api[_-]?key"
  "secret"
  "token"
  "credential"
  "private[_-]?key"
  "sk-[a-zA-Z0-9]{32,}"  # OpenAI API key pattern
)

for pattern in "${SECRET_PATTERNS[@]}"; do
  if echo "$PROMPT" | grep -qiE "$pattern"; then
    echo "Potential secret detected in prompt" >&2
    echo "Pattern: $pattern" >&2
    echo "Please remove sensitive information before proceeding" >&2
    exit 2
  fi
done

exit 0
```

---

## Secret Management

### Best Practices

**1. Never hardcode secrets in hooks:**

**Bad:**
```bash
#!/bin/bash
API_KEY="sk-abc123..."  # NEVER DO THIS
```

**Good:**
```bash
#!/bin/bash
if [[ -z "$SLACK_WEBHOOK_URL" ]]; then
  echo "SLACK_WEBHOOK_URL not set" >&2
  exit 1
fi
```

---

**2. Use environment variables:**

```bash
#!/bin/bash
# Load from .env file (gitignored)
if [[ -f "$CLAUDE_PROJECT_DIR/.env" ]]; then
  export $(grep -v '^#' "$CLAUDE_PROJECT_DIR/.env" | xargs)
fi

curl -X POST "$SLACK_WEBHOOK_URL" -d "$DATA"
```

---

**3. Use secret management tools:**

```bash
#!/bin/bash
# Use 1Password CLI
OP_TOKEN=$(op item get "Slack Webhook" --fields url 2>/dev/null)
if [[ -z "$OP_TOKEN" ]]; then
  echo "Failed to retrieve secret from 1Password" >&2
  exit 1
fi

curl -X POST "$OP_TOKEN" -d "$DATA"
```

---

**4. Skip sensitive files:**

```bash
#!/bin/bash
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Skip .env files
if [[ "$FILE" == *".env"* ]]; then
  echo "Skipping sensitive file: $FILE" >&2
  exit 0
fi

# Continue with hook logic...
```

---

### SessionStart Environment Setup

**Securely load environment:**

```bash
#!/bin/bash
# SessionStart hook

# Load secrets from secure storage
if command -v op &>/dev/null; then
  # 1Password CLI
  GITHUB_TOKEN=$(op item get "GitHub Token" --fields token)
  SLACK_WEBHOOK=$(op item get "Slack Webhook" --fields url)
elif command -v aws &>/dev/null; then
  # AWS Secrets Manager
  GITHUB_TOKEN=$(aws secretsmanager get-secret-value --secret-id github-token --query SecretString --output text)
fi

# Persist to session environment
if [[ -n "$CLAUDE_ENV_FILE" ]]; then
  echo "export GITHUB_TOKEN=\"$GITHUB_TOKEN\"" >> "$CLAUDE_ENV_FILE"
  echo "export SLACK_WEBHOOK=\"$SLACK_WEBHOOK\"" >> "$CLAUDE_ENV_FILE"
fi

exit 0
```

---

## Audit Logging

### Complete Audit Trail

**Log all hook executions:**

```bash
#!/bin/bash
# .claude/hooks/audit-logger.sh

INPUT=$(cat)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LOG_FILE="$HOME/.claude/audit.log"

# Append full event to audit log
echo "$INPUT" | jq -c ". + {timestamp: \"$TIMESTAMP\"}" >> "$LOG_FILE"

# Continue normally
exit 0
```

**Configuration:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/audit-logger.sh",
            "async": true
          }
        ]
      }
    ]
  }
}
```

---

### Security Event Logging

**Log security-relevant events:**

```bash
#!/bin/bash
# .claude/hooks/security-audit.sh

INPUT=$(cat)
EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
TOOL=$(echo "$INPUT" | jq -r '.tool_name // "N/A"')
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

AUDIT_LOG="$HOME/.claude/security-audit.log"

# Log security events
case "$EVENT" in
  PreToolUse)
    COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')
    if [[ -n "$COMMAND" ]]; then
      echo "$TIMESTAMP [PreToolUse:Bash] $COMMAND" >> "$AUDIT_LOG"
    fi
    ;;
  PermissionRequest)
    echo "$TIMESTAMP [PermissionRequest:$TOOL]" >> "$AUDIT_LOG"
    ;;
  ConfigChange)
    SOURCE=$(echo "$INPUT" | jq -r '.source')
    FILE=$(echo "$INPUT" | jq -r '.file_path')
    echo "$TIMESTAMP [ConfigChange:$SOURCE] $FILE" >> "$AUDIT_LOG"
    ;;
esac

exit 0
```

---

### Log Rotation

```bash
#!/bin/bash
# Rotate audit logs daily

LOG_FILE="$HOME/.claude/audit.log"
MAX_SIZE=$((10 * 1024 * 1024))  # 10MB

if [[ -f "$LOG_FILE" ]] && [[ $(stat -f%z "$LOG_FILE") -gt $MAX_SIZE ]]; then
  ARCHIVE="$HOME/.claude/audit-$(date +%Y%m%d).log.gz"
  gzip -c "$LOG_FILE" > "$ARCHIVE"
  > "$LOG_FILE"  # Truncate
  echo "Rotated audit log to $ARCHIVE" >&2
fi
```

---

## Enterprise Security Patterns

### Compliance Enforcement

**Block operations in production:**

```bash
#!/bin/bash
INPUT=$(cat)

# Detect production environment
if [[ -f "/etc/production" ]] || [[ "$ENVIRONMENT" == "production" ]]; then
  TOOL=$(echo "$INPUT" | jq -r '.tool_name')

  # Block destructive operations
  if [[ "$TOOL" == "Write" ]] || [[ "$TOOL" == "Edit" ]]; then
    echo "Direct file modifications blocked in production" >&2
    echo "Use deployment pipeline instead" >&2
    exit 2
  fi
fi

exit 0
```

---

### Multi-Approval Workflow

**Require approval for critical operations:**

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Critical patterns require approval
if echo "$COMMAND" | grep -qE "DROP TABLE|DELETE FROM.*WHERE 1=1"; then
  echo "Critical database operation detected" >&2
  echo "Approval required from: security@company.com" >&2

  # Send notification
  curl -X POST "$APPROVAL_WEBHOOK" \
    -H "Content-Type: application/json" \
    -d "{\"command\": \"$COMMAND\", \"user\": \"$USER\"}"

  exit 2
fi

exit 0
```

---

### Rate Limiting

**Limit expensive operations:**

```bash
#!/bin/bash
RATE_FILE="$HOME/.claude/rate-limit"
RATE_LIMIT=10  # Max 10 per hour

# Count executions in last hour
COUNT=$(find "$RATE_FILE" -mmin -60 2>/dev/null | wc -l)

if [[ $COUNT -ge $RATE_LIMIT ]]; then
  echo "Rate limit exceeded ($RATE_LIMIT/hour)" >&2
  exit 2
fi

# Record execution
touch "$RATE_FILE-$(date +%s)"

exit 0
```

---

## Common Security Hooks

### 1. Block rm -rf

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if echo "$COMMAND" | grep -qE "rm\s+-[rf]{1,2}\s+/"; then
  echo "Blocked: rm -rf with root path" >&2
  exit 2
fi

exit 0
```

---

### 2. Block sudo

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if echo "$COMMAND" | grep -qE "^sudo\s"; then
  echo "Blocked: sudo commands not allowed" >&2
  echo "Run Claude Code with appropriate permissions" >&2
  exit 2
fi

exit 0
```

---

### 3. Block curl | sh

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if echo "$COMMAND" | grep -qE "(curl|wget).*\|.*(sh|bash)"; then
  echo "Blocked: Piping remote scripts to shell" >&2
  echo "Download and review scripts before executing" >&2
  exit 2
fi

exit 0
```

---

### 4. Validate File Paths

```bash
#!/bin/bash
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Block absolute paths outside project
if [[ "$FILE" == /* ]] && [[ ! "$FILE" =~ ^"$CLAUDE_PROJECT_DIR" ]]; then
  echo "Blocked: Absolute path outside project" >&2
  exit 2
fi

# Block path traversal
if [[ "$FILE" == *".."* ]]; then
  echo "Blocked: Path traversal detected" >&2
  exit 2
fi

exit 0
```

---

### 5. Require Tests Before Deployment

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Detect deployment commands
if echo "$COMMAND" | grep -qE "(deploy|ship|release)"; then
  # Require tests to pass
  if ! npm test &>/dev/null; then
    echo "Blocked: Tests must pass before deployment" >&2
    exit 2
  fi

  # Require git clean state
  if [[ -n $(git status --porcelain) ]]; then
    echo "Blocked: Uncommitted changes detected" >&2
    exit 2
  fi
fi

exit 0
```

---

## See Also

- [Scripting & Execution](./scripting.md) - Exit codes and error handling
- [Examples](./examples.md) - Ready-to-use security patterns
- [Configuration](./configuration.md) - Policy hooks setup

---

**Sources:**
- [Official Security Warning](https://code.claude.com/docs/en/hooks#security)
- [Trail of Bits Config](https://github.com/trailofbits/claude-code-config)
- [GitHub Examples](https://github.com/disler/claude-code-hooks-mastery)

**Document Version:** 1.0
**Research Date:** 2026-02-21
