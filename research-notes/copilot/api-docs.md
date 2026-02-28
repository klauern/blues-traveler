# GitHub Copilot API Documentation - Programmatic Access

**Research Date:** 2026-02-21

## Overview

GitHub Copilot provides **two distinct API surfaces**:
1. **REST API** - For management and metrics (administrative)
2. **SDK/Runtime API** - For programmatic integration (agent capabilities)

There is **NO public API** for direct code generation capabilities outside of these contexts.

## REST API - Management & Metrics

### Purpose
The REST API is designed for **organizational management and monitoring**, not for accessing Copilot's code generation capabilities.

### Available Endpoints

#### 1. User Management & Billing
- Manage GitHub Copilot Business subscriptions
- Manage seats and team access
- Billing assignments
- User seat allocation

**Documentation:** [REST API endpoints for Copilot user management](https://docs.github.com/en/rest/copilot/copilot-user-management)

#### 2. Metrics & Usage
- Get aggregated metrics for various GitHub Copilot features
- View Copilot usage metrics
- Track feature adoption across organization

**Example Use Cases:**
- Monitor completion acceptance rates
- Track active user counts
- Analyze feature usage patterns

**Documentation:** [REST API endpoints for Copilot metrics](https://docs.github.com/en/rest/copilot/copilot-metrics)

#### 3. Enterprise Reporting
- Retrieve download links for Copilot enterprise usage metrics
- Get comprehensive usage data for specific days
- Export data for business intelligence

**Documentation:** [REST API endpoints for Copilot usage metrics](https://docs.github.com/rest/copilot/copilot-usage-metrics)

### Limitations

**IMPORTANT:** GitHub Copilot does NOT provide an official API for programmatic access to its code generation capabilities. The REST API is purely for administrative purposes.

As stated in community discussions:
> "GitHub Copilot does not offer a public API for direct programmatic access from languages like Python or Java."

## GitHub Copilot SDK - Agent Runtime

### Overview

The GitHub Copilot SDK exposes the **Copilot Agent Runtime** - the same intelligence that powers Copilot CLI - as a library you can consume in your applications.

**GitHub Repository:** [github/copilot-sdk](https://github.com/github/copilot-sdk)

**Status:** Technical Preview (functional for development and testing, may not yet be suitable for production use)

### Supported Languages

- **Node.js** (TypeScript/JavaScript)
- **Python**
- **Go**
- **.NET** (C#)

### Architecture

All SDKs communicate with the Copilot CLI server via **JSON-RPC protocol**.

### Event System

The SDK provides a comprehensive event system for listening to agent actions:

#### Session Events
- `session.created`
- `session.deleted`
- `session.updated`
- `session.foreground`
- `session.background`

#### Message Events
- `assistant.message_delta` - Streaming incremental content
- `AssistantMessageEvent` - Complete messages
- `SessionIdleEvent` - Session state changes

#### Event Subscription Pattern
```javascript
// Example: Subscribe to specific events
session.on("assistant.message_delta", (event) => {
  // Process incremental content
});
```

The event system supports **pattern matching** and **filtering by event type**.

### Session Hooks in SDK

The SDK exposes `SessionHooks` for programmatic automation:

#### OnPreToolUse Callback
```csharp
// .NET SDK Example
SessionHooks.OnPreToolUse = (toolName, toolArgs) => {
  // Implement logic for:
  // - Auditing
  // - Permission decisions
  // - Security validation
  return new ToolPermissionDecision {
    Decision = "allow" // or "deny"
  };
};
```

**Capabilities:**
- Audit tool usage
- Implement custom permission logic
- Block dangerous operations
- Log compliance data

### Documentation Links

- [GitHub Copilot SDK Repository](https://github.com/github/copilot-sdk)
- [Getting Started Guide](https://github.com/github/copilot-sdk/blob/main/docs/getting-started.md)
- [.NET SDK Documentation](https://deepwiki.com/github/copilot-sdk/6.4-.net-sdk)
- [Examples and Cookbook](https://deepwiki.com/github/copilot-sdk/10-examples-and-cookbook)

### Recent Updates (February 2026)

**InfoQ Article:** [GitHub Copilot SDK Lets Developers Integrate Copilot CLI's Engine into Apps](https://www.infoq.com/news/2026/02/github-copilot-sdk/)

**Blog Post:** [Building Custom AI Tooling with the GitHub Copilot SDK for .NET](https://benjamin-abt.com/blog/2026/02/03/github-copilot-sdk-dotnet-tooling/)

Key capabilities highlighted:
- Production-tested agent runtime
- Programmatic invocation
- Custom AI tooling integration
- Session management
- Tool execution control

## Third-Party Workarounds

### copilot-api Project
**Repository:** [ericc-ch/copilot-api](https://github.com/ericc-ch/copilot-api)

**Description:** "Turn GitHub Copilot into OpenAI/Anthropic API compatible server. Usable with Claude Code!"

This is a **community-built** workaround that:
- Wraps Copilot's internal API
- Provides OpenAI-compatible endpoints
- Enables programmatic access outside official channels
- **Not officially supported by GitHub**

## Comparison: REST API vs SDK

| Feature | REST API | SDK |
|---------|----------|-----|
| **Purpose** | Management & metrics | Agent integration |
| **Code Generation** | ❌ No | ✅ Yes (via agent runtime) |
| **Hooks/Callbacks** | ❌ No | ✅ Yes (SessionHooks) |
| **Event System** | ❌ No | ✅ Yes (session/message events) |
| **Languages** | HTTP/Any | Node.js, Python, Go, .NET |
| **Use Case** | Admin/reporting | Building AI apps |
| **Status** | GA | Technical Preview |
| **Authentication** | GitHub tokens | GitHub tokens |

## Webhook Support

**Does GitHub Copilot support webhooks?**

**No.** GitHub Copilot does **not** support traditional webhooks for receiving notifications about events.

**Alternatives:**
- Use hooks (JSON-based shell commands) in repositories
- Use SDK event system for programmatic notification
- Poll REST API for metrics changes (not real-time)

## Rate Limits

GitHub Copilot has rate limits that apply to API usage:
- Documented at: [Rate limits for GitHub Copilot](https://docs.github.com/en/copilot/concepts/rate-limits)
- Limits vary by subscription tier
- Applies to completions, chat, and other interactive features

**Workarounds:** [Work Around GitHub Copilot Rate Limits](https://markaicode.com/bypass-github-copilot-rate-limits/)

## Summary

### What IS Possible:
- ✅ Administrative management via REST API
- ✅ Metrics and usage tracking via REST API
- ✅ Programmatic agent integration via SDK
- ✅ Event-driven automation via SDK event system
- ✅ Session hooks for auditing and security via SDK
- ✅ JSON-based hooks in repositories

### What Is NOT Possible:
- ❌ Direct API access to code generation (without SDK)
- ❌ Traditional webhook notifications
- ❌ Public API for arbitrary programmatic completion requests
- ❌ Standalone code generation API separate from agent context

**Key Takeaway:** GitHub Copilot's "API" is split between administrative REST endpoints and SDK-based agent integration. For automation and callbacks, use the hooks system (repository-based JSON configuration) or the SDK (programmatic integration).
