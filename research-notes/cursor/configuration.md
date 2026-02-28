# Cursor Hooks Configuration Reference

**Research Date:** 2026-02-21

## Configuration File Format

### File Name
`hooks.json`

### File Locations (Priority Order)

1. **Project-specific** (Highest Priority)
   - Path: `.cursor/hooks.json` (relative to project root)
   - Scope: Single project only
   - Committed: Usually yes (version control)
   - Use: Team-wide policies and project-specific rules

2. **User-specific** (Medium Priority)
   - Path: `~/.cursor/hooks.json` (user home directory)
   - Scope: All projects for that user
   - Committed: No (user-local)
   - Use: Personal preferences and automation

3. **Global/System** (Lowest Priority)
   - Path: System-dependent (not well-documented)
   - Scope: All users on system
   - Committed: No
   - Use: Organization-wide policies

### Precedence Rules

- **All applicable hooks execute**: If hooks exist at multiple levels, ALL are run
- **Execution order**: System → User → Project (within each level, array order)
- **Override behavior**: Later hooks can override earlier decisions (deny overrides allow)
- **No merge logic**: hooks.json files do NOT merge; each is evaluated independently

---

## JSON Schema

### Official Schema URL
```
https://unpkg.com/cursor-hooks/schema/hooks.schema.json
```

### Complete Schema Definition

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://unpkg.com/cursor-hooks/schema/hooks.schema.json",
  "title": "Cursor Hooks Configuration",
  "type": "object",
  "additionalProperties": false,
  "required": ["version", "hooks"],
  "definitions": {
    "hookDefinition": {
      "type": "object",
      "additionalProperties": false,
      "required": ["command"],
      "properties": {
        "command": {
          "type": "string",
          "minLength": 1,
          "description": "Command to execute. Supports absolute paths, paths relative to hooks.json, or shell snippets."
        }
      }
    },
    "hooksArray": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/hookDefinition"
      }
    }
  },
  "properties": {
    "version": {
      "type": "integer",
      "const": 1,
      "description": "Configuration schema version. Currently only 1 is supported."
    },
    "hooks": {
      "type": "object",
      "minProperties": 1,
      "description": "Hook event name to hook definition mappings.",
      "properties": {
        "beforeShellExecution": {
          "$ref": "#/definitions/hooksArray"
        },
        "beforeMCPExecution": {
          "$ref": "#/definitions/hooksArray"
        },
        "afterFileEdit": {
          "$ref": "#/definitions/hooksArray"
        },
        "beforeReadFile": {
          "$ref": "#/definitions/hooksArray"
        },
        "beforeSubmitPrompt": {
          "$ref": "#/definitions/hooksArray"
        },
        "stop": {
          "$ref": "#/definitions/hooksArray"
        }
      },
      "additionalProperties": false
    }
  }
}
```

---

## Configuration Structure

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

## Field Reference

### Root Level

#### `version` (required)
- **Type:** Integer
- **Value:** Must be `1`
- **Description:** Schema version identifier
- **Future:** May support version 2+ with breaking changes

```json
{
  "version": 1
}
```

#### `hooks` (required)
- **Type:** Object
- **Properties:** Event names → Hook arrays
- **Validation:** Must have at least 1 property
- **additionalProperties:** `false` (only documented events allowed)

```json
{
  "hooks": {
    "beforeShellExecution": [ ... ],
    "afterFileEdit": [ ... ]
  }
}
```

---

### Hook Event Properties

Each event maps to an array of hook definitions.

#### Supported Event Names
- `beforeShellExecution`
- `beforeMCPExecution`
- `beforeReadFile`
- `afterFileEdit`
- `beforeSubmitPrompt`
- `stop`

#### Hook Definition Object

Each hook in the array is an object with:

**`command` (required)**
- **Type:** String
- **MinLength:** 1
- **Format:** Executable path or shell snippet
- **Resolution:** Relative to `hooks.json` location

```json
{ "command": "./hooks/script.sh" }
{ "command": "/absolute/path/to/script" }
{ "command": "bun run hooks/check.ts" }
{ "command": "python -m my_hooks.validate" }
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

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/check.sh" }
    ]
  }
}
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

**Supported:** Yes (shell expansion)

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

**Supported:** Yes (via shell)

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

**Supported:** Yes (arbitrary shell snippets)

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "bun run hooks/check.ts" },
      { "command": "python -m hooks.validate" },
      { "command": "uvx --from /path/to/project python -m hooks.run" }
    ]
  }
}
```

---

## Multiple Hooks per Event

### Array Execution Order

Hooks execute **top to bottom** in array order:

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
- First `deny` → Block action, may skip remaining hooks
- All must `allow` → Action proceeds

**For flow-control events:**
- First `continue: false` → Block prompt, may skip remaining hooks

**For informational events:**
- All hooks run regardless of individual failures

---

## IDE Integration

### VS Code / Cursor JSON Schema

Enable autocomplete and validation.

#### Option 1: Add $schema to hooks.json

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
- Schema-aware suggestions

#### Option 2: Global Cursor Settings

**File:** `~/Library/Application Support/Cursor/User/settings.json` (macOS)
**File:** `~/.config/Cursor/User/settings.json` (Linux)
**File:** `%APPDATA%\Cursor\User\settings.json` (Windows)

```json
{
  "json.validate.enable": true,
  "json.format.enable": true,
  "json.schemaDownload.enable": true,
  "json.schemas": [
    {
      "fileMatch": [".cursor/hooks.json"],
      "url": "https://unpkg.com/cursor-hooks/schema/hooks.schema.json"
    }
  ]
}
```

**Advantages:**
- Applies to all projects
- No per-project modification needed
- Centralized configuration

---

## Configuration Examples by Use Case

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

### Audit: Log Everything

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/audit.sh" }
    ],
    "beforeMCPExecution": [
      { "command": "./hooks/audit.sh" }
    ],
    "beforeReadFile": [
      { "command": "./hooks/audit.sh" }
    ],
    "afterFileEdit": [
      { "command": "./hooks/audit.sh" }
    ],
    "beforeSubmitPrompt": [
      { "command": "./hooks/audit.sh" }
    ],
    "stop": [
      { "command": "./hooks/audit.sh" }
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

### Malware Detection: Endor Labs Integration

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./.cursor/hooks/shell-malware-audit-hook.sh" }
    ],
    "afterFileEdit": [
      { "command": "./.cursor/hooks/malware-audit-hook.sh" }
    ],
    "stop": [
      { "command": "./.cursor/hooks/session-end.sh" }
    ]
  }
}
```

### MCP Governance: ToolHive

```json
{
  "version": 1,
  "hooks": {
    "beforeMCPExecution": [
      { "command": "~/.cursor/hooks/stacklok-hook.sh" }
    ]
  }
}
```

---

## Multi-Language Hook Configurations

### TypeScript (Bun)

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "bun run hooks/before-shell-execution.ts" }
    ],
    "afterFileEdit": [
      { "command": "bun run hooks/after-file-edit.ts" }
    ]
  }
}
```

**Directory Structure:**
```
.cursor/
├── hooks.json
└── hooks/
    ├── package.json
    ├── before-shell-execution.ts
    └── after-file-edit.ts
```

### Python

```json
{
  "version": 1,
  "hooks": {
    "beforeReadFile": [
      {
        "command": "uvx --from /absolute/path/to/project python -m hooks.run --hook beforeReadFile"
      }
    ]
  }
}
```

**Directory Structure:**
```
.cursor/
├── hooks.json
project/
├── pyproject.toml
└── hooks/
    ├── __init__.py
    ├── models.py
    └── run.py
```

### Bash

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security.sh" }
    ]
  }
}
```

**Directory Structure:**
```
.cursor/
├── hooks.json
└── hooks/
    ├── security.sh
    ├── audit.sh
    └── format.sh
```

---

## Environment Variable Support

### No Direct Configuration

**Limitation:** Cannot pass environment variables via `hooks.json`

**Workaround:** Use wrapper scripts

```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/wrapper.sh" }
    ]
  }
}
```

```bash
#!/bin/bash
# wrapper.sh
export HOOK_DEBUG=1
export API_KEY=$(security find-generic-password -s my-service -w)
./hooks/actual-hook.sh
```

### Available Environment Variables

Cursor does NOT document environment variables passed to hooks.
Hooks rely on **stdin JSON** for context, not env vars.

**Potential env vars (undocumented):**
- `CURSOR_*` - Cursor-specific variables
- `PWD` - Current working directory
- Standard shell env (`HOME`, `USER`, etc.)

---

## Validation and Testing

### Validate JSON Schema

**Using jq:**
```bash
jq empty .cursor/hooks.json
# No output = valid JSON
# Error output = syntax error
```

**Using online validator:**
1. Copy `.cursor/hooks.json` contents
2. Visit https://www.jsonschemavalidator.net/
3. Paste schema URL: `https://unpkg.com/cursor-hooks/schema/hooks.schema.json`
4. Validate

### Test Hook Execution

**Manual stdin test:**
```bash
echo '{"command":"test"}' | .cursor/hooks/my-hook.sh
```

**Expected output:**
```json
{"permission":"allow"}
```

### Verify Hook Registration

After creating `hooks.json`:
1. **Restart Cursor** (required to load new config)
2. Trigger hook event (e.g., run shell command)
3. Check hook execution (logs, outputs, behavior)

---

## Common Configuration Errors

### Error 1: Invalid Event Name

```json
{
  "hooks": {
    "beforeCommand": [ ... ]  // ❌ Not a valid event
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

**Correct:**
```json
{ "command": "./check.sh" }
```

### Error 3: Empty Command

```json
{ "command": "" }  // ❌ MinLength: 1
```

### Error 4: Wrong Type

```json
{
  "version": "1"  // ❌ Should be integer, not string
}
```

**Correct:**
```json
{ "version": 1 }
```

### Error 5: Additional Properties

```json
{
  "version": 1,
  "hooks": { ... },
  "timeout": 5000  // ❌ Not allowed (additionalProperties: false)
}
```

---

## Migration Patterns

### From Global to Project

**Before:** `~/.cursor/hooks.json`
```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "~/.cursor/hooks/check.sh" }
    ]
  }
}
```

**After:** `.cursor/hooks.json` (project)
```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/check.sh" }
    ]
  }
}
```

**Steps:**
1. Copy hook scripts to `.cursor/hooks/`
2. Create `.cursor/hooks.json` with relative paths
3. Commit to version control
4. Team members get hooks automatically

### From Bash to TypeScript

**Before:** Bash script
```json
{ "command": "./hooks/check.sh" }
```

**After:** TypeScript with Bun
```json
{ "command": "bun run hooks/check.ts" }
```

**Steps:**
1. Install Bun and cursor-hooks: `bun add cursor-hooks`
2. Rewrite hook logic in TypeScript
3. Update hooks.json command
4. Test with sample input

---

## Performance Optimization

### Configuration Impact

**Multiple hooks per event:**
- Execute sequentially (not parallel)
- Total time = sum of individual hook times
- Keep critical hooks fast

**Recommended:**
```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./fast-security-check.sh" },  // < 100ms
      { "command": "./audit-log.sh" }             // < 50ms
    ]
  }
}
```

**Not recommended:**
```json
{
  "hooks": {
    "beforeShellExecution": [
      { "command": "./slow-api-call.sh" },       // 2+ seconds
      { "command": "./complex-scan.sh" }         // 5+ seconds
    ]
  }
}
```

### Conditional Execution

Use hook logic to skip unnecessary work:

```bash
#!/bin/bash
# Only run for Python files in afterFileEdit
input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path')

if [[ ! "$file_path" =~ \.py$ ]]; then
  exit 0  # Skip for non-Python files
fi

# Run Python formatter
ruff format "$file_path"
```

---

## Security Considerations

### Trusted Sources Only

Hooks execute arbitrary code with full user permissions.

**Safe:**
- Official examples from Cursor docs
- Well-maintained GitHub repos (e.g., 1Password, Endor Labs)
- Internal team repositories

**Risky:**
- Random internet scripts
- Unverified third-party hooks
- Hooks without code review

### Code Review

Before adding hooks to project:
1. **Read the entire script**
2. **Understand what it does**
3. **Check for malicious patterns**:
   - Network calls to unknown domains
   - File system modifications outside workspace
   - Credential theft attempts
   - Command injection vulnerabilities

### Permissions

Hooks run with **user's full permissions**:
- Can read any file user can read
- Can write any file user can write
- Can execute any command user can execute
- **No sandbox isolation**

**Best practice:** Principle of least privilege
- Don't use `sudo` in hooks
- Limit file system access
- Validate inputs rigorously

---

## Troubleshooting

### Hooks Not Running

**Checklist:**
1. ✅ `hooks.json` in correct location (`.cursor/hooks.json`)
2. ✅ Cursor restarted after config changes
3. ✅ Hook scripts are executable (`chmod +x`)
4. ✅ Paths are correct (relative to `hooks.json`)
5. ✅ JSON syntax is valid (`jq empty hooks.json`)
6. ✅ Event names match schema exactly

### Invalid JSON Errors

**Symptoms:** Hook blocks all actions with "invalid JSON" error

**Causes:**
- Extra output to stdout before JSON
- Malformed JSON (missing quotes, commas)
- Non-JSON output (e.g., debug echo statements)

**Fix:**
- Use `jq` to validate output
- Move debug output to stderr or log files
- Test hook in isolation: `echo '{"test":"data"}' | ./hook.sh | jq`

### Path Resolution Errors

**Symptom:** "Command not found" or "Permission denied"

**Fixes:**
```bash
# Check if path is correct
ls -la .cursor/hooks/script.sh

# Ensure executable
chmod +x .cursor/hooks/script.sh

# Test direct execution
.cursor/hooks/script.sh <<< '{"command":"test"}'
```

---

## References

- JSON Schema Spec: https://unpkg.com/cursor-hooks/schema/hooks.schema.json
- TypeScript SDK: https://github.com/johnlindquist/cursor-hooks
- Example Configs: https://github.com/hamzafer/cursor-hooks
- Official Docs: https://cursor.com/docs/agent/hooks
