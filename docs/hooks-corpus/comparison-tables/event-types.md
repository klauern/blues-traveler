# Event Types Across Systems

**Cross-system event comparison for AI coding assistant hooks**

This table maps equivalent events across all five systems, enabling developers to understand how different AI coding assistants handle similar automation triggers.

---

## Quick Reference

| Event Category | Claude Code | Cursor IDE | Copilot | Codex | Gemini |
|----------------|-------------|------------|---------|-------|--------|
| **Pre-execution** | PreToolUse | ? | ? | ? | ? |
| **Post-execution** | PostToolUse | ? | ? | ? | ? |
| **User interaction** | UserPromptSubmit | ? | ? | ? | ? |
| **Session lifecycle** | SessionStart/End | ? | ? | ? | ? |
| **Notifications** | Notification | ? | ? | ? | ? |

**Legend:**
- ✓ = Fully supported
- ~ = Partial/limited support
- ✗ = Not supported
- ? = Unknown/needs research

---

## Pre-Execution Events

Events that fire **before** an action is executed, allowing validation, blocking, or modification.

### General Pre-Execution Hook

**Purpose:** Intercept and validate any action before execution

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `PreToolUse` | Tool-based matching | ✓ Yes (exit code 1) | Tool name, parameters, user prompt |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Security validation (block dangerous commands)
- Policy enforcement (require code reviews)
- Custom workflow triggers (notify on specific actions)

**Cross-System Notes:**
- Claude Code uses a single `PreToolUse` event with tool-specific filtering via `matcher` field
- Other systems may provide granular per-action events
- See individual system docs for filtering/matching syntax

---

### Pre-Shell Execution

**Purpose:** Intercept shell/bash commands before execution

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `PreToolUse` (matcher: "Bash") | Via matcher field | ✓ Yes | Command string, working dir |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Block destructive commands (rm -rf, sudo)
- Enforce safe command patterns
- Log command execution

---

### Pre-File Edit

**Purpose:** Intercept file modifications before they occur

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `PreToolUse` (matcher: "Edit\|Write") | Via matcher field | ✓ Yes | File path, old/new content |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Auto-format code before saving
- Validate file permissions
- Enforce file naming conventions

---

## Post-Execution Events

Events that fire **after** an action completes, allowing logging, follow-up actions, or rollback.

### General Post-Execution Hook

**Purpose:** React to completed actions

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `PostToolUse` | Tool-based matching | ✗ No (already executed) | Tool name, result, exit code |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Log successful operations
- Trigger follow-up actions (tests after code changes)
- Collect metrics

---

### Post-File Edit

**Purpose:** React to completed file modifications

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `PostToolUse` (matcher: "Edit\|Write") | Via matcher field | ✗ No | File path, changes made |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Run tests after code changes
- Update documentation
- Trigger builds

---

## User Interaction Events

Events related to user input and prompts.

### User Prompt Submission

**Purpose:** Intercept user prompts before they're processed

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `UserPromptSubmit` | N/A | ✓ Yes (exit code 1) | User prompt text |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Inject context into prompts
- Block inappropriate prompts
- Log user interactions

---

## Session Lifecycle Events

Events related to session start, end, and state changes.

### Session Start

**Purpose:** Triggered when AI assistant session begins

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `SessionStart` | N/A | ✗ No | Session ID, project info |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Initialize project context
- Load custom settings
- Display welcome messages

---

### Session End

**Purpose:** Triggered when AI assistant session terminates

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `SessionEnd` | N/A | ✗ No | Session ID, duration |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Cleanup temporary files
- Save session state
- Generate reports

---

## Subagent/Task Events

Events related to spawned subagents or background tasks.

### Subagent Stop

**Purpose:** Triggered when a subagent/background task completes

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `SubagentStop` | N/A | ✗ No | Subagent ID, result |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Collect results from parallel tasks
- Log task completion
- Trigger dependent workflows

---

## Context Management Events

Events related to conversation context and memory.

### Pre-Context Compaction

**Purpose:** Triggered before conversation history is compressed

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `PreCompact` | N/A | ✗ No | Messages to be compacted |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Save important context before compaction
- Extract key decisions
- Archive conversation history

---

## Notification Events

Events triggered by system notifications or alerts.

### General Notification

**Purpose:** React to system notifications

| System | Event Name | Filtering | Cancellable | Context Available |
|--------|------------|-----------|-------------|-------------------|
| **Claude Code** | `Notification` | Type-based filtering | ✗ No | Notification type, data |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Example Use Cases:**
- Route notifications to external systems
- Log important events
- Trigger alerts

---

## System-Specific Events

Events unique to specific systems that don't map across platforms.

### Claude Code Only

| Event Name | Purpose | Notes |
|------------|---------|-------|
| `Stop` | Session stopped by user | Cleanup hook |

### Cursor Only

| Event Name | Purpose | Notes |
|------------|---------|-------|
| ? | ? | ? |

### Copilot Only

| Event Name | Purpose | Notes |
|------------|---------|-------|
| ? | ? | ? |

### Codex Only

| Event Name | Purpose | Notes |
|------------|---------|-------|
| ? | ? | ? |

### Gemini Only

| Event Name | Purpose | Notes |
|------------|---------|-------|
| ? | ? | ? |

---

## Event Ordering & Lifecycle

### Typical Event Sequence

Example event sequence for a file edit operation:

```
Claude Code:          Cursor:               Copilot:
1. SessionStart       ?                     ?
2. UserPromptSubmit   ?                     ?
3. PreToolUse (Edit)  ?                     ?
4. [Edit executes]    [Edit executes]       [Edit executes]
5. PostToolUse (Edit) ?                     ?
6. SessionEnd         ?                     ?
```

### Event Precedence

When multiple hooks match an event:

| System | Execution Order | Stop on Failure |
|--------|----------------|-----------------|
| **Claude Code** | Sequential (config order) | Yes (exit code 1) |
| **Cursor** | ? | ? |
| **Copilot** | ? | ? |
| **Codex** | ? | ? |
| **Gemini** | ? | ? |

---

## Sources & References

### Claude Code
- [Claude Code Hooks Documentation](https://docs.anthropic.com/claude-code/hooks)
- [blues-traveler: Custom Hooks Guide](../claude/events-reference.md)

### Cursor
- [Cursor Documentation](https://cursor.com/docs)
- [blues-traveler: Cursor Compatibility](../cursor/events-reference.md)

### Copilot
- [GitHub Copilot Documentation](https://docs.github.com/en/copilot)
- Research notes: `research-notes/copilot/`

### Codex
- [OpenAI API Documentation](https://platform.openai.com/docs/)
- Research notes: `research-notes/codex/`

### Gemini
- [Google Gemini Documentation](https://ai.google.dev/)
- Research notes: `research-notes/gemini/`

---

## See Also

- **[Configuration Comparison](./configuration.md)** - Config format comparison
- **[Capabilities Matrix](./capabilities.md)** - Feature support comparison
- **[Migration Matrix](./migration-matrix.md)** - System-to-system migration
- **[Corpus Home](../README.md)** - Main documentation index

---

**Last Updated:** 2026-02-21
**Status:** In Progress - Claude Code complete, others in research phase
