# ADR 0009: CI Failure Ontology As Catalog Taxonomy

- Status: Accepted
- Date: 2026-04-25

## Context

As the playbook catalog expands, the catalog needs a shared classification vocabulary that:

- answers "what gaps exist?" across domains and failure classes
- gives fixture organization a consistent ownership model
- guides contributors toward consistent playbook placement
- enables coverage metrics and gap analysis without changing matching behavior

Without a shared taxonomy, coverage gaps are opaque, playbooks accumulate in ad hoc category directories, and cross-cutting analysis (e.g. "how complete is auth coverage?") requires manual inspection.

The ontology design is documented across [docs/ontology.md](../ontology.md), [docs/ontology-quick-reference.md](../ontology-quick-reference.md), [docs/ontology-examples.md](../ontology-examples.md), and [docs/ontology-implementation.md](../ontology-implementation.md). Commit `039f7d0` applied the taxonomy to the existing catalog.

## Decision

Faultline adopts a five-level CI failure taxonomy as **additive, read-only playbook metadata**. The ontology does not drive matching; it classifies playbooks for human and tooling understanding.

The five levels are:

1. **Domain** — broad operational area (`dependency`, `runtime`, `container`, `auth`, `network`, `ci-config`, `test-runner`, `database`, `filesystem`, `platform`, `source`)
2. **Class** — specific failure family within a domain (e.g. `lockfile-drift`, `missing-executable`, `service-not-ready`)
3. **Mode** — concrete root cause distinguishing this playbook from neighbors in the same class (e.g. `npm-ci-requires-package-lock`, `node-missing`)
4. **Evidence pattern** — the deterministic signals that confirm the mode
5. **Remediation strategy** — the fix approach category

Playbooks declare their taxonomy membership through YAML fields (`domain`, `class`, `mode`). The fields are optional and additive: existing playbooks without ontology tags continue to work and match normally. Classification happens over time and does not require a flag-day migration.

As of v0.4.1, 170 of 181 bundled playbooks carry ontology metadata.

## Consequences

- Coverage analysis can be driven by catalog introspection rather than manual audit
- New playbooks must declare their domain, class, and mode as part of authoring
- The taxonomy is extensible: new domains and classes can be added without breaking existing playbooks
- Playbooks with identical class and mode should be considered candidates for consolidation or inheritance (see [ADR 0008](0008-playbook-catalog-scalability-through-composition-and-inheritance.md))
- Ontology metadata must not change matching behavior — any behavioral change is an engine concern, not a taxonomy concern
- The 11-domain, 40-class taxonomy defines the current coverage contract; gaps within this space are explicit and reportable

## References

- [docs/ontology.md](../ontology.md) — Full design, five-level hierarchy, schema
- [docs/ontology-quick-reference.md](../ontology-quick-reference.md) — Condensed reference
- [docs/ontology-implementation.md](../ontology-implementation.md) — Migration phases and contributor guidance
- [docs/ontology-examples.md](../ontology-examples.md) — Six complete annotated examples
- [ADR 0008](0008-playbook-catalog-scalability-through-composition-and-inheritance.md) — Composition model that the taxonomy informs
- Git history: `039f7d0` apply ontology metadata to existing catalog
