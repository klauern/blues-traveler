# Cursor Hooks: GitHub Examples and Real-World Implementations

**Research Date:** 2026-02-21

## Overview

This document catalogs real-world Cursor hooks implementations found on GitHub, organized by use case and implementation language.

---

## Major Example Repositories

### 1. hamzafer/cursor-hooks
**URL:** https://github.com/hamzafer/cursor-hooks
**Language:** Bash
**Purpose:** Ready-to-use examples and documentation
**Last Updated:** 2024+ (actively maintained)

#### Files Structure
```
.cursor/
├── hooks.json          # Configuration with all 6 events
└── hooks/
    ├── audit.sh        # Universal audit logging
    ├── block-git.sh    # Git command enforcement
    ├── redact-secrets.sh  # Secret detection
    └── format.sh       # Minimal formatter hook
```

#### hooks.json Configuration
```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/audit.sh" },
      { "command": "./hooks/block-git.sh" }
    ],
    "beforeMCPExecution": [
      { "command": "./hooks/audit.sh" }
    ],
    "beforeReadFile": [
      { "command": "./hooks/redact-secrets.sh" }
    ],
    "afterFileEdit": [
      { "command": "./hooks/audit.sh" },
      { "command": "./hooks/format.sh" }
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

#### audit.sh - Universal Logging
```bash
#!/bin/bash
# Append every hook payload to /tmp/agent-audit.log for debugging

json_input=$(cat)
timestamp=$(date '+%Y-%m-%d %H:%M:%S')
mkdir -p "$(dirname /tmp/agent-audit.log)"
echo "[$timestamp] $json_input" >> /tmp/agent-audit.log
exit 0
```

**Use Case:** Comprehensive audit trail of all agent actions

#### block-git.sh - Git Command Enforcement
```bash
#!/bin/bash
# Guard shell commands, deny raw git usage, and prefer gh

echo "Hook execution started" >> /tmp/hooks.log
input=$(cat)
command=$(echo "$input" | jq -r '.command // empty')

if [[ "$command" =~ git[[:space:]] ]] || [[ "$command" == "git" ]]; then
    echo "Git command detected - blocking: '$command'" >> /tmp/hooks.log
    cat <<'__DENY__'
{
  "continue": true,
  "permission": "deny",
  "userMessage": "Git command blocked. Please use the GitHub CLI (gh) tool instead.",
  "agentMessage": "The git command '$command' has been blocked by a project hook. Instead of using raw git commands, please use the 'gh' tool."
}
__DENY__
elif [[ "$command" =~ gh[[:space:]] ]] || [[ "$command" == "gh" ]]; then
    echo "GitHub CLI command detected - asking for permission: '$command'" >> /tmp/hooks.log
    cat <<'__ASK__'
{
  "continue": true,
  "permission": "ask",
  "userMessage": "GitHub CLI command requires permission: $command",
  "agentMessage": "The command '$command' uses the GitHub CLI (gh). Please review and approve this command if you want to proceed."
}
__ASK__
else
    echo "Non-git/non-gh command detected - allowing: '$command'" >> /tmp/hooks.log
    cat <<'__ALLOW__'
{
  "continue": true,
  "permission": "allow"
}
__ALLOW__
fi
```

**Key Features:**
- Blocks raw `git` commands
- Prompts user for `gh` commands
- Logs all decisions to `/tmp/hooks.log`
- Demonstrates `deny` vs `ask` permissions

#### redact-secrets.sh - Secret Detection
```bash
#!/bin/bash
# Deny reads if probable GitHub tokens are detected

input=$(cat)
file_path=$(echo "$input" | jq -r '.file_path // empty')
content=$(echo "$input" | jq -r '.content // empty')

if echo "$content" | grep -qE 'gh[ps]_[A-Za-z0-9]{36}|gh_api_[A-Za-z0-9]+'; then
    echo "GitHub API key detected in file: '$file_path'" >> /tmp/hooks.log
    cat <<'__DENY__'
{
  "permission": "deny"
}
__DENY__
    exit 3
else
    echo "No GitHub API key detected - allowing" >> /tmp/hooks.log
    cat <<'__ALLOW__'
{
  "permission": "allow"
}
__ALLOW__
fi
```

**Detection Patterns:**
- `ghp_*` - Personal access tokens
- `ghs_*` - Server tokens
- `gh_api_*` - API tokens

---

### 2. johnlindquist/cursor-hooks
**URL:** https://github.com/johnlindquist/cursor-hooks
**Language:** TypeScript (Bun)
**Purpose:** Type definitions and TypeScript helpers
**Package:** npm: `cursor-hooks`
**Last Updated:** 2024+ (actively maintained)

#### Installation
```bash
cd .cursor/hooks
bun init -y
bun install cursor-hooks
```

#### Example: Block npm, Enforce bun
```typescript
// before-shell-execution.ts
import type {
  BeforeShellExecutionPayload,
  BeforeShellExecutionResponse
} from "cursor-hooks";

const input: BeforeShellExecutionPayload = await Bun.stdin.json();

const startsWithNpm = input.command.startsWith("npm")
  || input.command.includes(" npm ");

const output: BeforeShellExecutionResponse = {
  permission: startsWithNpm ? "deny" : "allow",
  agentMessage: startsWithNpm
    ? "npm is not allowed, always use bun instead"
    : undefined,
};

console.log(JSON.stringify(output, null, 2));
```

#### Example: Auto-format TypeScript Files
```typescript
// after-file-edit.ts
import type { AfterFileEditPayload } from "cursor-hooks";

const input: AfterFileEditPayload = await Bun.stdin.json();

if (input.file_path.endsWith(".ts")) {
  const result = await Bun.$`bunx @biomejs/biome lint --fix --unsafe --verbose ${input.file_path}`;

  const output = {
    timestamp: new Date().toISOString(),
    ...input,
    stdout: result.stdout.toString(),
    stderr: result.stderr.toString(),
    exitCode: result.exitCode,
  };

  // Only console.errors show in logs
  console.error(JSON.stringify(output, null, 2));
}
```

#### Example: Prompt Gating
```typescript
// before-submit-prompt.ts
import type {
  BeforeSubmitPromptPayload,
  BeforeSubmitPromptResponse
} from "cursor-hooks";

const input: BeforeSubmitPromptPayload = await Bun.stdin.json();

const output: BeforeSubmitPromptResponse = {
  continue: input.prompt.includes("allow"),
};

console.log(JSON.stringify(output, null, 2));
```

#### Type Safety Helper
```typescript
import { isHookPayloadOf } from "cursor-hooks";

const rawInput: unknown = await Bun.stdin.json();

if (isHookPayloadOf(rawInput, "beforeShellExecution")) {
  // TypeScript now knows rawInput is BeforeShellExecutionPayload
  console.log(rawInput.command);
}
```

**Available Types:**
- `BeforeShellExecutionPayload` / `BeforeShellExecutionResponse`
- `BeforeMCPExecutionPayload` / `BeforeMCPExecutionResponse`
- `BeforeReadFilePayload` / `BeforeReadFileResponse`
- `AfterFileEditPayload` (no response type)
- `BeforeSubmitPromptPayload` / `BeforeSubmitPromptResponse`
- `StopPayload` (no response type)

---

### 3. DevonFulcher/py-cursor-hooks
**URL:** https://github.com/DevonFulcher/py-cursor-hooks
**Language:** Python
**Purpose:** Typed Python library for Cursor hooks
**Package:** PyPI: `py-cursor-hooks`
**Last Updated:** 2024+ (actively maintained)

#### Installation
```bash
uv add py-cursor-hooks
```

#### Implementation Pattern
```python
# my_hooks.py
from hooks.interfaces import CursorHooks
from hooks.models import BeforeReadFileInput, BeforeReadFileOutput

class MyHooks(CursorHooks):
    def before_read_file(
        self, input: BeforeReadFileInput
    ) -> BeforeReadFileOutput:
        # Block reading .env files
        if input.file_path.endswith(".env"):
            return BeforeReadFileOutput(
                permission="deny",
                user_message="Cannot read .env files",
                agent_message="Reading .env files is not allowed",
            )
        return BeforeReadFileOutput(permission="allow")

hooks = MyHooks()
```

#### pyproject.toml Configuration
```toml
[project.entry-points."py_cursor_hooks.hooks"]
default = "my_package.my_hooks:hooks"
```

#### Cursor hooks.json
```json
{
  "version": 1,
  "hooks": {
    "beforeReadFile": [
      {
        "command": "uvx --from /absolute/path/to/your/project python -m hooks.run --hook beforeReadFile"
      }
    ]
  }
}
```

**Key Features:**
- Pydantic models for validation
- Type-safe interfaces
- CLI that works with Cursor
- Object-oriented hook implementation

---

### 4. endorlabs/cursor-hook-examples
**URL:** https://github.com/endorlabs/cursor-hook-examples
**Language:** Bash
**Purpose:** Malware detection via Endor Labs API
**Last Updated:** 2024+ (actively maintained)

#### Architecture
```
┌─────────────────────────────────────────────────────────┐
│             Cursor Agent Session                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   ┌─────────────┐    ┌──────────────┐    ┌──────────┐  │
│   │ File Edit   │───▶│afterFileEdit │───▶│ Endor    │  │
│   │ (manifest)  │    │ Hook         │    │ Labs API │  │
│   └─────────────┘    └──────────────┘    └──────────┘  │
│                            │                     │      │
│                            ▼                     │      │
│                      ┌──────────┐                │      │
│                      │ Log/File │◀───────────────┘      │
│                      └──────────┘                       │
│                                                         │
│   ┌─────────────┐    ┌──────────────┐    ┌──────────┐  │
│   │ pkg install │───▶│beforeShell   │───▶│ Endor    │  │
│   │ command     │    │ Hook         │    │ Labs API │  │
│   └─────────────┘    └──────────────┘    └──────────┘  │
│                            │                     │      │
│                            ▼                     │      │
│                      ┌──────────┐                │      │
│                      │Deny/Allow│◀───────────────┘      │
│                      └──────────┘                       │
│                                                         │
│   ┌─────────────┐    ┌──────────────┐                  │
│   │ Session End │───▶│ stop Hook    │───▶ Summary      │
│   └─────────────┘    └──────────────┘                  │
└─────────────────────────────────────────────────────────┘
```

#### hooks.json
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

#### Supported Package Managers
- **JavaScript/Node.js**: npm, yarn, pnpm, bun
- **Python**: pip, pip3, poetry, pipenv, uv
- **Go**: `go get`, `go install`
- **Rust**: `cargo add`, `cargo install`
- **PHP**: `composer require`
- **Java/Maven**: `mvn ... -Dartifact=...`

#### Example Detection Pattern (npm)
```bash
# Detect: npm install lodash@4.17.21
if echo "$command" | grep -qE '^npm (install|i|add) ([^@]+)@([0-9.]+)'; then
  package_name=$(echo "$command" | sed -E 's/^npm (install|i|add) ([^@]+)@.*/\2/')
  version=$(echo "$command" | sed -E 's/^npm (install|i|add) [^@]+@([0-9.]+).*/\2/')

  # Query Endor Labs
  result=$(endorctl api create -r QueryMalware -n oss \
    -d "{\"spec\":{\"package_version_names\":{\"names\":[\"npm://${package_name}@${version}\"]}}}")

  if echo "$result" | jq -e '.status == "MALWARE"' > /dev/null; then
    echo "{\"permission\":\"deny\",\"userMessage\":\"Malware detected: npm://${package_name}@${version}\"}"
    exit 0
  fi
fi
```

**State Sharing Pattern:**
```bash
# afterFileEdit: Write detections to file
generation_id=$(echo "$input" | jq -r '.generation_id')
echo "npm://nyc-config@0.9.0" >> "./malware_detected_packages_${generation_id}.txt"

# stop: Read and report
generation_id=$(echo "$input" | jq -r '.generation_id')
if [ -f "./malware_detected_packages_${generation_id}.txt" ]; then
  summary=$(cat "./malware_detected_packages_${generation_id}.txt")
  echo "{\"followup_message\":\"Malware detected: $summary\"}"
  rm "./malware_detected_packages_${generation_id}.txt"
fi
```

---

### 5. StacklokLabs/cursor-hooks
**URL:** https://github.com/StacklokLabs/cursor-hooks
**Language:** Bash
**Purpose:** MCP governance with ToolHive integration
**Last Updated:** 2024+ (actively maintained)

#### Purpose
Restrict MCP calls to only servers managed by ToolHive (Stacklok's MCP platform).

#### Hook Implementation
```bash
#!/bin/bash
# stacklok-hook.sh

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')

# Query ToolHive for managed servers
allowed_servers=$(thv list | jq -r '.[].url')

# Extract MCP server URL from arguments
server_url=$(echo "$input" | jq -r '.arguments.server_url // ""')

# Check if server is ToolHive-managed
if echo "$allowed_servers" | grep -qF "$server_url"; then
  # Optional: Registry-only mode
  if [ "$THV_REGISTRY_ONLY" = "true" ]; then
    # Check if server is from configured registry
    registry_servers=$(thv list --registry-only | jq -r '.[].url')
    if echo "$registry_servers" | grep -qF "$server_url"; then
      echo '{"permission":"allow"}'
    else
      echo '{
        "permission":"deny",
        "userMessage":"Server not in approved registry",
        "agentMessage":"Contact administrator to add server to ToolHive registry"
      }'
    fi
  else
    echo '{"permission":"allow"}'
  fi
else
  echo '{
    "permission":"deny",
    "userMessage":"Unapproved MCP server",
    "agentMessage":"MCP server not managed by ToolHive"
  }'
fi
```

#### Installation Modes

**Default mode:**
```bash
./install.sh
```
Allows any ToolHive-managed server.

**Registry-only mode:**
```bash
./install.sh --registry-only
```
Only allows servers from configured ToolHive registry.

**Manual environment variable:**
```bash
THV_REGISTRY_ONLY=true ~/.cursor/hooks/stacklok-hook.sh
```

---

### 6. 1Password/cursor-hooks
**URL:** https://github.com/1Password/cursor-hooks
**Language:** Bash
**Purpose:** 1Password configuration validation
**Last Updated:** 2024+ (actively maintained)

#### Hook: 1password-validate-mounted-env-files
Validates that required 1Password resources are properly set up before commands execute.

#### Use Case
- Prevent errors from missing 1Password mounts
- Validate environment file configurations
- Ensure secrets are available before execution

---

## Use Case Categories

### Security & Compliance

| Repository | Focus | Language |
|------------|-------|----------|
| hamzafer/cursor-hooks | General security examples | Bash |
| endorlabs/cursor-hook-examples | Malware detection | Bash |
| StacklokLabs/cursor-hooks | MCP governance | Bash |
| 1Password/cursor-hooks | Secret management | Bash |

### Code Quality

| Repository | Focus | Language |
|------------|-------|----------|
| johnlindquist/cursor-hooks | Auto-formatting | TypeScript |
| DevonFulcher/py-cursor-hooks | Python patterns | Python |

### Tool Enforcement

| Repository | Pattern | Example |
|------------|---------|---------|
| hamzafer/cursor-hooks | Block git, enforce gh | Bash |
| johnlindquist/cursor-hooks | Block npm, enforce bun | TypeScript |

### Audit & Logging

| Repository | Approach | Output |
|------------|----------|--------|
| hamzafer/cursor-hooks | Universal audit hook | `/tmp/agent-audit.log` |
| endorlabs/cursor-hook-examples | Malware scan logs | `/tmp/endorctl.log` |

---

## Common Patterns

### Pattern 1: Secret Detection
```bash
# Regex patterns for common secrets
AWS_PATTERN='AKIA[0-9A-Z]{16}'
GH_PATTERN='gh[ps]_[A-Za-z0-9]{36}'
JWT_PATTERN='eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+'

if echo "$content" | grep -qE "($AWS_PATTERN|$GH_PATTERN|$JWT_PATTERN)"; then
  echo '{"permission":"deny","userMessage":"Secret detected"}'
fi
```

### Pattern 2: Package Manager Detection
```bash
# Extract package@version from various package managers
case "$command" in
  npm\ install*|npm\ i*|npm\ add*)
    package=$(echo "$command" | sed -E 's/.*npm (install|i|add) ([^@]+)@.*/\2/')
    version=$(echo "$command" | sed -E 's/.*@([0-9.]+).*/\1/')
    ;;
  pip*install*)
    package=$(echo "$command" | sed -E 's/.*pip[0-9]* install ([^=]+)==.*/\1/')
    version=$(echo "$command" | sed -E 's/.*==([0-9.]+).*/\1/')
    ;;
  cargo\ add*|cargo\ install*)
    package=$(echo "$command" | sed -E 's/.*cargo (add|install) ([^@]+)@.*/\2/')
    version=$(echo "$command" | sed -E 's/.*@([0-9.]+).*/\1/')
    ;;
esac
```

### Pattern 3: File Extension Filtering
```bash
case "$file_path" in
  *.ts|*.tsx|*.js|*.jsx)
    # Run prettier
    prettier --write "$file_path"
    ;;
  *.py)
    # Run ruff
    ruff format "$file_path"
    ;;
  *.go)
    # Run gofmt
    gofmt -w "$file_path"
    ;;
esac
```

### Pattern 4: State Sharing via Files
```bash
# Hook 1 (beforeShellExecution): Write state
generation_id=$(echo "$input" | jq -r '.generation_id')
echo "$data" > "/tmp/cursor-${generation_id}.state"

# Hook 2 (stop): Read state
generation_id=$(echo "$input" | jq -r '.generation_id')
if [ -f "/tmp/cursor-${generation_id}.state" ]; then
  data=$(cat "/tmp/cursor-${generation_id}.state")
  echo "{\"followup_message\":\"Found: $data\"}"
  rm "/tmp/cursor-${generation_id}.state"
fi
```

### Pattern 5: External API Integration
```bash
# Call external API for validation
result=$(curl -s -X POST https://api.service.com/validate \
  -H "Content-Type: application/json" \
  -d "{\"command\":\"$command\"}")

if echo "$result" | jq -e '.safe == false' > /dev/null; then
  echo '{"permission":"deny","userMessage":"API validation failed"}'
else
  echo '{"permission":"allow"}'
fi
```

---

## Language-Specific SDKs

### TypeScript (Bun)
**Package:** `cursor-hooks` (npm)
**Pros:**
- Full type safety
- Modern async/await
- Fast execution with Bun
- JSON Schema validation

**Cons:**
- Requires Bun runtime
- Larger footprint than shell scripts

### Python
**Package:** `py-cursor-hooks` (PyPI)
**Pros:**
- Pydantic validation
- Object-oriented design
- Familiar to Python developers
- CLI integration

**Cons:**
- Requires Python runtime
- Slower startup than shell scripts

### Bash
**Package:** None (raw scripts)
**Pros:**
- No dependencies
- Fast startup
- Universal availability
- Simple for basic checks

**Cons:**
- No type safety
- Complex JSON parsing with jq
- Error-prone string manipulation

---

## Testing Patterns

### Unit Testing (TypeScript)
```typescript
// test-hook.test.ts
import { test, expect } from "bun:test";
import type { BeforeShellExecutionPayload } from "cursor-hooks";

test("blocks npm commands", async () => {
  const input: BeforeShellExecutionPayload = {
    command: "npm install lodash",
    conversation_id: "test",
    generation_id: "test",
    hook_event_name: "beforeShellExecution",
    workspace_roots: ["/test"]
  };

  const proc = Bun.spawn(["bun", "run", "before-shell-execution.ts"], {
    stdin: "pipe"
  });

  proc.stdin.write(JSON.stringify(input));
  proc.stdin.end();

  const output = await new Response(proc.stdout).json();
  expect(output.permission).toBe("deny");
});
```

### Manual Testing (Bash)
```bash
#!/bin/bash
# test-hook.sh

echo "Test 1: Allow safe command"
echo '{"command":"ls -la"}' | ./hook.sh
echo ""

echo "Test 2: Block dangerous command"
echo '{"command":"rm -rf /"}' | ./hook.sh
echo ""

echo "Test 3: Validate JSON output"
output=$(echo '{"command":"safe"}' | ./hook.sh)
echo "$output" | jq -e '.permission' > /dev/null || echo "FAIL: Invalid JSON"
```

---

## Installation Patterns

### Project-Level (Recommended for Teams)
```bash
# Clone or copy hooks into project
mkdir -p .cursor/hooks
cp /path/to/hooks/*.sh .cursor/hooks/
chmod +x .cursor/hooks/*.sh

# Create .cursor/hooks.json
cat > .cursor/hooks.json <<EOF
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      { "command": "./hooks/security.sh" }
    ]
  }
}
EOF

# Commit to version control
git add .cursor/
git commit -m "Add Cursor hooks for security"
```

### User-Level (Personal Automation)
```bash
# Install to home directory
mkdir -p ~/.cursor/hooks
cp /path/to/hooks/*.sh ~/.cursor/hooks/
chmod +x ~/.cursor/hooks/*.sh

# Create ~/.cursor/hooks.json
cat > ~/.cursor/hooks.json <<EOF
{
  "version": 1,
  "hooks": {
    "afterFileEdit": [
      { "command": "~/.cursor/hooks/format.sh" }
    ]
  }
}
EOF
```

---

## Performance Considerations

### Fast Hooks (< 100ms)
- Simple regex checks
- File extension filtering
- Local state reads

### Medium Hooks (100ms - 1s)
- External command execution (formatters)
- Small file operations
- JSON parsing with jq

### Slow Hooks (1s+)
- API calls to external services
- Package manager operations
- Complex file scanning

**Recommendation:** Keep hooks under 2 seconds to avoid UX lag.

---

## References

- hamzafer/cursor-hooks: https://github.com/hamzafer/cursor-hooks
- johnlindquist/cursor-hooks: https://github.com/johnlindquist/cursor-hooks
- DevonFulcher/py-cursor-hooks: https://github.com/DevonFulcher/py-cursor-hooks
- endorlabs/cursor-hook-examples: https://github.com/endorlabs/cursor-hook-examples
- StacklokLabs/cursor-hooks: https://github.com/StacklokLabs/cursor-hooks
- 1Password/cursor-hooks: https://github.com/1Password/cursor-hooks
