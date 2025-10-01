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

- **Architecture and design** - System design, execution flow, benefits
- Adding new hooks - Step-by-step guide
- Development workflow - Build, test, and contribute
- Best practices - Hook design, code style, performance
- Testing guidelines - Patterns and examples

## Project Structure

```
docs/
├── index.md                  # This documentation index
├── quick_start.md            # Getting started guide
├── developer_guide.md        # Complete developer reference (includes architecture)
├── custom-hooks-guide.md     # Custom hooks documentation
├── cursor-support.md         # Cursor platform support
└── archive/                  # Historical documents
    ├── cobra-to-urfave-cli-migration.md
    ├── xdg-migration-spec.md
    ├── code_review_2024.md
    └── cursor-planning/      # Cursor implementation planning docs
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

1. Read the [Developer Guide](developer_guide.md) (includes architecture)
2. Examine existing hook implementations in `internal/hooks/`
3. Review test patterns in `*_test.go` files

### For AI Assistants

1. Follow the guidance in [CLAUDE.md](../CLAUDE.md)
2. Use the [Developer Guide](developer_guide.md) for implementation details
3. Reference existing code patterns

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
