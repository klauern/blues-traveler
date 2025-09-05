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

## Developer Resources

### 🛠️ [Developer Guide](developer_guide.md)
### 🧩 [Custom Hooks Guide](custom_hooks.md)

**For contributors and developers** - Learn how to extend Blues Traveler.

- Architecture overview
- Adding new hooks
- Development workflow
- Best practices
- Testing guidelines

### 🏗️ [Architecture Design](unified_pipeline_design.md)

**Technical deep dive** - Understand the internal design and decisions.

- Current architecture
- Design principles
- Hook execution flow
- Future considerations

## For AI Assistants

### 🤖 [CLAUDE.md](../CLAUDE.md)

**AI assistant guidance** - Specific instructions for working with this codebase.

- Project overview
- Architecture details
- Development patterns
- What to do/not do

## Project Structure

```
blues-traveler/
├── README.md                 # Main user documentation (urfave/cli v3 based)
├── CLAUDE.md                 # AI assistant guidance
├── docs/
│   ├── index.md             # This documentation index
│   ├── quick_start.md       # Getting started guide
│   ├── developer_guide.md   # Developer reference
│   ├── custom_hooks.md      # Custom hooks usage
│   └── unified_pipeline_design.md  # Architecture details
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
blues-traveler list

# Install a hook
blues-traveler install <hook-name> --event <event-type>

# Run a hook manually
blues-traveler run <hook-name> --log

# Check configuration
blues-traveler list-installed
```

### Key Hooks

| Hook | Purpose | Best Event |
|------|---------|------------|
| `security` | Block dangerous commands | `PreToolUse` |
| `format` | Auto-format code | `PostToolUse` |
| `debug` | Log operations | Any event |
| `audit` | JSON audit logging | Any event |

### Configuration Files

- **Project**: `./.claude/settings.json`
- **Global**: `~/.claude/settings.json`

## Getting Help

### For Users

1. Start with the [Quick Start Guide](quick_start.md)
2. Reference the [Main README](../README.md) for detailed information
3. Check the troubleshooting sections

### For Developers

1. Read the [Developer Guide](developer_guide.md)
2. Review the [Architecture Design](unified_pipeline_design.md)
3. Examine existing hook implementations in `internal/hooks/`

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
