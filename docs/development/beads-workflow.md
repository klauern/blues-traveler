# Beads Issue Tracking Workflow

This project uses [beads](https://github.com/beads-marketplace/beads) for issue tracking instead of markdown-based backlog files.

## Why Beads?

- **Structured data**: Issues stored in a database with proper relationships
- **Dependencies**: Track blocking relationships between issues
- **Status tracking**: Open, in progress, blocked, closed
- **Priorities**: Numeric priority levels (1=low, 2=medium, 3=high)
- **Types**: bug, feature, task, epic, chore
- **CLI-first**: Fast command-line interface for developers and AI assistants

## Quick Start

### Installation

Beads is already initialized in this repository (`.beads/` directory).

### Basic Commands

```bash
# List all issues
bd list

# List only open issues
bd list --status open

# Find issues ready to work on (no blockers)
bd ready

# Show detailed information
bd show blues-traveler-1

# Create a new issue
bd create "Add new hook for X" --type feature --priority 2

# Update issue status
bd update blues-traveler-1 --status in_progress

# Close completed issue
bd close blues-traveler-1 "Implemented and tested"

# View statistics
bd stats
```

## Issue Types

| Type | Use For |
|------|---------|
| **feature** | New functionality or capabilities |
| **bug** | Defects or issues to fix |
| **task** | General work items (tests, refactoring, etc.) |
| **epic** | Large features split into smaller issues |
| **chore** | Maintenance tasks (cleanup, docs, etc.) |

## Priority Levels

- **3**: High priority (critical features, security issues)
- **2**: Medium priority (important improvements)
- **1**: Low priority (nice-to-haves, minor improvements)

## Issue Workflow

### 1. Finding Work

```bash
# See all ready-to-work issues
bd ready

# Filter by priority
bd ready --priority 3
```

### 2. Claiming Work

```bash
# Update status to in_progress
bd update blues-traveler-2 --status in_progress
```

### 3. Adding Details

```bash
# Add design notes
bd update blues-traveler-2 --design "Implementation approach: ..."

# Add acceptance criteria
bd update blues-traveler-2 --acceptance "- Test coverage > 80%\n- Documentation updated"

# Add general notes
bd update blues-traveler-2 --notes "Found during code review"
```

### 4. Managing Dependencies

```bash
# Make issue 2 block issue 3
bd dep blues-traveler-2 blues-traveler-3

# See blocked issues
bd blocked

# See what an issue depends on
bd show blues-traveler-3
```

### 5. Completing Work

```bash
# Close the issue
bd close blues-traveler-2 "Implemented with tests and docs"

# Reopen if needed
bd reopen blues-traveler-2 "Found regression"
```

## Best Practices

### For Developers

1. **Check ready issues first**: Use `bd ready` to find unblocked work
2. **Update status**: Keep status current (open → in_progress → closed)
3. **Add acceptance criteria**: Define "done" before starting work
4. **Track dependencies**: Use `bd dep` to show blocking relationships
5. **Close completed work**: Use `bd close` with a brief summary
6. **Create issues for discovered work**: Found a bug while working? Create an issue!

### For AI Assistants

When assisting with code:

1. **Check open issues**: `bd list --status open` to see what needs work
2. **Understand requirements**: `bd show <id>` for details and acceptance criteria
3. **Update status**: Mark issues `in_progress` when working on them
4. **Close when complete**: Use `bd close <id>` after implementing and testing
5. **Create new issues**: For bugs found or improvements identified
6. **Link related work**: Use `bd dep` to connect related issues

**Do NOT**:
- Create backlog.md or similar markdown files
- Track TODOs in comments (create issues instead)
- Mix issue tracking with documentation

## Issue Metadata

Issues support rich metadata:

- **ID**: Unique identifier (e.g., `blues-traveler-1`)
- **Title**: Short description
- **Description**: Detailed explanation
- **Type**: bug, feature, task, epic, chore
- **Priority**: 1 (low), 2 (medium), 3 (high)
- **Status**: open, in_progress, blocked, closed
- **Design**: Implementation notes and approach
- **Acceptance**: Criteria for completion
- **Notes**: General observations and context
- **External ref**: Link to external resources (PRs, docs, etc.)
- **Assignee**: Who's working on it
- **Labels**: Tags for categorization
- **Dependencies**: Issues that block or are blocked by this one

## Querying Issues

### By Status

```bash
bd list --status open
bd list --status in_progress
bd list --status blocked
bd list --status closed
```

### By Type

```bash
bd list --type feature
bd list --type bug
bd list --type task
```

### By Priority

```bash
bd list --priority 3  # High priority only
bd ready --priority 2 # Medium priority ready issues
```

### By Assignee

```bash
bd list --assignee alice
bd ready --assignee alice
```

## Integration with Development Workflow

### During Code Review

Extract action items into issues:

```bash
bd create "Improve error messages" \
  --type chore \
  --priority 1 \
  --description "Error messages need more context per code review"
```

### During Feature Development

Track sub-tasks:

```bash
# Create epic
bd create "Add performance hook" --type epic --priority 2

# Create sub-tasks
bd create "Implement performance.go" --type task --priority 2
bd create "Add performance tests" --type task --priority 2
bd create "Document performance hook" --type task --priority 1

# Link them
bd dep blues-traveler-8 blues-traveler-9  # impl blocks tests
bd dep blues-traveler-9 blues-traveler-10 # tests block docs
```

### During Bug Fixing

Track bugs with reproduction steps:

```bash
bd create "Format hook fails on large files" \
  --type bug \
  --priority 3 \
  --description "Format hook times out on files > 10MB" \
  --acceptance "- Handle files up to 100MB\n- Timeout after 30s\n- Clear error message"
```

## Viewing Statistics

```bash
bd stats
```

Shows:
- Total issues
- Open/closed/blocked counts
- Average lead time
- Issues in progress

## Advanced Usage

### Custom Queries

The beads database is SQLite in `.beads/blues-traveler.db`. You can query it directly:

```bash
sqlite3 .beads/blues-traveler.db "SELECT id, title, priority FROM issues WHERE status='open' ORDER BY priority DESC"
```

### Bulk Operations

```bash
# Close multiple issues
for id in blues-traveler-5 blues-traveler-6; do
  bd close "$id" "Completed in bulk"
done

# Update priority for related issues
bd update blues-traveler-7 --priority 3
bd update blues-traveler-8 --priority 3
```

## Migration from Backlog.md

If you have existing backlog items:

1. Extract each TODO/action item
2. Create a beads issue with appropriate type and priority
3. Add acceptance criteria and design notes
4. Delete the backlog.md file
5. Update documentation to reference beads

## Troubleshooting

### Beads Not Initialized

```bash
bd init --prefix blues-traveler
```

### View Issue Details

```bash
bd show <issue-id>
```

### Debug Information

```bash
bd debug-env
```

## Further Reading

- [Beads Documentation](https://github.com/beads-marketplace/beads)
- [Issue Types Best Practices](https://github.com/beads-marketplace/beads/docs/types.md)
- [Dependency Management](https://github.com/beads-marketplace/beads/docs/dependencies.md)

## Summary

Use beads for **all** issue tracking in this project:

✅ **Do**:
- Create issues with `bd create`
- Track progress with `bd update`
- Find work with `bd ready`
- Link issues with `bd dep`
- Close completed work with `bd close`

❌ **Don't**:
- Create backlog.md files
- Use TODO comments for tracking (create issues instead)
- Track issues in documentation files
