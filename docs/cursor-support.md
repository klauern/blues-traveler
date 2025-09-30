# Cursor Hooks Support

> **Status**: ✅ Phase 3 Complete - Full hook execution with JSON transformation
> **Version**: v0.3.0-alpha (Ready for Testing)

## Overview

blues-traveler now supports Cursor IDE hooks in addition to Claude Code. Cursor hooks use a different protocol (JSON over stdin/stdout) compared to Claude Code (environment variables), so blues-traveler implements a **hybrid adapter pattern** to bridge the two systems.

## Architecture

```text
Cursor Agent
    ↓ JSON stdin
Wrapper Script (auto-generated)
    ↓ Parse JSON, set env vars
blues-traveler run <hook> --cursor-mode
    ↓ Execute hook, return JSON
Wrapper Script
    ↓ JSON stdout
Cursor Agent
```

### Key Differences from Claude Code

| Aspect          | Claude Code             | Cursor                   |
| --------------- | ----------------------- | ------------------------ |
| **Protocol**    | Environment variables   | JSON stdin/stdout        |
| **Config File** | `.claude/settings.json` | `~/.cursor/hooks.json`   |
| **Matchers**    | Regex in config         | None (filter in scripts) |
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
```

### Wrapper Scripts

Wrapper scripts are automatically generated when you install hooks for Cursor using `hooks install --platform cursor`:

```bash
# Example generated wrapper for security hook
#!/bin/bash
input=$(cat)
export EVENT_NAME=$(echo "$input" | jq -r '.hook_event_name')
export TOOL_ARGS=$(echo "$input" | jq -r '.command')

if blues-traveler run security --cursor-mode <<< "$input"; then
  exit 0
else
  exit 3  # Deny permission
fi
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
- Special: `TOOL_NAME` is set to `"shell"`

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
- [x] Wrapper script generator with templates
- [x] Cursor config handler (`~/.cursor/hooks.json`)
- [x] Platform detection logic
- [x] Event name mapper (Claude ↔ Cursor)
- [x] Unit tests for core functionality

### ✅ Phase 2: CLI Integration (Complete)

- [x] `--platform cursor` flag on install command
- [x] Automatic wrapper script generation and installation
- [x] Platform auto-detection in commands
- [x] `blues-traveler platform detect` command
- [x] Basic integration (hooks allow all operations)

### ✅ Phase 3: Full Hook Execution (Complete)

- [x] Full hook execution in Cursor mode (convert events, call handlers)
- [x] JSON transformation from Cursor format to Claude Code format
- [x] Hook logic now executes properly (security checks, formatting, etc.)
- [ ] Custom hooks support for Cursor (future enhancement)
- [ ] Cross-platform sync command (future enhancement)
- [ ] Migration guide (future enhancement)

**Status**: All core functionality is complete! Hooks now properly execute in Cursor mode by transforming Cursor JSON events to Claude Code format before passing them to hook handlers.

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

# Auto-detect platform (works if you're in a directory with .cursor/)
blues-traveler hooks install security --event PreToolUse
```

This automatically:

- Generates wrapper script
- Makes it executable
- Updates `~/.cursor/hooks.json`
- Maps events correctly

### Manual Setup (Advanced)

If you need manual control, you can set up Cursor hooks manually:

1. **Create wrapper script** (`~/.cursor/hooks/blues-traveler-security.sh`):

   ```bash
   #!/bin/bash
   input=$(cat)
   blues-traveler run security --cursor-mode <<< "$input"
   ```

2. **Make executable**:

   ```bash
   chmod +x ~/.cursor/hooks/blues-traveler-security.sh
   ```

3. **Update Cursor config** (`~/.cursor/hooks.json`):

   ```json
   {
     "version": 1,
     "hooks": {
       "beforeShellExecution": [
         { "command": "~/.cursor/hooks/blues-traveler-security.sh" }
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

## Next Steps

- Custom hooks support for Cursor (use YAML/JSON config to define Cursor-specific hooks)
- Cross-platform sync command to sync hooks between Claude Code and Cursor
- Migration guide for users switching platforms

---

**Last Updated**: 2024-09-30
**Status**: ✅ Phase 3 Complete - Full Cursor hooks support with proper hook execution
