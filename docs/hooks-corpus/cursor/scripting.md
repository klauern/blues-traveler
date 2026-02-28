# Cursor IDE Scripting & Execution

**Languages, exit codes, and error handling for Cursor hooks**

---

## Supported Languages

### Native Support (Any Language)

Cursor hooks run as **shell commands**, supporting any executable:

**Bash/Shell:**
```json
{ "command": "./hooks/check.sh" }
```

**TypeScript (Bun):**
```json
{ "command": "bun run hooks/check.ts" }
```

**Python:**
```json
{ "command": "python hooks/check.py" }
```

**Node.js:**
```json
{ "command": "node hooks/check.js" }
```

**Compiled Binary:**
```json
{ "command": "./hooks/check-binary" }
```

**Ruby, Go, Rust, etc.:**
```json
{ "command": "ruby hooks/check.rb" }
{ "command": "go run hooks/check.go" }
{ "command": "./target/release/check" }
```

### Language-Specific Considerations

**Bash:**
- ✅ Fast startup (~10-20ms)
- ✅ Universal availability
- ✅ Simple JSON parsing with `jq`
- ❌ Limited type safety
- ❌ Error-prone string manipulation

**TypeScript (Bun):**
- ✅ Type safety via `cursor-hooks` package
- ✅ Modern syntax
- ✅ Fast runtime (Bun)
- ❌ Slower startup (~50-200ms)
- ❌ Requires Bun installation

**Python:**
- ✅ Rich ecosystem
- ✅ Type hints available (`py-cursor-hooks`)
- ✅ Good for complex logic
- ❌ Slower startup (~100-300ms)
- ❌ Dependency management

**Compiled (Go/Rust):**
- ✅ Fastest execution (~5-10ms)
- ✅ Type safety
- ✅ No runtime dependency
- ❌ Requires compilation step
- ❌ Larger binaries

---

## Exit Codes & Control Flow

### Exit Code Meaning

| Exit Code | JSON Output | Result |
|-----------|-------------|--------|
| 0 | Valid JSON | Process JSON response |
| 0 | No JSON | Implicit allow (silent success) |
| 0 | Invalid JSON | Block with error |
| Non-zero | Valid JSON | Process JSON (exit code ignored) |
| Non-zero | No JSON | Block with error alert |
| Non-zero | Invalid JSON | Block with error |

### Exit Code Examples

**Success with explicit allow:**
```bash
#!/bin/bash
echo '{"permission":"allow"}'
exit 0
```

**Success with silent allow:**
```bash
#!/bin/bash
exit 0  # No output = implicit allow
```

**Failure with deny:**
```bash
#!/bin/bash
echo '{"permission":"deny","userMessage":"Blocked"}'
exit 1  # Exit code ignored if JSON valid
```

**Crash without JSON:**
```bash
#!/bin/bash
exit 1  # No JSON = block with error
```

### Best Practices

**DO:**
- ✅ Always output valid JSON for clarity
- ✅ Use exit 0 for success paths
- ✅ Include helpful messages in JSON
- ✅ Log errors to stderr or files

**DON'T:**
- ❌ Rely on implicit allow behavior
- ❌ Output debug info to stdout
- ❌ Mix stdout logging with JSON output
- ❌ Use non-zero exits without JSON

---

## Error Handling

### Stdout vs Stderr

**Stdout:** JSON output only
```bash
echo '{"permission":"allow"}' # ✅ Stdout
```

**Stderr:** Debug/logging
```bash
echo "Debug: checking command" >&2  # ✅ Stderr
echo '{"permission":"allow"}'       # ✅ Stdout
```

**Anti-Pattern:**
```bash
echo "Debug: checking command"  # ❌ Corrupts JSON
echo '{"permission":"allow"}'
# Result: Invalid JSON error
```

### Error Reporting

**Bash:**
```bash
#!/bin/bash
set -euo pipefail  # Exit on error, undefined vars, pipe failures

input=$(cat)
command=$(echo "$input" | jq -r '.command') || {
  echo '{"permission":"deny","userMessage":"Invalid input"}' >&2
  exit 1
}

# Process command...
```

**TypeScript:**
```typescript
try {
  const input = await Bun.stdin.json();
  // Process input...
  console.log(JSON.stringify({ permission: "allow" }));
} catch (error) {
  console.error("Hook error:", error);
  console.log(JSON.stringify({
    permission: "deny",
    userMessage: "Hook execution failed"
  }));
  process.exit(1);
}
```

**Python:**
```python
import sys
import json

try:
    input_data = json.load(sys.stdin)
    # Process input...
    print(json.dumps({"permission": "allow"}))
except Exception as e:
    print(json.dumps({
        "permission": "deny",
        "userMessage": "Hook execution failed"
    }))
    sys.exit(1)
```

---

## Timeouts & Resource Limits

### Timeout Configuration

**In hooks.json (undocumented):**
```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hook.sh", "timeout": 5000 }
    ]
  }
}
```

**Timeout behavior:** Unknown (testing needed)

### Resource Limits

**Undocumented:**
- No official memory limits
- No CPU limits
- No payload size limits

**Community Recommendations:**
- Timeout: < 2 seconds (normal), < 10 seconds (max)
- Memory: < 512MB
- JSON payload: < 10MB

### Handling Long Operations

**Anti-Pattern (Blocking):**
```bash
#!/bin/bash
# This blocks agent for 30 seconds!
curl -m 30 https://api.example.com/slow-endpoint
echo '{"permission":"allow"}'
```

**Better (Background):**
```bash
#!/bin/bash
# Quick validation, async logging
echo '{"permission":"allow"}'

# Background audit (doesn't block agent)
(curl -X POST https://api.example.com/audit -d "$input" &)
```

---

## JSON Generation

### Bash (Heredoc)

```bash
#!/bin/bash
cat << 'JSON'
{
  "permission": "deny",
  "userMessage": "Command blocked",
  "agentMessage": "Matched dangerous pattern: rm -rf"
}
JSON
```

### Bash (jq)

```bash
#!/bin/bash
jq -n --arg msg "Blocked" '{
  permission: "deny",
  userMessage: $msg
}'
```

### TypeScript

```typescript
const output = {
  permission: "allow" as const,
  userMessage: "Approved",
};

console.log(JSON.stringify(output));
```

### Python

```python
import json

output = {
    "permission": "allow",
    "userMessage": "Approved"
}

print(json.dumps(output))
```

---

## Debugging Techniques

### Log to File

```bash
#!/bin/bash
{
  echo "=== Hook Debug Log ==="
  echo "Timestamp: $(date -Iseconds)"
  echo "Input: $(cat)"
  echo "Command: $command"
  echo "Result: $permission"
} >> /tmp/cursor-hook-debug.log
```

### Test Hooks Manually

```bash
# Test beforeShellExecution hook
echo '{"command":"npm install","hook_event_name":"beforeShellExecution"}' \
  | ./hooks/security-check.sh

# Expected output:
# {"permission":"deny","agentMessage":"npm is not allowed"}
```

### Verbose Mode

```bash
#!/bin/bash
VERBOSE="${CURSOR_HOOK_VERBOSE:-0}"

if [ "$VERBOSE" = "1" ]; then
  echo "Debug: Processing command: $command" >&2
fi
```

### Dry Run Mode

```bash
#!/bin/bash
DRY_RUN="${CURSOR_HOOK_DRY_RUN:-0}"

if [ "$DRY_RUN" = "1" ]; then
  echo "Dry run: Would block command: $command" >&2
  echo '{"permission":"allow"}'  # Always allow in dry run
  exit 0
fi
```

---

## Common Patterns

### Validation with Early Exit

```bash
#!/bin/bash
set -euo pipefail

input=$(cat)

# Quick rejection
command=$(echo "$input" | jq -r '.command')
if echo "$command" | grep -qE '^rm -rf /'; then
  echo '{"permission":"deny","userMessage":"Dangerous command"}'
  exit 0
fi

# More expensive checks...
echo '{"permission":"allow"}'
```

### Multi-Stage Validation

```bash
#!/bin/bash

# Stage 1: Fast regex checks
if echo "$command" | grep -qE '(rm -rf|sudo rm)'; then
  echo '{"permission":"deny"}'
  exit 0
fi

# Stage 2: Moderate complexity
if echo "$command" | grep -q "curl.*sh"; then
  echo '{"permission":"deny"}'
  exit 0
fi

# Stage 3: Expensive (API call)
if curl -s "https://api.example.com/validate" -d "$command" | grep -q "malicious"; then
  echo '{"permission":"deny"}'
  exit 0
fi

echo '{"permission":"allow"}'
```

### Caching Results

```bash
#!/bin/bash
command=$(echo "$input" | jq -r '.command')
cache_key=$(echo -n "$command" | md5sum | cut -d' ' -f1)
cache_file="/tmp/hook-cache-${cache_key}"

# Check cache
if [ -f "$cache_file" ]; then
  cat "$cache_file"
  exit 0
fi

# Expensive validation...
result='{"permission":"allow"}'

# Write cache
echo "$result" > "$cache_file"
echo "$result"
```

---

## SDK-Specific Guidance

### TypeScript (cursor-hooks)

**Installation:**
```bash
cd .cursor
bun init -y
bun add cursor-hooks
```

**Example:**
```typescript
import type {
  BeforeShellExecutionPayload,
  BeforeShellExecutionResponse,
} from "cursor-hooks";

const input: BeforeShellExecutionPayload = await Bun.stdin.json();

const output: BeforeShellExecutionResponse = {
  permission: input.command.includes("danger") ? "deny" : "allow",
  userMessage: input.command.includes("danger")
    ? "Dangerous command blocked"
    : undefined,
};

console.log(JSON.stringify(output));
```

### Python (py-cursor-hooks)

**Installation:**
```bash
pip install py-cursor-hooks
```

**Example:**
```python
import sys
import json
from cursor_hooks import BeforeShellExecutionPayload

input_data = BeforeShellExecutionPayload(**json.load(sys.stdin))

output = {
    "permission": "deny" if "danger" in input_data.command else "allow",
    "userMessage": "Dangerous command blocked" if "danger" in input_data.command else None
}

print(json.dumps(output))
```

---

## Performance Profiling

### Measure Hook Execution Time

```bash
#!/bin/bash
start_time=$(date +%s%N)

# Hook logic here...
input=$(cat)
command=$(echo "$input" | jq -r '.command')
echo '{"permission":"allow"}'

end_time=$(date +%s%N)
elapsed=$((($end_time - $start_time) / 1000000))  # Convert to ms

echo "Hook execution time: ${elapsed}ms" >&2
```

### Log Performance Metrics

```bash
#!/bin/bash
{
  echo "$(date -Iseconds),beforeShellExecution,${elapsed}ms"
} >> /tmp/hook-performance.csv
```

---

**Sources:**
- [Cursor Official Documentation](https://cursor.com/docs/agent/hooks)
- [TypeScript SDK](https://github.com/johnlindquist/cursor-hooks)
- [Python SDK](https://github.com/DevonFulcher/py-cursor-hooks)
- [GitButler Deep Dive](https://blog.gitbutler.com/cursor-hooks-deep-dive)
