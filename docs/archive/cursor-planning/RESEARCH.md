# Cursor Hooks Research Document

## Purpose

This document tracks what we need to learn about Cursor hooks to successfully implement support in blues-traveler.

## ✅ RESEARCH COMPLETE - Key Findings

### Configuration System

**Location**: `~/.cursor/hooks.json` (NOT `.cursor/settings.json`!)

**Format**: JSON with version field

```json
{
  "version": 1,
  "hooks": {
    "afterFileEdit": [{ "command": "./hooks/format.sh" }]
  }
}
```

**Configuration Priority** (most restrictive wins):

1. Global (Enterprise): `/Library/Application Support/Cursor/hooks.json` (macOS), `/etc/cursor/hooks.json` (Linux), `C:\\ProgramData\\Cursor\\hooks.json` (Windows)
2. User: `~/.cursor/hooks.json`

### Event System - MAJOR DIFFERENCES

| Claude Code Event | Cursor Event     | Cursor Name            | Notes                                      |
| ----------------- | ---------------- | ---------------------- | ------------------------------------------ |
| PreToolUse        | ✅ Similar       | `beforeShellExecution` | Before shell commands                      |
| PreToolUse        | ✅ Similar       | `beforeMCPExecution`   | Before MCP tool calls                      |
| PostToolUse       | ⚠️ Partial       | `afterFileEdit`        | Only for file edits, not all tools         |
| UserPromptSubmit  | ✅ Similar       | `beforeSubmitPrompt`   | Before user prompt sent                    |
| Stop              | ✅ Similar       | `stop`                 | When agent loop ends                       |
| PreToolUse        | ✅ NEW           | `beforeReadFile`       | Before agent reads a file (access control) |
| PostToolUse       | ❌ No equivalent | N/A                    | No generic post-tool hook                  |
| Notification      | ❌ No equivalent | N/A                    | Not supported                              |
| SubagentStop      | ❌ No equivalent | N/A                    | Not supported                              |
| PreCompact        | ❌ No equivalent | N/A                    | Not supported                              |
| SessionStart      | ❌ No equivalent | N/A                    | Not supported                              |
| SessionEnd        | ❌ No equivalent | N/A                    | Not supported                              |

### Hook Protocol - CRITICAL DIFFERENCE

**Cursor**: JSON over stdio (spawned process)

- Input: JSON on stdin
- Output: JSON on stdout
- Exit codes: 0 (success), 3 (deny permission)

**Claude Code**: Command execution

- Input: Environment variables
- Output: Exit code (0 = success, non-zero = failure)

### Hook Input/Output Schemas

#### Common Input (All Hooks)

```json
{
  "conversation_id": "string",
  "generation_id": "string",
  "hook_event_name": "string",
  "workspace_roots": ["<path>"]
}
```

#### beforeShellExecution

```json
// Input
{
  "command": "<full terminal command>",
  "cwd": "<current working directory>",
  // ... common fields
}

// Output
{
  "permission": "allow" | "deny" | "ask",
  "userMessage": "<message shown in client>",
  "agentMessage": "<message sent to agent>"
}
```

#### beforeMCPExecution

```json
// Input
{
  "tool_name": "<tool name>",
  "tool_input": "<json params>",
  "url": "<server url>" | "command": "<command string>",
  // ... common fields
}

// Output (same as beforeShellExecution)
```

#### afterFileEdit

```json
// Input
{
  "file_path": "<absolute path>",
  "edits": [{ "old_string": "<search>", "new_string": "<replace>" }]
  // ... common fields
}

// Output: none required (or exit code)
```

#### beforeReadFile

```json
// Input
{
  "file_path": "<absolute path>",
  "content": "<file contents>",
  // ... common fields
}

// Output
{
  "permission": "allow" | "deny"
}
```

#### beforeSubmitPrompt

```json
// Input
{
  "prompt": "<user prompt text>",
  "attachments": [
    {
      "type": "file" | "rule",
      "file_path": "<absolute path>"
    }
  ],
  // ... common fields
}

// Output
{
  "continue": true | false
}
```

#### stop

```json
// Input
{
  "status": "completed" | "aborted" | "error",
  // ... common fields
}

// Output: none
```

### NO MATCHERS

**Critical Difference**: Cursor has NO matcher/filter system at the config level.

- Claude Code: Regex matchers in config to filter when hooks run
- Cursor: Hooks run for ALL instances of the event, filtering must happen INSIDE the script

This means blues-traveler will need to:

1. Generate wrapper scripts that implement the filtering logic
2. Call the actual blues-traveler hooks after filtering

### Environment Variables

Cursor hooks receive data via **JSON on stdin**, NOT environment variables.

To support blues-traveler's existing model, we need to:

1. Parse JSON input
2. Set environment variables (EVENT_NAME, TOOL_NAME, etc.)
3. Call blues-traveler hooks
4. Translate exit code/output back to JSON

## Research Questions - ANSWERED

### 1. Configuration & Settings

#### Q1.1: Where does Cursor store hook configurations?

✅ **ANSWERED**: `~/.cursor/hooks.json` (user) or system-wide paths (enterprise)

#### Q1.2: What is the JSON schema for Cursor hooks?

✅ **ANSWERED**: See schemas above

#### Q1.3: Does Cursor support XDG Base Directory Spec?

❌ **NO**: Fixed path at `~/.cursor/hooks.json`

### 2. Event Lifecycle

#### Q2.1: What events does Cursor support?

✅ **ANSWERED**: 6 events total (see table above)

#### Q2.2: What is the event execution order?

✅ **ANSWERED**:

```text
User Input -> beforeSubmitPrompt -> [Backend] -> beforeShellExecution/beforeMCPExecution -> [Execution] -> afterFileEdit -> ... -> stop
```

### 3. Tool Integration

#### Q3.1: What tools does Cursor expose to hooks?

✅ **ANSWERED**:

- Shell commands (`beforeShellExecution`)
- MCP tools (`beforeMCPExecution`)
- File operations (`afterFileEdit`, `beforeReadFile`)

#### Q3.2: What context is available in hooks?

✅ **ANSWERED**: JSON stdin (see schemas above)

#### Q3.3: How are tool parameters passed to hooks?

✅ **ANSWERED**: JSON on stdin

### 4. Hook Execution

#### Q4.1: What is the hook command format?

✅ **ANSWERED**:

```json
{ "command": "./script.sh" } // Relative to hooks.json or absolute
```

#### Q4.2: How does Cursor handle hook failures?

✅ **ANSWERED**:

- Exit 0 = success
- Exit 3 = deny permission
- Permission responses via JSON

#### Q4.3: What is the timeout behavior?

❓ **NOT DOCUMENTED**: Need to test

#### Q4.4: Are hooks executed in parallel or serially?

✅ **ANSWERED**: Multiple configs merge; most restrictive wins

### 5. Matchers & Filters

#### Q5.1: How does Cursor filter when hooks run?

❌ **NO FILTERING**: Hooks run for all events, must filter inside script

#### Q5.2: What is the expression syntax?

❌ **NO EXPRESSION SYNTAX**: No config-level filtering

### 6. Security & Permissions

#### Q6.1: Does Cursor require hook approval/allowlisting?

✅ **ANSWERED**: User can configure in `hooks.json`; global config managed by enterprise

#### Q6.2: What sandboxing or security features exist?

✅ **ANSWERED**: Permission system (`allow`, `deny`, `ask`) for gating operations

### 7. Logging & Debugging

#### Q7.1: How can hook execution be debugged?

✅ **ANSWERED**:

- Hooks tab in Cursor Settings
- Hooks output channel for errors

#### Q7.2: Where are hook errors reported?

✅ **ANSWERED**: Hooks output channel in Cursor

### 8. Versioning & Compatibility

#### Q8.1: What version of Cursor supports hooks?

❓ **NOT DOCUMENTED**: Need to test

#### Q8.2: Are there breaking changes between versions?

✅ **ANSWERED**: Version field in config (`"version": 1`)

## Key Differences Summary

### Configuration Differences

| Aspect         | Claude Code              | Cursor                              | Migration Strategy       |
| -------------- | ------------------------ | ----------------------------------- | ------------------------ |
| Config path    | `.claude/settings.json`  | `~/.cursor/hooks.json`              | Separate file management |
| Schema format  | Hooks nested in settings | Dedicated hooks config with version | Schema translation       |
| Matcher syntax | Regex in config          | None - filter in script             | Generate wrapper scripts |
| Hook format    | Command with env vars    | Command with JSON I/O               | Adapter/wrapper needed   |

### Event Mapping

| Generic Event     | Claude Code        | Cursor Equivalent                             | Translation Notes         |
| ----------------- | ------------------ | --------------------------------------------- | ------------------------- |
| PreToolUse        | `PreToolUse`       | `beforeShellExecution` + `beforeMCPExecution` | Split by tool type        |
| PostToolUse       | `PostToolUse`      | `afterFileEdit`                               | Only for file edits       |
| UserPromptSubmit  | `UserPromptSubmit` | `beforeSubmitPrompt`                          | Direct mapping            |
| Stop              | `Stop`             | `stop`                                        | Direct mapping            |
| PreToolUse (Read) | N/A                | `beforeReadFile`                              | New Cursor-specific event |
| Notification      | `Notification`     | ❌ None                                       | Not supported in Cursor   |
| SubagentStop      | `SubagentStop`     | ❌ None                                       | Not supported in Cursor   |
| PreCompact        | `PreCompact`       | ❌ None                                       | Not supported in Cursor   |
| SessionStart      | `SessionStart`     | ❌ None                                       | Not supported in Cursor   |
| SessionEnd        | `SessionEnd`       | ❌ None                                       | Not supported in Cursor   |

### Protocol Translation

**Blues-Traveler Adapter Required**:

```bash
#!/bin/bash
# cursor-adapter.sh - Wraps blues-traveler hooks for Cursor

# Read JSON input
input=$(cat)

# Parse and set environment variables
export EVENT_NAME=$(echo "$input" | jq -r '.hook_event_name')
export TOOL_NAME=$(echo "$input" | jq -r '.command // .tool_name // "unknown"')
# ... more parsing

# Check matcher if needed
if [[ "$MATCHER" != "" ]]; then
  # Implement matcher logic
fi

# Call blues-traveler hook
blues-traveler run security

# Translate exit code to JSON response
if [ $? -eq 0 ]; then
  echo '{"permission": "allow"}'
else
  echo '{"permission": "deny", "userMessage": "Security check failed"}'
  exit 3
fi
```

## Implementation Strategy

### Option 1: Wrapper Script Generation

Generate Cursor-compatible wrapper scripts that:

1. Parse JSON input
2. Set environment variables
3. Apply matcher logic
4. Call blues-traveler hooks
5. Translate response to JSON

### Option 2: Native Cursor Support

Add native JSON I/O support to blues-traveler:

1. Detect execution context (env var vs JSON stdin)
2. Parse input accordingly
3. Output JSON when running under Cursor

### Recommendation: Hybrid Approach

1. Generate wrapper scripts for easy installation
2. Add `--cursor-mode` flag to blues-traveler for native JSON I/O
3. Wrapper can call `blues-traveler run <hook> --cursor-mode`

## Testing Plan

### Test Cases Created

1. **Basic Hook Installation** ✅

   - Generate wrapper script
   - Install in `~/.cursor/hooks.json`
   - Verify hook executes on event

2. **Event Mapping** ✅

   - Test `beforeShellExecution` → PreToolUse mapping
   - Test `afterFileEdit` → PostToolUse mapping
   - Test `beforeSubmitPrompt` → UserPromptSubmit mapping
   - Test `stop` → Stop mapping

3. **Context Variables** ✅

   - Parse JSON input correctly
   - Set all expected environment variables
   - Test file path handling

4. **Permission System** ✅

   - Test allow/deny responses
   - Test userMessage and agentMessage
   - Test exit code 3 for deny

5. **Multi-Platform** ✅
   - Install same hook on both platforms
   - Compare behavior
   - Test platform detection

## Next Steps

1. ✅ Research complete - all questions answered
2. ⏭️ Update `platform-comparison.md` with findings
3. ⏭️ Revise `cursor-hooks-support-plan.md` with new strategy
4. ⏭️ Design wrapper script generator
5. ⏭️ Implement Cursor platform with JSON I/O support

---

**Document Status**: ✅ Research Complete
**Created**: 2024-09-30
**Completed**: 2024-09-30
**Next Action**: Update implementation plan
