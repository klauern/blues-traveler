# Cursor Hooks: Knowledge Gaps and Unknown Areas

**Research Date:** 2026-02-21

This document identifies areas where information is incomplete, contradictory, or entirely missing from available documentation. These gaps require hands-on testing or clarification from Cursor developers.

---

## Critical Gaps (High Priority)

### 1. Timeout Behavior

**What We Know:**
- Timeout is configurable in hooks.json
- Format: `"timeout": 5000` (milliseconds)
- Example from community: 5000ms, 10000ms

**What We Don't Know:**
- ❌ What happens when timeout expires?
  - Does hook fail (deny)?
  - Does execution proceed (allow)?
  - Is there an error message?
- ❌ Is there a maximum timeout value?
- ❌ Is there a default timeout if not specified?
- ❌ How is timeout measured?
  - Wall clock time?
  - CPU time?
  - Does it include hook startup overhead?
- ❌ Can timeout be disabled (set to 0 or infinity)?

**Impact:** High - Critical for production hooks
**Testing Needed:** Yes - Create hook with deliberate delays

**Recommended Tests:**
```bash
# Test 1: Hook that sleeps longer than timeout
sleep 10  # with timeout: 5000

# Test 2: Hook with no timeout specified
# (check default behavior)

# Test 3: Hook with timeout: 0
# (check if disabled)
```

---

### 2. Error Recovery and Retry Logic

**What We Know:**
- Hooks that fail (non-zero exit + no JSON) → block action
- Invalid JSON → block with error
- Exit 0 + no JSON → allow

**What We Don't Know:**
- ❌ Are failed hooks retried automatically?
- ❌ Can users configure retry behavior?
- ❌ Is there exponential backoff?
- ❌ What happens to subsequent hooks if one fails?
  - Do they still run?
  - Or does first failure short-circuit?
- ❌ How are transient failures handled?
  - Network errors
  - Temporary file system issues
  - External API timeouts
- ❌ Can hooks return "retry" status?

**Impact:** High - Affects reliability
**Testing Needed:** Yes - Simulate failures

**Recommended Tests:**
```bash
# Test 1: Hook that randomly fails
exit $((RANDOM % 2))

# Test 2: Multiple hooks where first fails
# Check if second runs

# Test 3: Hook that fails with different exit codes
exit 1   # vs exit 2 vs exit 255
```

---

### 3. Performance Characteristics

**What We Know:**
- Hooks are synchronous (block agent)
- Community recommendation: < 2 seconds
- Community max: < 10 seconds

**What We Don't Know:**
- ❌ What is Cursor's internal timeout default?
- ❌ How does hook latency impact user experience?
  - Is there a progress indicator?
  - Can users cancel slow hooks?
- ❌ Is there a performance budget?
- ❌ Are hooks executed in parallel for different events?
- ❌ What is the startup overhead?
  - Process spawn time
  - Shell initialization
- ❌ Are there caching mechanisms?
- ❌ How does hook performance scale with file size/complexity?

**Impact:** High - User experience critical
**Testing Needed:** Yes - Benchmarking required

**Recommended Benchmarks:**
```bash
# Measure hook execution time
time ./hook.sh <<< '{"command":"test"}'

# Test with various payload sizes
# - Small: 100 bytes
# - Medium: 10KB
# - Large: 1MB

# Parallel event test
# Trigger multiple events simultaneously
```

---

### 4. Multi-Hook Execution Order and Coordination

**What We Know:**
- Multiple hooks per event execute in array order
- First deny may block action

**What We Don't Know:**
- ❌ Exact short-circuit behavior
  - Does first `deny` skip remaining hooks?
  - Or do all hooks always run?
- ❌ How are results combined?
  - If hook 1 allows, hook 2 denies → which wins?
  - If hook 1 denies with message, hook 2 denies with different message → which message shown?
- ❌ Can hooks communicate with each other?
  - Via filesystem (generation_id)?
  - Via environment variables?
  - Via IPC?
- ❌ Are hook results aggregated?
- ❌ What happens if hooks conflict?
  - Hook 1: `permission: "allow", userMessage: "OK"`
  - Hook 2: `permission: "deny", userMessage: "BLOCKED"`

**Impact:** Medium-High - Multi-hook setups affected
**Testing Needed:** Yes - Multiple hook scenarios

**Recommended Tests:**
```json
{
  "beforeShellExecution": [
    { "command": "./hook1-allow.sh" },
    { "command": "./hook2-deny.sh" },
    { "command": "./hook3-allow.sh" }
  ]
}
```
Test: Which decision wins? Do all 3 run?

---

### 5. Security Model and Sandbox Interaction

**What We Know:**
- Hooks run OUTSIDE Cursor's agent sandbox
- Hooks have full user permissions
- Malicious hooks can compromise system

**What We Don't Know:**
- ❌ Can hooks be sandboxed separately?
- ❌ Are there plans for hook sandboxing?
- ❌ What security checks does Cursor perform on hooks?
  - Code signing?
  - Checksum validation?
  - Permission verification?
- ❌ Can hooks escape to modify Cursor internals?
- ❌ Are there rate limits to prevent hook abuse?
- ❌ How does Cursor verify hook authenticity?
- ❌ Can hooks be disabled remotely (kill switch)?

**Impact:** Critical - Security implications
**Testing Needed:** Ethical hacking assessment

---

## Important Gaps (Medium Priority)

### 6. Hook State Management

**What We Know:**
- conversation_id and generation_id provided
- Hooks can write to filesystem

**What We Don't Know:**
- ❌ Is there persistent state between hooks?
- ❌ Where should hooks store temporary data?
  - `/tmp/` recommended?
  - Workspace-specific location?
- ❌ Who cleans up hook state?
  - Cursor?
  - Hook scripts?
  - User responsibility?
- ❌ How long do conversation_id and generation_id persist?
- ❌ Are there state size limits?

**Impact:** Medium - Affects stateful hooks
**Testing Needed:** State persistence tests

---

### 7. Environment Variable Propagation

**What We Know:**
- Cursor doesn't document env vars passed to hooks
- Standard shell env available (`HOME`, `USER`, `PWD`)

**What We Don't Know:**
- ❌ What Cursor-specific env vars are available?
  - `CURSOR_VERSION`?
  - `CURSOR_SESSION_ID`?
  - `CURSOR_WORKSPACE`?
- ❌ Are user-set env vars passed through?
- ❌ Can hooks modify env for subsequent hooks?
- ❌ Are there security-sensitive env vars filtered?
- ❌ How are secrets handled in env?

**Impact:** Medium - Integration patterns affected
**Testing Needed:** env var inspection

**Recommended Test:**
```bash
#!/bin/bash
# Print all env vars to debug
env > /tmp/cursor-hook-env.txt
```

---

### 8. Hook Installation and Discovery

**What We Know:**
- Hooks defined in `.cursor/hooks.json`
- Multiple levels: project, user, global

**What We Don't Know:**
- ❌ How does Cursor discover hooks.json?
  - Does it watch for file changes?
  - Or only read on startup?
- ❌ Can hooks be hot-reloaded?
  - Or is Cursor restart required?
- ❌ What happens with malformed hooks.json?
  - Silent failure?
  - Error message?
  - Cursor won't start?
- ❌ Are there hook installation helpers?
- ❌ Can hooks be disabled without deleting?
- ❌ Is there a hook validation command?

**Impact:** Medium - Developer experience
**Testing Needed:** Installation workflows

---

### 9. Cross-Platform Behavior

**What We Know:**
- Cursor runs on macOS, Linux, Windows
- Hooks are shell scripts

**What We Don't Know:**
- ❌ How do hooks work on Windows?
  - WSL required?
  - PowerShell support?
  - cmd.exe support?
- ❌ Are there platform-specific limitations?
- ❌ How are paths resolved cross-platform?
- ❌ Are there Windows-specific best practices?
- ❌ Do timeouts behave identically on all platforms?

**Impact:** Medium - Windows users affected
**Testing Needed:** Windows testing required

---

### 10. JSON Schema Versioning

**What We Know:**
- Current version: 1
- Schema URL: https://unpkg.com/cursor-hooks/schema/hooks.schema.json

**What We Don't Know:**
- ❌ What changes in version 2 (if planned)?
- ❌ Will version 1 be deprecated?
- ❌ How will migrations be handled?
- ❌ Are there version compatibility guarantees?
- ❌ Can multiple versions coexist?

**Impact:** Low-Medium - Future-proofing
**Testing Needed:** Monitor release notes

---

## Minor Gaps (Low Priority)

### 11. Logging and Observability

**What We Know:**
- Hooks can write to stderr or log files
- No built-in logging framework

**What We Don't Know:**
- ❌ Does Cursor log hook execution?
- ❌ Where are Cursor's hook logs?
- ❌ Is there a hook execution history?
- ❌ Can hook metrics be exported?
- ❌ Are there debugging tools for hooks?
- ❌ Can hooks emit structured logs?

**Impact:** Low - Affects debugging
**Testing Needed:** Log inspection

---

### 12. Hook Metadata

**What We Know:**
- Hooks defined with `command` field only

**What We Don't Know:**
- ❌ Can hooks have names/descriptions?
- ❌ Can hooks declare dependencies?
- ❌ Can hooks specify required tools/versions?
- ❌ Is there a hook manifest format?
- ❌ Can hooks be tagged/categorized?

**Impact:** Low - Nice-to-have features
**Testing Needed:** Not critical

---

### 13. Testing Frameworks

**What We Know:**
- Community has ad-hoc testing patterns
- No official testing tools

**What We Don't Know:**
- ❌ Are there official hook testing utilities?
- ❌ Can hooks be tested in isolation?
- ❌ Is there a mock Cursor environment?
- ❌ Are there integration test patterns?
- ❌ Can hooks be CI/CD tested?

**Impact:** Low-Medium - Developer productivity
**Testing Needed:** Framework development

---

### 14. Rate Limiting

**What We Know:**
- Cursor has API rate limits (Pro plan)

**What We Don't Know:**
- ❌ Are hooks counted against rate limits?
- ❌ Can hooks trigger rate limiting?
- ❌ Is there a hook execution quota?
- ❌ Are there throttling mechanisms?

**Impact:** Low - Edge cases
**Testing Needed:** High-volume hook tests

---

## Contradictions and Ambiguities

### 15. Permission Field vs Continue Field

**Contradiction:**
- Some events use `permission` (beforeShellExecution)
- Some use `continue` (beforeSubmitPrompt)
- Documentation doesn't explain why

**Questions:**
- ❌ Why two different patterns?
- ❌ Can they be mixed?
- ❌ What happens if both are specified?

**Testing Needed:**
```json
{
  "permission": "deny",
  "continue": true
}
```
Which takes precedence?

---

### 16. Ask Permission Fallback

**Ambiguity:**
- Blues-traveler documents `ask` fallback to `allow`
- Cursor native `ask` prompts user

**Questions:**
- ❌ Does Cursor have a timeout for user response?
- ❌ What if user ignores prompt?
- ❌ Can `ask` be globally disabled?
- ❌ Is there an audit log of user decisions?

**Testing Needed:** User interaction tests

---

### 17. afterFileEdit Response Behavior

**Contradiction:**
- Docs say "no response expected"
- But hooks can output JSON
- Unclear what Cursor does with it

**Questions:**
- ❌ Is JSON output ignored?
- ❌ Or processed for logging?
- ❌ Can afterFileEdit block (retroactively)?

**Testing Needed:**
```bash
# afterFileEdit hook returns deny
echo '{"permission":"deny"}'
```
What happens?

---

## Undocumented Features

### 18. Variable Interpolation

**Observed in Examples:**
```json
{
  "command": "script.sh \"${command}\" \"${file}\""
}
```

**Questions:**
- ❌ What variables are available?
  - `${command}` ✓ (confirmed)
  - `${file}` ✓ (confirmed)
  - `${workspace}` ?
  - `${generation_id}` ?
- ❌ Is this shell expansion or Cursor feature?
- ❌ Are there escaping rules?

**Testing Needed:** Interpolation tests

---

### 19. Timeout Units

**Ambiguity in Examples:**
```json
// Example 1
{ "timeout": 5000 }  // milliseconds?

// Example 2 (blues-traveler)
{ "timeout": 5 }     // seconds?
```

**Questions:**
- ❌ What are the units?
- ❌ Is it documented consistently?

**Testing Needed:** Empirical timeout tests

---

## Future Roadmap Gaps

### 20. Beta Status Duration

**What We Know:**
- Hooks introduced October 2025 (Cursor 1.7)
- Still beta as of February 2026

**What We Don't Know:**
- ❌ When will hooks exit beta?
- ❌ What needs to happen for GA?
- ❌ Will API change before GA?
- ❌ Are there breaking changes planned?

**Impact:** Medium-High - Affects production adoption
**Source:** Need official roadmap

---

### 21. Async Hooks

**Community Request:**
- Multiple users want async hooks
- Current hooks are synchronous only

**Questions:**
- ❌ Are async hooks planned?
- ❌ What would async semantics be?
- ❌ How would results be delivered?

**Impact:** Medium - Performance improvement
**Source:** Feature request tracking

---

### 22. Hook Marketplace

**Community Pattern:**
- Many GitHub repos sharing hooks
- No centralized registry

**Questions:**
- ❌ Will Cursor provide hook marketplace?
- ❌ Are there plans for curated hooks?
- ❌ Will there be hook signing/verification?

**Impact:** Low - Community ecosystem
**Source:** Feature requests

---

## Hands-On Testing Plan

To fill critical gaps, the following tests should be conducted:

### Priority 1: Timeout Behavior
```bash
# Test script
#!/bin/bash
sleep_time=${1:-10}
echo "Sleeping for $sleep_time seconds..." >&2
sleep $sleep_time
echo '{"permission":"allow"}'
```

**hooks.json:**
```json
{
  "beforeShellExecution": [
    { "command": "./timeout-test.sh 15", "timeout": 5000 }
  ]
}
```

**Expected Questions Answered:**
- What happens when hook times out?
- Is there an error message?
- Does action proceed or block?

---

### Priority 2: Multi-Hook Coordination
```json
{
  "beforeShellExecution": [
    { "command": "./hook-allow.sh" },
    { "command": "./hook-deny.sh" },
    { "command": "./hook-allow.sh" }
  ]
}
```

**Expected Questions Answered:**
- Do all hooks run?
- Which decision wins?
- Are all messages shown?

---

### Priority 3: State Sharing
```bash
# Hook 1 (beforeShellExecution)
echo "data" > /tmp/cursor-$generation_id.state

# Hook 2 (stop)
cat /tmp/cursor-$generation_id.state
```

**Expected Questions Answered:**
- Does generation_id persist?
- Can hooks share state?
- Who cleans up?

---

### Priority 4: Environment Variables
```bash
#!/bin/bash
env | grep -i cursor > /tmp/cursor-env.txt
env >> /tmp/all-env.txt
```

**Expected Questions Answered:**
- What Cursor-specific env vars exist?
- What's available to hooks?

---

### Priority 5: Error Scenarios
```bash
# Test various failure modes
exit 1      # Non-zero, no JSON
exit 0      # Zero, no JSON
echo "invalid json" && exit 0
echo '{"permission":"invalid"}' && exit 0
```

**Expected Questions Answered:**
- How are errors handled?
- What error messages appear?
- Does behavior match documentation?

---

## Testing Infrastructure Needs

To comprehensively test Cursor hooks, we need:

1. **Mock Cursor Environment**
   - Simulate Cursor's hook execution
   - Test hooks without IDE

2. **Automated Test Suite**
   - Hook behavior tests
   - Integration tests
   - Performance benchmarks

3. **CI/CD Integration**
   - Run tests on commits
   - Cross-platform testing

4. **Observability Tools**
   - Hook execution tracing
   - Performance monitoring
   - Error tracking

---

## Documentation Improvement Opportunities

### For Cursor Team
1. **Expand Official Docs**
   - Document timeout behavior
   - Clarify error handling
   - Explain multi-hook coordination
   - Provide performance guidelines

2. **Add Examples**
   - Official hook repository
   - Best practices guide
   - Common patterns

3. **Provide Tools**
   - Hook validator
   - Testing framework
   - Debugging utilities

### For Community
1. **Testing Framework**
   - Build mock environment
   - Share test patterns

2. **Best Practices Document**
   - Performance guidelines
   - Security checklist
   - Common pitfalls

3. **Hook Registry**
   - Curated examples
   - Verified hooks
   - Security audits

---

## Summary of Gaps by Impact

### Critical (Requires Urgent Attention)
1. Timeout behavior
2. Error recovery
3. Security model
4. Multi-hook coordination

### Important (Should Be Addressed)
5. Performance characteristics
6. State management
7. Environment variables
8. Installation/discovery

### Nice-to-Have (Can Wait)
9. Cross-platform details
10. Logging/observability
11. Testing frameworks
12. Hook metadata

### Future (Roadmap Dependent)
13. Beta graduation
14. Async hooks
15. Hook marketplace

---

## Recommended Next Steps

1. **Empirical Testing** (Priority 1-5 tests above)
2. **Cursor Team Engagement** (request clarifications)
3. **Community Collaboration** (share findings)
4. **Documentation Updates** (fill gaps as discovered)
5. **Testing Framework Development** (infrastructure)

---

## Contributing to Gap Closure

If you test any of these gaps, please:

1. Document your findings
2. Share test scripts
3. Update this document
4. Submit PRs to blues-traveler
5. Report to Cursor team

---

**Last Updated:** 2026-02-21
**Status:** Living document - update as gaps are filled
**Contact:** Open issues at https://github.com/klauern/blues-traveler/issues
