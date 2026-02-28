# Gemini - Security & Safety

**Last Updated:** 2026-02-21
**Version:** Gemini CLI v0.26.0+

---

## Overview

Gemini CLI hooks provide powerful security capabilities for protecting your development environment, enforcing policies, and preventing dangerous operations. This document covers security patterns, best practices, and real-world examples.

**Sources:**
- [Gemini CLI Hooks Best Practices](https://geminicli.com/docs/hooks/best-practices/)
- [Writing Hooks Documentation](https://geminicli.com/docs/hooks/writing-hooks/)

---

## Permission Model

### Execution Environment

**Hook Permissions:**
- Hooks run with **CLI user's full permissions**
- No sandboxing or isolation
- Full filesystem access
- Full network access
- Can execute arbitrary code

**Security Implications:**

| Access Type | Hook Capability | Risk Level |
|-------------|----------------|------------|
| Filesystem | Read/write any file user can access | ⚠️ High |
| Network | Make external requests | ⚠️ High |
| Process | Spawn child processes | ⚠️ High |
| Environment | Access all environment variables | ⚠️ Medium |
| System | Execute shell commands | 🔴 Critical |

**Trust Model:**
- User must trust all hook code
- Review hooks before enabling
- Use hooks from trusted sources only
- Prefer audited, version-controlled hooks

---

## Blocking Dangerous Operations

### File Operations Security

#### Block Sensitive File Writes

**Pattern: Prevent writing to sensitive files**

```bash
#!/bin/bash
# .gemini/hooks/block-sensitive-files.sh
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool.name // empty')
FILE=$(echo "$INPUT" | jq -r '.tool.params.path // empty')

# Define sensitive patterns
SENSITIVE_PATTERNS=(
  '\.env$'
  '\.env\.'
  'credentials'
  'secrets'
  '\.ssh/'
  '\.aws/'
  '\.config/gcloud/'
  'id_rsa'
  'id_ed25519'
  '\.pem$'
  '\.key$'
  'password'
  'token'
)

# Check if writing to file
if [[ "$TOOL" =~ ^(write|edit|create|update)_file$ ]]; then
  for pattern in "${SENSITIVE_PATTERNS[@]}"; do
    if [[ "$FILE" =~ $pattern ]]; then
      echo "{\"decision\": \"deny\", \"systemMessage\": \"Blocked: Cannot modify sensitive file $FILE\"}"
      exit 0
    fi
  done
fi

echo '{"decision": "allow"}'
```

**Configure:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "block-sensitive-files",
        "matcher": "(write|edit|create|update)_file",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": [".gemini/hooks/block-sensitive-files.sh"]
          }
        ]
      }
    ]
  }
}
```

#### Detect Secrets in Content

**Pattern: Scan file content for secrets before writing**

```python
#!/usr/bin/env python3
# .gemini/hooks/scan-secrets.py
import json
import sys
import re

SECRET_PATTERNS = [
    # API Keys
    (r'[A-Za-z0-9]{32,}', 'API key pattern detected'),

    # AWS Keys
    (r'AKIA[0-9A-Z]{16}', 'AWS Access Key ID detected'),

    # Private Keys
    (r'-----BEGIN (RSA |EC )?PRIVATE KEY-----', 'Private key detected'),

    # GitHub Tokens
    (r'gh[ps]_[a-zA-Z0-9]{36}', 'GitHub token detected'),

    # Generic secrets
    (r'(secret|password|token|api[_-]?key)\s*[:=]\s*["\']?[a-zA-Z0-9]{10,}', 'Secret pattern detected'),
]

def scan_content(content):
    """Scan content for secrets"""
    for pattern, message in SECRET_PATTERNS:
        if re.search(pattern, content, re.IGNORECASE):
            return False, message
    return True, None

def main():
    input_data = json.load(sys.stdin)

    tool = input_data.get('tool', {})
    if tool.get('name') not in ['write_file', 'create_file']:
        print(json.dumps({'decision': 'allow'}))
        return

    content = tool.get('params', {}).get('content', '')
    is_safe, message = scan_content(content)

    if not is_safe:
        output = {
            'decision': 'deny',
            'systemMessage': f'Blocked: {message}. Review content before committing.'
        }
    else:
        output = {'decision': 'allow'}

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

---

### Command Execution Security

#### Block Dangerous Commands

**Pattern: Prevent dangerous shell commands**

```bash
#!/bin/bash
# .gemini/hooks/block-dangerous-commands.sh
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool.name // empty')

if [[ "$TOOL" != "execute_command" ]]; then
  echo '{"decision": "allow"}'
  exit 0
fi

CMD=$(echo "$INPUT" | jq -r '.tool.params.command // empty')

# Dangerous patterns
DANGEROUS_PATTERNS=(
  'rm -rf /'           # Delete root
  'sudo rm -rf'        # Sudo delete
  'curl.*\|.*sh'       # Pipe curl to shell
  'wget.*\|.*sh'       # Pipe wget to shell
  'eval'               # Eval command
  ':(){ :|:& };:'      # Fork bomb
  'dd if=/dev/zero'    # Disk fill
  'chmod 777'          # Insecure permissions
  'chown.*:.*-R\s+/'   # Recursive ownership change
  'mkfs\.'             # Format disk
  '>/dev/sd'           # Write to disk
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if [[ "$CMD" =~ $pattern ]]; then
    echo "{\"decision\": \"deny\", \"systemMessage\": \"Blocked: Dangerous command pattern detected: $pattern\"}"
    exit 0
  fi
done

# Check for suspicious sudo usage
if [[ "$CMD" =~ ^sudo ]] && [[ ! "$CMD" =~ ^sudo\s+(apt|yum|brew) ]]; then
  echo '{"decision": "deny", "systemMessage": "Blocked: Sudo commands require manual approval"}'
  exit 0
fi

echo '{"decision": "allow"}'
```

#### Whitelist Safe Commands

**Pattern: Only allow known safe commands**

```python
#!/usr/bin/env python3
# .gemini/hooks/whitelist-commands.py
import json
import sys
import shlex

SAFE_COMMANDS = {
    # Development tools
    'npm', 'yarn', 'pnpm', 'bun',
    'git', 'gh',
    'node', 'python', 'python3', 'ruby', 'go', 'cargo',

    # Build tools
    'make', 'cmake', 'gradle', 'mvn',

    # Testing
    'jest', 'pytest', 'mocha', 'vitest',

    # Linting/formatting
    'eslint', 'prettier', 'black', 'ruff', 'gofmt',

    # Safe system commands
    'ls', 'cat', 'grep', 'find', 'echo', 'pwd',
}

def is_safe_command(cmd):
    """Check if command uses safe executable"""
    try:
        parts = shlex.split(cmd)
        if not parts:
            return False

        executable = parts[0].split('/')[-1]  # Get base command
        return executable in SAFE_COMMANDS
    except:
        return False

def main():
    input_data = json.load(sys.stdin)

    tool = input_data.get('tool', {})
    if tool.get('name') != 'execute_command':
        print(json.dumps({'decision': 'allow'}))
        return

    cmd = tool.get('params', {}).get('command', '')

    if is_safe_command(cmd):
        output = {'decision': 'allow'}
    else:
        output = {
            'decision': 'deny',
            'systemMessage': f'Blocked: Command not in whitelist. Review and approve manually.'
        }

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

---

### Network Security

#### Block External Network Requests

**Pattern: Prevent data exfiltration**

```python
#!/usr/bin/env python3
# .gemini/hooks/block-external-requests.py
import json
import sys
import re
from urllib.parse import urlparse

ALLOWED_DOMAINS = [
    'api.github.com',
    'api.openai.com',
    'registry.npmjs.org',
    'pypi.org',
    # Add your trusted domains
]

def is_allowed_url(url):
    """Check if URL is allowed"""
    try:
        parsed = urlparse(url)
        domain = parsed.netloc.lower()

        # Allow localhost
        if domain in ['localhost', '127.0.0.1', '::1']:
            return True

        # Check whitelist
        return domain in ALLOWED_DOMAINS
    except:
        return False

def main():
    input_data = json.load(sys.stdin)

    tool = input_data.get('tool', {})
    params = tool.get('params', {})

    # Check for URL parameters
    url = params.get('url') or params.get('endpoint')
    if not url:
        print(json.dumps({'decision': 'allow'}))
        return

    if is_allowed_url(url):
        output = {'decision': 'allow'}
    else:
        output = {
            'decision': 'deny',
            'systemMessage': f'Blocked: External request to {url} not allowed'
        }

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

---

## Secret Management

### Environment Variables

**Best Practice: Never hardcode secrets in hooks**

```bash
# ❌ BAD - Hardcoded secret
SLACK_WEBHOOK="https://hooks.slack.com/services/ABC123/..."

# ✅ GOOD - Use environment variable
SLACK_WEBHOOK="$SLACK_WEBHOOK_URL"
```

**Configuration:**
```json
{
  "hooks": {
    "AfterTool": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": [".gemini/hooks/notify.sh"],
            "env": {
              "SLACK_WEBHOOK_URL": "${SLACK_WEBHOOK_URL}"
            }
          }
        ]
      }
    ]
  }
}
```

**Load from `.env` file:**
```bash
#!/bin/bash
# Load secrets from .env (not in git)
if [ -f .gemini/.env ]; then
  set -a
  source .gemini/.env
  set +a
fi

WEBHOOK="$SLACK_WEBHOOK_URL"
# Use webhook...
```

### Secret Scanning in Git Context

**Pattern: Prevent committing secrets**

```python
#!/usr/bin/env python3
# .gemini/hooks/pre-commit-scan.py
import json
import sys
import subprocess
import re

SECRET_PATTERNS = [
    r'AKIA[0-9A-Z]{16}',  # AWS
    r'gh[ps]_[a-zA-Z0-9]{36}',  # GitHub
    r'sk-[a-zA-Z0-9]{48}',  # OpenAI
    # Add more patterns
]

def scan_staged_files():
    """Scan files staged for commit"""
    # Get staged files
    result = subprocess.run(
        ['git', 'diff', '--cached', '--name-only'],
        capture_output=True,
        text=True
    )

    files = result.stdout.strip().split('\n')

    for file in files:
        if not file:
            continue

        # Get staged content
        result = subprocess.run(
            ['git', 'show', f':{file}'],
            capture_output=True,
            text=True
        )

        content = result.stdout

        # Scan for secrets
        for pattern in SECRET_PATTERNS:
            if re.search(pattern, content):
                return False, f'Secret detected in {file}'

    return True, None

def main():
    input_data = json.load(sys.stdin)

    tool = input_data.get('tool', {})
    if tool.get('name') != 'git_commit':
        print(json.dumps({'decision': 'allow'}))
        return

    is_safe, message = scan_staged_files()

    if not is_safe:
        output = {
            'decision': 'deny',
            'systemMessage': f'Blocked: {message}'
        }
    else:
        output = {'decision': 'allow'}

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

---

## Audit Logging

### Comprehensive Audit Trail

**Pattern: Log all tool executions**

```python
#!/usr/bin/env python3
# .gemini/hooks/audit-log.py
import json
import sys
from datetime import datetime
import os

LOG_FILE = '.gemini/audit.jsonl'

def log_event(event_data):
    """Append event to audit log"""
    os.makedirs(os.path.dirname(LOG_FILE), exist_ok=True)

    log_entry = {
        'timestamp': datetime.now().isoformat(),
        'event': event_data.get('event'),
        'tool': event_data.get('tool', {}).get('name'),
        'params': event_data.get('tool', {}).get('params'),
        'session': event_data.get('sessionId'),
        'request': event_data.get('requestId')
    }

    with open(LOG_FILE, 'a') as f:
        f.write(json.dumps(log_entry) + '\n')

def main():
    input_data = json.load(sys.stdin)

    # Log the event
    log_event(input_data)

    # Always allow (observability only)
    print(json.dumps({'decision': 'allow'}))

if __name__ == '__main__':
    main()
```

**Configure for all tools:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "audit-log",
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "python3",
            "args": [".gemini/hooks/audit-log.py"]
          }
        ]
      }
    ]
  }
}
```

### Security Event Logging

**Pattern: Log security decisions**

```bash
#!/bin/bash
# .gemini/hooks/security-audit.sh
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool.name // "unknown"')
FILE=$(echo "$INPUT" | jq -r '.tool.params.path // empty')

SECURITY_LOG=".gemini/security.log"

# Log security-relevant events
if [[ "$FILE" =~ (\.env|credentials|secrets) ]]; then
  echo "[$(date -Iseconds)] SECURITY: Attempt to modify $FILE by $TOOL" >> "$SECURITY_LOG"
  echo '{"decision": "deny", "systemMessage": "Security policy violation logged"}'
  exit 0
fi

echo '{"decision": "allow"}'
```

---

## Policy Enforcement

### Organization-Wide Policies

**Pattern: Enforce coding standards**

```python
#!/usr/bin/env python3
# .gemini/hooks/enforce-policies.py
import json
import sys

POLICIES = {
    'no_console_log': {
        'pattern': r'console\.log',
        'message': 'console.log() not allowed - use proper logging library'
    },
    'no_todo_comments': {
        'pattern': r'//\s*TODO',
        'message': 'TODO comments not allowed - create Jira tickets instead'
    },
    'require_error_handling': {
        'check': lambda content: 'try' in content or 'catch' not in content,
        'message': 'Functions must include error handling'
    }
}

def enforce_policies(content, file_path):
    """Enforce organization policies"""
    import re

    # Only check certain file types
    if not file_path.endswith(('.js', '.ts', '.jsx', '.tsx')):
        return True, None

    for policy_name, policy in POLICIES.items():
        if 'pattern' in policy:
            if re.search(policy['pattern'], content):
                return False, policy['message']

    return True, None

def main():
    input_data = json.load(sys.stdin)

    tool = input_data.get('tool', {})
    if tool.get('name') != 'write_file':
        print(json.dumps({'decision': 'allow'}))
        return

    content = tool.get('params', {}).get('content', '')
    path = tool.get('params', {}).get('path', '')

    is_compliant, message = enforce_policies(content, path)

    if not is_compliant:
        output = {
            'decision': 'deny',
            'systemMessage': f'Policy violation: {message}'
        }
    else:
        output = {'decision': 'allow'}

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

### Approval Workflows

**Pattern: Require manual approval for sensitive operations**

```bash
#!/bin/bash
# .gemini/hooks/approval-workflow.sh
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool.name // empty')

# Operations requiring approval
APPROVAL_REQUIRED=(
  "deploy_application"
  "delete_database"
  "create_user"
  "grant_permissions"
)

for op in "${APPROVAL_REQUIRED[@]}"; do
  if [[ "$TOOL" == "$op" ]]; then
    # Request manual approval
    echo '{"decision": "deny", "systemMessage": "This operation requires manual approval. Run with --approve flag."}'
    exit 0
  fi
done

echo '{"decision": "allow"}'
```

---

## PII and Data Protection

### Redact PII from Responses

**Pattern: Filter personally identifiable information**

```python
#!/usr/bin/env python3
# .gemini/hooks/redact-pii.py
import json
import sys
import re

PII_PATTERNS = [
    (r'\b[\w.-]+@[\w.-]+\.\w+\b', '[EMAIL REDACTED]'),  # Email
    (r'\b\d{3}-\d{2}-\d{4}\b', '[SSN REDACTED]'),  # SSN
    (r'\b\d{3}[- ]?\d{3}[- ]?\d{4}\b', '[PHONE REDACTED]'),  # Phone
    (r'\b\d{16}\b', '[CREDIT_CARD REDACTED]'),  # Credit card
]

def redact_pii(text):
    """Redact PII from text"""
    for pattern, replacement in PII_PATTERNS:
        text = re.sub(pattern, replacement, text)
    return text

def main():
    input_data = json.load(sys.stdin)

    model_output = input_data.get('modelOutput', {})
    text = model_output.get('text', '')

    redacted_text = redact_pii(text)

    output = {
        'decision': 'continue',
        'modelOutput': {
            **model_output,
            'text': redacted_text
        }
    }

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

**Configure for AfterModel:**
```json
{
  "hooks": {
    "AfterModel": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "python3",
            "args": [".gemini/hooks/redact-pii.py"]
          }
        ]
      }
    ]
  }
}
```

---

## Security Checklist

### Hook Development

**Before Deploying Hooks:**

- [ ] Review all hook code for vulnerabilities
- [ ] Test with malicious inputs
- [ ] Validate JSON parsing is safe
- [ ] Check for injection vulnerabilities
- [ ] Ensure proper error handling
- [ ] Verify timeouts are set
- [ ] Test fail-open behavior
- [ ] Add comprehensive logging
- [ ] Document security decisions

### Production Deployment

**Security Hardening:**

- [ ] Use principle of least privilege
- [ ] Enable audit logging
- [ ] Set appropriate timeouts
- [ ] Validate all inputs
- [ ] Sanitize all outputs
- [ ] Review logs regularly
- [ ] Update hooks regularly
- [ ] Version control all hooks
- [ ] Test in staging first
- [ ] Have rollback plan

---

## Common Security Patterns

### Pattern: Defense in Depth

**Multiple layers of security**

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "layer-1-file-path-validation",
        "matcher": "write_file",
        "hooks": [{"type": "command", "command": "bash", "args": ["./check-path.sh"]}]
      },
      {
        "name": "layer-2-content-scanning",
        "matcher": "write_file",
        "hooks": [{"type": "command", "command": "python3", "args": ["./scan-content.py"]}]
      },
      {
        "name": "layer-3-audit-logging",
        "matcher": "*",
        "hooks": [{"type": "command", "command": "python3", "args": ["./audit.py"]}]
      }
    ]
  }
}
```

### Pattern: Allowlist > Denylist

**Prefer allowlisting over denylisting**

```python
# ✅ GOOD - Allowlist approach
ALLOWED_TOOLS = ['read_file', 'write_file', 'execute_npm']

def is_allowed(tool_name):
    return tool_name in ALLOWED_TOOLS

# ❌ BAD - Denylist approach (easy to bypass)
BLOCKED_TOOLS = ['delete_database', 'sudo_command']

def is_blocked(tool_name):
    return tool_name in BLOCKED_TOOLS
```

### Pattern: Fail Securely

**Fail closed for security-critical operations**

```python
def validate_security(tool):
    try:
        # Security-critical validation
        result = security_check(tool)

        if not result:
            # Fail closed - deny by default
            return {'decision': 'deny', 'systemMessage': 'Security check failed'}

        return {'decision': 'allow'}

    except Exception as e:
        logger.error(f"Security validation error: {e}")
        # Fail closed for security-critical operations
        return {'decision': 'deny', 'systemMessage': 'Security validation error'}
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Hook events for security
- [Scripting & Execution](./scripting.md) - Writing secure hooks
- [Environment & Context](./environment-context.md) - Security-relevant context
- [Examples](./examples.md) - More security examples

---

**Document Version:** 1.0
**Last Updated:** 2026-02-21
**Sources:** Official documentation, security best practices
