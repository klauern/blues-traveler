# OpenAI Codex - Security & Safety

**Last Updated:** 2026-02-21

---

## Overview

Codex's security model is built around **cloud sandbox isolation**, **approval-based execution**, and **environment-level secret management**. Unlike local tools, Codex runs code in isolated containers with controlled permissions.

---

## Permission Model

### Approval-Based Execution

**High-risk operations require user approval:**

```
Agent wants to execute: rm -rf node_modules
    ↓
Sends approval_request event to client
    ↓
User sees: "Delete node_modules directory?"
[Approve] [Deny]
    ↓ (User clicks)
Client sends approval response
    ↓
Agent proceeds or cancels
```

### Approval Request Example

```json
{
  "jsonrpc": "2.0",
  "method": "approval_request",
  "params": {
    "id": "approval_789",
    "operation": "bash_command",
    "command": "curl https://example.com/script.sh | bash",
    "reason": "Installing dependencies",
    "risk_level": "high"
  }
}
```

### Risk Levels

| Level | Example Operations | Auto-Approve |
|-------|-------------------|--------------|
| **Low** | `ls`, `cat`, `grep` | ✅ (configurable) |
| **Medium** | `npm install`, file writes | ❌ |
| **High** | `rm -rf`, `curl \| bash`, `sudo` | ❌ |
| **Critical** | System modifications, network requests | ❌ |

---

## Blocking Dangerous Operations

### Via Hooks

**Block operations before execution:**

```toml
# ~/.codex/config.toml
[hooks.tool.before]
command = "/usr/local/bin/safety-check"
```

```bash
#!/bin/bash
# /usr/local/bin/safety-check

TOOL_NAME=$1
TOOL_ARGS=$2

# Block dangerous bash commands
if [ "$TOOL_NAME" = "Bash" ]; then
    COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command')

    # Block rm -rf
    if echo "$COMMAND" | grep -q "rm -rf"; then
        echo "ERROR: rm -rf is blocked for safety" >&2
        exit 1  # Block execution
    fi

    # Block curl | bash
    if echo "$COMMAND" | grep -q "curl.*|.*bash"; then
        echo "ERROR: Piping curl to bash is blocked" >&2
        exit 1
    fi

    # Block sudo
    if echo "$COMMAND" | grep -q "^sudo "; then
        echo "ERROR: sudo is not allowed" >&2
        exit 1
    fi
fi

exit 0  # Allow
```

### Blocked Operations List

**Common dangerous patterns:**

```bash
# Dangerous file operations
rm -rf /
rm -rf *
find . -delete

# Dangerous downloads
curl | bash
wget | sh
eval $(curl ...)

# System modifications
sudo *
chmod 777 *
chown -R root *

# Fork bombs
:(){ :|:& };:

# Network attacks
ping -f
hping3 --flood
```

---

## Sandbox Isolation

### Cloud Sandbox Architecture

```
┌─────────────────────────────────────────┐
│        Isolated Container               │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐   │
│  │  Codex Agent                    │   │
│  │  - Limited filesystem access     │   │
│  │  - No sudo/root                  │   │
│  │  - Controlled network egress     │   │
│  │  - Resource limits enforced      │   │
│  └─────────────────────────────────┘   │
├─────────────────────────────────────────┤
│  Resources:                             │
│  - CPU: 4 cores max                     │
│  - Memory: 8 GB max                     │
│  - Disk: 50 GB ephemeral                │
│  - Network: Outbound only (HTTPS)       │
└─────────────────────────────────────────┘
```

### Sandbox Restrictions

| Resource | Restriction | Reason |
|----------|-------------|--------|
| **Filesystem** | `/sandbox/*` only | Prevent access to host |
| **Network** | Outbound HTTPS only | Prevent attacks |
| **Processes** | Limited to 100 | Prevent fork bombs |
| **Root** | No sudo/root access | Prevent privilege escalation |
| **Devices** | No device access | Security |

---

## Secret Management

### Best Practices

**✅ Good: Environment Variables**

```python
# Configure in OpenAI dashboard or local config
response = client.responses.create(
    model="gpt-5.2-codex",
    environment={
        "DATABASE_URL": os.environ["DB_URL"],
        "API_KEY": os.environ["API_KEY"],
        "GITHUB_TOKEN": os.environ["GH_TOKEN"]
    },
    messages=[{...}]
)
```

**❌ Bad: Hardcoded Secrets**

```python
# Never do this!
environment={
    "DATABASE_URL": "postgres://user:password@host/db",
    "API_KEY": "sk-abc123",
}
```

### Secret Detection Hook

```bash
#!/bin/bash
# /usr/local/bin/detect-secrets

FILE_PATH=$1

# Read file content
CONTENT=$(cat "$FILE_PATH")

# Detect common secret patterns
if echo "$CONTENT" | grep -E '(api_key|password|secret|token).*=.*["'\''][a-zA-Z0-9]{20,}["'\'']'; then
    echo "ERROR: Potential secret detected in $FILE_PATH" >&2
    echo "Use environment variables instead" >&2
    exit 1  # Block write
fi

# Detect AWS keys
if echo "$CONTENT" | grep -E 'AKIA[0-9A-Z]{16}'; then
    echo "ERROR: AWS access key detected" >&2
    exit 1
fi

# Detect GitHub tokens
if echo "$CONTENT" | grep -E 'ghp_[a-zA-Z0-9]{36}'; then
    echo "ERROR: GitHub token detected" >&2
    exit 1
fi

exit 0  # Allow
```

### Secret Rotation

```bash
# Rotate secrets regularly
export OPENAI_API_KEY="sk-new-key"
export DATABASE_URL="postgres://new-credentials"

# Update in OpenAI dashboard
# Update in CI/CD secrets
# Update in production environment
```

---

## Audit Logging

### Hook Execution Logs

```bash
#!/bin/bash
# /usr/local/bin/audit-log

TOOL_NAME=$1
TOOL_ARGS=$2
USER=${CODEX_USER_ID:-unknown}
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Log to audit file
echo "$TIMESTAMP|$USER|$CODEX_RESPONSE_ID|$TOOL_NAME|$TOOL_ARGS" \
  >> /var/log/codex-audit.log

# Send to SIEM
curl -X POST https://siem.example.com/events \
  -H "Content-Type: application/json" \
  -d "{
    \"timestamp\": \"$TIMESTAMP\",
    \"user\": \"$USER\",
    \"response_id\": \"$CODEX_RESPONSE_ID\",
    \"tool\": \"$TOOL_NAME\",
    \"arguments\": $(echo "$TOOL_ARGS" | jq -c .)
  }"
```

### Log Locations

| Log Type | Location | Retention |
|----------|----------|-----------|
| Hook execution | `~/.codex/hooks.log` | 30 days |
| Tool usage | `~/.codex/tools.log` | 30 days |
| API requests | OpenAI dashboard | 90 days |
| Audit logs | Custom SIEM | Per policy |

---

## Network Security

### Egress Control

**Sandbox allows only HTTPS outbound:**

```bash
# ✅ Allowed
curl https://api.github.com/repos/...
wget https://example.com/file.tar.gz

# ❌ Blocked
curl http://internal-server/  # HTTP not HTTPS
nc internal-host 22  # Raw TCP blocked
ssh user@host  # SSH blocked
```

### Webhook Security

**Verify webhook signatures:**

```python
import hmac
import hashlib

def verify_webhook(request):
    signature = request.headers.get("webhook-signature")
    timestamp = request.headers.get("webhook-timestamp")
    body = request.body

    secret = os.environ["OPENAI_WEBHOOK_SECRET"]

    expected_sig = hmac.new(
        secret.encode(),
        f"{timestamp}.{body.decode()}".encode(),
        hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(signature, f"v1,{expected_sig}")

# Use in webhook handler
@app.post("/webhooks/codex")
async def handle_webhook(request):
    if not verify_webhook(request):
        raise HTTPException(status_code=401)

    # Process webhook safely
```

---

## Code Injection Prevention

### Validate Inputs

```bash
#!/bin/bash
# /usr/local/bin/validate-bash-command

COMMAND=$1

# Block command injection attempts
if echo "$COMMAND" | grep -E '(;|\||&|`|\$\()'; then
    echo "ERROR: Potential command injection detected" >&2
    exit 1
fi

# Block path traversal
if echo "$COMMAND" | grep -E '\.\./'; then
    echo "ERROR: Path traversal detected" >&2
    exit 1
fi

exit 0
```

### Sanitize File Paths

```python
#!/usr/bin/env python3
import os
import sys

file_path = sys.argv[1]

# Prevent path traversal
if ".." in file_path:
    print("ERROR: Path traversal detected", file=sys.stderr)
    sys.exit(1)

# Ensure within workspace
workspace = os.environ.get("CODEX_WORKSPACE", "/sandbox/workspace")
abs_path = os.path.abspath(file_path)

if not abs_path.startswith(workspace):
    print(f"ERROR: Access outside workspace denied", file=sys.stderr)
    sys.exit(1)

sys.exit(0)
```

---

## Compliance & Governance

### GDPR / Data Privacy

**Data handling:**
- Sandbox data is ephemeral (deleted after completion)
- No persistent storage of user code
- Opt-in data retention policies
- Data encrypted in transit and at rest

**User controls:**
- Delete conversation history
- Opt out of training data
- Export all data
- Request account deletion

### SOC 2 Compliance

**OpenAI's security controls:**
- Annual SOC 2 Type II audit
- Penetration testing
- Security incident response
- Vendor risk management

### Access Controls

```toml
# Restrict Codex usage to approved users
[access]
allowed_users = ["alice@example.com", "bob@example.com"]
allowed_teams = ["engineering", "devops"]
require_approval = ["sudo", "rm", "deploy"]
```

---

## Incident Response

### Webhook for Security Events

```toml
[hooks.event.security]
command = "/usr/local/bin/security-alert"
events = ["suspicious_command", "secret_detected", "approval_denied"]
```

```bash
#!/bin/bash
# /usr/local/bin/security-alert

EVENT_TYPE=$1
EVENT_DATA=$2

# Alert security team
curl -X POST https://security-siem.example.com/alert \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SIEM_TOKEN" \
  -d "{
    \"severity\": \"high\",
    \"event_type\": \"$EVENT_TYPE\",
    \"user\": \"$CODEX_USER_ID\",
    \"response_id\": \"$CODEX_RESPONSE_ID\",
    \"details\": $(echo "$EVENT_DATA" | jq -c .)
  }"

# Page on-call if critical
if [ "$EVENT_TYPE" = "critical_violation" ]; then
    curl -X POST https://pagerduty.com/incidents \
      -H "Content-Type: application/json" \
      -d "{\"incident\": {\"title\": \"Codex Security Event\"}}"
fi
```

---

## Best Practices Summary

### 1. Use Environment Variables for Secrets

```python
# ✅ Good
environment={"API_KEY": os.environ["API_KEY"]}

# ❌ Bad
environment={"API_KEY": "hardcoded-secret"}
```

### 2. Implement Pre-Execution Hooks

```toml
[hooks.tool.before]
command = "/usr/local/bin/safety-check"
```

### 3. Enable Audit Logging

```toml
[hooks.tool.after]
command = "/usr/local/bin/audit-log"
```

### 4. Verify Webhook Signatures

```python
if not verify_webhook(request):
    raise HTTPException(status_code=401)
```

### 5. Restrict Network Access

```toml
[security]
allowed_domains = ["api.github.com", "registry.npmjs.org"]
block_http = true  # Only allow HTTPS
```

### 6. Review Approvals Regularly

```bash
# Review approval logs
codex logs approvals --last-7-days

# Analyze patterns
codex logs approvals --denied | grep "rm -rf"
```

---

## Sources

- [OpenAI Security Practices](https://openai.com/security)
- [Codex Sandbox Documentation](https://developers.openai.com/codex/security/)
- [Standard Webhooks Specification](https://www.standardwebhooks.com/)
- Research: [API Capabilities](../../../research-notes/codex/api-capabilities.md)

---

**Last Updated:** 2026-02-21
