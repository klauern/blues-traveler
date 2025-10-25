# Architecture Documentation

This directory contains architecture documents, design decisions, and technical specifications for Blues Traveler.

## Documents

### Active Architecture

- **[Unified Pipeline Design](unified-pipeline.md)** - Current architecture using static hook registry with independent execution model
- **[XDG Migration](xdg-migration.md)** - Completed migration to XDG-compliant configuration structure

## Architecture Principles

Blues Traveler follows these core architectural principles:

1. **Static Registration**: All hooks are registered at startup via `init()` functions for security and predictability
2. **Independent Execution**: Each hook runs in isolation to ensure reliability and prevent cascading failures
3. **No Dynamic Loading**: Prevents security risks and ensures predictable behavior
4. **Simple Lifecycle**: Create → Execute → Cleanup

## Key Components

- **CLI Layer** (`internal/cmd/`): urfave/cli v3 command implementations
- **Registry** (`internal/core/registry.go`): Static hook registration and management
- **Hooks** (`internal/hooks/`): Concrete hook implementations
- **Settings** (`internal/config/`): Configuration management with hierarchical precedence
- **Core** (`internal/core/`): Event handling and execution

## Configuration Hierarchy

1. **Project Settings**: `./.claude/settings.json` (takes precedence)
2. **Global Settings**: `~/.claude/settings.json` (fallback)
3. **XDG Config**: `~/.config/blues-traveler/` (application-specific settings)

## Related Documentation

- [Developer Guide](../developer_guide.md) - How to extend Blues Traveler
- [Code Reviews](../reviews/code-review-2024.md) - Code review findings and recommendations
- [Quick Start](../quick_start.md) - Getting started guide
