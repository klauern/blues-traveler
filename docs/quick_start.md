# Quick Start Guide

Get up and running with Blues Traveler in minutes.

## Prerequisites

- Go 1.25.0 or later
- Claude Code installed and configured

## Installation

### Option 1: Install from Source (Recommended)

```bash
go install github.com/klauern/blues-traveler@latest
```

### Option 2: Build from Source

```bash
git clone https://github.com/klauern/blues-traveler.git
cd blues-traveler
task build
```

## First Steps

### Use Custom Hooks (Recommended)

Define project-specific automation with custom hooks, then sync them into Claude Code settings:

```yaml
# ~/.config/blues-traveler/projects/my-project.yml
my-project:
  PreToolUse:
    jobs:
      - name: security-check
        run: |
          if echo "$TOOL_ARGS" | grep -E "(rm -rf|sudo|curl.*\\|.*sh)"; then
            echo "Dangerous command detected"; exit 1; fi
        only: ${TOOL_NAME} == "Bash"
  PostToolUse:
    jobs:
      - name: format-go
        run: gofmt -w ${TOOL_OUTPUT_FILE}
        only: ${TOOL_NAME} == "Edit" || ${TOOL_NAME} == "Write"
        glob: ["*.go"]
```

Sync and test:

```bash
blues-traveler hooks custom validate
blues-traveler hooks custom sync
blues-traveler hooks run config:my-project:format-go
```

Built-in hooks remain available for quick setup, but custom hooks can replace or extend them to fit your workflow.

### 1. Verify Installation

```bash
blues-traveler hooks list
```

You should see available hooks like:

- `security` - Blocks dangerous commands
- `format` - Auto-formats code
- `debug` - Logs tool usage
- `audit` - JSON audit logging
- `vet` - Code quality checks

### 2. Install Your First Hook

Start with the security hook to protect against dangerous commands:

```bash
blues-traveler hooks install security --event PreToolUse
```

This will:

- Add the security hook to your project settings
- Configure it to run before any tool execution
- Block dangerous commands like `rm -rf /` or `sudo`

### 3. Test the Hook

Run the security hook manually to test:

```bash
blues-traveler hooks run security --log
```

## Common Configurations

### Security + Format Pipeline

Set up a complete code quality pipeline:

```bash
# Security: Block dangerous commands
blues-traveler hooks install security --event PreToolUse

# Format: Auto-format code after editing
blues-traveler hooks install format --event PostToolUse --matcher "Edit,Write"

# Debug: Log all operations
blues-traveler hooks install debug --event PreToolUse --log --log-format pretty
```

### Production Monitoring

For production environments, add comprehensive logging:

```bash
# Audit all operations globally
blues-traveler hooks install audit --event PreToolUse --global
blues-traveler hooks install audit --event PostToolUse --global
```

## Verify Configuration

Check what's currently installed:

```bash
blues-traveler hooks list --installed
```

This shows all hooks configured in your Claude Code settings.

## Next Steps

- **Explore Events**: Run `blues-traveler hooks list --events` to see all available events
- **Customize Settings**: Edit `.claude/settings.json` for advanced configuration
- **Add More Hooks**: Install additional hooks based on your needs
- **Check Logs**: Use `--log` flag when running hooks to see detailed output

## Troubleshooting

### Hook Not Found

```bash
# Check available hooks
blues-traveler hooks list

# Verify installation
which blues-traveler
```

### Hook Not Working

```bash
# Check if hook is enabled
blues-traveler hooks list --installed

# Run with logging
blues-traveler hooks run <hook-name> --log
```

### Settings Issues

```bash
# Check project settings
cat ./.claude/settings.json

# Check global settings
cat ~/.claude/settings.json
```

## What Happens Next?

Once configured, Blues Traveler will automatically:

1. **Intercept Commands**: Security hooks run before tool execution
2. **Process Results**: Format and vet hooks run after code changes
3. **Log Operations**: Debug and audit hooks record all activity
4. **Maintain Quality**: Ensure code meets your standards

Your Claude Code experience is now enhanced with security, quality, and monitoring!
