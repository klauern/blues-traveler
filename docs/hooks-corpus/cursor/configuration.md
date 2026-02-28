# Cursor IDE Configuration & Setup

**Complete guide to configuring Cursor hooks**

---

## Configuration File Location

### File Name

`.cursor/hooks.json`

### File Locations (Priority Order)

**1. Project-specific** (Highest Priority)
- Path: `.cursor/hooks.json` (relative to project root)
- Scope: Single project only
- Committed: Usually yes (version control)
- Use: Team-wide policies

**2. User-specific** (Medium Priority)
- Path: `~/.cursor/hooks.json`
- Scope: All projects for that user
- Committed: No
- Use: Personal preferences

**3. Global/System** (Lowest Priority)
- Path: System-dependent (undocumented)
- Scope: All users on system
- Committed: No
- Use: Organization-wide policies

### Precedence Rules

- **All hooks execute**: Hooks from all levels run
- **Execution order**: System → User → Project
- **Override behavior**: Later `deny` overrides earlier `allow`
- **No merge logic**: Each hooks.json evaluated independently

---

## JSON Schema

### Official Schema URL

```
https://unpkg.com/cursor-hooks/schema/hooks.schema.json
```

### Minimal Configuration

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./security-check.sh" }
    ]
  }
}
```

### Complete Configuration (All Events)

```json
{
  "$schema": "https://unpkg.com/cursor-hooks/schema/hooks.schema.json",
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security.sh" },
      { "command": "./hooks/audit.sh" }
    ],
    "beforeMCPExecution": [
      { "command": "./hooks/mcp-validate.sh" }
    ],
    "beforeReadFile": [
      { "command": "./hooks/secret-scan.sh" }
    ],
    "afterFileEdit": [
      { "command": "./hooks/format.sh" },
      { "command": "./hooks/lint.sh" }
    ],
    "beforeSubmitPrompt": [
      { "command": "./hooks/prompt-gate.sh" }
    ],
    "stop": [
      { "command": "./hooks/notify.sh" },
      { "command": "./hooks/cleanup.sh" }
    ]
  }
}
```

---

## Schema Reference

### Root Level Fields

#### `version` (required)

- **Type:** Integer
- **Value:** Must be `1`
- **Description:** Schema version identifier

```json
{
  "version": 1
}
```

#### `hooks` (required)

- **Type:** Object
- **Properties:** Event names → Hook arrays
- **Validation:** Must have at least 1 property
- **additionalProperties:** `false`

```json
{
  "hooks": {
    "beforeShellExecution": [ ... ]
  }
}
```

### Hook Event Properties

Supported event names:
- `beforeShellExecution`
- `beforeMCPExecution`
- `beforeReadFile`
- `afterFileEdit`
- `beforeSubmitPrompt`
- `stop`

### Hook Definition Object

#### `command` (required)

- **Type:** String
- **MinLength:** 1
- **Format:** Executable path or shell snippet

```json
{ "command": "./hooks/script.sh" }
{ "command": "/absolute/path/to/script" }
{ "command": "bun run hooks/check.ts" }
{ "command": "python -m hooks.validate" }
```

---

## Path Resolution

### Relative Paths

**Base Directory:** Directory containing `hooks.json`

```
project/
├── .cursor/
│   ├── hooks.json          ← Base directory
│   └── hooks/
│       └── check.sh        ← Referenced as "./hooks/check.sh"
```

### Absolute Paths

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "/usr/local/bin/my-hook" }
    ]
  }
}
```

### Home Directory (~)

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "~/.cursor/hooks/global-check.sh" }
    ]
  }
}
```

### Environment Variables

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "$HOME/.cursor/hooks/check.sh" }
    ]
  }
}
```

### Shell Commands

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "bun run hooks/check.ts" },
      { "command": "python -m hooks.validate" }
    ]
  }
}
```

---

## Multiple Hooks per Event

### Array Execution Order

Hooks execute **top to bottom**:

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hook1.sh" },  // Runs first
      { "command": "./hook2.sh" },  // Runs second
      { "command": "./hook3.sh" }   // Runs third
    ]
  }
}
```

### Short-Circuit Behavior

**For permission-based events:**
- First `deny` → Block action
- All must `allow` → Action proceeds

**For informational events:**
- All hooks run regardless of failures

---

## IDE Integration

### VS Code / Cursor JSON Schema

**Option 1: Add $schema to hooks.json**

```json
{
  "$schema": "https://unpkg.com/cursor-hooks@latest/schema/hooks.schema.json",
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./check.sh" }
    ]
  }
}
```

**Benefits:**
- Autocomplete for event names
- Inline documentation on hover
- Validation errors for typos

**Option 2: Global Cursor Settings**

File: `~/Library/Application Support/Cursor/User/settings.json` (macOS)

```json
{
  "json.validate.enable": true,
  "json.schemas": [
    {
      "fileMatch": [".cursor/hooks.json"],
      "url": "https://unpkg.com/cursor-hooks/schema/hooks.schema.json"
    }
  ]
}
```

---

## Setup Instructions

### Initial Setup

```bash
# Step 1: Create directory structure
mkdir -p .cursor/hooks

# Step 2: Create hooks.json
cat > .cursor/hooks.json << 'HOOKS_JSON'
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security-check.sh" }
    ]
  }
}
HOOKS_JSON

# Step 3: Create hook script
cat > .cursor/hooks/security-check.sh << 'HOOK_SCRIPT'
#!/bin/bash
input=$(cat)
command=$(echo "$input" | jq -r '.command')

if echo "$command" | grep -qE '(rm -rf /|sudo rm)'; then
  echo '{"permission":"deny","userMessage":"Dangerous command blocked"}'
  exit 0
fi

echo '{"permission":"allow"}'
HOOK_SCRIPT

# Step 4: Make executable
chmod +x .cursor/hooks/security-check.sh

# Step 5: Restart Cursor
```

### Verification

```bash
# Validate JSON syntax
jq empty .cursor/hooks.json

# Test hook manually
echo '{"command":"test"}' | .cursor/hooks/security-check.sh

# Check output
# Expected: {"permission":"allow"}
```

---

## Configuration Examples

### Security: Block Dangerous Commands

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security-check.sh" }
    ],
    "beforeReadFile": [
      { "command": "./hooks/secret-scanner.sh" }
    ]
  }
}
```

### Code Quality: Auto-format

```json
{
  "version": 1,
  "hooks": {
    "afterFileEdit": [
      { "command": "bun run hooks/format.ts" }
    ]
  }
}
```

### Tool Enforcement: Bun over npm

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "bun run hooks/enforce-bun.ts" }
    ]
  }
}
```

---

## Common Configuration Errors

### Error 1: Invalid Event Name

```json
{
  "hooks": {
    "beforeCommand": [ ... ]  // ❌ Not valid
  }
}
```

**Valid names:**
- `beforeShellExecution`
- `beforeMCPExecution`
- `afterFileEdit`
- `beforeReadFile`
- `beforeSubmitPrompt`
- `stop`

### Error 2: Missing `command` Field

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "script": "./check.sh" }  // ❌ Should be "command"
    ]
  }
}
```

### Error 3: Empty Command

```json
{ "command": "" }  // ❌ MinLength: 1
```

### Error 4: Wrong Type

```json
{
  "version": "1"  // ❌ Should be integer
}
```

---

**Sources:**
- [JSON Schema](https://unpkg.com/cursor-hooks/schema/hooks.schema.json)
- [GitButler Configuration Guide](https://blog.gitbutler.com/cursor-hooks-deep-dive)
- [Cursor Official Docs](https://cursor.com/docs/agent/hooks)
