# OpenAI Codex - Configuration & Setup

**Last Updated:** 2026-02-21

---

## Overview

Codex configuration occurs at **three levels**:
1. **OpenAI Dashboard** - API keys, webhooks, billing
2. **config.toml** - Global Codex CLI/App settings
3. **Project Files** - Repository-specific configuration

---

## Configuration File Locations

### Global Configuration

**Location:** `~/.codex/config.toml`
**Scope:** All Codex CLI sessions
**Format:** TOML

```toml
# ~/.codex/config.toml
[model]
name = "gpt-5.2-codex"
reasoning_tier = "medium"  # low, medium, high, xhigh

[hooks.file.after_write]
command = "/usr/local/bin/format-code"

[notify]
command = "/usr/local/bin/notify-slack"
events = ["turn_complete", "error"]
```

### Project Configuration

**Location:** `<project-root>/codex.json` or `.codex/config.toml`
**Scope:** Current project only
**Format:** JSON or TOML

```json
{
  "model": "gpt-5.3-codex",
  "instructions_file": "AGENTS.md",
  "skills": [
    "code-review",
    "jira-integration"
  ],
  "mcp_servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    }
  }
}
```

### Project Instructions

**Location:** `<project-root>/AGENTS.md` or `.claude/CODEX.md`
**Purpose:** Repository-specific instructions for Codex
**Format:** Markdown

```markdown
# Project: MyApp

## Context
This is a TypeScript API using Express and PostgreSQL.

## Code Style
- Use functional components
- Prefer async/await over promises
- ESLint must pass before commits

## Testing
- Write unit tests for all business logic
- Integration tests for API endpoints
- Use Jest + Supertest

## Deployment
- Deploy to AWS via GitHub Actions
- Never merge without CI passing
```

---

## Configuration Precedence

```
Project codex.json
    ↓ (overrides)
Project .codex/config.toml
    ↓ (overrides)
Global ~/.codex/config.toml
    ↓ (overrides)
Default values
```

---

## Complete Configuration Schema

### Model Configuration

```toml
[model]
name = "gpt-5.2-codex"  # or "gpt-5.3-codex"
reasoning_tier = "medium"  # low, medium, high, xhigh (for 5.2)
temperature = 0.7  # 0.0-1.0
max_tokens = 4096  # Response length limit
```

### Hooks Configuration

```toml
# Tool hooks
[hooks.tool.before]
command = "/path/to/pre-tool-hook"
pattern = "Bash|Write|Edit"  # Regex pattern
timeout = 30  # seconds

[hooks.tool.after]
command = "/path/to/post-tool-hook"

# File hooks
[hooks.file.before_write]
command = "/path/to/validate"
args = ["--strict"]

[hooks.file.after_write]
command = "/path/to/format"
pattern = "\\.(ts|js|py)$"  # Regex for file extensions

# Event hooks
[hooks.event.stop]
command = "/path/to/on-complete"

[hooks.event.notification]
command = "/path/to/notify"
events = ["turn_complete", "approval_required", "error"]
```

### MCP Servers

```toml
[mcp_servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
env = { GITHUB_TOKEN = "${GITHUB_TOKEN}" }

[mcp_servers.context7]
command = "npx"
args = ["-y", "@upstash/context7-mcp"]

[mcp_servers.custom]
command = "/usr/local/bin/my-mcp-server"
args = ["--port", "8080"]
```

### Notification Configuration

```toml
[notify]
command = "/usr/local/bin/send-notification"
args = ["--webhook", "https://hooks.slack.com/services/xxx"]
events = [
    "turn_complete",
    "approval_required",
    "error",
    "background_start",
    "background_complete"
]
```

### Skills Configuration

```toml
[skills]
paths = [
    "~/.codex/skills",
    "./.codex/skills",
    "/usr/local/share/codex-skills"
]

enabled = [
    "code-review",
    "jira-integration",
    "security-scan"
]
```

---

## Setup Instructions

### Initial Setup

```bash
# Install Codex CLI
curl -fsSL https://developers.openai.com/install.sh | sh

# Authenticate
codex auth login

# Initialize project
cd /path/to/project
codex init

# Creates:
# - .codex/config.toml (project config)
# - AGENTS.md (project instructions)
```

### Configure Webhooks (OpenAI Dashboard)

```
1. Visit: https://platform.openai.com/settings/webhooks
2. Click "Add Webhook"
3. Enter:
   - URL: https://your-domain.com/webhooks/codex
   - Events: background_completion, batch.completed
4. Save (secret generated automatically)
5. Copy secret to environment:
   export OPENAI_WEBHOOK_SECRET="wh_secret_xxx"
```

### Configure Hooks

```bash
# Create hook script
cat > ~/.local/bin/format-code << 'EOF'
#!/bin/bash
FILE=$1

case "$FILE" in
  *.py)  black "$FILE" ;;
  *.js)  prettier --write "$FILE" ;;
  *.ts)  prettier --write "$FILE" ;;
esac
EOF

chmod +x ~/.local/bin/format-code

# Add to config
cat >> ~/.codex/config.toml << 'EOF'
[hooks.file.after_write]
command = "/Users/username/.local/bin/format-code"
EOF
```

### Configure MCP Servers

```bash
# Add GitHub MCP server
codex mcp add github -- npx -y @modelcontextprotocol/server-github

# Verify
codex mcp list
# Output:
# - github: npx -y @modelcontextprotocol/server-github
```

---

## Environment Variables

### OpenAI API

| Variable | Purpose | Required |
|----------|---------|----------|
| `OPENAI_API_KEY` | API authentication | ✅ |
| `OPENAI_WEBHOOK_SECRET` | Webhook signature verification | ❌ (if using webhooks) |
| `OPENAI_ORG_ID` | Organization ID | ❌ |

### Codex CLI

| Variable | Purpose | Default |
|----------|---------|---------|
| `CODEX_CONFIG_DIR` | Config directory | `~/.codex` |
| `CODEX_MODEL` | Default model | `gpt-5.2-codex` |
| `CODEX_DEBUG` | Enable debug logging | `false` |

### Custom Integration

```bash
# Set environment variables for MCP servers
export GITHUB_TOKEN="ghp_xxx"
export JIRA_API_TOKEN="xxx"
export SLACK_WEBHOOK="https://hooks.slack.com/services/xxx"
```

---

## Configuration Merging

### How Configs Merge

```toml
# Global ~/.codex/config.toml
[model]
name = "gpt-5.2-codex"
reasoning_tier = "medium"

[hooks.file.after_write]
command = "/usr/local/bin/global-format"

# Project .codex/config.toml
[model]
reasoning_tier = "high"  # Overrides global

[hooks.file.before_write]  # Adds to global hooks
command = "/usr/local/bin/project-validate"

# Result: Both hooks run
# before_write: project-validate (project)
# after_write: global-format (global)
# reasoning_tier: high (project overrides global)
```

### Array Handling

- **Lists append:** MCP servers, skills, hook events
- **Objects merge:** Model config, environment variables
- **Primitives override:** Strings, numbers, booleans

---

## Verification

### Test Configuration

```bash
# Verify config syntax
codex config validate

# Show effective config (merged)
codex config show

# Test hooks
codex config test-hooks

# Verify MCP servers
codex mcp list
codex mcp test github
```

### Debug Mode

```bash
# Enable debug logging
export CODEX_DEBUG=true
codex

# Or via config
[debug]
enabled = true
log_file = "~/.codex/debug.log"
log_level = "debug"  # error, warn, info, debug, trace
```

---

## Common Configuration Patterns

### Pattern: Auto-Format on Write

```toml
[hooks.file.after_write]
command = "/usr/local/bin/format-and-lint"
pattern = "\\.(ts|js|py|go)$"
timeout = 60
```

### Pattern: Secret Detection

```toml
[hooks.file.before_write]
command = "/usr/local/bin/detect-secrets"
args = ["--strict"]
```

### Pattern: Slack Notifications

```toml
[notify]
command = "/usr/local/bin/notify-slack"
args = ["--webhook", "${SLACK_WEBHOOK}"]
events = ["turn_complete", "error"]
```

### Pattern: GitHub Integration

```toml
[mcp_servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
env = { GITHUB_TOKEN = "${GITHUB_TOKEN}" }
```

---

## Troubleshooting

### Config Not Loading

```bash
# Check config file location
echo $CODEX_CONFIG_DIR  # Should be ~/.codex

# Verify file exists
ls -la ~/.codex/config.toml

# Check syntax
codex config validate
```

### Hooks Not Firing

```bash
# Test hook script directly
/path/to/hook script arg1 arg2

# Check hook timeout
codex config show | grep timeout

# Enable debug mode
export CODEX_DEBUG=true
```

### MCP Server Errors

```bash
# Test MCP server manually
npx -y @modelcontextprotocol/server-github

# Check logs
codex mcp logs github

# Reinstall
codex mcp remove github
codex mcp add github -- npx -y @modelcontextprotocol/server-github
```

---

## Sources

- [Codex Config Reference](https://developers.openai.com/codex/config-reference/)
- [Hooks Documentation](https://developers.openai.com/codex/config-reference/#hooks)
- [MCP Configuration](https://developers.openai.com/codex/mcp/)
- Research: [API Capabilities](../../../research-notes/codex/api-capabilities.md)

---

**Last Updated:** 2026-02-21
