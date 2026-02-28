# Hooks & Automation Encyclopedia

**The definitive reference for hooks, callbacks, and automation across AI coding assistants**

---

## 📚 Overview

This corpus provides comprehensive documentation for hook systems across the five major AI coding assistants:

- **[Claude Code](./claude/README.md)** - Anthropic's CLI tool with extensive hook support
- **[Cursor IDE](./cursor/README.md)** - AI-first code editor with callback system
- **[GitHub Copilot](./copilot/README.md)** - GitHub's AI pair programmer
- **[OpenAI Codex](./codex/README.md)** - OpenAI's code generation API
- **[Google Gemini](./gemini/README.md)** - Google's AI coding assistant

**Goal:** Provide developers with a unified reference for understanding, implementing, and migrating hook configurations across different AI coding tools.

---

## 🗺️ Quick Navigation

### By System

| System | Current State | Core Pages | Examples | Planned Pages |
|--------|---------------|------------|----------|---------------|
| [Claude Code](./claude/README.md) | Core reference complete | 7 | Guide + staging dir | Placeholder pages tracked |
| [Cursor IDE](./cursor/README.md) | Reference draft | 6 | Placeholder guide + staging dir | Placeholder pages tracked |
| [GitHub Copilot](./copilot/README.md) | Reference draft | 7 | Guide + staging dir | Placeholder pages tracked |
| [OpenAI Codex](./codex/README.md) | Reference draft | 7 | Guide + staging dir | Placeholder pages tracked |
| [Google Gemini](./gemini/README.md) | Reference draft | 6 | Placeholder guide + staging dir | Placeholder pages tracked |

### By Use Case

| Use Case | Claude | Cursor | Copilot | Codex | Gemini |
|----------|--------|--------|---------|-------|--------|
| **Code Formatting** | [Guide](./claude/examples.md) | [Guide](./cursor/examples.md) | [Guide](./copilot/examples.md) | [Guide](./codex/examples.md) | [Guide](./gemini/examples.md) |
| **Test Execution** | [Examples Staging](./claude/examples/) | [Examples Staging](./cursor/examples/) | [Examples Staging](./copilot/examples/) | [Examples Staging](./codex/examples/) | [Examples Staging](./gemini/examples/) |
| **Security Validation** | [Security](./claude/security.md) | [Security](./cursor/security.md) | [Security](./copilot/security.md) | [Security](./codex/security.md) | [Security](./gemini/security.md) |
| **Custom Workflows** | [Events](./claude/events-reference.md) | [Events](./cursor/events-reference.md) | [Events](./copilot/events-reference.md) | [Events](./codex/events-reference.md) | [Events](./gemini/events-reference.md) |
| **CI/CD Integration** | [Status](./claude/COMPLETION_STATUS.md) | [Status](./../hooks-corpus/IMPLEMENTATION_STATUS.md) | [Status](./IMPLEMENTATION_STATUS.md) | [Status](./IMPLEMENTATION_STATUS.md) | [Status](./IMPLEMENTATION_STATUS.md) |

---

## 🔍 Comparison Tables

Cross-system comparison tables for understanding equivalent features and migrating between systems:

- **[Event Types](./comparison-tables/event-types.md)** - Equivalent events across all systems
- **[Configuration Formats](./comparison-tables/configuration.md)** - Config file formats and locations
- **[Capabilities Matrix](./comparison-tables/capabilities.md)** - Feature support comparison
- **[Migration Matrix](./comparison-tables/migration-matrix.md)** - Bidirectional migration guides

---

## 📖 System Documentation

### Claude Code (Anthropic)

**Status:** ✅ Complete | **Depth:** Encyclopedia-level

Comprehensive documentation of Claude Code's hook system, including:
- 7+ event types (PreToolUse, PostToolUse, UserPromptSubmit, etc.)
- JSON configuration format
- Bash/shell script execution
- Security model and permissions
- examples guide plus staged example directory

**[→ Read Claude Code Documentation](./claude/README.md)**

---

### Cursor IDE

**Status:** 🚧 In Progress | **Depth:** Deep-dive

Detailed documentation of Cursor's callback system:
- Event-driven automation
- Configuration format
- Integration with IDE features
- examples guide plus staged example directory

**[→ Read Cursor Documentation](./cursor/README.md)**

---

### GitHub Copilot

**Status:** 📝 Research Phase | **Depth:** Under investigation

Documentation coverage:
- Event system research
- API integration points
- Extension possibilities

**[→ Read Copilot Documentation](./copilot/README.md)**

---

### OpenAI Codex

**Status:** 📝 Research Phase | **Depth:** Under investigation

Documentation coverage:
- API-level hooks
- Integration patterns
- Callback mechanisms

**[→ Read Codex Documentation](./codex/README.md)**

---

### Google Gemini

**Status:** 📝 Research Phase | **Depth:** Under investigation

Documentation coverage:
- Hook system research
- Event model
- Configuration approach

**[→ Read Gemini Documentation](./gemini/README.md)**

---

## 🚀 Quick Start

### For New Users

1. **Choose your AI coding assistant** from the list above
2. **Read the system overview** to understand the hook model
3. **Browse common patterns** for your use case
4. **Copy and adapt examples** for your workflow

### For Migrating Users

1. **Check the [Migration Matrix](./comparison-tables/migration-matrix.md)** for your source → target systems
2. **Review [Event Types comparison](./comparison-tables/event-types.md)** to map equivalent events
3. **Follow the migration guide** in your target system's documentation
4. **Test with simple examples** before full migration

### For Contributing

1. **See [Corpus Contributing Guide](./CONTRIBUTING.md)** for scope, status, and validation rules
2. **Use the [master template](./templates/system-template.md)** for new system documentation
3. **Add working examples** to the `examples/` directory
4. **Update comparison tables** with new system data

---

## 📊 Documentation Structure

Each system follows a consistent 12-section template:

1. **Overview & Philosophy** - What hooks are and when to use them
2. **Architecture & Internals** - How hooks are executed
3. **Event Types & Triggers** - Complete event catalog
4. **Configuration & Setup** - Config format and installation
5. **Environment & Context** - Available variables and data
6. **Scripting & Execution** - Exit codes, languages, timeouts
7. **Security & Safety** - Permissions and safety features
8. **Common Patterns & Recipes** - Working examples
9. **Integration & Extensibility** - Build tools and CI/CD
10. **Troubleshooting & Debugging** - Common issues and solutions
11. **API Reference** - Complete schema and API documentation
12. **Migration Guide** - How to migrate TO and FROM this system

**[→ View Master Template](./templates/system-template.md)**

---

## 🎯 Scope & Coverage

### Current Status

**Total Documentation Pages:** ~250-500 pages (estimated when complete)
**Working Code Examples:** Mixed; Codex and Copilot have authored examples, other systems have placeholder guides and staged directories
**Comparison Tables:** 4 comprehensive cross-system tables
**Migration Paths:** 10+ bidirectional migration guides

### Completion Criteria

The corpus is considered **complete** when:

- ✅ All 5 systems have full documentation (12 sections each)
- ✅ All 4 comparison tables populated with sources
- ✅ 80%+ of capabilities matrix cells have definitive data (not "?")
- ✅ Migration guides exist for 70%+ of system pairs
- ⏳ Each system has a stable examples guide and at least one committed example or clearly-marked placeholder
- ✅ All documentation passes quality checklist
- ✅ All cross-references validated

---

## 📚 Documentation Quality

### Quality Standards

Every system documentation must:

1. **Follow the master template** exactly (all 12 sections)
2. **Include source citations** for all claims
3. **Provide working code examples or clearly-marked placeholders**
4. **Document limitations** and edge cases
5. **Cross-reference** other systems and comparison tables
6. **Specify versions** and compatibility information

### Validation Checklist

- [ ] All 12 template sections completed
- [ ] Every event type documented with examples
- [ ] Configuration schema fully specified
- [ ] Examples links resolve to committed files or directories
- [ ] Migration guides TO and FROM 2+ other systems
- [ ] All comparison tables updated
- [ ] All source citations accessible
- [ ] All cross-references valid

**[→ View Full Quality Checklist](./templates/system-template.md#verification--quality-assurance)**

---

## 🔗 Related Resources

### Within blues-traveler

- **[Custom Hooks Guide](../custom-hooks.md)** - Blues-traveler hook development
- **[Cursor Compatibility](../cursor-compatibility.md)** - Quick Cursor setup guide
- **[Main README](../../README.md)** - Blues-traveler project overview

### External Resources

- **[Claude Code Documentation](https://docs.anthropic.com/claude-code/)** - Official Claude Code docs
- **[Cursor Documentation](https://cursor.com/docs)** - Official Cursor docs
- **[GitHub Copilot Docs](https://docs.github.com/en/copilot)** - Official Copilot docs
- **[OpenAI API Docs](https://platform.openai.com/docs/)** - OpenAI Codex API
- **[Google Gemini Docs](https://ai.google.dev/)** - Gemini AI documentation

---

## 🤝 Contributing

Contributions are welcome! This corpus is a community effort to document the hook systems across AI coding assistants.

**Ways to contribute:**
- Document additional systems or features
- Add working code examples
- Improve comparison tables
- Fix errors or outdated information
- Share migration experiences

**[→ Read Corpus Contributing Guidelines](./CONTRIBUTING.md)**

---

## 📜 License

This documentation corpus is part of the blues-traveler project and is licensed under the same terms.

**[→ View License](../../LICENSE)**

---

## 📞 Support & Community

- **Issues:** [GitHub Issues](https://github.com/yourusername/blues-traveler/issues)
- **Discussions:** [GitHub Discussions](https://github.com/yourusername/blues-traveler/discussions)
- **Project:** [blues-traveler](https://github.com/yourusername/blues-traveler)

---

**Last Updated:** 2026-02-21
**Corpus Version:** 1.0
**Status:** In Active Development

---

## Navigation

**[↑ Back to docs](../index.md)** | **[↑ Back to project](../../README.md)**
