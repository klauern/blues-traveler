# GitHub Copilot Hooks - Scripting & Execution

**Version:** 1.0
**Last Updated:** 2026-02-21

---

## Supported Languages

### Native Support

**bash (Unix/Linux/macOS):**
```json
{
  "bash": "./scripts/hook.sh"
}
```

**PowerShell (Windows):**
```json
{
  "powershell": "./scripts/hook.ps1"
}
```

### Via Generic Command

**Node.js:**
```json
{
  "command": "node ./scripts/hook.js"
}
```

**Python:**
```json
{
  "command": "python ./scripts/hook.py"
}
```

**Any language with interpreter:**
```json
{
  "command": "/path/to/interpreter ./scripts/hook.ext"
}
```

---

## Exit Codes & Control Flow

### Exit Code Meanings

| Exit Code | Meaning | Effect |
|-----------|---------|--------|
| `0` | Success | Parse stdout (preToolUse), continue (others) |
| `2` | Blocking error | Stop and report to model |
| `1` | Failure | Log warning, continue (non-blocking) |
| `3-255` | Failure | Log warning, continue (non-blocking) |

### Examples

**Success (allow):**
```bash
#!/bin/bash
echo '{"permissionDecision":"allow"}' | jq -c
exit 0
```

**Success (deny):**
```bash
#!/bin/bash
echo '{"permissionDecision":"deny","permissionDecisionReason":"Blocked"}' | jq -c
exit 0
```

**Blocking error:**
```bash
#!/bin/bash
echo "Critical validation error" >&2
exit 2
```

**Non-blocking warning:**
```bash
#!/bin/bash
echo "Warning: unexpected input" >&2
exit 1
```

---

## Input/Output Protocol

### Reading Input (stdin)

**bash:**
```bash
#!/bin/bash
INPUT=$(cat)
FIELD=$(echo "$INPUT" | jq -r '.fieldName')
```

**PowerShell:**
```powershell
#!/usr/bin/env pwsh
$input = $input | ConvertFrom-Json
$field = $input.fieldName
```

**Python:**
```python
#!/usr/bin/env python3
import json
import sys

input_data = json.load(sys.stdin)
field = input_data.get('fieldName')
```

**Node.js:**
```javascript
#!/usr/bin/env node
const fs = require('fs');

const input = JSON.parse(fs.readFileSync(0, 'utf-8'));
const field = input.fieldName;
```

### Writing Output (stdout)

**preToolUse hooks only** - return JSON decision:

```bash
#!/bin/bash
echo '{"permissionDecision":"allow"}' | jq -c
```

Or with reason:
```bash
echo '{
  "permissionDecision": "deny",
  "permissionDecisionReason": "Explanation shown to user"
}' | jq -c
```

**All other hooks** - output is ignored.

### Error Messages (stderr)

```bash
#!/bin/bash
echo "Debug: Processing command..." >&2
echo "Warning: Unusual pattern detected" >&2
```

stderr is logged but doesn't affect execution (unless exit code 2).

---

## Error Handling

### Validate Input

```bash
#!/bin/bash
INPUT=$(cat)

# Validate JSON
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON" >&2
  exit 2  # Blocking error
fi

# Validate required fields
if ! echo "$INPUT" | jq -e '.toolName' > /dev/null 2>&1; then
  echo "Missing toolName" >&2
  exit 1  # Non-blocking warning
fi
```

### Handle Missing Dependencies

```bash
#!/bin/bash

# Check for jq
if ! command -v jq &> /dev/null; then
  echo "ERROR: jq is required but not installed" >&2
  exit 2
fi

# Check for required files
if [ ! -f "$CONFIG_FILE" ]; then
  echo "WARNING: Config file not found, using defaults" >&2
fi
```

### Timeout Handling

```bash
#!/bin/bash

# Run expensive operation with timeout
timeout 3s expensive_validation || {
  echo "Validation timeout, failing open" >&2
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
}
```

---

## Timeouts & Resource Limits

### Timeout Configuration

```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./hook.sh",
      "timeoutSec": 5
    }]
  }
}
```

**Guidelines:**
- Security checks (preToolUse): 3-5 seconds
- Logging (postToolUse): 2-3 seconds
- Setup (sessionStart): 30-60 seconds
- Cleanup (sessionEnd): 30-90 seconds

### Handling Timeouts

If hook exceeds timeout:
- Process is terminated (SIGTERM, then SIGKILL)
- Treated as non-blocking failure
- Error logged
- Execution continues with default behavior

### Resource Management

**Memory:**
```bash
#!/bin/bash
# Stream large data instead of loading into memory
jq -c '.toolArgs.command' | grep -E "$PATTERN"
```

**CPU:**
```bash
#!/bin/bash
# Avoid expensive operations
# ❌ Bad
find / -name "*.log" -exec grep "pattern" {} \;

# ✅ Good
echo "$INPUT" | jq -r '.field' | grep -q "pattern"
```

---

## Best Practices

### Performance

**1. Keep scripts fast:**
```bash
#!/bin/bash
# Quick validation only
COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command')
if echo "$COMMAND" | grep -qE "$DANGEROUS_PATTERN"; then
  echo '{"permissionDecision":"deny"}' | jq -c
  exit 0
fi
echo '{"permissionDecision":"allow"}' | jq -c
```

**2. Cache expensive operations:**
```bash
#!/bin/bash
HASH=$(echo "$INPUT" | sha256sum | cut -d' ' -f1)
CACHE="/tmp/hook-cache-$HASH"

if [ -f "$CACHE" ] && [ $(($(date +%s) - $(stat -f %m "$CACHE"))) -lt 300 ]; then
  cat "$CACHE"
  exit 0
fi

RESULT=$(expensive_check)
echo "$RESULT" | tee "$CACHE"
```

**3. Defer heavy work:**
```bash
#!/bin/bash
# Log to queue for async processing
echo "$INPUT" >> /var/log/copilot-queue.jsonl

# Quick decision
echo '{"permissionDecision":"allow"}' | jq -c
exit 0
```

### Security

**1. Validate and sanitize:**
```bash
#!/bin/bash
# Extract safely
COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command // empty')
if [ -z "$COMMAND" ]; then
  echo '{"permissionDecision":"allow"}' | jq -c
  exit 0
fi

# Don't use eval or other dangerous patterns
```

**2. Avoid logging secrets:**
```bash
#!/bin/bash
# Sanitize before logging
SAFE_CMD=$(echo "$COMMAND" | sed -E 's/(password|token|key)=[^ ]+/\1=****/gi')
echo "Command: $SAFE_CMD" >&2
```

**3. Use safe patterns:**
```bash
#!/bin/bash
# ✅ Good - Pattern matching
if echo "$COMMAND" | grep -qE "$PATTERN"; then
  # ...
fi

# ❌ Bad - Command injection risk
if eval "$COMMAND" | grep -q "something"; then
  # ...
fi
```

### Reliability

**1. Handle errors gracefully:**
```bash
#!/bin/bash
set -e  # Exit on error

INPUT=$(cat) || {
  echo "Failed to read input" >&2
  exit 2
}

# Validate and continue...
```

**2. Provide clear error messages:**
```bash
#!/bin/bash
echo '{
  "permissionDecision": "deny",
  "permissionDecisionReason": "Command contains dangerous pattern: rm -rf. Please review and run manually if needed."
}' | jq -c
```

**3. Log for debugging:**
```bash
#!/bin/bash
LOG_FILE="/var/log/copilot-hooks.log"

{
  echo "$(date -Iseconds) [preToolUse] Processing..."
  echo "Tool: $TOOL_NAME"
  echo "Decision: $DECISION"
} >> "$LOG_FILE" 2>&1
```

---

## Platform-Specific Scripts

### bash (Unix/Linux/macOS)

```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

if [ "$TOOL_NAME" = "bash" ]; then
  COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command')
  if echo "$COMMAND" | grep -qE "(rm -rf|sudo)"; then
    echo '{"permissionDecision":"deny","permissionDecisionReason":"Dangerous command"}' | jq -c
    exit 0
  fi
fi

echo '{"permissionDecision":"allow"}' | jq -c
```

### PowerShell (Windows)

```powershell
#!/usr/bin/env pwsh
param()

$ErrorActionPreference = 'Stop'

$input = $input | ConvertFrom-Json
$toolName = $input.toolName

if ($toolName -eq 'bash') {
    $command = ($input.toolArgs | ConvertFrom-Json).command
    if ($command -match '(rm -rf|sudo)') {
        @{
            permissionDecision = 'deny'
            permissionDecisionReason = 'Dangerous command'
        } | ConvertTo-Json -Compress
        exit 0
    }
}

@{permissionDecision = 'allow'} | ConvertTo-Json -Compress
```

### Cross-Platform (Node.js)

```javascript
#!/usr/bin/env node
const fs = require('fs');

try {
  const input = JSON.parse(fs.readFileSync(0, 'utf-8'));

  if (input.toolName === 'bash') {
    const args = JSON.parse(input.toolArgs);
    if (/(rm -rf|sudo)/.test(args.command)) {
      console.log(JSON.stringify({
        permissionDecision: 'deny',
        permissionDecisionReason: 'Dangerous command'
      }));
      process.exit(0);
    }
  }

  console.log(JSON.stringify({ permissionDecision: 'allow' }));
} catch (err) {
  console.error('Error:', err.message);
  process.exit(2);
}
```

---

## Testing Hooks Locally

### Create Test Input

```bash
cat > test-input.json <<'EOF'
{
  "timestamp": 1708560000000,
  "cwd": "/workspace",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"rm -rf /\"}"
}
EOF
```

### Test Hook

```bash
cat test-input.json | .github/hooks/scripts/security-check.sh
```

### Validate Output

```bash
OUTPUT=$(cat test-input.json | .github/hooks/scripts/security-check.sh)
echo "$OUTPUT" | jq empty  # Validate JSON
echo "$OUTPUT" | jq -r '.permissionDecision'  # Check decision
```

### Automated Testing

```bash
#!/bin/bash
# test-hooks.sh

echo "Testing security-check.sh..."

# Test 1: Dangerous command should be denied
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"rm -rf /\"}"}' | \
  .github/hooks/scripts/security-check.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "deny"' > /dev/null; then
  echo "✅ Test 1 passed"
else
  echo "❌ Test 1 failed"
  exit 1
fi

# Test 2: Safe command should be allowed
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"ls -la\"}"}' | \
  .github/hooks/scripts/security-check.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "allow"' > /dev/null; then
  echo "✅ Test 2 passed"
else
  echo "❌ Test 2 failed"
  exit 1
fi

echo "All tests passed!"
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Hook input schemas
- [Environment & Context](./environment-context.md) - Accessing context data
- [Examples](./examples.md) - Working patterns
- [Troubleshooting](./troubleshooting.md) - Debug issues

---

**Sources:**
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [Using Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

---

**Document Version:** 1.0
**Status:** ✅ Complete
