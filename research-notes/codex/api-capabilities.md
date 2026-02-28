# OpenAI Codex - API Capabilities Overview

## Executive Summary

Codex has evolved from a simple code generation API to a comprehensive autonomous software engineering platform with multiple API surfaces, event-driven architecture, and extensive automation capabilities.

## Core API Surfaces

### 1. Responses API (Primary API - 2026)

**Status**: Current recommended API for Codex
**Replaces**: Chat Completions API (deprecated Feb 2026)

#### Key Features
- **Semantic Event Streaming**: Type-safe event system with predefined schemas
- **Background Mode**: Long-running tasks without timeout concerns
- **Polling Support**: Check response status over time
- **Combined Streaming + Background**: Create background responses with immediate streaming

#### Supported Operations
```
- Create Response (synchronous)
- Create Background Response (async)
- Stream Events (real-time)
- Poll Response Status
- Retrieve Response
```

#### Example Use Cases
- Multi-hour feature development
- Large codebase analysis
- Batch code transformations
- Complex refactoring tasks

### 2. App Server API

**Purpose**: Bidirectional protocol powering all Codex client experiences
**Transport**: JSON-RPC over stdio (JSONL format)
**Status**: Open source, actively developed

#### Protocol Capabilities

##### Client → Server
- Start conversations (threads)
- Send user messages
- Approve/deny agent actions
- Configure agent settings
- Manage MCP servers

##### Server → Client
- Stream agent events (semantic types)
- Request approvals
- Send notifications
- Update conversation state
- Emit tool execution events

##### Event Types
```json
{
  "events": [
    "user_message",
    "agent_message",
    "command_run",
    "file_change",
    "tool_call",
    "approval_request",
    "notification",
    "thread_archived",
    "thread_unarchived"
  ]
}
```

#### Integration Patterns

**Pattern 1: Local Clients (CLI, Desktop, IDE)**
```
Application → Launch App Server (child process)
           → Maintain stdio bidirectional channel
           → Exchange JSON-RPC messages
           → Handle events in real-time
```

**Pattern 2: Partner Integrations (Xcode, etc.)**
```
Client (stable) → Points to App Server binary
                → Decoupled release cycles
                → Server-side improvements without client updates
```

**Pattern 3: Web Runtime (Codex Cloud)**
```
Browser → HTTP + Server-Sent Events → Worker
                                    → Container
                                    → App Server (inside container)
```

### 3. MCP Server Interface

**Purpose**: Expose Codex as Model Context Protocol server
**Use Case**: Multi-agent orchestration, tool integration

#### Exposed Tools
- `codex()` - Start a conversation
- `codex-reply()` - Continue a conversation

#### Features
- Keeps Codex alive across multiple agent turns
- Compatible with OpenAI Agents SDK
- Enables complex multi-agent workflows
- Supports scoped contexts per agent

#### Configuration
```bash
codex mcp add <server-name> -- <command>
# Example: codex mcp add context7 -- npx -y @upstash/context7-mcp
```

### 4. GitHub Action API

**Repository**: https://github.com/openai/codex-action
**Purpose**: CI/CD integration

#### Capabilities
- Install Codex CLI in GitHub Actions
- Run non-interactive `codex exec` commands
- Apply patches automatically
- Post code reviews
- Start Responses API proxy

#### Permissions Model
```yaml
permissions:
  contents: write     # Apply changes
  pull-requests: write  # Post reviews
  issues: write       # Comment on issues
```

## Automation & Callback Mechanisms

### 1. Webhooks (OpenAI API Level)

**URL**: Standard Webhooks specification
**Endpoint**: Your controlled HTTP endpoint
**Events**: Async completion notifications

#### Supported Webhook Events
```
- Batch completion
- Background response completion
- Fine-tuning job completion
```

#### Implementation
```
1. Configure webhook endpoint in OpenAI dashboard
2. Verify webhook signature (Standard Webhooks spec)
3. Handle event payload
4. Return 200 OK
```

#### Use Cases
- Turn polling into event-driven systems
- Trigger downstream workflows on completion
- Monitor long-running background tasks
- Orchestrate multi-step pipelines

### 2. Event Hooks (Codex-Specific)

**Configuration**: `~/.codex/config.toml` or project `codex.json`
**Trigger Points**: Lifecycle events during Codex execution

#### Hook Types

##### Tool Hooks
```toml
[hooks.tool.before]
# Runs before tool execution
# Has access to tool input

[hooks.tool.after]
# Runs after tool execution
# Has access to tool output
```

##### File Hooks
```toml
[hooks.file.before_write]
# Runs before file write operations
# Can validate or transform content

[hooks.file.after_write]
# Runs after file write operations
# Can trigger linters, formatters, etc.
```

##### Event Hooks
```toml
[hooks.event.prompt_gating]
# Filter/modify prompts before sending

[hooks.event.stop]
# Execute on agent stop

[hooks.event.notification]
# Handle notification events
```

#### Hook Configuration Example
```toml
[hooks.notify]
command = "/usr/local/bin/notify-webhook"
events = ["turn_complete", "approval_required", "error"]
```

### 3. Notification System

**Purpose**: External program integration for side-channel alerts
**Trigger**: Supported Codex events

#### Notification Targets
- Desktop toast notifications
- Chat webhooks (Slack, Discord, Teams)
- CI status updates
- Custom monitoring systems
- Logging services

#### Configuration
```toml
[notify]
command = "/path/to/notification-handler"
args = ["--webhook", "https://hooks.slack.com/..."]
events = ["turn_complete", "error", "approval_required"]
```

### 4. Automations (Scheduled & Triggered)

**Interface**: Codex App
**Execution**: Background, unprompted

#### Automation Components
```
- Instruction (prompt)
- Optional skills
- Triggers (schedule or event-based)
- Reporting (inbox or auto-archive)
```

#### Built-in Automation Triggers
- **Schedule-based**: Daily, hourly, custom cron
- **Event-based** (future): GitHub push, CI failure, PR open

#### Example Automations at OpenAI
- Daily issue triage
- CI failure summarization
- Daily release briefs
- Bug detection scans
- Documentation updates

#### Automation Workflow
```
1. Define automation (instruction + schedule)
2. Codex executes in background
3. Findings added to inbox OR
4. Auto-archived if nothing to report
```

## Streaming & Async Capabilities

### Real-time Streaming

#### Semantic Event Stream
```
Each event has:
- Type (predefined schema)
- Timestamp
- Payload (type-specific data)
- Context (thread, turn, item IDs)
```

#### Stream Consumption
```javascript
// Pseudo-code
const response = await createResponse({
  stream: true,
  model: "gpt-5.2-codex",
  // ...
});

for await (const event of response.events) {
  switch(event.type) {
    case 'agent_message':
      console.log(event.content);
      break;
    case 'file_change':
      applyChange(event.file, event.diff);
      break;
    case 'tool_call':
      logToolUse(event.tool, event.args);
      break;
  }
}
```

### Background Mode

#### Use Cases
- Multi-hour development tasks
- Large-scale refactoring
- Codebase-wide migrations
- Comprehensive test generation

#### Background + Streaming
```
1. Create background response with stream: true
2. Start receiving events immediately
3. Connection can drop and reconnect
4. Poll for status separately
5. Retrieve final result when complete
```

#### Polling Pattern
```javascript
const response = await createBackgroundResponse({...});
const responseId = response.id;

// Poll for completion
while (true) {
  const status = await getResponseStatus(responseId);
  if (status.state === 'completed') {
    const result = await getResponse(responseId);
    break;
  }
  await sleep(5000);
}
```

## Multi-Agent Capabilities

### Parallel Execution

**Feature**: Run multiple Codex agents simultaneously
**Isolation**: Git worktrees (local) or containers (cloud)

#### Use Cases
- Feature development + bug fixing in parallel
- Multiple code reviews simultaneously
- Concurrent testing and implementation
- Distributed task delegation

#### Configuration
```bash
# CLI supports experimental multi-agent mode
codex --multi-agent task1 &
codex --multi-agent task2 &
codex --multi-agent task3 &
```

### Multi-Agent Orchestration

**Tool**: OpenAI Agents SDK + Codex MCP Server
**Pattern**: Hierarchical agent coordination

#### Example Architecture
```
Orchestrator Agent
├── Designer Agent (via Codex MCP)
│   └── Creates game design spec
└── Developer Agent (via Codex MCP)
    └── Implements game from spec
```

#### Benefits
- Scoped context per agent
- Reviewable agent traces
- Deterministic workflows
- Scalable to complex pipelines

## Tool & Skill System

### Skills

**Definition**: Folders of instructions, scripts, and resources
**Purpose**: Repeatable task patterns
**Repository**: https://github.com/openai/skills

#### Skill Structure
```
my-skill/
├── INSTRUCTIONS.md    # Agent instructions
├── resources/         # Reference materials
└── scripts/          # Executable tools
```

#### Skill Discovery
- Codex automatically discovers skills in path
- Can reference skills by name
- Skills package domain knowledge

### Custom Tools

**Integration**: Via MCP servers
**Configuration**: `~/.codex/config.toml`

#### Tool Best Practices
- **Semantic naming**: `semantic_search` not `search`
- **Clear descriptions**: When, why, and how to use
- **Good/bad examples**: Guide proper usage
- **Type safety**: Well-defined input/output schemas

## Integration Capabilities

### Source Control
- Git integration (native)
- GitHub API (via MCP or direct)
- Automatic branch/worktree management
- Commit generation
- PR creation and review

### CI/CD Systems
- GitHub Actions (codex-action)
- Generic CI via CLI
- Webhook-triggered workflows
- Autofix on test failures

### Issue Tracking
- Jira integration (via skills/MCP)
- GitHub Issues
- Automated triage
- Cross-system synchronization

### Communication
- Slack webhooks
- Discord notifications
- Email via external scripts
- Custom chat integrations

### Development Tools
- IDE integration (VS Code, Xcode, etc.)
- Terminal (CLI)
- Web browser (Codex Cloud)
- MCP-compatible tools

## API Language Support

### Official Client Libraries

The App Server protocol supports client implementations in:
- **Go** - Main implementation
- **Python** - Official bindings
- **TypeScript** - Web/Node.js
- **Swift** - iOS/macOS integrations
- **Kotlin** - Android/JVM

### Schema Generation
- TypeScript type definitions
- JSON Schema for validation
- Auto-generated from protocol spec

## Rate Limits & Quotas

### Responses API
- Model-dependent rate limits
- Background mode for long tasks (no timeout)
- Prompt caching (75% discount on cached tokens)

### Pricing (as of Feb 2026)
- **Codex-mini-latest** (deprecated): $1.50 input / $6 output per 1M tokens
- **GPT-5.2-Codex**: Variable by reasoning tier
- **GPT-5.3-Codex**: Pricing TBA for API access

## Security & Privacy

### Local Execution
- Codex CLI runs locally by default
- Code never leaves your machine (local mode)
- Full control over data

### Cloud Execution
- Isolated sandbox containers
- Ephemeral environments
- Configurable data retention

### Authentication
- ChatGPT auth for app/web
- API key for programmatic access
- OAuth for partner integrations

## API Limitations & Constraints

### Current Limitations
- GPT-5.3-Codex API access "coming soon" (not yet available)
- Chat Completions API deprecated (use Responses API)
- Codex-mini-latest removed (upgrade to newer models)
- HTTP-based MCP servers not yet supported (stdio only)

### Best Practices
- Use Responses API for new integrations
- Prefer background mode for long tasks
- Implement webhooks for event-driven workflows
- Use App Server for rich client integrations
- Configure hooks for lifecycle automation

## Future Capabilities (Roadmap Hints)

Based on official communications and documentation:

- **Codex Jobs**: Fully cloud-based automation with triggers (e.g., "on push")
- **HTTP MCP support**: Direct HTTP endpoint support for MCP servers
- **GPT-5.3-Codex API**: API access for latest model
- **Enhanced automations**: More event-based triggers
- **Deeper IDE integrations**: Expanded partner ecosystem

## Summary: Hooks & Callbacks Matrix

| Mechanism | Scope | Trigger | Use Case | Configuration |
|-----------|-------|---------|----------|---------------|
| **Webhooks** | OpenAI API | Async completion | Event-driven pipelines | OpenAI Dashboard |
| **Event Hooks** | Codex | Lifecycle events | Custom validation/processing | config.toml |
| **Notifications** | Codex | Agent events | Alerts, monitoring | config.toml |
| **Automations** | Codex App | Schedule/events | Recurring tasks | App UI |
| **GitHub Actions** | CI/CD | Git events | Code review, autofix | .github/workflows |
| **App Server Events** | Custom clients | Agent actions | Rich UI integration | App Server protocol |

## Last Updated

February 21, 2026
