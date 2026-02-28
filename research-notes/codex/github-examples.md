# OpenAI Codex - GitHub Examples & Real-World Integrations

## Official OpenAI Repositories

### 1. openai/codex
**URL**: https://github.com/openai/codex
**Description**: Main Codex repository - open-source CLI and App Server
**Language**: Go (primary), with client bindings in multiple languages

#### Key Features
- Complete App Server implementation
- JSON-RPC protocol over stdio
- CLI tool for terminal usage
- Multi-platform support (macOS, Linux, Windows)
- Extensive documentation

#### Directory Structure
```
openai/codex/
├── cmd/codex/           # CLI entry point
├── pkg/
│   ├── appserver/      # App Server implementation
│   ├── protocol/       # JSON-RPC protocol
│   ├── agent/          # Agent logic
│   └── tools/          # Built-in tools
├── docs/               # Documentation
└── examples/           # Usage examples
```

#### Notable Issues & Discussions
- **Event Hooks (#2109)**: Feature request for lifecycle hooks
  - Status: Implemented (PR #9796)
  - Allows custom scripts on tool, file, and event lifecycle
- **Hook System (#2150)**: Community discussion on hook requirements
  - Led to comprehensive hooks implementation
  - Pattern matching for flexible triggers

### 2. openai/codex-action
**URL**: https://github.com/openai/codex-action
**Description**: GitHub Action for CI/CD integration
**Purpose**: Run Codex in GitHub Actions workflows

#### Usage Examples

##### Basic Code Review
```yaml
# From repository examples
name: Codex Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - uses: openai/codex-action@v1
        with:
          command: codex exec "@codex review"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

##### Auto-fix CI Failures
```yaml
# From cookbook examples
name: Codex Autofix
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
      - uses: openai/codex-action@v1
        with:
          command: |
            codex exec "Fix the failing tests and create a commit"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

#### Features
- Automatic Codex CLI installation
- Responses API proxy setup
- Git configuration for commits
- PR review posting
- Patch application

### 3. openai/skills
**URL**: https://github.com/openai/skills
**Description**: Community skills catalog for Codex
**Purpose**: Shareable task patterns and domain knowledge

#### Skill Structure
```
skills/
├── code-review/
│   ├── INSTRUCTIONS.md     # Agent instructions
│   ├── checklist.md        # Review checklist
│   └── examples/           # Good/bad examples
├── jira-integration/
│   ├── INSTRUCTIONS.md
│   ├── scripts/
│   │   ├── create-pr.sh
│   │   └── sync-status.sh
│   └── resources/
│       └── jira-api.md
├── security-scan/
│   ├── INSTRUCTIONS.md
│   ├── patterns/           # Security patterns
│   └── tools/
│       └── scan.py
└── docs-generation/
    ├── INSTRUCTIONS.md
    └── templates/
```

#### Example Skills
1. **Code Review** - Systematic PR review process
2. **Jira Integration** - Issue tracking synchronization
3. **Security Scan** - Security vulnerability detection
4. **Documentation** - Auto-generate docs from code
5. **Test Generation** - Comprehensive test suite creation
6. **Migration** - Framework/language migrations
7. **Refactoring** - Safe code refactoring patterns

#### Using Skills
```bash
# Install skill from catalog
codex skill add code-review https://github.com/openai/skills/tree/main/code-review

# Use skill in conversation
codex
> Use the code-review skill to review this PR

# Reference skill in automation
codex exec --skill code-review "Review all open PRs"
```

## Community Examples

### 4. tuannvm/codex-mcp-server
**URL**: https://github.com/tuannvm/codex-mcp-server
**Description**: MCP server wrapper for Codex CLI
**Purpose**: Enable Claude Code to leverage Codex capabilities

#### Architecture
```
Claude Code → MCP Protocol → Codex MCP Server → Codex CLI → OpenAI API
```

#### Implementation Highlights
```python
# Simplified example from repository
from mcp import MCPServer

class CodexMCPServer(MCPServer):
    def __init__(self):
        super().__init__("codex")

    async def call_tool(self, name: str, args: dict):
        if name == "codex":
            # Start new conversation
            result = subprocess.run(
                ["codex", "exec", args["prompt"]],
                capture_output=True
            )
            return result.stdout

        elif name == "codex-reply":
            # Continue conversation
            result = subprocess.run(
                ["codex", "reply", args["thread_id"], args["message"]],
                capture_output=True
            )
            return result.stdout
```

#### Use Case
- Cross-agent collaboration
- Best-of-both-worlds (Claude reasoning + Codex coding)
- Multi-agent orchestration

### 5. codingmoh/open-codex
**URL**: https://github.com/codingmoh/open-codex
**Description**: Open-source alternative inspired by OpenAI Codex
**Purpose**: Local LLM support for Codex-like workflows

#### Key Differences from OpenAI Codex
- Supports local models (Ollama, LM Studio)
- Fully open source
- Privacy-focused (no cloud required)
- Community-driven development

#### Relevance for Research
- Shows how Codex architecture can work with any LLM
- Demonstrates demand for local/private coding agents
- Alternative for air-gapped environments

## Cookbook Examples

### 6. Jira ↔ GitHub Automation
**Source**: https://developers.openai.com/cookbook/examples/codex/jira-github/
**Repository**: Included in official cookbook

#### Workflow
```
1. Label Jira issue with "auto-implement"
2. GitHub Action triggers
3. Codex reads Jira ticket details
4. Codex implements feature
5. Codex creates GitHub PR
6. PR link added to Jira ticket
7. PR status updates Jira automatically
```

#### Key Code Snippets

```bash
#!/bin/bash
# From cookbook example: jira-github/sync.sh

JIRA_TICKET=$1

# Fetch ticket details
TICKET_JSON=$(curl -s -u ${JIRA_EMAIL}:${JIRA_TOKEN} \
  "https://${JIRA_DOMAIN}/rest/api/3/issue/${JIRA_TICKET}")

SUMMARY=$(echo $TICKET_JSON | jq -r '.fields.summary')
DESCRIPTION=$(echo $TICKET_JSON | jq -r '.fields.description')

# Run Codex to implement
codex exec "Implement ${JIRA_TICKET}: ${SUMMARY}

Details:
${DESCRIPTION}

Create a complete implementation with tests."

# Create PR
BRANCH="feature/${JIRA_TICKET}"
git checkout -b $BRANCH
git add .
git commit -m "${JIRA_TICKET}: ${SUMMARY}"
git push origin $BRANCH

PR_URL=$(gh pr create --title "${JIRA_TICKET}: ${SUMMARY}" \
  --body "Implements ${JIRA_TICKET}")

# Update Jira
curl -X POST -u ${JIRA_EMAIL}:${JIRA_TOKEN} \
  "https://${JIRA_DOMAIN}/rest/api/3/issue/${JIRA_TICKET}/comment" \
  -H "Content-Type: application/json" \
  -d "{\"body\":\"PR created: ${PR_URL}\"}"
```

### 7. Code Modernization
**Source**: https://developers.openai.com/cookbook/examples/codex/code_modernization/

#### Pattern: JavaScript to TypeScript Migration

```python
# From cookbook example
import os
from pathlib import Path
from openai import OpenAI

client = OpenAI()

def modernize_file(file_path):
    """Convert JS file to TypeScript"""

    with open(file_path) as f:
        js_code = f.read()

    response = client.responses.create(
        model="gpt-5.2-codex",
        messages=[{
            "role": "system",
            "content": "You are an expert at modernizing JavaScript to TypeScript."
        }, {
            "role": "user",
            "content": f"""Convert this JavaScript to TypeScript:

            ```javascript
            {js_code}
            ```

            Requirements:
            - Add proper type annotations
            - Use modern TypeScript features
            - Preserve all functionality
            - Add JSDoc comments
            """
        }]
    )

    ts_code = response.choices[0].message.content

    # Write TypeScript file
    ts_path = file_path.with_suffix('.ts')
    with open(ts_path, 'w') as f:
        f.write(ts_code)

    return ts_path

# Batch process
for js_file in Path('src').rglob('*.js'):
    ts_file = modernize_file(js_file)
    print(f"Converted {js_file} → {ts_file}")
```

### 8. Multi-Agent Workflows with Agents SDK
**Source**: https://cookbook.openai.com/examples/codex/codex_mcp_agents_sdk/building_consistent_workflows_codex_cli_agents_sdk

#### Pattern: Design → Develop → Test Pipeline

```python
# From cookbook example
from openai import OpenAI
from agents_sdk import Agent

client = OpenAI()

# Define agents
designer = Agent(
    name="designer",
    instructions="""You are a software designer.
    Create detailed design specifications for features.""",
    tools=["codex"]  # Via MCP
)

developer = Agent(
    name="developer",
    instructions="""You are a software developer.
    Implement features according to design specs.""",
    tools=["codex"]
)

tester = Agent(
    name="tester",
    instructions="""You are a QA engineer.
    Write comprehensive tests for implementations.""",
    tools=["codex"]
)

# Orchestrate workflow
def feature_pipeline(feature_request):
    # Design phase
    design_spec = designer.run(
        f"Design a solution for: {feature_request}"
    )

    # Development phase
    implementation = developer.run(
        f"Implement this design:\n{design_spec.output}"
    )

    # Testing phase
    tests = tester.run(
        f"Write tests for this implementation:\n{implementation.output}"
    )

    return {
        "design": design_spec.output,
        "code": implementation.output,
        "tests": tests.output
    }

# Execute
result = feature_pipeline("Add user authentication with JWT")
```

### 9. Code Review Automation
**Source**: https://developers.openai.com/cookbook/examples/codex/build_code_review_with_codex_sdk/

#### Pattern: Comprehensive PR Review

```python
# From cookbook example
def review_pull_request(pr_number):
    """Automated code review using Codex"""

    # Get PR details
    pr = gh.get_pull_request(pr_number)

    # Get changed files
    files = pr.get_files()

    review_comments = []

    for file in files:
        if file.additions > 0:  # Only review files with additions
            # Get file diff
            diff = file.patch

            # Review with Codex
            response = client.chat.completions.create(
                model="gpt-5.2-codex",
                messages=[{
                    "role": "system",
                    "content": """You are a code reviewer. Review code for:
                    - Bugs and edge cases
                    - Security vulnerabilities
                    - Performance issues
                    - Code style and best practices
                    - Test coverage"""
                }, {
                    "role": "user",
                    "content": f"""Review this code change:

                    File: {file.filename}

                    ```diff
                    {diff}
                    ```

                    Provide specific, actionable feedback."""
                }]
            )

            feedback = response.choices[0].message.content

            review_comments.append({
                "path": file.filename,
                "body": feedback,
                "position": file.patch.count('\n')
            })

    # Post review
    pr.create_review(
        body="Automated code review by Codex",
        comments=review_comments,
        event="COMMENT"
    )

# Use in GitHub Action
review_pull_request(os.environ['PR_NUMBER'])
```

## Real-World Integration Patterns

### 10. OpenAI's Internal Usage
**Source**: https://openai.com/business/guides-and-resources/how-openai-uses-codex/

#### Daily Automations
```yaml
# OpenAI runs these automations daily
automations:
  - name: "Issue Triage"
    schedule: "0 9 * * *"  # 9 AM daily
    instruction: "Categorize open issues, flag urgent ones"

  - name: "CI Failure Summary"
    schedule: "0 10 * * *"  # 10 AM daily
    instruction: "Summarize all CI failures from last 24 hours"

  - name: "Release Brief"
    schedule: "0 8 * * *"  # 8 AM daily
    instruction: "Generate daily release brief for standup"

  - name: "Bug Detection"
    schedule: "0 0 * * *"  # Midnight daily
    instruction: "Scan recent commits for potential bugs"

  - name: "Documentation Drift"
    schedule: "0 0 * * SUN"  # Weekly on Sunday
    instruction: "Check for outdated documentation"
```

#### Impact
- Reduced manual triage time by 60%
- Faster CI failure resolution
- Improved daily standup efficiency
- Proactive bug detection

### 11. Partner Integrations

#### Xcode Integration (Apple)
**Pattern**: Decoupled client/server architecture

```
Xcode Extension (stable)
    ↓
App Server Binary (can update independently)
    ↓
OpenAI API
```

**Benefits**:
- Xcode extension doesn't need updates
- Server-side improvements deploy immediately
- Stable user experience
- Rapid iteration on capabilities

## Example Projects & Templates

### 12. Starter Templates

#### Basic CLI Integration
```bash
#!/bin/bash
# codex-wrapper.sh - Simple Codex integration template

TASK="$1"
OUTPUT_FILE="${2:-output.patch}"

# Run Codex non-interactively
codex exec "$TASK" > "$OUTPUT_FILE"

if [ $? -eq 0 ]; then
    echo "✓ Codex completed successfully"
    echo "Output: $OUTPUT_FILE"

    # Optional: Apply patch
    read -p "Apply changes? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git apply "$OUTPUT_FILE"
    fi
else
    echo "✗ Codex failed"
    exit 1
fi
```

#### GitHub Action Template
```yaml
# .github/workflows/codex-template.yml
# Reusable template for Codex automations

name: Codex Automation Template

on:
  workflow_call:
    inputs:
      task:
        required: true
        type: string
      create_pr:
        required: false
        type: boolean
        default: false

jobs:
  codex:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4

      - uses: openai/codex-action@v1
        with:
          command: codex exec "${{ inputs.task }}"
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

      - name: Create PR
        if: inputs.create_pr
        run: |
          BRANCH="codex/$(date +%Y%m%d-%H%M%S)"
          git checkout -b $BRANCH
          git add .
          git commit -m "Codex: ${{ inputs.task }}"
          git push origin $BRANCH
          gh pr create --title "Codex: ${{ inputs.task }}" \
                       --body "Automated by Codex"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### Using the Template
```yaml
# .github/workflows/daily-tasks.yml
name: Daily Codex Tasks

on:
  schedule:
    - cron: '0 9 * * *'

jobs:
  update-deps:
    uses: ./.github/workflows/codex-template.yml
    with:
      task: "Update all dependencies to latest versions"
      create_pr: true

  fix-todos:
    uses: ./.github/workflows/codex-template.yml
    with:
      task: "Find and fix any TODO comments marked as 'urgent'"
      create_pr: true
```

## Learning Resources

### Documentation Examples
- **Quickstart**: Step-by-step first project
- **Cookbook**: 20+ practical examples
- **API Reference**: Complete endpoint documentation
- **Integration Guides**: Platform-specific tutorials

### Community Contributions
- **Skills Catalog**: 50+ community-contributed skills
- **Discussion Forum**: Active Q&A and patterns
- **Example Repositories**: Real-world integrations

## Best Practices from Examples

### 1. Error Handling
```python
# From multiple cookbook examples
try:
    result = codex_exec(task)
except CodexError as e:
    # Log error
    logging.error(f"Codex failed: {e}")

    # Notify team
    notify_slack(f"Codex automation failed: {e}")

    # Create issue for investigation
    create_github_issue(
        title=f"Codex automation failure",
        body=f"Task: {task}\nError: {e}"
    )
```

### 2. Validation Before Application
```python
# From code review example
changes = codex_exec("Fix the bug")

# Validate before applying
if validate_changes(changes):
    apply_changes(changes)
else:
    request_manual_review(changes)
```

### 3. Incremental Processing
```python
# From modernization example
for batch in chunk_files(all_files, size=10):
    results = process_batch(batch)
    if all_successful(results):
        commit_batch(results)
    else:
        rollback_batch()
```

## Analytics & Metrics

### Tracking Automation Success

```python
# Example metrics collection
metrics = {
    "automation_runs": count_executions(),
    "success_rate": successful / total,
    "avg_duration": sum(durations) / len(durations),
    "errors": collect_errors(),
    "cost": sum(api_costs)
}

# Send to analytics platform
send_metrics(metrics)
```

## Last Updated

February 21, 2026
