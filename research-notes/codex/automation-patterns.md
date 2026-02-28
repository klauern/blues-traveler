# OpenAI Codex - Automation Patterns & Implementation Guide

## Overview

This document provides practical patterns for automating development workflows with OpenAI Codex at the API and integration level.

## Core Automation Paradigms

### 1. Event-Driven Automation (Webhooks)

**When to Use**: React to completion of long-running tasks
**Mechanism**: OpenAI API Webhooks

#### Pattern: Async Task Completion Handler

```python
# 1. Start background task
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,
    messages=[{"role": "user", "content": "Refactor entire codebase to use TypeScript"}]
)

# 2. Configure webhook to receive completion event
# (Set up in OpenAI dashboard or via API)

# 3. Webhook handler receives event
@app.post("/webhooks/codex")
async def handle_completion(request: Request):
    event = await request.json()

    if event['type'] == 'background_completion':
        response_id = event['data']['response_id']

        # Retrieve completed work
        result = client.responses.retrieve(response_id)

        # Trigger downstream workflow
        await create_pull_request(result)
        await notify_team(result)

    return {"status": "ok"}
```

#### Benefits
- No polling required
- Immediate reaction to completion
- Scalable for multiple concurrent tasks
- Enables complex pipeline orchestration

#### Use Cases
- Large-scale refactoring completion → auto-create PR
- Test generation completion → trigger CI run
- Documentation update → notify technical writers
- Security scan completion → file tickets

### 2. Scheduled Automation (Codex Automations)

**When to Use**: Recurring, predictable tasks
**Mechanism**: Codex App Automations

#### Pattern: Daily Issue Triage

```yaml
# Automation Configuration (Codex App UI or config)
name: "Daily Issue Triage"
schedule: "0 9 * * MON-FRI"  # 9 AM weekdays
instruction: |
  Review all open GitHub issues for our project.
  Categorize by type (bug, feature, question).
  Identify any issues that need immediate attention.
  For high-priority bugs, check if they're duplicates.
  Summarize findings in a daily report.
skills:
  - github-triage
  - duplicate-detector
reporting:
  mode: "inbox"  # Add findings to inbox
  auto_archive_if_empty: true
```

#### Implementation Steps
1. Define automation in Codex App
2. Configure schedule (cron syntax)
3. Add skills for specific capabilities
4. Set reporting preferences
5. Monitor inbox for results

#### OpenAI's Internal Examples
- **Daily issue triage**: Categorize and prioritize issues
- **CI failure summarization**: Aggregate and explain build failures
- **Release brief generation**: Compile changes for daily standup
- **Bug detection**: Scan recent commits for potential issues
- **Documentation drift**: Check for outdated docs

### 3. CI/CD Integration (GitHub Actions)

**When to Use**: Git event triggers (push, PR, release)
**Mechanism**: Codex GitHub Action

#### Pattern: Auto-Review on PR

```yaml
# .github/workflows/codex-review.yml
name: Codex Code Review

on:
  pull_request:
    types: [opened, synchronize, ready_for_review]

permissions:
  contents: read
  pull-requests: write

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for context

      - uses: openai/codex-action@v1
        with:
          command: |
            codex exec "Review this PR for:
            - Code quality and style
            - Potential bugs or edge cases
            - Test coverage
            - Documentation completeness
            Post your findings as a review comment."
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

#### Pattern: Auto-Fix CI Failures

```yaml
# .github/workflows/codex-autofix.yml
name: Auto-Fix Test Failures

on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]

jobs:
  autofix:
    if: ${{ github.event.workflow_run.conclusion == 'failure' }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Get failure logs
        run: |
          gh run view ${{ github.event.workflow_run.id }} --log > failure.log
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - uses: openai/codex-action@v1
        with:
          command: |
            codex exec "Analyze failure.log and fix the failing tests.
            Create a commit with the fixes."
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

      - name: Create PR with fixes
        run: |
          git checkout -b autofix/${{ github.run_id }}
          git push origin autofix/${{ github.run_id }}
          gh pr create --title "Auto-fix: CI failures" \
                       --body "Automatically generated fixes for test failures"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 4. Lifecycle Hooks (Event-Driven)

**When to Use**: React to Codex agent actions
**Mechanism**: Codex event hooks in config.toml

#### Pattern: File Validation Hook

```toml
# ~/.codex/config.toml
[hooks.file.before_write]
command = "/usr/local/bin/validate-file"
args = ["--strict"]

# Validation script
#!/bin/bash
# /usr/local/bin/validate-file
FILE_PATH=$1
FILE_CONTENT=$2

# Run linter
if ! eslint "$FILE_PATH"; then
    echo "Linting failed, blocking write"
    exit 1
fi

# Check for secrets
if grep -E '(api_key|password|secret).*=.*["\']' "$FILE_PATH"; then
    echo "Potential secret detected, blocking write"
    exit 1
fi

exit 0
```

#### Pattern: Notification Hook

```toml
[hooks.notify]
command = "/usr/local/bin/send-notification"
events = ["turn_complete", "approval_required", "error"]

# Notification script
#!/bin/bash
EVENT_TYPE=$1
EVENT_DATA=$2

case $EVENT_TYPE in
  "turn_complete")
    curl -X POST https://hooks.slack.com/... \
      -d "{\"text\":\"Codex completed task\"}"
    ;;
  "approval_required")
    curl -X POST https://hooks.slack.com/... \
      -d "{\"text\":\"Codex needs approval\", \"priority\":\"high\"}"
    ;;
  "error")
    curl -X POST https://hooks.slack.com/... \
      -d "{\"text\":\"Codex encountered error\", \"priority\":\"urgent\"}"
    ;;
esac
```

#### Pattern: Tool Execution Logging

```toml
[hooks.tool.after]
command = "/usr/local/bin/log-tool-usage"

# Logging script
#!/bin/bash
TOOL_NAME=$1
TOOL_INPUT=$2
TOOL_OUTPUT=$3
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "{\"timestamp\":\"$TIMESTAMP\",\"tool\":\"$TOOL_NAME\",\"input\":\"$TOOL_INPUT\",\"output\":\"$TOOL_OUTPUT\"}" \
  >> ~/.codex/tool-usage.jsonl

# Optional: Send to analytics
curl -X POST https://analytics.example.com/events \
  -H "Content-Type: application/json" \
  -d "{\"event\":\"tool_used\",\"tool\":\"$TOOL_NAME\",\"timestamp\":\"$TIMESTAMP\"}"
```

### 5. Multi-Agent Orchestration

**When to Use**: Complex workflows requiring specialized agents
**Mechanism**: Agents SDK + Codex MCP Server

#### Pattern: Feature Development Pipeline

```python
from openai import OpenAI
from agents_sdk import Agent, Workflow

# Define specialized agents
designer = Agent(
    name="designer",
    instructions="You design software features based on requirements.",
    tools=[codex_mcp_tool]
)

developer = Agent(
    name="developer",
    instructions="You implement features from design specs.",
    tools=[codex_mcp_tool]
)

tester = Agent(
    name="tester",
    instructions="You write comprehensive tests for implemented features.",
    tools=[codex_mcp_tool]
)

reviewer = Agent(
    name="reviewer",
    instructions="You review code for quality, security, and best practices.",
    tools=[codex_mcp_tool]
)

# Define workflow
workflow = Workflow([
    ("designer", "Create design spec for user authentication"),
    ("developer", "Implement the authentication system per the design"),
    ("tester", "Write unit and integration tests"),
    ("reviewer", "Review all changes and provide feedback")
])

# Execute workflow
result = workflow.execute()
```

#### Pattern: Parallel Task Delegation

```bash
#!/bin/bash
# Delegate multiple independent tasks to Codex agents

# Start agents in parallel
codex exec "Add logging to all API endpoints" --output feature1.patch &
PID1=$!

codex exec "Update documentation for new features" --output docs.patch &
PID2=$!

codex exec "Refactor database queries for performance" --output perf.patch &
PID3=$!

codex exec "Add TypeScript types to JavaScript modules" --output types.patch &
PID4=$!

# Wait for all to complete
wait $PID1 $PID2 $PID3 $PID4

# Apply patches sequentially
git apply feature1.patch
git apply docs.patch
git apply perf.patch
git apply types.patch

# Create single PR with all changes
git checkout -b multi-agent-improvements
git add .
git commit -m "Multi-agent improvements: logging, docs, perf, types"
git push origin multi-agent-improvements
gh pr create --title "Multi-agent improvements" --body "Automated improvements from parallel Codex agents"
```

### 6. Jira-GitHub Synchronization

**When to Use**: Keep issue tracking and code in sync
**Mechanism**: Codex automation + GitHub Action

#### Pattern: Jira Issue → GitHub PR Pipeline

```yaml
# .github/workflows/jira-to-pr.yml
name: Jira Issue to PR

on:
  issue_comment:
    types: [created]

jobs:
  create_pr:
    if: contains(github.event.comment.body, '/implement')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Extract Jira ticket
        id: jira
        run: |
          TICKET=$(echo "${{ github.event.comment.body }}" | grep -oP 'PROJ-\d+')
          echo "ticket=$TICKET" >> $GITHUB_OUTPUT

      - name: Get Jira details
        id: jira_details
        run: |
          # Fetch Jira ticket details via API
          DETAILS=$(curl -u ${{ secrets.JIRA_EMAIL }}:${{ secrets.JIRA_API_TOKEN }} \
            "https://yourorg.atlassian.net/rest/api/3/issue/${{ steps.jira.outputs.ticket }}")
          echo "summary=$(echo $DETAILS | jq -r '.fields.summary')" >> $GITHUB_OUTPUT
          echo "description=$(echo $DETAILS | jq -r '.fields.description')" >> $GITHUB_OUTPUT

      - uses: openai/codex-action@v1
        with:
          command: |
            codex exec "Implement the following Jira ticket:

            Ticket: ${{ steps.jira.outputs.ticket }}
            Summary: ${{ steps.jira_details.outputs.summary }}
            Description: ${{ steps.jira_details.outputs.description }}

            Create a complete implementation with tests and documentation."
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

      - name: Create PR and link to Jira
        run: |
          BRANCH="feature/${{ steps.jira.outputs.ticket }}"
          git checkout -b $BRANCH
          git add .
          git commit -m "${{ steps.jira.outputs.ticket }}: ${{ steps.jira_details.outputs.summary }}"
          git push origin $BRANCH

          PR_URL=$(gh pr create \
            --title "${{ steps.jira.outputs.ticket }}: ${{ steps.jira_details.outputs.summary }}" \
            --body "Implements ${{ steps.jira.outputs.ticket }}\n\nJira: https://yourorg.atlassian.net/browse/${{ steps.jira.outputs.ticket }}")

          # Update Jira with PR link
          curl -X POST -u ${{ secrets.JIRA_EMAIL }}:${{ secrets.JIRA_API_TOKEN }} \
            -H "Content-Type: application/json" \
            "https://yourorg.atlassian.net/rest/api/3/issue/${{ steps.jira.outputs.ticket }}/comment" \
            -d "{\"body\":\"PR created: $PR_URL\"}"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 7. Streaming Progress Updates

**When to Use**: Long-running tasks requiring progress visibility
**Mechanism**: Responses API streaming + webhooks/websockets

#### Pattern: Real-time Progress Dashboard

```python
import asyncio
from openai import AsyncOpenAI

client = AsyncOpenAI()

async def run_with_progress(task_description):
    """Execute Codex task with real-time progress updates"""

    response = await client.responses.create(
        model="gpt-5.2-codex",
        stream=True,
        background=True,
        messages=[{"role": "user", "content": task_description}]
    )

    async for event in response.events:
        # Send progress to dashboard via websocket
        await websocket.send(json.dumps({
            "type": event.type,
            "timestamp": event.timestamp,
            "data": event.data
        }))

        # Handle specific event types
        if event.type == "agent_message":
            print(f"Agent: {event.content}")

        elif event.type == "file_change":
            print(f"Modified: {event.file}")
            await update_file_tree(event.file)

        elif event.type == "tool_call":
            print(f"Using tool: {event.tool}")
            await show_tool_usage(event.tool, event.args)

        elif event.type == "approval_request":
            print(f"Approval needed: {event.message}")
            approval = await request_user_approval(event)
            await send_approval_response(approval)

# Usage
await run_with_progress("Refactor entire authentication system")
```

### 8. Code Modernization Pipeline

**When to Use**: Large-scale codebase transformations
**Mechanism**: Background mode + streaming + webhooks

#### Pattern: Incremental Migration

```python
# Migrate from JavaScript to TypeScript incrementally
import os
from pathlib import Path

async def migrate_codebase():
    """Migrate JS to TS in batches with progress tracking"""

    js_files = list(Path("src").rglob("*.js"))
    batch_size = 10

    for i in range(0, len(js_files), batch_size):
        batch = js_files[i:i+batch_size]
        file_list = "\n".join(str(f) for f in batch)

        # Start background migration for this batch
        response = await client.responses.create(
            model="gpt-5.2-codex",
            background=True,
            stream=True,
            messages=[{
                "role": "user",
                "content": f"""Convert these JavaScript files to TypeScript:

                {file_list}

                Preserve all functionality, add proper type annotations,
                update imports, and ensure no type errors.
                """
            }]
        )

        # Track progress
        async for event in response.events:
            if event.type == "file_change":
                # Rename .js to .ts
                old_path = Path(event.file)
                new_path = old_path.with_suffix('.ts')

                # Update progress database
                await db.update_migration_status(
                    file=str(old_path),
                    status="completed",
                    new_file=str(new_path)
                )

        print(f"Completed batch {i//batch_size + 1}/{len(js_files)//batch_size + 1}")

await migrate_codebase()
```

### 9. Conditional Automation (Smart Triggers)

**When to Use**: Automate only when specific conditions met
**Mechanism**: App Server events + custom logic

#### Pattern: Smart Code Review

```yaml
# Only run Codex review if PR meets criteria
name: Conditional Codex Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  check_conditions:
    runs-on: ubuntu-latest
    outputs:
      should_review: ${{ steps.conditions.outputs.should_review }}
    steps:
      - id: conditions
        run: |
          # Review if: large PR, security-sensitive files, or contains "TODO: review"
          CHANGED_FILES=$(gh pr view ${{ github.event.pull_request.number }} --json files -q '.files | length')
          SECURITY_FILES=$(gh pr view ${{ github.event.pull_request.number }} --json files -q '.files[].path' | grep -E '(auth|crypto|security)' | wc -l)
          TODO_REVIEW=$(gh pr view ${{ github.event.pull_request.number }} --json body -q '.body' | grep -c "TODO: review" || true)

          if [ $CHANGED_FILES -gt 15 ] || [ $SECURITY_FILES -gt 0 ] || [ $TODO_REVIEW -gt 0 ]; then
            echo "should_review=true" >> $GITHUB_OUTPUT
          else
            echo "should_review=false" >> $GITHUB_OUTPUT
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  review:
    needs: check_conditions
    if: needs.check_conditions.outputs.should_review == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: openai/codex-action@v1
        with:
          command: codex exec "Perform thorough security and quality review"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### 10. Error Recovery Automation

**When to Use**: Automatically handle and recover from failures
**Mechanism**: Hooks + retry logic

#### Pattern: Auto-Retry with Fallback

```toml
# ~/.codex/config.toml
[hooks.event.error]
command = "/usr/local/bin/error-handler"
max_retries = 3
```

```bash
#!/bin/bash
# /usr/local/bin/error-handler

ERROR_TYPE=$1
ERROR_MESSAGE=$2
RETRY_COUNT=$3

# Log error
echo "[$(date)] Error: $ERROR_TYPE - $ERROR_MESSAGE (retry $RETRY_COUNT)" >> ~/.codex/errors.log

# Send alert if max retries exceeded
if [ $RETRY_COUNT -ge 3 ]; then
    curl -X POST https://hooks.slack.com/... \
      -d "{\"text\":\"🚨 Codex error after 3 retries: $ERROR_MESSAGE\"}"

    # Create GitHub issue for investigation
    gh issue create \
      --title "Codex automation failure: $ERROR_TYPE" \
      --body "Error: $ERROR_MESSAGE\n\nSee logs: ~/.codex/errors.log" \
      --label "automation,bug"
fi

# Return appropriate exit code
if [ $RETRY_COUNT -lt 3 ]; then
    exit 42  # Signal retry
else
    exit 1   # Signal failure
fi
```

## Best Practices

### 1. Idempotency
- Design automations to be safely re-runnable
- Check state before making changes
- Use atomic operations where possible

### 2. Observability
- Log all automation executions
- Track success/failure rates
- Monitor execution duration
- Alert on anomalies

### 3. Graceful Degradation
- Implement fallback behaviors
- Handle partial failures
- Provide manual override options

### 4. Testing
- Test automations in isolated environments
- Use dry-run modes where available
- Validate outputs before applying

### 5. Security
- Never hardcode secrets in automation configs
- Use environment variables or secret managers
- Validate all external inputs
- Audit automation access regularly

### 6. Resource Management
- Set timeouts for long-running tasks
- Implement rate limiting
- Use background mode for expensive operations
- Clean up resources after completion

## Common Pitfalls

### ❌ Anti-Patterns

1. **Polling instead of webhooks** - Wastes resources and adds latency
2. **No error handling** - Automations fail silently
3. **Hardcoded credentials** - Security risk
4. **Missing validation** - Bad changes applied automatically
5. **No logging** - Can't debug when things go wrong
6. **Blocking operations** - Ties up resources unnecessarily

### ✅ Solutions

1. Use webhooks for async notifications
2. Implement comprehensive error handlers with hooks
3. Use environment variables and secret managers
4. Add validation hooks before file writes
5. Log all automation activity with timestamps
6. Use background mode for long operations

## Automation Maturity Model

### Level 1: Manual with Codex CLI
- Run `codex` commands manually
- Copy-paste results
- No integration

### Level 2: Scripted Automation
- Shell scripts call `codex exec`
- Basic GitHub Actions
- Manual triggers

### Level 3: Event-Driven
- Webhooks configured
- CI/CD integration
- Automatic PR creation

### Level 4: Intelligent Automation
- Conditional logic
- Multi-agent orchestration
- Self-healing systems

### Level 5: Autonomous
- Scheduled automations
- Predictive triggers
- Closed-loop feedback

## Last Updated

February 21, 2026
