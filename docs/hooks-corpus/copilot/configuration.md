# GitHub Copilot Hooks - Configuration & Setup

**Version:** 1.0
**Last Updated:** 2026-02-21

---

## Configuration File Location

### Repository Configuration (Recommended)

**File:** `.github/hooks/*.json`

```
repository/
├── .github/
│   └── hooks/
│       ├── hooks.json            # Main configuration
│       ├── security-hooks.json   # Can split by category
│       └── scripts/
│           ├── security-check.sh
│           ├── audit-log.sh
│           └── session-start.sh
└── ...
```

**Requirements:**
- ✅ Must be on **default branch** (main/master)
- ✅ Can have multiple `*.json` files (merged together)
- ✅ Relative paths resolved from repository root
- ✅ Version controlled with repository

**Why repository configuration:**
- Team collaboration - everyone uses same hooks
- Version control - track hook changes over time
- Code review - hooks reviewed like code
- Deployment - automatic with repository

### Workspace Configuration (Not Applicable)

**Note:** Unlike Claude Code, GitHub Copilot does **not** support workspace-level or user-level hook configurations. All hooks must be in the repository.

**Limitation:** Cannot have personal hooks that differ from team hooks.

---

## Configuration Format

### Basic Structure

```json
{
  "version": 1,
  "hooks": {
    "hookName": [
      {
        "type": "command",
        "bash": "./path/to/script.sh",
        "powershell": "./path/to/script.ps1",
        "cwd": ".",
        "env": {
          "KEY": "value"
        },
        "timeoutSec": 30
      }
    ]
  }
}
```

### Minimal Example

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/security-check.sh"
    }]
  }
}
```

### Complete Example

```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{
      "type": "command",
      "bash": "./scripts/session-start.sh",
      "powershell": "./scripts/session-start.ps1",
      "cwd": ".",
      "env": {
        "LOG_LEVEL": "info",
        "PROJECT_NAME": "my-app"
      },
      "timeoutSec": 30
    }],
    "preToolUse": [
      {
        "type": "command",
        "bash": "./scripts/security-check.sh",
        "powershell": "./scripts/security-check.ps1",
        "timeoutSec": 5
      },
      {
        "type": "command",
        "bash": "./scripts/audit-log.sh",
        "timeoutSec": 3
      }
    ],
    "postToolUse": [{
      "type": "command",
      "bash": "./scripts/track-usage.sh",
      "env": {
        "METRICS_FILE": "metrics.jsonl"
      },
      "timeoutSec": 3
    }],
    "errorOccurred": [{
      "type": "command",
      "bash": "./scripts/error-notify.sh",
      "env": {
        "SLACK_WEBHOOK_URL": "${SLACK_WEBHOOK_URL}"
      },
      "timeoutSec": 10
    }],
    "sessionEnd": [{
      "type": "command",
      "bash": "./scripts/cleanup.sh",
      "timeoutSec": 60
    }]
  }
}
```

---

## Schema Reference

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | number | ✅ Yes | Configuration schema version (currently `1`) |
| `hooks` | object | ✅ Yes | Hook definitions by event type |

### Hook Event Keys

Valid keys for the `hooks` object:

- `sessionStart`
- `sessionEnd`
- `userPromptSubmitted`
- `preToolUse`
- `postToolUse`
- `errorOccurred`
- `agentStop`

### Hook Definition Fields

Each hook is an object with these fields:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | ✅ Yes | - | Hook type (currently only `"command"`) |
| `bash` | string | ⚠️ One required | - | Path to bash script (Unix/Linux/macOS) |
| `powershell` | string | ⚠️ One required | - | Path to PowerShell script (Windows) |
| `command` | string/object | ⚠️ Alternative | - | Generic command or OS-specific commands |
| `cwd` | string | ❌ No | `"."` | Working directory (relative to repo root) |
| `env` | object | ❌ No | `{}` | Environment variables |
| `timeoutSec` | number | ❌ No | `30` | Timeout in seconds (1-90 practical limit) |

**Note:** You must specify at least one of `bash`, `powershell`, or `command`.

### Platform-Specific Commands

**Option 1: Separate scripts (Recommended)**
```json
{
  "bash": "./scripts/hook.sh",
  "powershell": "./scripts/hook.ps1"
}
```

**Option 2: Generic command**
```json
{
  "command": "node ./scripts/hook.js"
}
```

**Option 3: OS-specific commands**
```json
{
  "command": {
    "windows": "powershell -File ./scripts/hook.ps1",
    "linux": "bash ./scripts/hook.sh",
    "osx": "bash ./scripts/hook.sh"
  }
}
```

---

## Configuration Merging

### Multiple Files

All `*.json` files in `.github/hooks/` are **merged together**:

```
.github/hooks/
├── security.json     # Security hooks
├── audit.json        # Audit logging
└── notifications.json # Slack/email alerts
```

**Merge Strategy:**
- Arrays are **concatenated** (hooks execute in order)
- Objects are **merged** (later values override)
- No deduplication (identical hooks execute multiple times)

**Example:**

`security.json`:
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {"bash": "./scripts/security.sh"}
    ]
  }
}
```

`audit.json`:
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {"bash": "./scripts/audit.sh"}
    ]
  }
}
```

**Result:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {"bash": "./scripts/security.sh"},
      {"bash": "./scripts/audit.sh"}
    ]
  }
}
```

Both hooks execute in order.

---

## Setup Instructions

### Initial Setup

**Step 1: Create hooks directory**
```bash
mkdir -p .github/hooks/scripts
```

**Step 2: Create configuration file**
```bash
cat > .github/hooks/hooks.json <<'EOF'
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/security-check.sh",
      "timeoutSec": 5
    }]
  }
}
EOF
```

**Step 3: Create hook script**
```bash
cat > .github/hooks/scripts/security-check.sh <<'EOF'
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

if [ "$TOOL_NAME" = "bash" ]; then
  COMMAND=$(echo "$INPUT" | jq -r '.toolArgs.command')
  if echo "$COMMAND" | grep -qE "(rm -rf|sudo)"; then
    echo '{"permissionDecision":"deny","permissionDecisionReason":"Dangerous command blocked"}' | jq -c
    exit 0
  fi
fi

echo '{"permissionDecision":"allow"}' | jq -c
EOF

chmod +x .github/hooks/scripts/security-check.sh
```

**Step 4: Test locally**
```bash
echo '{"toolName":"bash","toolArgs":"{\"command\":\"rm -rf /\"}"}' | \
  .github/hooks/scripts/security-check.sh
# Should output: {"permissionDecision":"deny",...}

echo '{"toolName":"bash","toolArgs":"{\"command\":\"ls -la\"}"}' | \
  .github/hooks/scripts/security-check.sh
# Should output: {"permissionDecision":"allow"}
```

**Step 5: Commit to default branch**
```bash
git add .github/hooks/
git commit -m "Add GitHub Copilot security hooks"
git push origin main
```

**Step 6: Verify in Copilot**
- Start GitHub Copilot session
- Try dangerous command: "run rm -rf /"
- Should be blocked with reason

---

## Verification

### Test Hook Configuration

**1. Validate JSON syntax:**
```bash
jq empty .github/hooks/*.json
# No output = valid JSON
```

**2. Test hooks locally:**
```bash
# Create test input
cat > test-input.json <<'EOF'
{
  "timestamp": 1708560000000,
  "cwd": "/workspace",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"sudo rm -rf /\"}"
}
EOF

# Test hook
cat test-input.json | .github/hooks/scripts/security-check.sh

# Verify output format
cat test-input.json | .github/hooks/scripts/security-check.sh | jq empty
```

**3. Check script permissions:**
```bash
ls -l .github/hooks/scripts/
# Should show +x (executable)
```

**4. Verify on default branch:**
```bash
git branch --show-current
# Should be main, master, or your default branch
```

---

## Advanced Configuration

### Environment Variables

**Set environment for hooks:**
```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./scripts/hook.sh",
      "env": {
        "LOG_LEVEL": "debug",
        "API_KEY": "${API_KEY}",
        "PROJECT": "my-app"
      }
    }]
  }
}
```

**Access in hook:**
```bash
#!/bin/bash
echo "Log level: $LOG_LEVEL" >&2
echo "Project: $PROJECT" >&2
```

**Note:** `${VAR}` syntax references environment variables from parent process.

### Custom Working Directory

**Run hook in subdirectory:**
```json
{
  "hooks": {
    "sessionStart": [{
      "bash": "./setup.sh",
      "cwd": "scripts"
    }]
  }
}
```

Script executes in `repository/scripts/` directory.

### Timeout Configuration

**Different timeouts per hook:**
```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./quick-check.sh",
      "timeoutSec": 3  // Fast security check
    }],
    "sessionEnd": [{
      "bash": "./cleanup.sh",
      "timeoutSec": 90  // Longer for cleanup
    }]
  }
}
```

**Guidelines:**
- Security checks: 3-5 seconds
- Logging: 2-3 seconds
- Setup/cleanup: 30-90 seconds
- External API calls: 10-15 seconds

---

## Best Practices

### Organization

**1. Separate by purpose:**
```
.github/hooks/
├── security.json          # Security enforcement
├── audit.json            # Compliance logging
├── notifications.json    # External integrations
└── scripts/
    ├── security/
    │   ├── check-commands.sh
    │   └── scan-secrets.sh
    ├── audit/
    │   ├── log-session.sh
    │   └── log-tools.sh
    └── notify/
        ├── slack.sh
        └── email.sh
```

**2. Use descriptive names:**
```json
{
  "hooks": {
    "preToolUse": [
      {"bash": "./scripts/security/check-dangerous-commands.sh"},
      {"bash": "./scripts/security/scan-for-secrets.sh"},
      {"bash": "./scripts/audit/log-tool-usage.sh"}
    ]
  }
}
```

**3. Document hooks:**
```bash
#!/bin/bash
# Purpose: Block dangerous system commands
# Event: preToolUse
# Returns: {"permissionDecision": "allow"|"deny"}
# Timeout: 5 seconds
```

### Version Control

**1. Commit hooks with code:**
```bash
git add .github/hooks/
git commit -m "Add Copilot hooks for security enforcement"
```

**2. Review hook changes:**
```bash
git diff main...feature-branch -- .github/hooks/
```

**3. Test before merging:**
- Test locally with sample inputs
- Review in pull request
- Deploy to staging first

### Testing Strategy

**1. Unit test individual hooks:**
```bash
./test-hooks.sh
```

`test-hooks.sh`:
```bash
#!/bin/bash
echo "Testing security-check.sh..."

# Test 1: Dangerous command
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"rm -rf /\"}"}' | \
  .github/hooks/scripts/security-check.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "deny"' > /dev/null; then
  echo "✅ Test 1 passed"
else
  echo "❌ Test 1 failed"
  exit 1
fi

# Test 2: Safe command
RESULT=$(echo '{"toolName":"bash","toolArgs":"{\"command\":\"ls -la\"}"}' | \
  .github/hooks/scripts/security-check.sh)
if echo "$RESULT" | jq -e '.permissionDecision == "allow"' > /dev/null; then
  echo "✅ Test 2 passed"
else
  echo "❌ Test 2 failed"
  exit 1
fi

echo "All tests passed!"
```

**2. Integration test with Copilot:**
- Start session in test repository
- Try various commands
- Verify hooks behave correctly

**3. Gradual rollout:**
```json
// Phase 1: Logging only (Week 1-2)
{
  "hooks": {
    "preToolUse": [{"bash": "./scripts/log-only.sh"}]
  }
}

// Phase 2: Warnings (Week 3-4)
// Script warns but allows

// Phase 3: Enforcement (Week 5+)
// Script denies dangerous operations
```

---

## Troubleshooting Configuration

### Common Issues

**Issue: Hooks not executing**
```bash
# Check 1: On default branch?
git branch --show-current

# Check 2: Valid JSON?
jq empty .github/hooks/*.json

# Check 3: Correct path?
ls -la .github/hooks/scripts/

# Check 4: Executable?
ls -l .github/hooks/scripts/*.sh
```

**Issue: Permission denied**
```bash
chmod +x .github/hooks/scripts/*.sh
git add .github/hooks/scripts/
git commit -m "Fix hook permissions"
git push
```

**Issue: Hook times out**
```json
{
  "hooks": {
    "preToolUse": [{
      "bash": "./slow-hook.sh",
      "timeoutSec": 60  // Increase timeout
    }]
  }
}
```

---

## Migration from Other Systems

### From Claude Code

**Claude Code:**
```json
{
  "hooks": {
    "PreToolUse": [{
      "type": "command",
      "command": "./hook.sh"
    }]
  }
}
```

**GitHub Copilot:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./hook.sh",
      "powershell": "./hook.ps1"
    }]
  }
}
```

**Changes:**
- ✅ Add `"version": 1`
- ✅ Change event name casing (`PreToolUse` → `preToolUse`)
- ✅ Specify `bash` and/or `powershell` instead of generic `command`
- ⚠️ Move to `.github/hooks/*.json` instead of `.claude/settings.json`
- ⚠️ Must be on default branch (no workspace/user configs)

**See:** [Migration Guide](./migration.md) for complete details

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Available hooks
- [Scripting & Execution](./scripting.md) - Writing hook scripts
- [Examples](./examples.md) - Ready-to-use patterns
- [API Reference](./api-reference.md) - Complete schemas

---

**Sources:**
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [Using Hooks with Copilot Agents](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/use-hooks)
- [Copilot CLI Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

---

**Document Version:** 1.0
**Status:** ✅ Complete
