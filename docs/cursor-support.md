# Cursor Hooks Support

> **Status**: ✅ Phase 1 Complete - Core infrastructure implemented
> **Version**: v0.1.0-alpha (In Development)

## Overview

blues-traveler now supports Cursor IDE hooks in addition to Claude Code. Cursor hooks use a different protocol (JSON over stdin/stdout) compared to Claude Code (environment variables), so blues-traveler implements a **hybrid adapter pattern** to bridge the two systems.

## Architecture

```
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

Generate wrapper scripts for Cursor (automatic with `hooks install --platform cursor` - coming in Phase 2):

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

**Current Limitation**: Hooks in Cursor mode currently allow all operations without executing hook logic. This is because the cchooks Runner tries to read from stdin, which has already been consumed in Cursor mode. Full hook execution (converting Cursor events to cchooks events and calling handlers directly) will be implemented in Phase 3.

### 📋 Phase 3: Polish (Future)

- [ ] Full hook execution in Cursor mode (convert events, call handlers)
- [ ] Custom hooks support for Cursor
- [ ] Cross-platform sync command
- [ ] Migration guide
- [ ] Full documentation
- [ ] Beta release

**Next Priority**: Implement executeCursorHook to convert Cursor JSON to cchooks events and call hook handlers directly, enabling security checks and other hook logic to actually run.

## Testing

All core components have unit tests:

```bash
# Run Cursor platform tests
go test ./internal/platform/cursor/... -v

# Run all tests
go test ./...
```

## Example: Using in Cursor (Manual Setup)

Until Phase 2 is complete, you can manually set up Cursor hooks:

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

## Next Steps

See [tmp/cursor-hooks/PLAN.md](../tmp/cursor-hooks/PLAN.md) for the complete implementation plan.

---

**Last Updated**: 2024-09-30
**Status**: Phase 1 Complete
