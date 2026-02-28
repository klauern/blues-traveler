# Hooks Corpus Contributing Guide

This corpus is intentionally commit-friendly: keep navigation stable, keep claims sourced, and keep research notes durable enough to survive outside one workstation.

## Research Notes Policy

Commit research notes when they are reusable project knowledge:

- Keep notes that summarize official docs, community findings, examples, or gaps that future contributors will need.
- Keep notes that are cited from `docs/hooks-corpus/`.
- Keep notes that explain contradictions or unknowns across systems.
- Do not publish scratchpad-only prompts, transient local experiments, or workstation-specific instructions.
- If a note is local-only, do not link to it from committed corpus pages.

## Examples Policy

- Do not advertise an `examples/` directory unless it exists in git.
- If runnable examples are not ready yet, add a clearly marked placeholder guide or README instead of a broken link.
- Prefer committed examples that are copy-paste friendly and tied back to a source note.

## Portability Rules

- Do not commit workstation-specific absolute paths.
- Use repo-relative links for cross-references inside the repository.
- Prefer plain relative paths when a prose reference is clearer than a markdown link.

## Validation

Run:

```bash
task docs-validate
```

The validator currently checks:

- forbidden workstation-specific absolute paths
- markdown links that point at missing local files or directories

## Status Discipline

- Mark placeholders as placeholders.
- Update `docs/hooks-corpus/IMPLEMENTATION_STATUS.md` when the corpus meaningfully changes state.
- Keep per-system status pages aligned with what is actually committed, not what is planned.
