# Architecture Decision Records

This directory captures durable architectural decisions that shaped the current Faultline repository.

These ADRs are intentionally brief and point back to the authoritative product and architecture docs plus the git history that introduced or clarified the decision.

## Index

- [0001: Deterministic CLI Architecture](0001-deterministic-cli-architecture.md)
- [0002: Catalog Packaging And Extra Pack Composition](0002-catalog-packaging-and-extra-pack-composition.md)
- [0003: Structured Playbooks With Markdown Rendering](0003-structured-playbooks-with-markdown-rendering.md)
- [0004: Deterministic Review And Ranking Gates](0004-deterministic-review-and-ranking-gates.md)
- [0005: Command Maturity And Release Boundary](0005-command-maturity-and-release-boundary.md)
- [0006: Snapshot-Backed Companion Surface Validation](0006-snapshot-backed-companion-surface-validation.md)
- [0007: Silent Failures As First-Class Detection](0007-silent-failures-as-first-class-detection.md)
- [0008: Playbook Catalog Scalability Through Composition And Inheritance](0008-playbook-catalog-scalability-through-composition-and-inheritance.md)
- [0009: CI Failure Ontology As Catalog Taxonomy](0009-ci-failure-ontology-as-catalog-taxonomy.md)

## Source Material

- [SYSTEM.md](../../SYSTEM.md)
- [docs/architecture.md](../architecture.md)
- [docs/release-boundary.md](../release-boundary.md)
- [docs/fixture-corpus.md](../fixture-corpus.md)
- [docs/playbooks.md](../playbooks.md)
- [docs/distribution.md](../distribution.md)
- [docs/silent-failures.md](../silent-failures.md)
- [docs/releases/v0.4.0.md](../releases/v0.4.0.md) and [v0.4.1.md](../releases/v0.4.1.md)
- [ROADMAP.md](../../ROADMAP.md) — v0.4 theme "Fleet-safe deterministic operations"
- Git commits: `27b6919` (pack split), `22b1a83` (markdown rendering), `fccab7e` (conflict management), `68f6587` (IDF weighting), `b7e9a9b` (silent failures), `d709932` (silent detector hardening), `039f7d0` (ontology applied to catalog)