# Claude Code Hooks & Automation Reference

**Version:** 1.0 (as of February 2026)
**Last Updated:** 2026-02-21
**Status:** Core reference complete; advanced sections are placeholder pages

---

## Table of Contents

1. [Overview & Philosophy](#1-overview--philosophy)
2. [Architecture & Internals](./architecture.md)
3. [Event Types & Triggers](./events-reference.md)
4. [Configuration & Setup](./configuration.md)
5. [Environment & Context](./environment-context.md)
6. [Scripting & Execution](./scripting.md)
7. [Security & Safety](./security.md)
8. [Common Patterns & Recipes](./examples.md)
9. [Integration & Extensibility](./integration.md)
10. [Troubleshooting & Debugging](./troubleshooting.md)
11. [API Reference](./api-reference.md)
12. [Migration Guide](./migration.md)

---

## 1. Overview & Philosophy

### What are Hooks in Claude Code?

Hooks are **user-defined callbacks** that execute automatically at specific points in Claude Code's lifecycle. They enable you to:

- **Automate workflows** (format code, run tests, send notifications)
- **Enforce policies** (block dangerous commands, validate inputs)
- **Customize behavior** (inject context, modify tool parameters)
- **Integrate external systems** (CI/CD, Slack, Jira, analytics)

Hooks run as shell commands, LLM prompts, or autonomous subagents, receiving event data via JSON stdin and controlling execution through exit codes and JSON output.

**Source:** [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)

### Design Philosophy

Claude Code's hook system embodies several core principles:

**Key Principles:**

1. **Event-driven automation** - Hooks fire at precise lifecycle moments
2. **Flexible control** - Three hook types (command, prompt, agent) support different complexity levels
3. **Decision power** - Hooks can block actions, modify inputs, inject context
4. **JSON-based communication** - Standard stdin/stdout protocol with well-defined schemas
5. **Matcher filtering** - Regular expressions filter when hooks fire
6. **Security-conscious** - Hooks run with user permissions; explicit warnings about safety
7. **Composable** - Multiple hooks from different sources merge and execute
8. **Transparent** - Debug mode reveals execution details; `/hooks` UI manages configuration

**Sources:**
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [eesel.ai Developer Guide](https://www.eesel.ai/blog/settings-json-claude-code)

### When to Use Hooks

**Use hooks when:**
- ✅ Enforcing code style (auto-format with Prettier, Black, gofmt)
- ✅ Running tests automatically after changes
- ✅ Blocking dangerous operations (rm -rf, sudo, curl | sh)
- ✅ Injecting project context at session start
- ✅ Implementing custom approval workflows

**Don't use hooks when:**
- ❌ You need to modify Claude's internal behavior (hooks are external)
- ❌ The operation requires modifying Claude's prompt (use CLAUDE.md specs instead)
- ❌ You want to change Claude's personality (use custom agents/skills)
- ❌ The task should be a skill (reusable command) instead of a hook

**Sources:**
- [DataCamp Tutorial](https://www.datacamp.com/tutorial/claude-code-hooks)
- [Official Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

### Version & Compatibility

| Version | Released | Status | Notes |
|---------|----------|--------|-------|
| 1.0+ | 2025-2026 | Stable | 17 events, 3 hook types |
| 0.x | 2024-2025 | Legacy | Earlier versions (9 events) |

**Platform Support:** macOS, Linux, Windows (WSL), Web (code.claude.com), GitHub Codespaces

**See:** [Configuration & Setup](./configuration.md) for full details

---

## Quick Start

### Basic Hook Example

**Auto-format Python files after edits:**

```bash
#!/bin/bash
# .claude/hooks/format-python.sh

INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_response.filePath // empty')

if [[ "$FILE" == *.py ]]; then
  black "$FILE"
  isort "$FILE"
fi
exit 0
```

**Configure in `.claude/settings.json`:**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format-python.sh"
          }
        ]
      }
    ]
  }
}
```

---

## The 17 Hook Events

| Event | Trigger | Cancellable | Use Case |
|-------|---------|-------------|----------|
| SessionStart | Session begins | ❌ | Load project context |
| UserPromptSubmit | User submits prompt | ✅ | Validate content |
| PreToolUse | Before tool execution | ✅ | Block dangerous commands |
| PermissionRequest | Permission dialog | ✅ | Auto-approve safe ops |
| PostToolUse | After tool success | ❌ | Auto-format code |
| PostToolUseFailure | After tool failure | ❌ | Log errors |
| Stop | Claude finishes | ✅ | Verify work complete |
| SubagentStart | Subagent spawned | ❌ | Inject instructions |
| SubagentStop | Subagent finishes | ✅ | Verify completion |
| TeammateIdle | Teammate going idle | ✅ | Enforce quality gates |
| TaskCompleted | Task marked complete | ✅ | Verify criteria |
| Notification | Notification sent | ❌ | Forward to Slack |
| ConfigChange | Config file changes | ✅* | Audit changes |
| WorktreeCreate | Worktree created | ✅ | Custom VCS |
| WorktreeRemove | Worktree removed | ❌ | VCS cleanup |
| PreCompact | Before compaction | ❌ | Log events |
| SessionEnd | Session terminates | ❌ | Cleanup files |

\* Cannot block policy_settings

**See:** [Event Types & Triggers](./events-reference.md)

---

## Documentation Structure

This reference provides comprehensive coverage across 12 focused sections:

1. **[Architecture & Internals](./architecture.md)** - Execution model, lifecycle, performance
2. **[Event Types & Triggers](./events-reference.md)** - Complete catalog with schemas
3. **[Configuration & Setup](./configuration.md)** - Settings files, merging logic
4. **[Environment & Context](./environment-context.md)** - Variables, JSON fields
5. **[Scripting & Execution](./scripting.md)** - Languages, exit codes, error handling
6. **[Security & Safety](./security.md)** - Permissions, blocking operations
7. **[Common Patterns & Recipes](./examples.md)** - Copy-paste ready examples
8. **[Integration & Extensibility](./integration.md)** - Build tools, CI/CD
9. **[Troubleshooting & Debugging](./troubleshooting.md)** - Common issues, debug mode
10. **[API Reference](./api-reference.md)** - Complete schema reference
11. **[Migration Guide](./migration.md)** - Moving to/from other systems
12. **[Working Examples](./examples/)** - Ready-to-use scripts

---

## Additional Resources

### Official Documentation
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

### Community Resources
- [awesome-claude-code](https://github.com/hesreallyhim/awesome-claude-code)
- [claude-code-hooks-mastery](https://github.com/disler/claude-code-hooks-mastery)
- [everything-claude-code](https://github.com/affaan-m/everything-claude-code)

### Comparison Tables
- [Event Types Comparison](../comparison-tables/event-types.md)
- [Configuration Format Comparison](../comparison-tables/configuration.md)
- [Capabilities Matrix](../comparison-tables/capabilities.md)

---

**Document Version:** 1.0
**Research Sources:** 30+ (official docs, blogs, GitHub repos)
**Coverage:** 17/17 events documented (100%)
