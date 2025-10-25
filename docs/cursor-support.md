# Cursor Hooks Support

> **Status**: ✅ Phase 3 Complete - Full hook execution with JSON transformation
> **Version**: v0.3.0-alpha (Ready for Testing)

## Prerequisites

- **blues-traveler**: Must be installed and available in your PATH
- **Cursor IDE**: Version that supports hooks (usually the latest version)
- **jq**: Required for JSON processing in Cursor hook scripts (install via `brew install jq` or your package manager)

## Overview

blues-traveler now supports Cursor IDE hooks in addition to Claude Code. Cursor hooks use a different protocol (JSON over stdin/stdout) compared to Claude Code (environment variables), so blues-traveler implements **direct execution** with the `--cursor-mode` flag.

## Architecture

```text
Cursor Agent
    ↓ JSON stdin
blues-traveler run <hook> --cursor-mode [--matcher <pattern>]
    ↓ Transform JSON to Claude Code format
    ↓ Execute hook, return JSON
Cursor Agent
```

### Key Differences from Claude Code

| Aspect          | Claude Code             | Cursor                   |
| --------------- | ----------------------- | ------------------------ |
| **Protocol**    | Environment variables   | JSON stdin/stdout        |
| **Config File** | `.claude/settings.json` | `~/.cursor/hooks.json`   |
| **Matchers**    | Regex in config         | Regex in command args    |
| **Events**      | 9 events                | 6 events                 |

## Cursor Events Supported

| Cursor Event           | Generic Event    | Description                    |
| ---------------------- | ---------------- | ------------------------------ |
| `beforeShellExecution` | PreToolUse       | Before shell command execution |
| `beforeMCPExecution`   | PreToolUse       | Before MCP tool calls          |
| `afterFileEdit`        | PostToolUse      | After file edit operations     |
| `beforeReadFile`       | PreToolUse       | Before agent reads a file      |
| `beforeSubmitPrompt`   | UserPromptSubmit | Before user prompt is sent     |
| `stop`                 | Stop             | When agent loop ends           |

**Not supported in Cursor**: `Notification`, `SubagentStop`, `PreCompact`, `SessionStart`, `SessionEnd`

## Using Cursor Mode

### Manual Execution

Run any blues-traveler hook in Cursor mode:

```bash
# Read JSON from stdin, output JSON to stdout
echo '{"hook_event_name": "beforeShellExecution", "command": "rm -rf /"}' | \
  blues-traveler run security --cursor-mode

# With matcher filter
echo '{"hook_event_name": "beforeShellExecution", "command": "rm -rf /"}' | \
  blues-traveler run security --cursor-mode --matcher "rm.*"
```

### Direct Registration

Hooks are registered directly in `~/.cursor/hooks.json` using the `hooks install --platform cursor` command:

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      {
        "command": "/path/to/blues-traveler hooks run security --cursor-mode --matcher \".*\""
      }
    ]
  }
}
```

## Environment Variable Mapping

Cursor JSON fields are mapped to environment variables that blues-traveler hooks expect:

### Common Fields

- `conversation_id` → `CONVERSATION_ID`
- `generation_id` → `GENERATION_ID`
- `hook_event_name` → `EVENT_NAME`
- `workspace_roots` → `WORKSPACE_ROOTS` (colon-separated)

### Event-Specific Fields

#### beforeShellExecution

- `command` → `TOOL_ARGS`
- `cwd` → `CWD`
- Special: `TOOL_NAME` is set to `"Bash"`

#### beforeMCPExecution

- `tool_name` → `TOOL_NAME`
- `tool_input` → `TOOL_ARGS`
- `url` → `MCP_URL`

#### afterFileEdit

- `file_path` → `FILE_PATH`
- `edits` → `FILE_EDITS` (JSON array)

#### beforeReadFile

- `file_path` → `FILE_PATH`
- `content` → `FILE_CONTENT`

#### beforeSubmitPrompt

- `prompt` → `USER_PROMPT`
- `attachments` → `PROMPT_ATTACHMENTS` (JSON array)

#### stop

- `status` → `STOP_STATUS`

## Implementation Status

### ✅ Phase 1: Core Infrastructure (Complete)

- [x] Platform abstraction layer
- [x] Cursor JSON I/O types and schemas
- [x] `--cursor-mode` flag for JSON protocol
- [x] Cursor config handler (`~/.cursor/hooks.json`)
- [x] Platform detection logic
- [x] Event name mapper (Claude ↔ Cursor)
- [x] Unit tests for core functionality

### ✅ Phase 2: CLI Integration (Complete)

- [x] `--platform cursor` flag on install command
- [x] Direct command registration in hooks.json
- [x] Platform auto-detection in commands
- [x] `blues-traveler platform detect` command
- [x] Matcher support via command-line flags

### ✅ Phase 3: Full Hook Execution (Complete)

- [x] Full hook execution in Cursor mode (convert events, call handlers)
- [x] JSON transformation from Cursor format to Claude Code format
- [x] Hook logic executes properly (security checks, formatting, etc.)
- [x] Matcher filtering with regex support
- [x] Direct command registration (no wrapper scripts needed)
- [ ] Custom hooks support for Cursor (future enhancement)
- [ ] Cross-platform sync command (future enhancement)

**Status**: ✅ **Phase 3 Complete!** All core functionality implemented. Hooks execute in Cursor mode by transforming Cursor JSON events to Claude Code format. Installation simplified with direct command registration instead of wrapper scripts.

## Testing

All core components have unit tests:

```bash
# Run Cursor platform tests
go test ./internal/platform/cursor/... -v

# Run all tests
go test ./...
```

## Example: Using in Cursor

### Automated Installation (Recommended)

Use the built-in install command:

```bash
# Install security hook for Cursor
blues-traveler hooks install security --platform cursor --event PreToolUse

# With custom matcher
blues-traveler hooks install security --platform cursor --event PreToolUse --matcher "rm.*|sudo.*"

# Auto-detect platform (works if you're in a directory with .cursor/)
blues-traveler hooks install security --event PreToolUse
```

This automatically:

- Builds the appropriate command with `--cursor-mode` and optional `--matcher`
- Updates `~/.cursor/hooks.json`
- Maps events correctly

### Manual Setup (Advanced)

If you need manual control, you can set up Cursor hooks manually:

1. **Update Cursor config** (`~/.cursor/hooks.json`):

   ```json
   {
     "version": 1,
     "hooks": {
       "beforeShellExecution": [
         {
           "command": "/path/to/blues-traveler hooks run security --cursor-mode --matcher \"rm.*|sudo.*\""
         }
       ]
     }
   }
   ```

## Platform Detection

blues-traveler automatically detects the platform in this order:

1. `BLUES_TRAVELER_PLATFORM` environment variable
2. `.cursor` directory in current working directory
3. `.claude` directory in current working directory
4. `~/.cursor/hooks.json` exists in home directory
5. Default to Claude Code (backward compatibility)

## Architecture Implementation

The key to making Cursor hooks work is **JSON transformation**:

1. Cursor sends events like `beforeShellExecution` with Cursor-specific fields
2. `transformCursorToClaudeFormat()` converts to Claude Code format (`PreToolUse`)
3. Transformed JSON is fed to the hook via stdin pipe
4. Hook executes normally using cchooks Runner
5. Response is converted back to Cursor format

This allows all existing hooks (security, format, vet, etc.) to work in Cursor without modification!

## Matcher Support

Matchers filter which events trigger hooks, similar to Claude Code's config-based matchers:

- For `beforeShellExecution`: matches against the command string
- For `beforeMCPExecution`: matches against the tool name
- For `afterFileEdit` and `beforeReadFile`: matches against the file path
- Use standard regex syntax (e.g., `"rm.*|sudo.*"`, `".*\\.go$"`)
- Special value `"*"` or empty string matches all events

## Next Steps

- Custom hooks support for Cursor (use YAML/JSON config to define Cursor-specific hooks)
- Cross-platform sync command to sync hooks between Claude Code and Cursor
- Migration guide for users switching platforms

---

**Last Updated**: 2025-10-01
**Status**: ✅ Phase 3 Complete - Full Cursor support with direct command registration
