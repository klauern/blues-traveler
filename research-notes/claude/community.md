# Claude Code Hooks - Community Resources & Tips

**Research Date:** 2026-02-21
**Sources:** Blog posts, tutorials, community discussions
**Status:** ✅ Complete

---

## Overview

This document catalogs community-created tutorials, blog posts, guides, and tips for Claude Code hooks discovered through web research and community resources.

---

## Blog Posts & Tutorials

### 1. A Developer's Guide to settings.json in Claude Code (2025)

**Source:** https://www.eesel.ai/blog/settings-json-claude-code
**Date:** 2025
**Publisher:** eesel.ai

**Key Points:**
- Claude Code hooks configured in JSON at three levels
- `~/.claude/settings.json` (all projects)
- `.claude/settings.json` (single project, shareable via git)
- `.claude/settings.local.json` (local project settings)

**Hook Types Explained:**
- **Command hooks:** Run shell scripts, best for deterministic tasks
- **Prompt hooks:** Use LLM for yes/no decisions when logic too complex
- **Agent hooks:** Spawn multi-turn subagent with file access and commands

**Management Tips:**
- Use `/hooks` interactive UI command instead of manual editing
- Minimizes syntax errors
- Safe configuration through GUI

**Disable Feature:**
- Set `"disableAllHooks": true` to temporarily disable
- Use toggle in `/hooks` menu
- Hook changes don't take effect mid-session (security)

---

### 2. Claude Code Hooks: Complete Guide with 20+ Ready-to-Use Examples (2026)

**Source:** https://aiorg.dev/blog/claude-code-hooks
**Date:** 2026
**Publisher:** aiorg.dev

**Coverage:**
- 20+ production-ready hook examples
- Complete event catalog
- Configuration patterns
- Real-world use cases

**Notable Examples:**
- Security validation hooks
- Code formatting automation
- Test execution triggers
- Notification integrations

---

### 3. Understanding Claude Code Hooks Documentation

**Source:** https://blog.promptlayer.com/understanding-claude-code-hooks-documentation/
**Publisher:** PromptLayer

**Focus:**
- Breaking down official documentation
- Practical interpretations
- Common pitfalls
- Best practices from production use

**Key Insights:**
- Hook lifecycle explained
- Event timing clarifications
- Decision control patterns

---

### 4. Configure Claude Code Hooks to Automate Your Workflow

**Source:** https://www.gend.co/blog/configure-claude-code-hooks-automation
**Publisher:** Gend.co

**Workflow Focus:**
- Automation patterns
- Workflow integration
- CI/CD considerations
- Team collaboration

**Practical Patterns:**
- Auto-formatting workflows
- Security gates
- Quality checks
- Deployment automation

---

### 5. Claude Code Power User Customization: How to Configure Hooks

**Source:** https://claude.com/blog/how-to-configure-hooks
**Publisher:** Official Claude Blog

**Official Guidance:**
- First-party configuration guide
- Best practices from Anthropic
- Security considerations
- Performance optimization

**Key Recommendations:**
- Hook design principles
- When to use hooks vs. skills
- Permission model understanding

---

### 6. How I Use Claude Code (+ My Best Tips)

**Source:** https://www.builder.io/blog/claude-code
**Publisher:** Builder.io
**Author:** Steve Sewell

**Personal Workflow:**
- Real-world usage patterns
- Tips from production use
- Hook configuration examples
- Integration with development tools

**Insider Tips:**
- Hook performance optimization
- Debugging techniques
- Common mistakes to avoid

---

### 7. Claude Code Hooks: A Practical Guide to Workflow Automation

**Source:** https://www.datacamp.com/tutorial/claude-code-hooks
**Publisher:** DataCamp

**Tutorial Format:**
- Step-by-step guide
- Hands-on examples
- Exercise-based learning
- Quiz and verification

**Learning Path:**
- Hook basics
- Configuration setup
- Writing first hook
- Advanced patterns
- Production deployment

---

### 8. Claude Code Hooks Complete Guide (February 2026 Edition)

**Source:** https://smartscope.blog/en/generative-ai/claude/claude-code-hooks-guide/
**Publisher:** SmartScope Blog
**Date:** February 2026

**Updated Content:**
- Latest hook features
- 2026 updates
- New event types
- Enhanced capabilities

---

### 9. How to Use Claude Code: Specs, Skills, Commands and Hooks

**Source:** https://levelup.gitconnected.com/how-to-use-claude-code-bed73d273638
**Publisher:** Level Up Coding
**Author:** Jarek Orzel
**Date:** January 2026

**Comprehensive Overview:**
- Specs (CLAUDE.md)
- Skills (custom commands)
- Hooks (automation)
- Command reference

**Integration Focus:**
- How hooks work with skills
- Combining with CLAUDE.md
- Complete workflow examples

---

### 10. Shipyard Claude Code CLI Cheatsheet

**Source:** https://shipyard.build/blog/claude-code-cheat-sheet/
**Publisher:** Shipyard

**Quick Reference:**
- Config patterns
- Command examples
- Prompts
- Best practices
- Copy-paste snippets

---

### 11. Claude Code Hooks Guide (GitButler Docs)

**Source:** https://docs.gitbutler.com/features/ai-integration/claude-code-hooks
**Publisher:** GitButler

**Git Integration:**
- Hooks for git workflows
- Branch management automation
- Commit validation
- PR automation

**GitButler-Specific:**
- Virtual branch integration
- GitButler + Claude Code patterns

---

### 12. ClaudeLog - Documentation Hub

**Source:** https://claudelog.com/
**Publisher:** Community project

**Resource Aggregation:**
- Docs compilation
- Guides collection
- Tutorials index
- Best practices library

**Configuration Section:**
- Hook examples
- Settings patterns
- Common configurations

---

## Community Insights & Tips

### From Blog Posts & Discussions

#### Communication: stdin, stdout, stderr, Exit Codes

**How Hooks Work:**
- Event fires → JSON to stdin
- Script reads data, does work
- Signals via exit code

**Exit Code Meanings:**
- `0` = Success, parse stdout for JSON or add to context
- `2` = Blocking error, stderr becomes error message
- Other = Non-blocking error, logged in verbose

**Source:** [eesel.ai](https://www.eesel.ai/blog/settings-json-claude-code)

---

#### PreToolUse Decision Options

Three options specific to PreToolUse:
1. **"allow"** - Proceed without permission prompt
2. **"deny"** - Cancel tool call, send reason to Claude
3. **"ask"** - Show permission prompt as normal

**JSON Format:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "Database writes not allowed"
  }
}
```

**Source:** [aiorg.dev](https://aiorg.dev/blog/claude-code-hooks)

---

#### PostToolUse Limitations

**Important:** PostToolUse hooks cannot undo actions since tool already executed.

**Use Cases:**
- Formatting (Prettier, Black)
- Linting
- Test execution (async)
- Logging
- Notifications

**Cannot Do:**
- Prevent file writes (too late)
- Rollback changes
- Block operations

**Source:** [DataCamp tutorial](https://www.datacamp.com/tutorial/claude-code-hooks)

---

#### SessionStart Context Injection

**Special Behavior:**
- stdout added to Claude's context
- Perfect for loading TODOs, git status, recent tickets

**Example Pattern:**
```bash
#!/bin/bash
echo "## Current Context"
git status --short
echo ""
echo "### Recent TODOs"
rg "TODO" --max-count 10
```

**Source:** [Official Claude Blog](https://claude.com/blog/how-to-configure-hooks)

---

#### JSON Validation Issues

**Common Problem:**
- Shell profile prints to stdout
- Interferes with JSON parsing
- Hooks fail silently

**Solution:**
- Redirect debug output to stderr: `>&2`
- Keep stdout clean for JSON only
- Test JSON parsing separately

**Source:** [PromptLayer blog](https://blog.promptlayer.com/understanding-claude-code-hooks-documentation/)

---

#### Hook Performance Best Practices

**Guidelines:**
1. Keep PreToolUse hooks **fast** (<5 seconds)
2. Use `async: true` for slow operations
3. Cache expensive checks
4. Avoid network calls in sync hooks
5. Profile hook execution times

**Async Pattern:**
```json
{
  "type": "command",
  "command": "./run-tests.sh",
  "async": true,
  "timeout": 300
}
```

**Source:** [Builder.io](https://www.builder.io/blog/claude-code), [gend.co](https://www.gend.co/blog/configure-claude-code-hooks-automation)

---

#### Matcher Patterns Tips

**Remember:**
- Matchers are **regex**, not globs
- Use `|` for OR: `"Edit|Write"`
- Use `.*` for wildcard: `"Notebook.*"`
- Empty string or omit = match all: `"*"` or `""`

**MCP Tools:**
- Pattern: `mcp__<server>__<tool>`
- Match all from server: `"mcp__memory__.*"`
- Match operation: `"mcp__.*__write.*"`

**Source:** [Official Hooks Reference](https://code.claude.com/docs/en/hooks)

---

#### Testing Hooks Locally

**Best Practice:** Test hooks before deploying

```bash
# Create test input
echo '{
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": {"command": "rm -rf /tmp/test"}
}' > test-input.json

# Test hook
cat test-input.json | ./.claude/hooks/security-check.sh

# Check exit code
echo $?

# Check output
cat test-input.json | ./.claude/hooks/security-check.sh 2>&1
```

**Source:** Multiple community sources

---

#### Security Best Practices

**From Community:**

1. **Never trust input blindly**
   ```bash
   # Bad
   eval "$COMMAND"

   # Good
   if [[ "$COMMAND" =~ ^[a-zA-Z0-9\ ]+$ ]]; then
     echo "Safe command"
   fi
   ```

2. **Always quote variables**
   ```bash
   # Bad
   rm $FILE

   # Good
   rm "$FILE"
   ```

3. **Validate file paths**
   ```bash
   # Check for path traversal
   if [[ "$PATH" == *".."* ]]; then
     echo "Path traversal detected" >&2
     exit 2
   fi
   ```

4. **Use absolute paths**
   ```bash
   # Use CLAUDE_PROJECT_DIR
   SCRIPT="$CLAUDE_PROJECT_DIR/.claude/hooks/validate.sh"
   ```

5. **Skip sensitive files**
   ```bash
   # Check for sensitive patterns
   if [[ "$FILE" =~ \.(env|key|pem|secret)$ ]]; then
     echo "Skipping sensitive file" >&2
     exit 0
   fi
   ```

**Source:** [trailofbits config](https://github.com/trailofbits/claude-code-config), community best practices

---

#### Stop Hook Infinite Loop Prevention

**Problem:** Stop hooks can prevent Claude from ever stopping

**Solution:** Check `stop_hook_active` field

```bash
#!/bin/bash
INPUT=$(cat)
ACTIVE=$(echo "$INPUT" | jq -r '.stop_hook_active')

if [[ "$ACTIVE" == "true" ]]; then
  # Already in a stop hook loop, allow stopping
  exit 0
fi

# Your validation logic
# ...
```

**Source:** [Official docs](https://code.claude.com/docs/en/hooks), community discussions

---

#### Prompt Hook vs Agent Hook

**When to Use Prompt:**
- Simple yes/no decision
- Can evaluate from input JSON alone
- Fast response needed (<30s)

**When to Use Agent:**
- Need to inspect files
- Requires tool use (Read, Grep, Glob)
- Complex verification (run tests, check builds)
- Can tolerate longer execution (up to 60s+)

**Example Decision Matrix:**

| Task | Hook Type | Reason |
|------|-----------|--------|
| Check command safety | Prompt | Can evaluate from command string |
| Verify tests pass | Agent | Needs to read test output |
| Validate file format | Prompt | Can check from file content in input |
| Ensure builds succeed | Agent | Needs to run build command |

**Source:** [DataCamp](https://www.datacamp.com/tutorial/claude-code-hooks), community examples

---

#### Environment Variable Persistence (SessionStart)

**Special Feature:** `CLAUDE_ENV_FILE`

```bash
#!/bin/bash
# SessionStart hook

if [ -n "$CLAUDE_ENV_FILE" ]; then
  # Individual exports
  echo 'export NODE_ENV=production' >> "$CLAUDE_ENV_FILE"
  echo 'export DEBUG=true' >> "$CLAUDE_ENV_FILE"

  # Or capture all changes
  ENV_BEFORE=$(export -p | sort)
  source ~/.nvm/nvm.sh
  nvm use 20
  ENV_AFTER=$(export -p | sort)
  comm -13 <(echo "$ENV_BEFORE") <(echo "$ENV_AFTER") >> "$CLAUDE_ENV_FILE"
fi
```

**Source:** [Official Hooks Reference](https://code.claude.com/docs/en/hooks)

---

## Undocumented Features & Tips

### From Community Experimentation

#### 1. Hook Deduplication

**Discovery:** Claude Code automatically deduplicates identical hooks

**Implication:** Multiple matcher groups can safely reference same script

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit",
        "hooks": [{"type": "command", "command": "./format.sh"}]
      },
      {
        "matcher": "Write",
        "hooks": [{"type": "command", "command": "./format.sh"}]
      }
    ]
  }
}
```

Only runs `./format.sh` once if both matchers fire.

---

#### 2. Multiple Hook Sources Merge

**Discovery:** Hooks from different sources (user, project, plugin) all execute

**Pattern:**
- User settings: Global security hook
- Project settings: Project-specific formatting
- Plugin: Additional validation

All three run in parallel.

---

#### 3. Hook Execution Order

**Finding:** Hooks from same source run in definition order

**Within a matcher group:**
```json
{
  "hooks": [
    {"command": "./first.sh"},   // Runs first
    {"command": "./second.sh"},  // Runs second
    {"command": "./third.sh"}    // Runs third
  ]
}
```

**Across matcher groups:** Parallel execution (order not guaranteed)

---

#### 4. Silent Failures

**Issue:** Hooks can fail silently if JSON malformed

**Detection:**
```bash
# Add validation to hooks
INPUT=$(cat)
if ! echo "$INPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON input" >&2
  exit 1
fi
```

---

#### 5. Timeout Behavior

**Findings:**
- Default timeout: 600s (10 minutes) for command hooks
- Timeout doesn't kill background processes
- Use `timeout` command for reliable limits:

```bash
#!/bin/bash
timeout 30s npm test || {
  echo "Test suite timed out" >&2
  exit 1
}
```

---

## Common Pitfalls (From Community)

### 1. Forgetting to Quote Variables

**Wrong:**
```bash
rm $FILE
```

**Right:**
```bash
rm "$FILE"
```

---

### 2. Using Glob Instead of Regex in Matcher

**Wrong:**
```json
{
  "matcher": "*.js"  // This is regex, not glob!
}
```

**Right:**
```json
{
  "matcher": ".*\\.js$"  // Proper regex
}
```

---

### 3. Not Checking if Fields Exist

**Wrong:**
```bash
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')
# FILE could be null or "null" string
```

**Right:**
```bash
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
if [[ -z "$FILE" ]]; then
  echo "No file path in input" >&2
  exit 0
fi
```

---

### 4. Blocking PostToolUse Actions

**Wrong Expectation:**
```bash
# PostToolUse hook
exit 2  # Won't prevent file write (already happened!)
```

**Right Understanding:**
- PostToolUse can only provide feedback
- Use PreToolUse to block
- Use PostToolUse for cleanup/formatting

---

### 5. Not Handling Async Output

**Wrong:**
```json
{
  "async": true,
  "decision": "block"  // Ignored in async hooks!
}
```

**Right:**
```json
{
  "async": true
  // No decision control, only systemMessage/additionalContext
}
```

---

## Tool Integration Examples (From Community)

### 1. Slack Integration

```bash
#!/bin/bash
INPUT=$(cat)
EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
MESSAGE=$(echo "$INPUT" | jq -r '.last_assistant_message // "Event triggered"' | head -c 100)

curl -X POST "$SLACK_WEBHOOK_URL" \
  -H 'Content-Type: application/json' \
  -d "{\"text\":\"Claude Code [$EVENT]: $MESSAGE\"}"
```

---

### 2. GitHub Actions Trigger

```bash
#!/bin/bash
# Trigger workflow via repository_dispatch
curl -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/owner/repo/dispatches" \
  -d '{"event_type":"claude-code-hook"}'
```

---

### 3. Sentry Error Tracking

```bash
#!/bin/bash
INPUT=$(cat)
ERROR=$(echo "$INPUT" | jq -r '.error // empty')

if [[ -n "$ERROR" ]]; then
  sentry-cli send-event \
    --message "Claude Code Error: $ERROR" \
    --level error
fi
```

---

### 4. Jira Ticket Creation

```bash
#!/bin/bash
# On Stop hook - create ticket if TODOs found
TODOS=$(rg "TODO" --count --no-heading | wc -l)

if [[ $TODOS -gt 0 ]]; then
  jira issue create \
    --project=PROJ \
    --type=Task \
    --summary="Address $TODOS TODOs from Claude Code session"
fi
```

---

## Performance Benchmarks (Community Data)

### Hook Execution Times

**From community reports:**

| Hook Type | Typical Time | Max Recommended |
|-----------|--------------|-----------------|
| PreToolUse (validation) | <100ms | 5s |
| PostToolUse (formatting) | 1-5s | 30s |
| SessionStart (context load) | 100-500ms | 10s |
| Stop (verification) | 5-30s | 60s |
| Agent hooks | 10-60s | 120s |

---

## Advanced Patterns (From Community)

### 1. Conditional Hook Chains

```bash
#!/bin/bash
# Run multiple formatters conditionally
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path')

case "$FILE" in
  *.py)
    black "$FILE" && isort "$FILE" && flake8 "$FILE"
    ;;
  *.ts|*.tsx)
    prettier --write "$FILE" && eslint --fix "$FILE"
    ;;
  *.go)
    gofmt -w "$FILE" && golangci-lint run "$FILE"
    ;;
esac
```

---

### 2. Hook State Management

```bash
#!/bin/bash
# Track hook executions across session
STATE_FILE="/tmp/claude-hooks-state.json"

# Initialize if doesn't exist
if [[ ! -f "$STATE_FILE" ]]; then
  echo '{"executions":0}' > "$STATE_FILE"
fi

# Increment counter
jq '.executions += 1' "$STATE_FILE" > "${STATE_FILE}.tmp"
mv "${STATE_FILE}.tmp" "$STATE_FILE"

# Read state
EXECUTIONS=$(jq -r '.executions' "$STATE_FILE")
echo "Hook executed $EXECUTIONS times this session" >&2
```

---

### 3. Multi-Stage Validation

```bash
#!/bin/bash
# Stage 1: Fast validation
if ! validate_fast; then
  exit 2
fi

# Stage 2: Medium validation (only if fast passes)
if ! validate_medium; then
  exit 2
fi

# Stage 3: Expensive validation (only if needed)
if [[ "$REQUIRE_DEEP_CHECK" == "true" ]]; then
  if ! validate_deep; then
    exit 2
  fi
fi

exit 0
```

---

## Sources

**Blog Posts & Tutorials:**
- https://www.eesel.ai/blog/settings-json-claude-code
- https://aiorg.dev/blog/claude-code-hooks
- https://blog.promptlayer.com/understanding-claude-code-hooks-documentation/
- https://www.gend.co/blog/configure-claude-code-hooks-automation
- https://claude.com/blog/how-to-configure-hooks
- https://www.builder.io/blog/claude-code
- https://www.datacamp.com/tutorial/claude-code-hooks
- https://smartscope.blog/en/generative-ai/claude/claude-code-hooks-guide/
- https://levelup.gitconnected.com/how-to-use-claude-code-bed73d273638
- https://shipyard.build/blog/claude-code-cheat-sheet/
- https://docs.gitbutler.com/features/ai-integration/claude-code-hooks
- https://claudelog.com/

**GitHub Issues:**
- anthropics/claude-code#5489 - Environment variable substitution
- anthropics/claude-code#9567 - Hook environment variables empty

Research date: 2026-02-21
