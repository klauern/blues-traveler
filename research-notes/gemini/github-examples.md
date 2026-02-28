# Gemini - GitHub Examples and Real-World Usage

**Research Date:** February 21, 2026
**Focus:** Real-world examples, configurations, and community implementations

---

## Official Gemini CLI Repository

**Repository:** [google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli)
**Description:** Official open-source AI agent that brings Gemini directly into your terminal

### Key Resources

**Documentation Files:**
- `docs/hooks/index.md` - Main hooks documentation
- `docs/get-started/configuration.md` - Configuration guide
- `docs/hooks/writing-hooks.md` - Tutorial for creating hooks (implied)
- `docs/hooks/best-practices.md` - Best practices guide (implied)
- `docs/hooks/reference.md` - Technical reference (implied)

### Important Issues and Discussions

#### Feature Requests and Development

**1. Feature Request: Implement a Hooks System**
- **Issue:** [#2779](https://github.com/google-gemini/gemini-cli/issues/2779)
- **Title:** "Feature Request: Implement a Hooks System for Custom Automation and Workflow Integration"
- **Status:** Implemented in v0.26.0
- **Significance:** Original feature request that led to hooks system

**2. Comprehensive Hooking System**
- **Issue:** [#9070](https://github.com/google-gemini/gemini-cli/issues/9070)
- **Title:** "Feature: Comprehensive Hooking System"
- **Significance:** Detailed discussion of hook architecture and requirements

**3. Comprehensive System of Hooks**
- **Issue:** [#11703](https://github.com/google-gemini/gemini-cli/issues/11703)
- **Title:** "Feature: Comprehensive System of Hooks"
- **Significance:** Further refinement of hooks system design

**4. Hook Support in Extensions**
- **Issue:** [#14449](https://github.com/google-gemini/gemini-cli/issues/14449)
- **Title:** "Hook Support in Extensions"
- **Status:** Implemented
- **Significance:** Allows extensions to bundle hooks

**5. Skills Create Command**
- **Issue:** [#16031](https://github.com/google-gemini/gemini-cli/issues/16031)
- **Title:** "feat: Implement /skills create for AI-assisted skill scaffolding"
- **Significance:** AI-assisted creation of new skills

**6. Workflow-Based Development**
- **Issue:** [#13333](https://github.com/google-gemini/gemini-cli/issues/13333)
- **Title:** "Add workflow-based development functionality to Gemini CLI, enabling users to define, execute, and manage multi-step development workflows as reusable templates"
- **Significance:** Extending automation beyond hooks to full workflows

#### Weekly Updates and Releases

**Gemini CLI v0.26.0 Release**
- **Discussion:** [#17812](https://github.com/google-gemini/gemini-cli/discussions/17812)
- **Title:** "Gemini CLI Weekly Update [v0.26.0]: Skills, Hooks and the ability to take a step back with /rewind"
- **Significance:** Major release introducing hooks system
- **Features:**
  - Skills system
  - Hooks system (enabled by default)
  - `/rewind` command for stepping back

**Official Release**
- **Release:** [v0.26.0](https://github.com/google-gemini/gemini-cli/releases/tag/v0.26.0)
- **Key Changes:**
  - Hooks enabled by default
  - New `/hooks` management commands
  - Skills improvements
  - Performance enhancements

### Pull Requests

**1. Hooks Commands Panel**
- **PR:** [#14225](https://github.com/google-gemini/gemini-cli/pull/14225)
- **Title:** "feat(hooks): Hooks Commands Panel, Enable/Disable, and Migrate"
- **Author:** Edilmo
- **Significance:** Adds UI for managing hooks

**2. Hook System Documentation**
- **PR:** [#14307](https://github.com/google-gemini/gemini-cli/pull/14307)
- **Title:** "feat(hooks): Hook System Documentation"
- **Author:** Edilmo
- **Significance:** Comprehensive documentation for hooks

**3. Support Extension Hooks with Security Warning**
- **PR:** [#14460](https://github.com/google-gemini/gemini-cli/pull/14460)
- **Title:** "feat: Support Extension Hooks with Security Warning"
- **Author:** abhipatel12
- **Significance:** Security considerations for extension-provided hooks

**4. Align Hooks Enable/Disable with Skills**
- **PR:** [#16822](https://github.com/google-gemini/gemini-cli/pull/16822)
- **Title:** "feat(cli): align hooks enable/disable with skills and improve completion"
- **Author:** sehoon38
- **Significance:** Consistent UX across hooks and skills

---

## Community Examples and Integrations

### MCP Server Implementations

**1. gemini-mcp by RLabs-Inc**
- **Repository:** [RLabs-Inc/gemini-mcp](https://github.com/RLabs-Inc/gemini-mcp)
- **Description:** MCP Server that enables Claude Code to interact with Gemini
- **Significance:** Cross-AI-assistant integration
- **Use Case:** Use Gemini capabilities from Claude Code

**2. mcp-server-gemini**
- **Repository:** [aliargun/mcp-server-gemini](https://github.com/aliargun/mcp-server-gemini)
- **Description:** MCP server implementation for Google's Gemini API
- **Significance:** Standard MCP server for Gemini integration

### Live API Examples

**live-api-web-console**
- **Repository:** [google-gemini/live-api-web-console](https://github.com/google-gemini/live-api-web-console)
- **Description:** React-based starter app for using Live API over WebSockets with Gemini
- **Technologies:** React, WebSocket
- **Features:**
  - WebSocket connection management
  - Audio/video streaming
  - Function calling examples
  - UI for real-time conversations

### Streaming Examples

**gemini-stream**
- **Repository:** [TechWithTy/gemini-stream](https://github.com/TechWithTy/gemini-stream)
- **Description:** Gemini API integration for streaming SSE responses
- **Features:**
  - Route handlers
  - Client adapters
  - Health checks for Google GenAI Live API
  - Production-ready with CORS
  - Typed hooks
  - Provider switching

### SDK Implementations

**gemini_ex (Elixir)**
- **Repository:** [nshkrdotcom/gemini_ex](https://github.com/nshkrdotcom/gemini_ex)
- **Description:** Elixir Interface / Adapter for Google Gemini LLM
- **Platforms:** AI Studio and Vertex AI
- **Significance:** Community SDK for Elixir ecosystem

---

## Configuration Examples

### Basic Hooks Configuration

**File:** `.gemini/settings.json` or `~/.gemini/settings.json`

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "name": "security-check",
        "matcher": "write_.*",
        "command": "node",
        "args": ["./hooks/security-check.js"],
        "enabled": true
      }
    ],
    "AfterTool": [
      {
        "name": "log-tool-usage",
        "matcher": "*",
        "command": "python",
        "args": ["./hooks/log-usage.py"],
        "enabled": true
      }
    ],
    "BeforeAgent": [
      {
        "name": "inject-context",
        "matcher": "",
        "command": "./hooks/inject-context.sh",
        "enabled": true
      }
    ],
    "SessionStart": [
      {
        "name": "init-session",
        "command": "node",
        "args": ["./hooks/init.js"],
        "enabled": true
      }
    ]
  }
}
```

### Security Check Hook Example

**File:** `hooks/security-check.js`

```javascript
#!/usr/bin/env node

// Read input from stdin
const input = JSON.parse(await readStdin());

// Check for sensitive patterns
const sensitivePatterns = [
  /api[_-]?key/i,
  /password/i,
  /secret/i,
  /token/i,
  /aws[_-]?access/i,
  /private[_-]?key/i
];

const content = JSON.stringify(input.args);

for (const pattern of sensitivePatterns) {
  if (pattern.test(content)) {
    // Block the operation
    console.log(JSON.stringify({
      decision: "deny",
      systemMessage: `Security check failed: Potential sensitive data detected (${pattern})`
    }));
    process.exit(0);
  }
}

// Allow the operation
console.log(JSON.stringify({
  decision: "allow"
}));
process.exit(0);

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks).toString('utf8');
}
```

### Context Injection Hook Example

**File:** `hooks/inject-context.sh`

```bash
#!/bin/bash

# Read input from stdin
input=$(cat)

# Get recent git commits
git_log=$(git log --oneline -5 2>/dev/null || echo "")

# Get current git status
git_status=$(git status --short 2>/dev/null || echo "")

# Build context
context="Recent commits:\n$git_log\n\nCurrent status:\n$git_status"

# Output JSON with additional context
echo "$input" | jq --arg ctx "$context" '.context += $ctx'
```

### Logging Hook Example

**File:** `hooks/log-usage.py`

```python
#!/usr/bin/env python3

import json
import sys
from datetime import datetime

# Read input from stdin
input_data = json.load(sys.stdin)

# Log to file
with open('.gemini/tool-usage.log', 'a') as f:
    log_entry = {
        'timestamp': datetime.now().isoformat(),
        'tool': input_data.get('tool', 'unknown'),
        'args': input_data.get('args', {})
    }
    f.write(json.dumps(log_entry) + '\n')

# Continue without blocking
print(json.dumps({'decision': 'continue'}))
sys.exit(0)
```

### MCP Server Configuration

**File:** `.gemini/settings.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "mcp-server-github",
      "args": ["--token", "${GITHUB_TOKEN}"],
      "enabled": true
    },
    "postgres": {
      "command": "mcp-server-postgres",
      "args": ["--connection", "${DATABASE_URL}"],
      "enabled": true
    },
    "custom-api": {
      "command": "node",
      "args": ["./servers/custom-api-server.js"],
      "enabled": true
    }
  }
}
```

### Skills Example Structure

**Directory:** `.gemini/skills/test-generator/`

```
test-generator/
├── prompt.md           # Main skill prompt
├── scripts/
│   └── run-tests.sh    # Optional automation scripts
├── references/
│   └── test-patterns.md # Documentation and examples
└── assets/
    └── templates/      # Code templates
```

**File:** `.gemini/skills/test-generator/prompt.md`

```markdown
You are a test generation specialist. Generate comprehensive unit tests for the provided code.

Follow these guidelines:
1. Use the testing framework specified in the project
2. Cover edge cases and error scenarios
3. Include setup and teardown where appropriate
4. Follow the test patterns in references/test-patterns.md
5. Aim for 80%+ code coverage

Generate tests for the code that will be provided.
```

---

## Real-World Use Cases from Community

### 1. Security Automation

**Pattern:** BeforeTool hook to prevent credential leaks

**Implementation:**
- Hook: BeforeTool
- Matcher: `write_.*` (all write operations)
- Action: Scan content for API keys, passwords, secrets
- Decision: Deny if found, allow otherwise

**Benefits:**
- Prevents accidental credential commits
- Automatic enforcement
- No manual review needed
- Works across entire team

### 2. Context Enhancement

**Pattern:** BeforeAgent hook to inject project-specific context

**Implementation:**
- Hook: BeforeAgent
- Trigger: Before every model request
- Action: Add git status, recent commits, open issues
- Result: Model has full context without manual prompting

**Benefits:**
- More relevant suggestions
- Fewer clarifying questions needed
- Automatic context awareness
- Time savings

### 3. Workflow Automation

**Pattern:** Skills for common multi-step tasks

**Implementation:**
- Skill: PR review workflow
- Steps:
  1. Analyze changed files
  2. Run tests
  3. Check for security issues
  4. Generate review comments
  5. Create PR description

**Benefits:**
- Consistent review quality
- Faster reviews
- Catches common issues
- Reduces manual effort

### 4. Compliance and Auditing

**Pattern:** AfterTool hook for logging all operations

**Implementation:**
- Hook: AfterTool
- Matcher: `*` (all tools)
- Action: Log tool name, args, results, timestamp
- Storage: Append to audit log file

**Benefits:**
- Complete audit trail
- Compliance with regulations
- Debug assistance
- Usage analytics

### 5. Cost Optimization

**Pattern:** BeforeToolSelection hook to limit expensive tools

**Implementation:**
- Hook: BeforeToolSelection
- Action: Remove expensive tools from available set based on context
- Decision: Force use of cheaper alternatives when appropriate

**Benefits:**
- Reduce API costs
- Prevent accidental expensive operations
- Budget control
- Usage optimization

### 6. Quality Gates

**Pattern:** AfterAgent hook to validate outputs

**Implementation:**
- Hook: AfterAgent
- Action: Check generated code against linting rules
- Decision: Request regeneration if quality checks fail

**Benefits:**
- Maintain code quality
- Automatic enforcement of standards
- Reduce manual review
- Consistent output quality

---

## Integration Patterns

### CI/CD Integration

**GitHub Actions Example:**

```yaml
name: Gemini Code Review

on: [pull_request]

jobs:
  gemini-review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install Gemini CLI
        run: npm install -g @google/gemini-cli@latest

      - name: Run Gemini Review
        run: |
          gemini "Review this PR and provide feedback" \
            --context "@diff" \
            --output review.md
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}

      - name: Post Review
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const review = fs.readFileSync('review.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: review
            });
```

### Docker Integration

**Dockerfile for Gemini CLI with Hooks:**

```dockerfile
FROM node:18-alpine

# Install Gemini CLI
RUN npm install -g @google/gemini-cli@latest

# Install hook dependencies
RUN apk add --no-cache python3 bash git

# Copy hooks and configuration
COPY .gemini /app/.gemini
WORKDIR /app

# Set environment variables
ENV GEMINI_API_KEY=${GEMINI_API_KEY}

CMD ["gemini"]
```

### Pre-commit Hook Integration

**File:** `.git/hooks/pre-commit`

```bash
#!/bin/bash

# Use Gemini CLI to review staged changes
gemini "Review these changes for potential issues" \
  --context "@staged" \
  --mode security-check

if [ $? -ne 0 ]; then
  echo "Gemini security check failed. Commit aborted."
  exit 1
fi
```

---

## Community Resources

### Tutorial Series

**Romin Irani's Gemini CLI Tutorial Series (Medium):**
1. Getting Started
2. Basic Usage
3. Configuration settings via settings.json and .env files
4. Skills
5. Hooks
6. MCP Servers
7. Extensions
8. Advanced Workflows

**Link Pattern:** `https://medium.com/google-cloud/gemini-cli-tutorial-series-part-{N}-...`

### Google Codelabs

**Available Codelabs:**
1. **Gemini CLI Hands-on**
   - URL: https://codelabs.developers.google.com/gemini-cli-hands-on
   - Focus: Getting started, basic commands

2. **Gemini CLI Deep-Dive**
   - URL: https://codelabs.developers.google.com/gemini-cli-deep-dive
   - Focus: Advanced features, hooks, skills

3. **Build an MCP Server with Gemini CLI and Go**
   - URL: https://codelabs.developers.google.com/cloud-gemini-cli-mcp-go
   - Focus: Creating custom MCP servers

4. **Code Customization with Gemini Code Assist Enterprise**
   - URL: https://codelabs.developers.google.com/codelabs/code-assist-enterprise
   - Focus: Enterprise code customization setup

---

## Notable Third-Party Tools and Integrations

### Integration Platforms

**1. Pipedream**
- **URL:** https://pipedream.com/apps/google-gemini/integrations/http
- **Purpose:** Serverless platform for connecting Gemini API with other apps
- **Features:** Event-driven workflows, no infrastructure management

**2. Make (formerly Integromat)**
- **Purpose:** Visual workflow automation with Gemini
- **Integration:** Google AI Studio (Gemini) + Webhooks
- **Use Cases:** Automated content generation, data processing

**3. n8n**
- **Purpose:** Open-source workflow automation
- **Integration:** Webhook + Google AI Studio (Gemini)
- **Benefits:** Self-hosted, extensible, visual workflow builder

**4. Zapier**
- **Integration:** Webhooks by Zapier + Google AI Studio
- **Use Cases:** Automated AI deployments, chatbot workflows

**5. Albato**
- **Integration:** Gemini AI + HTTP Request/Outgoing webhook
- **Purpose:** Easy integration without coding

### Pipecat Transport

**Package:** `@pipecat-ai/gemini-live-websocket-transport`
- **Platform:** npm
- **Purpose:** Transport layer for Gemini Live API in Pipecat framework
- **Use Cases:** Voice agents, real-time AI conversations

---

## Extension Examples

### Extension Structure

```
my-gemini-extension/
├── package.json
├── hooks/
│   └── hooks.json
├── skills/
│   └── my-skill/
│       └── prompt.md
├── mcp-servers/
│   └── my-server.js
└── README.md
```

**hooks.json Example:**

```json
{
  "BeforeTool": [
    {
      "name": "extension-security",
      "matcher": "write_.*",
      "command": "node",
      "args": ["./hooks/security.js"],
      "enabled": true
    }
  ]
}
```

---

## Resources

### Official Repositories
- [Gemini CLI](https://github.com/google-gemini/gemini-cli)
- [Gemini Cookbook](https://github.com/google-gemini/cookbook)
- [Live API Web Console](https://github.com/google-gemini/live-api-web-console)

### Community Repositories
- [gemini-mcp (RLabs-Inc)](https://github.com/RLabs-Inc/gemini-mcp)
- [mcp-server-gemini (aliargun)](https://github.com/aliargun/mcp-server-gemini)
- [gemini-stream (TechWithTy)](https://github.com/TechWithTy/gemini-stream)
- [gemini_ex (nshkrdotcom)](https://github.com/nshkrdotcom/gemini_ex)

### Documentation
- [Gemini CLI Docs](https://geminicli.com/docs/)
- [Extensions Browse](https://geminicli.com/extensions/)

---

## Summary

The Gemini ecosystem has strong GitHub presence with:

1. **Official Repository** - Active development with comprehensive issues/PRs tracking
2. **Community Implementations** - MCP servers, SDKs, integration tools
3. **Real-World Examples** - Security, automation, workflow patterns
4. **Integration Platforms** - Pipedream, Make, n8n, Zapier, Albato
5. **Educational Resources** - Codelabs, tutorials, blog posts
6. **Extension Ecosystem** - Growing library of pre-built extensions

The hooks system in v0.26.0 represents a major advancement in automation capabilities, with active community adoption and real-world usage patterns emerging.
