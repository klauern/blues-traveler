# OpenAI Codex - Architecture & Internals

**Last Updated:** 2026-02-21

---

## Overview

Codex's architecture is fundamentally different from local coding tools. It's a **cloud-first autonomous agent platform** with distributed execution, not a local assistant tool.

---

## Execution Architecture

### Cloud-First Model

```
User (CLI/Web/IDE)
    ↓ (JSON-RPC)
App Server (local or cloud)
    ↓ (HTTPS)
OpenAI API
    ↓
Cloud Sandbox (isolated container)
    ↓
Agent Execution (hours-long tasks)
    ↓
Results (streamed back)
```

### vs Local Tools (Claude Code, Cursor)

| Aspect | Codex | Local Tools |
|--------|-------|-------------|
| **Execution** | Cloud sandboxes | User's machine |
| **Parallelism** | Native (multiple containers) | Experimental (worktrees) |
| **Timeout** | None (background mode) | User must keep alive |
| **Resource Limits** | Cloud infrastructure | User's hardware |
| **Isolation** | Full container isolation | Process/directory isolation |

---

## App Server Protocol

### Architecture Overview

**Protocol:** JSON-RPC 2.0 over stdio
**Transport:** JSONL (one JSON object per line)
**Direction:** Bidirectional (client ↔ server)

```
┌─────────────┐                 ┌──────────────┐                 ┌─────────────┐
│   Client    │────stdio────────│  App Server  │────HTTPS────────│  OpenAI API │
│ (CLI/IDE)   │  (JSON-RPC)     │   (Go)       │                 │   (Cloud)   │
└─────────────┘                 └──────────────┘                 └─────────────┘
       ↑                                │
       │                                │
       └────────────────────────────────┘
         Server-initiated requests
         (approvals, notifications)
```

### Why JSON-RPC over stdio?

**Advantages:**
1. **Decoupled Releases**: Client and server can update independently
2. **Cross-Platform**: Works on all OSes without modification
3. **Simple Integration**: No network configuration required
4. **Bidirectional**: Server can request approvals from client
5. **Streamable**: Events stream in real-time

**Source:** [Unlocking the Codex Harness](https://openai.com/index/unlocking-the-codex-harness/)

### Integration Patterns

#### Pattern 1: Local Clients (CLI, Desktop)

```
Application Process
    ↓
Launch App Server (child process)
    ↓
Communicate via stdin/stdout
    ↓
App Server → OpenAI API → Cloud Sandbox
```

#### Pattern 2: Partner Integrations (Xcode, VS Code)

```
IDE Extension (stable binary)
    ↓ (points to)
App Server (separate binary, can update)
    ↓
OpenAI API
```

**Benefit:** IDE doesn't need updates when server improves

#### Pattern 3: Web Runtime (Codex Cloud)

```
Browser
    ↓ (HTTP + Server-Sent Events)
Worker/Container
    ↓
App Server (inside container)
    ↓
OpenAI API
```

---

## Background Mode Architecture

### Long-Running Tasks

**Problem:** Traditional APIs have timeouts (30s, 2min, 10min max)
**Solution:** Background mode with polling/webhooks

```python
# Start background task (no timeout!)
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,  # ← Key parameter
    messages=[{
        "role": "user",
        "content": "Refactor entire codebase to TypeScript"
    }]
)

# Task runs for hours if needed
# Get notified via webhook when complete
```

### Execution Model

```
1. Client creates background request
     ↓
2. OpenAI allocates cloud sandbox
     ↓
3. Agent executes (hours/days)
     ↓
4. Sandbox persists across reconnections
     ↓
5. Webhook fires on completion
     ↓
6. Client retrieves final result
```

### Streaming + Background

```python
# Get both: immediate progress + long-running execution
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,
    stream=True,  # ← Stream events in real-time
    messages=[{...}]
)

# Events stream immediately
for event in response.events:
    print(event.type, event.data)

# Connection can drop and reconnect
# Task continues in background
```

---

## Multi-Agent Orchestration

### Parallel Execution

**Codex Native:** Multiple agents run simultaneously in separate sandboxes

```
Orchestrator
    ├── Designer Agent (Sandbox 1)
    ├── Developer Agent (Sandbox 2)
    ├── Tester Agent (Sandbox 3)
    └── Reviewer Agent (Sandbox 4)

All run in parallel, results aggregated
```

### Agents SDK Integration

```python
from agents_sdk import Agent

# Each agent gets own Codex instance via MCP
designer = Agent(
    name="designer",
    tools=["codex"],  # MCP server connection
    instructions="Design features"
)

developer = Agent(
    name="developer",
    tools=["codex"],
    instructions="Implement features"
)

# Orchestrator coordinates
design = designer.run("Design auth system")
code = developer.run(f"Implement: {design.output}")
```

**Architecture:**

```
Agents SDK (Orchestrator)
    ↓ (calls via MCP)
Codex MCP Server
    ↓ (spawns)
Multiple Codex CLI instances
    ↓ (each connects to)
Separate Cloud Sandboxes
```

---

## Performance Characteristics

### Latency

| Operation | Latency | Notes |
|-----------|---------|-------|
| API Call | 100-500ms | Initial request |
| First Token | 1-3s | Start of response |
| Tool Execution | 100ms-30s | Depends on tool |
| File Write | 100-500ms | Local or cloud |
| Background Task | Hours | No timeout |

### Throughput

| Metric | Value | Notes |
|--------|-------|-------|
| Max Parallel Agents | Unlimited | Cloud sandboxes |
| Concurrent Tools | ~10 | Per agent |
| Files/Second | 100+ | Parallel writes |
| Tokens/Second | 50-100 | Streaming |

### Cost Optimization

**Prompt Caching:** 75% discount on cached tokens

```python
# First request: Full price
response = client.responses.create(
    model="gpt-5.2-codex",
    messages=[
        {"role": "system", "content": huge_codebase},  # Expensive
        {"role": "user", "content": "Add error handling"}
    ]
)

# Second request: 75% cheaper for cached codebase
response2 = client.responses.create(
    model="gpt-5.2-codex",
    messages=[
        {"role": "system", "content": huge_codebase},  # Cached!
        {"role": "user", "content": "Add logging"}
    ]
)
```

---

## Security Model

### Sandbox Isolation

**Each agent runs in isolated container:**
- Separate filesystem
- Network isolation (controlled egress)
- Resource limits (CPU, memory, disk)
- Ephemeral (destroyed after completion)

### Permission Model

**Approval-based execution:**

```
Agent wants to execute dangerous command
    ↓
Sends approval_request event
    ↓
User approves/denies via client
    ↓
Agent proceeds or cancels
```

### Secret Management

**Best Practice:** Use environment variables, not hardcoded secrets

```python
# Configure in OpenAI dashboard or local config
# Secrets injected into sandbox, not in code
response = client.responses.create(
    model="gpt-5.2-codex",
    environment={
        "DATABASE_URL": os.environ["DB_URL"],  # Passed securely
        "API_KEY": os.environ["API_KEY"]
    },
    messages=[{...}]
)
```

---

## Resource Limits

### Cloud Sandbox Limits

| Resource | Limit | Configurable |
|----------|-------|--------------|
| CPU | 4 cores | ❌ |
| Memory | 8 GB | ❌ |
| Disk | 50 GB | ❌ |
| Network | Metered | ❌ |
| Time | Unlimited | ✅ (background mode) |

### Hook Execution Limits

| Hook Type | Timeout | Configurable |
|-----------|---------|--------------|
| Tool Hooks | 30s | ✅ |
| File Hooks | 30s | ✅ |
| Event Hooks | 30s | ✅ |
| Notifications | 10s | ✅ |

---

## Lifecycle & State Management

### Agent Lifecycle

```
1. Create → Sandbox provisioned
2. Execute → Agent runs autonomously
3. Pause → Sandbox persists (background mode)
4. Resume → Agent continues from checkpoint
5. Complete → Results returned
6. Cleanup → Sandbox destroyed
```

### State Persistence

**Background Mode:**
- Sandbox state persists across disconnections
- File changes preserved
- Environment maintained
- Agent context retained

**Regular Mode:**
- Sandbox destroyed on completion
- No state persistence
- Fresh environment each time

---

## Comparison to Other Architectures

### vs Claude Code (Local-First)

| Aspect | Codex | Claude Code |
|--------|-------|-------------|
| Execution | Cloud sandboxes | User's machine |
| State | Persistent (background) | Session-based |
| Parallelism | Native | Worktrees (experimental) |
| Resource Limits | Cloud infrastructure | User's hardware |
| Network | Always connected | Can work offline |

### vs GitHub Copilot (Inline)

| Aspect | Codex | Copilot |
|--------|-------|---------|
| Architecture | Autonomous agent | Inline completion |
| Scope | Complete features | Code snippets |
| Execution | Cloud sandboxes | N/A (suggestions only) |
| State | Multi-turn conversations | Stateless suggestions |

---

## Sources

- [Unlocking the Codex Harness (Architecture)](https://openai.com/index/unlocking-the-codex-harness/)
- [Responses API Guide](https://platform.openai.com/docs/guides/)
- [Background Mode Documentation](https://platform.openai.com/docs/guides/background)
- Research: [API Capabilities](../../../research-notes/codex/api-capabilities.md)

---

**Last Updated:** 2026-02-21
