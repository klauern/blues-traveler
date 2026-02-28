# Claude Code Hooks - Configuration & Setup

**Version:** 1.0
**Last Updated:** 2026-02-21
**Status:** Complete

---

## Table of Contents

1. [Configuration File Locations](#configuration-file-locations)
2. [Configuration Format](#configuration-format)
3. [Schema Reference](#schema-reference)
4. [Merging Logic](#merging-logic)
5. [Hook Types Configuration](#hook-types-configuration)
6. [Matcher Patterns](#matcher-patterns)
7. [Setup Instructions](#setup-instructions)
8. [Management & Updates](#management--updates)
9. [Troubleshooting](#troubleshooting)

---

## Configuration File Locations

Claude Code reads hook configurations from multiple locations, merging them according to precedence rules.

### File Hierarchy

| Location | Scope | Shareable | Precedence | Gitignore |
|----------|-------|-----------|------------|-----------|
| Managed policy settings | Organization | Admin-controlled | 1 (highest) | N/A |
| `~/.claude/settings.json` | All projects (global) | No (local machine) | 2 | N/A |
| `.claude/settings.json` | Single project | Yes (committable) | 3 | No |
| `.claude/settings.local.json` | Single project | No (local overrides) | 4 | **Yes** |
| Plugin `hooks/hooks.json` | When plugin enabled | Yes (bundled) | 5 | N/A |
| Skill/Agent frontmatter | Component-specific | Yes (in YAML) | 6 (lowest) | N/A |

**Source:** [Official Hooks Documentation](https://code.claude.com/docs/en/hooks)

---

### Global Configuration

**Path:** `~/.claude/settings.json`

**Use for:**
- Personal preferences across all projects
- Security hooks (dangerous command blocking)
- Global notification hooks
- Personal workflow automation

**Example:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$HOME/.claude/hooks/security-check.sh"
          }
        ]
      }
    ]
  }
}
```

---

### Project Configuration

**Path:** `.claude/settings.json`

**Use for:**
- Project-specific workflows
- Team-shared hooks
- Build automation
- Project formatting standards

**Example:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format.sh",
            "statusMessage": "Auto-formatting code..."
          }
        ]
      }
    ]
  }
}
```

**Shareable:** Commit to git for team use

---

### Local Project Configuration

**Path:** `.claude/settings.local.json`

**Use for:**
- Machine-specific overrides
- Local testing hooks
- Personal preferences for shared project
- Credentials/API keys (keep gitignored!)

**Example:**
```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$HOME/local-dev-setup.sh"
          }
        ]
      }
    ]
  }
}
```

**Must gitignore:** Add to `.gitignore`:
```
.claude/settings.local.json
```

**Source:** [eesel.ai Developer Guide](https://www.eesel.ai/blog/settings-json-claude-code)

---

## Configuration Format

### Basic Structure

All hook configurations follow this JSON schema:

```json
{
  "hooks": {
    "EventName": [
      {
        "matcher": "optional_regex_pattern",
        "hooks": [
          {
            "type": "command|prompt|agent",
            "command": "/path/to/script.sh",  // command type only
            "prompt": "LLM prompt text",      // prompt/agent only
            "model": "haiku|sonnet",          // prompt/agent only
            "timeout": 30,                    // optional
            "statusMessage": "Custom message", // optional
            "async": false,                   // command only
            "once": false                     // skills only
          }
        ]
      }
    ]
  }
}
```

**Source:** [Hooks Reference](https://code.claude.com/docs/en/hooks)

---

### Minimal Example

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "prettier --write $FILE"
          }
        ]
      }
    ]
  }
}
```

---

### Complete Example

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/init-context.sh",
            "timeout": 10,
            "statusMessage": "Loading project context..."
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/security-check.sh",
            "timeout": 5
          }
        ]
      },
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/branch-protection.sh",
            "timeout": 3
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format-and-test.sh",
            "timeout": 60,
            "statusMessage": "Formatting and testing...",
            "async": false
          },
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/notify-slack.sh",
            "async": true
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "agent",
            "prompt": "Before stopping, verify all tests pass. $ARGUMENTS",
            "timeout": 120
          }
        ]
      }
    ]
  },
  "disableAllHooks": false
}
```

**Source:** Composite from [GitHub Examples](https://github.com/disler/claude-code-hooks-mastery)

---

## Schema Reference

### Top-Level Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `hooks` | object | No | `{}` | Event-to-hook mappings |
| `disableAllHooks` | boolean | No | `false` | Disable all user/project hooks (not policy) |

---

### Event Configuration

Each event key (e.g., `"PreToolUse"`) contains an array of matcher-hook pairs:

```json
{
  "EventName": [
    {
      "matcher": "regex",
      "hooks": [...]
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `matcher` | string | No | Regex to filter when hooks fire (event-specific) |
| `hooks` | array | Yes | Array of hook definitions |

---

### Hook Definition

| Field | Type | Required | Default | Applies To | Description |
|-------|------|----------|---------|------------|-------------|
| `type` | string | Yes | - | All | `"command"`, `"prompt"`, or `"agent"` |
| `command` | string | Yes* | - | command | Shell command to execute |
| `prompt` | string | Yes* | - | prompt, agent | LLM prompt (use `$ARGUMENTS` placeholder) |
| `model` | string | No | `haiku` (prompt), `sonnet` (agent) | prompt, agent | Model for evaluation |
| `timeout` | number | No | 600 (cmd), 30 (prompt), 60 (agent) | All | Timeout in seconds |
| `statusMessage` | string | No | - | All | Custom spinner message |
| `async` | boolean | No | `false` | command | Run in background |
| `once` | boolean | No | `false` | All (skills only) | Run once per session |

\* Required for respective type

**Source:** [Official Hooks Reference](https://code.claude.com/docs/en/hooks)

---

## Merging Logic

### Merge Strategy

Claude Code **concatenates** hooks from all sources (no deduplication):

```
Policy hooks
    ↓
User hooks (appended)
    ↓
Project hooks (appended)
    ↓
Local hooks (appended)
    ↓
Plugin hooks (appended)
    ↓
Skill hooks (appended)
```

**Example:**

**User config (`~/.claude/settings.json`):**
```json
{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "~/check.sh"}]}
    ]
  }
}
```

**Project config (`.claude/settings.json`):**
```json
{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Write", "hooks": [{"type": "command", "command": "./format.sh"}]}
    ]
  }
}
```

**Merged result:**
```json
{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "~/check.sh"}]},
      {"matcher": "Write", "hooks": [{"type": "command", "command": "./format.sh"}]}
    ]
  }
}
```

**Both hooks execute** when PreToolUse fires (if matchers match).

---

### Execution Order

Within the same event:
1. Policy hooks run first
2. User hooks
3. Project hooks
4. Local hooks
5. Plugin hooks
6. Skill hooks

**All matching hooks execute sequentially** (not parallel).

**Source:** [Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

---

### Override Behavior

**Cannot override:** Hooks from higher precedence sources execute alongside lower precedence hooks.

**To disable inherited hooks:**
- Set `"disableAllHooks": true` in lower-precedence config
- **Note:** Only disables user/project/local hooks, not policy hooks

---

## Hook Types Configuration

### Command Hooks

**Type:** `"command"`

```json
{
  "type": "command",
  "command": "/absolute/path/to/script.sh",
  "timeout": 30,
  "statusMessage": "Running security checks...",
  "async": false
}
```

**Fields:**
- `command` (required): Shell command or script path
- `timeout`: Seconds before kill (default: 600)
- `statusMessage`: Custom spinner text
- `async`: Run in background (default: false)

**Path Resolution:**
- Absolute paths recommended
- Use `$CLAUDE_PROJECT_DIR` for project-relative paths
- Use `$CLAUDE_PLUGIN_ROOT` for plugin scripts
- Environment variables expanded by shell

---

### Prompt Hooks

**Type:** `"prompt"`

```json
{
  "type": "prompt",
  "prompt": "Is this command safe to run? Command: $ARGUMENTS. Respond with {\"ok\": true} or {\"ok\": false, \"reason\": \"why\"}",
  "model": "haiku",
  "timeout": 30
}
```

**Fields:**
- `prompt` (required): LLM evaluation prompt
- `model`: `"haiku"` (fast) or `"sonnet"` (powerful)
- `timeout`: Seconds (default: 30)
- Use `$ARGUMENTS` placeholder for event context

**Supported Events:**
- PermissionRequest
- PostToolUse
- PostToolUseFailure
- PreToolUse
- Stop
- SubagentStop
- TaskCompleted
- UserPromptSubmit

**Response Format:**
```json
{
  "ok": true,
  "reason": "Required when ok=false"
}
```

---

### Agent Hooks

**Type:** `"agent"`

```json
{
  "type": "agent",
  "prompt": "Verify all tests pass before stopping. Use Read and Grep to check. $ARGUMENTS",
  "model": "sonnet",
  "timeout": 120
}
```

**Fields:**
- `prompt` (required): Instructions for agent
- `model`: Usually `"sonnet"` for complex tasks
- `timeout`: Seconds (default: 60)

**Agent Capabilities:**
- Multi-turn conversation (up to 50 turns)
- Tools: Read, Grep, Glob
- Cannot execute Bash commands
- Cannot write/edit files

**Response Format:** Same as prompt hooks

**Source:** [Hooks Guide - Agent Hooks](https://code.claude.com/docs/en/hooks-guide#agent-hooks)

---

## Matcher Patterns

### Regex Syntax

Matchers use full JavaScript regex syntax (not glob patterns).

**Common Patterns:**

| Pattern | Matches | Example Use |
|---------|---------|-------------|
| `Bash` | Exact tool name | PreToolUse with Bash tool |
| `Write\|Edit` | Either Write or Edit | File modification tools |
| `mcp__.*` | Any MCP tool | All MCP server tools |
| `mcp__memory__.*` | All memory tools | Specific MCP server |
| `startup` | SessionStart with startup | New sessions only |
| `permission_prompt` | Notification type | Permission dialogs |

---

### Event-Specific Matcher Values

| Event | Matches On | Example Values |
|-------|-----------|----------------|
| SessionStart | Session source | `startup`, `resume`, `clear`, `compact` |
| UserPromptSubmit | - | No matcher support |
| PreToolUse | Tool name | `Bash`, `Edit`, `Write`, `Read`, `Glob`, `Grep`, `Task`, `WebFetch`, `WebSearch`, `mcp__*` |
| PermissionRequest | Tool name | Same as PreToolUse |
| PostToolUse | Tool name | Same as PreToolUse |
| PostToolUseFailure | Tool name | Same as PreToolUse |
| Notification | Notification type | `permission_prompt`, `idle_prompt`, `auth_success`, `elicitation_dialog` |
| SubagentStart | Agent type | `Bash`, `Explore`, `Plan`, custom agent names |
| SubagentStop | Agent type | Same as SubagentStart |
| Stop | - | No matcher support |
| TeammateIdle | - | No matcher support |
| TaskCompleted | - | No matcher support |
| ConfigChange | Config source | `user_settings`, `project_settings`, `local_settings`, `policy_settings`, `skills` |
| WorktreeCreate | - | No matcher support |
| WorktreeRemove | - | No matcher support |
| PreCompact | Trigger type | `manual`, `auto` |
| SessionEnd | Exit reason | `clear`, `logout`, `prompt_input_exit`, `bypass_permissions_disabled`, `other` |

**Source:** [Events Reference](./events-reference.md)

---

### MCP Tool Patterns

MCP tools follow naming: `mcp__<server>__<tool>`

**Examples:**
```json
{
  "matcher": "mcp__memory__create_entities",  // Specific tool
  "hooks": [...]
}
```

```json
{
  "matcher": "mcp__memory__.*",  // All memory tools
  "hooks": [...]
}
```

```json
{
  "matcher": "mcp__.*__write.*",  // Any write tool from any MCP server
  "hooks": [...]
}
```

---

## Setup Instructions

### Initial Setup

**1. Create configuration directory:**
```bash
mkdir -p .claude/hooks
```

**2. Create settings file:**
```bash
touch .claude/settings.json
```

**3. Add basic configuration:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Hook triggered!'",
            "statusMessage": "Running post-edit hook..."
          }
        ]
      }
    ]
  }
}
```

**4. Test hook:**
```bash
claude
```

Then ask Claude to write a file and observe the hook execution.

---

### Verification

**Check hook configuration:**
```bash
claude
```

Then use `/hooks` command to view and manage hooks interactively.

**Debug mode:**
```bash
claude --debug
```

Shows detailed hook execution information.

**Source:** [Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

---

## Management & Updates

### The `/hooks` Interactive Menu

**Access:** Type `/hooks` in any Claude Code session

**Features:**
- View all configured hooks
- See hook sources ([User], [Project], [Local], [Plugin])
- Add new hooks (guided UI)
- Delete existing hooks
- Toggle `disableAllHooks`
- Review pending configuration changes

**Source Labels:**
- `[User]` - from `~/.claude/settings.json`
- `[Project]` - from `.claude/settings.json`
- `[Local]` - from `.claude/settings.local.json`
- `[Plugin]` - from plugin (read-only, cannot delete)
- `[Policy]` - from managed policy (read-only, cannot disable)

**Best Practice:** Use `/hooks` menu instead of manually editing JSON to avoid syntax errors.

**Source:** [eesel.ai Guide](https://www.eesel.ai/blog/settings-json-claude-code)

---

### Mid-Session Changes

**Important Behavior:**
- Hook configurations captured at session start
- Direct edits to settings files don't apply immediately
- Must review in `/hooks` menu before changes take effect

**To Apply Changes:**
1. Edit `.claude/settings.json` or `~/.claude/settings.json`
2. Open `/hooks` menu in active session
3. Review pending changes
4. Accept to apply to current session

**Security Feature:** Prevents malicious config changes from taking effect without user awareness.

---

### Disabling Hooks

**Temporary disable all:**

Add to any settings file:
```json
{
  "disableAllHooks": true
}
```

**Or:** Toggle in `/hooks` menu

**Effects:**
- Disables all user, project, and local hooks
- **Does not** disable managed policy hooks
- Per-session setting (doesn't persist)

**Source:** [Hooks Reference](https://code.claude.com/docs/en/hooks)

---

## Troubleshooting

### Common Configuration Errors

#### Invalid JSON Syntax

**Error:** Hooks don't execute, debug mode shows parse errors

**Solution:**
```bash
# Validate JSON
cat .claude/settings.json | jq empty
```

Fix syntax errors (trailing commas, missing quotes, etc.)

---

#### Hook Not Executing

**Symptoms:** Hook script never runs

**Checklist:**
1. Verify matcher pattern matches tool/event
2. Check file permissions (`chmod +x script.sh`)
3. Verify script has shebang (`#!/bin/bash`)
4. Use debug mode: `claude --debug`
5. Check `/hooks` menu for configuration

**Debug:**
```bash
claude --debug
```

Look for:
```
[DEBUG] Getting matching hook commands for PreToolUse with query: Bash
[DEBUG] Matched 0 hooks for query "Bash"
```

---

#### Timeout Issues

**Error:** Hook killed after timeout

**Solutions:**
1. Increase timeout:
   ```json
   {
     "type": "command",
     "command": "/path/to/slow-script.sh",
     "timeout": 300  // 5 minutes
   }
   ```

2. Make async:
   ```json
   {
     "type": "command",
     "command": "/path/to/slow-script.sh",
     "async": true
   }
   ```

3. Optimize hook script

---

#### Path Resolution Issues

**Error:** Command not found

**Solution:** Use absolute paths with environment variables:

**Bad:**
```json
{
  "command": ".claude/hooks/format.sh"  // Relative path fails
}
```

**Good:**
```json
{
  "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format.sh"
}
```

---

#### Matcher Not Matching

**Error:** Hook doesn't fire for expected events

**Debug:** Check matcher regex syntax

**Examples:**

```json
{
  "matcher": "Write|Edit",  // Correct: OR pattern
  "hooks": [...]
}
```

```json
{
  "matcher": "mcp__.*",  // Correct: Any MCP tool
  "hooks": [...]
}
```

**Avoid:**
```json
{
  "matcher": "*.py",  // Wrong: This is glob, not regex
  "hooks": [...]
}
```

---

### Validation Script

Create `.claude/hooks/validate-config.sh`:

```bash
#!/bin/bash
# Validate hook configuration

for file in ~/.claude/settings.json .claude/settings.json .claude/settings.local.json; do
  if [[ -f "$file" ]]; then
    echo "Validating $file..."
    if ! jq empty "$file" 2>/dev/null; then
      echo "ERROR: Invalid JSON in $file" >&2
      exit 1
    fi
    echo "✓ Valid"
  fi
done

echo "All hook configurations valid!"
```

---

## See Also

- [Events Reference](./events-reference.md) - Complete event catalog
- [Scripting & Execution](./scripting.md) - Writing hook scripts
- [Examples](./examples.md) - Ready-to-use patterns
- [Architecture](./architecture.md) - Execution model

---

**Sources:**
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)
- [eesel.ai Developer Guide](https://www.eesel.ai/blog/settings-json-claude-code)
- [GitHub Examples](https://github.com/disler/claude-code-hooks-mastery)

**Document Version:** 1.0
**Research Date:** 2026-02-21
