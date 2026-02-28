# Cursor IDE Hooks & Automation Reference

**Version:** 1.0 (Beta as of February 2026)
**Last Updated:** 2026-02-21
**Status:** Reference draft; advanced sections are placeholder pages

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

### What are Hooks in Cursor IDE?

Hooks are **external scripts** that execute automatically at specific lifecycle events in Cursor's AI agent. They enable you to:

- **Control agent behavior** (approve/deny shell commands, MCP tools, file reads)
- **Automate post-actions** (auto-format edited files, send notifications)
- **Enforce policies** (block dangerous commands, validate prompts)
- **Integrate external systems** (security scanning, audit logging, CI/CD)

Hooks run as standalone subprocesses, receiving event data via JSON stdin and controlling execution through JSON stdout responses.

**Introduced:** October 2025 (Cursor 1.7)
**Status:** Beta (as of February 2026)

**Sources:**
- [Cursor Official Documentation](https://cursor.com/docs/agent/hooks)
- [InfoQ Announcement](https://www.infoq.com/news/2025/10/cursor-hooks/)

### Design Philosophy

Cursor's hook system embodies several core principles:

**Key Principles:**

1. **Lifecycle interception** - Hooks fire at precise moments (before/after actions)
2. **Permission-based control** - Native support for allow/deny/ask modes
3. **JSON-based communication** - Standard stdin/stdout protocol with well-defined schemas
4. **Standalone execution** - Each hook runs as isolated subprocess
5. **Multi-level configuration** - Project, user, and system-level hooks merge
6. **Simple configuration** - Single JSON file (`.cursor/hooks.json`)
7. **Beta transparency** - Cursor explicitly labels this feature as beta

**Sources:**
- [GitButler Deep Dive](https://blog.gitbutler.com/cursor-hooks-deep-dive)
- [Skywork AI Guide](https://skywork.ai/blog/how-to-cursor-1-7-hooks-guide/)

### When to Use Hooks

**Use hooks when:**
- ✅ Blocking dangerous shell commands (`rm -rf /`, `sudo rm`)
- ✅ Enforcing tool preferences (bun over npm, gh over git)
- ✅ Scanning for secrets/credentials in files
- ✅ Auto-formatting code after edits
- ✅ Auditing agent actions for compliance
- ✅ Integrating malware detection (Endor Labs, Semgrep)
- ✅ Managing MCP server access control

**Don't use hooks when:**
- ❌ You need to modify Cursor's internal behavior (hooks are external)
- ❌ The operation requires complex state management (use proper tools)
- ❌ You want to inject context into prompts (not supported)
- ❌ Performance is critical (hooks add synchronous latency)
- ❌ You need async/parallel execution (all hooks are synchronous)

**Sources:**
- [DataCamp Tutorial](https://www.datacamp.com/tutorial/cursor-ai-hooks)
- [egghead.io Course](https://egghead.io/courses/advanced-cursor-hooks~swp89)

### Version & Compatibility

| Version | Released | Status | Notes |
|---------|----------|--------|-------|
| 1.0 (Beta) | October 2025 | Beta | 6 events, JSON config |
| Future | TBD | Planned | API may change |

**Platform Support:** macOS, Linux, Windows (WSL)

**Beta Implications:**
- API stability not guaranteed
- Breaking changes possible in future versions
- Community-driven documentation supplements official docs
- Production use at your own risk

**See:** [Configuration & Setup](./configuration.md) for version-specific details

---

## Quick Start

### Basic Hook Example

**Block dangerous shell commands:**

```bash
#!/bin/bash
# .cursor/hooks/security-check.sh

input=$(cat)
command=$(echo "$input" | jq -r '.command')

if echo "$command" | grep -qE '(rm -rf /|sudo rm|mkfs)'; then
  echo '{
    "permission": "deny",
    "userMessage": "Dangerous command blocked",
    "agentMessage": "Blocked pattern: filesystem destruction"
  }'
  exit 0
fi

echo '{"permission":"allow"}'
```

**Configure in `.cursor/hooks.json`:**

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security-check.sh" }
    ]
  }
}
```

**Make executable and restart Cursor:**
```bash
chmod +x .cursor/hooks/security-check.sh
# Restart Cursor IDE
```

---

## The 6 Hook Events

| Event | Trigger | Cancellable | Permission Modes | Use Case |
|-------|---------|-------------|------------------|----------|
| beforeShellExecution | Before shell command | ✅ | allow/deny/ask | Security, tool enforcement |
| beforeMCPExecution | Before MCP tool | ✅ | allow/deny/ask | MCP governance |
| beforeReadFile | Before file read | ✅ | allow/deny | Secret protection |
| afterFileEdit | After file modified | ❌ | N/A | Formatting, audit |
| beforeSubmitPrompt | Before prompt sent | ✅ | continue only | Prompt gating |
| stop | Agent completes | ❌ | followup only | Cleanup, summary |

**See:** [Event Types & Triggers](./events-reference.md) for complete schemas

---

## Documentation Structure

This reference provides comprehensive coverage across 12 focused sections:

1. **[Architecture & Internals](./architecture.md)** - Execution model, lifecycle, performance
2. **[Event Types & Triggers](./events-reference.md)** - Complete catalog with schemas
3. **[Configuration & Setup](./configuration.md)** - JSON format, file locations, validation
4. **[Environment & Context](./environment-context.md)** - JSON payloads, IDs, workspace roots
5. **[Scripting & Execution](./scripting.md)** - Languages, exit codes, error handling
6. **[Security & Safety](./security.md)** - Permissions, sandboxing, best practices
7. **[Common Patterns & Recipes](./examples.md)** - Copy-paste ready examples
8. **[Integration & Extensibility](./integration.md)** - SDKs, enterprise integrations
9. **[Troubleshooting & Debugging](./troubleshooting.md)** - Common issues, debug techniques
10. **[API Reference](./api-reference.md)** - Complete schema reference
11. **[Migration Guide](./migration.md)** - Moving to/from Claude Code, Copilot, etc.
12. **[Working Examples](./examples/)** - Ready-to-use scripts

---

## Key Differences vs Other IDEs

### Cursor vs Claude Code

| Feature | Cursor | Claude Code |
|---------|--------|-------------|
| Events | 6 specific | 17 generic + matchers |
| Config Format | JSON only | JSON + YAML |
| Ask Permission | ✅ Native | ⚠️ Fallback |
| Matchers | ❌ No | ✅ Regex matchers |
| Inline Scripts | ❌ External only | ✅ YAML inline |
| SessionStart | ❌ No | ✅ Yes |
| Cross-Compatible | Via blues-traveler | ✅ Native |

**Winner:** Cursor (simpler), Claude Code (more powerful)

**See:** [Migration Guide](./migration.md) for cross-IDE compatibility

---

## Community Ecosystem

### Official Resources
- [Cursor Documentation](https://cursor.com/docs/agent/hooks)
- [JSON Schema](https://unpkg.com/cursor-hooks/schema/hooks.schema.json)

### SDKs and Libraries
- **TypeScript:** [cursor-hooks](https://github.com/johnlindquist/cursor-hooks) (npm)
- **Python:** [py-cursor-hooks](https://github.com/DevonFulcher/py-cursor-hooks) (PyPI)
- **Bash Examples:** [hamzafer/cursor-hooks](https://github.com/hamzafer/cursor-hooks)

### Enterprise Integrations
- **1Password:** [Validation hooks](https://github.com/1Password/cursor-hooks)
- **Endor Labs:** [Malware detection](https://github.com/endorlabs/cursor-hook-examples)
- **StacklokLabs:** [MCP governance](https://github.com/StacklokLabs/cursor-hooks)

### Tutorials and Guides
- [GitButler Deep Dive](https://blog.gitbutler.com/cursor-hooks-deep-dive) ⭐⭐⭐⭐⭐
- [egghead.io Video Course](https://egghead.io/courses/advanced-cursor-hooks~swp89) ⭐⭐⭐⭐⭐
- [Skywork AI Comprehensive Guide](https://skywork.ai/blog/how-to-cursor-1-7-hooks-guide/) ⭐⭐⭐⭐

**See:** [Integration & Extensibility](./integration.md) for complete ecosystem

---

## Additional Resources

### Comparison Tables
- [Event Types Comparison](../comparison-tables/event-types.md)
- [Configuration Format Comparison](../comparison-tables/configuration.md)
- [Capabilities Matrix](../comparison-tables/capabilities.md)
- [Migration Matrix](../comparison-tables/migration-matrix.md)

### Related Systems
- [Claude Code](../claude/README.md) - 17 events, powerful matchers
- [Copilot](../copilot/README.md) - Extension-based approach
- [Codex](../codex/README.md) - API-driven hooks

---

## Production Readiness

**✅ Production-Ready For:**
- Security controls (command blocking)
- Audit logging
- Secret scanning
- Tool enforcement

**⚠️ Use with Caution:**
- Auto-formatting (performance impact)
- External API calls (timeout risk)
- Complex workflows (beta status)

**❌ Not Recommended:**
- Long-running operations (> 10s)
- Untrusted third-party hooks
- Critical business logic (beta feature)

**Overall Rating:** ⭐⭐⭐⭐ (Very Good - pending GA)

---

## Known Gaps

**Critical Unknowns:**
- Timeout behavior (what happens on timeout?)
- Multi-hook coordination (which result wins?)
- Error recovery (retry logic?)
- Performance limits (official benchmarks?)
- Environment variables (what's available?)

**See:** Research notes at `/research-notes/cursor/gaps.md` for testing priorities

---

**Document Version:** 1.0
**Research Sources:** 77 verified sources
**Coverage:** 6/6 events documented (100%)
**Community:** 60+ blog posts, tutorials, videos analyzed
