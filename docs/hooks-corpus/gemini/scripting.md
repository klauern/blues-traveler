# Gemini - Scripting & Execution

**Last Updated:** 2026-02-21
**Version:** Gemini CLI v0.26.0+

---

## Overview

Gemini CLI hooks can be implemented in **any programming language** that supports:
1. Reading JSON from stdin
2. Writing JSON to stdout
3. Executable permissions (for command hooks)

This document covers language support, execution models, and best practices for writing hooks.

**Sources:**
- [Writing Hooks Documentation](https://geminicli.com/docs/hooks/writing-hooks/)
- [Hooks Best Practices](https://geminicli.com/docs/hooks/best-practices/)

---

## Supported Languages

### Native Support (Command Hooks)

**Any language with stdin/stdout:** All programming languages supported via shebang or explicit command specification.

**Common Languages:**

| Language | Execution Method | Example |
|----------|------------------|---------|
| **Bash/Shell** | Shebang or `bash` | `#!/bin/bash` |
| **Python** | Shebang or `python3` | `#!/usr/bin/env python3` |
| **Node.js** | Shebang or `node` | `#!/usr/bin/env node` |
| **Ruby** | Shebang or `ruby` | `#!/usr/bin/env ruby` |
| **Go** | Compiled binary | `./hook-binary` |
| **Rust** | Compiled binary | `./hook-binary` |
| **Perl** | Shebang or `perl` | `#!/usr/bin/env perl` |
| **PHP** | Shebang or `php` | `#!/usr/bin/env php` |
| **Deno** | `deno run` | `deno run --allow-all hook.ts` |

### Plugin Hooks (JavaScript/TypeScript Only)

**npm Package Format:**
```json
{
  "name": "my-gemini-plugin",
  "keywords": ["geminicli-plugin"],
  "geminicli": {
    "apiVersion": "1.0"
  }
}
```

**Implementation:**
```typescript
// TypeScript with types
export interface HookContext {
  event: string;
  sessionId: string;
  tool?: {
    name: string;
    params: Record<string, any>;
  };
}

export interface HookServices {
  logger: Logger;
  config: Config;
  httpClient: HttpClient;
}

export async function beforeTool(
  context: HookContext,
  services: HookServices
): Promise<HookDecision> {
  const { tool } = context;
  const { logger } = services;

  logger.info(`Validating tool: ${tool?.name}`);

  if (shouldBlock(tool)) {
    return {
      decision: 'deny',
      systemMessage: 'Tool blocked'
    };
  }

  return { decision: 'allow' };
}
```

---

## Language-Specific Examples

### Bash/Shell

**Minimal Hook:**
```bash
#!/bin/bash
set -euo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Parse with jq
TOOL=$(echo "$INPUT" | jq -r '.tool.name // "unknown"')
FILE=$(echo "$INPUT" | jq -r '.tool.params.path // empty')

# Validation logic
if [[ "$FILE" =~ \.env$ ]]; then
  echo '{"decision": "deny", "systemMessage": "Cannot modify .env files"}'
  exit 0
fi

echo '{"decision": "allow"}'
```

**With Logging:**
```bash
#!/bin/bash
set -euo pipefail

INPUT=$(cat)

# Log to stderr (appears in CLI logs)
echo "Hook executing at $(date)" >&2

TOOL=$(echo "$INPUT" | jq -r '.tool.name')
echo "Tool: $TOOL" >&2

# Return decision
echo '{"decision": "allow"}'
```

**Error Handling:**
```bash
#!/bin/bash
set -euo pipefail

trap 'echo "{\"decision\": \"allow\"}" && exit 0' ERR

INPUT=$(cat)

TOOL=$(echo "$INPUT" | jq -r '.tool.name // "unknown"')

# Risky operation
result=$(risky_validation "$TOOL" 2>&1) || {
  echo "Validation failed, allowing by default" >&2
  echo '{"decision": "allow"}'
  exit 0
}

echo '{"decision": "allow"}'
```

---

### Python

**Minimal Hook:**
```python
#!/usr/bin/env python3
import json
import sys

# Read JSON from stdin
input_data = json.load(sys.stdin)

# Extract fields
tool_name = input_data.get('tool', {}).get('name', '')
file_path = input_data.get('tool', {}).get('params', {}).get('path', '')

# Validation
if file_path.endswith('.env'):
    output = {
        'decision': 'deny',
        'systemMessage': 'Cannot modify .env files'
    }
else:
    output = {'decision': 'allow'}

# Write JSON to stdout
print(json.dumps(output))
```

**With Logging:**
```python
#!/usr/bin/env python3
import json
import sys
import logging

# Configure logging to stderr
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    stream=sys.stderr
)

logger = logging.getLogger(__name__)

try:
    input_data = json.load(sys.stdin)
    tool = input_data.get('tool', {})

    logger.info(f"Validating tool: {tool.get('name')}")

    # Validation logic
    if should_block(tool):
        logger.warning(f"Blocking tool: {tool.get('name')}")
        output = {
            'decision': 'deny',
            'systemMessage': 'Tool blocked by policy'
        }
    else:
        output = {'decision': 'allow'}

except Exception as e:
    logger.error(f"Hook error: {e}")
    output = {'decision': 'allow'}  # Fail open

print(json.dumps(output))
```

**With External Libraries:**
```python
#!/usr/bin/env python3
import json
import sys
import re
from pathlib import Path

def is_sensitive_file(path: str) -> bool:
    """Check if file is sensitive"""
    sensitive_patterns = [
        r'\.env$',
        r'credentials',
        r'secrets',
        r'\.ssh/',
        r'\.aws/'
    ]

    return any(re.search(pattern, path) for pattern in sensitive_patterns)

def main():
    input_data = json.load(sys.stdin)

    tool = input_data.get('tool', {})
    params = tool.get('params', {})
    file_path = params.get('path', '')

    if is_sensitive_file(file_path):
        output = {
            'decision': 'deny',
            'systemMessage': f'Blocked: {file_path} is a sensitive file'
        }
    else:
        output = {'decision': 'allow'}

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

---

### Node.js/JavaScript

**Minimal Hook:**
```javascript
#!/usr/bin/env node

const fs = require('fs');

// Read stdin
const input = JSON.parse(fs.readFileSync(0, 'utf-8'));

// Extract fields
const tool = input.tool || {};
const filePath = tool.params?.path || '';

// Validation
if (filePath.endsWith('.env')) {
  console.log(JSON.stringify({
    decision: 'deny',
    systemMessage: 'Cannot modify .env files'
  }));
} else {
  console.log(JSON.stringify({decision: 'allow'}));
}
```

**Async Hook:**
```javascript
#!/usr/bin/env node

const fs = require('fs');

async function main() {
  // Read stdin
  const chunks = [];
  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }
  const input = JSON.parse(Buffer.concat(chunks).toString('utf8'));

  // Async validation
  const isValid = await validateTool(input.tool);

  if (!isValid) {
    console.log(JSON.stringify({
      decision: 'deny',
      systemMessage: 'Validation failed'
    }));
  } else {
    console.log(JSON.stringify({decision: 'allow'}));
  }
}

async function validateTool(tool) {
  // Async operation (e.g., API call, database query)
  return new Promise(resolve => {
    setTimeout(() => resolve(true), 100);
  });
}

main().catch(err => {
  console.error('Error:', err);
  console.log(JSON.stringify({decision: 'allow'}));
  process.exit(0);
});
```

**TypeScript:**
```typescript
#!/usr/bin/env node

import * as fs from 'fs';

interface HookInput {
  event: string;
  tool?: {
    name: string;
    params: Record<string, any>;
  };
}

interface HookOutput {
  decision: 'allow' | 'deny' | 'continue';
  systemMessage?: string;
}

function readStdin(): Promise<HookInput> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    process.stdin.on('data', chunk => chunks.push(chunk));
    process.stdin.on('end', () => {
      try {
        const data = Buffer.concat(chunks).toString('utf8');
        resolve(JSON.parse(data));
      } catch (e) {
        reject(e);
      }
    });
  });
}

async function main() {
  const input = await readStdin();

  const output: HookOutput = shouldBlock(input)
    ? { decision: 'deny', systemMessage: 'Blocked' }
    : { decision: 'allow' };

  console.log(JSON.stringify(output));
}

function shouldBlock(input: HookInput): boolean {
  // Validation logic
  return false;
}

main().catch(err => {
  console.error('Error:', err);
  console.log(JSON.stringify({decision: 'allow'}));
  process.exit(0);
});
```

---

### Go (Compiled)

**Complete Hook:**
```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type HookInput struct {
	Event string `json:"event"`
	Tool  *Tool  `json:"tool,omitempty"`
}

type Tool struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

type HookOutput struct {
	Decision      string `json:"decision"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

func main() {
	var input HookInput

	// Read stdin
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding input: %v\n", err)
		output := HookOutput{Decision: "allow"}
		json.NewEncoder(os.Stdout).Encode(output)
		return
	}

	// Validation
	if shouldBlock(&input) {
		output := HookOutput{
			Decision:      "deny",
			SystemMessage: "Blocked by security policy",
		}
		json.NewEncoder(os.Stdout).Encode(output)
		return
	}

	output := HookOutput{Decision: "allow"}
	json.NewEncoder(os.Stdout).Encode(output)
}

func shouldBlock(input *HookInput) bool {
	if input.Tool == nil {
		return false
	}

	// Check for sensitive files
	if path, ok := input.Tool.Params["path"].(string); ok {
		return strings.HasSuffix(path, ".env") ||
			strings.Contains(path, "credentials")
	}

	return false
}
```

**Build and Use:**
```bash
# Build
go build -o hook-validate hook.go

# Make executable
chmod +x hook-validate

# Configure
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "./hook-validate"
          }
        ]
      }
    ]
  }
}
```

---

## Exit Codes & Control Flow

### Exit Code Behavior

**Important:** Exit codes **do NOT determine decision** - only JSON output matters.

```bash
#!/bin/bash

# ❌ WRONG - Exit code doesn't deny
exit 1

# ✅ CORRECT - JSON output denies
echo '{"decision": "deny"}'
exit 0
```

**Best Practice:**
- Always exit 0 (success)
- Use JSON output to control decision
- Reserve non-zero exits for crashes/errors

```bash
#!/bin/bash
set -euo pipefail

trap 'echo "{\"decision\": \"allow\"}" && exit 0' ERR

# Your logic here

echo '{"decision": "allow"}'
exit 0
```

---

## Decision Types

### Allow

**Continue normally without modifications:**
```json
{"decision": "allow"}
```

**Use when:**
- Validation passes
- No modifications needed
- Default behavior desired

### Deny

**Block operation and show message:**
```json
{
  "decision": "deny",
  "systemMessage": "Blocked: Cannot write to .env files"
}
```

**Use when:**
- Validation fails
- Operation violates policy
- Security check fails

**Requirements:**
- `systemMessage` should explain why blocked
- User-friendly message

### Continue

**Modify and proceed:**
```json
{
  "decision": "continue",
  "context": "additional context to inject",
  "tool": {
    "name": "write_file",
    "params": {
      "path": "/modified/path",
      "content": "modified content"
    }
  }
}
```

**Use when:**
- Need to modify parameters
- Injecting context
- Sanitizing inputs

**Modifiable Fields (Event-Specific):**
- `context` - Add/modify context (BeforeAgent)
- `tool.params` - Modify tool parameters (BeforeTool)
- `availableTools` - Filter tools (BeforeToolSelection)
- `modelInput` - Modify model input (BeforeModel)
- `modelOutput` - Filter response (AfterModel)

---

## Error Handling

### Best Practices

**1. Fail Open (Default Allow):**
```python
try:
    # Risky validation
    result = validate_operation()
    if not result:
        output = {'decision': 'deny'}
    else:
        output = {'decision': 'allow'}
except Exception as e:
    logger.error(f"Validation error: {e}")
    # Fail open - allow by default
    output = {'decision': 'allow'}

print(json.dumps(output))
```

**2. Provide Useful Error Messages:**
```bash
#!/bin/bash
INPUT=$(cat)

RESULT=$(validate_tool "$INPUT" 2>&1) || {
  ERROR_MSG=$(echo "$RESULT" | jq -R -s .)
  echo "{\"decision\": \"deny\", \"systemMessage\": \"Validation error: $ERROR_MSG\"}"
  exit 0
}

echo '{"decision": "allow"}'
```

**3. Log Errors to stderr:**
```python
import sys
import logging

logging.basicConfig(stream=sys.stderr, level=logging.ERROR)

try:
    # Logic
    pass
except Exception as e:
    logging.error(f"Hook error: {e}", exc_info=True)
    print(json.dumps({'decision': 'allow'}))
```

---

## Timeouts & Resource Limits

### Timeout Configuration

**Default Timeout:** 5000ms (5 seconds)

**Configure Per Hook:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "*",
        "timeout": 10000,  // 10 seconds
        "hooks": [...]
      }
    ]
  }
}
```

**Timeout Behavior:**
- Hook exceeds timeout → Process killed
- Treated as `allow` (fail open)
- Error logged to CLI logs

**Best Practices:**
- Keep hooks under 1-2 seconds
- Set realistic timeouts
- Avoid network calls in critical path

### Resource Management

**1. Avoid Long-Running Operations:**
```python
# ❌ BAD - Slow synchronous validation
def validate():
    time.sleep(10)  # Long operation
    return True

# ✅ GOOD - Fast validation with caching
@lru_cache(maxsize=128)
def validate(tool_name):
    # Fast lookup
    return tool_name in allowed_tools
```

**2. Use Background Tasks for Non-Critical Work:**
```bash
#!/bin/bash
INPUT=$(cat)

# Fire and forget for logging
(
  # Background task
  echo "$INPUT" | log_to_external_system &
)

# Immediate response
echo '{"decision": "allow"}'
```

---

## Async & Parallel Execution

### Sequential Execution

All hooks execute **sequentially** (one after another):

```
Hook 1 executes (3s)
    ↓
Hook 2 executes (2s)
    ↓
Hook 3 executes (1s)
    ↓
Total: 6 seconds
```

**No parallel execution currently supported.**

### Async Within Hooks

**Python asyncio:**
```python
#!/usr/bin/env python3
import asyncio
import json
import sys

async def validate_async(tool):
    # Async operations
    result = await external_api_call(tool)
    return result

async def main():
    input_data = json.load(sys.stdin)

    is_valid = await validate_async(input_data.get('tool'))

    output = {
        'decision': 'allow' if is_valid else 'deny'
    }

    print(json.dumps(output))

if __name__ == '__main__':
    asyncio.run(main())
```

**Node.js async:**
```javascript
#!/usr/bin/env node

async function main() {
  const input = await readStdin();

  // Parallel async operations
  const [result1, result2] = await Promise.all([
    checkSecurity(input),
    checkCompliance(input)
  ]);

  const output = result1 && result2
    ? {decision: 'allow'}
    : {decision: 'deny'};

  console.log(JSON.stringify(output));
}

main();
```

---

## Debugging Hooks

### Enable Debug Logging

**Configuration:**
```json
{
  "options": {
    "logLevel": "debug"
  }
}
```

**Hook Logging:**
```bash
#!/bin/bash
INPUT=$(cat)

# Log to stderr (appears in CLI debug logs)
echo "Hook started at $(date)" >&2
echo "Input: $INPUT" >&2

# Your logic

echo "Output: allow" >&2
echo '{"decision": "allow"}'
```

### Test Hooks Locally

**Prepare test input:**
```bash
cat > test-input.json << 'EOF'
{
  "event": "BeforeTool",
  "tool": {
    "name": "write_file",
    "params": {
      "path": "/test/file.txt",
      "content": "test"
    }
  }
}
EOF
```

**Test hook:**
```bash
cat test-input.json | .gemini/hooks/your-hook.sh

# Should output valid JSON
# {"decision": "allow"}
```

**Validate JSON output:**
```bash
cat test-input.json | .gemini/hooks/your-hook.sh | jq .

# Should parse without errors
```

---

## Best Practices

### Performance

1. **Keep hooks fast** (< 1 second ideal)
2. **Cache expensive operations**
3. **Avoid network calls** in critical path
4. **Use compiled languages** for performance-critical hooks

### Reliability

1. **Always output valid JSON**
2. **Fail open** (default to `allow` on errors)
3. **Handle missing fields** gracefully
4. **Log errors** to stderr

### Security

1. **Validate all inputs**
2. **Sanitize outputs**
3. **Never trust tool parameters**
4. **Log security decisions**

### Maintainability

1. **Add comments** explaining logic
2. **Use consistent formatting**
3. **Version your hooks**
4. **Test thoroughly**

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Hook events
- [Environment & Context](./environment-context.md) - Available context
- [Security & Safety](./security.md) - Security patterns
- [Examples](./examples.md) - Ready-to-use hooks

---

**Document Version:** 1.0
**Last Updated:** 2026-02-21
**Sources:** Official documentation, research analysis
