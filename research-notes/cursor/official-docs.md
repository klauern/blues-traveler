# Cursor IDE Official Hooks Documentation

**Research Date:** 2026-02-21
**Official Documentation URL:** https://cursor.com/docs/agent/hooks

## Overview

Cursor introduced a Hooks system in version 1.7 (released in October 2025) that allows developers to intercept and modify agent behavior at defined lifecycle events. Hooks are external scripts that run at defined stages of the agent loop.

## Key Features

- **Lifecycle Interception**: Hooks execute at specific events emitted by Cursor
- **JSON-based Communication**: Each hook receives structured JSON over stdin and returns JSON on stdout
- **Standalone Execution**: Hooks run as separate subprocess with isolated execution
- **Multiple Hook Support**: Each event can trigger multiple hooks in sequence
- **Multi-level Configuration**: Hooks can be defined at project, user, or system level

## Supported Hook Events

Cursor supports 6 lifecycle hook events:

### 1. `beforeShellExecution`
- **When**: Before executing shell commands initiated by the AI agent
- **Purpose**: Validate, block, or approve shell commands
- **Response**: Supports `permission` field (allow/deny/ask)
- **Common Use Cases**:
  - Block dangerous commands (e.g., `rm -rf /`, `git push --force`)
  - Restrict package installations
  - Enforce use of specific tools (e.g., prefer `gh` over `git`)

### 2. `beforeMCPExecution`
- **When**: Before MCP (Model Context Protocol) tool execution
- **Purpose**: Control which MCP tools can be executed
- **Response**: Supports `permission` field (allow/deny/ask)
- **Common Use Cases**:
  - Restrict MCP servers to approved/managed instances
  - Enforce governance policies
  - Validate tool parameters

### 3. `beforeReadFile`
- **When**: Before the AI agent reads a file
- **Purpose**: Control file access and redact sensitive content
- **Response**: Supports `permission` field (allow/deny)
- **Common Use Cases**:
  - Block reading of sensitive files (.env, .ssh/config, credentials)
  - Redact secrets/tokens before sending to model
  - Protect critical configuration files

### 4. `afterFileEdit`
- **When**: After the AI agent modifies a file
- **Purpose**: Post-processing, validation, or notification
- **Response**: No response expected (informational hook)
- **Common Use Cases**:
  - Auto-format edited files
  - Run linters
  - Audit/log file changes
  - Update documentation

### 5. `beforeSubmitPrompt`
- **When**: When user submits a prompt, before sending to the model
- **Purpose**: Gate or filter prompts
- **Response**: Boolean `continue` field only
- **Common Use Cases**:
  - Require specific keywords for approval
  - Prompt validation
  - Session control
- **Note**: Context injection is NOT supported

### 6. `stop`
- **When**: When the AI agent finishes execution/task
- **Purpose**: Cleanup, notification, or summary
- **Response**: Optional `followup_message` to agent
- **Common Use Cases**:
  - Display notifications
  - Generate session summaries
  - Report detections from other hooks
  - Cleanup temporary files

## Configuration Format

### File Locations

Hooks are configured via `hooks.json` file at one or more levels:

1. **Project-specific**: `.cursor/hooks.json` in project root
   - Applies only to that project
   - Checked into version control for team consistency
   - Highest precedence

2. **User-specific**: `~/.cursor/hooks.json` or similar user config directory
   - Applies to all projects for that user
   - Personal automation and preferences
   - Medium precedence

3. **Global/system-level**: System-wide configuration location
   - Applies to all users on the system
   - Lowest precedence

**Precedence Rule**: Project-level > User-level > Global-level. All applicable hooks from all levels are executed.

### JSON Schema

Official schema: https://unpkg.com/cursor-hooks/schema/hooks.schema.json

```json
{
  "$schema": "https://unpkg.com/cursor-hooks/schema/hooks.schema.json",
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "path/to/script.sh" }
    ],
    "beforeMCPExecution": [
      { "command": "path/to/mcp-check.sh" }
    ],
    "beforeReadFile": [
      { "command": "path/to/secret-scanner.sh" }
    ],
    "afterFileEdit": [
      { "command": "path/to/formatter.sh" }
    ],
    "beforeSubmitPrompt": [
      { "command": "path/to/prompt-gate.sh" }
    ],
    "stop": [
      { "command": "path/to/notify.sh" }
    ]
  }
}
```

### Schema Properties

- **`version`** (required): Must be `1` (only supported version)
- **`hooks`** (required): Object mapping event names to hook arrays
- **`command`** (required per hook): String path to executable
  - Supports absolute paths
  - Supports paths relative to `hooks.json` location
  - Supports shell snippets
  - Minimum length: 1

### Path Resolution

Hook command paths are **relative to the `.cursor/hooks.json` file location**.

Example:
```
.cursor/
  hooks.json          # Contains: "./hooks/script.sh"
  hooks/
    script.sh         # Actual hook script
```

## Hook Input Format (stdin)

All hooks receive JSON on stdin with event-specific payloads. Common fields:

### Common Fields (all events)

```json
{
  "conversation_id": "uuid-for-chat-session",
  "generation_id": "uuid-for-prompt-within-conversation",
  "hook_event_name": "beforeShellExecution",
  "workspace_roots": ["/path/to/project"]
}
```

**Field Descriptions:**

- **`conversation_id`**: Unique ID for each new chat/conversation
- **`generation_id`**: Unique ID for each prompt within a conversation
- **`hook_event_name`**: The event that triggered this hook
- **`workspace_roots`**: Array of workspace root paths
  - Usually contains one path
  - Can contain multiple paths in VS Code multi-root workspaces

### Event-Specific Fields

#### `beforeShellExecution`
```json
{
  "command": "npm install lodash@4.17.21",
  "conversation_id": "...",
  "generation_id": "...",
  "hook_event_name": "beforeShellExecution",
  "workspace_roots": ["/path/to/project"]
}
```

#### `beforeMCPExecution`
```json
{
  "tool_name": "mcp-server-name",
  "arguments": { "arg1": "value1" },
  "conversation_id": "...",
  "generation_id": "...",
  "hook_event_name": "beforeMCPExecution",
  "workspace_roots": ["/path/to/project"]
}
```

#### `beforeReadFile`
```json
{
  "file_path": "/path/to/file.txt",
  "content": "file contents here",
  "attachments": [],
  "conversation_id": "...",
  "generation_id": "...",
  "hook_event_name": "beforeReadFile",
  "workspace_roots": ["/path/to/project"]
}
```

#### `afterFileEdit`
```json
{
  "file_path": "/path/to/edited.js",
  "edits": [
    {
      "old_string": "original content",
      "new_string": "modified content"
    }
  ],
  "conversation_id": "...",
  "generation_id": "...",
  "hook_event_name": "afterFileEdit",
  "workspace_roots": ["/path/to/project"]
}
```

#### `beforeSubmitPrompt`
```json
{
  "prompt": "User's prompt text",
  "attachments": [],
  "conversation_id": "...",
  "generation_id": "...",
  "hook_event_name": "beforeSubmitPrompt",
  "workspace_roots": ["/path/to/project"]
}
```

#### `stop`
```json
{
  "status": "completion status",
  "conversation_id": "...",
  "generation_id": "...",
  "hook_event_name": "stop",
  "workspace_roots": ["/path/to/project"]
}
```

## Hook Output Format (stdout)

Hooks must output valid JSON to stdout. The schema varies by event type.

### Permission-based Events (beforeShellExecution, beforeMCPExecution, beforeReadFile)

```json
{
  "permission": "allow",
  "userMessage": "Optional: Shown to the user in UI",
  "agentMessage": "Optional: Technical details for the AI agent"
}
```

**Fields:**

- **`permission`** (required): One of:
  - `"allow"` - Permit the action
  - `"deny"` - Block the action
  - `"ask"` - Prompt user for manual approval (beforeShellExecution and beforeMCPExecution only)

- **`userMessage`** (optional): User-friendly message displayed in Cursor UI
  - Keep concise and actionable
  - Explain what happened and why
  - Example: "Command blocked for security reasons"

- **`agentMessage`** (optional): Technical details for the AI agent
  - Include error codes, patterns matched, diagnostic info
  - Helps agent understand and potentially retry
  - Example: "Blocked dangerous pattern: sudo rm -rf / (filesystem root deletion)"

### Flow Control Events (beforeSubmitPrompt)

```json
{
  "continue": true
}
```

**Fields:**

- **`continue`** (required): Boolean
  - `true` - Allow prompt submission
  - `false` - Block prompt submission

### Informational Events (afterFileEdit)

No JSON output expected. Hook exit code indicates success/failure.
Use `console.error()` or stderr for logging (stdout logging interferes with JSON parsing).

### Completion Events (stop)

```json
{
  "followup_message": "Optional message to send back to agent"
}
```

**Fields:**

- **`followup_message`** (optional): String message sent to the agent
  - Can include session summaries
  - Can report findings from earlier hooks
  - Example: "Session audit: 3 files edited, 2 shell commands blocked"

## Execution Model

### Process Model

- Each hook runs as a **standalone subprocess**
- Hook receives input via **stdin** (JSON)
- Hook returns output via **stdout** (JSON)
- Cursor waits for hook completion before proceeding
- **Synchronous execution**: Blocks agent action until hook completes

### Exit Codes

Hook exit codes are processed as follows:

1. **Exit 0 + valid JSON**: Process JSON response (allow/deny based on fields)
2. **Exit 0 + no JSON**: Implicit allow (silent success)
3. **Exit non-zero + no JSON**: Block with error alert
4. **Exit any + invalid JSON**: Block with "hook broken" error message
5. **Exit any + partial JSON**: Process available fields (missing fields use defaults)

### Error Handling

- Invalid JSON → Block execution with error message
- Partial JSON → Process with defaults for missing fields
- Missing required fields → Use sensible defaults
- Hook crashes → Fail-safe behavior (typically block)

### Timeout Behavior

- No explicit timeout documented in official spec
- Best practice: Keep hooks under 2 seconds for typical operations
- Maximum recommended: 10 seconds for complex operations
- Long-running hooks will delay agent actions

### Security Considerations

**Hook Execution Context:**

- Hooks run in the user's security context (not sandboxed)
- Hooks have full filesystem and network access
- Hooks can execute arbitrary code
- **Risk**: Malicious hooks can compromise system

**Best Practices:**

- Only run trusted hook scripts
- Review hook scripts before enabling
- Use minimum necessary permissions
- Avoid `sudo` in hooks unless absolutely necessary
- Never hardcode secrets in hook scripts
- Use environment variables or OS keychains for credentials
- Sanitize and validate all hook inputs

**Sandboxing Notes:**

- Hooks themselves are NOT sandboxed
- Cursor's agent sandbox (for shell commands) uses:
  - macOS: Seatbelt (sandbox-exec)
  - Linux: Landlock and seccomp
  - Windows: Linux sandbox via WSL2
- Hook execution happens OUTSIDE the agent sandbox
- MCP initialization commands ingested during startup run outside sandbox

## IDE Integration

### JSON Schema Validation

Enable autocomplete and validation in Cursor settings:

**Option A: Add `$schema` to hooks.json**
```json
{
  "$schema": "https://unpkg.com/cursor-hooks@latest/schema/hooks.schema.json",
  "version": 1,
  "hooks": { ... }
}
```

**Option B: Configure globally in Cursor settings**

File: `~/Library/Application Support/Cursor/User/settings.json` (macOS)

```json
{
  "json.validate.enable": true,
  "json.format.enable": true,
  "json.schemaDownload.enable": true,
  "json.schemas": [
    {
      "fileMatch": [".cursor/hooks.json"],
      "url": "https://unpkg.com/cursor-hooks/schema/hooks.schema.json"
    }
  ]
}
```

Benefits:
- Autocomplete for event names
- Inline documentation on hover
- Validation warnings for invalid config
- Schema-aware error messages

## Limitations and Gaps

### Current Limitations

1. **No timeout configuration**: Cannot specify per-hook timeouts in hooks.json
2. **No async hooks**: All hooks are synchronous and block execution
3. **No hook chaining control**: Cannot control order of multiple hooks for same event
4. **Limited beforeSubmitPrompt**: No context injection support
5. **afterFileEdit is write-only**: Cannot modify files or send messages to user/agent
6. **No environment variable passing**: Cannot pass custom env vars to hooks via config
7. **No retry logic**: Hooks that fail do not retry automatically
8. **No rate limiting**: No built-in throttling for hooks
9. **No hook metrics**: No built-in performance monitoring or statistics

### Documentation Gaps

1. **Timeout behavior**: No documented timeout limits or behavior
2. **Error recovery**: Limited documentation on error handling strategies
3. **Performance guidance**: No official performance benchmarks or guidelines
4. **Multi-hook execution order**: Unclear how multiple hooks for same event are ordered
5. **Environment variables**: No documented env vars available to hooks (community-discovered only)
6. **Sandbox interaction**: Limited docs on how hooks interact with agent sandbox

### Beta Status

As of February 2026:
- Hooks are still a **beta feature**
- API may change in future versions
- Community feedback indicates occasional instability
- Documentation is minimal (community-driven docs fill gaps)

## Community Resources

- **TypeScript SDK**: https://github.com/johnlindquist/cursor-hooks (npm: `cursor-hooks`)
- **Python SDK**: https://github.com/DevonFulcher/py-cursor-hooks (PyPI: `py-cursor-hooks`)
- **Example Repository**: https://github.com/hamzafer/cursor-hooks (Shell scripts)
- **Security Examples**: https://github.com/endorlabs/cursor-hook-examples (Malware detection)
- **1Password Integration**: https://github.com/1Password/cursor-hooks (Validation)

## Official vs Community Knowledge

**Official (Documented by Cursor):**
- 6 hook events
- JSON schema format
- Basic input/output formats
- Configuration file locations

**Community (Reverse-engineered/Discovered):**
- Exact input payloads for each event
- Environment variable handling
- Error handling behavior
- Performance characteristics
- Best practices for security
- Multi-language SDK implementations
- Real-world use case patterns

## References

- Official Documentation: https://cursor.com/docs/agent/hooks
- JSON Schema: https://unpkg.com/cursor-hooks/schema/hooks.schema.json
- InfoQ Article (Oct 2025): https://www.infoq.com/news/2025/10/cursor-hooks/
- GitButler Deep Dive: https://blog.gitbutler.com/cursor-hooks-deep-dive
- Skywork AI Guide: https://skywork.ai/blog/how-to-cursor-1-7-hooks-guide/
