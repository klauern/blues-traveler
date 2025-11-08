# Dual-Message Response Wrapper Design

## Overview

Blues Traveler needs to support **separate messages for end-users and AI agents** to enable better Cursor IDE compatibility and improved user experience. Currently, hooks return single messages via the `cchooks` library, which cannot differentiate between:

- **User messages**: Simple, friendly explanations (e.g., "Code formatting failed")
- **Agent messages**: Technical details for the AI to learn from (e.g., "Black formatter exited with code 2: invalid syntax on line 42")

This design describes a wrapper layer that adds dual-message support without forking the `cchooks` library.

## Design Goals

1. **Wrap, Don't Fork**: Extend `cchooks` responses without modifying the library
2. **Backward Compatible**: Existing single-parameter calls continue working unchanged
3. **Flexible API**: Support both single and dual-message patterns through variadic parameters
4. **Cross-Platform Ready**: Enable Cursor JSON response format support
5. **Type Safe**: Maintain type safety for Pre/Post tool use responses
6. **Clean Integration**: Minimal changes to existing hook code

## Architecture

### High-Level Flow

```text
Hook Code (security.go, format.go, etc.)
         ↓
         Block("user msg", "agent msg")  ← Call dual-message wrapper
         ↓
Response wrapper layer ([internal/core/response.go](mdc:internal/core/response.go))
         ↓
cchooks.Block(userMsg)  ← Single message to cchooks
         ↓
Metadata attached (userMessage, agentMessage in response object)
         ↓
Claude Code / Cursor route messages to appropriate recipients
```

### Core Design Decisions

#### 1. Wrapper Location
- **File**: `internal/core/response.go` (new)
- **Package**: `core` package
- **Rationale**: Centralized response handling, available to all hooks

#### 2. Variadic Parameter Pattern

```go
// Single parameter: same message to both audiences
BlockWithMessages("Operation failed")
// → userMessage: "Operation failed"
// → agentMessage: "Operation failed"

// Dual parameters: separate messages
BlockWithMessages("Formatting failed", "Black formatter error: ...")
// → userMessage: "Formatting failed"
// → agentMessage: "Black formatter error: ..."
```

**Rationale**: Intuitive API that handles 95% of use cases with a single call while supporting advanced scenarios.

#### 3. Message Routing Mechanism

Two implementation approaches (to be finalized during implementation phase):

**Option A: Response Wrapper Struct** (Recommended)
```go
type DualMessageResponse struct {
    underlying   cchooks.PreToolUseResponseInterface // Original response
    userMessage  string
    agentMessage string
}

// Implement cchooks.PreToolUseResponseInterface to wrap and add metadata
func (r *DualMessageResponse) Permission() string {
    return r.underlying.Permission()
}
// Add metadata-carrying methods
```

##### Option B: Metadata in Environment Variables

- Store userMessage/agentMessage in environment during response
- Claude Code reads and routes based on env vars
- Simpler but less type-safe

**Recommendation**: Start with Option A (wrapper struct), fallback to Option B if integration issues arise.

#### 4. Response Type Coverage

Wrapper functions needed for all hook response scenarios:

| Hook Type | Current Function | New Wrapper | Use Case |
|-----------|------------------|-------------|----------|
| PreToolUse (allow) | `cchooks.Approve()` | `ApproveWithMessages()` | Security/fetch blocker - allow with optional messages |
| PreToolUse (deny) | `cchooks.Block(msg)` | `BlockWithMessages()` | Security - block with user-friendly + technical reason |
| PostToolUse (allow) | `cchooks.Allow()` | `AllowWithMessages()` | Format/vet pass - success with optional details |
| PostToolUse (deny) | `cchooks.PostBlock(msg)` | `PostBlockWithMessages()` | Format/vet failure - error with user + agent details |
| Permission (ask) | `cchooks.Ask()` | `AskWithMessages()` | Custom hooks - seek user approval with context |

## Detailed API Specification

### Function Signatures

```go
package core

// PreToolUse Response Wrappers
// ============================

// BlockWithMessages creates a blocking response for PreToolUse events.
// If agentMsg is omitted, userMsg is sent to both audiences.
func BlockWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface

// ApproveWithMessages creates an approval response for PreToolUse events.
// If agentMsg is omitted, userMsg is sent to both audiences.
// Used when hook permits the tool but wants to communicate context.
func ApproveWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface

// AskWithMessages creates a permission request for PreToolUse events.
// Returns *core.AskPreToolResponse so Cursor serializes `permission: "ask"`
// while Claude gracefully falls back to approve semantics.
func AskWithMessages(userMsg string, agentMsg ...string) cchooks.PreToolUseResponseInterface


// PostToolUse Response Wrappers
// =============================

// PostBlockWithMessages creates a blocking response for PostToolUse events.
// If agentMsg is omitted, userMsg is sent to both audiences.
func PostBlockWithMessages(userMsg string, agentMsg ...string) cchooks.PostToolUseResponseInterface

// AllowWithMessages creates an allow response for PostToolUse events.
// If agentMsg is omitted, userMsg is sent to both audiences.
// Used when hook allows continuation but wants to communicate status.
func AllowWithMessages(userMsg string, agentMsg ...string) cchooks.PostToolUseResponseInterface
```

### Usage Examples

#### Example 1: Security Hook (Deny with Technical Details)

```go
// Current code:
return cchooks.Block("blocked dangerous command pattern: sudo")

// New code:
return core.BlockWithMessages(
    "This command was blocked for security reasons.",
    "Blocked dangerous pattern: sudo. Type: privilege_escalation",
)
```

#### Example 2: Format Hook (Failure with Formatter Output)

```go
// Current code:
return cchooks.PostBlock(fmt.Sprintf("Formatting failed for %s: %v", filePath, err))

// New code:
return core.PostBlockWithMessages(
    fmt.Sprintf("Code formatting failed for %s", filePath),
    fmt.Sprintf("Black formatter failed: %s\nStderr: %v", filePath, err),
)
```

#### Example 3: Fetch Blocker (Block with Alternatives)

```go
// Current code:
return cchooks.Block("Failed to fetch: URL requires authentication. Use 'gh' or 'curl' with credentials.")

// New code:
return core.BlockWithMessages(
    "This URL requires authentication.",
    "URL prefix blocked: github.com/private\nAlternatives:\n  - Use 'gh' CLI\n  - Use 'curl' with token",
)
```

#### Example 4: Custom Config Hook (Allow with Status)

```go
// Current code:
return cchooks.Approve()

// New code (with optional context):
return core.ApproveWithMessages(
    "Hook 'linter' executed successfully",
    "Job completed in 245ms, 0 issues found",
)
```

## Implementation Roadmap

### Phase 1: Core Infrastructure
1. Create `internal/core/response.go` with wrapper functions
2. Define `DualMessageResponse` struct implementing `PreToolUseResponseInterface`
3. Implement variadic parameter handling
4. Add unit tests for wrapper functions

### Phase 2: Integration Points
1. Wire metadata transport through cchooks runner
2. Add environment variable extraction for message routing
3. Test with Claude Code hook execution

### Phase 3: Hook Migration (6 tasks)
Migrate each hook to use dual messages:
1. Security hook: Technical + user-friendly messages
2. Find-blocker: Command suggestions for agent
3. Fetch-blocker: URL details for agent
4. Format hook: Formatter-specific errors for agent
5. Vet hook: Type check details for agent
6. Config hook: Command output for agent

### Phase 4: Ask Mode Support
1. Implement `AskWithMessages()` wrapper
2. Add permission request handling
3. Test user prompt workflow

### Phase 5: Cursor JSON Support (blues-traveler-59)
1. Parse hook script JSON responses
2. Route to appropriate audience
3. Test cross-platform compatibility

## Message Routing Mechanism Details

### Runtime Flow

1. **Hook calls wrapper**:
   ```go
   return core.BlockWithMessages("User sees this", "Agent sees this")
   ```

2. **Wrapper creates metadata**:
   ```go
   resp := &DualMessageResponse{
       underlying: cchooks.Block("User sees this"),
       userMessage: "User sees this",
       agentMessage: "Agent sees this",
   }
   ```

3. **cchooks runner processes response**:
   - Calls `resp.Permission()` → "deny" (from underlying)
   - Calls `resp.Message()` → "User sees this"

4. **Metadata extraction** (implementation-specific):
   - Option A: Attach via response object fields (if modifying cchooks integration)
   - Option B: Store in HookContext environment for retrieval

5. **Claude Code routes messages**:
   - Display userMessage to UI
   - Send agentMessage to AI model

### Key Assumptions

- Claude Code hook runner has visibility into DualMessageResponse metadata
- Message routing happens in Claude Code, not in Blues Traveler
- Backward compatibility preserved: single message works everywhere

## Error Handling

### Validation Rules

```go
// Valid
BlockWithMessages("msg") → userMsg="msg", agentMsg="msg"
BlockWithMessages("user", "agent") → userMsg="user", agentMsg="agent"
BlockWithMessages("user", "agent", "extra") → ERROR (too many args)

// Empty strings handled
BlockWithMessages("") → userMsg="", agentMsg="" (valid, explicit silence)
BlockWithMessages("user", "") → userMsg="user", agentMsg="" (valid, context only)
```

### Error Cases

1. **Too many parameters**: Compile-time error (variadic bounds)
2. **Nil caller**: Runtime panic (same as cchooks)
3. **Message encoding**: UTF-8 validation (if Claude Code requires)

## Backward Compatibility Strategy

### Existing Code Continues Working

```go
// All existing calls remain valid and work unchanged
cchooks.Block("message")           // ✅ Still works
cchooks.PostBlock("message")       // ✅ Still works
cchooks.Approve()                  // ✅ Still works

// New wrapper calls live alongside
core.BlockWithMessages("msg")      // ✅ New code uses wrappers
core.BlockWithMessages("u", "a")   // ✅ Dual messages
```

### Migration Path

1. **Phase 1-2**: Wrappers available but unused
2. **Phase 3**: Migrate built-in hooks (one at a time)
3. **Phase 4**: Custom hooks can adopt at own pace
4. **No breaking changes**: Old and new code coexist indefinitely

## Testing Strategy

### Unit Tests

```go
// Test single parameter
TestBlockWithMessagesSingleParam() {
    resp := BlockWithMessages("error")
    assert resp.UserMessage == "error"
    assert resp.AgentMessage == "error"
}

// Test dual parameters
TestBlockWithMessagesDualParams() {
    resp := BlockWithMessages("user error", "agent detail")
    assert resp.UserMessage == "user error"
    assert resp.AgentMessage == "agent detail"
}

// Test interface compliance
TestBlockWithMessagesImplementsInterface() {
    var resp cchooks.PreToolUseResponseInterface = BlockWithMessages("msg")
    assert resp.Permission() == "deny"
}

// Test all wrapper functions
TestAllWrapperFunctions() {
    // ApproveWithMessages, AskWithMessages, PostBlockWithMessages, AllowWithMessages
}
```

### Integration Tests

```go
// Test with actual hook execution
TestSecurityHookDualMessages() {
    // Run security hook with BlockWithMessages
    // Verify both messages appear in audit log
}

// Test with cchooks runner
TestDualMessageResponseWithRunner() {
    // Pass DualMessageResponse to cchooks runner
    // Verify underlying response works correctly
}
```

## Migration Examples

### Before and After

#### Security Hook
```go
// BEFORE
func (h *SecurityHook) preToolUseHandler(...) cchooks.PreToolUseResponseInterface {
    if blocked, reason := h.checkCommand(...); blocked {
        return cchooks.Block(reason)  // ← Single message
    }
    return cchooks.Approve()
}

// AFTER
func (h *SecurityHook) preToolUseHandler(...) cchooks.PreToolUseResponseInterface {
    if blocked, reason, checkType := h.checkCommand(...); blocked {
        return core.BlockWithMessages(
            "This command was blocked for security reasons.",
            fmt.Sprintf("Pattern matched: %s (type: %s)", reason, checkType),
        )
    }
    return cchooks.Approve()
}
```

#### Format Hook
```go
// BEFORE
func (h *FormatHook) postToolUseHandler(...) cchooks.PostToolUseResponseInterface {
    if err := h.formatFile(filePath); err != nil {
        return cchooks.PostBlock(fmt.Sprintf("Formatting failed: %v", err))
    }
    return cchooks.Allow()
}

// AFTER
func (h *FormatHook) postToolUseHandler(...) cchooks.PostToolUseResponseInterface {
    if err := h.formatFile(filePath); err != nil {
        return core.PostBlockWithMessages(
            fmt.Sprintf("Code formatting failed for %s", filePath),
            fmt.Sprintf("Formatter error:\n%v\n\nCheck syntax and try again.", err),
        )
    }
    return cchooks.Allow()
}
```

## Open Questions & Decisions

| Question | Status | Decision |
|----------|--------|----------|
| Use wrapper struct vs env vars? | Design | Wrapper struct (Option A) - type-safe, cleaner |
| How does Claude Code access metadata? | Implementation | TBD during integration phase |
| Should ask mode be included in Phase 1? | Design | No, Phase 4 after base is stable |
| Support for custom hook scripts? | Design | Via JSON response parsing (blues-traveler-59) |
| Versioning of response format? | Future | Consider for Phase 4+ |

## References & Dependencies

- **cchooks library**: `github.com/brads3290/cchooks` - base response types
- **Cursor compatibility**: blues-traveler-59 (JSON response parsing)
- **Ask mode support**: blues-traveler-54 (permission requests)

## Future Enhancements

1. **Response Builder Pattern**: Fluent API for complex messages
   ```go
   core.NewBlockResponse().
       WithUserMessage("...").
       WithAgentMessage("...").
       WithSuggestion("...").
       Build()
   ```

2. **Message Formatting**: Markdown support for structured messages
   ```go
   core.BlockWithMessages(
       "Code formatting failed",
       "## Formatter Output\n```\n" + stderr + "\n```",
   )
   ```

3. **Response Metadata**: Additional context (timestamp, hook key, etc.)
   ```go
   resp.WithMetadata("formatter", "black").WithMetadata("exit_code", "2")
   ```

4. **Localization**: Support translated messages
   ```go
   core.BlockWithMessages(
       i18n.T("formatting_failed"),
       technicalDetails,
   )
   ```
