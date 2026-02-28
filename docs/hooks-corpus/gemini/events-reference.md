# Gemini CLI - Complete Event Types & Hooks Reference

**Version:** Gemini CLI v0.26.0+
**Last Updated:** 2026-02-21
**Event Count:** 13+ distinct hook events

---

## Overview

Gemini CLI provides **the most comprehensive hook system** among AI coding assistants, with 13+ distinct event types covering the entire CLI lifecycle. Hooks fire synchronously at specific points, allowing you to:

- **Observe** operations (logging, telemetry)
- **Modify** inputs/outputs (context injection, filtering)
- **Block** dangerous operations (security, validation)
- **Trigger** follow-up actions (notifications, workflows)

**Sources:**
- [Gemini CLI Hooks Documentation](https://geminicli.com/docs/hooks/)
- [Hooks Reference](https://geminicli.com/docs/hooks/reference/)
- [Writing Hooks](https://geminicli.com/docs/hooks/writing-hooks/)

---

## Complete Event Catalog

| Event Name | Trigger | Timing | Cancellable | Context Available | Flow Control |
|------------|---------|--------|-------------|-------------------|--------------|
| **SessionStart** | CLI starts/resumes/clear | Session init | ❌ | Session info | Continue only |
| **BeforeAgent** | Before LLM request | Pre-agent | ✅ | Full request | Allow/Deny/Continue |
| **BeforeToolSelection** | Before tool choice | Tool planning | ✅ | Available tools, query | Allow/Deny/Continue |
| **BeforeTool** | Before tool execution | Pre-tool | ✅ | Tool name, params | Allow/Deny/Continue |
| **AfterTool** | After tool success | Post-tool | ❌ | Tool result | Continue only |
| **BeforeModel** | Before model processing | Pre-inference | ✅ | Model input | Allow/Deny/Continue |
| **AfterModel** | After LLM response | Post-inference | ❌ | Model output | Continue only |
| **BeforeResponse** | Each response chunk | Streaming | ❌ | Chunk data | Continue only |
| **AfterChunk** | Each response chunk | Streaming | ❌ | Chunk data | **Limited*** |
| **AfterAgent** | After agent turn | Turn completion | ❌ | Turn summary | Continue only |
| **SessionEnd** | CLI exits/cleared | Session end | ❌ | Session stats | Continue only |
| **Notification** | Needs user attention | Async | ❌ | Notification type | Continue only |
| **BeforeCompress** | Before context compression | Async | ❌** | Compression info | **Ignored** |
| **AfterAlert** | After alert shown | Observability | ❌** | Alert details | **Ignored** |

\* AfterChunk does not support `decision`, `continue`, or `systemMessage` fields
\*\* Flow-control fields are ignored (observability only)

---

## Event Categories

### 1. Session Lifecycle Events

#### SessionStart

**Purpose:** Initialize session state, load project context

**Trigger:**
- Application startup
- Resuming existing session
- After `/clear` command execution

**Context Available:**
```json
{
  "event": "SessionStart",
  "sessionId": "unique-session-identifier",
  "timestamp": "2026-02-21T12:00:00Z",
  "projectPath": "/path/to/project",
  "user": {
    "id": "user-identifier",
    "email": "user@example.com"
  }
}
```

**Flow Control:**
- `continue` - Proceed normally (add systemMessage to inject context)

**Common Use Cases:**
- Load project-specific context from files
- Initialize session-scoped variables
- Log session start for audit trails
- Display welcome messages or warnings

**Example Hook:**
```bash
#!/bin/bash
# Load git context at session start

GIT_STATUS=$(git status --short 2>/dev/null || echo "Not a git repo")
GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
RECENT_COMMITS=$(git log --oneline -3 2>/dev/null || echo "No commits")

CONTEXT="Git Branch: $GIT_BRANCH\n\nRecent Commits:\n$RECENT_COMMITS\n\nCurrent Status:\n$GIT_STATUS"

echo "{\"decision\": \"continue\", \"systemMessage\": \"$CONTEXT\"}"
```

**Cancellable:** ❌ No
**Can Modify:** ✅ Yes (inject context via systemMessage)

---

#### SessionEnd

**Purpose:** Cleanup, final logging, session teardown

**Trigger:**
- CLI exits normally
- Session cleared via `/clear`
- CLI crashes (best-effort)

**Context Available:**
```json
{
  "event": "SessionEnd",
  "sessionId": "unique-session-identifier",
  "timestamp": "2026-02-21T13:00:00Z",
  "stats": {
    "toolsExecuted": 42,
    "tokensUsed": 15000,
    "duration": 3600
  }
}
```

**Flow Control:**
- `continue` - Cleanup continues

**Common Use Cases:**
- Save session logs
- Clean up temporary files
- Send final telemetry
- Save session statistics

**Example Hook:**
```python
#!/usr/bin/env python3
import json
import sys
from datetime import datetime

input_data = json.load(sys.stdin)

# Log session end
with open('.gemini/session-log.jsonl', 'a') as f:
    log_entry = {
        'event': 'SessionEnd',
        'timestamp': datetime.now().isoformat(),
        'stats': input_data.get('stats', {})
    }
    f.write(json.dumps(log_entry) + '\n')

print(json.dumps({'decision': 'continue'}))
```

**Cancellable:** ❌ No
**Can Modify:** ❌ No

---

### 2. Agent/LLM Lifecycle Events

#### BeforeAgent

**Purpose:** Modify or augment prompts before sending to LLM

**Trigger:**
- Before each request to the LLM
- After user submits prompt or agent decides to call LLM again

**Context Available:**
```json
{
  "event": "BeforeAgent",
  "requestId": "unique-request-id",
  "query": "user's input prompt",
  "conversationHistory": [
    {"role": "user", "content": "previous message"},
    {"role": "assistant", "content": "previous response"}
  ],
  "context": "existing context string",
  "availableTools": [
    {"name": "write_file", "description": "..."}
  ]
}
```

**Flow Control:**
- `allow` - Send request unmodified
- `deny` - Block request, show systemMessage to user
- `continue` - Modify request (add context, change query)

**Common Use Cases:**
- Inject project-specific context
- Add git status, issue tracking data
- Append documentation snippets
- Modify unsafe prompts

**Example Hook:**
```javascript
#!/usr/bin/env node
const input = JSON.parse(require('fs').readFileSync(0, 'utf-8'));

// Add Jira context if ticket mentioned
const ticketMatch = input.query.match(/([A-Z]+-\d+)/);
if (ticketMatch) {
  const ticket = ticketMatch[1];
  const jiraData = fetchJiraTicket(ticket); // Your function

  const enhanced = {
    decision: 'continue',
    context: input.context + `\n\nJira ${ticket}:\n${jiraData}`
  };

  console.log(JSON.stringify(enhanced));
} else {
  console.log(JSON.stringify({decision: 'allow'}));
}

function fetchJiraTicket(id) {
  // Implementation
  return `Title: Example\nDescription: Details...`;
}
```

**Cancellable:** ✅ Yes
**Can Modify:** ✅ Yes (query, context)

---

#### AfterAgent

**Purpose:** Post-process agent responses, log turn completion

**Trigger:**
- After agent finishes final response for a turn
- Once per conversation turn

**Context Available:**
```json
{
  "event": "AfterAgent",
  "requestId": "unique-request-id",
  "response": "agent's full response text",
  "toolsUsed": ["write_file", "read_file"],
  "tokensUsed": 1500,
  "duration": 2.5
}
```

**Flow Control:**
- `continue` - Normal completion (can add systemMessage)

**Common Use Cases:**
- Log completed turns for audit
- Validate response quality
- Track tool usage patterns
- Send notifications on completion

**Example Hook:**
```python
#!/usr/bin/env python3
import json
import sys

input_data = json.load(sys.stdin)

# Log agent completion
with open('.gemini/agent-log.jsonl', 'a') as f:
    json.dump({
        'event': 'AfterAgent',
        'tools': input_data.get('toolsUsed', []),
        'tokens': input_data.get('tokensUsed', 0)
    }, f)
    f.write('\n')

print(json.dumps({'decision': 'continue'}))
```

**Cancellable:** ❌ No
**Can Modify:** ✅ Yes (add systemMessage only)

---

#### BeforeModel

**Purpose:** Advanced prompt modification before model inference

**Trigger:**
- Before model processes input
- After tool selection, before actual LLM call

**Context Available:**
```json
{
  "event": "BeforeModel",
  "requestId": "unique-request-id",
  "modelInput": {
    "prompt": "full prompt including context",
    "tools": ["available", "tools"],
    "parameters": {
      "temperature": 0.7,
      "maxTokens": 2048
    }
  }
}
```

**Flow Control:**
- `allow` - Send to model unmodified
- `deny` - Block model call
- `continue` - Modify model input

**Common Use Cases:**
- Sophisticated prompt engineering
- Dynamic parameter adjustment
- Cost optimization (reduce maxTokens for simple queries)
- Advanced context injection

**Example Hook:**
```bash
#!/bin/bash
INPUT=$(cat)

# Reduce tokens for simple queries
QUERY=$(echo "$INPUT" | jq -r '.modelInput.prompt')
LENGTH=${#QUERY}

if [ $LENGTH -lt 100 ]; then
  # Simple query, reduce max tokens
  echo "$INPUT" | jq '.modelInput.parameters.maxTokens = 500 | {decision: "continue", modelInput: .modelInput}'
else
  echo '{"decision": "allow"}'
fi
```

**Cancellable:** ✅ Yes
**Can Modify:** ✅ Yes (model input, parameters)

---

#### AfterModel

**Purpose:** Real-time response filtering, PII redaction

**Trigger:**
- Immediately after LLM response chunk received
- Before response shown to user

**Context Available:**
```json
{
  "event": "AfterModel",
  "requestId": "unique-request-id",
  "modelOutput": {
    "text": "model's response text",
    "functionCalls": [
      {"name": "write_file", "params": {...}}
    ]
  }
}
```

**Flow Control:**
- `continue` - Modify response (redact PII, filter content)

**Common Use Cases:**
- Real-time PII filtering
- Redact sensitive information
- Block inappropriate content
- Log model outputs

**Example Hook:**
```python
#!/usr/bin/env python3
import json
import sys
import re

input_data = json.load(sys.stdin)
text = input_data.get('modelOutput', {}).get('text', '')

# Redact email addresses
text = re.sub(r'\b[\w.-]+@[\w.-]+\.\w+\b', '[EMAIL REDACTED]', text)

# Redact API keys (simple pattern)
text = re.sub(r'\b[A-Za-z0-9]{32,}\b', '[API_KEY REDACTED]', text)

output = {
    'decision': 'continue',
    'modelOutput': {
        **input_data.get('modelOutput', {}),
        'text': text
    }
}

print(json.dumps(output))
```

**Cancellable:** ❌ No (but can modify output)
**Can Modify:** ✅ Yes (response text, function calls)

---

#### BeforeResponse / AfterChunk

**Purpose:** Process each streaming response chunk

**Trigger:**
- For every chunk generated by the model
- Multiple times per response

**Context Available:**
```json
{
  "event": "AfterChunk",
  "requestId": "unique-request-id",
  "chunk": {
    "text": "partial response text",
    "index": 5,
    "isLast": false
  }
}
```

**Flow Control:**
- `continue` - Proceed to next chunk
- **Note:** AfterChunk does NOT support `decision`, `continue`, or `systemMessage`

**Common Use Cases:**
- Real-time logging of streaming responses
- Track response progress
- Observe model output patterns

**Example Hook:**
```javascript
#!/usr/bin/env node
const input = JSON.parse(require('fs').readFileSync(0, 'utf-8'));

// Log chunk (observability only)
console.error(`Chunk ${input.chunk.index}: ${input.chunk.text.length} chars`);

// No modification possible for AfterChunk
console.log(JSON.stringify({decision: 'continue'}));
```

**Cancellable:** ❌ No
**Can Modify:** ❌ No (observability only for AfterChunk)

---

### 3. Tool Lifecycle Events

#### BeforeToolSelection

**Purpose:** Filter available tools, optimize tool selection

**Trigger:**
- Before LLM decides which tools to call
- After understanding user intent, before tool planning

**Context Available:**
```json
{
  "event": "BeforeToolSelection",
  "requestId": "unique-request-id",
  "query": "user's request",
  "availableTools": [
    {
      "name": "write_file",
      "description": "Write content to a file",
      "expensive": false
    },
    {
      "name": "search_codebase",
      "description": "Search entire codebase",
      "expensive": true
    }
  ]
}
```

**Flow Control:**
- `allow` - Use all available tools
- `deny` - Block tool selection entirely
- `continue` - Filter tools (modify availableTools array)

**Common Use Cases:**
- Cost optimization (remove expensive tools for simple queries)
- Security (disable dangerous tools in production)
- Performance (reduce tool set for faster selection)
- Context-specific tool filtering

**Example Hook:**
```javascript
#!/usr/bin/env node
const input = JSON.parse(require('fs').readFileSync(0, 'utf-8'));

const query = input.query.toLowerCase();
const isSimple = query.length < 50 && !query.includes('analyze');

if (isSimple) {
  // Filter out expensive tools for simple queries
  const filtered = input.availableTools.filter(t => !t.expensive);

  console.log(JSON.stringify({
    decision: 'continue',
    availableTools: filtered
  }));
} else {
  console.log(JSON.stringify({decision: 'allow'}));
}
```

**Cancellable:** ✅ Yes
**Can Modify:** ✅ Yes (availableTools array)

---

#### BeforeTool

**Purpose:** Validate and potentially block tool execution

**Trigger:**
- Before any tool executes
- After LLM decides to call a tool

**Context Available:**
```json
{
  "event": "BeforeTool",
  "requestId": "unique-request-id",
  "tool": {
    "name": "write_file",
    "params": {
      "path": "/path/to/file.js",
      "content": "console.log('hello');"
    }
  }
}
```

**Flow Control:**
- `allow` - Execute tool unmodified
- `deny` - Block tool execution, show systemMessage
- `continue` - Modify tool parameters

**Common Use Cases:**
- **Security validation** (block dangerous operations)
- **Argument validation** (sanitize file paths)
- **Policy enforcement** (prevent writes to sensitive files)
- **Parameter rewriting** (add prefixes, sanitize inputs)

**Example Hook - Security Validation:**
```bash
#!/bin/bash
INPUT=$(cat)

TOOL=$(echo "$INPUT" | jq -r '.tool.name')
FILE=$(echo "$INPUT" | jq -r '.tool.params.path // empty')
CONTENT=$(echo "$INPUT" | jq -r '.tool.params.content // empty')

# Block dangerous file operations
if [[ "$FILE" =~ (\.env|credentials|secrets|\.ssh|\.aws) ]]; then
  echo '{"decision": "deny", "systemMessage": "Blocked: Cannot modify sensitive files"}'
  exit 0
fi

# Block writing secrets to files
if [[ "$CONTENT" =~ (API_KEY|SECRET|PASSWORD|TOKEN).*[A-Za-z0-9]{20,} ]]; then
  echo '{"decision": "deny", "systemMessage": "Blocked: Detected potential secrets in content"}'
  exit 0
fi

# Block dangerous shell commands
if [[ "$TOOL" == "execute_command" ]]; then
  CMD=$(echo "$INPUT" | jq -r '.tool.params.command')
  if [[ "$CMD" =~ (rm -rf /|curl.*\|.*sh|eval|sudo) ]]; then
    echo '{"decision": "deny", "systemMessage": "Blocked: Dangerous command detected"}'
    exit 0
  fi
fi

echo '{"decision": "allow"}'
```

**Example Hook - Parameter Modification:**
```python
#!/usr/bin/env python3
import json
import sys
import os

input_data = json.load(sys.stdin)

tool = input_data.get('tool', {})
if tool.get('name') == 'write_file':
    # Add project prefix to relative paths
    path = tool['params']['path']
    if not os.path.isabs(path):
        project_root = os.environ.get('GEMINI_PROJECT_DIR', '.')
        tool['params']['path'] = os.path.join(project_root, path)

    output = {
        'decision': 'continue',
        'tool': tool
    }
    print(json.dumps(output))
else:
    print(json.dumps({'decision': 'allow'}))
```

**Cancellable:** ✅ Yes
**Can Modify:** ✅ Yes (tool parameters)

**Sources:**
- [Hooks Best Practices - Security](https://geminicli.com/docs/hooks/best-practices/)

---

#### AfterTool

**Purpose:** Log results, trigger follow-up actions

**Trigger:**
- After tool executes successfully
- Before result returned to LLM

**Context Available:**
```json
{
  "event": "AfterTool",
  "requestId": "unique-request-id",
  "tool": {
    "name": "write_file",
    "params": {...}
  },
  "result": {
    "success": true,
    "output": "File written successfully",
    "duration": 0.15
  }
}
```

**Flow Control:**
- `continue` - Normal completion (can add systemMessage)

**Common Use Cases:**
- Log tool execution for audit trails
- Trigger notifications (file written, test passed)
- Update external systems
- Collect usage statistics

**Example Hook:**
```python
#!/usr/bin/env python3
import json
import sys
from datetime import datetime

input_data = json.load(sys.stdin)

# Log tool execution
with open('.gemini/tool-log.jsonl', 'a') as f:
    log_entry = {
        'timestamp': datetime.now().isoformat(),
        'tool': input_data['tool']['name'],
        'success': input_data['result']['success'],
        'duration': input_data['result']['duration']
    }
    f.write(json.dumps(log_entry) + '\n')

# Send notification for important tools
if input_data['tool']['name'] == 'deploy_application':
    # Send Slack notification
    send_slack_notification(f"Deployment completed: {input_data['result']['output']}")

print(json.dumps({'decision': 'continue'}))

def send_slack_notification(msg):
    # Implementation
    pass
```

**Cancellable:** ❌ No
**Can Modify:** ✅ Yes (add systemMessage only)

---

### 4. Notification and Alert Events

#### Notification

**Purpose:** Monitor when CLI needs user attention

**Trigger:**
- CLI is idle, waiting for user input
- Agent awaiting confirmation
- Long-running operation needs attention

**Context Available:**
```json
{
  "event": "Notification",
  "notificationType": "idle" | "confirmation" | "attention",
  "message": "Description of notification",
  "timestamp": "2026-02-21T12:30:00Z"
}
```

**Flow Control:**
- `continue` - Normal notification flow

**Common Use Cases:**
- Forward idle notifications to Slack
- Track user engagement patterns
- Send mobile push notifications
- Log attention requests

**Example Hook:**
```javascript
#!/usr/bin/env node
const input = JSON.parse(require('fs').readFileSync(0, 'utf-8'));

if (input.notificationType === 'idle') {
  // Send Slack message
  sendSlackMessage('Gemini CLI is idle and ready for input');
}

console.log(JSON.stringify({decision: 'continue'}));

function sendSlackMessage(text) {
  // Implementation
}
```

**Cancellable:** ❌ No
**Can Modify:** ❌ No (observability only)

---

#### AfterAlert

**Purpose:** Observability for alerts and permission requests

**Trigger:**
- After alert shown to user
- After permission dialog displayed
- Observability only

**Context Available:**
```json
{
  "event": "AfterAlert",
  "alertType": "permission" | "warning" | "error",
  "alertMessage": "Alert text shown to user",
  "userResponse": "allow" | "deny" | null
}
```

**Flow Control:**
- **Ignored** - Cannot block or modify alerts

**Common Use Cases:**
- Audit logging of permission requests
- Track user approval patterns
- Security monitoring
- Compliance reporting

**Example Hook:**
```python
#!/usr/bin/env python3
import json
import sys

input_data = json.load(sys.stdin)

# Log all permission requests for audit
if input_data['alertType'] == 'permission':
    with open('.gemini/permissions.log', 'a') as f:
        json.dump({
            'timestamp': input_data.get('timestamp'),
            'alert': input_data['alertMessage'],
            'response': input_data.get('userResponse')
        }, f)
        f.write('\n')

# Flow control ignored, but still output
print(json.dumps({'decision': 'continue'}))
```

**Cancellable:** ❌ No
**Can Modify:** ❌ No (observability only)

**Note:** Flow-control fields are ignored for this event

---

### 5. System Events

#### BeforeCompress

**Purpose:** Observability for context compression

**Trigger:**
- Before CLI compresses conversation context
- Fired asynchronously when context window nearly full

**Context Available:**
```json
{
  "event": "BeforeCompress",
  "contextSize": 128000,
  "targetSize": 64000,
  "compressionRatio": 0.5,
  "itemsToCompress": ["message-1", "message-2"]
}
```

**Flow Control:**
- **Ignored** - Cannot block or modify compression

**Common Use Cases:**
- Log compression events
- Track context usage patterns
- Monitor token consumption
- Telemetry for optimization

**Example Hook:**
```bash
#!/bin/bash
INPUT=$(cat)

# Log compression for telemetry
echo "$INPUT" | jq '{event: .event, contextSize: .contextSize, targetSize: .targetSize}' >> .gemini/compression.log

# Flow control ignored
echo '{"decision": "continue"}'
```

**Cancellable:** ❌ No
**Can Modify:** ❌ No (observability only)

**Note:** Flow-control fields are ignored. Hook cannot block compression.

---

## Hook Execution Flow

### Simple Request Flow

```
User Input
    ↓
SessionStart (if new session)
    ↓
BeforeAgent (can modify prompt)
    ↓
BeforeModel (can modify model input)
    ↓
AfterModel (can filter response)
    ↓
BeforeResponse/AfterChunk (streaming)
    ↓
AfterAgent (log turn)
    ↓
Display to User
```

### Tool Execution Flow

```
User Input
    ↓
BeforeAgent
    ↓
BeforeToolSelection (filter tools)
    ↓
LLM selects tool
    ↓
BeforeTool (validate/block)
    ↓
Tool Execution
    ↓
AfterTool (log result)
    ↓
Result to LLM
    ↓
AfterAgent
```

### Multiple Tools Flow

```
BeforeToolSelection
    ↓
LLM selects [ToolA, ToolB]
    ↓
BeforeTool (ToolA) → Execute → AfterTool (ToolA)
    ↓
BeforeTool (ToolB) → Execute → AfterTool (ToolB)
    ↓
Results aggregated
    ↓
Back to LLM
```

---

## Event Ordering Guarantees

**Guaranteed Order:**
1. SessionStart fires first (if applicable)
2. BeforeAgent fires before any LLM request
3. BeforeToolSelection fires before tool selection
4. BeforeTool fires before each tool execution
5. AfterTool fires after each tool completion
6. AfterAgent fires last in a turn
7. SessionEnd fires on session termination

**No Guarantee:**
- Exact timing of Notification events
- BeforeCompress timing (async)
- AfterAlert timing relative to other events

---

## Matcher Patterns

### Tool Event Matchers

**Regular Expression Matching:**
```json
{
  "BeforeTool": [
    {
      "matcher": "write_.*",  // Matches write_file, write_to_disk, etc.
      "hooks": [...]
    },
    {
      "matcher": "execute_.*|run_.*",  // Matches execute_ or run_ prefixes
      "hooks": [...]
    }
  ]
}
```

**Wildcard Matching:**
```json
{
  "BeforeTool": [
    {
      "matcher": "*",  // Matches ALL tools
      "hooks": [...]
    },
    {
      "matcher": "",  // Also matches ALL (empty string)
      "hooks": [...]
    }
  ]
}
```

### Lifecycle Event Matchers

**Exact String Matching:**
```json
{
  "SessionStart": [
    {
      "matcher": "startup",  // Exact match
      "hooks": [...]
    }
  ]
}
```

---

## Best Practices

### Performance

1. **Keep hooks fast** - Synchronous execution blocks CLI
2. **Use timeouts** - Configure reasonable timeouts (3-5 seconds)
3. **Avoid expensive operations** - No heavy I/O in critical path
4. **Filter aggressively** - Use matchers to limit hook invocations

### Security

1. **Validate all inputs** - Never trust tool parameters
2. **Use deny-by-default** - Explicitly allow safe operations
3. **Log security events** - Track all blocks and denials
4. **Sanitize outputs** - Filter PII and sensitive data

### Reliability

1. **Handle errors gracefully** - Catch exceptions, return valid JSON
2. **Provide useful error messages** - systemMessage should help users
3. **Test thoroughly** - Test allow, deny, and continue paths
4. **Version your hooks** - Track changes to hook logic

### Maintainability

1. **Document hook purpose** - Add comments explaining intent
2. **Use consistent naming** - Follow naming conventions
3. **Separate concerns** - One hook per responsibility
4. **Share via extensions** - Package team hooks as extensions

---

## Sources

**Official Documentation:**
- [Gemini CLI Hooks](https://geminicli.com/docs/hooks/)
- [Hooks Reference](https://geminicli.com/docs/hooks/reference/)
- [Writing Hooks](https://geminicli.com/docs/hooks/writing-hooks/)
- [Best Practices](https://geminicli.com/docs/hooks/best-practices/)

**Research:**
- [Research Notes - Official Docs](../../../research-notes/gemini/official-docs.md)
- [Research Notes - Automation Patterns](../../../research-notes/gemini/automation-patterns.md)

---

## See Also

- [Architecture & Internals](./architecture.md) - Execution model details
- [Configuration & Setup](./configuration.md) - How to configure hooks
- [Security & Safety](./security.md) - Security patterns
- [Examples](./examples.md) - Ready-to-use hook examples
- [API Reference](./api-reference.md) - Complete JSON schemas

---

**Document Version:** 1.0
**Event Coverage:** 13/13 events (100%)
**Last Verified:** 2026-02-21
