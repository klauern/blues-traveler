# Configuration Formats Across Systems

**Cross-system configuration comparison for AI coding assistant hooks**

This table compares how different AI coding assistants configure their hook systems, including file formats, locations, schemas, and precedence rules.

---

## Quick Reference

| System | Format | Project Config | Global Config | Schema Validation |
|--------|--------|----------------|---------------|-------------------|
| **Claude Code** | JSON | `.claude/settings.json` | `~/.claude/settings.json` | ✓ Yes |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

---

## File Locations

### Project-Level Configuration

Configuration that applies to a specific project/repository.

| System | Path | Format | Auto-Created | Version Control |
|--------|------|--------|--------------|-----------------|
| **Claude Code** | `.claude/settings.json` | JSON | No (manual) | ✓ Recommended |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Best Practices:**
- Claude Code: Commit `.claude/settings.json` to share hooks across team
- Cursor: ?
- Copilot: ?
- Codex: ?
- Gemini: ?

---

### Global/User Configuration

Configuration that applies to all projects for a user.

| System | Path | Format | Auto-Created | Scope |
|--------|------|--------|--------------|-------|
| **Claude Code** | `~/.claude/settings.json` | JSON | Yes (on first run) | All projects |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

**Use Cases:**
- Claude Code: Personal hooks that apply across all projects
- Cursor: ?
- Copilot: ?
- Codex: ?
- Gemini: ?

---

## Configuration Formats

### Claude Code (JSON)

**File:** `.claude/settings.json`

**Minimal Example:**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "command": "/usr/local/bin/validate-command"
      }]
    }]
  }
}
```

**Complete Schema:**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "ToolName",
      "hooks": [{
        "command": "/path/to/script",
        "env": {
          "KEY": "value"
        }
      }]
    }],
    "PostToolUse": [/* same structure */],
    "UserPromptSubmit": [/* same structure */],
    "SessionStart": [/* same structure */],
    "SessionEnd": [/* same structure */],
    "SubagentStop": [/* same structure */],
    "PreCompact": [/* same structure */],
    "Notification": [/* same structure */],
    "Stop": [/* same structure */]
  }
}
```

**[→ Full Claude Code Configuration Reference](../claude/configuration.md)**

---

### Cursor (Format TBD)

**File:** ?

**Minimal Example:**
```
{Unknown - needs research}
```

**[→ Full Cursor Configuration Reference](../cursor/configuration.md)**

---

### Copilot (Format TBD)

**File:** ?

**Minimal Example:**
```
{Unknown - needs research}
```

**[→ Full Copilot Configuration Reference](../copilot/configuration.md)**

---

### Codex (Format TBD)

**File:** ?

**Minimal Example:**
```
{Unknown - needs research}
```

**[→ Full Codex Configuration Reference](../codex/configuration.md)**

---

### Gemini (Format TBD)

**File:** ?

**Minimal Example:**
```
{Unknown - needs research}
```

**[→ Full Gemini Configuration Reference](../gemini/configuration.md)**

---

## Configuration Precedence

How different configuration sources are merged when both project and global configs exist.

### Claude Code

**Precedence Order:**
1. Project config (`.claude/settings.json`)
2. Global config (`~/.claude/settings.json`)

**Merge Strategy:**
- Hooks are **concatenated** (both project and global hooks execute)
- Project hooks execute **before** global hooks
- Arrays are merged, not replaced
- Object fields in project config **override** global config

**Example:**
```json
// Global: ~/.claude/settings.json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{ "command": "/usr/local/bin/global-check" }]
    }]
  }
}

// Project: .claude/settings.json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{ "command": "/path/to/project-check" }]
    }]
  }
}

// Effective configuration:
// 1. /path/to/project-check (project)
// 2. /usr/local/bin/global-check (global)
```

---

### Cursor

**Precedence Order:** ?

**Merge Strategy:** ?

---

### Copilot

**Precedence Order:** ?

**Merge Strategy:** ?

---

### Codex

**Precedence Order:** ?

**Merge Strategy:** ?

---

### Gemini

**Precedence Order:** ?

**Merge Strategy:** ?

---

## Schema & Validation

### Schema Validation Support

| System | Schema Format | Validation Tool | IDE Support | Error Messages |
|--------|---------------|-----------------|-------------|----------------|
| **Claude Code** | JSON Schema | Built-in validator | ✓ VS Code | Detailed |
| **Cursor** | ? | ? | ? | ? |
| **Copilot** | ? | ? | ? | ? |
| **Codex** | ? | ? | ? | ? |
| **Gemini** | ? | ? | ? | ? |

### Claude Code Schema

**JSON Schema Location:** Built into Claude Code binary

**Validation:**
```bash
# Claude Code validates on startup
claude --validate-config

# Also validates on config reload
```

**Common Validation Errors:**
- Invalid JSON syntax
- Unknown event types
- Invalid matcher patterns
- Missing required fields

---

## Environment Variable Configuration

Some systems allow environment variables to configure or override hook settings.

### Claude Code

**Environment Variables:**

| Variable | Purpose | Example |
|----------|---------|---------|
| `CLAUDE_SETTINGS_PATH` | Override settings file location | `/custom/path/settings.json` |
| `CLAUDE_DISABLE_HOOKS` | Disable all hooks | `1` or `true` |
| `CLAUDE_HOOK_TIMEOUT` | Override default hook timeout (ms) | `5000` |

**Usage:**
```bash
export CLAUDE_DISABLE_HOOKS=1
claude  # Runs with hooks disabled
```

---

### Cursor

**Environment Variables:** ?

---

### Copilot

**Environment Variables:** ?

---

### Codex

**Environment Variables:** ?

---

### Gemini

**Environment Variables:** ?

---

## Configuration Examples by Use Case

### Security Validation Hook

Block dangerous shell commands.

**Claude Code:**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "command": "/usr/local/bin/security-check"
      }]
    }]
  }
}
```

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

---

### Code Formatting Hook

Auto-format code before file edits.

**Claude Code:**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Edit|Write",
      "hooks": [{
        "command": "/usr/local/bin/format-code"
      }]
    }]
  }
}
```

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

---

### Test Execution Hook

Run tests after code changes.

**Claude Code:**
```json
{
  "hooks": {
    "PostToolUse": [{
      "matcher": "Edit|Write",
      "hooks": [{
        "command": "/usr/local/bin/run-tests"
      }]
    }]
  }
}
```

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

---

## Migration Examples

### From Claude Code to Cursor

**Before (Claude Code):**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "command": "/usr/local/bin/validate"
      }]
    }]
  }
}
```

**After (Cursor):**
```
{Unknown - needs research}
```

---

### From Cursor to Copilot

**Before (Cursor):**
```
{Unknown - needs research}
```

**After (Copilot):**
```
{Unknown - needs research}
```

---

## Configuration Testing

### Testing Configuration Validity

**Claude Code:**
```bash
# Validate configuration file
claude --validate-config

# Dry-run with verbose output
claude --dry-run --verbose
```

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

---

### Testing Hook Execution

**Claude Code:**
```bash
# Test a specific hook
HOOK_EVENT=PreToolUse TOOL_NAME=Bash /usr/local/bin/your-hook

# Debug hook execution
claude --debug-hooks
```

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

---

## Configuration Management Best Practices

### Version Control

| System | Recommended Files to Commit | Files to Ignore |
|--------|----------------------------|-----------------|
| **Claude Code** | `.claude/settings.json` | `.claude/.local_*` |
| **Cursor** | ? | ? |
| **Copilot** | ? | ? |
| **Codex** | ? | ? |
| **Gemini** | ? | ? |

### Secret Management

**Claude Code:**
- ✗ **Don't** put secrets directly in `settings.json`
- ✓ **Do** use environment variables in hook scripts
- ✓ **Do** use external secret managers (Bitwarden, 1Password)

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

### Team Sharing

**Claude Code:**
- Commit `.claude/settings.json` for team-wide hooks
- Document hook script installation in README
- Use relative paths when possible
- Provide hook script installation script

**Cursor:** ?

**Copilot:** ?

**Codex:** ?

**Gemini:** ?

---

## Configuration Troubleshooting

### Common Issues

#### Invalid JSON Syntax

**System:** Claude Code

**Error:**
```
Error: Invalid JSON in settings file
```

**Solution:**
```bash
# Validate JSON
jq . .claude/settings.json

# Fix formatting
jq . .claude/settings.json > .claude/settings.json.tmp
mv .claude/settings.json.tmp .claude/settings.json
```

---

#### Hook Not Executing

**System:** Claude Code

**Debugging:**
```bash
# Check configuration
claude --validate-config

# Enable hook debugging
claude --debug-hooks

# Check hook script permissions
ls -la /path/to/hook-script
chmod +x /path/to/hook-script
```

---

## Sources & References

### Claude Code
- [Claude Code Hooks Documentation](https://docs.anthropic.com/claude-code/hooks)
- [blues-traveler: Configuration Guide](../claude/configuration.md)
- [Example configurations](../claude/examples/)

### Cursor
- [Cursor Documentation](https://cursor.com/docs)
- Research notes: `research-notes/cursor/`

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

- **[Event Types Comparison](./event-types.md)** - Cross-system event comparison
- **[Capabilities Matrix](./capabilities.md)** - Feature support comparison
- **[Migration Matrix](./migration-matrix.md)** - System-to-system migration
- **[Corpus Home](../README.md)** - Main documentation index

---

**Last Updated:** 2026-02-21
**Status:** In Progress - Claude Code complete, others in research phase
