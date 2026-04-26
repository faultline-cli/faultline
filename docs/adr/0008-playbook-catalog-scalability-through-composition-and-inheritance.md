# ADR 0008: Playbook Catalog Scalability Through Composition And Inheritance

- Status: Accepted
- Date: 2026-04-25

## Context

As the bundled playbook catalog grew from 77 playbooks at v0.3.0 to 178+ by v0.4.1, naive copy-paste expansion becomes untenable: duplicate matching logic, divergent guidance text, and brittle maintenance.

The repository already handles duplicates and conflict resolution through `make review` (`internal/playbooks/conflicts.go`). But a larger catalog still creates pressure toward:

- redundant match blocks across similar failure signatures
- guidance text that diverges for the same root cause in different environments
- no clear mechanism to express "this is a narrower variant of that"

The scaling model is documented in [docs/playbooks.md](../playbooks.md) and the inheritance resolver is implemented in `internal/playbooks/inheritance.go`.

## Decision

Faultline's catalog scaling model relies on four mechanisms, each addressing a different axis of growth:

1. **`extends` inheritance**: A child playbook can reference a parent by id. The engine merges parent fields into the child at load time, with child values overriding parent values. Cycles are rejected. This keeps a shared root cause in one place while allowing environment-specific guidance and match constraints to live in child playbooks.

2. **Reusable signal fragments** via `match.use` or `faultline-matchers.yaml`: Common evidence patterns can be factored out and composed into multiple playbooks without duplicating signal logic.

3. **`match.partial` groups**: Multiple soft signals that are individually inconclusive can be combined into a decisive threshold. This avoids broadening individual `match` rules while capturing multi-signal evidence.

4. **Explicit constraint fields** (`tags`, `stage_hints`, `context_filters`, `source`): Narrowing without modifying match patterns. Playbooks stay broad in their core signal and use constraint fields to sharpen scope.

The catalog scales through composition of these primitives rather than accumulating monolithic playbooks or unconstrained duplication.

## Current Status

As of v0.4.1, the `extends` field is supported by the engine (`internal/playbooks/inheritance.go`) and present in the model schema, but no bundled playbooks use inheritance yet. The current catalog uses flat composition (constraint fields + conflict review gates) to maintain quality. Inheritance-based playbook families are the intended next step for common root causes with environment-specific variants.

## Consequences

- New playbooks for the same root cause should extend a shared base rather than duplicating match and guidance content
- The conflict review gate (`make review`) remains the primary quality check until inheritance-based families are adopted
- Inheritance cycles are a load-time error; they will not silently produce wrong results
- The composition model gives a deterministic equivalent of "components" without introducing a second matching language
- `match.partial` is the correct tool for combining weak signals; broadening individual match rules to compensate is an anti-pattern

## References

- [docs/playbooks.md](../playbooks.md) — Scaling model and authoring guidance
- [internal/playbooks/inheritance.go](../../internal/playbooks/inheritance.go)
- [internal/playbooks/conflicts.go](../../internal/playbooks/conflicts.go)
- [docs/releases/v0.4.0.md](../releases/v0.4.0.md) — Catalog grew from 77 to 123 playbooks
- [docs/releases/v0.4.1.md](../releases/v0.4.1.md) — Catalog at 170 bundled playbooks
