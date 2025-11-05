# Cursor Compatibility Guide

This guide shows how to write hooks that work seamlessly in both Cursor IDE and Claude Code (via blues-traveler).

## Overview

Blues-traveler supports the [Cursor hooks specification](https://cursor.com/docs/agent/hooks), enabling you to write hooks once and run them in both environments. This compatibility layer includes:

- **Event name aliases**: Use Cursor event names that auto-translate to Claude Code events
- **JSON response format**: Standard response structure for both systems ✅ **Implemented**
- **Dual-message support**: Separate messages for users vs AI agents ✅ **Implemented**
- **Permission model**: allow/deny/ask permissions ⚠️ **Partial** (ask mode falls back to allow with messages)

## Event Name Mapping

Blues-traveler accepts both canonical Claude Code event names and Cursor aliases:

| Cursor Event Name | Claude Code Event | Description | Matcher Pattern |
|-------------------|-------------------|-------------|-----------------|
| `beforeShellExecution` | `PreToolUse` | Before executing Bash commands | `TOOL_NAME=Bash` |
| `afterFileEdit` | `PostToolUse` | After Edit/Write tool completes | `TOOL_NAME=Edit,Write` |
| `beforeFileRead` | `PreToolUse` | Before reading files | `TOOL_NAME=Read` |
| `onUserPromptSubmit` | `UserPromptSubmit` | When user submits a prompt | N/A |
| `onSessionEnd` | `SessionEnd` | When session ends | N/A |

**Usage**: You can use either name when installing hooks:

```bash
# Using Cursor event name (auto-translated)
blues-traveler hooks install security --event beforeShellExecution

# Using Claude Code canonical name (works identically)
blues-traveler hooks install security --event PreToolUse
```

## JSON Response Format

Hooks can output JSON to control execution and provide messages. Both systems support the same format:

### Response Schema

```json
{
  "permission": "allow|deny|ask",
  "userMessage": "Message shown to the user",
  "agentMessage": "Technical details sent to AI agent",
  "continue": true|false
}
```

### Field Descriptions

- **`permission`** (string, optional): Control execution flow
  - `"allow"` - Continue execution (default if omitted) ✅
  - `"deny"` - Block execution ✅
  - `"ask"` - Prompt user for manual approval
    - ✅ **Works in Cursor IDE** - Native support for user prompting
    - ⚠️ **Falls back in Claude Code** - Approves with contextual messages (Claude Code doesn't support ask mode)
    - Logged for audit visibility

- **`userMessage`** (string, optional): User-friendly message displayed in the UI
  - Keep concise and actionable
  - Explain what happened and why
  - Example: "This command was blocked for security reasons"

- **`agentMessage`** (string, optional): Technical details for the AI agent
  - Include error codes, patterns matched, diagnostic info
  - Helps agent understand and potentially retry
  - Example: "blocked dangerous command pattern: sudo (check_type: regex)"

- **`continue`** (boolean, optional): Override permission-based flow
  - `false` - Block execution regardless of permission field
  - `true` or omitted - Use permission field logic

### Response Processing Rules

Blues-traveler follows these rules when processing hook output:

1. **Non-zero exit + no JSON** → Block with alert + error message
2. **Partial JSON** → Proceed with available fields (missing fields use defaults)
3. **Invalid JSON** → Block with "hook broken" message
4. **Exit 0 + no JSON** → Allow (silent success)

## Writing Compatible Hook Scripts

### Basic Example: Security Check

This script works in both Cursor and blues-traveler:

```bash
#!/bin/bash
# security-check.sh - Blocks dangerous sudo commands

COMMAND="$1"

if echo "$COMMAND" | grep -q "sudo rm -rf /"; then
  cat <<EOF
{
  "permission": "deny",
  "userMessage": "Command blocked for security",
  "agentMessage": "Blocked dangerous pattern: sudo rm -rf / (filesystem root deletion)"
}
EOF
  exit 0
fi

# Allow by default
echo '{"permission": "allow"}'
exit 0
```

### Advanced Example: Dual Messages

```bash
#!/bin/bash
# format-checker.sh - Validates code formatting

FILE="$1"

if ! prettier --check "$FILE" 2>/dev/null; then
  cat <<EOF
{
  "permission": "deny",
  "userMessage": "Code formatting failed for $(basename "$FILE")",
  "agentMessage": "Prettier check failed for $FILE. Run: prettier --write \"$FILE\""
}
EOF
  exit 0
fi

echo '{"permission": "allow"}'
exit 0
```

### Best Practices

1. **Always output valid JSON** - Invalid JSON blocks execution with error
2. **Exit 0 for controlled blocks** - Use `permission: deny` + exit 0, not exit 1
3. **Separate user/agent messages** - Users get friendly text, agents get technical details
4. **Handle missing tools gracefully** - Check tool availability before use
5. **Use partial JSON strategically** - Omit optional fields for cleaner output

## Configuration Examples

### Cursor-style hooks.json

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

### Blues-traveler settings.json

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

### Custom Hooks YAML (blues-traveler only)

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

## Environment Variables

Both systems provide hook scripts with contextual environment variables:

| Variable | Available In | Description |
|----------|-------------|-------------|
| `HOOK_EVENT_NAME` | Both | Event that triggered the hook |
| `TOOL_NAME` | Both | Name of tool being used (Bash, Edit, Write, etc.) |
| `TOOL_INPUT` | Both | Input/arguments to the tool |
| `FILES_CHANGED` | Both | List of files modified |
| `USER_PROMPT` | Both (UserPromptSubmit only) | User's prompt text |
| `CWD` | Both | Current working directory |

## Migration Guide

### From Cursor to Blues-traveler

1. **Install blues-traveler**:
   ```bash
   go install github.com/klauern/blues-traveler@latest
   ```

2. **Copy hook scripts** - Your existing Cursor hook scripts work as-is

3. **Convert configuration** - Translate event names and settings format:
   ```bash
   # Cursor
   "beforeShellExecution" → PreToolUse (matcher: "*" or "Bash")
   "afterFileEdit" → PostToolUse (matcher: "Edit,Write")
   ```

4. **Install hooks**:
   ```bash
   blues-traveler hooks custom install security
   ```

5. **Test**:
   ```bash
   blues-traveler hooks run security
   ```

### From Blues-traveler to Cursor

1. Extract hook scripts from your blues-traveler configuration
2. Update event names to Cursor format (use mapping table above)
3. Create Cursor `hooks.json` with your scripts
4. Test in Cursor IDE

### Keeping Both in Sync

For teams using both systems:

1. **Store hook scripts in version control** (e.g., `.claude/hooks/scripts/`)
2. **Use Cursor event names** - Blues-traveler auto-translates them
3. **Standardize on JSON responses** - Works in both systems
4. **Document dependencies** - List required tools (prettier, ruff, etc.)
5. **Test in both environments** - Ensure consistent behavior

## Testing Hook Compatibility

### Test Script Template

```bash
#!/bin/bash
# test-hook.sh - Verify hook works in both systems

set -e

echo "Testing hook output format..."

# Test 1: Valid JSON allow
output=$(.claude/hooks/scripts/my-hook.sh "safe command")
if ! echo "$output" | jq -e '.permission == "allow"' > /dev/null; then
  echo "FAIL: Allow case"
  exit 1
fi

# Test 2: Valid JSON deny
output=$(.claude/hooks/scripts/my-hook.sh "dangerous command")
if ! echo "$output" | jq -e '.permission == "deny"' > /dev/null; then
  echo "FAIL: Deny case"
  exit 1
fi

# Test 3: Both messages present
if ! echo "$output" | jq -e '.userMessage and .agentMessage' > /dev/null; then
  echo "FAIL: Missing messages"
  exit 1
fi

echo "✅ All compatibility tests passed"
```

### Manual Testing Checklist

- [ ] Hook outputs valid JSON (test with `jq`)
- [ ] Both `userMessage` and `agentMessage` are meaningful
- [ ] Exit codes follow rules (0 for controlled blocks)
- [ ] Works with expected environment variables
- [ ] Handles missing tools gracefully
- [ ] Performance acceptable (< 2s typical, < 10s max)
- [ ] No hardcoded paths or system-specific assumptions

## Troubleshooting

### "Hook returned invalid JSON"

**Problem**: Hook output isn't valid JSON or doesn't match schema

**Solution**:
- Test JSON with `jq`: `./my-hook.sh | jq .`
- Check for extra output (echo statements, debug info)
- Ensure heredoc is properly formatted
- Validate against schema: `{"permission": "allow|deny", ...}`

### "Permission field not recognized"

**Problem**: Typo in permission value or wrong type

**Solution**:
- Must be lowercase string: `"allow"`, `"deny"`, or `"ask"`
- Check quotes: `"permission": "allow"` not `permission: allow`

### Hook works in Cursor but not blues-traveler

**Problem**: Event name or environment variable mismatch

**Solution**:
- Use Cursor event names (blues-traveler translates them)
- Check environment variables are documented in both systems
- Test with: `blues-traveler hooks run <hook-type> --log`

### Hook blocks but no message shown

**Problem**: Missing `userMessage` field

**Solution**:
- Always include `userMessage` when `permission: deny`
- Provide actionable guidance for the user
- Example: `"userMessage": "Code formatting required. Run: npm run format"`

## Implementation Status

Blues-traveler's Cursor compatibility features have been implemented in phases:

### ✅ Fully Implemented

1. **JSON Response Parsing** (Issue #59)
   - Parses Cursor-style JSON responses from hook scripts
   - Supports both permission-based and flow-control formats
   - Handles partial JSON gracefully
   - Comprehensive error handling and validation

2. **Dual-Message Responses** (Issue #55)
   - Separate `userMessage` and `agentMessage` fields
   - Wrapper functions: `BlockWithMessages()`, `ApproveWithMessages()`, `PostBlockWithMessages()`, `AllowWithMessages()`
   - Single-parameter support (same message to both audiences)
   - Backward compatible with existing code

3. **Response Format Support**
   - `permission: "allow"` - ✅ Working
   - `permission: "deny"` - ✅ Working
   - `continue: true/false` - ✅ Working
   - `userMessage` / `agentMessage` routing - ✅ Working

### ✅ Platform-Specific Features

1. **Ask Permission Mode** (Issue #54)
   - **Status**: Works in Cursor IDE, graceful fallback in Claude Code
   - **Cursor IDE**: Native "ask" support - prompts user for approval
   - **Claude Code**: Falls back to `Approve()` with contextual messages (Claude Code doesn't support ask mode natively)
   - **Logging**: All "ask" events are logged for audit visibility
   - **Function Available**: `AskWithMessages()` provides consistent API across both platforms

### 📋 Usage Recommendations

**For Cursor IDE users:**
- Use `permission: "ask"` when you want user confirmation for non-critical operations
- User will see a prompt and can approve/deny
- Great for operations that might be disruptive but not dangerous

**For Claude Code users:**
- `permission: "ask"` will allow execution with messages (logged for audit)
- Use `permission: "deny"` for security-critical blocks
- Use `permission: "allow"` for informational messages

**Cross-platform hooks:**
- Safe to use `permission: "ask"` - works optimally in Cursor, degrades gracefully in Claude Code
- Check logs to see when "ask" events occur

## Additional Resources

- [Cursor Hooks Documentation](https://cursor.com/docs/agent/hooks)
- [Blues-traveler README](../README.md)
- [Blues-traveler Custom Hooks Guide](../README.md#custom-hooks)
- [cchooks Library](https://github.com/brads3290/cchooks)

## Contributing

Found an incompatibility or have a compatibility tip? Please open an issue or PR at:
https://github.com/klauern/blues-traveler/issues
