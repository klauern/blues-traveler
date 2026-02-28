# Capabilities Matrix

**Feature support comparison across AI coding assistant hook systems**

This matrix provides a comprehensive comparison of hook system capabilities across all five AI coding assistants, helping developers understand what features are available in each system.

---

## Legend

- ✓ = Fully supported
- ~ = Partial/limited support (see notes)
- ✗ = Not supported
- ? = Unknown/needs research

---

## Event System Capabilities

### Event Types

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Pre-execution hooks** | ✓ | ? | ? | ? | ? |
| **Post-execution hooks** | ✓ | ? | ? | ? | ? |
| **User prompt hooks** | ✓ | ? | ? | ? | ? |
| **Session lifecycle hooks** | ✓ | ? | ? | ? | ? |
| **Context management hooks** | ✓ | ? | ? | ? | ? |
| **Notification hooks** | ✓ | ? | ? | ? | ? |
| **Subagent/task hooks** | ✓ | ? | ? | ? | ? |
| **Custom event types** | ✗ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** 9 built-in event types, cannot define custom events
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Event Filtering & Matching

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Tool/action matching** | ✓ (regex) | ? | ? | ? | ? |
| **File path matching** | ~ (via env vars) | ? | ? | ? | ? |
| **Conditional execution** | ✓ (in script) | ? | ? | ? | ? |
| **Event payload filtering** | ✓ | ? | ? | ? | ? |
| **Glob pattern matching** | ✗ (manual in script) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Uses `matcher` field with regex patterns; file filtering done in hook script
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Event Control Flow

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Cancel/block actions** | ✓ (exit code 1) | ? | ? | ? | ? |
| **Modify action parameters** | ✗ | ? | ? | ? | ? |
| **Chain multiple hooks** | ✓ | ? | ? | ? | ? |
| **Async execution** | ✗ (sequential only) | ? | ? | ? | ? |
| **Parallel hook execution** | ✗ | ? | ? | ? | ? |
| **Hook execution order control** | ✓ (config order) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Hooks execute sequentially in config order; first failure stops chain
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Configuration Capabilities

### Configuration Format

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **JSON support** | ✓ | ? | ? | ? | ? |
| **YAML support** | ✗ | ? | ? | ? | ? |
| **TOML support** | ✗ | ? | ? | ? | ? |
| **Schema validation** | ✓ | ? | ? | ? | ? |
| **IDE autocomplete** | ✓ (VS Code) | ? | ? | ? | ? |
| **Comments in config** | ✗ (JSON limitation) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** JSON only, no comments (use external documentation)
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Configuration Scope

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Project-level config** | ✓ | ? | ? | ? | ? |
| **Global/user config** | ✓ | ? | ? | ? | ? |
| **Config merging** | ✓ (concatenation) | ? | ? | ? | ? |
| **Environment-specific config** | ~ (via env vars) | ? | ? | ? | ? |
| **Per-file config** | ✗ | ? | ? | ? | ? |
| **Dynamic config reload** | ~ (restart required) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Requires restart to reload configuration changes
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Scripting & Execution Capabilities

### Supported Languages

| Language | Claude Code | Cursor | Copilot | Codex | Gemini |
|----------|-------------|--------|---------|-------|--------|
| **Bash/Shell** | ✓ | ? | ? | ? | ? |
| **Python** | ✓ (via shebang) | ? | ? | ? | ? |
| **Node.js** | ✓ (via shebang) | ? | ? | ? | ? |
| **Ruby** | ✓ (via shebang) | ? | ? | ? | ? |
| **Native binaries** | ✓ | ? | ? | ? | ? |
| **Inline scripts** | ✗ (must be files) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Any executable with proper shebang works
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Script Execution Features

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Exit code handling** | ✓ | ? | ? | ? | ? |
| **Stdout capture** | ✓ | ? | ? | ? | ? |
| **Stderr capture** | ✓ | ? | ? | ? | ? |
| **Timeout support** | ✓ (default 30s) | ? | ? | ? | ? |
| **Resource limits** | ~ (OS-level only) | ? | ? | ? | ? |
| **Working directory control** | ✓ (project root) | ? | ? | ? | ? |
| **Custom environment variables** | ✓ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** 30-second default timeout, configurable via env var
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Context & Environment Capabilities

### Available Context Data

| Context | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Tool/action name** | ✓ (`TOOL_NAME`) | ? | ? | ? | ? |
| **Tool parameters** | ✓ (`TOOL_INPUT`) | ? | ? | ? | ? |
| **User prompt** | ✓ (UserPromptSubmit) | ? | ? | ? | ? |
| **File paths** | ✓ (via TOOL_INPUT) | ? | ? | ? | ? |
| **Git information** | ✗ (manual query) | ? | ? | ? | ? |
| **Project metadata** | ✗ (manual query) | ? | ? | ? | ? |
| **Session ID** | ✓ (SessionStart) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Context via environment variables; git/project data requires manual queries
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Environment Variables

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **System env vars** | ✓ | ? | ? | ? | ? |
| **Custom env vars** | ✓ (in config) | ? | ? | ? | ? |
| **Event-specific vars** | ✓ | ? | ? | ? | ? |
| **Secret injection** | ~ (via external tools) | ? | ? | ? | ? |
| **Dynamic env vars** | ✗ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Static env vars in config; secrets should use external managers
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Security Capabilities

### Permission & Sandboxing

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Script sandboxing** | ✗ | ? | ? | ? | ? |
| **Permission prompts** | ✗ | ? | ? | ? | ? |
| **Allowlist/denylist** | ~ (via scripts) | ? | ? | ? | ? |
| **File system isolation** | ✗ | ? | ? | ? | ? |
| **Network isolation** | ✗ | ? | ? | ? | ? |
| **Resource limits** | ~ (timeout only) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** No built-in sandboxing; hooks run with full user permissions
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Safety Features

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Dangerous command blocking** | ✓ (via hooks) | ? | ? | ? | ? |
| **Audit logging** | ~ (hook output) | ? | ? | ? | ? |
| **Rollback support** | ✗ | ? | ? | ? | ? |
| **Dry-run mode** | ✗ | ? | ? | ? | ? |
| **Hook disable flag** | ✓ (env var) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Safety via custom hook scripts; no built-in rollback
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Integration Capabilities

### External Tool Integration

| Integration | Claude Code | Cursor | Copilot | Codex | Gemini |
|-------------|-------------|--------|---------|-------|--------|
| **Git hooks** | ✓ | ? | ? | ? | ? |
| **CI/CD integration** | ✓ | ? | ? | ? | ? |
| **Build tools** | ✓ | ? | ? | ? | ? |
| **Linters/formatters** | ✓ | ? | ? | ? | ? |
| **Test runners** | ✓ | ? | ? | ? | ? |
| **Notification services** | ✓ (via scripts) | ? | ? | ? | ? |
| **Monitoring/metrics** | ~ (manual logging) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Integration via hook scripts calling external tools
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### API & Extensibility

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Hook development API** | ✗ | ? | ? | ? | ? |
| **Plugin system** | ✗ | ? | ? | ? | ? |
| **Custom event types** | ✗ | ? | ? | ? | ? |
| **Hook marketplace** | ✗ | ? | ? | ? | ? |
| **Shared hook libraries** | ~ (manual sharing) | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** No formal API; hooks are shell scripts
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Debugging & Observability

### Debugging Features

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Verbose logging** | ✓ | ? | ? | ? | ? |
| **Hook execution trace** | ✓ (stdout/stderr) | ? | ? | ? | ? |
| **Error messages** | ✓ | ? | ? | ? | ? |
| **Hook testing mode** | ~ (manual testing) | ? | ? | ? | ? |
| **Performance profiling** | ✗ | ? | ? | ? | ? |
| **Debug breakpoints** | ✗ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Basic logging via hook stdout/stderr
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Observability

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Execution logs** | ✓ (stdout/stderr) | ? | ? | ? | ? |
| **Metrics collection** | ✗ | ? | ? | ? | ? |
| **Performance monitoring** | ✗ | ? | ? | ? | ? |
| **Hook success/failure tracking** | ~ (via exit codes) | ? | ? | ? | ? |
| **Execution time tracking** | ✗ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Manual logging in hook scripts; no built-in metrics
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Performance Capabilities

### Execution Performance

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Async execution** | ✗ (sequential) | ? | ? | ? | ? |
| **Parallel hooks** | ✗ | ? | ? | ? | ? |
| **Lazy evaluation** | ✗ | ? | ? | ? | ? |
| **Caching** | ~ (manual in scripts) | ? | ? | ? | ? |
| **Optimization hints** | ✗ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Sequential execution only; caching requires custom implementation
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Resource Management

| Feature | Claude Code | Cursor | Copilot | Codex | Gemini |
|---------|-------------|--------|---------|-------|--------|
| **Timeout configuration** | ✓ (env var) | ? | ? | ? | ? |
| **Memory limits** | ✗ | ? | ? | ? | ? |
| **CPU limits** | ✗ | ? | ? | ? | ? |
| **Disk I/O limits** | ✗ | ? | ? | ? | ? |
| **Network limits** | ✗ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Only timeout limits; other limits via OS-level tools
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Platform Support

### Operating Systems

| Platform | Claude Code | Cursor | Copilot | Codex | Gemini |
|----------|-------------|--------|---------|-------|--------|
| **macOS** | ✓ | ? | ? | ? | ? |
| **Linux** | ✓ | ? | ? | ? | ? |
| **Windows** | ~ (WSL recommended) | ? | ? | ? | ? |
| **Docker/Containers** | ✓ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Native macOS/Linux; Windows via WSL
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

### Development Environments

| Environment | Claude Code | Cursor | Copilot | Codex | Gemini |
|-------------|-------------|--------|---------|-------|--------|
| **CLI** | ✓ (native) | ? | ? | ? | ? |
| **VS Code** | ✓ | ? | ? | ? | ? |
| **JetBrains IDEs** | ✗ | ? | ? | ? | ? |
| **Vim/Neovim** | ~ (via CLI) | ? | ? | ? | ? |
| **Emacs** | ~ (via CLI) | ? | ? | ? | ? |
| **GitHub Codespaces** | ✓ | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Primary CLI tool with IDE integrations
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Documentation & Support

### Documentation Quality

| Aspect | Claude Code | Cursor | Copilot | Codex | Gemini |
|--------|-------------|--------|---------|-------|--------|
| **Official docs** | ✓ Good | ? | ? | ? | ? |
| **Examples** | ✓ Yes | ? | ? | ? | ? |
| **API reference** | ~ Limited | ? | ? | ? | ? |
| **Migration guides** | ✗ | ? | ? | ? | ? |
| **Community resources** | ~ Growing | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Good official docs, this corpus fills migration gaps
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Maturity & Stability

### System Maturity

| Aspect | Claude Code | Cursor | Copilot | Codex | Gemini |
|--------|-------------|--------|---------|-------|--------|
| **Stable release** | ✓ v1.0+ | ? | ? | ? | ? |
| **Breaking changes** | ~ Rare | ? | ? | ? | ? |
| **Backward compatibility** | ✓ Good | ? | ? | ? | ? |
| **Deprecation warnings** | ✓ | ? | ? | ? | ? |
| **LTS support** | ? | ? | ? | ? | ? |

**Notes:**
- **Claude Code:** Stable and mature hook system
- **Cursor:** ?
- **Copilot:** ?
- **Codex:** ?
- **Gemini:** ?

---

## Sources & References

### Claude Code
- [Claude Code Documentation](https://docs.anthropic.com/claude-code/)
- [blues-traveler: System Documentation](../claude/README.md)
- Community examples and reports

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
- **[Configuration Comparison](./configuration.md)** - Config format comparison
- **[Migration Matrix](./migration-matrix.md)** - System-to-system migration
- **[Corpus Home](../README.md)** - Main documentation index

---

**Last Updated:** 2026-02-21
**Status:** In Progress - Claude Code complete, others in research phase
