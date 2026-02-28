# Gemini - Architecture & Internals

**Last Updated:** 2026-02-21
**Version:** Gemini CLI v0.26.0+

---

## Overview

Gemini provides multiple automation architectures depending on the product area:

1. **Gemini CLI Hooks** - Synchronous middleware execution
2. **Function Calling (API)** - Callback-based tool execution
3. **Streaming APIs** - Event-driven real-time processing
4. **Batch Processing** - Asynchronous large-scale automation
5. **Agent Mode** - Multi-step autonomous workflows

This document focuses primarily on the **Gemini CLI Hooks architecture**, with references to other automation mechanisms.

**Sources:**
- [Gemini CLI Hooks Documentation](https://geminicli.com/docs/hooks/)
- [Gemini API Reference](https://ai.google.dev/api)
- [Live API Documentation](https://ai.google.dev/gemini-api/docs/live)

---

## 1. Gemini CLI Hooks Architecture

### Execution Model

**Synchronous Middleware Pattern:**

```
User Input
    ↓
┌─────────────────────────────────────┐
│ Hook Event Fires (e.g., BeforeTool)│
├─────────────────────────────────────┤
│ CLI finds matching hooks            │
│ (based on matcher patterns)         │
└──────────────┬──────────────────────┘
               ↓
    ┌──────────────────────┐
    │ Execute Hook 1       │ ←── Command or Plugin
    │ (synchronously)      │
    └──────────┬───────────┘
               ↓
    Read stdin (JSON context)
               ↓
    Execute logic
               ↓
    Write stdout (JSON decision)
               ↓
    ┌──────────────────────┐
    │ CLI evaluates result │
    └──────────┬───────────┘
               ↓
    ┌──────────────────────────────────┐
    │ allow: Continue normally         │
    │ deny: Block operation, show msg  │
    │ continue: Apply modifications    │
    └──────────────────────────────────┘
               ↓
    Next hook (if multiple)
               ↓
    Proceed with operation (if allowed)
```

**Key Characteristics:**

1. **Synchronous Execution** - CLI waits for all hooks to complete
2. **Sequential Processing** - Hooks execute in order (project → user → extension)
3. **Blocking** - Slow hooks slow down CLI operations
4. **Timeout Protected** - Configurable timeout per hook (default varies)
5. **Error Handling** - Hook failures logged but don't crash CLI

**Sources:**
- [Hooks Reference - Execution Model](https://geminicli.com/docs/hooks/reference/)

---

### Hook Lifecycle

**Detailed Lifecycle Stages:**

#### 1. Registration Phase (CLI Startup)

```
CLI starts
    ↓
Load .gemini/settings.json (project)
    ↓
Load ~/.gemini/settings.json (user)
    ↓
Load extension hooks
    ↓
Merge configurations (precedence: project > user > extension)
    ↓
Validate hook configurations
    ↓
Register hooks in hook registry
```

**Configuration Precedence:**
1. **Project settings:** `.gemini/settings.json` (highest priority)
2. **User settings:** `~/.gemini/settings.json`
3. **Extension settings:** Installed extensions (lowest priority)

**Merge Strategy:**
- Hooks with same event type are combined (not replaced)
- All matching hooks execute sequentially
- Order: project hooks → user hooks → extension hooks

#### 2. Event Detection Phase

```
CLI operation triggered
    ↓
Determine event type (e.g., BeforeTool)
    ↓
Lookup hooks registered for this event
    ↓
Apply matcher filtering
    ↓
Build list of hooks to execute
```

**Matcher Evaluation:**
- Tool events: Regular expression matching (e.g., `write_.*`)
- Lifecycle events: Exact string matching
- Wildcards: `*` or `""` (empty) matches all

#### 3. Execution Phase

```
For each matching hook:
    ↓
Prepare JSON context (event data)
    ↓
Start timer (timeout protection)
    ↓
Execute hook (command or plugin)
    ↓
    ┌─────────────────────────┐
    │ Command Hook:           │
    │ - Spawn process         │
    │ - Write JSON to stdin   │
    │ - Read JSON from stdout │
    │ - Capture stderr (logs) │
    └─────────────────────────┘
    OR
    ┌─────────────────────────┐
    │ Plugin Hook:            │
    │ - Load npm package      │
    │ - Inject services       │
    │ - Call hook method      │
    │ - Return result         │
    └─────────────────────────┘
    ↓
Parse hook output (JSON)
    ↓
Validate decision format
    ↓
Apply decision logic
```

**Timeout Handling:**
- Hook exceeds timeout → Kill process/abort
- Log timeout error
- Treat as "allow" (fail-open for safety)

**Error Handling:**
- Invalid JSON output → Log error, treat as "allow"
- Hook crashes → Log error, treat as "allow"
- Non-zero exit code → Check output for decision (exit code alone doesn't deny)

#### 4. Decision Application Phase

```
Hook returns decision
    ↓
┌──────────────────────────────────┐
│ Decision Type:                   │
├──────────────────────────────────┤
│ allow:                           │
│   - Proceed unmodified           │
│   - Execute next hook (if any)   │
├──────────────────────────────────┤
│ deny:                            │
│   - Block operation              │
│   - Show systemMessage to user   │
│   - Skip remaining hooks         │
│   - Return to user               │
├──────────────────────────────────┤
│ continue:                        │
│   - Apply modifications          │
│   - Merge context/params         │
│   - Execute next hook (if any)   │
└──────────────────────────────────┘
    ↓
If all hooks allow/continue:
    Continue with operation
```

**Decision Precedence:**
- First `deny` blocks operation (short-circuit)
- Multiple `continue` decisions merge changes
- Final `allow` after `continue` series proceeds with modifications

---

### Hook Types

Gemini supports two hook types:

#### 1. Command Hooks

**Architecture:**

```
┌────────────────────────────────┐
│ CLI                            │
│  ↓                             │
│  Spawn child process           │
│  (bash, python, node, etc.)    │
│  ↓                             │
│  Write JSON to stdin           │
│  ↓                             │
│  Wait for completion           │
│  (with timeout)                │
│  ↓                             │
│  Read JSON from stdout         │
│  ↓                             │
│  Read logs from stderr         │
│  ↓                             │
│  Parse and apply decision      │
└────────────────────────────────┘
```

**Communication Protocol:**
- **Input:** JSON via stdin
- **Output:** JSON via stdout (decision)
- **Logs:** Text via stderr (logging/debugging)

**Example Configuration:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "security-check",
        "matcher": "write_.*",
        "command": "bash",
        "args": ["/path/to/script.sh"],
        "timeout": 5000,
        "enabled": true
      }
    ]
  }
}
```

**Supported Languages:**
- Bash/Zsh (via shebang or explicit command)
- Python (via `python3` or shebang)
- Node.js (via `node` or shebang)
- Any executable with JSON stdin/stdout support

#### 2. Plugin Hooks

**Architecture:**

```
┌────────────────────────────────┐
│ CLI discovers plugins          │
│ (npm packages with             │
│  geminicli-plugin keyword)     │
│  ↓                             │
│  Validate API version          │
│  ↓                             │
│  Load plugin module            │
│  ↓                             │
│  Inject services:              │
│  - Logger                      │
│  - Config                      │
│  - HttpClient                  │
│  ↓                             │
│  Call hook method              │
│  (e.g., beforeTool())          │
│  ↓                             │
│  Receive result object         │
│  ↓                             │
│  Apply decision                │
└────────────────────────────────┘
```

**Plugin Structure:**
```javascript
// package.json
{
  "name": "my-gemini-plugin",
  "keywords": ["geminicli-plugin"],
  "geminicli": {
    "apiVersion": "1.0"
  }
}

// index.js
module.exports = {
  beforeTool: async (context, services) => {
    const { tool } = context;
    const { logger, config } = services;

    logger.info(`Validating tool: ${tool.name}`);

    if (shouldBlock(tool)) {
      return {
        decision: 'deny',
        systemMessage: 'Tool blocked by policy'
      };
    }

    return { decision: 'allow' };
  }
};
```

**Injected Services:**
- **Logger:** Structured logging interface
- **Config:** Access to CLI configuration
- **HttpClient:** Pre-configured HTTP client

**Advantages over Command Hooks:**
- Better performance (no process spawning)
- Access to injected services
- Easier debugging
- TypeScript support
- npm package distribution

---

### Performance Considerations

#### Execution Timing

**Synchronous Nature:**
- Hooks block CLI operations
- All matching hooks execute sequentially
- Total delay = sum of all hook execution times

**Performance Budget:**
```
Simple hook: < 100ms (ideal)
Moderate hook: < 500ms (acceptable)
Heavy hook: < 2000ms (problematic)
```

**Timeout Configuration:**
```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "fast-check",
        "timeout": 1000,  // 1 second
        "hooks": [...]
      },
      {
        "name": "slow-validation",
        "timeout": 10000,  // 10 seconds (avoid)
        "hooks": [...]
      }
    ]
  }
}
```

#### Optimization Strategies

**1. Use Matcher Filtering:**
```json
{
  "BeforeTool": [
    {
      "matcher": "write_file",  // Specific tool only
      "hooks": [...]
    }
  ]
}
```

**2. Cache Expensive Operations:**
```python
import json
import sys
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_validation(tool_name):
    # Cache results
    return validate(tool_name)

input_data = json.load(sys.stdin)
result = expensive_validation(input_data['tool']['name'])
```

**3. Use Plugin Hooks for Performance:**
- Avoid process spawning overhead
- Reuse services and connections
- Better caching opportunities

**4. Defer Non-Critical Work:**
```bash
#!/bin/bash
# Fire and forget for non-critical logging
(
  # Background task
  log_to_analytics "$INPUT" &
)

# Immediate response
echo '{"decision": "allow"}'
```

**5. Optimize I/O:**
- Avoid file operations in hot path
- Use in-memory caching
- Batch database operations

---

### Security Model

#### Execution Environment

**Permissions:**
- Hooks run with **CLI user's permissions**
- No sandboxing or isolation by default
- Full filesystem and network access

**Security Implications:**
- Hooks can read sensitive files
- Hooks can make network requests
- Hooks can execute arbitrary code
- **Trust model:** User must trust hook code

#### Security Best Practices

**1. Validate Hook Sources:**
```bash
# Only use hooks from trusted sources
# Review code before enabling
cat .gemini/hooks/security-check.sh
```

**2. Use Read-Only Operations When Possible:**
```python
# Prefer read-only checks
def is_safe_path(path):
    return not path.startswith(('/etc', '/sys', '/.ssh'))
```

**3. Audit Hook Execution:**
```json
{
  "hooks": {
    "AfterTool": [
      {
        "matcher": "*",
        "hooks": [{
          "type": "command",
          "command": "bash",
          "args": ["audit-log.sh"]
        }]
      }
    ]
  }
}
```

**4. Limit Hook Capabilities:**
- Use minimal permissions
- Avoid sudo/elevated access
- Restrict network access if not needed

**Sources:**
- [Hooks Best Practices - Security](https://geminicli.com/docs/hooks/best-practices/)

---

## 2. Function Calling Architecture (API)

### Execution Flow

```
User Query
    ↓
┌────────────────────────────────┐
│ Gemini API                     │
│ (model analyzes query)         │
└────────────┬───────────────────┘
             ↓
    Determines function needed
             ↓
┌────────────────────────────────┐
│ FunctionCall Request           │
│ {                              │
│   name: "get_weather",         │
│   args: {location: "SF"}       │
│ }                              │
└────────────┬───────────────────┘
             ↓
┌────────────────────────────────┐
│ Your Code (callback)           │
│ - Receives function call       │
│ - Executes actual function     │
│ - Returns result               │
└────────────┬───────────────────┘
             ↓
    Send result to API
             ↓
┌────────────────────────────────┐
│ Gemini API                     │
│ (generates final response      │
│  using function result)        │
└────────────┬───────────────────┘
             ↓
    Final Response to User
```

**Two Modes:**

1. **Manual Mode** - Developer controls execution
2. **Automatic Mode** - SDK auto-executes (Python only)

**Sources:**
- [Function Calling Documentation](https://ai.google.dev/gemini-api/docs/function-calling)

---

## 3. Streaming Architecture

### Server-Sent Events (SSE)

```
Client Request
    ↓
┌────────────────────────────────┐
│ Gemini API                     │
│ (streamGenerateContent)        │
└────────────┬───────────────────┘
             ↓
    HTTP/2 SSE stream
             ↓
┌────────────────────────────────┐
│ Client receives chunks         │
│ Chunk 1: "Hello"               │
│ Chunk 2: " world"              │
│ Chunk 3: "!"                   │
│ ...                            │
└────────────────────────────────┘
```

**Protocol:** HTTP/2 with Server-Sent Events

### WebSocket (Live API)

```
Client
    ↓
WebSocket Connection
    ↓
┌────────────────────────────────┐
│ Live API                       │
│ (bi-directional)               │
└────────────┬───────────────────┘
             ↓
    ┌──────────────┐
    │ Callbacks:   │
    │ - onopen     │
    │ - onmessage  │
    │ - onerror    │
    │ - onclose    │
    └──────────────┘
```

**Features:**
- Bi-directional communication
- Audio/video streaming
- Real-time function calling
- Low-latency responses

**Sources:**
- [Live API Documentation](https://ai.google.dev/gemini-api/docs/live)

---

## 4. Batch Processing Architecture

### Asynchronous Processing Model

```
Prepare JSONL file
    ↓
┌────────────────────────────────┐
│ Upload batch (up to 2GB)       │
└────────────┬───────────────────┘
             ↓
    Batch API queues job
             ↓
┌────────────────────────────────┐
│ Asynchronous Processing        │
│ (50% cost discount)            │
│ (~24 hour turnaround)          │
└────────────┬───────────────────┘
             ↓
    Poll for completion
             ↓
┌────────────────────────────────┐
│ Download results (JSONL)       │
└────────────────────────────────┘
```

**Advantages:**
- **50% cost savings**
- Handles large datasets (up to 2GB)
- Asynchronous - no waiting
- Multimodal support

**Sources:**
- [Batch API Documentation](https://ai.google.dev/gemini-api/docs/batch-api)

---

## 5. Agent Mode Architecture

### Multi-Step Workflow

```
User Request
    ↓
┌────────────────────────────────┐
│ Agent analyzes task            │
│ (understand scope)             │
└────────────┬───────────────────┘
             ↓
    Create execution plan
             ↓
┌────────────────────────────────┐
│ Show plan to user              │
│ (approval required)            │
└────────────┬───────────────────┘
             ↓
    User approves
             ↓
┌────────────────────────────────┐
│ Execute plan:                  │
│ - Step 1: Read files           │
│ - Step 2: Analyze dependencies │
│ - Step 3: Modify File A        │
│ - Step 4: Modify File B        │
│ - Step 5: Run tests            │
└────────────┬───────────────────┘
             ↓
    Show results for review
             ↓
    User accepts/modifies
```

**Key Features:**
- Multi-file awareness
- Planning phase with approval
- Built-in tools + MCP servers
- Human oversight at key points

**Sources:**
- [Agent Mode Documentation](https://developers.google.com/gemini-code-assist/docs/agent-mode)

---

## Performance Comparison

| Approach | Latency | Throughput | Cost | Use Case |
|----------|---------|------------|------|----------|
| **CLI Hooks** | Low (synchronous) | Medium | Free | Real-time validation |
| **Function Calling** | Medium | High | Standard API | Tool integration |
| **SSE Streaming** | Low (first token) | High | Standard API | Interactive UIs |
| **WebSocket** | Very Low | Very High | Standard API | Real-time audio/video |
| **Batch** | High (~24h) | Very High | **50% discount** | Large-scale offline |
| **Agent Mode** | High (multi-step) | Low | Code Assist | Complex workflows |

---

## See Also

- [Event Types & Triggers](./events-reference.md) - Hook event catalog
- [Configuration & Setup](./configuration.md) - Setup instructions
- [Scripting & Execution](./scripting.md) - Writing hooks
- [Security & Safety](./security.md) - Security patterns

---

**Document Version:** 1.0
**Last Updated:** 2026-02-21
**Sources:** Official documentation, research analysis
