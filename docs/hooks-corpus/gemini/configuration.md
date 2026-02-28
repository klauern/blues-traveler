# Gemini - Configuration & Setup

**Last Updated:** 2026-02-21
**Version:** Gemini CLI v0.26.0+

---

## Overview

Gemini's automation capabilities require configuration across multiple product areas:

1. **Gemini CLI Hooks** - Configure via `settings.json`
2. **Gemini Code Assist** - IDE extension settings
3. **Gemini API** - API keys and SDK configuration

This document focuses primarily on **Gemini CLI Hooks configuration**.

**Sources:**
- [Gemini CLI Hooks Documentation](https://geminicli.com/docs/hooks/)
- [Gemini Code Assist Setup](https://developers.google.com/gemini-code-assist/docs/set-up-gemini)
- [Gemini API Quickstart](https://ai.google.dev/gemini-api/docs/get-started)

---

## Configuration File Locations

### Gemini CLI

**Project Configuration (Highest Priority):**
```
.gemini/settings.json
```
- Located in project root
- Project-specific hooks
- Version controlled (recommended)
- Overrides user and extension settings

**User Configuration:**
```
~/.gemini/settings.json
```
- User's home directory
- Global hooks across all projects
- Not version controlled

**Extension Configuration (Lowest Priority):**
```
~/.gemini/extensions/{extension-name}/settings.json
```
- Provided by installed extensions
- Shared hooks from extension packages
- Automatically loaded

**Precedence Order:**
1. **Project settings** (`.gemini/settings.json`) - Highest
2. **User settings** (`~/.gemini/settings.json`)
3. **Extension settings** - Lowest

**Merge Behavior:**
- Hooks with same event type are **combined** (not replaced)
- All matching hooks execute in order: project → user → extension
- Field-level merging for non-hook settings

---

## Configuration Format

### Basic Structure

```json
{
  "hooks": {
    "SessionStart": [...],
    "BeforeAgent": [...],
    "BeforeToolSelection": [...],
    "BeforeTool": [...],
    "AfterTool": [...],
    "BeforeModel": [...],
    "AfterModel": [...],
    "BeforeResponse": [...],
    "AfterChunk": [...],
    "AfterAgent": [...],
    "SessionEnd": [...],
    "Notification": [...],
    "BeforeCompress": [...],
    "AfterAlert": [...]
  },
  "options": {
    "logLevel": "info",
    "timeout": 5000
  }
}
```

### Complete Schema

```json
{
  "hooks": {
    "{EventType}": [
      {
        "name": "string (optional - for identification)",
        "matcher": "string or regex (event-specific)",
        "enabled": true,
        "timeout": 5000,
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["./hooks/script.sh"],
            "cwd": "${GEMINI_PROJECT_DIR}",
            "env": {
              "CUSTOM_VAR": "value"
            }
          }
        ]
      }
    ]
  },
  "options": {
    "logLevel": "debug" | "info" | "warn" | "error",
    "timeout": 5000,
    "concurrency": 1
  }
}
```

---

## Schema Reference

### Hook Configuration Object

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | ❌ | (none) | Human-readable identifier for this hook group |
| `matcher` | string | ✅ | N/A | Pattern to match events (regex for tools, exact for lifecycle) |
| `enabled` | boolean | ❌ | `true` | Enable/disable this hook |
| `timeout` | number | ❌ | 5000 | Timeout in milliseconds |
| `hooks` | array | ✅ | N/A | Array of hook definitions (command or plugin) |

### Command Hook Definition

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | ✅ | N/A | Must be `"command"` |
| `command` | string | ✅ | N/A | Executable to run (e.g., `"bash"`, `"python3"`, `"node"`) |
| `args` | array | ✅ | `[]` | Arguments to pass to command |
| `cwd` | string | ❌ | CLI's cwd | Working directory for command |
| `env` | object | ❌ | `{}` | Environment variables to set |

### Plugin Hook Definition

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | ✅ | N/A | Must be `"plugin"` |
| `package` | string | ✅ | N/A | npm package name (with `geminicli-plugin` keyword) |
| `options` | object | ❌ | `{}` | Plugin-specific options |

### Global Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `logLevel` | string | `"info"` | Logging verbosity: `"debug"`, `"info"`, `"warn"`, `"error"` |
| `timeout` | number | 5000 | Default timeout for all hooks (milliseconds) |
| `concurrency` | number | 1 | Hook concurrency (currently only 1 supported) |

---

## Configuration Examples

### Minimal Configuration

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "write_file",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["./hooks/validate-write.sh"]
          }
        ]
      }
    ]
  }
}
```

### Complete Configuration

```json
{
  "hooks": {
    "SessionStart": [
      {
        "name": "load-project-context",
        "matcher": "startup",
        "enabled": true,
        "timeout": 2000,
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["${GEMINI_PROJECT_DIR}/.gemini/hooks/load-context.sh"],
            "cwd": "${GEMINI_PROJECT_DIR}",
            "env": {
              "PROJECT_NAME": "my-project"
            }
          }
        ]
      }
    ],
    "BeforeAgent": [
      {
        "name": "inject-git-context",
        "matcher": "*",
        "enabled": true,
        "hooks": [
          {
            "type": "command",
            "command": "python3",
            "args": [".gemini/hooks/git-context.py"]
          }
        ]
      }
    ],
    "BeforeToolSelection": [
      {
        "name": "optimize-tool-selection",
        "matcher": "*",
        "enabled": true,
        "hooks": [
          {
            "type": "plugin",
            "package": "gemini-tool-optimizer",
            "options": {
              "costLimit": 0.01
            }
          }
        ]
      }
    ],
    "BeforeTool": [
      {
        "name": "security-validation",
        "matcher": "write_.*|execute_.*|delete_.*",
        "enabled": true,
        "timeout": 3000,
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": [".gemini/hooks/security-check.sh"]
          },
          {
            "type": "command",
            "command": "node",
            "args": [".gemini/hooks/validate-params.js"]
          }
        ]
      }
    ],
    "AfterTool": [
      {
        "name": "log-operations",
        "matcher": "*",
        "enabled": true,
        "hooks": [
          {
            "type": "command",
            "command": "python3",
            "args": [".gemini/hooks/audit-log.py"]
          }
        ]
      }
    ],
    "AfterAgent": [
      {
        "name": "send-notification",
        "matcher": "*",
        "enabled": true,
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": [".gemini/hooks/notify-completion.sh"]
          }
        ]
      }
    ]
  },
  "options": {
    "logLevel": "debug",
    "timeout": 5000
  }
}
```

### Multiple Hooks Per Event

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "write-file-security",
        "matcher": "write_file",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["./hooks/check-sensitive-files.sh"]
          }
        ]
      },
      {
        "name": "execute-command-security",
        "matcher": "execute_command",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["./hooks/check-dangerous-commands.sh"]
          }
        ]
      },
      {
        "name": "all-tools-audit",
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "python3",
            "args": ["./hooks/audit-all.py"]
          }
        ]
      }
    ]
  }
}
```

---

## Matcher Patterns

### Tool Event Matchers (Regex)

**Events:** `BeforeTool`, `AfterTool`

**Examples:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "write_file",  // Exact match
        "hooks": [...]
      },
      {
        "matcher": "write_.*",  // Starts with "write_"
        "hooks": [...]
      },
      {
        "matcher": "(read|write)_file",  // read_file OR write_file
        "hooks": [...]
      },
      {
        "matcher": "execute_.*|run_.*",  // Starts with execute_ OR run_
        "hooks": [...]
      },
      {
        "matcher": "*",  // All tools
        "hooks": [...]
      },
      {
        "matcher": "",  // All tools (empty string)
        "hooks": [...]
      }
    ]
  }
}
```

**Common Patterns:**
- `write_.*` - All write operations
- `read_.*` - All read operations
- `execute_.*` - All execute operations
- `.*_file` - All file operations
- `(create|update|delete)_.*` - CRUD operations
- `*` or `""` - All tools

### Lifecycle Event Matchers (Exact String)

**Events:** `SessionStart`, `SessionEnd`, `Notification`

**Examples:**
```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",  // Exact match
        "hooks": [...]
      }
    ],
    "Notification": [
      {
        "matcher": "idle",  // Specific notification type
        "hooks": [...]
      },
      {
        "matcher": "*",  // All notifications
        "hooks": [...]
      }
    ]
  }
}
```

---

## Environment Variables

### System-Provided Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `GEMINI_PROJECT_DIR` | string | Project root directory | `/path/to/project` |
| `GEMINI_USER_DIR` | string | User's home directory | `/home/user` |
| `GEMINI_SESSION_ID` | string | Current session ID | `abc123...` |
| `GEMINI_EVENT_TYPE` | string | Current event type | `BeforeTool` |

**Usage in Configuration:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["${GEMINI_PROJECT_DIR}/.gemini/hooks/check.sh"],
            "cwd": "${GEMINI_PROJECT_DIR}"
          }
        ]
      }
    ]
  }
}
```

### Custom Environment Variables

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["./hooks/notify.sh"],
            "env": {
              "SLACK_WEBHOOK_URL": "https://hooks.slack.com/...",
              "NOTIFICATION_LEVEL": "high",
              "PROJECT_NAME": "my-app"
            }
          }
        ]
      }
    ]
  }
}
```

---

## Setup Instructions

### Initial Setup

**1. Install Gemini CLI (v0.26.0+):**
```bash
# Installation method depends on platform
# Check official docs for latest instructions
npm install -g @google/gemini-cli

# Or via Homebrew (macOS)
brew install gemini-cli
```

**2. Verify Installation:**
```bash
gemini --version
# Should show v0.26.0 or higher
```

**3. Create Project Configuration:**
```bash
cd /path/to/your/project

# Create .gemini directory
mkdir -p .gemini/hooks

# Create settings.json
cat > .gemini/settings.json << 'EOF'
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": [".gemini/hooks/load-context.sh"]
          }
        ]
      }
    ]
  }
}
EOF
```

**4. Create Your First Hook:**
```bash
cat > .gemini/hooks/load-context.sh << 'EOF'
#!/bin/bash
# Load git context at session start

GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
GIT_STATUS=$(git status --short 2>/dev/null || echo "Not a git repo")

CONTEXT="Git Branch: $GIT_BRANCH\n\nStatus:\n$GIT_STATUS"

echo "{\"decision\": \"continue\", \"systemMessage\": \"$CONTEXT\"}"
EOF

chmod +x .gemini/hooks/load-context.sh
```

**5. Test the Hook:**
```bash
# Start Gemini CLI
gemini

# Hook should fire on SessionStart
# You should see git context injected
```

---

### Verification Commands

**View Hook Configuration:**
```bash
# Open hooks management panel
gemini
> /hooks panel
```

**Enable Debug Logging:**
```bash
# Edit settings.json
{
  "options": {
    "logLevel": "debug"
  }
}

# Restart Gemini CLI
# You'll see hook execution details in logs
```

**Test Hook Manually:**
```bash
# Create test input
echo '{"event": "SessionStart", "sessionId": "test-123"}' | \
  .gemini/hooks/load-context.sh

# Should output valid JSON decision
```

---

## Configuration Management

### Version Control

**Recommended `.gitignore`:**
```gitignore
# Ignore user-specific hooks
.gemini/local-hooks/

# Ignore hook logs
.gemini/*.log
.gemini/audit-log.jsonl

# Ignore sensitive configs (if any)
.gemini/secrets.json
```

**Commit to Git:**
```bash
git add .gemini/settings.json
git add .gemini/hooks/
git commit -m "Add Gemini CLI hooks configuration"
```

### Sharing Across Team

**1. Document Required Setup:**
```markdown
# README.md

## Gemini CLI Setup

1. Install Gemini CLI v0.26.0+
2. Run `npm install` (if using plugin hooks)
3. Copy `.gemini/settings.example.json` to `.gemini/settings.json`
4. Update environment variables in hooks as needed
5. Make hooks executable: `chmod +x .gemini/hooks/*.sh`
```

**2. Use Extensions for Team Hooks:**
```bash
# Create team extension
mkdir -p team-gemini-hooks
cd team-gemini-hooks

# package.json
{
  "name": "@mycompany/gemini-hooks",
  "version": "1.0.0",
  "keywords": ["geminicli-plugin"],
  "geminicli": {
    "apiVersion": "1.0"
  }
}

# Publish to private npm registry
npm publish --registry=https://npm.mycompany.com
```

**3. Install Team Extension:**
```bash
gemini extension install @mycompany/gemini-hooks
```

---

## Hooks Management Commands

### CLI Commands

**View Hooks Panel:**
```bash
gemini
> /hooks panel
```

**Enable Specific Hook:**
```bash
gemini
> /hooks enable security-validation
```

**Disable Specific Hook:**
```bash
gemini
> /hooks disable security-validation
```

**Migrate Hooks Configuration:**
```bash
gemini
> /hooks migrate
```

---

## Troubleshooting

### Common Issues

#### Hook Not Executing

**Symptoms:** Hook doesn't fire when expected

**Diagnosis:**
1. Check matcher pattern matches event
2. Verify hook is enabled (`"enabled": true`)
3. Check timeout isn't too short
4. Enable debug logging

**Solution:**
```bash
# Enable debug logging
{
  "options": {
    "logLevel": "debug"
  }
}

# Check logs for hook execution details
```

#### Invalid JSON Output

**Symptoms:** Hook executes but CLI treats as "allow"

**Diagnosis:**
```bash
# Test hook manually
echo '{"event": "BeforeTool", "tool": {"name": "write_file"}}' | \
  .gemini/hooks/your-hook.sh

# Should output valid JSON
```

**Solution:**
```bash
# Ensure valid JSON output
echo '{"decision": "allow"}'  # ✅ Valid
echo "decision: allow"         # ❌ Invalid
```

#### Timeout Errors

**Symptoms:** Hook killed before completion

**Solution:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "*",
        "timeout": 10000,  // Increase timeout
        "hooks": [...]
      }
    ]
  }
}
```

#### Permission Denied

**Symptoms:** "Permission denied" error

**Solution:**
```bash
# Make hooks executable
chmod +x .gemini/hooks/*.sh
chmod +x .gemini/hooks/*.py
chmod +x .gemini/hooks/*.js
```

---

## Best Practices

### Configuration Organization

**1. Separate Hooks by Responsibility:**
```
.gemini/
├── settings.json
└── hooks/
    ├── security/
    │   ├── validate-write.sh
    │   └── check-dangerous.sh
    ├── context/
    │   ├── load-git.sh
    │   └── inject-jira.py
    └── logging/
        └── audit-log.py
```

**2. Use Descriptive Names:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "prevent-sensitive-file-writes",  // ✅ Clear
        "matcher": "write_file",
        "hooks": [...]
      }
    ]
  }
}
```

**3. Document Configuration:**
```json
{
  "// Purpose": "Security validation hooks for production",
  "// Last Updated": "2026-02-21",
  "// Maintainer": "security-team@company.com",
  "hooks": {
    "BeforeTool": [...]
  }
}
```

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Hook event catalog
- [Architecture & Internals](./architecture.md) - Execution model
- [Scripting & Execution](./scripting.md) - Writing hooks
- [Security & Safety](./security.md) - Security patterns
- [Examples](./examples.md) - Ready-to-use configurations

---

**Document Version:** 1.0
**Last Updated:** 2026-02-21
**Sources:** Official documentation, research analysis
