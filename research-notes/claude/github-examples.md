# Claude Code Hooks - GitHub Examples & Real-World Usage

**Research Date:** 2026-02-21
**Sources:** GitHub repositories and community projects
**Status:** ✅ Complete

---

## Overview

This document catalogs real-world Claude Code hook examples found in public GitHub repositories, providing practical implementations and patterns from the community.

---

## Major Hook Repositories

### 1. disler/claude-code-hooks-mastery

**URL:** https://github.com/disler/claude-code-hooks-mastery
**Description:** Complete hook lifecycle coverage with all 13+ hook events
**Hackathon Winner:** Anthropic Claude Code Hackathon participant

**Features:**
- Intelligent TTS (Text-to-Speech) system
- Security enhancements
- Automatic logging
- Complete event coverage

**Key Implementations:**
- All 17 hook events implemented
- Production-ready examples
- Advanced patterns (TTS integration)

**Source:** [GitHub - disler/claude-code-hooks-mastery](https://github.com/disler/claude-code-hooks-mastery)

---

### 2. affaan-m/everything-claude-code

**URL:** https://github.com/affaan-m/everything-claude-code
**Status:** Anthropic Hackathon Winner (Feb 2026)
**Description:** Complete Claude Code configuration collection

**What's Included:**
- Production-ready agents
- Skills library
- Comprehensive hooks
- Custom commands
- Rules configuration
- MCP (Model Context Protocol) setups

**Battle-Tested:** 10+ months of intensive daily use

**Notable Files:**
- `hooks/hooks.json` - Complete hook configurations
- Production patterns evolved over months

**Source:** [GitHub - affaan-m/everything-claude-code](https://github.com/affaan-m/everything-claude-code)

---

### 3. karanb192/claude-code-hooks

**URL:** https://github.com/karanb192/claude-code-hooks
**Description:** Copy-paste ready hook collection

**Categories:**
- Safety hooks
- Automation hooks
- Notification hooks
- Custom workflows

**Philosophy:** Growing collection of practical, ready-to-use hooks

**Source:** [GitHub - karanb192/claude-code-hooks](https://github.com/karanb192/claude-code-hooks)

---

### 4. ChrisWiles/claude-code-showcase

**URL:** https://github.com/ChrisWiles/claude-code-showcase
**Description:** Comprehensive project configuration example

**Features:**
- Auto-format code hooks
- Test-on-change hooks
- TypeScript type-checking
- Branch protection (block edits on main)
- GitHub Actions workflows

**Use Case:** Reference implementation for complete project setup

**Source:** [GitHub - ChrisWiles/claude-code-showcase](https://github.com/ChrisWiles/claude-code-showcase)

---

### 5. johnlindquist/claude-hooks

**URL:** https://github.com/johnlindquist/claude-hooks
**Description:** TypeScript-powered hook system

**Unique Features:**
- Full TypeScript support
- Type safety for hooks
- Auto-completion
- Strongly-typed payloads

**Philosophy:** Bring IDE intelligence to hook development

**Source:** [GitHub - johnlindquist/claude-hooks](https://github.com/johnlindquist/claude-hooks)

---

### 6. decider/claude-hooks

**URL:** https://github.com/decider/claude-hooks
**Description:** Python-based lightweight hook system

**Features:**
- Automatic validation
- Quality checks during sessions
- Minimal overhead
- Production-focused

**Language:** Python

**Source:** [GitHub - decider/claude-hooks](https://github.com/decider/claude-hooks)

---

### 7. trailofbits/claude-code-config

**URL:** https://github.com/trailofbits/claude-code-config
**Description:** Opinionated defaults and workflows from Trail of Bits

**Features:**
- Enterprise-grade configurations
- Security-focused patterns
- Documentation and best practices
- CI/CD integration examples

**Source:** [GitHub - trailofbits/claude-code-config](https://github.com/trailofbits/claude-code-config)

---

### 8. hesreallyhim/awesome-claude-code

**URL:** https://github.com/hesreallyhim/awesome-claude-code
**Description:** Curated list of Claude Code resources

**Categories:**
- Skills
- Hooks
- Slash commands
- Agent orchestrators
- Applications
- Plugins

**Use Case:** Discovery and exploration of community resources

**Source:** [GitHub - hesreallyhim/awesome-claude-code](https://github.com/hesreallyhim/awesome-claude-code)

---

### 9. wesammustafa/Claude-Code-Everything-You-Need-to-Know

**URL:** https://github.com/wesammustafa/Claude-Code-Everything-You-Need-to-Know
**Description:** Ultimate all-in-one guide to mastering Claude Code

**Coverage:**
- Setup and configuration
- Prompt engineering
- Commands and hooks
- Workflows and automation
- Integrations
- MCP servers and tools
- BMAD method

**Format:** Step-by-step tutorials, real-world examples, expert strategies

**Source:** [GitHub - wesammustafa/Claude-Code-Everything-You-Need-to-Know](https://github.com/wesammustafa/Claude-Code-Everything-You-Need-to-Know)

---

## Common Hook Patterns (From GitHub)

### Pattern 1: Auto-Format After Edits

**Source:** Multiple repositories

**Configuration:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "prettier --write \"$CLAUDE_PROJECT_DIR\"/**/*.{js,ts,jsx,tsx}"
          }
        ]
      }
    ]
  }
}
```

**Variations:**
- **Python:** `black` + `isort`
- **Go:** `gofmt` or `goimports`
- **Rust:** `cargo fmt`
- **Multiple:** Chain formatters with `&&`

---

### Pattern 2: Block Dangerous Commands

**Source:** disler/claude-code-hooks-mastery, trail of bits

**Bash Script:**
```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

# Dangerous patterns
DANGEROUS="rm -rf|sudo rm|mkfs|dd if=|:(){|curl.*sh"

if echo "$COMMAND" | grep -qE "$DANGEROUS"; then
  echo "Blocked dangerous command: $COMMAND" >&2
  exit 2
fi

exit 0
```

**Configuration:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/block-dangerous.sh"
          }
        ]
      }
    ]
  }
}
```

---

### Pattern 3: Run Tests on File Changes

**Source:** ChrisWiles/claude-code-showcase

**Configuration:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "npm test -- --findRelatedTests $(echo \"$INPUT\" | jq -r '.tool_input.file_path')"
          }
        ]
      }
    ]
  }
}
```

**Script Alternative:**
```bash
#!/bin/bash
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path')

# Only run tests for test files
if [[ "$FILE_PATH" == *test* ]] || [[ "$FILE_PATH" == *spec* ]]; then
  npm test -- "$FILE_PATH"
fi
```

---

### Pattern 4: Notification Integration

**Source:** karanb192/claude-code-hooks

**Slack Notification:**
```bash
#!/bin/bash
INPUT=$(cat)
EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
TOOL=$(echo "$INPUT" | jq -r '.tool_name')

curl -X POST "$SLACK_WEBHOOK_URL" \
  -H 'Content-Type: application/json' \
  -d "{\"text\":\"Claude Code: $EVENT - $TOOL\"}"
```

**macOS Notification:**
```bash
#!/bin/bash
osascript -e 'display notification "Claude Code hook triggered" with title "Hook Event"'
```

---

### Pattern 5: Context Injection (SessionStart)

**Source:** affaan-m/everything-claude-code

**Load TODOs and Git Status:**
```bash
#!/bin/bash

# Output goes to Claude's context
echo "## Current Context"
echo ""
echo "### Git Status"
git status --short
echo ""
echo "### Recent TODOs"
rg "TODO|FIXME" --max-count 10
echo ""
echo "### Recent Commits"
git log --oneline -5
```

---

### Pattern 6: Security Audit Logging

**Source:** trailofbits/claude-code-config

**Audit Logger:**
```bash
#!/bin/bash
INPUT=$(cat)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LOG_FILE="$HOME/.claude/audit.log"

# Append JSON event to audit log
echo "$INPUT" | jq -c ". + {timestamp: \"$TIMESTAMP\"}" >> "$LOG_FILE"

# Always allow (this is just logging)
exit 0
```

---

### Pattern 7: Branch Protection

**Source:** ChrisWiles/claude-code-showcase

**Block Edits on Main:**
```bash
#!/bin/bash
INPUT=$(cat)
BRANCH=$(git branch --show-current)
TOOL=$(echo "$INPUT" | jq -r '.tool_name')

if [[ "$BRANCH" == "main" || "$BRANCH" == "master" ]]; then
  if [[ "$TOOL" == "Edit" || "$TOOL" == "Write" ]]; then
    echo "Cannot edit files on protected branch: $BRANCH" >&2
    echo "Create a feature branch first: git checkout -b feature/your-feature" >&2
    exit 2
  fi
fi

exit 0
```

---

### Pattern 8: Type-Check TypeScript Files

**Source:** ChrisWiles/claude-code-showcase

**TypeScript Hook:**
```bash
#!/bin/bash
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path')

# Only type-check .ts/.tsx files
if [[ "$FILE_PATH" =~ \.(ts|tsx)$ ]]; then
  if ! npx tsc --noEmit "$FILE_PATH"; then
    echo "TypeScript type errors detected" >&2
    exit 2
  fi
fi

exit 0
```

---

### Pattern 9: Prompt-Based Approval

**Source:** Official docs + community adaptations

**LLM-Based Security Check:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "prompt",
            "prompt": "Evaluate if this command is safe to run: $ARGUMENTS. Consider: 1) Does it delete files? 2) Does it modify system settings? 3) Does it run untrusted code? Respond with {\"ok\": true} to allow or {\"ok\": false, \"reason\": \"explanation\"} to block."
          }
        ]
      }
    ]
  }
}
```

---

### Pattern 10: Agent-Based Verification

**Source:** Advanced examples from hackathon projects

**Verify Tests Pass Before Stop:**
```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "agent",
            "prompt": "Before allowing Claude to stop, verify: 1) Run the test suite 2) Check all tests pass 3) Check for linting errors. If anything fails, respond with {\"ok\": false, \"reason\": \"explanation\"}. Context: $ARGUMENTS",
            "timeout": 120
          }
        ]
      }
    ]
  }
}
```

---

## Real-World Configuration Examples

### Example 1: Full-Stack Web App Setup

**Source:** Composite from multiple repos

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/security-check.sh",
            "timeout": 10
          }
        ]
      },
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/branch-protection.sh",
            "timeout": 5
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format-code.sh",
            "timeout": 30
          },
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/type-check.sh",
            "timeout": 60
          },
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/run-tests.sh",
            "async": true,
            "timeout": 300
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/load-context.sh"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "agent",
            "prompt": "Verify all work is complete before stopping. Check: 1) Tests pass 2) No lint errors 3) All TODOs addressed. $ARGUMENTS",
            "timeout": 120
          }
        ]
      }
    ]
  }
}
```

---

### Example 2: Python Data Science Project

**Source:** Community examples + adaptations

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "black \"$(echo \"$INPUT\" | jq -r '.tool_input.file_path')\"",
            "timeout": 30
          },
          {
            "type": "command",
            "command": "isort \"$(echo \"$INPUT\" | jq -r '.tool_input.file_path')\"",
            "timeout": 30
          },
          {
            "type": "command",
            "command": "flake8 \"$(echo \"$INPUT\" | jq -r '.tool_input.file_path')\"",
            "timeout": 30
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/check-venv.sh"
          }
        ]
      }
    ]
  }
}
```

**check-venv.sh:**
```bash
#!/bin/bash
if [[ -z "$VIRTUAL_ENV" ]]; then
  echo "Warning: Not in a virtual environment" >&2
  echo "Activate with: source venv/bin/activate" >&2
fi
exit 0  # Warning, but don't block
```

---

### Example 3: Enterprise Security-First

**Source:** trailofbits/claude-code-config

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "/opt/security/validate-command.sh",
            "timeout": 15
          }
        ]
      },
      {
        "matcher": "WebFetch",
        "hooks": [
          {
            "type": "command",
            "command": "/opt/security/validate-url.sh",
            "timeout": 10
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "/opt/security/audit-log.sh",
            "async": true,
            "timeout": 5
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/opt/security/scan-prompt.sh",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

---

## Community Best Practices

### From GitHub Examples

**1. File Organization**
```
.claude/
├── settings.json          # Main config
├── settings.local.json    # Local overrides (gitignored)
└── hooks/
    ├── security-check.sh
    ├── format-code.sh
    ├── run-tests.sh
    ├── branch-protection.sh
    └── load-context.sh
```

**2. Script Template**
```bash
#!/bin/bash
set -euo pipefail  # Strict error handling

# Read JSON input
INPUT=$(cat)

# Extract fields with jq
EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
TOOL=$(echo "$INPUT" | jq -r '.tool_name // "unknown"')

# Your logic here
echo "Processing $EVENT for $TOOL" >&2

# Return decision (if applicable)
exit 0
```

**3. Testing Hooks**
```bash
# Simulate hook input
echo '{
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": {"command": "npm test"}
}' | ./.claude/hooks/security-check.sh
```

**4. Error Handling**
- Always use `set -euo pipefail` in Bash scripts
- Quote variables: `"$VAR"` not `$VAR`
- Validate JSON fields exist before use
- Provide helpful error messages to stderr

**5. Performance**
- Keep PreToolUse hooks fast (<5s)
- Use async for slow operations (tests, builds)
- Cache expensive checks when possible

---

## Specialized Hooks from Community

### TTS Integration (disler)

Text-to-Speech notifications when Claude completes:

```bash
#!/bin/bash
INPUT=$(cat)
MESSAGE=$(echo "$INPUT" | jq -r '.last_assistant_message' | head -c 100)

# macOS TTS
say "Claude has finished: $MESSAGE"
```

### Git Commit Message Generation

```bash
#!/bin/bash
DIFF=$(git diff --cached)
if [[ -n "$DIFF" ]]; then
  echo "## Staged Changes for Commit"
  echo "$DIFF" | head -50
fi
```

### Dependency Check

```bash
#!/bin/bash
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')

if [[ "$FILE" == "package.json" ]] || [[ "$FILE" == "requirements.txt" ]]; then
  echo "Dependency file changed - consider running install" >&2
fi
```

---

## Known Limitations (From GitHub Issues)

### Documented Issues

1. **Environment Variable Substitution**
   - GitHub Issue: anthropics/claude-code#5489, #9567
   - `CLAUDE_TOOL_NAME` and `CLAUDE_TOOL_PARAMS` sometimes fail
   - Workaround: Use JSON stdin instead

2. **Hook Execution Context**
   - Hooks run with user's full permissions
   - No sandboxing
   - Be careful with untrusted input

3. **JSON Parsing Interference**
   - Shell profile output can break JSON
   - Use `2>&1` redirects carefully
   - Test JSON parsing separately

---

## Migration Examples

### From Cursor to Claude Code

**Cursor Config:**
```json
{
  "hooks": {
    "beforeFileEdit": ["./hooks/format.sh"]
  }
}
```

**Claude Code Equivalent:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "./hooks/format.sh"
          }
        ]
      }
    ]
  }
}
```

---

## Sources

**GitHub Repositories:**
- https://github.com/disler/claude-code-hooks-mastery
- https://github.com/affaan-m/everything-claude-code
- https://github.com/karanb192/claude-code-hooks
- https://github.com/ChrisWiles/claude-code-showcase
- https://github.com/johnlindquist/claude-hooks
- https://github.com/decider/claude-hooks
- https://github.com/trailofbits/claude-code-config
- https://github.com/hesreallyhim/awesome-claude-code
- https://github.com/wesammustafa/Claude-Code-Everything-You-Need-to-Know

**Community Resources:**
- Anthropic Hackathon Winners (Feb 2026)
- Production configurations (10+ months battle-tested)

Research date: 2026-02-21
