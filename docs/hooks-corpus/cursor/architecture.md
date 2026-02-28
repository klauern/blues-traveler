# Cursor IDE Architecture & Internals

**Understanding Cursor's hook execution model and lifecycle**

---

## Execution Architecture

### Process Model

Cursor implements a **subprocess-based execution model** for hooks:

```
┌──────────────────────────────────────┐
│         Cursor IDE (Parent)          │
│  ┌────────────────────────────────┐  │
│  │      AI Agent Process          │  │
│  │                                │  │
│  │  1. Event triggered            │  │
│  │     │                          │  │
│  │     ▼                          │  │
│  │  2. Spawn hook subprocess ────┼──┼───> Hook Script
│  │     │                          │  │         │
│  │     │  3. JSON via stdin       │  │         │
│  │     ├──────────────────────────┼──┼────────>│
│  │     │                          │  │         │
│  │     │  4. JSON via stdout      │  │         │
│  │     │<─────────────────────────┼──┼─────────┤
│  │     │                          │  │         │
│  │     ▼                          │  │         X Process terminates
│  │  5. Process result             │  │
│  │     │                          │  │
│  │     ▼                          │  │
│  │  6. Allow/Deny action          │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

**Key Characteristics:**

- **Isolated execution** - Each hook runs in separate subprocess
- **JSON communication** - stdin for input, stdout for output
- **Synchronous** - Agent waits for hook completion
- **Shell-based** - Hooks invoked via shell command
- **No state persistence** - Each invocation is independent

---

## Hook Lifecycle

### Lifecycle Stages

**1. Event Detection**
- Agent action triggers registered event
- Cursor identifies applicable hooks from `.cursor/hooks.json`
- Prepares event-specific JSON payload

**2. Hook Resolution**
- Load hooks from all levels (system → user → project)
- Resolve command paths relative to `hooks.json` location
- Build execution queue (array order)

**3. Process Spawning**
- Fork subprocess for hook script
- Set up stdin/stdout pipes
- Execute command via shell

**4. Input Transmission**
- Serialize event data to JSON
- Write JSON to hook's stdin
- Close stdin stream

**5. Hook Execution**
- Hook script runs
- Processes JSON input
- Performs validation/checks
- Generates JSON output

**6. Output Processing**
- Read stdout from hook
- Parse JSON response
- Validate against schema
- Extract permission/continue fields

**7. Action Control**
- Apply hook result (allow/deny/ask)
- Display user messages if present
- Send agent messages to AI
- Proceed or block original action

**8. Cleanup**
- Close subprocess
- Release resources
- Log execution (if enabled)

---

## Performance Considerations

### Execution Model

**Synchronous Blocking:**
- Agent waits for hook completion
- No timeout fallback (by default)
- User sees delay if hooks are slow
- Recommendation: Keep hooks under 2 seconds

**Sequential Execution:**
- Multiple hooks run one after another
- Total time = sum of individual hook times
- No parallelization within event

**Process Overhead:**
- Process spawn: ~10-50ms
- Shell initialization: ~10-30ms
- Script parsing (Python/Node): ~50-200ms
- Compiled binaries: ~5-10ms

**Performance Budget Example:**

```
Event: beforeShellExecution
├── Hook 1 (Bash security check): 50ms
├── Hook 2 (Python audit log):    150ms
└── Hook 3 (API call):             1500ms
────────────────────────────────────────
Total latency:                     1700ms  ✅ Acceptable

Event: beforeReadFile
├── Hook 1 (Secret scanner):       2000ms
├── Hook 2 (External API):         5000ms
└── Hook 3 (Database log):         3000ms
────────────────────────────────────────
Total latency:                     10000ms ❌ Too slow
```

### Performance Best Practices

**DO:**
- ✅ Use compiled binaries when possible
- ✅ Cache expensive computations
- ✅ Optimize regex patterns
- ✅ Exit early on common cases
- ✅ Use local file operations over network calls
- ✅ Profile hooks during development

**DON'T:**
- ❌ Make blocking HTTP requests
- ❌ Query external databases synchronously
- ❌ Perform heavy computations
- ❌ Load large dependencies unnecessarily
- ❌ Use slow interpreted languages for hot paths

**Optimization Example:**

```bash
#!/bin/bash
# SLOW: Check every command against API
curl -X POST api.example.com/validate -d "$command"

# FAST: Check locally first, API only if needed
if echo "$command" | grep -qE '^(rm|sudo|mkfs)'; then
  # Only call API for suspicious commands
  curl -X POST api.example.com/validate -d "$command"
else
  echo '{"permission":"allow"}'
fi
```

---

## Multi-Level Configuration

### Configuration Hierarchy

Cursor supports hooks at three levels:

```
1. System-wide
   └── Applies to all users, all projects
       Example: Organization security policies

2. User-specific
   └── Applies to one user, all projects
       Example: Personal preferences

3. Project-specific
   └── Applies to one project only
       Example: Team-wide standards
```

### Precedence and Merging

**Execution Order:** System → User → Project

**Merge Behavior:**
- All applicable hooks execute (no override)
- Later hooks can override earlier decisions
- `deny` typically overrides `allow`

**Example:**

```
System hooks.json:
{
  "beforeShellExecution": [
    { "command": "/usr/local/bin/security-baseline.sh" }
  ]
}

User hooks.json:
{
  "beforeShellExecution": [
    { "command": "~/.cursor/hooks/my-preferences.sh" }
  ]
}

Project hooks.json:
{
  "beforeShellExecution": [
    { "command": "./hooks/project-policy.sh" }
  ]
}

Actual execution order:
1. /usr/local/bin/security-baseline.sh  (system)
2. ~/.cursor/hooks/my-preferences.sh     (user)
3. ./hooks/project-policy.sh             (project)
```

---

## Security Model

### Sandboxing

**Hook Execution Context:**
- Hooks run **outside** Cursor's agent sandbox
- Hooks execute with **full user permissions**
- No isolation or privilege separation
- Can access filesystem, network, system resources

**Agent Sandbox (for comparison):**
- Shell commands run in sandboxed environment:
  - macOS: Seatbelt (sandbox-exec)
  - Linux: Landlock and seccomp
  - Windows: Linux sandbox via WSL2
- Hooks bypass this sandbox entirely

**Security Implications:**

```
┌─────────────────────────────────────┐
│  Cursor IDE Process                 │
│  ┌────────────────────────────────┐ │
│  │  Agent (Sandboxed)             │ │
│  │  - Limited filesystem access   │ │
│  │  - Restricted network          │ │
│  │  - Controlled syscalls         │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌────────────────────────────────┐ │
│  │  Hooks (NOT Sandboxed)         │ │
│  │  - Full filesystem access      │ │
│  │  - Full network access         │ │
│  │  - Full user permissions       │ │
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### Permission Model

**No Permission System:**
- Hooks do not request permissions
- No capability restrictions
- Trust is all-or-nothing

**Best Practices:**
1. Only run hooks from trusted sources
2. Review all hook code before enabling
3. Avoid `sudo` in hooks
4. Use principle of least privilege
5. Log all hook executions

---

## Error Handling

### Exit Code Behavior

| Exit Code | JSON Output | Result |
|-----------|-------------|--------|
| 0 | Valid JSON | Process JSON (allow/deny based on fields) |
| 0 | No JSON | Implicit allow (silent success) |
| 0 | Invalid JSON | Block with "hook broken" error |
| Non-zero | Valid JSON | Process JSON (can still allow) |
| Non-zero | No JSON | Block with error alert |
| Non-zero | Invalid JSON | Block with "hook broken" error |

**Example Scenarios:**

```bash
# Scenario 1: Success with explicit allow
echo '{"permission":"allow"}'
exit 0
# Result: Allow

# Scenario 2: Success with no output (silent)
exit 0
# Result: Allow (implicit)

# Scenario 3: Failure with deny message
echo '{"permission":"deny","userMessage":"Blocked"}'
exit 1
# Result: Deny (exit code doesn't matter if JSON valid)

# Scenario 4: Crash without JSON
exit 1
# Result: Block with error

# Scenario 5: Invalid JSON
echo '{"permission":"allow"'  # Missing closing brace
exit 0
# Result: Block with "hook broken" error
```

### Timeout Behavior

**Documented:** Timeout configurable in hooks.json

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hook.sh", "timeout": 5000 }
    ]
  }
}
```

**Undocumented:**
- ❌ What happens when timeout expires?
- ❌ Default timeout value?
- ❌ Maximum timeout limit?
- ❌ Can timeout be disabled?

**Testing Needed:** See [research-notes/cursor/gaps.md](../../../research-notes/cursor/gaps.md)

---

## Conversation & Generation IDs

### ID Semantics

**Conversation ID:**
- Unique per chat session
- Persists across multiple prompts
- Resets when starting new chat
- Use for: Session-level state

**Generation ID:**
- Unique per prompt within conversation
- New ID for each user message
- Use for: Prompt-level state

**Example Timeline:**

```
User starts new chat
  └── conversation_id: abc123
      User sends prompt 1
        └── generation_id: gen001
            beforeShellExecution fires
              → conversation_id: abc123
              → generation_id: gen001

      User sends prompt 2 (same chat)
        └── generation_id: gen002
            beforeShellExecution fires
              → conversation_id: abc123  (same)
              → generation_id: gen002    (new)

User starts NEW chat
  └── conversation_id: xyz789  (new)
      User sends prompt 1
        └── generation_id: gen003  (new)
```

### State Management

**Per-Generation State:**

```bash
#!/bin/bash
generation_id=$(echo "$input" | jq -r '.generation_id')

# Write state for this generation
echo "data" > "/tmp/state-${generation_id}.txt"

# Read state in later hook (same generation)
data=$(cat "/tmp/state-${generation_id}.txt")
```

**Per-Conversation State:**

```bash
#!/bin/bash
conversation_id=$(echo "$input" | jq -r '.conversation_id')

# Shared state across all prompts in conversation
echo "$data" >> "/tmp/conversation-${conversation_id}.log"
```

---

## Workspace Roots

### Single Workspace

```json
{
  "workspace_roots": ["/Users/username/projects/myapp"]
}
```

### Multi-Root Workspace (VS Code)

```json
{
  "workspace_roots": [
    "/Users/username/projects/frontend",
    "/Users/username/projects/backend",
    "/Users/username/projects/shared-lib"
  ]
}
```

**Use Cases:**
- Validate file paths against workspace roots
- Restrict operations to workspace only
- Apply different rules per workspace

**Example:**

```bash
#!/bin/bash
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')
workspace_roots=$(echo "$input" | jq -r '.workspace_roots[]')

# Check if file is within workspace
in_workspace=false
for root in $workspace_roots; do
  if [[ "$file_path" == "$root"* ]]; then
    in_workspace=true
    break
  fi
done

if [ "$in_workspace" = false ]; then
  echo '{"permission":"deny","userMessage":"File outside workspace"}'
  exit 0
fi

echo '{"permission":"allow"}'
```

---

## Resource Limits

### Current State

**Undocumented:**
- Memory limits per hook
- CPU limits
- Concurrent hook limit
- Maximum JSON payload size

**Community Observations:**
- Hooks can consume arbitrary memory
- No CPU throttling observed
- Large payloads (MB+) handled successfully

**Recommendations:**
- Assume 512MB memory budget
- Keep JSON payloads under 10MB
- Don't spawn subprocesses unnecessarily
- Clean up resources on exit

---

## Debugging Architecture

### Execution Visibility

**Limited Observability:**
- No built-in hook execution logs
- No performance metrics
- No error aggregation
- Manual logging required

**Recommended Logging:**

```bash
#!/bin/bash
# Log all hook executions
{
  echo "=== Hook Execution ==="
  echo "Timestamp: $(date -Iseconds)"
  echo "Event: $(echo "$input" | jq -r '.hook_event_name')"
  echo "Input: $input"
  echo "Output: $output"
  echo "Exit Code: $?"
} >> /tmp/cursor-hooks-debug.log
```

---

**Sources:**
- [Cursor Official Documentation](https://cursor.com/docs/agent/hooks)
- [GitButler Deep Dive](https://blog.gitbutler.com/cursor-hooks-deep-dive)
- [Skywork AI Guide](https://skywork.ai/blog/how-to-cursor-1-7-hooks-guide/)
- Community testing and observation
