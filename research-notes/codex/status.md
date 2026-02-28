# OpenAI Codex - Current Status (February 2026)

## Active Development Status

**Codex is ACTIVE and evolving** - not deprecated. However, there have been significant changes and deprecations of specific components:

## Recent Deprecations (2025-2026)

### 1. Codex-Mini-Latest Model
- **Deprecation Date**: November 17, 2025
- **Removal Date**: February 12, 2026
- **Impact**: The codex-mini-latest model was removed from the API
- **Migration Path**: Users should upgrade to newer Codex models (GPT-5.2-Codex, GPT-5.3-Codex)

### 2. Chat Completions API Support
- **Deprecation Timeline**: Early February 2026
- **Status**: Support for the Chat Completions API is deprecated and will be removed in future releases
- **Migration Path**: Users are encouraged to use the **Responses API** instead
- **CLI Impact**: The Codex CLI emits deprecation warnings if configured to use chat/completions API; will transition to hard error in February 2026

## Current Model Lineup (February 2026)

### Available Models

1. **GPT-5.3-Codex** (Recommended)
   - Available for ChatGPT-authenticated Codex sessions
   - Supported in: Codex app, CLI, IDE extension, Codex Cloud
   - **API access coming soon** (as of Feb 2026)
   - Best for most coding tasks

2. **GPT-5.2-Codex**
   - Available on Responses API
   - Multiple reasoning tiers: medium, high, xhigh
   - Medium reasoning recommended for balanced intelligence and speed

3. **Codex-Mini-Latest**
   - **DEPRECATED** (removed Feb 12, 2026)
   - Was priced at $1.50 per 1M input tokens, $6 per 1M output tokens
   - 75% prompt caching discount available

## Current Architecture

Codex has evolved from a simple API model to a comprehensive autonomous coding agent platform:

### Three Primary Surfaces

1. **Codex CLI** - Terminal-based coding agent
2. **Codex App** (Web/Desktop) - Cloud-based IDE integration
3. **Codex IDE Extensions** - VS Code, Xcode, etc.

### Core Infrastructure

- **App Server**: Powers all Codex experiences with unified bidirectional protocol
- **Responses API**: Modern API for interacting with Codex models
- **MCP Server**: Model Context Protocol support for agent orchestration
- **Cloud Runtime**: Isolated sandbox environments for autonomous work

## What Codex Is Today (2026)

Codex has transformed from a code completion API into an **autonomous software engineering agent** that:

- Works asynchronously on complete features
- Operates in isolated cloud sandbox environments
- Handles multi-file projects independently
- Supports parallel task execution with multiple agents
- Provides automation capabilities for CI/CD workflows
- Integrates with GitHub, Jira, and other development tools

## Key Differences from Original Codex (2021)

| Aspect | Original Codex (2021) | Current Codex (2026) |
|--------|----------------------|---------------------|
| **Nature** | Code generation model (API) | Autonomous software agent |
| **Execution** | Synchronous completions | Asynchronous task delegation |
| **Scope** | Code snippets | Complete features/projects |
| **Integration** | API-only | CLI, App, IDE extensions, API |
| **Automation** | Manual API calls | Built-in automations, hooks, webhooks |
| **Multi-tasking** | Single request | Parallel agents in cloud sandboxes |

## Migration Guidance

### If You're Using Old Codex API

1. **Migrate from Chat Completions to Responses API**
   - Update API endpoints from `/v1/chat/completions` to Responses API
   - Follow migration guide in official documentation

2. **Upgrade Model References**
   - Replace `codex-mini-latest` with `gpt-5.2-codex` or newer
   - Consider GPT-5.3-Codex when API access becomes available

3. **Consider Modern Integration Patterns**
   - Explore Codex CLI for terminal workflows
   - Try Codex App for cloud-based autonomous work
   - Use App Server for custom integrations

## Future Direction

Based on OpenAI's public communications, Codex is positioned as:

- Primary agentic coding solution from OpenAI
- Foundation for autonomous software development
- Platform for multi-agent orchestration
- Integration point for CI/CD automation

The product is actively developed with regular updates, new features, and expanding capabilities.

## Official Resources

- Main Documentation: https://developers.openai.com/codex
- Changelog: https://developers.openai.com/codex/changelog/
- API Deprecations: https://platform.openai.com/docs/deprecations
- GitHub Repository: https://github.com/openai/codex

## Last Updated

February 21, 2026
