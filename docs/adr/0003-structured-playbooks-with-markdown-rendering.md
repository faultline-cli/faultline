# ADR 0003: Structured Playbooks With Markdown Rendering

- Status: Accepted
- Date: 2026-04-13

## Context

Faultline playbooks must carry both machine-meaningful matching data and human-readable diagnosis guidance. Prior to the current rendering model, those concerns were less clearly separated. Commit `22b1a83` introduced the markdown-centric content model and the dedicated renderer package, and the current authoring guidance in [docs/playbooks.md](../playbooks.md) codifies the boundary.

## Decision

Playbooks use structured fields for matching and ranking, while markdown-capable fields carry operator-facing diagnosis, fix, and validation content.

The runtime renders the same content model through:

- raw terminal-oriented text
- markdown output
- stable JSON that bypasses terminal styling

Presentation stays separate from matching semantics.

## Consequences

- Playbooks remain auditable because matching logic is not hidden inside prose.
- Human-facing output can improve without changing scoring behavior.
- Markdown output becomes a first-class handoff format for issue templates, notes, and agent workflows.
- Renderer concerns stay isolated from engine and matcher logic.

## References

- [docs/playbooks.md](../playbooks.md)
- [docs/architecture.md](../architecture.md)
- [README.md](../../README.md)
- Git history: `22b1a83` Markdown for all