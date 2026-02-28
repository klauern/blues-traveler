# Claude Code Official Documentation - Research Notes

**Research Date:** 2026-02-21
**Primary Source:** https://code.claude.com/docs/en/hooks
**Status:** ✅ Complete

---

## Overview

Claude Code is Anthropic's official CLI tool for AI-assisted software development. The hooks system allows developers to automate workflows, enforce policies, and customize behavior through event-driven callbacks.

**Key Concept:** Hooks are user-defined shell commands, LLM prompts, or subagents that execute automatically at specific points in Claude Code's lifecycle.

---

## Hook Types

Claude Code supports three types of hooks:

1. **Command hooks** (`type: "command"`) - Run shell commands
2. **Prompt hooks** (`type: "prompt"`) - Single-turn LLM evaluation
3. **Agent hooks** (`type: "agent"`) - Multi-turn subagent with tool access

---

## Complete Event Catalog

| Event | Trigger | Timing | Cancellable | Matcher Support |
|-------|---------|--------|-------------|-----------------|
| `SessionStart` | Session begins/resumes | Once per session | No | Yes (startup/resume/clear/compact) |
| `UserPromptSubmit` | User submits prompt | Before processing | Yes | No |
| `PreToolUse` | Before tool execution | Pre-action | Yes | Yes (tool name) |
| `PermissionRequest` | Permission dialog appears | Pre-permission | Yes | Yes (tool name) |
| `PostToolUse` | After tool success | Post-action | No | Yes (tool name) |
| `PostToolUseFailure` | After tool failure | Post-action | No | Yes (tool name) |
| `Notification` | Claude sends notification | During | No | Yes (notification type) |
| `SubagentStart` | Subagent spawned | Pre-subagent | No | Yes (agent type) |
| `SubagentStop` | Subagent finishes | Post-subagent | Yes | Yes (agent type) |
| `Stop` | Claude finishes responding | Post-response | Yes | No |
| `TeammateIdle` | Team teammate going idle | Pre-idle | Yes | No |
| `TaskCompleted` | Task marked complete | Pre-completion | Yes | No |
| `ConfigChange` | Config file changes | During | Yes (except policy) | Yes (config source) |
| `WorktreeCreate` | Worktree being created | Pre-creation | Yes | No |
| `WorktreeRemove` | Worktree being removed | Pre-removal | No | No |
| `PreCompact` | Before compaction | Pre-compact | No | Yes (manual/auto) |
| `SessionEnd` | Session terminates | Post-session | No | Yes (reason) |

**Total Events:** 17 documented events

---

## Hook Lifecycle Diagram

From official documentation:

```
SessionStart
    ↓
UserPromptSubmit
    ↓
┌───────────────────┐
│  Agentic Loop     │
│                   │
│  PreToolUse       │
│  PermissionRequest│
│  PostToolUse      │
│  PostToolUseFailure│
│  Notification     │
│  SubagentStart    │
│  SubagentStop     │
│  Stop             │
│  TeammateIdle     │
│  TaskCompleted    │
│                   │
└───────────────────┘
    ↓
PreCompact (if needed)
    ↓
SessionEnd

Standalone:
- WorktreeCreate (setup)
- WorktreeRemove (teardown)
- ConfigChange (anytime)
```

---

## Configuration Schema

### Location Hierarchy

| Location | Scope | Shareable |
|----------|-------|-----------|
| `~/.claude/settings.json` | All projects | No (local machine) |
| `.claude/settings.json` | Single project | Yes (git committable) |
| `.claude/settings.local.json` | Single project | No (gitignored) |
| Managed policy settings | Organization | Yes (admin-controlled) |
| Plugin `hooks/hooks.json` | When enabled | Yes (bundled) |
| Skill/agent frontmatter | Component active | Yes (in component) |

### Basic Structure

```json
{
  "hooks": {
    "EventName": [
      {
        "matcher": "regex_pattern",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/script.sh",
            "timeout": 30,
            "statusMessage": "Custom message...",
            "async": false
          }
        ]
      }
    ]
  }
}
```

### Common Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | Yes | - | "command", "prompt", or "agent" |
| `timeout` | number | No | 600 (command), 30 (prompt), 60 (agent) | Seconds before canceling |
| `statusMessage` | string | No | - | Custom spinner message |
| `once` | boolean | No | false | Run once per session (skills only) |

### Command Hook Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | Yes | Shell command to execute |
| `async` | boolean | No | Run in background without blocking |

### Prompt/Agent Hook Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | Prompt text (use `$ARGUMENTS` placeholder) |
| `model` | string | No | Model for evaluation (defaults to fast model) |

---

## Environment Variables

### System-Provided Variables

| Variable | Available In | Description |
|----------|--------------|-------------|
| `$CLAUDE_PROJECT_DIR` | All events | Project root directory |
| `$CLAUDE_PLUGIN_ROOT` | Plugin hooks | Plugin's root directory |
| `$CLAUDE_CODE_REMOTE` | All events | Set to "true" in remote web environments |
| `$CLAUDE_ENV_FILE` | SessionStart only | File path for persisting env vars |

### Common Input Fields (JSON via stdin)

All hooks receive these fields:

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/directory",
  "permission_mode": "default|plan|acceptEdits|dontAsk|bypassPermissions",
  "hook_event_name": "EventName"
}
```

---

## Exit Code Behavior

| Exit Code | Meaning | Effect |
|-----------|---------|--------|
| `0` | Success | Parse stdout for JSON, continue normally |
| `2` | Blocking error | Feed stderr to Claude, block action (if cancellable) |
| Other | Non-blocking error | Log stderr in verbose mode, continue |

### Exit Code 2 Effects by Event

| Event | Can Block? | Effect |
|-------|------------|--------|
| PreToolUse | Yes | Blocks tool call |
| PermissionRequest | Yes | Denies permission |
| UserPromptSubmit | Yes | Blocks prompt, erases from context |
| Stop | Yes | Prevents stopping, continues |
| SubagentStop | Yes | Prevents subagent from stopping |
| TeammateIdle | Yes | Keeps teammate working |
| TaskCompleted | Yes | Prevents task completion |
| ConfigChange | Yes | Blocks config change (except policy) |
| WorktreeCreate | Yes | Fails worktree creation |
| PostToolUse | No | Shows stderr to Claude |
| PostToolUseFailure | No | Shows stderr to Claude |
| Notification | No | Shows stderr to user |
| SubagentStart | No | Shows stderr to user |
| SessionStart | No | Shows stderr to user |
| SessionEnd | No | Shows stderr to user |
| PreCompact | No | Shows stderr to user |
| WorktreeRemove | No | Logged in debug mode only |

---

## JSON Output Schema

### Universal Fields

Available to all hooks when exiting 0:

```json
{
  "continue": false,  // default: true
  "stopReason": "Message for user",
  "suppressOutput": false,  // default: false
  "systemMessage": "Warning message"
}
```

### Decision Control Patterns

**Pattern 1: Top-level decision** (UserPromptSubmit, PostToolUse, PostToolUseFailure, Stop, SubagentStop, ConfigChange)

```json
{
  "decision": "block",
  "reason": "Explanation for Claude"
}
```

**Pattern 2: PreToolUse decision**

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

**Pattern 3: PermissionRequest decision**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PermissionRequest",
    "decision": {
      "behavior": "allow|deny",
      "updatedInput": { "command": "modified" },
      "updatedPermissions": {},
      "message": "Denial reason",
      "interrupt": false
    }
  }
}
```

---

## Matcher Patterns

Matchers are regex strings that filter when hooks fire.

### Matcher Values by Event

| Event | Matches On | Example Values |
|-------|-----------|----------------|
| PreToolUse, PostToolUse, etc. | Tool name | `Bash`, `Edit\|Write`, `mcp__.*` |
| SessionStart | Session source | `startup`, `resume`, `clear`, `compact` |
| SessionEnd | Exit reason | `clear`, `logout`, `prompt_input_exit`, etc. |
| Notification | Notification type | `permission_prompt`, `idle_prompt`, etc. |
| SubagentStart/Stop | Agent type | `Bash`, `Explore`, `Plan`, custom names |
| PreCompact | Trigger type | `manual`, `auto` |
| ConfigChange | Config source | `user_settings`, `project_settings`, etc. |

### MCP Tool Matching

MCP tools follow naming: `mcp__<server>__<tool>`

Examples:
- `mcp__memory__create_entities`
- `mcp__filesystem__read_file`
- `mcp__github__search_repositories`

Match patterns:
- `mcp__memory__.*` - all memory server tools
- `mcp__.*__write.*` - any write tool from any server

---

## Event-Specific Details

### PreToolUse

**Input Schema:**

```json
{
  "tool_name": "Bash|Edit|Write|Read|Glob|Grep|Task|WebFetch|WebSearch|mcp__*",
  "tool_input": {
    // Tool-specific fields
  },
  "tool_use_id": "toolu_..."
}
```

**Tool Input Schemas:**

**Bash:**
```json
{
  "command": "npm test",
  "description": "Run test suite",
  "timeout": 120000,
  "run_in_background": false
}
```

**Write:**
```json
{
  "file_path": "/absolute/path/file.txt",
  "content": "file content"
}
```

**Edit:**
```json
{
  "file_path": "/absolute/path/file.txt",
  "old_string": "original",
  "new_string": "replacement",
  "replace_all": false
}
```

**Read:**
```json
{
  "file_path": "/absolute/path/file.txt",
  "offset": 10,
  "limit": 50
}
```

**Glob:**
```json
{
  "pattern": "**/*.ts",
  "path": "/search/directory"
}
```

**Grep:**
```json
{
  "pattern": "TODO.*fix",
  "path": "/search/path",
  "glob": "*.ts",
  "output_mode": "content|files_with_matches|count",
  "-i": true,
  "multiline": false
}
```

**WebFetch:**
```json
{
  "url": "https://example.com",
  "prompt": "Extract endpoints"
}
```

**WebSearch:**
```json
{
  "query": "search terms",
  "allowed_domains": ["docs.example.com"],
  "blocked_domains": ["spam.com"]
}
```

**Task:**
```json
{
  "prompt": "Task description",
  "description": "Short desc",
  "subagent_type": "Explore|Bash|Plan|custom",
  "model": "sonnet"
}
```

### PostToolUse

**Input Schema:**

```json
{
  "tool_name": "Write",
  "tool_input": { /* same as PreToolUse */ },
  "tool_response": {
    "filePath": "/path/to/file.txt",
    "success": true
  },
  "tool_use_id": "toolu_..."
}
```

**Decision Control:**

```json
{
  "decision": "block",
  "reason": "Tests failed",
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "Additional info",
    "updatedMCPToolOutput": { /* MCP tools only */ }
  }
}
```

### SessionStart

**Input Schema:**

```json
{
  "source": "startup|resume|clear|compact",
  "model": "claude-sonnet-4-6",
  "agent_type": "agent_name"  // if --agent used
}
```

**Decision Control:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Added to Claude's context"
  }
}
```

**Special Feature:** `$CLAUDE_ENV_FILE` environment variable

```bash
#!/bin/bash
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export NODE_ENV=production' >> "$CLAUDE_ENV_FILE"
  echo 'export PATH="$PATH:./bin"' >> "$CLAUDE_ENV_FILE"
fi
```

### UserPromptSubmit

**Input Schema:**

```json
{
  "prompt": "User's submitted text"
}
```

**Decision Control:**

```json
{
  "decision": "block",  // prevents processing, erases prompt
  "reason": "Shown to user",
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "Added to context"
  }
}
```

Or output plain text to stdout (exit 0) to add context without JSON.

### Stop & SubagentStop

**Input Schema:**

```json
{
  "stop_hook_active": true,  // already continuing from hook
  "last_assistant_message": "Claude's final response text",
  // SubagentStop adds:
  "agent_id": "def456",
  "agent_type": "Explore",
  "agent_transcript_path": "path/to/subagent/transcript.jsonl"
}
```

**Decision Control:**

```json
{
  "decision": "block",  // prevents stopping
  "reason": "Required when blocking"
}
```

### TeammateIdle

**Input Schema:**

```json
{
  "teammate_name": "researcher",
  "team_name": "my-project"
}
```

**Decision:** Exit code 2 only (no JSON). Stderr fed to teammate.

### TaskCompleted

**Input Schema:**

```json
{
  "task_id": "task-001",
  "task_subject": "Title",
  "task_description": "Details",
  "teammate_name": "implementer",
  "team_name": "my-project"
}
```

**Decision:** Exit code 2 only (no JSON).

### ConfigChange

**Input Schema:**

```json
{
  "source": "user_settings|project_settings|local_settings|policy_settings|skills",
  "file_path": "/path/to/changed/file"
}
```

**Decision Control:**

```json
{
  "decision": "block",  // blocks change (except policy_settings)
  "reason": "Requires admin approval"
}
```

### WorktreeCreate

**Input Schema:**

```json
{
  "name": "feature-auth"  // slug identifier
}
```

**Output:** Print absolute path to stdout. Non-zero exit = failure.

```bash
#!/bin/bash
NAME=$(jq -r .name)
DIR="$HOME/.claude/worktrees/$NAME"
svn checkout https://svn.example.com/repo/trunk "$DIR" >&2
echo "$DIR"  # This path is used by Claude Code
```

### WorktreeRemove

**Input Schema:**

```json
{
  "worktree_path": "/absolute/path/to/worktree"
}
```

**Behavior:** No decision control. Failures logged in debug only.

### Notification

**Input Schema:**

```json
{
  "message": "Notification text",
  "title": "Notification title",
  "notification_type": "permission_prompt|idle_prompt|auth_success|elicitation_dialog"
}
```

**Decision Control:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "Notification",
    "additionalContext": "Added to context"
  }
}
```

### SubagentStart

**Input Schema:**

```json
{
  "agent_id": "agent-abc123",
  "agent_type": "Explore|Bash|Plan|custom"
}
```

**Decision Control:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SubagentStart",
    "additionalContext": "Context for subagent"
  }
}
```

### PostToolUseFailure

**Input Schema:**

```json
{
  "tool_name": "Bash",
  "tool_input": { /* same as PreToolUse */ },
  "tool_use_id": "toolu_...",
  "error": "Error description",
  "is_interrupt": false  // optional
}
```

**Decision Control:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUseFailure",
    "additionalContext": "Additional context about failure"
  }
}
```

### PreCompact

**Input Schema:**

```json
{
  "trigger": "manual|auto",
  "custom_instructions": "User's compact instructions or empty"
}
```

**Behavior:** Cannot block compaction.

### SessionEnd

**Input Schema:**

```json
{
  "reason": "clear|logout|prompt_input_exit|bypass_permissions_disabled|other"
}
```

**Behavior:** Cannot block termination.

---

## Prompt-Based Hooks

Events supporting `type: "prompt"`:
- PermissionRequest
- PostToolUse
- PostToolUseFailure
- PreToolUse
- Stop
- SubagentStop
- TaskCompleted
- UserPromptSubmit

**Configuration:**

```json
{
  "type": "prompt",
  "prompt": "Evaluate if Claude should stop: $ARGUMENTS",
  "model": "haiku",
  "timeout": 30
}
```

**Response Schema:**

```json
{
  "ok": true,  // or false
  "reason": "Required when ok is false"
}
```

---

## Agent-Based Hooks

Same events as prompt hooks.

**Configuration:**

```json
{
  "type": "agent",
  "prompt": "Verify tests pass. $ARGUMENTS",
  "model": "sonnet",
  "timeout": 120
}
```

Agent can use tools: Read, Grep, Glob (up to 50 turns).

**Response:** Same as prompt hooks (`ok` + `reason`).

---

## Async Hooks

**Only for `type: "command"` hooks.**

```json
{
  "type": "command",
  "command": "/path/to/long-running.sh",
  "async": true,
  "timeout": 300
}
```

**Limitations:**
- Cannot block actions
- Decision fields ignored
- Output delivered on next conversation turn
- No deduplication across firings

---

## Security Model

### Warnings

**From official docs:**
> Hooks execute shell commands with your full user permissions. They can modify, delete, or access any files your user account can access.

### Best Practices

1. **Validate and sanitize inputs**
2. **Always quote shell variables:** `"$VAR"` not `$VAR`
3. **Block path traversal:** Check for `..` in paths
4. **Use absolute paths:** Specify full paths with `$CLAUDE_PROJECT_DIR`
5. **Skip sensitive files:** Avoid `.env`, `.git/`, keys

### Permission Modes

| Mode | Description |
|------|-------------|
| `default` | Standard permission prompts |
| `plan` | Plan mode |
| `acceptEdits` | Auto-accept edits |
| `dontAsk` | Minimal prompts |
| `bypassPermissions` | No prompts (dangerous) |

---

## Debug & Troubleshooting

### Debug Mode

```bash
claude --debug
```

Shows:
- Hook execution details
- Matching logic
- Exit codes
- Output

### Verbose Mode

Press `Ctrl+O` to toggle verbose mode (shows hook progress in transcript).

### Debug Output Example

```
[DEBUG] Executing hooks for PostToolUse:Write
[DEBUG] Getting matching hook commands for PostToolUse with query: Write
[DEBUG] Found 1 hook matchers in settings
[DEBUG] Matched 1 hooks for query "Write"
[DEBUG] Found 1 hook commands to execute
[DEBUG] Executing hook command: <command> with timeout 600000ms
[DEBUG] Hook command completed with status 0: <stdout>
```

---

## Management Features

### The `/hooks` Menu

Interactive UI for managing hooks:
- View all configured hooks
- Add new hooks
- Delete existing hooks
- Toggle `disableAllHooks`

**Hook Source Labels:**
- `[User]` - from `~/.claude/settings.json`
- `[Project]` - from `.claude/settings.json`
- `[Local]` - from `.claude/settings.local.json`
- `[Plugin]` - from plugin (read-only)

### Disabling Hooks

**Temporary disable all:**

```json
{
  "disableAllHooks": true
}
```

Or use toggle in `/hooks` menu.

**Note:** Managed policy hooks cannot be disabled by user/project settings.

### Hook Modification Behavior

**Important:** Direct edits to hooks in settings files don't take effect immediately.

Claude Code captures hooks at startup. Mid-session changes require review in `/hooks` menu before applying (security feature).

---

## Limitations & Gotchas

### Known Issues

1. **Environment variable substitution** - Some users report issues with `CLAUDE_TOOL_NAME` and `CLAUDE_TOOL_PARAMS` not substituting properly
2. **JSON parsing** - Shell profile output can interfere with JSON parsing
3. **Stop hook loops** - Watch `stop_hook_active` to prevent infinite loops
4. **Matcher regex** - Uses full regex, not glob patterns (except in job-level `glob` field)

### Events Without Matcher Support

These always fire:
- UserPromptSubmit
- Stop
- TeammateIdle
- TaskCompleted
- WorktreeCreate
- WorktreeRemove

---

## Sources

Primary documentation:
- **Official Hooks Reference:** https://code.claude.com/docs/en/hooks
- **Hooks Guide:** https://code.claude.com/docs/en/hooks-guide
- **Overview:** https://code.claude.com/docs

Research date: 2026-02-21
