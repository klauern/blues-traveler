# Migration Matrix

**Bidirectional migration guides for hook systems across AI coding assistants**

This matrix provides comprehensive migration guides for moving hook configurations between different AI coding assistant systems, including complete before/after examples and translation patterns.

---

## Quick Navigation

**Migration Paths:**
- [Claude Code ↔ Cursor](#claude-code--cursor)
- [Claude Code ↔ Copilot](#claude-code--copilot)
- [Claude Code ↔ Codex](#claude-code--codex)
- [Claude Code ↔ Gemini](#claude-code--gemini)
- [Cursor ↔ Copilot](#cursor--copilot)
- [Cursor ↔ Codex](#cursor--codex)
- [Cursor ↔ Gemini](#cursor--gemini)
- [Copilot ↔ Codex](#copilot--codex)
- [Copilot ↔ Gemini](#copilot--gemini)
- [Codex ↔ Gemini](#codex--gemini)

---

## Migration Path Status

| From → To | Status | Difficulty | Completeness |
|-----------|--------|------------|--------------|
| **Claude Code → Cursor** | 🚧 Partial | Medium | 30% |
| **Cursor → Claude Code** | 🚧 Partial | Medium | 30% |
| **Claude Code → Copilot** | 📝 Research | Unknown | 0% |
| **Copilot → Claude Code** | 📝 Research | Unknown | 0% |
| **Claude Code → Codex** | 📝 Research | Unknown | 0% |
| **Codex → Claude Code** | 📝 Research | Unknown | 0% |
| **Claude Code → Gemini** | 📝 Research | Unknown | 0% |
| **Gemini → Claude Code** | 📝 Research | Unknown | 0% |
| **Cursor → Copilot** | 📝 Research | Unknown | 0% |
| **Copilot → Cursor** | 📝 Research | Unknown | 0% |
| **Cursor → Codex** | 📝 Research | Unknown | 0% |
| **Codex → Cursor** | 📝 Research | Unknown | 0% |
| **Cursor → Gemini** | 📝 Research | Unknown | 0% |
| **Gemini → Cursor** | 📝 Research | Unknown | 0% |
| **Copilot → Codex** | 📝 Research | Unknown | 0% |
| **Codex → Copilot** | 📝 Research | Unknown | 0% |
| **Copilot → Gemini** | 📝 Research | Unknown | 0% |
| **Gemini → Copilot** | 📝 Research | Unknown | 0% |
| **Codex → Gemini** | 📝 Research | Unknown | 0% |
| **Gemini → Codex** | 📝 Research | Unknown | 0% |

---

## Claude Code ↔ Cursor

### Claude Code → Cursor

**Migration Difficulty:** Medium
**Status:** 🚧 Partial documentation

#### Event Mapping

| Claude Code Event | Cursor Equivalent | Notes |
|-------------------|-------------------|-------|
| `PreToolUse` | ? | Unknown - needs research |
| `PostToolUse` | ? | Unknown - needs research |
| `UserPromptSubmit` | ? | Unknown - needs research |
| `SessionStart` | ? | Unknown - needs research |
| `SessionEnd` | ? | Unknown - needs research |

#### Configuration Translation

**Before (Claude Code):**
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

**After (Cursor):**
```
{Unknown - needs research}
```

#### Migration Steps

1. **Identify hook types** in your Claude Code config
2. **Map events** to Cursor equivalents (see table above)
3. **Translate configuration format** from JSON to Cursor format
4. **Adapt hook scripts** if needed for Cursor's execution environment
5. **Test thoroughly** in Cursor environment

#### Known Limitations

- Event mapping incomplete (research needed)
- Configuration format unknown
- Hook script compatibility unknown

#### Full Migration Example

**Scenario:** Security validation hook for shell commands

**Claude Code (Before):**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "command": "/usr/local/bin/security-check",
        "env": {
          "SECURITY_LEVEL": "high"
        }
      }]
    }]
  }
}
```

**Hook Script:** `/usr/local/bin/security-check`
```bash
#!/bin/bash
if echo "$TOOL_INPUT" | grep -E "(rm -rf|sudo)"; then
  echo "Blocked dangerous command"
  exit 1
fi
exit 0
```

**Cursor (After):**
```
{Unknown - needs research}
```

---

### Cursor → Claude Code

**Migration Difficulty:** Medium
**Status:** 🚧 Partial documentation

#### Event Mapping

| Cursor Event | Claude Code Equivalent | Notes |
|--------------|------------------------|-------|
| ? | `PreToolUse` | Unknown - needs research |
| ? | `PostToolUse` | Unknown - needs research |
| ? | `UserPromptSubmit` | Unknown - needs research |

#### Configuration Translation

**Before (Cursor):**
```
{Unknown - needs research}
```

**After (Claude Code):**
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "command": "/usr/local/bin/hook-script"
      }]
    }]
  }
}
```

#### Migration Steps

1. **Export Cursor hook configuration**
2. **Map Cursor events** to Claude Code equivalents
3. **Convert to Claude Code JSON format**
4. **Place config** in `.claude/settings.json`
5. **Test hooks** with Claude Code CLI

#### Known Limitations

- Cursor event types unknown (research needed)
- Configuration export process unknown
- Feature parity gaps unknown

---

## Claude Code ↔ Copilot

### Claude Code → Copilot

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

#### Event Mapping

{Needs research}

#### Configuration Translation

{Needs research}

#### Migration Steps

{Needs research}

---

### Copilot → Claude Code

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

#### Event Mapping

{Needs research}

#### Configuration Translation

{Needs research}

#### Migration Steps

{Needs research}

---

## Claude Code ↔ Codex

### Claude Code → Codex

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Codex → Claude Code

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Claude Code ↔ Gemini

### Claude Code → Gemini

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Gemini → Claude Code

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Cursor ↔ Copilot

### Cursor → Copilot

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Copilot → Cursor

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Cursor ↔ Codex

### Cursor → Codex

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Codex → Cursor

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Cursor ↔ Gemini

### Cursor → Gemini

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Gemini → Cursor

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Copilot ↔ Codex

### Copilot → Codex

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Codex → Copilot

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Copilot ↔ Gemini

### Copilot → Gemini

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Gemini → Copilot

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Codex ↔ Gemini

### Codex → Gemini

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

### Gemini → Codex

**Migration Difficulty:** Unknown
**Status:** 📝 Research needed

{Needs research}

---

## Common Migration Patterns

### Security Validation Hook

**Pattern:** Block dangerous shell commands before execution

#### Claude Code
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

#### Cursor
```
{Unknown - needs research}
```

#### Copilot
```
{Unknown - needs research}
```

#### Codex
```
{Unknown - needs research}
```

#### Gemini
```
{Unknown - needs research}
```

---

### Code Formatting Hook

**Pattern:** Auto-format code before file edits

#### Claude Code
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

#### Cursor
```
{Unknown - needs research}
```

#### Copilot
```
{Unknown - needs research}
```

#### Codex
```
{Unknown - needs research}
```

#### Gemini
```
{Unknown - needs research}
```

---

### Test Execution Hook

**Pattern:** Run tests after code changes

#### Claude Code
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

#### Cursor
```
{Unknown - needs research}
```

#### Copilot
```
{Unknown - needs research}
```

#### Codex
```
{Unknown - needs research}
```

#### Gemini
```
{Unknown - needs research}
```

---

## Migration Tools & Utilities

### Automated Migration Tools

**Status:** No automated tools exist yet

**Future Possibilities:**
- Configuration format converters
- Event mapping translators
- Hook script adapters
- Compatibility testing frameworks

### Manual Migration Checklist

**Pre-Migration:**
- [ ] Document current hooks and their purposes
- [ ] Identify event types used
- [ ] Review hook scripts for system dependencies
- [ ] Backup current configuration
- [ ] Test hooks in current system

**Migration:**
- [ ] Map events to target system equivalents
- [ ] Translate configuration format
- [ ] Adapt hook scripts if needed
- [ ] Install hooks in target system
- [ ] Verify configuration syntax

**Post-Migration:**
- [ ] Test each hook individually
- [ ] Verify event triggers work correctly
- [ ] Check environment variable availability
- [ ] Monitor for errors or unexpected behavior
- [ ] Document any feature gaps or workarounds

---

## Compatibility Layers

### Potential Compatibility Approaches

**Wrapper Scripts:**
- Create adapter scripts that translate between hook systems
- Use environment variable mapping
- Provide event emulation

**Configuration Converters:**
- Build tools to convert between config formats
- Automated event mapping
- Syntax translation

**Universal Hook Format:**
- Define common denominator hook format
- Build adapters for each system
- Allow single config for multiple systems

**Status:** No compatibility layers exist yet (research opportunity)

---

## Migration Gotchas & Pitfalls

### Common Issues

#### Event Timing Differences

**Problem:** Events may fire at different times in different systems

**Solution:**
- Test hooks thoroughly in target system
- Adjust hook logic for timing differences
- Add defensive checks in hook scripts

#### Environment Variable Differences

**Problem:** Different systems provide different context data

**Solution:**
- Document required environment variables
- Add fallback logic for missing variables
- Use system-agnostic data sources where possible

#### Configuration Syntax

**Problem:** Config format differences can cause subtle errors

**Solution:**
- Validate configuration syntax
- Use schema validation tools
- Test with minimal config first

#### Hook Script Compatibility

**Problem:** Scripts may rely on system-specific features

**Solution:**
- Use portable shell constructs
- Avoid system-specific commands
- Test scripts on target system

---

## Sources & References

### Claude Code
- [Claude Code Hooks Documentation](https://docs.anthropic.com/claude-code/hooks)
- [blues-traveler: Migration Guides](../claude/README.md#migration-guide)

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

## Contributing Migration Guides

**Help expand this matrix!**

If you've successfully migrated hooks between systems:
1. Document your event mappings
2. Share configuration translations
3. Note any gotchas or issues
4. Provide working examples
5. Submit a pull request or issue

**[→ Corpus Contributing Guidelines](../CONTRIBUTING.md)**

---

## See Also

- **[Event Types Comparison](./event-types.md)** - Cross-system event comparison
- **[Configuration Comparison](./configuration.md)** - Config format comparison
- **[Capabilities Matrix](./capabilities.md)** - Feature support comparison
- **[Corpus Home](../README.md)** - Main documentation index

---

**Last Updated:** 2026-02-21
**Status:** In Progress - Claude Code section started, others need research
**Total Migration Paths:** 20 (10 bidirectional pairs)
**Documented Paths:** 2 partial (10%)
**Research Needed:** 18 paths (90%)
