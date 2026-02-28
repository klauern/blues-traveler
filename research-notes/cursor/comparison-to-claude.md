# Cursor Hooks vs Claude Code Hooks: Comprehensive Comparison

**Research Date:** 2026-02-21
**Based on:** Cursor documentation, blues-traveler implementation, cursor-compatibility.md

---

## Executive Summary

Both Cursor and Claude Code support lifecycle hooks for AI agent control, but with significant architectural differences. **Blues-traveler provides a compatibility layer** that allows hooks to work in both environments.

**Key Finding:** Cursor's hook system is **more limited but simpler**, while Claude Code (via blues-traveler) is **more powerful but complex**.

---

## Event System Comparison

### Event Name Mapping

Blues-traveler accepts both Cursor and Claude Code event names:

| Cursor Event Name | Claude Code Event | Blues-traveler Matcher | Description |
|-------------------|-------------------|------------------------|-------------|
| `beforeShellExecution` | `PreToolUse` | `TOOL_NAME=Bash` | Before shell commands |
| `afterFileEdit` | `PostToolUse` | `TOOL_NAME=Edit,Write` | After file modifications |
| `beforeReadFile` | `PreToolUse` | `TOOL_NAME=Read` | Before reading files |
| `beforeMCPExecution` | `PreToolUse` | Custom matcher | Before MCP tool execution |
| `beforeSubmitPrompt` | `UserPromptSubmit` | N/A | When user submits prompt |
| `stop` | `SessionEnd` | N/A | When session ends |

### Event Coverage

**Cursor Events:** 6 total
1. beforeShellExecution
2. beforeMCPExecution
3. beforeReadFile
4. afterFileEdit
5. beforeSubmitPrompt
6. stop

**Claude Code Events (blues-traveler):** 5 total
1. PreToolUse (covers multiple Cursor events)
2. PostToolUse (afterFileEdit equivalent)
3. UserPromptSubmit (beforeSubmitPrompt equivalent)
4. SessionStart (Cursor lacks this)
5. SessionEnd (stop equivalent)

**Winner:** Cursor (more granular events) / Claude Code (SessionStart event)

---

## Configuration Format Comparison

### Cursor: hooks.json

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      {
        "command": "~/.local/bin/security-check.sh \"${command}\"",
        "timeout": 5000
      }
    ],
    "afterFileEdit": [
      {
        "command": "prettier --check \"${file}\"",
        "timeout": 10000
      }
    ]
  }
}
```

**Location:** `.cursor/hooks.json`

**Features:**
- Simple JSON structure
- Event name → command array mapping
- Timeout per hook (in milliseconds)
- Variable interpolation: `${command}`, `${file}`

### Claude Code: settings.json (blues-traveler)

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.local/bin/security-check.sh",
            "timeout": 5
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit,Write",
        "hooks": [
          {
            "type": "command",
            "command": "prettier --check",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

**Location:** `.claude/settings.json`

**Features:**
- Matcher-based filtering (granular control)
- Multiple hook types (command, inline)
- Timeout in seconds
- No variable interpolation (env vars used instead)

### Claude Code: Custom YAML (blues-traveler only)

```yaml
# .claude/hooks.yml
security:
  PreToolUse:
    jobs:
      - name: check-dangerous
        run: |
          if echo "$TOOL_INPUT" | grep -q "sudo rm -rf /"; then
            cat <<EOF
          {
            "permission": "deny",
            "userMessage": "Command blocked",
            "agentMessage": "Dangerous pattern detected"
          }
          EOF
          else
            echo '{"permission": "allow"}'
          fi
        glob: ["*"]
        timeout: 5
```

**Features:**
- YAML format (more readable for complex hooks)
- Inline scripts (no separate file needed)
- Named jobs
- Glob-based filtering

**Winner:** Claude Code (more flexible with matchers and YAML support)

---

## Hook Input/Output Comparison

### Input Format

**Cursor:** JSON via stdin (well-documented)
```json
{
  "command": "npm install lodash@4.17.21",
  "conversation_id": "uuid",
  "generation_id": "uuid",
  "hook_event_name": "beforeShellExecution",
  "workspace_roots": ["/path/to/project"]
}
```

**Claude Code:** Environment variables (documented in blues-traveler)
```bash
HOOK_EVENT_NAME=PreToolUse
TOOL_NAME=Bash
TOOL_INPUT="npm install lodash@4.17.21"
FILES_CHANGED=""
USER_PROMPT=""
CWD="/path/to/project"
BT_HOOK_PLATFORM=claude
```

**Key Difference:**
- Cursor: **JSON stdin** (more structured)
- Claude Code: **Environment variables** (more Unix-friendly)

### Output Format

**Both support the same JSON response schema** (standardized by blues-traveler):

```json
{
  "permission": "allow|deny|ask",
  "userMessage": "Message for user",
  "agentMessage": "Message for AI agent",
  "continue": true|false
}
```

**Compatibility:** ✅ Full (blues-traveler standardizes this)

---

## Permission Models

### Cursor Permissions

**beforeShellExecution, beforeMCPExecution:**
- `allow` - Permit action ✅
- `deny` - Block action ✅
- `ask` - Prompt user for approval ✅ (native support)

**beforeReadFile:**
- `allow` - Permit read ✅
- `deny` - Block read ✅
- `ask` - ❌ NOT supported

**beforeSubmitPrompt:**
- `continue: true` - Allow ✅
- `continue: false` - Block ✅

**afterFileEdit, stop:**
- No permission control (informational)

### Claude Code Permissions (blues-traveler)

**PreToolUse, PostToolUse:**
- `allow` - Permit action ✅
- `deny` - Block action ✅
- `ask` - ⚠️ Falls back to allow with messages (Claude Code doesn't support native prompts)

**Platform Detection:**
Blues-traveler auto-detects Cursor environment and uses native "ask" when available:

```bash
# Cursor detected → native ask prompt
# Claude Code → allow with contextual messages + audit log
```

**Winner:** Cursor (native "ask" support) / Claude Code (graceful degradation)

---

## Environment Variables

### Cursor

**Documented:** None (relies on JSON stdin)

**Undocumented but available:**
- `CURSOR_*` - Cursor-specific variables (auto-detection)
- Standard shell env: `HOME`, `USER`, `PWD`

**Variable interpolation in hooks.json:**
```json
{
  "command": "script.sh \"${command}\" \"${file}\""
}
```

### Claude Code (blues-traveler)

**Documented:**
- `HOOK_EVENT_NAME` - Event that triggered hook
- `TOOL_NAME` - Tool being invoked (Bash, Edit, Read, etc.)
- `TOOL_INPUT` - Input to the tool
- `FILES_CHANGED` - Space-separated list of modified files
- `USER_PROMPT` - User's prompt text (UserPromptSubmit only)
- `CWD` - Current working directory
- `BT_HOOK_PLATFORM` - Override platform (`cursor` or `claude`)

**No variable interpolation** (use env vars directly in scripts)

**Winner:** Claude Code (comprehensive, documented env vars)

---

## Execution Model

### Cursor

**Process Model:**
- Standalone subprocess per hook
- JSON input via stdin
- JSON output via stdout
- Synchronous (blocks agent)

**Timeout:**
- Configurable per hook in hooks.json (milliseconds)
- Example: `"timeout": 5000` (5 seconds)

**Error Handling:**
- Exit 0 + valid JSON → Process response
- Exit 0 + no JSON → Allow (silent success)
- Exit non-zero + no JSON → Block with error
- Invalid JSON → Block with "hook broken" error

### Claude Code (blues-traveler)

**Process Model:**
- Standalone subprocess per hook
- Environment variables for input
- JSON output via stdout
- Synchronous (blocks agent)

**Timeout:**
- Configurable per hook in settings.json (seconds)
- Example: `"timeout": 5` (5 seconds)

**Error Handling:**
- Exit 0 + valid JSON → Process response
- Exit 0 + no JSON → Allow (silent success)
- Exit non-zero + no JSON → Block with error
- Invalid JSON → Block with "hook broken" error

**Identical behavior** (blues-traveler standardized this)

**Winner:** Tie (both well-defined, Cursor has millisecond precision)

---

## Matcher System

### Cursor

**No matcher system** - Events are tool-specific:
- `beforeShellExecution` → Only shell commands
- `beforeMCPExecution` → Only MCP tools
- `beforeReadFile` → Only file reads
- `afterFileEdit` → Only file edits

**Limitation:** Cannot filter within event type (e.g., only Python files in afterFileEdit)

**Workaround:** Hook script must check file extension/path

### Claude Code (blues-traveler)

**Powerful matcher system:**

```json
{
  "PreToolUse": [
    {
      "matcher": "Bash",           // Only Bash tool
      "hooks": [ ... ]
    },
    {
      "matcher": "Read",           // Only Read tool
      "hooks": [ ... ]
    },
    {
      "matcher": "Edit,Write",     // Edit OR Write tool
      "hooks": [ ... ]
    },
    {
      "matcher": "*",              // All tools
      "hooks": [ ... ]
    }
  ]
}
```

**YAML glob support:**
```yaml
jobs:
  - name: format-python
    run: ruff format "$TOOL_INPUT"
    glob: ["*.py"]           # Only Python files
```

**Winner:** Claude Code (powerful filtering without script logic)

---

## Message Routing

### Dual Messages (Both)

Both support separate messages for users vs AI agents:

**Cursor:**
```json
{
  "userMessage": "Code formatting failed for index.ts",
  "agentMessage": "Prettier check failed. Run: prettier --write \"index.ts\""
}
```

**Claude Code:**
```json
{
  "userMessage": "Code formatting failed for index.ts",
  "agentMessage": "Prettier check failed. Run: prettier --write \"index.ts\""
}
```

**Implementation:** ✅ Identical (blues-traveler standardized)

### Message Functions (blues-traveler wrappers)

Claude Code provides convenience functions:

```go
// Permission-based (PreToolUse)
BlockWithMessages(userMsg, agentMsg)
ApproveWithMessages(userMsg, agentMsg)
AskWithMessages(userMsg, agentMsg)        // Falls back to approve in Claude

// Informational (PostToolUse)
AllowWithMessages(userMsg, agentMsg)
PostBlockWithMessages(userMsg, agentMsg)
```

**Cursor:** Direct JSON output only (no wrapper functions)

**Winner:** Claude Code (better developer ergonomics with wrappers)

---

## Hook Types

### Cursor

**One type:** External command

```json
{
  "command": "path/to/script.sh"
}
```

Must be executable script (shell, Python, Node, etc.)

### Claude Code (blues-traveler)

**Two types:**

1. **Command** (external script)
```json
{
  "type": "command",
  "command": "path/to/script.sh"
}
```

2. **Inline** (YAML only)
```yaml
jobs:
  - name: quick-check
    run: |
      if [ "$TOOL_NAME" = "Bash" ]; then
        echo '{"permission":"allow"}'
      fi
```

**Winner:** Claude Code (inline scripts eliminate file management)

---

## Hook Lifecycle Features

### Cursor

**Available:**
- beforeShellExecution (pre-action)
- beforeMCPExecution (pre-action)
- beforeReadFile (pre-action)
- beforeSubmitPrompt (pre-action)
- afterFileEdit (post-action)
- stop (post-action)

**Missing:**
- ❌ SessionStart (no hook for session beginning)
- ❌ PreToolUse generic (must specify exact tool type)

### Claude Code (blues-traveler)

**Available:**
- PreToolUse (pre-action, covers Bash, Read, Edit, Write, MCP)
- PostToolUse (post-action, covers Edit, Write)
- UserPromptSubmit (pre-action)
- SessionStart (pre-action) ⭐
- SessionEnd (post-action)

**Missing:**
- ❌ Granular MCP vs Shell separation (use matcher instead)

**Winner:** Claude Code (SessionStart event, generic PreToolUse)

---

## Cross-Compatibility (blues-traveler)

### Cursor → Blues-traveler

**Event Translation:**
```
beforeShellExecution → PreToolUse (matcher: Bash)
beforeMCPExecution   → PreToolUse (custom matcher)
beforeReadFile       → PreToolUse (matcher: Read)
afterFileEdit        → PostToolUse (matcher: Edit,Write)
beforeSubmitPrompt   → UserPromptSubmit
stop                 → SessionEnd
```

**Automatic:** Blues-traveler accepts Cursor event names

### Blues-traveler → Cursor

**Manual translation required:**
```
PreToolUse (Bash)       → beforeShellExecution
PreToolUse (Read)       → beforeReadFile
PostToolUse (Edit,Write)→ afterFileEdit
UserPromptSubmit        → beforeSubmitPrompt
SessionEnd              → stop
```

**Extract hook scripts** from blues-traveler config and create Cursor hooks.json

### Write-Once Compatibility

**Best Practice:** Use Cursor event names everywhere

```bash
# Works in both Cursor and blues-traveler
blues-traveler hooks install security --event beforeShellExecution
```

**Winner:** Blues-traveler (cross-platform abstraction)

---

## Feature Comparison Matrix

| Feature | Cursor | Claude Code (blues-traveler) |
|---------|--------|------------------------------|
| **Events** | 6 specific events | 5 generic + matchers |
| **Config Format** | JSON only | JSON + YAML |
| **Input Method** | JSON stdin | Environment variables |
| **Output Format** | JSON stdout | JSON stdout |
| **Permission Modes** | allow/deny/ask | allow/deny/ask (with fallback) |
| **Ask Permission** | ✅ Native | ⚠️ Fallback to allow + messages |
| **Timeout Config** | ✅ Per-hook (ms) | ✅ Per-hook (seconds) |
| **Matcher System** | ❌ None | ✅ Tool/file matchers |
| **Variable Interpolation** | ✅ ${command}, ${file} | ❌ Use env vars |
| **Environment Vars** | ❌ Not documented | ✅ Fully documented |
| **Hook Types** | Command only | Command + Inline |
| **Dual Messages** | ✅ Yes | ✅ Yes |
| **SessionStart Event** | ❌ No | ✅ Yes |
| **JSON Schema** | ✅ Official | ✅ Via blues-traveler |
| **SDK Support** | TypeScript, Python | Go (native) |
| **Official Docs** | ⚠️ Minimal | ✅ Comprehensive (blues-traveler) |
| **Community Examples** | ✅ Many | ⚠️ Growing |
| **Cross-compatibility** | ❌ Cursor-only | ✅ Works in both |

---

## Use Case Recommendations

### Choose Cursor If:
- ✅ You only use Cursor IDE
- ✅ You need native "ask" permission prompts
- ✅ You want simpler configuration (single JSON file)
- ✅ You prefer event-specific hooks (no matchers)
- ✅ You value large community examples

### Choose Claude Code (blues-traveler) If:
- ✅ You use multiple AI IDEs (Cursor + Claude Code)
- ✅ You need powerful matchers (file globs, tool filtering)
- ✅ You want YAML configuration for complex hooks
- ✅ You prefer environment variables over JSON stdin
- ✅ You need SessionStart event
- ✅ You want inline hooks (no separate files)
- ✅ You need better documentation

### Use Both (Cross-Compatible Hooks) If:
- ✅ Team uses mixed IDE environments
- ✅ You want maximum portability
- ✅ You prefer Cursor event names (blues-traveler auto-translates)

---

## Migration Strategies

### Cursor → Blues-traveler

**Step 1:** Install blues-traveler
```bash
go install github.com/klauern/blues-traveler@latest
```

**Step 2:** Copy hook scripts (no changes needed)
```bash
cp -r .cursor/hooks .claude/hooks
```

**Step 3:** Translate hooks.json → settings.json

**Before (Cursor):**
```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security.sh" }
    ]
  }
}
```

**After (Claude Code):**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/security.sh"
          }
        ]
      }
    ]
  }
}
```

**Or use Cursor event names (auto-translates):**
```bash
blues-traveler hooks install security --event beforeShellExecution
```

### Blues-traveler → Cursor

**Step 1:** Extract hook scripts

**Step 2:** Create .cursor/hooks.json

**Step 3:** Translate event names:
- PreToolUse (Bash) → beforeShellExecution
- PreToolUse (Read) → beforeReadFile
- PostToolUse (Edit,Write) → afterFileEdit
- UserPromptSubmit → beforeSubmitPrompt
- SessionEnd → stop

**Step 4:** Update scripts to read JSON stdin instead of env vars

**Before (env vars):**
```bash
command="$TOOL_INPUT"
```

**After (JSON stdin):**
```bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')
```

---

## Best Practices for Cross-Compatibility

### 1. Use Cursor Event Names

Blues-traveler auto-translates them:
```bash
blues-traveler hooks install sec --event beforeShellExecution
# Works in both Cursor and Claude Code
```

### 2. Standardize on JSON Output

Both systems support identical JSON response format:
```json
{
  "permission": "deny",
  "userMessage": "User-friendly message",
  "agentMessage": "Technical details for AI"
}
```

### 3. Document Dependencies

List required tools in README:
```markdown
## Requirements
- jq (JSON parsing)
- prettier (formatting)
- ruff (Python linting)
```

### 4. Store Scripts in Version Control

```
.claude/hooks/scripts/    # or .cursor/hooks/
├── security-check.sh
├── format.sh
└── audit.sh
```

### 5. Test in Both Environments

```bash
# Test with Cursor-style JSON stdin
echo '{"command":"npm install"}' | ./security-check.sh

# Test with Claude Code env vars
TOOL_INPUT="npm install" ./security-check.sh
```

---

## Conclusion

**Cursor:** Simpler, event-specific, better community resources, native "ask" support

**Claude Code (blues-traveler):** More powerful, matcher-based, better documentation, cross-compatible

**Blues-traveler Compatibility Layer:** Best of both worlds - write once, run everywhere

---

## References

- Cursor Documentation: https://cursor.com/docs/agent/hooks
- Blues-traveler: https://github.com/klauern/blues-traveler
- Cursor Compatibility Guide: `docs/cursor-compatibility.md`
- TypeScript SDK: https://github.com/johnlindquist/cursor-hooks
- Python SDK: https://github.com/DevonFulcher/py-cursor-hooks
