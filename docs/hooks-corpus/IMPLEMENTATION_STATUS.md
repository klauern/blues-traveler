# Hooks Corpus Implementation Status

**Project:** Hooks & Automation Encyclopedia for AI Coding Assistants  
**Started:** 2026-02-21  
**Last Updated:** 2026-02-27  
**Status:** Research imported, core reference pages committed, advanced sections still being filled

## Current Snapshot

The repository now contains:

- Core reference pages for all five systems covering architecture, configuration, events, environment, scripting, and security.
- Authored examples guides for Copilot and Codex.
- Placeholder examples guides and staging directories for Claude, Cursor, and Gemini so navigation is stable while examples are still being written.
- Research note trees for `claude`, `cursor`, `copilot`, `codex`, and `gemini`.
- Comparison tables and a corpus contribution guide.

## What Is Complete

### Repository structure

- `docs/hooks-corpus/README.md` is the top-level navigation hub.
- `docs/hooks-corpus/comparison-tables/` contains committed comparison material.
- `research-notes/` contains durable source notes for all five systems.

### System documentation

| System | Core docs present | Examples guide | Advanced pages |
|--------|-------------------|----------------|----------------|
| Claude | Yes | Placeholder | Placeholder tracked in status docs |
| Cursor | Yes | Placeholder | Placeholder tracked in corpus |
| Copilot | Yes | Authored | Placeholder tracked in corpus |
| Codex | Yes | Authored | Placeholder tracked in corpus |
| Gemini | Yes | Placeholder | Placeholder tracked in corpus |

## What Is Still Placeholder Work

- Real worked examples in the `examples/` directories.
- System-specific integration pages.
- Troubleshooting playbooks.
- API reference pages.
- Detailed migration guides.

Those pages now exist either as placeholders or are tracked explicitly so that links remain valid and current status is obvious.

## Research Status

Phase 2 research is no longer "not started". It is materially complete enough to support draft docs:

- Official documentation was gathered for all systems.
- Public examples and community references were collected.
- Cross-system comparisons and gaps were recorded.
- Remaining work is curation and synthesis, not initial discovery.

## Merge-Readiness Checklist

- [x] Core hooks corpus pages committed
- [x] Research notes committed
- [x] Broken internal links removed or replaced
- [x] Workstation-specific absolute paths removed from committed docs
- [x] Empty examples directories anchored with committed README files
- [x] Docs validation task added
- [ ] Real runnable example scripts added for each system
- [ ] Advanced pages promoted from placeholder to authored reference

## Next Recommended Work

1. Replace placeholder example guides with runnable scripts in each `examples/` directory.
2. Promote the highest-value placeholder pages: troubleshooting and migration first.
3. Expand docs validation if new doc sections introduce richer link patterns.
