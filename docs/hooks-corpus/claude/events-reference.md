# Claude Code Hooks - Event Types & Triggers

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete - All 17 Events Documented

---

## Event Summary Table

| # | Event | Trigger | Timing | Cancellable | Matcher Support | Hook Types |
|---|-------|---------|--------|-------------|-----------------|------------|
| 1 | SessionStart | Session begins/resumes | Session Init | ❌ No | ✅ Yes | command, prompt, agent |
| 2 | UserPromptSubmit | User submits prompt | Before Processing | ✅ Yes | ❌ No | command, prompt, agent |
| 3 | PreToolUse | Before tool execution | Pre-Action | ✅ Yes | ✅ Yes | command, prompt, agent |
| 4 | PermissionRequest | Permission dialog | Pre-Permission | ✅ Yes | ✅ Yes | command, prompt, agent |
| 5 | PostToolUse | After tool success | Post-Action | ❌ No | ✅ Yes | command, prompt, agent |
| 6 | PostToolUseFailure | After tool failure | Post-Action | ❌ No | ✅ Yes | command, prompt, agent |
| 7 | Notification | Notification sent | During | ❌ No | ✅ Yes | command only |
| 8 | SubagentStart | Subagent spawned | Pre-Subagent | ❌ No | ✅ Yes | command only |
| 9 | SubagentStop | Subagent finishes | Post-Subagent | ✅ Yes | ✅ Yes | command, prompt, agent |
| 10 | Stop | Claude finishes | Post-Response | ✅ Yes | ❌ No | command, prompt, agent |
| 11 | TeammateIdle | Teammate going idle | Pre-Idle | ✅ Yes | ❌ No | command only |
| 12 | TaskCompleted | Task marked complete | Pre-Completion | ✅ Yes | ❌ No | command, prompt, agent |
| 13 | ConfigChange | Config file changes | During | ✅ Yes* | ✅ Yes | command only |
| 14 | WorktreeCreate | Worktree being created | Pre-Creation | ✅ Yes | ❌ No | command only |
| 15 | WorktreeRemove | Worktree being removed | Pre-Removal | ❌ No | ❌ No | command only |
| 16 | PreCompact | Before compaction | Pre-Compact | ❌ No | ✅ Yes | command only |
| 17 | SessionEnd | Session terminates | Post-Session | ❌ No | ✅ Yes | command only |

\* ConfigChange: Cannot block `policy_settings` changes

---

## 1. SessionStart

**Trigger:** When Claude Code starts a new session or resumes an existing session
**Timing:** Once per session (startup, resume, clear, or compact)
**Frequency:** Low (session-level)
**Hook Types:** command, prompt, agent

### When It Fires

| Matcher Value | Trigger |
|---------------|---------|
| `startup` | New session |
| `resume` | `--resume`, `--continue`, or `/resume` |
| `clear` | `/clear` command |
| `compact` | Auto or manual compaction |

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "SessionStart",
  "source": "startup|resume|clear|compact",
  "model": "claude-sonnet-4-6",
  "agent_type": "agent_name"  // Optional, if --agent used
}
```

### Decision Control

**Cannot block session start.**

**Can add context:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Added to Claude's context"
  }
}
```

Or output plain text to stdout (automatically added to context).

### Special Features

**Environment Variable Persistence:** `$CLAUDE_ENV_FILE`

```bash
#!/bin/bash
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export NODE_ENV=production' >> "$CLAUDE_ENV_FILE"
  echo 'export PATH="$PATH:./bin"' >> "$CLAUDE_ENV_FILE"
fi
```

Variables written to this file available in all subsequent Bash commands.

### Common Use Cases

- Load project context (git status, TODOs)
- Initialize environment variables
- Check project state
- Load recent tickets/issues
- Set up logging

### Example

```bash
#!/bin/bash
echo "## Current Project Context"
echo ""
echo "### Git Status"
git status --short
echo ""
echo "### Recent TODOs"
rg "TODO|FIXME" --max-count 10
echo ""
echo "### Recent Commits"
git log --oneline -5
```

---

## 2. UserPromptSubmit

**Trigger:** When user submits a prompt, before Claude processes it
**Timing:** Pre-processing
**Frequency:** High (every user message)
**Hook Types:** command, prompt, agent

### When It Fires

Every time user submits a prompt. No matcher support.

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "UserPromptSubmit",
  "prompt": "User's submitted text"
}
```

### Decision Control

**Can block prompt processing:**

```json
{
  "decision": "block",
  "reason": "Shown to user (not to Claude)",
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "Added to context if allowed"
  }
}
```

**Or output plain text to stdout** (exit 0) to add context without JSON.

### Effects

- **Block:** Prevents processing, erases prompt from context
- **Allow:** Continue normally, optionally add context

### Common Use Cases

- Validate prompt content
- Add context based on prompt
- Block certain types of requests
- Inject instructions
- Load relevant documentation

### Example

```bash
#!/bin/bash
INPUT=$(cat)
PROMPT=$(echo "$INPUT" | jq -r '.prompt')

# Check for sensitive patterns
if echo "$PROMPT" | grep -qi "password\|secret\|api.key"; then
  echo "Prompt contains sensitive terms" >&2
  exit 2  # Block
fi

# Add context
echo "Note: Follow security guidelines in SECURITY.md"
exit 0
```

---

## 3. PreToolUse

**Trigger:** After Claude creates tool parameters, before processing the tool call
**Timing:** Pre-action
**Frequency:** Very high (every tool use)
**Hook Types:** command, prompt, agent

### When It Fires

Before any tool execution. Matcher filters by tool name.

### Matcher Values

Tool names:
- `Bash` - Shell command execution
- `Edit` - File editing
- `Write` - File creation/overwrite
- `Read` - File reading
- `Glob` - File pattern matching
- `Grep` - Content searching
- `Task` - Subagent spawning
- `WebFetch` - Fetching web content
- `WebSearch` - Web searching
- `mcp__<server>__<tool>` - MCP tool names

Regex patterns:
- `Edit|Write` - Either Edit or Write
- `mcp__.*` - Any MCP tool
- `mcp__memory__.*` - All memory server tools

### Input Schema

**Common fields:**
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": { /* tool-specific */ },
  "tool_use_id": "toolu_01ABC123..."
}
```

**Tool-specific inputs:** See [Tool Input Schemas](#tool-input-schemas) section.

### Decision Control

**Most flexible control of all events:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|deny|ask",
    "permissionDecisionReason": "Explanation",
    "updatedInput": { "field": "new_value" },
    "additionalContext": "Context for Claude"
  }
}
```

| Decision | Effect |
|----------|--------|
| `allow` | Bypass permission system, proceed |
| `deny` | Block tool call, feed reason to Claude |
| `ask` | Show permission prompt to user |

### Effects

- **allow:** Tool executes immediately, user not prompted
- **deny:** Tool blocked, Claude sees reason
- **ask:** User gets permission dialog

**Note:** Can modify tool input with `updatedInput` before execution.

### Common Use Cases

- Block dangerous commands
- Validate file paths
- Check security constraints
- Modify tool parameters
- Auto-approve safe operations
- Inject additional context

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Block dangerous Bash commands
if [[ "$TOOL" == "Bash" ]] && echo "$COMMAND" | grep -qE "rm -rf|sudo|curl.*\\|.*sh"; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Dangerous command blocked by security hook"
    }
  }'
  exit 0
fi

# Allow safe operations
exit 0
```

---

## 4. PermissionRequest

**Trigger:** When permission dialog is about to be shown
**Timing:** Pre-permission
**Frequency:** Medium (when permissions needed)
**Hook Types:** command, prompt, agent

### When It Fires

When user would normally see a permission prompt. Matcher filters by tool name (same values as PreToolUse).

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "PermissionRequest",
  "tool_name": "Bash",
  "tool_input": { /* same as PreToolUse */ },
  "permission_suggestions": [
    { "type": "toolAlwaysAllow", "tool": "Bash" }
  ]
}
```

**Note:** No `tool_use_id` (unlike PreToolUse).

### Decision Control

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PermissionRequest",
    "decision": {
      "behavior": "allow|deny",
      "updatedInput": { /* modify tool input */ },
      "updatedPermissions": { /* permission rules */ },
      "message": "Denial reason for Claude",
      "interrupt": false  // stop Claude on deny
    }
  }
}
```

| Behavior | Effect |
|----------|--------|
| `allow` | Grant permission, optionally modify input/rules |
| `deny` | Deny permission, optionally stop Claude |

### Effects

- **allow:** Permission granted on behalf of user
- **deny:** Permission denied, optionally interrupt Claude

### Common Use Cases

- Auto-approve safe operations
- Deny dangerous operations
- Modify tool input before approval
- Set permission rules
- Implement custom approval logic

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Auto-approve safe npm commands
if [[ "$TOOL" == "Bash" ]] && echo "$COMMAND" | grep -q "^npm \\(test\\|run\\|install\\)"; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PermissionRequest",
      decision: {
        behavior: "allow"
      }
    }
  }'
  exit 0
fi

# Ask user for everything else
exit 0
```

---

## 5. PostToolUse

**Trigger:** Immediately after tool completes successfully
**Timing:** Post-action
**Frequency:** Very high (every successful tool use)
**Hook Types:** command, prompt, agent

### When It Fires

After tool has executed and succeeded. Matcher filters by tool name (same values as PreToolUse).

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "PostToolUse",
  "tool_name": "Write",
  "tool_input": { /* same as PreToolUse */ },
  "tool_response": {
    "filePath": "/path/to/file.txt",
    "success": true
  },
  "tool_use_id": "toolu_01ABC123..."
}
```

### Decision Control

**Cannot undo action (tool already ran).**

**Can provide feedback:**
```json
{
  "decision": "block",  // Prompts Claude with reason
  "reason": "Explanation for Claude",
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "Additional context",
    "updatedMCPToolOutput": { /* MCP tools only */ }
  }
}
```

### Effects

- **"block" decision:** Prompts Claude with reason (tool already executed)
- **additionalContext:** Adds context for Claude
- **updatedMCPToolOutput:** Replaces MCP tool output

### Common Use Cases

- Auto-format files after edits
- Run linters
- Execute tests (async)
- Audit logging
- Notifications
- Verify outputs

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
FILE=$(echo "$INPUT" | jq -r '.tool_response.filePath // empty')

# Format Python files after Write/Edit
if [[ "$TOOL" == "Write" || "$TOOL" == "Edit" ]] && [[ "$FILE" == *.py ]]; then
  black "$FILE"
  isort "$FILE"
  echo "Formatted $FILE" >&2
fi

exit 0
```

---

## 6. PostToolUseFailure

**Trigger:** When tool execution fails
**Timing:** Post-action
**Frequency:** Low-Medium (when tools fail)
**Hook Types:** command, prompt, agent

### When It Fires

After tool call throws error or returns failure. Matcher filters by tool name.

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "PostToolUseFailure",
  "tool_name": "Bash",
  "tool_input": { /* same as PreToolUse */ },
  "tool_use_id": "toolu_01ABC123...",
  "error": "Command exited with non-zero status code 1",
  "is_interrupt": false  // Optional
}
```

### Decision Control

**Can add context about failure:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUseFailure",
    "additionalContext": "Additional information about failure"
  }
}
```

### Effects

- Adds context to help Claude understand/recover from failure

### Common Use Cases

- Log errors
- Send alerts
- Provide debugging hints
- Suggest fixes
- Track failure patterns

### Example

```bash
#!/bin/bash
INPUT=$(cat)
ERROR=$(echo "$INPUT" | jq -r '.error')
TOOL=$(echo "$INPUT" | jq -r '.tool_name')

# Log to file
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") [$TOOL] $ERROR" >> ~/.claude/error.log

# Send alert (async)
if command -v osascript &>/dev/null; then
  osascript -e "display notification \"$ERROR\" with title \"Claude Code Error\"" &
fi

exit 0
```

---

## 7. Notification

**Trigger:** When Claude Code sends a notification
**Timing:** During session
**Frequency:** Low (permission prompts, idle alerts, etc.)
**Hook Types:** command only

### When It Fires

When notifications sent. Matcher filters by notification type.

### Matcher Values

| Matcher | Description |
|---------|-------------|
| `permission_prompt` | Permission needed |
| `idle_prompt` | Claude idle |
| `auth_success` | Auth successful |
| `elicitation_dialog` | Dialog shown |

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "Notification",
  "message": "Claude needs your permission to use Bash",
  "title": "Permission needed",
  "notification_type": "permission_prompt"
}
```

### Decision Control

**Cannot block notifications.**

**Can add context:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "Notification",
    "additionalContext": "Context for conversation"
  }
}
```

### Common Use Cases

- Forward notifications to other systems
- Log notification events
- Trigger external workflows
- Send to Slack/email

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TYPE=$(echo "$INPUT" | jq -r '.notification_type')
MESSAGE=$(echo "$INPUT" | jq -r '.message')

# Send to Slack
if [[ "$TYPE" == "permission_prompt" ]]; then
  curl -X POST "$SLACK_WEBHOOK_URL" \
    -H 'Content-Type: application/json' \
    -d "{\"text\":\"Claude needs permission: $MESSAGE\"}"
fi

exit 0
```

---

## 8. SubagentStart

**Trigger:** When subagent is spawned via Task tool
**Timing:** Pre-subagent
**Frequency:** Low (when using subagents)
**Hook Types:** command only

### When It Fires

When Task tool spawns a subagent. Matcher filters by agent type.

### Matcher Values

- `Bash` - Bash specialist agent
- `Explore` - Exploration agent
- `Plan` - Planning agent
- Custom agent names from `.claude/agents/`

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "SubagentStart",
  "agent_id": "agent-abc123",
  "agent_type": "Explore"
}
```

### Decision Control

**Cannot block subagent creation.**

**Can inject context:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "SubagentStart",
    "additionalContext": "Context for subagent"
  }
}
```

### Common Use Cases

- Inject instructions for subagent
- Log subagent creation
- Track agent hierarchy
- Add subagent-specific context

### Example

```bash
#!/bin/bash
INPUT=$(cat)
AGENT_TYPE=$(echo "$INPUT" | jq -r '.agent_type')

# Provide exploration guidelines
if [[ "$AGENT_TYPE" == "Explore" ]]; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "SubagentStart",
      additionalContext: "Focus on security implications during exploration"
    }
  }'
fi

exit 0
```

---

## 9. SubagentStop

**Trigger:** When subagent finishes responding
**Timing:** Post-subagent
**Frequency:** Low (when subagents complete)
**Hook Types:** command, prompt, agent

### When It Fires

After subagent completes. Matcher filters by agent type (same values as SubagentStart).

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/main/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "SubagentStop",
  "stop_hook_active": false,
  "agent_id": "def456",
  "agent_type": "Explore",
  "agent_transcript_path": "/path/subagents/agent-def456.jsonl",
  "last_assistant_message": "Analysis complete..."
}
```

### Decision Control

**Same as Stop event:**

```json
{
  "decision": "block",  // Prevents subagent from stopping
  "reason": "Required when blocking"
}
```

### Effects

- **"block":** Subagent continues working with reason
- **Allow (no JSON or exit 0):** Subagent stops normally

### Common Use Cases

- Verify subagent completed task
- Check quality of results
- Enforce completion criteria
- Log subagent results

### Example

```bash
#!/bin/bash
INPUT=$(cat)
MESSAGE=$(echo "$INPUT" | jq -r '.last_assistant_message')

# Check if subagent found issues
if echo "$MESSAGE" | grep -qi "incomplete\|failed"; then
  jq -n '{
    decision: "block",
    reason: "Please complete the investigation fully"
  }'
  exit 0
fi

exit 0
```

---

## 10. Stop

**Trigger:** When main Claude Code agent finishes responding
**Timing:** Post-response
**Frequency:** High (end of every response)
**Hook Types:** command, prompt, agent

### When It Fires

After Claude completes response. No matcher support.

**Important:** Does not run if user interrupts.

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "Stop",
  "stop_hook_active": true,  // Already continuing from hook
  "last_assistant_message": "I've completed the refactoring..."
}
```

### Decision Control

**Can prevent Claude from stopping:**

```json
{
  "decision": "block",
  "reason": "Required when blocking - tells Claude why to continue"
}
```

### Effects

- **"block":** Claude continues with reason as next instruction
- **Allow:** Claude stops normally

**WARNING:** Check `stop_hook_active` to prevent infinite loops!

### Common Use Cases

- Verify work complete
- Check tests pass
- Enforce quality gates
- Ensure all TODOs addressed
- Run final validation

### Example

```bash
#!/bin/bash
INPUT=$(cat)
ACTIVE=$(echo "$INPUT" | jq -r '.stop_hook_active')

# Prevent infinite loop
if [[ "$ACTIVE" == "true" ]]; then
  exit 0
fi

# Check if tests pass
if ! npm test &>/dev/null; then
  jq -n '{
    decision: "block",
    reason: "Tests are failing. Please fix before stopping."
  }'
  exit 0
fi

exit 0
```

---

## 11. TeammateIdle

**Trigger:** When agent team teammate going idle
**Timing:** Pre-idle
**Frequency:** Low (agent teams only)
**Hook Types:** command only

### When It Fires

When teammate about to go idle after finishing turn. No matcher support.

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "TeammateIdle",
  "teammate_name": "researcher",
  "team_name": "my-project"
}
```

### Decision Control

**Exit code only (no JSON):**
- Exit 0: Allow idle
- Exit 2: Prevent idle, stderr fed to teammate

### Effects

- **Exit 2:** Teammate receives stderr as feedback, continues working
- **Exit 0:** Teammate goes idle normally

### Common Use Cases

- Enforce quality gates
- Verify deliverables
- Check test status
- Ensure completeness

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TEAMMATE=$(echo "$INPUT" | jq -r '.teammate_name')

# Check teammate's work
if [[ "$TEAMMATE" == "implementer" ]]; then
  if ! npm test &>/dev/null; then
    echo "Tests failing. Fix before going idle." >&2
    exit 2
  fi
fi

exit 0
```

---

## 12. TaskCompleted

**Trigger:** When task being marked as completed
**Timing:** Pre-completion
**Frequency:** Low (when using tasks)
**Hook Types:** command, prompt, agent

### When It Fires

When any agent marks task complete via TaskUpdate tool, or agent team teammate finishes with in-progress tasks. No matcher support.

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "TaskCompleted",
  "task_id": "task-001",
  "task_subject": "Implement user authentication",
  "task_description": "Add login and signup endpoints",
  "teammate_name": "implementer",  // Optional
  "team_name": "my-project"  // Optional
}
```

### Decision Control

**Exit code only (no JSON):**
- Exit 0: Allow completion
- Exit 2: Prevent completion, stderr fed to model

### Effects

- **Exit 2:** Task not marked complete, model sees feedback
- **Exit 0:** Task completes normally

### Common Use Cases

- Verify completion criteria
- Check tests pass
- Validate deliverables
- Enforce quality standards

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TASK=$(echo "$INPUT" | jq -r '.task_subject')

# Run tests before completing
if ! npm test 2>&1; then
  echo "Tests failing for: $TASK. Fix before completing." >&2
  exit 2
fi

exit 0
```

---

## 13. ConfigChange

**Trigger:** When configuration file changes during session
**Timing:** During session
**Frequency:** Very low (config changes)
**Hook Types:** command only

### When It Fires

When settings files, policy settings, or skill files change. Matcher filters by config source.

### Matcher Values

| Matcher | File Changed |
|---------|--------------|
| `user_settings` | `~/.claude/settings.json` |
| `project_settings` | `.claude/settings.json` |
| `local_settings` | `.claude/settings.local.json` |
| `policy_settings` | Managed policy settings |
| `skills` | `.claude/skills/*` files |

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "ConfigChange",
  "source": "project_settings",
  "file_path": "/path/to/.claude/settings.json"
}
```

### Decision Control

**Can block config changes (except policy):**

```json
{
  "decision": "block",
  "reason": "Requires admin approval"
}
```

Or exit code 2.

**Exception:** `policy_settings` changes cannot be blocked (always apply).

### Effects

- **"block" or exit 2:** Config change not applied to running session
- **Allow:** Config change applied

### Common Use Cases

- Audit config changes
- Enforce approval workflow
- Security validation
- Log configuration updates

### Example

```bash
#!/bin/bash
INPUT=$(cat)
SOURCE=$(echo "$INPUT" | jq -r '.source')
FILE=$(echo "$INPUT" | jq -r '.file_path')

# Log all changes
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") Config changed: $SOURCE ($FILE)" >> ~/.claude/config-audit.log

# Block project settings changes (require approval)
if [[ "$SOURCE" == "project_settings" ]]; then
  echo "Project settings changes require team approval" >&2
  exit 2
fi

exit 0
```

---

## 14. WorktreeCreate

**Trigger:** When worktree being created
**Timing:** Pre-creation
**Frequency:** Very low (worktree usage)
**Hook Types:** command only

### When It Fires

When `--worktree` flag used or subagent uses `isolation: "worktree"`. No matcher support.

**Purpose:** Replace default git behavior with custom VCS (SVN, Perforce, Mercurial).

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "hook_event_name": "WorktreeCreate",
  "name": "feature-auth"  // Slug identifier
}
```

### Decision Control

**Hook must print absolute path to stdout:**

```bash
#!/bin/bash
NAME=$(jq -r .name)
DIR="$HOME/.claude/worktrees/$NAME"

# Create worktree (your VCS command)
svn checkout https://svn.example.com/repo/trunk "$DIR" >&2

# Print path for Claude Code
echo "$DIR"
```

Non-zero exit = failure.

### Effects

- **Success (exit 0 + print path):** Claude Code uses printed path
- **Failure (non-zero exit):** Worktree creation fails

### Common Use Cases

- Custom VCS integration (SVN, Perforce, Mercurial)
- Special checkout logic
- Template-based worktrees
- Pre-configured environments

---

## 15. WorktreeRemove

**Trigger:** When worktree being removed
**Timing:** Pre-removal
**Frequency:** Very low (worktree cleanup)
**Hook Types:** command only

### When It Fires

When exiting `--worktree` session (and choosing to remove), or when subagent with `isolation: "worktree"` finishes. No matcher support.

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "hook_event_name": "WorktreeRemove",
  "worktree_path": "/absolute/path/to/worktree"
}
```

### Decision Control

**No decision control.** Cannot block removal.

**Use for cleanup:**

```bash
#!/bin/bash
PATH=$(jq -r .worktree_path)

# Your cleanup logic
rm -rf "$PATH"
```

### Effects

- Hook performs cleanup
- Failures logged in debug mode only

### Common Use Cases

- VCS cleanup (paired with WorktreeCreate)
- Archive worktree
- Save state before removal
- Custom cleanup logic

---

## 16. PreCompact

**Trigger:** Before Claude Code runs compaction
**Timing:** Pre-compact
**Frequency:** Low (context window fills)
**Hook Types:** command only

### When It Fires

Before auto-compact (context full) or manual `/compact`. Matcher filters by trigger.

### Matcher Values

| Matcher | Trigger |
|---------|---------|
| `manual` | `/compact` command |
| `auto` | Automatic (context window full) |

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "PreCompact",
  "trigger": "manual|auto",
  "custom_instructions": "User's compact instructions or empty"
}
```

### Decision Control

**Cannot block compaction.**

Used for side effects (logging, notifications).

### Common Use Cases

- Log compaction events
- Notify user
- Save pre-compact state
- Analytics

### Example

```bash
#!/bin/bash
INPUT=$(cat)
TRIGGER=$(echo "$INPUT" | jq -r '.trigger')

# Log compaction
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") Compaction triggered: $TRIGGER" >> ~/.claude/compact.log

# Notify user
if [[ "$TRIGGER" == "auto" ]]; then
  echo "Auto-compaction happening now" >&2
fi

exit 0
```

---

## 17. SessionEnd

**Trigger:** When session terminates
**Timing:** Post-session
**Frequency:** Low (session cleanup)
**Hook Types:** command only

### When It Fires

When session ends. Matcher filters by reason.

### Matcher Values

| Matcher | Reason |
|---------|--------|
| `clear` | `/clear` command |
| `logout` | User logged out |
| `prompt_input_exit` | Exit during prompt input |
| `bypass_permissions_disabled` | Bypass permissions disabled |
| `other` | Other exit reasons |

### Input Schema

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/directory",
  "permission_mode": "default",
  "hook_event_name": "SessionEnd",
  "reason": "clear|logout|prompt_input_exit|bypass_permissions_disabled|other"
}
```

### Decision Control

**Cannot block termination.**

Used for cleanup and logging.

### Common Use Cases

- Cleanup temporary files
- Save session state
- Log session stats
- Send completion notifications
- Archive transcripts

### Example

```bash
#!/bin/bash
INPUT=$(cat)
REASON=$(echo "$INPUT" | jq -r '.reason')
TRANSCRIPT=$(echo "$INPUT" | jq -r '.transcript_path')

# Log session end
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") Session ended: $REASON" >> ~/.claude/sessions.log

# Archive transcript
cp "$TRANSCRIPT" "$HOME/.claude/archive/$(basename "$TRANSCRIPT")"

# Cleanup
rm -f /tmp/claude-*.tmp

exit 0
```

---

## Tool Input Schemas

### Bash

```json
{
  "command": "npm test",
  "description": "Run test suite",  // Optional
  "timeout": 120000,  // Optional (milliseconds)
  "run_in_background": false  // Optional
}
```

### Write

```json
{
  "file_path": "/absolute/path/to/file.txt",
  "content": "file contents here"
}
```

### Edit

```json
{
  "file_path": "/absolute/path/to/file.txt",
  "old_string": "text to find",
  "new_string": "replacement text",
  "replace_all": false  // Optional
}
```

### Read

```json
{
  "file_path": "/absolute/path/to/file.txt",
  "offset": 10,  // Optional (line number to start)
  "limit": 50  // Optional (number of lines)
}
```

### Glob

```json
{
  "pattern": "**/*.ts",
  "path": "/search/directory"  // Optional (defaults to cwd)
}
```

### Grep

```json
{
  "pattern": "TODO.*fix",
  "path": "/search/path",  // Optional
  "glob": "*.ts",  // Optional
  "output_mode": "content|files_with_matches|count",  // Optional
  "-i": true,  // Optional (case insensitive)
  "multiline": false  // Optional
}
```

### WebFetch

```json
{
  "url": "https://example.com/api",
  "prompt": "Extract the API endpoints"
}
```

### WebSearch

```json
{
  "query": "react hooks best practices",
  "allowed_domains": ["docs.react.dev"],  // Optional
  "blocked_domains": ["spam.com"]  // Optional
}
```

### Task (Subagent)

```json
{
  "prompt": "Find all API endpoints in the codebase",
  "description": "API endpoint discovery",  // Optional
  "subagent_type": "Explore|Bash|Plan|custom",  // Optional
  "model": "sonnet"  // Optional (model alias)
}
```

---

## Event Categories

### Pre-Execution (Can Block)

- PreToolUse
- PermissionRequest
- UserPromptSubmit
- Stop
- SubagentStop
- TeammateIdle
- TaskCompleted
- ConfigChange (except policy)
- WorktreeCreate

### Post-Execution (Cannot Block)

- PostToolUse
- PostToolUseFailure
- Notification
- SubagentStart
- SessionEnd
- PreCompact
- WorktreeRemove

### Session Lifecycle

- SessionStart
- UserPromptSubmit
- PreCompact
- SessionEnd

### Tool Lifecycle

- PreToolUse
- PermissionRequest
- PostToolUse
- PostToolUseFailure

### Agent/Task Lifecycle

- SubagentStart
- SubagentStop
- TeammateIdle
- TaskCompleted

### System Events

- ConfigChange
- WorktreeCreate
- WorktreeRemove
- Notification

---

## Sources

- **Official Documentation:** https://code.claude.com/docs/en/hooks
- **Hooks Guide:** https://code.claude.com/docs/en/hooks-guide

Research date: 2026-02-21
