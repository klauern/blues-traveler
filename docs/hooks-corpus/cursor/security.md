# Cursor IDE Security & Safety

**Security model, permissions, and best practices for Cursor hooks**

---

## Permission Model

### Hook Execution Context

**Hooks run with FULL user permissions:**
- ✅ Can read any file user can read
- ✅ Can write any file user can write
- ✅ Can execute any command user can execute
- ✅ Full network access
- ❌ **NO sandboxing or isolation**
- ❌ **NO privilege separation**

**Critical:** Hooks execute outside Cursor's agent sandbox

```
┌─────────────────────────────────────┐
│  Cursor IDE Process                 │
│  ┌────────────────────────────────┐ │
│  │  Agent (Sandboxed)             │ │
│  │  ├─ Limited filesystem         │ │
│  │  ├─ Restricted network         │ │
│  │  └─ Controlled syscalls        │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌────────────────────────────────┐ │
│  │  Hooks (NOT Sandboxed)         │ │
│  │  ├─ Full filesystem access     │ │
│  │  ├─ Full network access        │ │
│  │  └─ Full user permissions      │ │
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

---

## Sandboxing

### Agent Sandbox (for comparison)

Cursor's AI agent runs in a sandboxed environment:

**macOS:**
- Seatbelt (sandbox-exec)
- Limited filesystem access
- Network restrictions

**Linux:**
- Landlock LSM
- seccomp syscall filtering
- Namespace isolation

**Windows:**
- Linux sandbox via WSL2
- Similar restrictions as Linux

### Hook Execution (NOT Sandboxed)

**Hooks bypass ALL agent sandbox restrictions:**
- Run in parent process context
- Full user privileges
- No capability restrictions
- No syscall filtering

**Security Implication:** Malicious hooks can compromise entire system

---

## Blocking Dangerous Operations

### Security Validation Examples

**Block Destructive Commands:**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

# Block filesystem destruction
if echo "$command" | grep -qE '(rm -rf /|mkfs|dd if=.*of=/dev/)'; then
  echo '{
    "permission": "deny",
    "userMessage": "Destructive filesystem operation blocked",
    "agentMessage": "Blocked pattern: filesystem destruction"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Block Privilege Escalation:**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

# Block sudo and su
if echo "$command" | grep -qE '(^sudo |^su | sudo | su )'; then
  echo '{
    "permission": "deny",
    "userMessage": "Privilege escalation not allowed",
    "agentMessage": "Blocked: sudo/su commands are restricted"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Block Download-and-Execute:**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

# Block curl/wget piped to sh
if echo "$command" | grep -qE '(curl.*\|.*sh|wget.*\|.*sh)'; then
  echo '{
    "permission": "deny",
    "userMessage": "Download-and-execute blocked",
    "agentMessage": "Security: curl/wget | sh is dangerous"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Block Dangerous Git Operations:**
```bash
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

# Block force push to main/master
if echo "$command" | grep -qE 'git push.*--force.*(main|master)'; then
  echo '{
    "permission": "deny",
    "userMessage": "Force push to main branch blocked"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

## Secret Management

### Detecting Secrets in Files

**GitHub Token Detection:**
```bash
#!/bin/bash
input=$(cat)
content=$(echo "$input" | jq -r '.content')

# Detect GitHub tokens
if echo "$content" | grep -qE 'gh[ps]_[A-Za-z0-9]{36}'; then
  echo '{
    "permission": "deny",
    "userMessage": "File contains GitHub API token"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**AWS Credentials:**
```bash
#!/bin/bash
input=$(cat)
content=$(echo "$input" | jq -r '.content')

# Detect AWS keys
if echo "$content" | grep -qE 'AKIA[0-9A-Z]{16}'; then
  echo '{
    "permission": "deny",
    "userMessage": "File contains AWS access key"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Generic Secret Patterns:**
```bash
#!/bin/bash
input=$(cat)
content=$(echo "$input" | jq -r '.content')

# Detect common secret patterns
patterns=(
  'api[_-]?key["\s]*[:=]'
  'secret[_-]?key["\s]*[:=]'
  'password["\s]*[:=]'
  'private[_-]?key'
  'BEGIN (RSA |DSA |EC |OPENSSH )?PRIVATE KEY'
)

for pattern in "${patterns[@]}"; do
  if echo "$content" | grep -qiE "$pattern"; then
    echo '{
      "permission": "deny",
      "userMessage": "File contains potential secret"
    }'
    exit 0
  fi
done

echo '{"permission":"allow"}'
```

### Best Practices for Secrets

**DO:**
- ✅ Use environment variables for secrets
- ✅ Use OS keychain (macOS Keychain, Windows Credential Manager)
- ✅ Use secret management tools (1Password, Vault)
- ✅ Never hardcode secrets in hook scripts

**DON'T:**
- ❌ Hardcode API keys in hooks
- ❌ Store secrets in `.cursor/hooks.json`
- ❌ Log secrets to files
- ❌ Include secrets in error messages

**Example (Secure Secret Access):**
```bash
#!/bin/bash
# Retrieve secret from macOS Keychain
API_KEY=$(security find-generic-password -s my-service -w)

# Use API_KEY without logging it
curl -H "Authorization: Bearer $API_KEY" https://api.example.com/validate
```

---

## Audit Logging

### Comprehensive Audit Trail

**Log All Hook Executions:**
```bash
#!/bin/bash
input=$(cat)

# Extract key fields
timestamp=$(date -Iseconds)
hook_event=$(echo "$input" | jq -r '.hook_event_name')
conversation_id=$(echo "$input" | jq -r '.conversation_id')
generation_id=$(echo "$input" | jq -r '.generation_id')

# Log to audit file
{
  echo "=== Audit Entry ==="
  echo "Timestamp: $timestamp"
  echo "Event: $hook_event"
  echo "Conversation: $conversation_id"
  echo "Generation: $generation_id"
  echo "Full Input: $input"
  echo ""
} >> /var/log/cursor-hooks-audit.log

# Continue with hook logic...
echo '{"permission":"allow"}'
```

**Structured JSON Audit Log:**
```bash
#!/bin/bash
input=$(cat)

# Structured audit entry
jq -n \
  --arg ts "$(date -Iseconds)" \
  --argjson input "$input" \
  '{
    timestamp: $ts,
    event: $input.hook_event_name,
    conversation_id: $input.conversation_id,
    generation_id: $input.generation_id,
    payload: $input
  }' >> /var/log/cursor-hooks-audit.jsonl
```

---

## Permission Modes Explained

### allow

**Meaning:** Permit the action
**Use When:** Action is safe and approved
**User Experience:** No interruption, action proceeds

```bash
echo '{"permission":"allow"}'
```

### deny

**Meaning:** Block the action
**Use When:** Action violates policy or is dangerous
**User Experience:** Action blocked, agent receives rejection message

```bash
echo '{
  "permission": "deny",
  "userMessage": "Shown to user in UI",
  "agentMessage": "Technical details for AI agent"
}'
```

### ask (beforeShellExecution and beforeMCPExecution only)

**Meaning:** Request user confirmation
**Use When:** Action requires manual approval
**User Experience:** Cursor displays approval dialog
**Unique to Cursor:** Not available in all IDE hook systems

```bash
echo '{
  "permission": "ask",
  "userMessage": "This command requires approval",
  "agentMessage": "User approval needed for: $command"
}'
```

**Example Flow:**
1. Hook returns `ask`
2. Cursor shows dialog: "Command requires approval: npm install"
3. User clicks "Approve" or "Deny"
4. If approved: Command executes
5. If denied: Command blocked

---

## Common Security Vulnerabilities

### Vulnerability 1: Command Injection

**Vulnerable:**
```bash
#!/bin/bash
command=$(echo "$input" | jq -r '.command')
eval "$command"  # ❌ DANGEROUS
```

**Secure:**
```bash
#!/bin/bash
command=$(echo "$input" | jq -r '.command')
# Validate command, don't execute it
echo '{"permission":"allow"}'
```

### Vulnerability 2: Path Traversal

**Vulnerable:**
```bash
#!/bin/bash
file_path=$(echo "$input" | jq -r '.file_path')
cat "$file_path"  # ❌ Could read /etc/passwd
```

**Secure:**
```bash
#!/bin/bash
file_path=$(echo "$input" | jq -r '.file_path')
workspace_roots=$(echo "$input" | jq -r '.workspace_roots[0]')

# Validate file is within workspace
if [[ "$file_path" != "$workspace_roots"* ]]; then
  echo '{"permission":"deny","userMessage":"File outside workspace"}'
  exit 0
fi

echo '{"permission":"allow"}'
```

### Vulnerability 3: Secret Leakage in Logs

**Vulnerable:**
```bash
#!/bin/bash
echo "Processing command: $command" >> /tmp/hooks.log  # ❌ Might log secrets
```

**Secure:**
```bash
#!/bin/bash
# Redact secrets before logging
safe_command=$(echo "$command" | sed 's/password=[^ ]*/password=REDACTED/g')
echo "Processing command: $safe_command" >> /tmp/hooks.log
```

### Vulnerability 4: Untrusted Input

**Vulnerable:**
```bash
#!/bin/bash
# Blindly trust input
curl "https://api.example.com?cmd=$command"  # ❌ No validation
```

**Secure:**
```bash
#!/bin/bash
# Validate and sanitize
if ! echo "$command" | grep -qE '^[a-zA-Z0-9 .-]+$'; then
  echo '{"permission":"deny","userMessage":"Invalid command format"}'
  exit 0
fi

# URL-encode before using
safe_command=$(echo "$command" | jq -sRr @uri)
curl "https://api.example.com?cmd=$safe_command"
```

---

## Trusted Sources Only

### Verifying Hook Scripts

**Before enabling ANY hook:**
1. ✅ Read the entire script
2. ✅ Understand what it does
3. ✅ Check for malicious patterns:
   - Network calls to unknown domains
   - File modifications outside workspace
   - Credential theft attempts
   - Obfuscated code
   - Command execution via eval
4. ✅ Review dependencies
5. ✅ Test in isolated environment

**Red Flags:**
- ❌ Obfuscated or base64-encoded code
- ❌ Unexpected network calls
- ❌ Reading sensitive files (.ssh, .aws, .env)
- ❌ Modifying system files
- ❌ Using `eval`, `exec`, or `source` with untrusted input

### Safe Sources

**Trusted:**
- ✅ Official Cursor documentation examples
- ✅ Well-maintained GitHub repos (1Password, Endor Labs)
- ✅ Internal team repositories
- ✅ Hooks you write yourself

**Untrusted:**
- ❌ Random internet scripts
- ❌ Unverified third-party hooks
- ❌ Hooks from unknown sources
- ❌ Hooks without code review

---

## Enterprise Security Integrations

### 1Password Integration

```bash
#!/bin/bash
# Validate secrets using 1Password
input=$(cat)
content=$(echo "$input" | jq -r '.content')

# Use 1Password CLI to check for secrets
if echo "$content" | op inject; then
  echo '{"permission":"deny","userMessage":"1Password secret detected"}'
  exit 0
fi

echo '{"permission":"allow"}'
```

### Endor Labs Malware Detection

```bash
#!/bin/bash
# Scan package.json for malware
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

if [[ "$file_path" == *"package.json" ]]; then
  # Run Endor Labs scan
  endor scan "$file_path" || {
    echo '{"permission":"deny","userMessage":"Malware detected in dependencies"}'
    exit 0
  }
fi

echo '{"permission":"allow"}'
```

### Semgrep Security Scanning

```bash
#!/bin/bash
# Scan code changes with Semgrep
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

# Run Semgrep on edited file
semgrep --config auto "$file_path" --json > /tmp/semgrep-results.json

# Check for findings
if jq -e '.results | length > 0' /tmp/semgrep-results.json; then
  echo '{
    "permission": "deny",
    "userMessage": "Security issues detected by Semgrep"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

## Security Checklist

### For Hook Developers

- [ ] Hook validates all input from JSON
- [ ] Hook sanitizes data before logging
- [ ] Hook uses absolute paths or validates relative paths
- [ ] Hook doesn't execute untrusted commands
- [ ] Hook handles errors gracefully
- [ ] Hook doesn't leak secrets in logs or errors
- [ ] Hook has timeouts for external calls
- [ ] Hook follows principle of least privilege
- [ ] Hook is code-reviewed
- [ ] Hook has security tests

### For Hook Users

- [ ] Reviewed hook source code
- [ ] Understand what hook does
- [ ] Verified hook source is trusted
- [ ] Tested hook in safe environment
- [ ] Hooks are version-controlled
- [ ] Audit logs are enabled
- [ ] Regular security reviews scheduled

---

**Sources:**
- [Cursor Official Documentation](https://cursor.com/docs/agent/hooks)
- [1Password Cursor Hooks](https://github.com/1Password/cursor-hooks)
- [Endor Labs Examples](https://github.com/endorlabs/cursor-hook-examples)
- [GitButler Security Guide](https://blog.gitbutler.com/cursor-hooks-deep-dive)
