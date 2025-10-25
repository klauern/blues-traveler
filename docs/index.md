# Documentation Index

Welcome to the Blues Traveler documentation. This index will help you find the information you need.

## Getting Started

### 🚀 [Quick Start Guide](quick_start.md)

**For new users** - Get up and running in minutes with step-by-step instructions.

- Installation and setup
- First hook configuration
- Common usage patterns
- Troubleshooting basics

## User Guides

### 📖 [Main README](../README.md)

**Complete user reference** - Comprehensive documentation of all features and commands.

- Feature overview
- Command reference
- Configuration examples
- Usage patterns

### 🧩 [Custom Hooks Guide](custom-hooks-guide.md)

**Flexible automation** - Define project-specific hooks using YAML/JSON configuration.

- Configuration format and examples
- Environment variables and expressions
- Replacing built-in hooks
- Best practices and patterns

### 🖱️ [Cursor Support](cursor-support.md)

**Cursor IDE integration** - Full support for Cursor hooks alongside Claude Code.

- Platform differences
- Installation and usage
- Event mapping
- Matcher support

## Developer Resources

### 🛠️ [Developer Guide](developer_guide.md)

**For contributors and developers** - Learn how to extend Blues Traveler.

- Architecture overview
- Adding new hooks
- Development workflow
- Best practices
- Testing guidelines

### 🏗️ [Architecture Documentation](architecture/)

**Technical deep dive** - Understand the internal design and decisions.

- [Unified Pipeline Design](architecture/unified-pipeline.md) - Current architecture
- [XDG Migration](architecture/xdg-migration.md) - Configuration migration
- Design principles and patterns
- Hook execution flow

### 📋 [Code Reviews](reviews/)

**Quality assurance** - Code reviews and audit findings.

- [Code Review 2024](reviews/code-review-2024.md) - Comprehensive audit
- Issues tracked in beads (`.beads/` directory)

## For AI Assistants

### 🤖 [AGENTS.md](../AGENTS.md)

**AI assistant guidance** - Specific instructions for working with this codebase.

- Project overview
- Architecture details
- Development patterns with beads
- What to do/not do
- Issue tracking workflow

## Project Structure

```
blues-traveler/
├── README.md                 # Main user documentation (urfave/cli v3 based)
├── AGENTS.md                 # AI assistant guidance (formerly CLAUDE.md)
├── .beads/                   # Issue tracking with beads
├── docs/
│   ├── index.md             # This documentation index
│   ├── quick_start.md       # Getting started guide
│   ├── developer_guide.md   # Developer reference
│   ├── custom_hooks.md      # Custom hooks usage
│   ├── architecture/        # Architecture documentation
│   │   ├── README.md        # Architecture index
│   │   ├── unified-pipeline.md  # Current architecture
│   │   └── xdg-migration.md     # XDG config migration
│   ├── reviews/             # Code reviews and audits
│   │   ├── README.md        # Reviews index
│   │   └── code-review-2024.md  # 2024 audit
│   ├── development/         # Development workflows
│   │   └── beads-workflow.md    # Issue tracking workflow
│   └── archive/             # Historical documents
│       ├── cobra-to-urfave-cli-migration.md
│       ├── xdg-migration-spec.md
│       ├── code_review_2024.md
│       └── cursor-planning/ # Cursor implementation planning docs
├── internal/
│   ├── cmd/                 # CLI command implementations
│   ├── hooks/               # Hook implementations
│   ├── core/                # Core functionality
│   └── config/              # Configuration management
└── Taskfile.yml             # Build and development tasks
```

## Quick Reference

### Common Commands

```bash
# List available hooks
blues-traveler hooks list

# Install a hook (auto-detects platform)
blues-traveler hooks install <hook-name> --event <event-type>

# Install for specific platform
blues-traveler hooks install <hook-name> --platform cursor --event <event-type>

# Run a hook manually
blues-traveler hooks run <hook-name> --log

# Check configuration
blues-traveler hooks list --installed

# Platform detection
blues-traveler platform detect
```

### Key Built-in Hooks

| Hook | Purpose | Best Event |
|------|---------|------------|
| `security` | Block dangerous commands | `PreToolUse` |
| `format` | Auto-format code | `PostToolUse` |
| `vet` | Code quality checks | `PostToolUse` |
| `debug` | Log operations | Any event |
| `audit` | JSON audit logging | Any event |

### Configuration Files

| Platform | Project | Global |
|----------|---------|--------|
| **Claude Code** | `./.claude/settings.json` | `~/.claude/settings.json` |
| **Cursor** | N/A | `~/.cursor/hooks.json` |
| **Custom Hooks** | `~/.config/blues-traveler/projects/<name>.yml` | `~/.config/blues-traveler/global.yml` |

## Getting Help

### For Users

1. Start with the [Quick Start Guide](quick_start.md)
2. Reference the [Main README](../README.md) for detailed information
3. Check [Custom Hooks Guide](custom-hooks-guide.md) for advanced automation
4. See [Cursor Support](cursor-support.md) if using Cursor IDE

### For Developers

1. Read the [Developer Guide](developer_guide.md)
2. Review the [Architecture Documentation](architecture/)
3. Check [Code Reviews](reviews/) for improvement areas
4. Use beads for issue tracking: `bd list`, `bd ready`, `bd show <id>`
5. Examine existing hook implementations in `internal/hooks/`
6. Review test patterns in `*_test.go` files

### For AI Assistants

1. Follow the guidance in [AGENTS.md](../AGENTS.md)
2. Use the [Developer Guide](developer_guide.md) for implementation details
3. Track work with beads: `bd list`, `bd create`, `bd update`
4. Reference existing code patterns

## Contributing

Want to improve the documentation?

1. **Report Issues**: Open an issue for unclear or missing information
2. **Suggest Improvements**: Propose changes via pull requests
3. **Add Examples**: Help others by adding practical examples
4. **Fix Typos**: Even small improvements help!

## Documentation Standards

- **Clear and Concise**: Write for the intended audience
- **Examples First**: Show before telling
- **Consistent Format**: Follow existing patterns
- **Up-to-Date**: Keep documentation current with code changes
- **Searchable**: Use descriptive headings and clear structure
- **Single Source of Truth**: One topic per file, no duplication

---

**Last Updated**: 2025-10-01
**Cleanup**: Reduced from 14 files to 5 core docs (64% reduction)
