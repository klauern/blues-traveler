# Claude Code Hooks - Architecture & Internals

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete

---

## Table of Contents

1. [Execution Architecture](#execution-architecture)
2. [Hook Lifecycle](#hook-lifecycle)
3. [Hook Types](#hook-types)
4. [Performance Considerations](#performance-considerations)
5. [Security Model](#security-model)
6. [Configuration Merging](#configuration-merging)

---

## Execution Architecture

### Process Model

Claude Code hooks execute as **separate processes** spawned from the main Claude Code runtime:

```
┌─────────────────────────────────────────┐
│      Claude Code Main Process           │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │     Event System                   │ │
│  │  (SessionStart, PreToolUse, etc.)  │ │
│  └────────────────────────────────────┘ │
│                │                         │
│                │ Event fires             │
│                ↓                         │
│  ┌────────────────────────────────────┐ │
│  │   Hook Matcher & Filter            │ │
│  │   - Regex matching                 │ │
│  │   - Event filtering                │ │
│  └────────────────────────────────────┘ │
│                │                         │
│                │ Matched hooks           │
│                ↓                         │
│  ┌────────────────────────────────────┐ │
│  │   Hook Executor (Sequential)       │ │
│  │   - Spawn subprocess                │ │
│  │   - Pipe JSON to stdin             │ │
│  │   - Wait for exit code             │ │
│  │   - Parse stdout                   │ │
│  └────────────────────────────────────┘ │
│                │                         │
└────────────────┼─────────────────────────┘
                 │
                 ↓
        ┌────────────────┐
        │  Hook Process  │
        │  (shell/LLM/   │
        │   agent)       │
        └────────────────┘
```

**Key Characteristics:**
- Each hook runs in a separate process
- Hooks do not share state
- Process isolation prevents one hook from affecting another
- No shared memory or global variables

**Source:** [Official Hooks Documentation](https://code.claude.com/docs/en/hooks)

---

### Communication Protocol

**Input (stdin):** JSON event payload
**Output (stdout):** Plain text or JSON decision
**Errors (stderr):** Error messages, shown to user or Claude
**Control (exit code):** 0 (success), 2 (block), other (error)

```
┌──────────────┐
│ Claude Code  │
└──────┬───────┘
       │ JSON payload via stdin
       │ {"tool_name": "Bash", ...}
       ↓
┌──────────────┐
│ Hook Script  │
│  (reads stdin)
│  (processes)  │
│  (outputs)    │
└──────┬───────┘
       │
       ├─→ stdout: JSON decision or text
       ├─→ stderr: error messages
       └─→ exit code: 0, 2, or other
       │
       ↓
┌──────────────┐
│ Claude Code  │
│ (processes   │
│  response)   │
└──────────────┘
```

**Source:** [Claude Code Official Documentation](https://code.claude.com/docs/en/hooks)

---

## Hook Lifecycle

### Complete Lifecycle Stages

**1. Discovery Phase**

At session start, Claude Code:
- Reads all settings files (user, project, local, policy, plugin)
- Merges configurations (see [Configuration Merging](#configuration-merging))
- Validates hook definitions
- Caches hook registry for session

**2. Event Triggering**

When an event occurs:
- Event fires with context (tool name, file path, etc.)
- Event payload constructed as JSON
- Hook registry queried for matching hooks

**3. Filtering Phase**

For each registered hook:
- Check if event type matches
- Apply matcher regex (if present)
- Skip if matcher doesn't match
- Build execution list of matched hooks

**4. Execution Phase**

For each matched hook (sequential):
- Spawn subprocess (shell command, LLM prompt, or agent)
- Pipe JSON payload to stdin
- Start timeout timer (default: 10 minutes for commands, 30s for prompts, 60s for agents)
- Wait for process completion

**5. Response Processing**

When hook completes:
- Capture exit code
- Read stdout (for JSON or text output)
- Read stderr (for error messages)
- Parse JSON decision (if applicable)
- Apply decision logic (allow, deny, block, context injection)

**6. Decision Application**

Based on exit code and JSON output:
- **Exit 0:** Parse stdout for decisions or context
- **Exit 2:** Block action (if event is cancellable), feed stderr to Claude
- **Other:** Log error, continue execution

**Event Flow Example (PreToolUse):**

```
User prompt → Claude decides to use Bash tool
                      ↓
            PreToolUse event fires
                      ↓
            Matcher checks: "Bash" matches "Bash" regex
                      ↓
            Hook script spawned with JSON stdin
                      ↓
            Script analyzes command: rm -rf /
                      ↓
            Script outputs: {"permissionDecision": "deny"}
                      ↓
            Exit code: 0 (success)
                      ↓
            Claude Code blocks tool execution
                      ↓
            Claude sees reason: "Dangerous command blocked"
```

**Source:** [Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

---

## Hook Types

Claude Code supports three distinct hook types, each with different execution models.

### 1. Command Hooks

**Type:** `"command"`
**Execution:** Shell command spawned as subprocess
**Default Timeout:** 600 seconds (10 minutes)

**Characteristics:**
- Most common hook type
- Run any executable (shell scripts, binaries, etc.)
- Full access to system commands
- Synchronous or async execution
- Best for deterministic logic

**Example:**
```json
{
  "type": "command",
  "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format.sh",
  "timeout": 30,
  "statusMessage": "Formatting code...",
  "async": false
}
```

**Async Mode:**
- Set `"async": true` to run in background
- Output delivered on next conversation turn
- Cannot block actions
- No deduplication across firings

**Source:** [Official Hooks Reference](https://code.claude.com/docs/en/hooks)

---

### 2. Prompt Hooks

**Type:** `"prompt"`
**Execution:** Single-turn LLM evaluation
**Default Timeout:** 30 seconds

**Characteristics:**
- Quick LLM-based decisions
- No tool access
- Best for complex conditional logic
- Uses faster Claude model by default
- Returns boolean decision

**Example:**
```json
{
  "type": "prompt",
  "prompt": "Evaluate if this command is safe: $ARGUMENTS. Respond with {\"ok\": true} or {\"ok\": false, \"reason\": \"why\"}",
  "model": "haiku",
  "timeout": 30
}
```

**Response Schema:**
```json
{
  "ok": true|false,
  "reason": "Required when ok is false"
}
```

**Supported Events:**
- PermissionRequest
- PostToolUse
- PostToolUseFailure
- PreToolUse
- Stop
- SubagentStop
- TaskCompleted
- UserPromptSubmit

**Source:** [Hooks Guide - Prompt Hooks](https://code.claude.com/docs/en/hooks-guide#prompt-hooks)

---

### 3. Agent Hooks

**Type:** `"agent"`
**Execution:** Multi-turn subagent with tool access
**Default Timeout:** 60 seconds

**Characteristics:**
- Full subagent with tools: Read, Grep, Glob
- Up to 50 conversation turns
- Can explore codebase
- Best for complex verification
- Uses Sonnet by default

**Example:**
```json
{
  "type": "agent",
  "prompt": "Verify all tests pass before allowing stop. $ARGUMENTS",
  "model": "sonnet",
  "timeout": 120
}
```

**Response Schema:**
```json
{
  "ok": true|false,
  "reason": "Required when ok is false"
}
```

**Supported Events:** Same as prompt hooks

**Tool Access:**
- Read files
- Search with Grep
- Find files with Glob
- Cannot execute Bash commands
- Cannot write/edit files

**Source:** [Hooks Guide - Agent Hooks](https://code.claude.com/docs/en/hooks-guide#agent-hooks)

---

## Performance Considerations

### Execution Model

**Sequential Execution:**
- Hooks for the same event execute sequentially (not parallel)
- Order is deterministic but implementation-specific
- One hook must complete before next starts

**Performance Impact:**
```
Event fires
    │
    ├─→ Hook 1 (500ms) ──→ complete
    │                      │
    └─→ Hook 2 (300ms) ────→ complete
                           │
                  Total: 800ms delay
```

**Optimization Strategies:**

1. **Use Async Hooks for Non-Critical Tasks**
   ```json
   {
     "type": "command",
     "command": "/path/to/notification.sh",
     "async": true
   }
   ```

2. **Set Aggressive Timeouts**
   ```json
   {
     "type": "command",
     "command": "/path/to/quick-check.sh",
     "timeout": 5
   }
   ```

3. **Optimize Hook Scripts**
   - Cache expensive computations
   - Early exit when possible
   - Avoid network calls in sync hooks

4. **Use Matcher Filters Effectively**
   ```json
   {
     "matcher": "Write|Edit",  // Only fire for Write/Edit
     "hooks": [...]
   }
   ```

**Source:** [Community Best Practices](https://www.eesel.ai/blog/settings-json-claude-code)

---

### Timeout Handling

**Default Timeouts:**
- Command hooks: 600 seconds (10 minutes)
- Prompt hooks: 30 seconds
- Agent hooks: 60 seconds

**Timeout Behavior:**
- Process killed after timeout
- Treated as non-blocking error (exit code != 0, != 2)
- Stderr logged in verbose mode
- Execution continues

**Custom Timeouts:**
```json
{
  "type": "command",
  "command": "/path/to/long-test.sh",
  "timeout": 300  // 5 minutes
}
```

**Source:** [Official Hooks Documentation](https://code.claude.com/docs/en/hooks)

---

### Resource Limits

**No Hard Limits:**
- Claude Code does not enforce CPU or memory limits on hooks
- Hooks run with full user permissions
- Resource usage limited only by OS

**Best Practices:**
- Avoid infinite loops in hooks
- Use timeouts to prevent hangs
- Monitor hook performance in production
- Use async hooks for expensive operations

---

## Security Model

### Permission Model

**No Sandboxing:**
- Hooks execute with full user permissions
- Can read, write, delete any files user can access
- Can execute arbitrary commands
- Can make network requests

**Security Warning from Official Docs:**
> Hooks execute shell commands with your full user permissions. They can modify, delete, or access any files your user account can access.

**Implications:**
- Hooks are trusted code
- Review all hook scripts carefully
- Never run untrusted hooks
- Validate inputs from JSON payload

**Source:** [Official Security Warning](https://code.claude.com/docs/en/hooks#security)

---

### Policy Settings

**Managed Policy Hooks:**
- Organization admins can enforce hooks via policy settings
- Users cannot disable or modify policy hooks
- Policy hooks always execute
- `"disableAllHooks": true` does not affect policy hooks

**ConfigChange Event:**
- Fires when settings change
- Can block user/project settings changes
- **Cannot** block `policy_settings` changes

**Source:** [Hooks Guide - ConfigChange](https://code.claude.com/docs/en/hooks-guide#configchange)

---

### Input Validation

**Always Validate:**
- File paths (prevent path traversal: `../`)
- Commands (block dangerous patterns)
- User input (sanitize before use)

**Example Validation:**
```bash
#!/bin/bash
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Validate file path
if [[ "$FILE" == *".."* ]]; then
  echo "Path traversal detected" >&2
  exit 2
fi

# Check file exists in project
if [[ ! "$FILE" =~ ^"$CLAUDE_PROJECT_DIR" ]]; then
  echo "File outside project directory" >&2
  exit 2
fi
```

**Source:** [Community Security Patterns](https://github.com/trailofbits/claude-code-config)

---

## Configuration Merging

### Configuration Sources

Claude Code merges hooks from multiple sources (in order of precedence):

1. **Managed policy settings** (highest precedence)
2. **User settings** (`~/.claude/settings.json`)
3. **Project settings** (`.claude/settings.json`)
4. **Local project settings** (`.claude/settings.local.json`)
5. **Plugin settings** (`plugin-dir/hooks/hooks.json`)
6. **Skill/Agent frontmatter** (in component YAML)

**Source:** [Hooks Configuration](https://code.claude.com/docs/en/hooks-guide#configuration)

---

### Merge Strategy

**Array Concatenation:**
- Hooks from all sources are concatenated
- No deduplication
- Order: policy → user → project → local → plugin → skill

**Example:**

**User settings (`~/.claude/settings.json`):**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{"type": "command", "command": "~/security-check.sh"}]
      }
    ]
  }
}
```

**Project settings (`.claude/settings.json`):**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [{"type": "command", "command": "./format.sh"}]
      }
    ]
  }
}
```

**Merged Result:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{"type": "command", "command": "~/security-check.sh"}]
      },
      {
        "matcher": "Write|Edit",
        "hooks": [{"type": "command", "command": "./format.sh"}]
      }
    ]
  }
}
```

**Both hooks execute when PreToolUse fires** (if matchers match).

---

### Mid-Session Changes

**Important Behavior:**
- Hook configurations are captured at session start
- Changes to settings files **do not take effect immediately**
- Must review changes in `/hooks` menu before applying
- Security feature to prevent malicious config changes

**To Apply Changes:**
1. Edit settings file
2. Open `/hooks` menu
3. Review changes
4. Accept to apply to current session

**Source:** [Hooks Guide - Management](https://code.claude.com/docs/en/hooks-guide#management)

---

### Disabling Hooks

**Temporary Disable:**
```json
{
  "disableAllHooks": true
}
```

**Effects:**
- Disables all user, project, and local hooks
- **Does not** disable policy hooks
- Can toggle in `/hooks` menu
- Per-session setting

**Source:** [Hooks Reference](https://code.claude.com/docs/en/hooks)

---

## Known Limitations

### Technical Constraints

1. **Sequential Execution Only**
   - Hooks cannot run in parallel
   - Total execution time = sum of all hook times

2. **No Shared State**
   - Each hook is isolated
   - Cannot pass data between hooks
   - Must use filesystem or external storage

3. **No Guaranteed Order**
   - Hook execution order within same event is deterministic but implementation-specific
   - Don't rely on specific ordering

4. **Limited Async Support**
   - Async only for command hooks
   - Async hooks cannot block actions
   - No guaranteed delivery of async output

5. **Session-Scoped**
   - Hooks captured at session start
   - Mid-session changes require manual review
   - No hot-reload

**Source:** [Official Documentation](https://code.claude.com/docs/en/hooks)

---

## Debug & Introspection

### Debug Mode

```bash
claude --debug
```

**Shows:**
- Hook discovery process
- Matcher evaluation
- Hook execution start/end
- Exit codes
- Stdout/stderr output
- Timing information

**Example Output:**
```
[DEBUG] Executing hooks for PostToolUse:Write
[DEBUG] Getting matching hook commands for PostToolUse with query: Write
[DEBUG] Found 1 hook matchers in settings
[DEBUG] Matched 1 hooks for query "Write"
[DEBUG] Found 1 hook commands to execute
[DEBUG] Executing hook command: <command> with timeout 600000ms
[DEBUG] Hook command completed with status 0: <stdout>
```

**Source:** [Hooks Guide - Debugging](https://code.claude.com/docs/en/hooks-guide#debugging)

---

### Verbose Mode

**Toggle:** Press `Ctrl+O` during session

**Shows:**
- Hook execution progress in transcript
- Hook status messages
- Real-time feedback

---

## See Also

- [Events Reference](./events-reference.md) - Complete event catalog
- [Configuration](./configuration.md) - Settings file format
- [Scripting](./scripting.md) - Writing hook scripts
- [Security](./security.md) - Security best practices

---

**Sources:**
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)
- [eesel.ai Developer Guide](https://www.eesel.ai/blog/settings-json-claude-code)
- [Community Examples](https://github.com/disler/claude-code-hooks-mastery)

**Document Version:** 1.0
**Research Date:** 2026-02-21
