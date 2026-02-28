# OpenAI Codex - Official Documentation Overview

## Primary Documentation Resources

### Main Documentation Hub
- **URL**: https://developers.openai.com/codex
- **Purpose**: Central hub for all Codex documentation
- **Content**: Architecture, guides, API references, tutorials

### API Documentation
- **Responses API**: https://platform.openai.com/docs/guides/
- **Webhooks Guide**: https://developers.openai.com/api/docs/guides/webhooks/
- **Code Generation**: https://platform.openai.com/docs/guides/code-generation
- **Background Mode**: https://platform.openai.com/docs/guides/background
- **Streaming Responses**: https://platform.openai.com/docs/guides/streaming-responses

### Changelog & Updates
- **Codex Changelog**: https://developers.openai.com/codex/changelog/
- **Developer Changelog**: https://developers.openai.com/changelog/
- **Deprecations**: https://platform.openai.com/docs/deprecations
- **Model Release Notes**: https://help.openai.com/en/articles/9624314-model-release-notes

## Codex CLI Documentation

### Getting Started
- **Quickstart**: https://developers.openai.com/codex/quickstart/
- **CLI Features**: https://developers.openai.com/codex/cli/features/
- **Command Reference**: https://developers.openai.com/codex/cli/reference/

### CLI Capabilities
- Interactive coding sessions in terminal
- Non-interactive execution with `codex exec`
- GitHub integration for PR reviews and CI/CD
- Skills system for repeatable tasks
- Multi-agent parallel execution

## Codex App Documentation

### App Server Architecture
- **Main Guide**: https://developers.openai.com/codex/app-server/
- **Architecture Blog**: https://openai.com/index/unlocking-the-codex-harness/
- **App Automations**: https://developers.openai.com/codex/app/automations/

### Codex Web
- **Cloud Runtime**: https://developers.openai.com/codex/cloud/
- **Web Interface**: https://developers.openai.com/codex/web/

## Configuration & Customization

### Configuration Files
- **Config Reference**: https://developers.openai.com/codex/config-reference/
- **Advanced Configuration**: https://developers.openai.com/codex/config-advanced/
- **Sample Configuration**: https://developers.openai.com/codex/config-sample/

### Key Configuration Topics
- `~/.codex/config.toml` - Main configuration file
- `AGENTS.md` - Repository-specific instructions
- MCP server configuration
- Tool and skill configuration
- Notification hooks

## Integration Guides

### Model Context Protocol (MCP)
- **MCP Overview**: https://developers.openai.com/codex/mcp/
- **Docs MCP Server**: https://developers.openai.com/resources/docs-mcp/

### GitHub Integration
- **GitHub Guide**: https://developers.openai.com/codex/integrations/github/
- **GitHub Action**: https://developers.openai.com/codex/github-action/
- **GitHub Action Repo**: https://github.com/openai/codex-action

### Agents SDK
- **SDK Guide**: https://developers.openai.com/codex/guides/agents-sdk/
- **Multi-Agent Workflows**: https://developers.openai.com/cookbook/examples/codex/codex_mcp_agents_sdk/building_consistent_workflows_codex_cli_agents_sdk

## Cookbook & Examples

### OpenAI Cookbook
- **Base URL**: https://developers.openai.com/cookbook/
- **Codex Examples**: https://developers.openai.com/cookbook/examples/codex/

### Specific Cookbook Recipes

#### Prompting & Configuration
- **Prompting Guide**: https://developers.openai.com/cookbook/examples/gpt-5/codex_prompting_guide/
- **Code Modernization**: https://developers.openai.com/cookbook/examples/codex/code_modernization/

#### Automation Examples
- **Jira-GitHub Integration**: https://developers.openai.com/cookbook/examples/codex/jira-github/
- **Auto-fix CI Failures**: https://developers.openai.com/cookbook/examples/codex/autofix-github-actions/
- **Code Review with SDK**: https://developers.openai.com/cookbook/examples/codex/build_code_review_with_codex_sdk/

#### Multi-Agent Patterns
- **Codex CLI & Agents SDK**: https://cookbook.openai.com/examples/codex/codex_mcp_agents_sdk/building_consistent_workflows_codex_cli_agents_sdk

## Model Documentation

### Current Models
- **Models Overview**: https://developers.openai.com/codex/models/
- **GPT-5.3-Codex**: Recommended for most coding tasks
- **GPT-5.2-Codex**: Multiple reasoning tiers (medium/high/xhigh)

### Model Features
- Trained on large, complex codebases
- Multi-file reasoning capabilities
- Code modernization specialization
- Test and documentation generation
- Refactoring and migration support

## GitHub Resources

### Official Repositories

#### Main Codex Repository
- **URL**: https://github.com/openai/codex
- **Purpose**: Open-source Codex CLI and App Server
- **Languages**: Go, Python, TypeScript, Swift, Kotlin support
- **Issues**: https://github.com/openai/codex/issues
- **Discussions**: https://github.com/openai/codex/discussions

#### Codex Action
- **URL**: https://github.com/openai/codex-action
- **Purpose**: GitHub Action for CI/CD integration
- **Use Cases**: Automated reviews, CI failure fixes, testing

#### Skills Catalog
- **URL**: https://github.com/openai/skills
- **Purpose**: Community skills library
- **Content**: Reusable instructions, scripts, and resources

## Workflows & Automations

### Workflow Documentation
- **Main Guide**: https://developers.openai.com/codex/workflows/
- **Automations**: https://developers.openai.com/codex/app/automations/

### Automation Capabilities
- Scheduled automations (e.g., daily issue triage)
- CI/CD integration hooks
- Event-driven triggers
- Background task execution
- Multi-agent orchestration

## API Architecture Documentation

### App Server Protocol
- JSON-RPC over stdio (JSONL format)
- Bidirectional communication
- Server-initiated requests (approvals, notifications)
- Event streaming support
- Semantic event types

### Integration Patterns
1. **Local Clients**: Platform-specific binaries, stdio communication
2. **Partner Integrations**: Decoupled client/server releases
3. **Web Runtime**: Container-based, HTTP + Server-Sent Events

## Webhooks Documentation

### OpenAI API Webhooks
- **Guide**: https://developers.openai.com/api/docs/guides/webhooks/
- **Standard**: Follows Standard Webhooks specification
- **Events**: Batch completion, background completion, fine-tuning completion

### Event Types
- Batch job completions
- Background response generation
- Fine-tuning job status
- Thread archival/unarchival (App Server v2)

## Community Resources

### Developer Community
- **Forum**: https://community.openai.com/
- **Codex Discussions**: Dedicated section for Codex topics
- **Best Practices**: https://community.openai.com/t/best-practices-for-using-codex/

### Official Blogs
- **Introducing Codex**: https://openai.com/index/introducing-codex/
- **Codex App Launch**: https://openai.com/index/introducing-the-codex-app/
- **App Server Architecture**: https://openai.com/index/unlocking-the-codex-harness/
- **GPT-5.3-Codex-Spark**: https://openai.com/index/introducing-gpt-5-3-codex-spark/
- **How OpenAI Uses Codex**: https://openai.com/business/guides-and-resources/how-openai-uses-codex/

## Support & Help

### Help Center
- **Model Release Notes**: https://help.openai.com/en/articles/9624314-model-release-notes
- **API Documentation**: https://platform.openai.com/docs/

### Getting Help
1. Search official documentation first
2. Check GitHub discussions for community questions
3. Review cookbook examples for patterns
4. Consult changelog for recent changes
5. Contact OpenAI support for API/billing issues

## Documentation Structure Summary

```
developers.openai.com/codex/
├── quickstart/              # Getting started guide
├── cli/                     # CLI documentation
│   ├── features/
│   └── reference/
├── app/                     # App documentation
│   └── automations/
├── app-server/              # App Server architecture
├── web/                     # Codex Cloud/Web
├── integrations/            # Third-party integrations
│   └── github/
├── guides/                  # Integration guides
│   └── agents-sdk/
├── config-reference/        # Configuration docs
├── mcp/                     # MCP documentation
├── models/                  # Model documentation
├── changelog/               # Version history
└── workflows/               # Workflow patterns

developers.openai.com/cookbook/examples/codex/
├── prompting guides
├── automation examples
├── integration patterns
└── multi-agent workflows

platform.openai.com/docs/
├── guides/
│   ├── code-generation
│   ├── background
│   ├── streaming-responses
│   └── webhooks/
└── deprecations
```

## Key Takeaways for Developers

1. **Start Here**: Quickstart guide for initial setup
2. **CLI Users**: Focus on CLI features and command reference
3. **App Users**: Review app-server and automations
4. **Integrators**: Study App Server architecture and MCP guides
5. **Automation**: Check cookbook for CI/CD patterns
6. **API Users**: Responses API is the modern standard (not Chat Completions)

## Last Updated

February 21, 2026

## Next Steps

For hooks/callbacks research, pay special attention to:
- Webhooks documentation (event-driven patterns)
- App Server protocol (bidirectional communication)
- Automations (scheduled and triggered execution)
- GitHub Action integration (CI/CD hooks)
- Notification hooks in config
