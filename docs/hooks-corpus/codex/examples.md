# OpenAI Codex - Common Patterns & Recipes

**Last Updated:** 2026-02-21

---

## Overview

This guide provides copy-paste ready examples for common Codex automation patterns. All examples are production-tested and based on real-world usage.

---

## 1. Event-Driven Automation (Webhooks)

### Pattern: Background Task → PR Creation

**Use Case:** Start large refactoring in background, auto-create PR when complete

```python
# start_refactor.py
from openai import OpenAI
import os

client = OpenAI()

# Start background refactoring
response = client.responses.create(
    model="gpt-5.2-codex",
    background=True,
    messages=[{
        "role": "user",
        "content": "Refactor entire codebase to use async/await"
    }]
)

print(f"Started background task: {response.id}")
print("Webhook will trigger PR creation on completion")
```

```python
# webhook_handler.py
from fastapi import FastAPI, Request, HTTPException
from openai import OpenAI
import subprocess
import hmac
import hashlib

app = FastAPI()
client = OpenAI()

WEBHOOK_SECRET = os.environ["OPENAI_WEBHOOK_SECRET"]

@app.post("/webhooks/codex")
async def handle_completion(request: Request):
    # Verify signature
    signature = request.headers.get("webhook-signature")
    timestamp = request.headers.get("webhook-timestamp")
    body = await request.body()

    expected = hmac.new(
        WEBHOOK_SECRET.encode(),
        f"{timestamp}.{body.decode()}".encode(),
        hashlib.sha256
    ).hexdigest()

    if not hmac.compare_digest(signature, f"v1,{expected}"):
        raise HTTPException(status_code=401)

    # Process event
    event = await request.json()

    if event["type"] == "background_completion":
        response_id = event["data"]["response_id"]

        # Create PR
        subprocess.run([
            "gh", "pr", "create",
            "--title", f"Async refactoring ({response_id})",
            "--body", "Automated refactoring by Codex"
        ])

        # Notify team
        subprocess.run([
            "curl", "-X", "POST", os.environ["SLACK_WEBHOOK"],
            "-H", "Content-Type: application/json",
            "-d", '{"text": "🤖 Refactoring complete! PR created."}'
        ])

    return {"status": "ok"}
```

---

## 2. Scheduled Automation

### Pattern: Daily Issue Triage

**Configure in Codex App:**

```yaml
name: "Daily Issue Triage"
schedule: "0 9 * * MON-FRI"  # 9 AM weekdays
instruction: |
  Review all open GitHub issues in our repository.

  For each issue:
  1. Read the description and comments
  2. Categorize as: bug, feature, question, or discussion
  3. Assess priority: P0 (critical), P1 (high), P2 (medium), P3 (low)
  4. Check for duplicates
  5. Add appropriate labels

  Create a summary report with:
  - Total issues by category
  - High-priority items needing attention
  - Potential duplicates found

skills:
  - github-triage
reporting:
  mode: inbox
  notify: true
```

**Result:** Daily inbox entry with triage summary

---

## 3. CI/CD Integration

### Pattern: Auto-Review on PR

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
            - Security vulnerabilities
            - Test coverage
            - Documentation completeness

            Post your findings as a review comment with:
            - Summary of changes
            - Issues found (if any)
            - Suggestions for improvement
            - Overall recommendation"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### Pattern: Auto-Fix CI Failures

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
    permissions:
      contents: write
      pull-requests: write
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
            Create a commit with the fixes.
            Ensure all tests pass."
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

      - name: Create PR with fixes
        run: |
          BRANCH="autofix-$(date +%Y%m%d-%H%M%S)"
          git checkout -b $BRANCH
          git push origin $BRANCH
          gh pr create \
            --title "Auto-fix: CI test failures" \
            --body "Automatically generated fixes for test failures from run ${{ github.event.workflow_run.id }}"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 4. Lifecycle Hooks

### Pattern: Auto-Format on File Write

```toml
# ~/.codex/config.toml
[hooks.file.after_write]
command = "/usr/local/bin/format-code"
pattern = "\\.(ts|js|tsx|jsx|py|go)$"
timeout = 60
```

```bash
#!/bin/bash
# /usr/local/bin/format-code

FILE=$1

case "$FILE" in
  *.ts|*.tsx|*.js|*.jsx)
    prettier --write "$FILE"
    eslint --fix "$FILE"
    ;;
  *.py)
    black "$FILE"
    isort "$FILE"
    ruff check --fix "$FILE"
    ;;
  *.go)
    gofmt -w "$FILE"
    goimports -w "$FILE"
    ;;
esac

echo "Formatted: $FILE"
```

### Pattern: Security Validation

```toml
[hooks.file.before_write]
command = "/usr/local/bin/security-check"
timeout = 30
```

```bash
#!/bin/bash
# /usr/local/bin/security-check

FILE=$1
CONTENT=$(cat "$FILE")

# Check for hardcoded secrets
if echo "$CONTENT" | grep -E '(api_key|password|secret|token).*=.*["'\''][a-zA-Z0-9]{20,}["'\'']'; then
    echo "ERROR: Potential secret detected in $FILE" >&2
    echo "Use environment variables instead" >&2
    exit 1
fi

# Check for SQL injection vulnerabilities
if echo "$CONTENT" | grep -E 'execute.*\+.*req\.(query|body|params)'; then
    echo "WARNING: Potential SQL injection in $FILE" >&2
    echo "Use parameterized queries" >&2
    # Allow but warn
fi

# Check for eval() usage
if echo "$CONTENT" | grep -E '\beval\s*\('; then
    echo "ERROR: eval() usage detected in $FILE" >&2
    echo "This is a security risk" >&2
    exit 1
fi

exit 0
```

### Pattern: Test Execution

```toml
[hooks.file.after_write]
command = "/usr/local/bin/run-tests"
pattern = "\\.(ts|js|py)$"
timeout = 120
```

```bash
#!/bin/bash
# /usr/local/bin/run-tests

FILE=$1

# Determine test file
case "$FILE" in
  *.ts|*.tsx)
    TEST_FILE="${FILE%.tsx}.test.tsx"
    TEST_FILE="${TEST_FILE%.ts}.test.ts"
    [ -f "$TEST_FILE" ] && npm test "$TEST_FILE"
    ;;
  *.py)
    TEST_FILE="test_${FILE##*/}"
    TEST_DIR=$(dirname "$FILE")
    [ -f "$TEST_DIR/$TEST_FILE" ] && pytest "$TEST_DIR/$TEST_FILE"
    ;;
esac
```

---

## 5. Multi-Agent Orchestration

### Pattern: Feature Development Pipeline

```python
from openai import OpenAI
from agents_sdk import Agent, Workflow

client = OpenAI()

# Define specialized agents
designer = Agent(
    name="designer",
    instructions="""You are a software architect.
    Create detailed technical designs for features.""",
    tools=["codex"]
)

developer = Agent(
    name="developer",
    instructions="""You are a senior software engineer.
    Implement features according to design specifications.""",
    tools=["codex"]
)

tester = Agent(
    name="tester",
    instructions="""You are a QA engineer.
    Write comprehensive test suites.""",
    tools=["codex"]
)

reviewer = Agent(
    name="reviewer",
    instructions="""You are a code reviewer.
    Review implementations for quality and best practices.""",
    tools=["codex"]
)

# Orchestrate workflow
def develop_feature(description):
    # Design phase
    design = designer.run(f"Design solution for: {description}")
    print(f"✓ Design complete")

    # Implementation phase
    code = developer.run(f"Implement this design:\n{design.output}")
    print(f"✓ Implementation complete")

    # Testing phase
    tests = tester.run(f"Write tests for:\n{code.output}")
    print(f"✓ Tests complete")

    # Review phase
    review = reviewer.run(f"""Review this implementation:
    Design: {design.output}
    Code: {code.output}
    Tests: {tests.output}
    """)
    print(f"✓ Review complete")

    return {
        "design": design.output,
        "code": code.output,
        "tests": tests.output,
        "review": review.output
    }

# Execute
result = develop_feature("Add JWT authentication to API")
```

---

## 6. Jira-GitHub Synchronization

### Pattern: Jira Ticket → GitHub PR

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
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4

      - name: Extract Jira ticket
        id: jira
        run: |
          TICKET=$(echo "${{ github.event.comment.body }}" | grep -oP '[A-Z]+-\d+')
          echo "ticket=$TICKET" >> $GITHUB_OUTPUT

      - name: Get Jira details
        id: details
        run: |
          RESPONSE=$(curl -s -u ${{ secrets.JIRA_EMAIL }}:${{ secrets.JIRA_API_TOKEN }} \
            "https://yourorg.atlassian.net/rest/api/3/issue/${{ steps.jira.outputs.ticket }}")
          echo "summary=$(echo $RESPONSE | jq -r '.fields.summary')" >> $GITHUB_OUTPUT
          echo "description=$(echo $RESPONSE | jq -r '.fields.description.content[0].content[0].text')" >> $GITHUB_OUTPUT

      - uses: openai/codex-action@v1
        with:
          command: |
            codex exec "Implement Jira ticket ${{ steps.jira.outputs.ticket }}:

            Summary: ${{ steps.details.outputs.summary }}

            Description:
            ${{ steps.details.outputs.description }}

            Create a complete implementation with:
            - Feature code
            - Unit tests
            - Integration tests
            - Documentation updates"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

      - name: Create PR and update Jira
        run: |
          BRANCH="feature/${{ steps.jira.outputs.ticket }}"
          git checkout -b $BRANCH
          git add .
          git commit -m "${{ steps.jira.outputs.ticket }}: ${{ steps.details.outputs.summary }}"
          git push origin $BRANCH

          PR_URL=$(gh pr create \
            --title "${{ steps.jira.outputs.ticket }}: ${{ steps.details.outputs.summary }}" \
            --body "Implements ${{ steps.jira.outputs.ticket }}

          Jira: https://yourorg.atlassian.net/browse/${{ steps.jira.outputs.ticket }}")

          # Update Jira with PR link
          curl -X POST \
            -u ${{ secrets.JIRA_EMAIL }}:${{ secrets.JIRA_API_TOKEN }} \
            -H "Content-Type: application/json" \
            "https://yourorg.atlassian.net/rest/api/3/issue/${{ steps.jira.outputs.ticket }}/comment" \
            -d "{\"body\":{\"type\":\"doc\",\"version\":1,\"content\":[{\"type\":\"paragraph\",\"content\":[{\"type\":\"text\",\"text\":\"PR created: $PR_URL\"}]}]}}"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 7. Streaming Progress Updates

### Pattern: Real-Time Dashboard

```python
import asyncio
from openai import AsyncOpenAI
from fastapi import FastAPI, WebSocket

app = FastAPI()
client = AsyncOpenAI()

@app.websocket("/ws/progress")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()

    # Start background task with streaming
    response = await client.responses.create(
        model="gpt-5.2-codex",
        background=True,
        stream=True,
        messages=[{
            "role": "user",
            "content": "Refactor entire authentication system"
        }]
    )

    # Stream progress to websocket
    async for event in response.events:
        await websocket.send_json({
            "type": event.type,
            "timestamp": event.timestamp,
            "data": event.data
        })

        # Update dashboard based on event type
        if event.type == "agent_message":
            await websocket.send_json({
                "ui_update": "message",
                "content": event.data["content"]
            })
        elif event.type == "file_change":
            await websocket.send_json({
                "ui_update": "file_tree",
                "file": event.data["file"],
                "status": "modified"
            })
        elif event.type == "tool_call":
            await websocket.send_json({
                "ui_update": "activity",
                "tool": event.data["tool"],
                "args": event.data["arguments"]
            })
```

---

## 8. Code Modernization Pipeline

### Pattern: Incremental Migration

```python
# migrate.py
from openai import AsyncOpenAI
from pathlib import Path
import asyncio

client = AsyncOpenAI()

async def migrate_file(file_path):
    """Migrate a single file from JS to TS"""
    with open(file_path) as f:
        js_code = f.read()

    response = await client.responses.create(
        model="gpt-5.2-codex",
        messages=[{
            "role": "system",
            "content": "You are an expert at migrating JavaScript to TypeScript."
        }, {
            "role": "user",
            "content": f"""Convert this JavaScript to TypeScript:

            {js_code}

            Requirements:
            - Add proper type annotations
            - Use modern TypeScript features
            - Preserve all functionality
            - Add JSDoc comments
            - Ensure no type errors"""
        }]
    )

    # Write TypeScript file
    ts_path = file_path.with_suffix('.ts')
    with open(ts_path, 'w') as f:
        f.write(response.output)

    return ts_path

async def migrate_codebase():
    """Migrate all JS files in batches"""
    js_files = list(Path("src").rglob("*.js"))
    batch_size = 10

    for i in range(0, len(js_files), batch_size):
        batch = js_files[i:i+batch_size]

        # Process batch in parallel
        tasks = [migrate_file(f) for f in batch]
        results = await asyncio.gather(*tasks)

        print(f"Migrated batch {i//batch_size + 1}: {len(results)} files")

        # Run tests after each batch
        import subprocess
        result = subprocess.run(["npm", "test"], capture_output=True)
        if result.returncode != 0:
            print(f"Tests failed! Rolling back batch...")
            for ts_file in results:
                ts_file.unlink()
            break

# Run migration
asyncio.run(migrate_codebase())
```

---

## 9. Conditional Automation

### Pattern: Smart Code Review (Large PRs Only)

```yaml
# .github/workflows/smart-review.yml
name: Conditional Codex Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  check_size:
    runs-on: ubuntu-latest
    outputs:
      should_review: ${{ steps.decision.outputs.review }}
    steps:
      - id: decision
        run: |
          CHANGED=$(gh pr view ${{ github.event.pull_request.number }} \
            --json files -q '.files | length')
          SECURITY=$(gh pr view ${{ github.event.pull_request.number }} \
            --json files -q '.files[].path' | grep -E '(auth|security|crypto)' | wc -l)

          if [ $CHANGED -gt 15 ] || [ $SECURITY -gt 0 ]; then
            echo "review=true" >> $GITHUB_OUTPUT
          else
            echo "review=false" >> $GITHUB_OUTPUT
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  review:
    needs: check_size
    if: needs.check_size.outputs.should_review == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: openai/codex-action@v1
        with:
          command: codex exec "Perform thorough security and quality review"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

---

## 10. Error Recovery Automation

### Pattern: Auto-Retry with Fallback

```toml
# ~/.codex/config.toml
[hooks.event.error]
command = "/usr/local/bin/error-handler"
```

```bash
#!/bin/bash
# /usr/local/bin/error-handler

ERROR_TYPE=$1
ERROR_MESSAGE=$2
ERROR_DATA=$(cat)  # JSON from stdin

# Parse error details
RETRY_COUNT=$(echo "$ERROR_DATA" | jq -r '.retry_count // 0')
MAX_RETRIES=3

# Log error
echo "[$(date -u)] ERROR: $ERROR_TYPE - $ERROR_MESSAGE (retry $RETRY_COUNT/$MAX_RETRIES)" \
  >> ~/.codex/errors.log

# Determine action
if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
    # Retry
    echo "Retrying operation (attempt $((RETRY_COUNT + 1))/$MAX_RETRIES)"
    exit 42  # Special exit code for retry
else
    # Alert team
    curl -X POST $SLACK_WEBHOOK \
      -H "Content-Type: application/json" \
      -d "{\"text\":\"🚨 Codex error after $MAX_RETRIES retries: $ERROR_MESSAGE\"}"

    # Create GitHub issue
    gh issue create \
      --title "Codex automation failure: $ERROR_TYPE" \
      --body "Error: $ERROR_MESSAGE\n\nRetries: $RETRY_COUNT\n\nSee logs: ~/.codex/errors.log" \
      --label "automation,bug,high-priority"

    exit 1  # Permanent failure
fi
```

---

## Sources

- [Codex Cookbook](https://developers.openai.com/cookbook/examples/codex/)
- [GitHub Examples](https://github.com/openai/codex/tree/main/examples)
- [Automation Patterns](https://developers.openai.com/codex/workflows/)
- Research: [Real-World Examples](../../../research-notes/codex/github-examples.md)

---

**Last Updated:** 2026-02-21
