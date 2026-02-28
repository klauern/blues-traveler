# GitHub Copilot Official Documentation - Hooks & Automation

**Research Date:** 2026-02-21

## Overview

GitHub Copilot has a comprehensive **hooks system** that enables execution of custom shell commands at strategic points in an agent's workflow. This is GitHub's native automation and callback mechanism for Copilot agents.

## Hook System Architecture

### What Are Hooks?

Hooks enable you to execute custom shell commands at strategic points in an agent's workflow, such as when:
- An agent session starts or ends
- A prompt is entered
- A tool is called (before and after)
- An error occurs

Hooks receive detailed information about agent actions via JSON input, enabling context-aware automation.

## Available Hook Types

GitHub Copilot provides **seven distinct hook types**:

### 1. **sessionStart**
- **Triggers:** When a new agent session begins or when resuming an existing session
- **Input Fields:**
  - `timestamp`: Unix milliseconds
  - `cwd`: Current working directory
  - `source`: "new", "resume", or "startup"
  - `initialPrompt`: User's opening prompt
- **Output:** Ignored (no return value processed)
- **Use Cases:**
  - Initialize environments
  - Log session starts for auditing
  - Validate project state
  - Set up temporary resources

### 2. **sessionEnd**
- **Triggers:** When agent session completes or terminates
- **Input Fields:**
  - `timestamp`: Unix milliseconds
  - `cwd`: Current working directory
  - `reason`: "complete", "error", "abort", "timeout", or "user_exit"
- **Output:** Ignored
- **Use Cases:**
  - Cleanup temporary resources
  - Generate and archive session reports/logs
  - Send notifications about session completion

### 3. **userPromptSubmitted**
- **Triggers:** When the user submits input to the agent
- **Input Fields:**
  - `timestamp`: Unix milliseconds
  - `cwd`: Current working directory
  - `prompt`: Exact submitted text
- **Output:** Ignored (prompt modification not supported)
- **Use Cases:**
  - Log user requests for auditing
  - Usage analysis
  - Compliance tracking

### 4. **preToolUse** (Most Powerful)
- **Triggers:** Before tool invocation
- **Input Fields:**
  - `timestamp`: Unix milliseconds
  - `cwd`: Current working directory
  - `toolName`: Tool identifier (bash, edit, view, create)
  - `toolArgs`: JSON string containing tool parameters
- **Output JSON:**
  ```json
  {
    "permissionDecision": "deny",
    "permissionDecisionReason": "Human-readable explanation"
  }
  ```
- **Decision Values:** "allow", "deny", "ask" (only "deny" currently processed)
- **Use Cases:**
  - Block dangerous commands
  - Enforce security policies and coding standards
  - Require approval for sensitive operations
  - Log tool usage for compliance
  - Security validation

### 5. **postToolUse**
- **Triggers:** After tool execution completes
- **Input Fields:**
  - `timestamp`: Unix milliseconds
  - `cwd`: Current working directory
  - `toolName`: Executed tool name
  - `toolArgs`: Tool parameters
  - `toolResult`: Contains `resultType` ("success", "failure", "denied") and `textResultForLlm`
- **Output:** Ignored (result modification not supported)
- **Use Cases:**
  - Log results
  - Track statistics
  - Monitor performance metrics

### 6. **errorOccurred**
- **Triggers:** When agent encounters errors during operation
- **Input Fields:**
  - `timestamp`: Unix milliseconds
  - `cwd`: Current working directory
  - `error`: Object containing `message`, `name`, and `stack`
- **Output:** Ignored
- **Use Cases:**
  - Error logging
  - External notifications (Slack, email)
  - Error tracking integration

### 7. **agentStop / subagentStop**
- **Triggers:** When main agent or subagent finishes responding
- **Use Cases:**
  - Cleanup
  - Final logging
  - Post-processing

## Configuration Format

### File Location
Create `hooks.json` files in:
- Repository: `.github/hooks/*.json` (must be on default branch)
- VS Code workspace: `.claude/settings.json`, `.claude/settings.local.json`
- User level: `~/.claude/settings.json`

Workspace configurations take precedence over user-level settings.

### Basic Structure
```json
{
  "version": 1,
  "hooks": {
    "hookType": [{
      "type": "command",
      "bash": "./scripts/hook.sh",
      "powershell": "./scripts/hook.ps1",
      "cwd": "scripts",
      "env": {"KEY": "value"},
      "timeoutSec": 30
    }]
  }
}
```

### Multiple Hooks Per Event
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {"type": "command", "bash": "./security-check.sh"},
      {"type": "command", "bash": "./audit-log.sh"}
    ]
  }
}
```

### Platform-Specific Commands
Each hook supports:
- `bash`: Unix/Linux/macOS command
- `powershell`: Windows command
- `command`: Generic command or OS-specific overrides (`windows`, `linux`, `osx`)

## Script Communication Protocol

### Input: JSON via stdin
Hooks receive JSON input through stdin. Example script:
```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
```

### Output: JSON via stdout
```bash
# Success with decision
echo '{"permissionDecision":"deny","permissionDecisionReason":"Dangerous command"}' | jq -c

# Exit codes:
# 0: Success - parse stdout as JSON
# 2: Blocking error - stop and report to model
# Other: Non-blocking warning - continue processing
```

## Key Capabilities

### Security Control
- Block sensitive operations
- Enforce coding standards
- Built-in secret scanning to prevent credential leaks
- Pattern-based command blocking

### Compliance
- Generate audit trails
- Custom validation rules
- Usage tracking
- Tool execution logs

### Automation
- Environment initialization
- Resource cleanup
- External integrations (Slack, email, webhooks)
- Performance monitoring

## Best Practices

### Performance
- Keep execution under 5 seconds
- Set appropriate timeouts (default: 30 seconds)
- Avoid resource exhaustion

### Security
- Validate and sanitize all inputs
- Use proper shell escaping
- Avoid logging sensitive data
- Review scripts carefully, especially from untrusted repositories

### JSON Handling
- Use `jq` for JSON parsing in bash
- Output single-line compact JSON: `jq -c`
- Structured logging via JSON Lines format

### Development Workflow
1. Start with logging-only hooks (no deny rules)
2. Review logs to understand usage patterns
3. Gradually introduce deny rules based on actual usage
4. Test hooks locally by piping test JSON input

## Common Security Patterns to Block

Download-and-execute patterns that should never be auto-executed:
```bash
curl ... | bash
wget ... | sh
PowerShell iex commands
rm -rf (dangerous deletions)
sudo (elevated privileges)
mkfs (filesystem formatting)
```

## Documentation Links

- [About Hooks](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-hooks)
- [Hooks Configuration Reference](https://docs.github.com/en/copilot/reference/hooks-configuration)
- [Using Hooks with GitHub Copilot Agents](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/use-hooks)
- [VS Code Hooks Implementation](https://code.visualstudio.com/docs/copilot/customization/hooks)
- [Copilot CLI Hooks Tutorial](https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks)

## Summary

GitHub Copilot has **native hook/callback support** through a JSON-based configuration system. The hooks system is production-ready, well-documented, and provides powerful automation capabilities for security, compliance, and workflow customization.
